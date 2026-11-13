package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coffeepaste/internal/store"
)

func TestValidPhone(t *testing.T) {
	cases := map[string]bool{
		"081234567890":      true,
		"0812-3456-7890":    true,
		"+62 812 3456 7890": true,
		"12345":             false,
		"08123abc4567":      false,
		"":                  false,
	}
	for p, want := range cases {
		if got := validPhone(p); got != want {
			t.Errorf("validPhone(%q)=%v, want %v", p, got, want)
		}
	}
}

// Semua kasus berikut ditolak sebelum menyentuh database, jadi users boleh nil.
func TestRegisterValidation(t *testing.T) {
	h := &AuthHandler{}
	cases := []struct{ name, body string }{
		{"tanpa nama", `{"name":" ","email":"a@b.com","password":"Abcdefg1"}`},
		{"email salah", `{"name":"Ayu","email":"ayu","password":"Abcdefg1"}`},
		{"telepon salah", `{"name":"Ayu","email":"a@b.com","phone":"12","password":"Abcdefg1"}`},
		{"password lemah", `{"name":"Ayu","email":"a@b.com","password":"abc"}`},
	}
	for _, c := range cases {
		w := postJSON(h.Register, c.body)
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", c.name, w.Code)
		}
	}
}

func TestToMobileOrderTotals(t *testing.T) {
	o := store.Order{Items: []store.OrderItem{
		{Price: 22000, Qty: 2},
		{Price: 5000, Qty: 1},
	}}
	m := toMobileOrder(o)
	if m.Subtotal != 49000 || m.Tax != 5390 || m.Total != 54390 {
		t.Errorf("got subtotal=%d tax=%d total=%d", m.Subtotal, m.Tax, m.Total)
	}
}

// Semua kasus ditolak sebelum menyentuh database, jadi users boleh nil.
func TestCustomerProfileValidation(t *testing.T) {
	h := NewCustomerProfileHandler(nil, nil, 0)
	put := func(body string) int {
		w := httptest.NewRecorder()
		h.Update(w, withUser(httptest.NewRequest("PUT", "/", stringsReader(body))))
		return w.Code
	}
	bad := []string{
		`{"name":" ","phone":""}`,
		`{"name":"Ayu","gender":"Lainnya"}`,
		`{"name":"Ayu","birthDate":"12-07-2004"}`,
		`{"name":"Ayu","birthDate":"2999-01-01"}`,
		`{"name":"Ayu","phone":"12"}`,
	}
	for _, b := range bad {
		if code := put(b); code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", b, code)
		}
	}
}

func TestCustomerChangePasswordValidation(t *testing.T) {
	h := NewCustomerProfileHandler(nil, nil, 0)
	for _, b := range []string{
		`{"currentPassword":"","newPassword":"Abcdefg1"}`,
		`{"currentPassword":"x","newPassword":"abc"}`,
	} {
		if w := postJSON(h.ChangePassword, b); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", b, w.Code)
		}
	}
}
