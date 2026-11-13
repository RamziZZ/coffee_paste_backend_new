package handler

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/store"
)

// PaymentCallbackHandler menerima notifikasi "pembayaran QRIS berhasil" dari payment gateway.
// Pelunasan QRIS tanpa gateway tetap lewat konfirmasi kasir (PUT /kasir/pesanan/{id}).
type PaymentCallbackHandler struct {
	orders   *store.OrderStore
	secret   string
	notifier *orderNotifier
}

func NewPaymentCallbackHandler(orders *store.OrderStore, secret string) *PaymentCallbackHandler {
	return &PaymentCallbackHandler{orders: orders, secret: secret}
}

// QRIS: POST /api/payments/qris/callback
// Header X-Callback-Secret wajib sama dengan QRIS_CALLBACK_SECRET.
// Body: { "orderId": "#CP00001", "amount": 55500, "status": "paid" }
// Status selain "paid" diabaikan (200). Endpoint ini 404 bila QRIS_CALLBACK_SECRET kosong.
func (h *PaymentCallbackHandler) QRIS(w http.ResponseWriter, r *http.Request) {
	if h.secret == "" {
		respondNotFound(w, "Endpoint tidak ditemukan")
		return
	}
	got := r.Header.Get("X-Callback-Secret")
	if subtle.ConstantTimeCompare([]byte(got), []byte(h.secret)) != 1 {
		httpx.Error(w, http.StatusUnauthorized, "Rahasia callback salah")
		return
	}

	var req struct {
		OrderID string `json:"orderId"`
		Amount  int64  `json:"amount"`
		Status  string `json:"status"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	if !strings.EqualFold(strings.TrimSpace(req.Status), "paid") {
		httpx.JSON(w, http.StatusOK, map[string]any{"ok": true, "ignored": true})
		return
	}
	id, ok := parseOrderCode(req.OrderID)
	if !ok {
		unprocessable(w, "orderId tidak valid")
		return
	}
	if req.Amount <= 0 {
		unprocessable(w, "amount tidak valid")
		return
	}

	order, err := h.orders.MarkQRISPaid(r.Context(), id, req.Amount)
	var ve store.ValidationError
	switch {
	case errors.Is(err, store.ErrNotFound):
		respondNotFound(w, "Pesanan tidak ditemukan")
	case errors.As(err, &ve):
		unprocessable(w, ve.Error())
	case err != nil:
		httpx.ServerError(w, err)
	default:
		h.notifier.changed(r.Context(), id)
		httpx.JSON(w, http.StatusOK, map[string]any{"ok": true, "orderId": order.ID, "paid": order.Paid})
	}
}
