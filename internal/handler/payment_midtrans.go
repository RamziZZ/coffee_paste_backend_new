package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/midtrans"
	"coffeepaste/internal/store"
)

// MidtransHandler menerima notifikasi pembayaran dari Midtrans.
type MidtransHandler struct {
	orders    *store.OrderStore
	serverKey string
	notifier  *orderNotifier

	// Pemeriksa status berkala (lihat midtrans_reconcile.go); nil = hanya mengandalkan webhook.
	checker  MidtransStatusChecker
	reportMu sync.Mutex
	reported map[string]bool
}

func NewMidtransHandler(orders *store.OrderStore, serverKey string) *MidtransHandler {
	return &MidtransHandler{orders: orders, serverKey: serverKey}
}

// flexString menerima nilai JSON berupa string atau angka.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" {
		*f = ""
		return nil
	}
	if strings.HasPrefix(s, `"`) {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*f = flexString(v)
		return nil
	}
	*f = flexString(s)
	return nil
}

// parseGrossAmount mengubah "55500.00" (atau "55500") menjadi 55500.
func parseGrossAmount(s string) (int64, bool) {
	whole, frac, _ := strings.Cut(strings.TrimSpace(s), ".")
	if strings.Trim(frac, "0") != "" { // rupiah tidak punya pecahan
		return 0, false
	}
	n, err := strconv.ParseInt(whole, 10, 64)
	return n, err == nil && n > 0
}

// Notification: POST /api/payments/midtrans/notification
// Dipanggil Midtrans saat status transaksi berubah. Keaslian diperiksa lewat signature_key
// (SHA512 dari order_id + status_code + gross_amount + server key). 404 bila server key kosong.
//
// Balasan 200 berarti "sudah diterima" (Midtrans berhenti mengulang), termasuk untuk
// notifikasi yang sengaja diabaikan. Hanya signature salah (401) dan error server (500)
// yang membuat Midtrans mencoba lagi.
func (h *MidtransHandler) Notification(w http.ResponseWriter, r *http.Request) {
	if h.serverKey == "" {
		respondNotFound(w, "Endpoint tidak ditemukan")
		return
	}

	var n struct {
		OrderID           flexString `json:"order_id"`
		StatusCode        flexString `json:"status_code"`
		GrossAmount       flexString `json:"gross_amount"`
		SignatureKey      flexString `json:"signature_key"`
		TransactionStatus flexString `json:"transaction_status"`
		FraudStatus       flexString `json:"fraud_status"`
	}
	if !decodeOrFail(w, r, &n) {
		return
	}

	ref := string(n.OrderID)
	if !midtrans.VerifySignature(ref, string(n.StatusCode), string(n.GrossAmount), string(n.SignatureKey), h.serverKey) {
		httpx.Error(w, http.StatusUnauthorized, "Signature tidak valid")
		return
	}

	ignore := func(reason string) {
		httpx.JSON(w, http.StatusOK, map[string]any{"ok": true, "ignored": reason})
	}

	switch strings.ToLower(string(n.TransactionStatus)) {
	case "settlement", "capture":
		if fs := strings.ToLower(string(n.FraudStatus)); fs != "" && fs != "accept" {
			log.Printf("Midtrans %s: fraud_status=%s, diabaikan", ref, fs)
			ignore("fraud_status")
			return
		}
		amount, ok := parseGrossAmount(string(n.GrossAmount))
		if !ok {
			unprocessable(w, "gross_amount tidak valid")
			return
		}
		order, err := h.orders.MarkMidtransPaid(r.Context(), ref, amount)
		var ve store.ValidationError
		switch {
		case errors.Is(err, store.ErrNotFound):
			// Mis. tombol "Test notification" di dashboard Midtrans.
			log.Printf("Midtrans %s: tagihan tidak dikenal, diabaikan", ref)
			ignore("unknown order")
		case errors.As(err, &ve):
			// Uang sudah masuk tetapi pesanan tak bisa dilunasi: butuh tindakan manual/refund.
			log.Printf("PERHATIAN Midtrans %s (Rp%d) dibayar tetapi tidak bisa dilunasi: %s", ref, amount, ve.Error())
			ignore(ve.Error())
		case err != nil:
			httpx.ServerError(w, err)
		default:
			h.notifier.changedByCode(r.Context(), order.ID)
			httpx.JSON(w, http.StatusOK, map[string]any{"ok": true, "orderId": order.ID, "paid": order.Paid})
		}

	case "expire":
		h.setStatus(w, r, ref, "expired")
	case "cancel", "deny", "failure":
		h.setStatus(w, r, ref, "failed")
	default: // pending dan status lain
		ignore("status " + string(n.TransactionStatus))
	}
}

func (h *MidtransHandler) setStatus(w http.ResponseWriter, r *http.Request, ref, status string) {
	if err := h.orders.SetMidtransStatus(r.Context(), ref, status); err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true, "status": status})
}
