package store

import (
	"context"
	"database/sql"
	"errors"
)

// MarkQRISPaid menandai pesanan QRIS lunas berdasarkan notifikasi payment gateway.
// Idempoten: pesanan yang sudah lunas dikembalikan apa adanya. Nominal harus sama persis
// dengan total pesanan, dan pesanan yang sudah dibatalkan ditolak (perlu refund manual).
func (s *OrderStore) MarkQRISPaid(ctx context.Context, id, amount int64) (*Order, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var status, method string
	var paid bool
	var total int64
	err = tx.QueryRowContext(ctx,
		`SELECT status, paid, payment_method, total FROM orders WHERE id = ? FOR UPDATE`, id,
	).Scan(&status, &paid, &method, &total)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if method != "QRIS" {
		return nil, ValidationError("Pesanan ini bukan pembayaran QRIS")
	}
	if status == "dibatalkan" {
		return nil, ValidationError("Pesanan sudah dibatalkan")
	}

	if !paid {
		if amount != total {
			return nil, ValidationError("Nominal pembayaran tidak sesuai tagihan")
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE orders SET paid = 1, paid_at = NOW() WHERE id = ?`, id); err != nil {
			return nil, err
		}
		if err := recordQRIS(ctx, tx, id, total, true); err != nil {
			return nil, err
		}
	}

	o, err := getOrder(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return o, nil
}
