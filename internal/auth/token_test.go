package auth

import (
	"testing"
	"time"
)

var testSecret = []byte("rahasia-uji-minimal-16-karakter")

func TestTokenRoundTripCarriesVersion(t *testing.T) {
	tok, err := NewToken(testSecret, time.Hour, 42, "admin", 3)
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseToken(testSecret, tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.Subject != "42" || c.Role != "admin" || c.Ver != 3 {
		t.Errorf("klaim tidak sesuai: sub=%q role=%q ver=%d", c.Subject, c.Role, c.Ver)
	}
}

func TestTokenRejectsWrongSecretAndExpired(t *testing.T) {
	tok, _ := NewToken(testSecret, time.Hour, 1, "kasir", 0)
	if _, err := ParseToken([]byte("kunci-lain-yang-berbeda-xx"), tok); err == nil {
		t.Error("token dengan secret berbeda harus ditolak")
	}
	expired, _ := NewToken(testSecret, -time.Minute, 1, "kasir", 0)
	if _, err := ParseToken(testSecret, expired); err == nil {
		t.Error("token kedaluwarsa harus ditolak")
	}
}
