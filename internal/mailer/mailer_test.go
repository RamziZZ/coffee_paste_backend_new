package mailer

import (
	"bufio"
	"net"
	"strings"
	"testing"
	"time"
)

func TestNewPicksMailer(t *testing.T) {
	if _, ok := New("", "587", "", "", "").(LogMailer); !ok {
		t.Error("host kosong harus memakai LogMailer")
	}
	if _, ok := New("smtp.example.com", "587", "u", "p", "no-reply@example.com").(*SMTPMailer); !ok {
		t.Error("host terisi harus memakai SMTPMailer")
	}
}

func TestSendRejectsHeaderInjectionAndBadFrom(t *testing.T) {
	m := &SMTPMailer{Host: "127.0.0.1", Port: "1", From: "no-reply@example.com"}
	if err := m.Send("a@b.com\r\nBcc: x@y.com", "Halo", "isi"); err == nil {
		t.Error("CRLF pada alamat tujuan harus ditolak")
	}
	if err := m.Send("a@b.com", "Halo\r\nBcc: x@y.com", "isi"); err == nil {
		t.Error("CRLF pada subjek harus ditolak")
	}
	bad := &SMTPMailer{Host: "127.0.0.1", Port: "1", From: "bukan-alamat"}
	if err := bad.Send("a@b.com", "Halo", "isi"); err == nil || !strings.Contains(err.Error(), "SMTP_FROM") {
		t.Errorf("SMTP_FROM tidak valid harus memberi error yang jelas, dapat: %v", err)
	}
}

// fakeSMTP menjalankan server SMTP minimal (tanpa TLS/AUTH) dan mengirim isi pesan ke channel.
func fakeSMTP(t *testing.T) (port string, got chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	got = make(chan string, 1)
	go func() {
		defer ln.Close()
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		r := bufio.NewReader(c)
		say := func(s string) { c.Write([]byte(s + "\r\n")) }
		var data strings.Builder
		say("220 fake ESMTP")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "DATA"):
				say("354 go ahead")
				for {
					l, err := r.ReadString('\n')
					if err != nil {
						return
					}
					if l == ".\r\n" {
						break
					}
					data.WriteString(l)
				}
				say("250 queued")
			case strings.HasPrefix(cmd, "QUIT"):
				say("221 bye")
				got <- data.String()
				return
			default:
				say("250 ok")
			}
		}
	}()
	_, port, _ = net.SplitHostPort(ln.Addr().String())
	return port, got
}

func TestSendDeliversEncodedMessage(t *testing.T) {
	port, got := fakeSMTP(t)
	m := &SMTPMailer{Host: "127.0.0.1", Port: port, From: "no-reply@example.com"}
	if err := m.Send("budi@example.com", "Reset password — Café", "Halo\nIni tes"); err != nil {
		t.Fatalf("Send gagal: %v", err)
	}
	select {
	case msg := <-got:
		for _, want := range []string{
			"To: budi@example.com\r\n",
			"Subject: =?utf-8?q?",
			"Content-Type: text/plain; charset=UTF-8\r\n",
			"Halo\r\nIni tes",
		} {
			if !strings.Contains(msg, want) {
				t.Errorf("pesan tidak memuat %q:\n%s", want, msg)
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server SMTP palsu tidak menerima pesan")
	}
}
