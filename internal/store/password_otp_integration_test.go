package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// Tes integrasi: butuh MySQL/MariaDB dengan tabel `users` (skema proyek).
// Jalankan dengan: OTP_TEST_DSN='user:pass@tcp(127.0.0.1:3306)/db?parseTime=true' go test ./internal/store -run OTP
func otpTestStore(t *testing.T) (*UserStore, int64) {
	t.Helper()
	dsn := os.Getenv("OTP_TEST_DSN")
	if dsn == "" {
		t.Skip("OTP_TEST_DSN belum diisi")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	s := &UserStore{DB: db}
	ctx := context.Background()
	if err := s.EnsurePasswordOTPSchema(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.EnsurePasswordOTPSchema(ctx); err != nil { // idempoten
		t.Fatal(err)
	}
	var uid int64
	if err := db.QueryRow(`SELECT id FROM users ORDER BY id LIMIT 1`).Scan(&uid); err != nil {
		t.Fatal(err)
	}
	db.Exec(`DELETE FROM password_reset_otps`)
	db.Exec(`UPDATE users SET password_hash='old', token_version=0 WHERE id=?`, uid)
	return s, uid
}

func TestOTPFlow(t *testing.T) {
	s, uid := otpTestStore(t)
	ctx := context.Background()

	if err := s.CreatePasswordOTP(ctx, uid, "hash-ok", 10); err != nil {
		t.Fatal(err)
	}
	if recent, _ := s.RecentOTPExists(ctx, uid, 60); !recent {
		t.Error("cooldown harus aktif setelah OTP dibuat")
	}
	// Verify tidak memakai OTP: boleh berkali-kali.
	for i := 0; i < 3; i++ {
		if err := s.CheckPasswordOTP(ctx, uid, "hash-ok", 5); err != nil {
			t.Fatalf("verify ke-%d: %v", i, err)
		}
	}
	if err := s.CheckPasswordOTP(ctx, uid, "salah", 5); !errors.Is(err, ErrOTPInvalid) {
		t.Fatalf("kode salah: %v", err)
	}
	if err := s.ResetPasswordWithOTP(ctx, uid, "hash-ok", "new-hash", 5); err != nil {
		t.Fatal(err)
	}
	var h string
	var tv int
	s.DB.QueryRow(`SELECT password_hash, token_version FROM users WHERE id=?`, uid).Scan(&h, &tv)
	if h != "new-hash" || tv != 1 {
		t.Errorf("password=%q token_version=%d, want new-hash/1", h, tv)
	}
	// Sekali pakai.
	if err := s.ResetPasswordWithOTP(ctx, uid, "hash-ok", "x", 5); !errors.Is(err, ErrNotFound) {
		t.Errorf("pakai ulang: %v, want ErrNotFound", err)
	}
}

func TestOTPAttemptsExhausted(t *testing.T) {
	s, uid := otpTestStore(t)
	ctx := context.Background()
	s.CreatePasswordOTP(ctx, uid, "hash-ok", 10)

	for i := 1; i < 5; i++ {
		if err := s.CheckPasswordOTP(ctx, uid, "salah", 5); !errors.Is(err, ErrOTPInvalid) {
			t.Fatalf("percobaan %d: %v", i, err)
		}
	}
	if err := s.CheckPasswordOTP(ctx, uid, "salah", 5); !errors.Is(err, ErrOTPExhausted) {
		t.Fatalf("percobaan terakhir: %v", err)
	}
	// Kode benar pun ditolak setelah batas habis.
	if err := s.CheckPasswordOTP(ctx, uid, "hash-ok", 5); !errors.Is(err, ErrNotFound) {
		t.Errorf("setelah habis: %v, want ErrNotFound", err)
	}
}

func TestOTPExpiredAndReplaced(t *testing.T) {
	s, uid := otpTestStore(t)
	ctx := context.Background()

	s.CreatePasswordOTP(ctx, uid, "lama", 10)
	s.DB.Exec(`UPDATE password_reset_otps SET expires_at = DATE_SUB(NOW(), INTERVAL 1 MINUTE)`)
	if err := s.CheckPasswordOTP(ctx, uid, "lama", 5); !errors.Is(err, ErrNotFound) {
		t.Errorf("kedaluwarsa: %v, want ErrNotFound", err)
	}

	s.CreatePasswordOTP(ctx, uid, "lama2", 10)
	s.CreatePasswordOTP(ctx, uid, "baru", 10) // minta ulang -> yang lama tidak berlaku
	if err := s.CheckPasswordOTP(ctx, uid, "lama2", 5); !errors.Is(err, ErrOTPInvalid) {
		t.Errorf("OTP lama: %v, want ErrOTPInvalid", err)
	}
	if err := s.CheckPasswordOTP(ctx, uid, "baru", 5); err != nil {
		t.Errorf("OTP baru: %v", err)
	}
}

func TestOTPWrongCodeDoesNotChangePassword(t *testing.T) {
	s, uid := otpTestStore(t)
	ctx := context.Background()
	s.CreatePasswordOTP(ctx, uid, "hash-ok", 10)
	if err := s.ResetPasswordWithOTP(ctx, uid, "salah", "hacked", 5); !errors.Is(err, ErrOTPInvalid) {
		t.Fatalf("reset kode salah: %v", err)
	}
	var h string
	s.DB.QueryRow(`SELECT password_hash FROM users WHERE id=?`, uid).Scan(&h)
	if h != "old" {
		t.Errorf("password berubah menjadi %q padahal kode salah", h)
	}
}
