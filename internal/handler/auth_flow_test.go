package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func postJSON(fn http.HandlerFunc, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	fn(w, withUser(httptest.NewRequest("POST", "/", stringsReader(body))))
	return w
}

func messageOf(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var m struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("respons bukan JSON: %v", err)
	}
	return m.Message
}

func TestValidPassword(t *testing.T) {
	cases := map[string]bool{
		"Abcdefg1":                     true,
		"abcdefg1":                     false, // tanpa huruf besar
		"ABCDEFG1":                     false, // tanpa huruf kecil
		"Abcdefgh":                     false, // tanpa angka
		"Ab1":                          false, // terlalu pendek
		"Aa1" + strings.Repeat("x", 70): false, // lebih dari 72 byte
	}
	for p, want := range cases {
		if got := validPassword(p); got != want {
			t.Errorf("validPassword(%q)=%v, want %v", p, got, want)
		}
	}
}

func TestValidEmail(t *testing.T) {
	cases := map[string]bool{
		"a@b.com":            true,
		"":                   false,
		"a@b":                false, // tanpa titik di domain
		"Nama <a@b.com>":     false, // hanya alamat polos
		"a@b.com\r\nBcc: x":  false,
		strings.Repeat("a", 150) + "@b.com": false,
	}
	for e, want := range cases {
		if got := validEmail(e); got != want {
			t.Errorf("validEmail(%q)=%v, want %v", e, got, want)
		}
	}
}

func TestForgotPasswordRejectsBadEmail(t *testing.T) {
	h := NewPasswordResetHandler(nil, nil, "")
	for _, body := range []string{`{"email":""}`, `{"email":"bukan-email"}`, `{"email":"a@b"}`, `bukan json`} {
		if w := postJSON(h.Forgot, body); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", body, w.Code)
		}
	}
}

func TestResetPasswordValidation(t *testing.T) {
	h := NewPasswordResetHandler(nil, nil, "")
	for _, body := range []string{
		`{"token":"","newPassword":"Password123"}`,
		`{"token":"   ","newPassword":"Password123"}`,
		`{"token":"abc","newPassword":"lemah"}`,
	} {
		if w := postJSON(h.Reset, body); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", body, w.Code)
		}
	}
}

func TestAdminProfileRequiresCurrentPassword(t *testing.T) {
	h := NewAdminProfileHandler(nil, nil, "")
	for name, w := range map[string]*httptest.ResponseRecorder{
		"ganti email":    postJSON(h.RequestEmailChange, `{"email":"baru@example.com"}`),
		"ganti password": postJSON(h.ChangePassword, `{"newPassword":"Password123"}`),
	} {
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", name, w.Code)
			continue
		}
		if msg := messageOf(t, w); !strings.Contains(msg, "Password saat ini") {
			t.Errorf("%s: pesan %q harus menyebut password saat ini", name, msg)
		}
	}
}

func TestAdminEmailChangeRejectsBadEmailBeforePasswordCheck(t *testing.T) {
	h := NewAdminProfileHandler(nil, nil, "")
	if w := postJSON(h.RequestEmailChange, `{"email":"salah","currentPassword":"x"}`); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status %d, want 422", w.Code)
	}
}
