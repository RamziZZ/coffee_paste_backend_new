package store

import (
	"context"
	"database/sql"
	"errors"
)

// RecentResetExists true bila user sudah meminta reset dalam `seconds` detik terakhir
// (dipakai untuk membatasi spam email).
func (s *UserStore) RecentResetExists(ctx context.Context, userID int64, seconds int) (bool, error) {
	var n int
	err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM password_resets
		  WHERE user_id = ? AND created_at > DATE_SUB(NOW(), INTERVAL ? SECOND)`, userID, seconds).Scan(&n)
	return n > 0, err
}

// CreatePasswordReset menyimpan token reset (hash). Token lama yang belum dipakai dihapus.
func (s *UserStore) CreatePasswordReset(ctx context.Context, userID int64, tokenHash string, ttlMinutes int) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM password_resets WHERE user_id = ? AND used_at IS NULL`, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO password_resets (user_id, token_hash, expires_at)
		 VALUES (?, ?, DATE_ADD(NOW(), INTERVAL ? MINUTE))`, userID, tokenHash, ttlMinutes); err != nil {
		return err
	}
	return tx.Commit()
}

// DeletePasswordReset membatalkan token yang belum dipakai (mis. saat email gagal terkirim).
func (s *UserStore) DeletePasswordReset(ctx context.Context, userID int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM password_resets WHERE user_id = ? AND used_at IS NULL`, userID)
	return err
}

// ResetPassword memverifikasi token lalu mengganti password. Token hanya berlaku sekali.
func (s *UserStore) ResetPassword(ctx context.Context, tokenHash, newPasswordHash string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id, userID int64
	err = tx.QueryRowContext(ctx,
		`SELECT pr.id, pr.user_id
		   FROM password_resets pr
		   JOIN users u ON u.id = pr.user_id AND u.is_active = 1 AND u.deleted_at IS NULL
		  WHERE pr.token_hash = ? AND pr.used_at IS NULL AND pr.expires_at > NOW()
		  LIMIT 1 FOR UPDATE`, tokenHash).Scan(&id, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET password_hash = ?, token_version = token_version + 1 WHERE id = ?`, newPasswordHash, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE password_resets SET used_at = NOW() WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}
