package store

import (
	"context"
	"database/sql"
	"errors"
)

// CustomerOf mengembalikan id customer pemilik pesanan; 0 bila pesanan dibuat kasir tanpa akun customer.
func (s *OrderStore) CustomerOf(ctx context.Context, id int64) (int64, error) {
	var owner sql.NullInt64
	err := s.DB.QueryRowContext(ctx, `SELECT customer_id FROM orders WHERE id = ?`, id).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return owner.Int64, nil // NULL -> 0
}
