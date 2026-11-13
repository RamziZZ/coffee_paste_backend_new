package store

import (
	"context"
	"database/sql"
	"errors"
)

func (s *UserStore) UpdateName(ctx context.Context, id int64, name string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE users SET name = ? WHERE id = ?`, name, id)
	return err
}

func (s *UserStore) UpdatePassword(ctx context.Context, id int64, hash string) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE users SET password_hash = ?, token_version = token_version + 1 WHERE id = ?`, hash, id)
	return err
}

// PasswordHash mengambil hash password user (untuk memeriksa password lama).
func (s *UserStore) PasswordHash(ctx context.Context, id int64) (string, error) {
	var h string
	err := s.DB.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = ?`, id).Scan(&h)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return h, err
}

// EmailTaken true bila email dipakai user lain (termasuk akun yang sudah dihapus,
// karena kolom email bersifat unik di tabel users).
func (s *UserStore) EmailTaken(ctx context.Context, email string, exceptUserID int64) (bool, error) {
	var n int
	err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM users WHERE email = ? AND id <> ?`, email, exceptUserID).Scan(&n)
	return n > 0, err
}

// CreateEmailChange menyimpan permintaan ganti email. Permintaan lama yang belum dipakai dihapus.
func (s *UserStore) CreateEmailChange(ctx context.Context, userID int64, newEmail, tokenHash string, ttlMinutes int) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM email_change_requests WHERE user_id = ? AND used_at IS NULL`, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO email_change_requests (user_id, new_email, token_hash, expires_at)
		 VALUES (?, ?, ?, DATE_ADD(NOW(), INTERVAL ? MINUTE))`,
		userID, newEmail, tokenHash, ttlMinutes); err != nil {
		return err
	}
	return tx.Commit()
}

// CancelEmailChange membatalkan permintaan yang masih menunggu verifikasi.
func (s *UserStore) CancelEmailChange(ctx context.Context, userID int64) error {
	_, err := s.DB.ExecContext(ctx,
		`DELETE FROM email_change_requests WHERE user_id = ? AND used_at IS NULL`, userID)
	return err
}

// ConfirmEmailChange memverifikasi token dan mengganti email user. Mengembalikan email baru.
func (s *UserStore) ConfirmEmailChange(ctx context.Context, tokenHash string) (string, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var reqID, userID int64
	var newEmail string
	err = tx.QueryRowContext(ctx,
		`SELECT id, user_id, new_email
		   FROM email_change_requests
		  WHERE token_hash = ? AND used_at IS NULL AND expires_at > NOW()
		  LIMIT 1 FOR UPDATE`, tokenHash).Scan(&reqID, &userID, &newEmail)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE users SET email = ? WHERE id = ?`, newEmail, userID); err != nil {
		if isDuplicate(err) {
			return "", ErrDuplicateEmail
		}
		return "", err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE email_change_requests SET used_at = NOW() WHERE id = ?`, reqID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return newEmail, nil
}

// UpdateProfile menyimpan nama, telepon, dan alamat (dipakai profil kasir).
func (s *UserStore) UpdateProfile(ctx context.Context, id int64, name, phone, address string) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE users SET name = ?, phone = ?, address = ? WHERE id = ?`,
		name, nullIfEmpty(phone), nullIfEmpty(address), id)
	return err
}

// PhoneAddress mengambil telepon dan alamat user (string kosong bila belum diisi).
func (s *UserStore) PhoneAddress(ctx context.Context, id int64) (phone, address string, err error) {
	var p, a sql.NullString
	err = s.DB.QueryRowContext(ctx, `SELECT phone, address FROM users WHERE id = ?`, id).Scan(&p, &a)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return p.String, a.String, err
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// PendingEmail mengembalikan email baru yang masih menunggu verifikasi (nil bila tidak ada
// atau sudah kedaluwarsa). Dipakai agar status "menunggu verifikasi" tidak hilang saat halaman dimuat ulang.
func (s *UserStore) PendingEmail(ctx context.Context, userID int64) (*string, error) {
	var e string
	err := s.DB.QueryRowContext(ctx,
		`SELECT new_email FROM email_change_requests
		  WHERE user_id = ? AND used_at IS NULL AND expires_at > NOW()
		  ORDER BY id DESC LIMIT 1`, userID).Scan(&e)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// CustomerProfile = data profil yang bisa dilihat dan diubah customer dari aplikasi mobile.
type CustomerProfile struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
	Gender    *string `json:"gender"`
	BirthDate *string `json:"birthDate"` // YYYY-MM-DD
	// PendingEmail: email baru yang menunggu verifikasi kode (nil bila tidak ada).
	PendingEmail *string `json:"pendingEmail"`
}

func (s *UserStore) CustomerProfile(ctx context.Context, id int64) (*CustomerProfile, error) {
	var p CustomerProfile
	var phone, gender, birth sql.NullString
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, name, email, phone, gender, DATE_FORMAT(birth_date, '%Y-%m-%d')
		   FROM users WHERE id = ? AND deleted_at IS NULL`, id,
	).Scan(&p.ID, &p.Name, &p.Email, &phone, &gender, &birth)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Phone = phone.String
	p.Gender = nullStr(gender)
	p.BirthDate = nullStr(birth)
	return &p, nil
}

// UpdateCustomerProfile mengganti nama, telepon, jenis kelamin, dan tanggal lahir.
// gender / birthDate nil = dikosongkan. Email tidak bisa diubah dari sini.
func (s *UserStore) UpdateCustomerProfile(ctx context.Context, id int64, name, phone string, gender, birthDate *string) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE users SET name = ?, phone = ?, gender = ?, birth_date = ?
		  WHERE id = ? AND role = 'customer' AND deleted_at IS NULL`,
		name, nullIfEmpty(phone), gender, birthDate, id)
	return err
}
