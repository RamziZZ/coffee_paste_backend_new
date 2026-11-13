package mailer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

type Mailer interface {
	Send(to, subject, body string) error
}

// LogMailer hanya mencetak email ke log. Dipakai saat SMTP belum dikonfigurasi (development).
type LogMailer struct{}

func (LogMailer) Send(to, subject, body string) error {
	log.Printf("[EMAIL - belum ada SMTP] kepada=%s subjek=%q\n%s", to, subject, body)
	return nil
}

type SMTPMailer struct {
	Host, Port, User, Pass, From string
}

// Send mengirim email teks biasa lewat SMTP.
//   - Port 465: TLS langsung (implicit TLS).
//   - Port lain (mis. 587): STARTTLS bila server mendukung.
//
// Subjek boleh berisi huruf non-ASCII (di-encode otomatis). Alamat tujuan harus sudah
// divalidasi pemanggil.
func (m *SMTPMailer) Send(to, subject, body string) error {
	if strings.ContainsAny(to, "\r\n") || strings.ContainsAny(subject, "\r\n") {
		return errors.New("header email tidak valid")
	}
	from, err := mail.ParseAddress(m.From)
	if err != nil {
		return fmt.Errorf("SMTP_FROM tidak valid: %w", err)
	}

	body = strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n")
	msg := "From: " + m.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n" +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	addr := net.JoinHostPort(m.Host, m.Port)
	dialer := &net.Dialer{Timeout: 15 * time.Second}

	var conn net.Conn
	if m.Port == "465" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: m.Host})
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	c, err := smtp.NewClient(conn, m.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	if m.Port != "465" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: m.Host}); err != nil {
				return err
			}
		}
	}
	if m.User != "" {
		if err := c.Auth(smtp.PlainAuth("", m.User, m.Pass, m.Host)); err != nil {
			return err
		}
	}
	if err := c.Mail(from.Address); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// New memilih SMTPMailer bila host diisi, selain itu LogMailer.
func New(host, port, user, pass, from string) Mailer {
	if host == "" {
		return LogMailer{}
	}
	return &SMTPMailer{Host: host, Port: port, User: user, Pass: pass, From: from}
}
