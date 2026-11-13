package midtrans

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSignature(t *testing.T) {
	sig := Signature("CP00001-1", "200", "55500.00", "SB-Mid-server-abc")
	if len(sig) != 128 {
		t.Fatalf("panjang signature %d, want 128", len(sig))
	}
	if !VerifySignature("CP00001-1", "200", "55500.00", sig, "SB-Mid-server-abc") {
		t.Error("signature sah ditolak")
	}
	if !VerifySignature("CP00001-1", "200", "55500.00", strings.ToUpper(sig), "SB-Mid-server-abc") {
		t.Error("huruf besar harus tetap diterima")
	}
	bad := []struct{ order, code, amount, key string }{
		{"CP00002-1", "200", "55500.00", "SB-Mid-server-abc"},
		{"CP00001-1", "201", "55500.00", "SB-Mid-server-abc"},
		{"CP00001-1", "200", "1.00", "SB-Mid-server-abc"},
		{"CP00001-1", "200", "55500.00", "kunci-lain"},
	}
	for _, b := range bad {
		if VerifySignature(b.order, b.code, b.amount, sig, b.key) {
			t.Errorf("signature palsu diterima: %+v", b)
		}
	}
	if VerifySignature("a", "200", "1.00", "", "k") || VerifySignature("a", "200", "1.00", sig, "") {
		t.Error("signature atau kunci kosong harus ditolak")
	}
}

func TestChargeQRIS(t *testing.T) {
	var gotAuth, gotNotify string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/charge" || r.Method != http.MethodPost {
			t.Errorf("request tak terduga: %s %s", r.Method, r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		gotNotify = r.Header.Get("X-Override-Notification")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"status_code":"201","status_message":"QRIS transaction is created",
			"transaction_id":"tx-1","qr_string":"00020101021226","expiry_time":"2026-10-10 20:00:00",
			"actions":[{"name":"generate-qr-code","method":"GET","url":"https://example/qr.png"}]}`))
	}))
	defer srv.Close()

	c := New("SB-Mid-server-abc", false, "https://abc.ngrok.app/api/payments/midtrans/notification")
	c.baseURL = srv.URL

	ch, err := c.ChargeQRIS(context.Background(), "CP00001-1", 55500, 30)
	if err != nil {
		t.Fatal(err)
	}
	if ch.QRString != "00020101021226" || ch.TransactionID != "tx-1" || ch.QRURL != "https://example/qr.png" {
		t.Errorf("hasil salah: %+v", ch)
	}
	
	if gotAuth != "Basic U0ItTWlkLXNlcnZlci1hYmM6" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotNotify != "https://abc.ngrok.app/api/payments/midtrans/notification" {
		t.Errorf("X-Override-Notification = %q", gotNotify)
	}
	if gotBody["payment_type"] != "qris" {
		t.Errorf("payment_type = %v", gotBody["payment_type"])
	}
	td := gotBody["transaction_details"].(map[string]any)
	if td["order_id"] != "CP00001-1" || td["gross_amount"] != float64(55500) {
		t.Errorf("transaction_details = %v", td)
	}
}

func TestChargeQRISErrors(t *testing.T) {
	respond := func(body string) *Client {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(body))
		}))
		t.Cleanup(srv.Close)
		c := New("k", false, "")
		c.baseURL = srv.URL
		return c
	}

	if _, err := respond(`{"status_code":"401","status_message":"Access denied"}`).
		ChargeQRIS(context.Background(), "x", 1000, 30); err == nil || !strings.Contains(err.Error(), "Access denied") {
		t.Errorf("err = %v", err)
	}
	if _, err := respond(`{"status_code":"201","status_message":"ok"}`).
		ChargeQRIS(context.Background(), "x", 1000, 30); err == nil {
		t.Error("tanpa qr_string harus error")
	}
	if _, err := respond(`bukan json`).ChargeQRIS(context.Background(), "x", 1000, 30); err == nil {
		t.Error("respons bukan JSON harus error")
	}
	if _, err := New("k", false, "").ChargeQRIS(context.Background(), "x", 0, 30); err == nil {
		t.Error("nominal 0 harus error")
	}
}
