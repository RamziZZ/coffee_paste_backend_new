package handler

import (
	"net/http"
	"testing"
)

func TestOTPRequestRejectsBadEmail(t *testing.T) {
	h := NewPasswordOTPHandler(nil, nil)
	for _, body := range []string{`{"email":""}`, `{"email":"bukan-email"}`, `{"email":"a@b"}`, `bukan json`} {
		if w := postJSON(h.Request, body); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", body, w.Code)
		}
	}
}

func TestOTPVerifyValidation(t *testing.T) {
	h := NewPasswordOTPHandler(nil, nil)
	for _, body := range []string{
		`{"email":"bukan-email","code":"123456"}`,
		`{"email":"a@b.com","code":""}`,
		`{"email":"a@b.com","code":"12345"}`,   // kurang dari 6 digit
		`{"email":"a@b.com","code":"1234567"}`, // lebih dari 6 digit
		`{"email":"a@b.com","code":"12a456"}`,  // bukan angka
	} {
		if w := postJSON(h.Verify, body); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", body, w.Code)
		}
	}
}

func TestOTPResetValidation(t *testing.T) {
	h := NewPasswordOTPHandler(nil, nil)
	for _, body := range []string{
		`{"email":"bukan-email","code":"123456","newPassword":"Password123"}`,
		`{"email":"a@b.com","code":"12345","newPassword":"Password123"}`,
		`{"email":"a@b.com","code":"123456","newPassword":"lemah"}`,
		`{"email":"a@b.com","code":"123456","newPassword":""}`,
	} {
		if w := postJSON(h.Reset, body); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", body, w.Code)
		}
	}
}

func TestOTPHashBoundToUser(t *testing.T) {
	if otpHash(1, "123456") != otpHash(1, "123456") {
		t.Error("hash harus deterministik")
	}
	if otpHash(1, "123456") == otpHash(2, "123456") {
		t.Error("kode yang sama untuk user berbeda harus menghasilkan hash berbeda")
	}
	if otpHash(1, "123456") == otpHash(1, "654321") {
		t.Error("kode berbeda harus menghasilkan hash berbeda")
	}
}

func TestValidOTPCode(t *testing.T) {
	cases := map[string]bool{
		"123456": true, "000000": true,
		"": false, "12345": false, "1234567": false, "12345a": false, " 12345": false,
	}
	for c, want := range cases {
		if got := validOTPCode(c); got != want {
			t.Errorf("validOTPCode(%q)=%v, want %v", c, got, want)
		}
	}
}
