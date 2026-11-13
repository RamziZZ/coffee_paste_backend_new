package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"coffeepaste/internal/googleauth"
)

type fakeGoogle struct {
	claims *googleauth.Claims
	err    error
}

func (f fakeGoogle) Verify(context.Context, string) (*googleauth.Claims, error) {
	return f.claims, f.err
}

func newGoogleTestHandler(v googleauth.Verifier) *AuthHandler {
	h := NewAuthHandler(nil, []byte("rahasia-rahasia-rahasia"), time.Hour)
	if v != nil {
		h.WithGoogle(v)
	}
	return h
}

func TestGoogleLoginDisabledWithoutConfig(t *testing.T) {
	h := newGoogleTestHandler(nil)
	if w := postJSON(h.Google, `{"idToken":"x"}`); w.Code != http.StatusServiceUnavailable {
		t.Errorf("status %d, want 503", w.Code)
	}
}

func TestGoogleLoginRejectsBadRequests(t *testing.T) {
	h := newGoogleTestHandler(fakeGoogle{err: errors.New("jelek")})

	for _, body := range []string{`{}`, `{"idToken":"  "}`, `bukan json`} {
		if w := postJSON(h.Google, body); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", body, w.Code)
		}
	}
	if w := postJSON(h.Google, `{"idToken":"abc"}`); w.Code != http.StatusUnauthorized {
		t.Errorf("token ditolak verifier: status %d, want 401", w.Code)
	}
}

func TestGoogleLoginRequiresVerifiedEmail(t *testing.T) {
	h := newGoogleTestHandler(fakeGoogle{claims: &googleauth.Claims{
		Subject: "1", Email: "ayu@gmail.com", EmailVerified: false,
	}})
	if w := postJSON(h.Google, `{"idToken":"abc"}`); w.Code != http.StatusForbidden {
		t.Errorf("status %d, want 403", w.Code)
	}
}
