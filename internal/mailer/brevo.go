package mailer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"
)

const brevoEndpoint = "https://api.brevo.com/v3/smtp/email"

// BrevoMailer mengirim email lewat HTTPS API Brevo (port 443), bukan SMTP.
// Cocok untuk hosting yang memblokir port SMTP (mis. Railway paket Hobby).
// Alamat pengirim diambil dari SMTP_FROM dan harus sudah diverifikasi di akun Brevo.
type BrevoMailer struct {
	APIKey string
	From   string       // "Nama <alamat@email.com>" atau "alamat@email.com"
	URL    string       // kosong = endpoint resmi Brevo (diisi hanya oleh tes)
	Client *http.Client // kosong = http.Client dengan timeout 15 detik
}

type brevoAddress struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type brevoPayload struct {
	Sender      brevoAddress   `json:"sender"`
	To          []brevoAddress `json:"to"`
	Subject     string         `json:"subject"`
	TextContent string         `json:"textContent"`
}

// Send mengirim email teks biasa. Subjek boleh berisi huruf non-ASCII.
func (m *BrevoMailer) Send(to, subject, body string) error {
	if strings.ContainsAny(to, "\r\n") || strings.ContainsAny(subject, "\r\n") {
		return errors.New("header email tidak valid")
	}
	from, err := mail.ParseAddress(m.From)
	if err != nil {
		return fmt.Errorf("SMTP_FROM tidak valid: %w", err)
	}
	rcpt, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("alamat tujuan tidak valid: %w", err)
	}

	payload, err := json.Marshal(brevoPayload{
		Sender:      brevoAddress{Name: from.Name, Email: from.Address},
		To:          []brevoAddress{{Email: rcpt.Address}},
		Subject:     subject,
		TextContent: body,
	})
	if err != nil {
		return err
	}

	url := m.URL
	if url == "" {
		url = brevoEndpoint
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("api-key", m.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := m.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("Brevo menolak email (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}
	io.Copy(io.Discard, resp.Body)
	return nil
}

// Choose memilih pengirim email: Brevo bila brevoAPIKey diisi, selain itu SMTP bila
// host diisi, selain itu LogMailer (email hanya dicetak ke log).
func Choose(brevoAPIKey, host, port, user, pass, from string) Mailer {
	if key := strings.TrimSpace(brevoAPIKey); key != "" {
		return &BrevoMailer{APIKey: key, From: from}
	}
	return New(host, port, user, pass, from)
}
