package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerKeepsFlusher(t *testing.T) {
	var flushable bool
	h := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, flushable = w.(http.Flusher)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))

	if !flushable {
		t.Fatal("ResponseWriter dari Logger harus mengimplementasikan http.Flusher")
	}
	if !rec.Flushed {
		t.Fatal("Flush tidak diteruskan ke ResponseWriter asli")
	}
}
