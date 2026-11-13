package push

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestNextMessages(t *testing.T) {
	s := NewService(nil, nil)

	// Pesanan baru dari customer: belum ada yang perlu diberitahukan.
	if _, ok := s.next(Order{Code: "#CP00001", Status: "menunggu"}); ok {
		t.Fatal("pesanan baru berstatus menunggu tidak boleh memicu push")
	}
	// Hanya uang diterima/data lain berubah: status sama, tidak ada push.
	if _, ok := s.next(Order{Code: "#CP00001", Status: "menunggu"}); ok {
		t.Fatal("perubahan tanpa status/pembayaran baru tidak boleh memicu push")
	}

	m, ok := s.next(Order{Code: "#CP00001", Status: "proses"})
	if !ok || m.Title != "Pesanan sedang diproses" || m.Data["orderId"] != "#CP00001" || m.Data["status"] != "proses" {
		t.Fatalf("push proses salah: %+v ok=%v", m, ok)
	}
	if _, ok := s.next(Order{Code: "#CP00001", Status: "proses"}); ok {
		t.Fatal("status yang sama tidak boleh dikirim dua kali")
	}

	m, ok = s.next(Order{Code: "#CP00001", Status: "siap"})
	if !ok || m.Title != "Pesanan siap diambil" {
		t.Fatalf("push siap salah: %+v ok=%v", m, ok)
	}

	m, ok = s.next(Order{Code: "#CP00001", Status: "dibatalkan", CancelReason: "Stok habis"})
	if !ok || !strings.Contains(m.Body, "Stok habis") {
		t.Fatalf("push dibatalkan harus memuat alasan: %+v ok=%v", m, ok)
	}
}

func TestNextPaymentOnly(t *testing.T) {
	s := NewService(nil, nil)
	s.next(Order{Code: "#CP00002", Status: "menunggu"})

	m, ok := s.next(Order{Code: "#CP00002", Status: "menunggu", Paid: true})
	if !ok || m.Title != "Pembayaran diterima" {
		t.Fatalf("pelunasan harus memicu push pembayaran: %+v ok=%v", m, ok)
	}
	if _, ok := s.next(Order{Code: "#CP00002", Status: "menunggu", Paid: true}); ok {
		t.Fatal("pelunasan yang sama tidak boleh dikirim dua kali")
	}
}

func TestNextFirstSightAfterRestart(t *testing.T) {
	s := NewService(nil, nil)
	if _, ok := s.next(Order{Code: "#CP00003", Status: "siap"}); !ok {
		t.Fatal("status siap yang pertama kali terlihat harus tetap dikirim")
	}
}

type fakeTokens struct {
	mu      sync.Mutex
	tokens  []string
	deleted []string
}

func (f *fakeTokens) DeviceTokens(context.Context, int64) ([]string, error) { return f.tokens, nil }
func (f *fakeTokens) DeleteDeviceToken(_ context.Context, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, token)
	return nil
}

type fakeSender struct {
	sent []string
	fail map[string]error
}

func (f *fakeSender) Send(_ context.Context, token string, _ Message) error {
	f.sent = append(f.sent, token)
	return f.fail[token]
}

func TestDeliverRemovesUnregisteredTokens(t *testing.T) {
	tokens := &fakeTokens{tokens: []string{"ok", "mati"}}
	snd := &fakeSender{fail: map[string]error{"mati": ErrUnregistered}}
	s := NewService(tokens, snd)

	s.deliver(1, Message{Title: "t", Body: "b"})

	if len(snd.sent) != 2 {
		t.Fatalf("seharusnya mengirim ke 2 token, dapat %v", snd.sent)
	}
	if len(tokens.deleted) != 1 || tokens.deleted[0] != "mati" {
		t.Fatalf("token tidak berlaku harus dihapus, dapat %v", tokens.deleted)
	}
}

func TestNilServiceIsSafe(t *testing.T) {
	var s *Service
	s.OrderChanged(1, Order{Code: "#CP1", Status: "proses"})
}

func testCredentials(t *testing.T, tokenURL string) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	raw, err := json.Marshal(map[string]string{
		"project_id":   "coffee-test",
		"client_email": "sa@coffee-test.iam.gserviceaccount.com",
		"private_key":  string(pemKey),
		"token_uri":    tokenURL,
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestClientSend(t *testing.T) {
	var gotAuth, gotBody string
	var tokenCalls int

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		tokenCalls++
		_ = r.ParseForm()
		if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" || r.Form.Get("assertion") == "" {
			http.Error(w, "bad", http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(w, `{"access_token":"abc","expires_in":3600}`)
	})
	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		if strings.Contains(gotBody, `"mati"`) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `{"error":{"status":"NOT_FOUND","details":[{"errorCode":"UNREGISTERED"}]}}`)
			return
		}
		_, _ = io.WriteString(w, `{"name":"projects/coffee-test/messages/1"}`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := NewClient(testCredentials(t, srv.URL+"/token"))
	if err != nil {
		t.Fatal(err)
	}
	c.sendURL = srv.URL + "/send"

	msg := Message{Title: "Pesanan siap diambil", Body: "Silakan ambil", Data: map[string]string{"orderId": "#CP00001"}}
	if err := c.Send(context.Background(), "token-1", msg); err != nil {
		t.Fatalf("Send gagal: %v", err)
	}
	if gotAuth != "Bearer abc" {
		t.Fatalf("header Authorization salah: %q", gotAuth)
	}
	for _, want := range []string{`"token":"token-1"`, `"title":"Pesanan siap diambil"`, `"channel_id":"order_updates"`, `"priority":"HIGH"`, `"orderId":"#CP00001"`} {
		if !strings.Contains(gotBody, want) {
			t.Fatalf("payload tidak memuat %s: %s", want, gotBody)
		}
	}

	if err := c.Send(context.Background(), "mati", msg); !errors.Is(err, ErrUnregistered) {
		t.Fatalf("token mati harus ErrUnregistered, dapat %v", err)
	}
	if tokenCalls != 1 {
		t.Fatalf("access token harus di-cache, diminta %d kali", tokenCalls)
	}
}

func TestParseCredentialsRejectsIncomplete(t *testing.T) {
	if _, err := parseCredentials(`{"project_id":"x"}`); err == nil {
		t.Fatal("kredensial tidak lengkap harus ditolak")
	}
	if _, err := parseCredentials("bukan-json-dan-bukan-base64!!"); err == nil {
		t.Fatal("input acak harus ditolak")
	}
}
