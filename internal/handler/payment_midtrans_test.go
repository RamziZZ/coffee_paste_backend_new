package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coffeepaste/internal/midtrans"
)

const testServerKey = "SB-Mid-server-test"

func midtransBody(status string, key string) string {
	sig := midtrans.Signature("CP00001-1", "200", "55500.00", key)
	return `{"order_id":"CP00001-1","status_code":"200","gross_amount":"55500.00","transaction_status":"` +
		status + `","signature_key":"` + sig + `"}`
}

func TestMidtransNotificationGuards(t *testing.T) {
	call := func(serverKey, body string) int {
		h := NewMidtransHandler(nil, serverKey)
		w := httptest.NewRecorder()
		h.Notification(w, httptest.NewRequest("POST", "/", stringsReader(body)))
		return w.Code
	}

	// Ditolak/diabaikan sebelum menyentuh database, jadi store boleh nil.
	if got := call("", midtransBody("settlement", testServerKey)); got != http.StatusNotFound {
		t.Errorf("tanpa server key: %d, want 404", got)
	}
	if got := call(testServerKey, midtransBody("settlement", "kunci-lain")); got != http.StatusUnauthorized {
		t.Errorf("signature palsu: %d, want 401", got)
	}
	if got := call(testServerKey, `{"order_id":"CP00001-1"}`); got != http.StatusUnauthorized {
		t.Errorf("tanpa signature: %d, want 401", got)
	}
	if got := call(testServerKey, `bukan json`); got != http.StatusUnprocessableEntity {
		t.Errorf("bukan JSON: %d, want 422", got)
	}
	if got := call(testServerKey, midtransBody("pending", testServerKey)); got != http.StatusOK {
		t.Errorf("pending: %d, want 200", got)
	}
}

func TestParseGrossAmount(t *testing.T) {
	ok := map[string]int64{"55500.00": 55500, "55500": 55500, " 1000.0 ": 1000}
	for in, want := range ok {
		if got, valid := parseGrossAmount(in); !valid || got != want {
			t.Errorf("parseGrossAmount(%q) = %d,%v want %d", in, got, valid, want)
		}
	}
	for _, in := range []string{"", "abc", "0", "-5", "55500.50"} {
		if _, valid := parseGrossAmount(in); valid {
			t.Errorf("parseGrossAmount(%q) seharusnya tidak valid", in)
		}
	}
}
