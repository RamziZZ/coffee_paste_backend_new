package push

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func newTestSender(t *testing.T, srv *httptest.Server) *Sender {
	t.Helper()
	raw, _ := json.Marshal(map[string]string{
		"project_id":   "demo",
		"client_email": "svc@demo.iam.gserviceaccount.com",
		"private_key":  testKeyPEM(t),
		"token_uri":    srv.URL + "/token",
	})
	s, err := ParseCredentials(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	s.apiBase = srv.URL
	return s
}

func TestSendBuildsMessage(t *testing.T) {
	var got map[string]map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
		case "/v1/projects/demo/messages:send":
			if r.Header.Get("Authorization") != "Bearer tok" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewDecoder(r.Body).Decode(&got)
			_, _ = w.Write([]byte(`{"name":"ok"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	s := newTestSender(t, srv)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Send(ctx, "device-1", Message{Title: "Halo", Body: "Isi", Data: map[string]string{"orderId": "#CP00001"}})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	m := got["message"]
	if m["token"] != "device-1" {
		t.Fatalf("token = %v", m["token"])
	}
	if n := m["notification"].(map[string]any); n["title"] != "Halo" || n["body"] != "Isi" {
		t.Fatalf("notification = %v", n)
	}
	if a := m["android"].(map[string]any); a["priority"] != "HIGH" {
		t.Fatalf("android = %v", a)
	}
}

func TestSendUnregistered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "tok", "expires_in": 3600})
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"status":"NOT_FOUND","details":[{"errorCode":"UNREGISTERED"}]}}`))
	}))
	defer srv.Close()

	s := newTestSender(t, srv)
	err := s.Send(context.Background(), "mati", Message{Title: "a", Body: "b"})
	if err != ErrUnregistered {
		t.Fatalf("err = %v, want ErrUnregistered", err)
	}
}

func TestParseCredentialsInvalid(t *testing.T) {
	for _, raw := range []string{`{}`, `{"project_id":"x"}`, `bukan-json-atau-base64!`} {
		if _, err := ParseCredentials(raw); err == nil {
			t.Errorf("ParseCredentials(%q) harus error", raw)
		}
	}
}
