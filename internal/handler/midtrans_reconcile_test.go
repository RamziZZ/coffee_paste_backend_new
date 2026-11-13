package handler

import (
	"context"
	"testing"

	"coffeepaste/internal/midtrans"
)

type fakeChecker struct{ calls int }

func (f *fakeChecker) Status(ctx context.Context, orderID string) (*midtrans.TxStatus, error) {
	f.calls++
	return &midtrans.TxStatus{OrderID: orderID, TransactionStatus: "settlement", GrossAmount: "55500.00"}, nil
}

// Tanpa pemeriksa atau store, Reconcile tidak boleh panik dan tidak melakukan apa pun.
func TestReconcileDisabled(t *testing.T) {
	h := NewMidtransHandler(nil, testServerKey)
	if n, err := h.Reconcile(context.Background()); n != 0 || err != nil {
		t.Fatalf("Reconcile tanpa pemeriksa = %d, %v", n, err)
	}
	fc := &fakeChecker{}
	h.WithStatusChecker(fc)
	if n, err := h.Reconcile(context.Background()); n != 0 || err != nil || fc.calls != 0 {
		t.Fatalf("Reconcile tanpa store = %d, %v, calls=%d", n, err, fc.calls)
	}
}

func TestReportedRefsAreSkipped(t *testing.T) {
	h := NewMidtransHandler(nil, testServerKey)
	if h.alreadyReported("CP00001-1") {
		t.Fatal("ref baru tidak boleh dianggap sudah dilaporkan")
	}
	h.markReported("CP00001-1")
	if !h.alreadyReported("CP00001-1") {
		t.Fatal("ref yang sudah dilaporkan harus dilewati")
	}
}

// Status selain settlement/capture/expire/cancel tidak menyentuh database.
func TestApplyStatusPendingIgnored(t *testing.T) {
	h := NewMidtransHandler(nil, testServerKey)
	if h.applyStatus(context.Background(), "CP00001-1", &midtrans.TxStatus{TransactionStatus: "pending"}) {
		t.Fatal("pending tidak boleh dianggap lunas")
	}
	if h.applyStatus(context.Background(), "CP00001-1", &midtrans.TxStatus{TransactionStatus: "settlement", FraudStatus: "deny", GrossAmount: "55500.00"}) {
		t.Fatal("fraud_status deny tidak boleh dilunasi")
	}
	if h.applyStatus(context.Background(), "CP00001-2", &midtrans.TxStatus{TransactionStatus: "settlement", GrossAmount: "abc"}) {
		t.Fatal("gross_amount tidak valid tidak boleh dilunasi")
	}
}
