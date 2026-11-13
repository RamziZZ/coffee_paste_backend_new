package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"coffeepaste/internal/qris"
	"coffeepaste/internal/store"
)

// testStaticQRIS merakit QRIS statis contoh lengkap dengan CRC.
func testStaticQRIS() string {
	body := "000201" + "010211" +
		"2644" + "0014ID.CO.QRIS.WWW" + "0215ID1234567890123" + "0303UMI" +
		"52045812" + "5303360" + "5802ID" + "5911COFFEEPASTE" + "6006JEMBER" + "6304"
	return fmt.Sprintf("%s%04X", body, qris.CRC16(body))
}

func TestMobileOrderQRISPayload(t *testing.T) {
	static := testStaticQRIS()
	if !qris.Valid(static) {
		t.Fatal("fixture QRIS statis tidak valid")
	}
	h := NewMobileHandler(nil, nil, nil, nil).WithQRIS(static)

	base := store.Order{
		ID: "#CP00001", Status: "menunggu", PaymentMethod: "QRIS",
		Items: []store.OrderItem{{Price: 50000, Qty: 1}}, // subtotal 50000 + PPN 5500 = 55500
	}

	m := h.mobileOrder(base)
	if m.Total != 55500 {
		t.Fatalf("total = %d, want 55500", m.Total)
	}
	if !qris.Valid(m.QRISPayload) || !strings.Contains(m.QRISPayload, "540555500") {
		t.Errorf("payload QRIS salah: %q", m.QRISPayload)
	}

	noPayload := map[string]store.Order{
		"tunai":       {Status: "menunggu", PaymentMethod: "Tunai (Cash)", Items: base.Items},
		"sudah lunas": {Status: "menunggu", PaymentMethod: "QRIS", Paid: true, Items: base.Items},
		"dibatalkan":  {Status: "dibatalkan", PaymentMethod: "QRIS", Items: base.Items},
	}
	for name, o := range noPayload {
		if got := h.mobileOrder(o).QRISPayload; got != "" {
			t.Errorf("%s: payload harus kosong, dapat %q", name, got)
		}
	}

	// Tanpa QRIS merchant terkonfigurasi: tidak ada QR, pesanan tetap terkirim.
	if got := NewMobileHandler(nil, nil, nil, nil).mobileOrder(base).QRISPayload; got != "" {
		t.Errorf("tanpa konfigurasi, payload harus kosong, dapat %q", got)
	}
}

func TestPaymentCallbackGuards(t *testing.T) {
	call := func(secret, header, body string) int {
		h := NewPaymentCallbackHandler(nil, secret) // semua kasus ditolak sebelum menyentuh store
		r := httptest.NewRequest("POST", "/", stringsReader(body))
		if header != "" {
			r.Header.Set("X-Callback-Secret", header)
		}
		w := httptest.NewRecorder()
		h.QRIS(w, r)
		return w.Code
	}

	cases := []struct {
		name           string
		secret, header string
		body           string
		want           int
	}{
		{"nonaktif tanpa secret", "", "apa-saja", `{}`, http.StatusNotFound},
		{"secret salah", "rahasia", "salah", `{"orderId":"#CP00001","amount":1000,"status":"paid"}`, http.StatusUnauthorized},
		{"tanpa header", "rahasia", "", `{"orderId":"#CP00001","amount":1000,"status":"paid"}`, http.StatusUnauthorized},
		{"status bukan paid diabaikan", "rahasia", "rahasia", `{"orderId":"#CP00001","amount":1000,"status":"pending"}`, http.StatusOK},
		{"orderId salah", "rahasia", "rahasia", `{"orderId":"abc","amount":1000,"status":"paid"}`, http.StatusUnprocessableEntity},
		{"nominal 0", "rahasia", "rahasia", `{"orderId":"#CP00001","amount":0,"status":"paid"}`, http.StatusUnprocessableEntity},
	}
	for _, c := range cases {
		if got := call(c.secret, c.header, c.body); got != c.want {
			t.Errorf("%s: status %d, want %d", c.name, got, c.want)
		}
	}
}

// Semua kasus ditolak sebelum menyentuh database, jadi users boleh nil.
func TestCustomerEmailChangeValidation(t *testing.T) {
	h := NewCustomerProfileHandler(nil, nil, 0)

	for _, b := range []string{
		`{"email":"bukan-email","currentPassword":"x"}`,
		`{"email":"","currentPassword":"x"}`,
	} {
		if w := postJSON(h.RequestEmailChange, b); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", b, w.Code)
		}
	}

	// Email valid tetapi pengirim email belum dikonfigurasi.
	if w := postJSON(h.RequestEmailChange, `{"email":"baru@mail.com","currentPassword":"x"}`); w.Code != http.StatusServiceUnavailable {
		t.Errorf("tanpa mailer: status %d, want 503", w.Code)
	}

	for _, b := range []string{
		`{"email":"baru@mail.com","code":"12345"}`,
		`{"email":"baru@mail.com","code":"12a456"}`,
		`{"email":"baru@mail.com","code":""}`,
	} {
		if w := postJSON(h.VerifyEmailChange, b); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", b, w.Code)
		}
	}
}
