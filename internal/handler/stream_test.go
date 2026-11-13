package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"coffeepaste/internal/realtime"
)

// syncRecorder = ResponseRecorder yang aman dibaca dari goroutine lain saat handler masih menulis.
type syncRecorder struct {
	mu sync.Mutex
	*httptest.ResponseRecorder
}

func (s *syncRecorder) Write(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ResponseRecorder.Write(b)
}

func (s *syncRecorder) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ResponseRecorder.Flush()
}

func (s *syncRecorder) body() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Body.String()
}

func waitForBody(t *testing.T, rec *syncRecorder, want string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(rec.body(), want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("body tidak memuat %q, isi: %q", want, rec.body())
}

func TestStreamSendsEventWhenOrdersChange(t *testing.T) {
	hub := realtime.NewHub()
	h := (&MobileHandler{}).WithHub(hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := withUser(httptest.NewRequest("GET", "/api/orders/stream", nil).WithContext(ctx))

	rec := &syncRecorder{ResponseRecorder: httptest.NewRecorder()}
	done := make(chan struct{})
	go func() {
		h.Stream(rec, req)
		close(done)
	}()

	// Pesan pembuka ditulis setelah langganan terdaftar, jadi aman memanggil Notify sesudahnya.
	waitForBody(t, rec, "terhubung")
	hub.Notify(1) // withUser memakai customer id 1
	waitForBody(t, rec, "event: orders")

	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q", ct)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream tidak berhenti saat klien putus")
	}
}

func TestStreamWithoutHubIsNotFound(t *testing.T) {
	h := &MobileHandler{}
	w := httptest.NewRecorder()
	h.Stream(w, withUser(httptest.NewRequest("GET", "/api/orders/stream", nil)))
	if w.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404", w.Code)
	}
}

func TestKasirStreamSendsEventToAllStaff(t *testing.T) {
	hub := realtime.NewHub()
	h := NewKasirOrderHandler(nil).WithHub(hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := httptest.NewRequest("GET", "/api/kasir/pesanan/stream", nil).WithContext(ctx)

	rec := &syncRecorder{ResponseRecorder: httptest.NewRecorder()}
	done := make(chan struct{})
	go func() {
		h.Stream(rec, req)
		close(done)
	}()

	waitForBody(t, rec, "terhubung")
	hub.NotifyStaff()
	waitForBody(t, rec, "event: orders")

	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q", ct)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream tidak berhenti saat klien putus")
	}
}

func TestKasirStreamWithoutHubIsNotFound(t *testing.T) {
	h := NewKasirOrderHandler(nil)
	w := httptest.NewRecorder()
	h.Stream(w, httptest.NewRequest("GET", "/api/kasir/pesanan/stream", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("status %d, want 404", w.Code)
	}
}
