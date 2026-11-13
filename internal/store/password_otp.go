package store

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
)

// OTP lupa password untuk aplikasi mobile. Disimpan di tabel sendiri (bukan `password_resets`)
// supaya tidak saling menghapus dengan tautan reset yang dipakai web.

var (
	// ErrOTPInvalid: kode salah, masih ada sisa percobaan.
	ErrOTPInvalid = errors.New("kode otp salah")
	// ErrOTPExhausted: kode salah dan batas percobaan habis; OTP dihapus, user harus minta kode baru.
	ErrOTPExhausted = errors.New("percobaan kode otp habis")
)

const passwordOTPSchema = `CREATE TABLE IF NOT EXISTS password_reset_otps (
  id bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id bigint(20) UNSIGNED NOT NULL,
  code_hash varchar(255) NOT NULL,
  attempts tinyint(3) UNSIGNED NOT NULL DEFAULT 0,
  expires_at datetime NOT NULL,
  used_at datetime DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (id),
  KEY idx_pro_user (user_id),
  CONSTRAINT fk_pro_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

// EnsurePasswordOTPSchema membuat tabel OTP bila belum ada (aman dipanggil berulang).
func (s *UserStore) EnsurePasswordOTPSchema(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, passwordOTPSchema)
	return err
}

// RecentOTPExists true bila user sudah meminta OTP dalam `seconds` detik terakhir
// (membatasi spam email).
func (s *UserStore) RecentOTPExists(ctx context.Context, userID int64, seconds int) (bool, error) {
	var n int
	err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM password_reset_otps
		  WHERE user_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL ? SECOND)`, userID, seconds).Scan(&n)
	return n > 0, err
}

// CreatePasswordOTP menyimpan hash OTP baru. OTP lama yang belum dipakai dihapus.
func (s *UserStore) CreatePasswordOTP(ctx context.Context, userID int64, codeHash string, ttlMinutes int) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM password_reset_otps WHERE user_id = ? AND used_at IS NULL`, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO password_reset_otps (user_id, code_hash, expires_at)
		 VALUES (?, ?, DATE_ADD(NOW(), INTERVAL ? MINUTE))`, userID, codeHash, ttlMinutes); err != nil {
		return err
	}
	return tx.Commit()
}

// DeletePasswordOTP membatalkan OTP yang belum dipakai (mis. saat email gagal terkirim).
func (s *UserStore) DeletePasswordOTP(ctx context.Context, userID int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM password_reset_otps WHERE user_id = ? AND used_at IS NULL`, userID)
	return err
}

// checkOTPTx mengunci OTP aktif milik user lalu membandingkannya dengan codeHash.
// Mengembalikan id baris bila cocok. Bila salah, jumlah percobaan dinaikkan (atau baris dihapus
// saat batas habis) dan pemanggil WAJIB commit supaya hitungan itu tersimpan.
func checkOTPTx(ctx context.Context, tx *sql.Tx, userID int64, codeHash string, maxAttempts int) (int64, error) {
	var id int64
	var stored string
	var attempts int
	err := tx.QueryRowContext(ctx,
		`SELECT id, code_hash, attempts
		   FROM password_reset_otps
		  WHERE user_id = ? AND used_at IS NULL AND expires_at > NOW()
		  ORDER BY id DESC LIMIT 1 FOR UPDATE`, userID).Scan(&id, &stored, &attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if subtle.ConstantTimeCompare([]byte(stored), []byte(codeHash)) == 1 {
		return id, nil
	}

	attempts++
	if attempts >= maxAttempts {
		if _, err := tx.ExecContext(ctx, `DELETE FROM password_reset_otps WHERE id = ?`, id); err != nil {
			return 0, err
		}
		return 0, ErrOTPExhausted
	}
	if _, err := tx.ExecContext(ctx, `UPDATE password_reset_otps SET attempts = ? WHERE id = ?`, attempts, id); err != nil {
		return 0, err
	}
	return 0, ErrOTPInvalid
}

// CheckPasswordOTP memeriksa OTP tanpa memakainya (langkah "verifikasi kode" di aplikasi).
// Error: ErrNotFound (tidak ada/kedaluwarsa), ErrOTPInvalid, ErrOTPExhausted.
func (s *UserStore) CheckPasswordOTP(ctx context.Context, userID int64, codeHash string, maxAttempts int) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = checkOTPTx(ctx, tx, userID, codeHash, maxAttempts)
	if err != nil && !errors.Is(err, ErrOTPInvalid) && !errors.Is(err, ErrOTPExhausted) {
		return err
	}
	if cerr := tx.Commit(); cerr != nil {
		return cerr
	}
	return err
}

// ResetPasswordWithOTP memverifikasi OTP lalu mengganti password. OTP hanya berlaku sekali;
// token_version dinaikkan sehingga semua sesi lama otomatis tidak berlaku.
func (s *UserStore) ResetPasswordWithOTP(ctx context.Context, userID int64, codeHash, newPasswordHash string, maxAttempts int) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	id, err := checkOTPTx(ctx, tx, userID, codeHash, maxAttempts)
	if err != nil {
		if errors.Is(err, ErrOTPInvalid) || errors.Is(err, ErrOTPExhausted) {
			if cerr := tx.Commit(); cerr != nil {
				return cerr
			}
		}
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET password_hash = ?, token_version = token_version + 1 WHERE id = ?`,
		newPasswordHash, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE password_reset_otps SET used_at = NOW() WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
