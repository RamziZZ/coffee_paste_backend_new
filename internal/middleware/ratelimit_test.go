package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func hit(h http.Handler, xff string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	if xff != "" {
		r.Header.Set("X-Forwarded-For", xff)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func TestRateLimitBlocksAfterMax(t *testing.T) {
	h := RateLimit(3, time.Minute)(okHandler())
	for i := 1; i <= 3; i++ {
		if w := hit(h, "1.2.3.4"); w.Code != http.StatusOK {
			t.Fatalf("request %d: status %d, want 200", i, w.Code)
		}
	}
	w := hit(h, "1.2.3.4")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d, want 429", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("header Retry-After wajib ada")
	}
	if w := hit(h, "5.6.7.8"); w.Code != http.StatusOK {
		t.Errorf("IP lain tidak boleh ikut terblokir, status %d", w.Code)
	}
}

func TestRateLimitUsesRightmostForwardedIP(t *testing.T) {
	h := RateLimit(2, time.Minute)(okHandler())
	hit(h, "9.9.9.9, 2.2.2.2")
	hit(h, "8.8.8.8, 2.2.2.2")
	// Entri paling kiri bisa dipalsukan klien; yang dihitung adalah entri paling kanan.
	if w := hit(h, "7.7.7.7, 2.2.2.2"); w.Code != http.StatusTooManyRequests {
		t.Errorf("status %d, want 429", w.Code)
	}
}

func TestRateLimitWindowResets(t *testing.T) {
	h := RateLimit(1, 50*time.Millisecond)(okHandler())
	hit(h, "1.1.1.1")
	if w := hit(h, "1.1.1.1"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d, want 429", w.Code)
	}
	time.Sleep(80 * time.Millisecond)
	if w := hit(h, "1.1.1.1"); w.Code != http.StatusOK {
		t.Errorf("setelah jendela berakhir status %d, want 200", w.Code)
	}
}
