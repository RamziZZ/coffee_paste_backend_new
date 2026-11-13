package handler

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"coffeepaste/internal/midtrans"
	"coffeepaste/internal/store"
)

// MidtransStatusChecker menanyakan status transaksi langsung ke Midtrans.
type MidtransStatusChecker interface {
	Status(ctx context.Context, orderID string) (*midtrans.TxStatus, error)
}

// WithStatusChecker mengaktifkan pemeriksaan status berkala sebagai cadangan webhook.
func (h *MidtransHandler) WithStatusChecker(c MidtransStatusChecker) *MidtransHandler {
	h.checker = c
	return h
}

// RunReconciler memeriksa semua tagihan QRIS yang masih pending setiap `every` sampai ctx selesai.
//
// Webhook Midtrans hanya sampai bila server dapat dijangkau dari internet dan Payment Notification
// URL sudah diisi di dashboard. Tanpa itu pesanan tetap "belum lunas" di web dan mobile walau di
// Midtrans sudah Settlement. Dengan pemeriksa ini server yang menanyakan sendiri ke Midtrans
// (koneksi keluar saja), lalu menandai lunas dan menyiarkannya ke kasir dan customer.
func (h *MidtransHandler) RunReconciler(ctx context.Context, every time.Duration) {
	if h.checker == nil || h.orders == nil {
		return
	}
	if every <= 0 {
		every = 3 * time.Second
	}
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			rctx, cancel := context.WithTimeout(ctx, 20*time.Second)
			if _, err := h.Reconcile(rctx); err != nil && ctx.Err() == nil {
				log.Printf("Midtrans: gagal memeriksa tagihan pending: %v", err)
			}
			cancel()
		}
	}
}

// Reconcile satu putaran pemeriksaan. Mengembalikan jumlah pesanan yang baru menjadi lunas.
func (h *MidtransHandler) Reconcile(ctx context.Context) (int, error) {
	if h.checker == nil || h.orders == nil {
		return 0, nil
	}
	refs, err := h.orders.PendingMidtransRefs(ctx, 50)
	if err != nil {
		return 0, err
	}
	paid := 0
	for _, ref := range refs {
		if ctx.Err() != nil {
			break
		}
		if h.alreadyReported(ref) {
			continue
		}
		st, err := h.checker.Status(ctx, ref)
		if errors.Is(err, midtrans.ErrTransactionNotFound) {
			continue // belum ada di Midtrans
		}
		if err != nil {
			log.Printf("Midtrans %s: gagal membaca status: %v", ref, err)
			continue
		}
		if h.applyStatus(ctx, ref, st) {
			paid++
		}
	}
	return paid, nil
}

func (h *MidtransHandler) alreadyReported(ref string) bool {
	h.reportMu.Lock()
	defer h.reportMu.Unlock()
	return h.reported[ref]
}

func (h *MidtransHandler) markReported(ref string) {
	h.reportMu.Lock()
	defer h.reportMu.Unlock()
	if h.reported == nil {
		h.reported = map[string]bool{}
	}
	h.reported[ref] = true
}

// applyStatus menerapkan status dari Midtrans. true bila pesanan baru saja menjadi lunas.
func (h *MidtransHandler) applyStatus(ctx context.Context, ref string, st *midtrans.TxStatus) bool {
	switch st.TransactionStatus {
	case "settlement", "capture":
		if st.FraudStatus != "" && st.FraudStatus != "accept" {
			log.Printf("Midtrans %s: fraud_status=%s, diabaikan", ref, st.FraudStatus)
			h.markReported(ref)
			return false
		}
		amount, ok := parseGrossAmount(st.GrossAmount)
		if !ok {
			log.Printf("Midtrans %s: gross_amount %q tidak valid", ref, st.GrossAmount)
			h.markReported(ref)
			return false
		}
		order, err := h.orders.MarkMidtransPaid(ctx, ref, amount)
		var ve store.ValidationError
		switch {
		case errors.Is(err, store.ErrNotFound):
			h.markReported(ref)
			return false
		case errors.As(err, &ve):
			// Uang sudah masuk tetapi pesanan tak bisa dilunasi: butuh tindakan manual/refund.
			log.Printf("PERHATIAN Midtrans %s (Rp%d) dibayar tetapi tidak bisa dilunasi: %s", ref, amount, ve.Error())
			h.markReported(ref)
			return false
		case err != nil:
			log.Printf("Midtrans %s: gagal menandai lunas: %v", ref, err)
			return false
		}
		log.Printf("Midtrans %s: settlement terdeteksi, pesanan %s lunas", ref, order.ID)
		h.notifier.changedByCode(ctx, order.ID)
		return true

	case "expire":
		h.setPendingStatus(ctx, ref, "expired")
	case "cancel", "deny", "failure":
		h.setPendingStatus(ctx, ref, "failed")
	}
	return false
}

func (h *MidtransHandler) setPendingStatus(ctx context.Context, ref, status string) {
	if err := h.orders.SetMidtransStatus(ctx, ref, status); err != nil {
		log.Printf("Midtrans %s: gagal menyimpan status %s: %v", ref, status, err)
		return
	}
	log.Printf("Midtrans %s: tagihan %s", ref, strings.ToLower(status))
}
