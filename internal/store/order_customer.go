package store

import (
	"context"
	"database/sql"
	"errors"
)

// CancelByCustomer membatalkan pesanan milik customer. Hanya boleh selama pesanan masih
// berstatus menunggu (belum disentuh kasir) dan belum lunas; selebihnya kasir yang
// membatalkan lewat PUT /kasir/pesanan/:id. Pesanan milik customer lain dianggap tidak ada.
func (s *OrderStore) CancelByCustomer(ctx context.Context, id, customerID int64) (*Order, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var status string
	var paid bool
	var owner sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT status, paid, customer_id FROM orders WHERE id = ? FOR UPDATE`, id).Scan(&status, &paid, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !owner.Valid || owner.Int64 != customerID {
		return nil, ErrNotFound
	}
	if status == "dibatalkan" {
		return nil, ValidationError("Pesanan sudah dibatalkan")
	}
	if status != "menunggu" || paid {
		return nil, ValidationError("Pesanan sudah diproses kasir dan tidak dapat dibatalkan dari aplikasi. Hubungi kasir.")
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE orders SET status = 'dibatalkan', cancel_reason = ?, cancelled_at = NOW() WHERE id = ?`,
		"Dibatalkan oleh pelanggan", id); err != nil {
		return nil, err
	}
	if err := logStatus(ctx, tx, id, &status, "dibatalkan", customerID); err != nil {
		return nil, err
	}
	// Tagihan QRIS yang belum dibayar ikut gugur.
	if _, err := tx.ExecContext(ctx,
		`UPDATE payments SET status = 'failed' WHERE order_id = ? AND status = 'pending'`, id); err != nil {
		return nil, err
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
