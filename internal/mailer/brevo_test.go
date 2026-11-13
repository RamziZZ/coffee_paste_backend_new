package mailer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBrevoSendPostsExpectedRequest(t *testing.T) {
	var gotKey, gotType, gotMethod string
	var got brevoPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("api-key")
		gotType = r.Header.Get("Content-Type")
		gotMethod = r.Method
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("body bukan JSON: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"messageId":"<abc@brevo>"}`))
	}))
	defer srv.Close()

	m := &BrevoMailer{APIKey: "kunci-rahasia", From: "Coffee Paste <no-reply@example.com>", URL: srv.URL}
	if err := m.Send("budi@example.com", "Reset password — Café", "Halo\nIni tes"); err != nil {
		t.Fatalf("Send gagal: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, ingin POST", gotMethod)
	}
	if gotKey != "kunci-rahasia" {
		t.Errorf("header api-key = %q", gotKey)
	}
	if gotType != "application/json" {
		t.Errorf("Content-Type = %q", gotType)
	}
	if got.Sender.Email != "no-reply@example.com" || got.Sender.Name != "Coffee Paste" {
		t.Errorf("sender = %+v", got.Sender)
	}
	if len(got.To) != 1 || got.To[0].Email != "budi@example.com" {
		t.Errorf("to = %+v", got.To)
	}
	if got.Subject != "Reset password — Café" || got.TextContent != "Halo\nIni tes" {
		t.Errorf("subject/body tidak sesuai: %q / %q", got.Subject, got.TextContent)
	}
}

func TestBrevoSendReportsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Key not found"}`))
	}))
	defer srv.Close()

	m := &BrevoMailer{APIKey: "salah", From: "no-reply@example.com", URL: srv.URL}
	err := m.Send("budi@example.com", "Halo", "isi")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("status 401 harus menjadi error, dapat: %v", err)
	}
	if strings.Contains(err.Error(), "salah") {
		t.Error("error tidak boleh memuat API key")
	}
}

func TestBrevoSendRejectsBadInput(t *testing.T) {
	m := &BrevoMailer{APIKey: "k", From: "no-reply@example.com", URL: "http://127.0.0.1:1"}
	if err := m.Send("a@b.com\r\nBcc: x@y.com", "Halo", "isi"); err == nil {
		t.Error("CRLF pada alamat tujuan harus ditolak")
	}
	if err := m.Send("a@b.com", "Halo\r\nBcc: x@y.com", "isi"); err == nil {
		t.Error("CRLF pada subjek harus ditolak")
	}
	if err := m.Send("bukan-alamat", "Halo", "isi"); err == nil {
		t.Error("alamat tujuan tidak valid harus ditolak")
	}
	bad := &BrevoMailer{APIKey: "k", From: "bukan-alamat", URL: "http://127.0.0.1:1"}
	if err := bad.Send("a@b.com", "Halo", "isi"); err == nil || !strings.Contains(err.Error(), "SMTP_FROM") {
		t.Errorf("SMTP_FROM tidak valid harus memberi error yang jelas, dapat: %v", err)
	}
}

func TestChoosePrefersBrevoThenSMTPThenLog(t *testing.T) {
	if _, ok := Choose("kunci", "smtp.example.com", "587", "u", "p", "a@b.com").(*BrevoMailer); !ok {
		t.Error("API key terisi harus memakai BrevoMailer")
	}
	if _, ok := Choose("  ", "smtp.example.com", "587", "u", "p", "a@b.com").(*SMTPMailer); !ok {
		t.Error("API key kosong + host terisi harus memakai SMTPMailer")
	}
	if _, ok := Choose("", "", "587", "", "", "").(LogMailer); !ok {
		t.Error("semuanya kosong harus memakai LogMailer")
	}
}
