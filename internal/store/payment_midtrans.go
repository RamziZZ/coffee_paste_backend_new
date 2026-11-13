package store

import (
	"context"
	"database/sql"
	"errors"
)

const midtransProvider = "midtrans"

// ActiveMidtransQR mengembalikan isi QR dari tagihan Midtrans yang masih berlaku
// (pending dan belum kedaluwarsa) untuk pesanan itu. Kosong bila tidak ada.
func (s *OrderStore) ActiveMidtransQR(ctx context.Context, orderID int64) (string, error) {
	var qr sql.NullString
	err := s.DB.QueryRowContext(ctx,
		`SELECT qr_string FROM payments
		  WHERE order_id = ? AND provider = ? AND status = 'pending' AND qr_string IS NOT NULL
		    AND (expires_at IS NULL OR expires_at > NOW())
		  ORDER BY id DESC LIMIT 1`, orderID, midtransProvider).Scan(&qr)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return qr.String, nil
}

// SaveMidtransCharge menyimpan tagihan QRIS Midtrans. ref = order_id yang dikirim ke Midtrans.
// Batas berlaku disimpan 1 menit lebih awal dari expiry di Midtrans supaya QR yang hampir
// kedaluwarsa tidak ditampilkan.
func (s *OrderStore) SaveMidtransCharge(ctx context.Context, orderID int64, ref, qr string, amount int64, ttlMinutes int) error {
	margin := ttlMinutes - 1
	if margin < 1 {
		margin = 1
	}
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO payments (order_id, provider, provider_ref, amount, status, qr_string, expires_at)
		 VALUES (?, ?, ?, ?, 'pending', ?, DATE_ADD(NOW(), INTERVAL ? MINUTE))`,
		orderID, midtransProvider, ref, amount, qr, margin)
	return err
}

// FindMidtransPayment mencari tagihan Midtrans berdasarkan order_id Midtrans (ref).
func (s *OrderStore) FindMidtransPayment(ctx context.Context, ref string) (orderID int64, status string, err error) {
	err = s.DB.QueryRowContext(ctx,
		`SELECT order_id, status FROM payments WHERE provider = ? AND provider_ref = ? LIMIT 1`,
		midtransProvider, ref).Scan(&orderID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrNotFound
	}
	return orderID, status, err
}

// SetMidtransStatus mengubah tagihan Midtrans yang masih pending (mis. jadi expired/failed).
func (s *OrderStore) SetMidtransStatus(ctx context.Context, ref, status string) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE payments SET status = ? WHERE provider = ? AND provider_ref = ? AND status = 'pending'`,
		status, midtransProvider, ref)
	return err
}

// MarkMidtransPaid menandai pesanan lunas berdasarkan notifikasi settlement Midtrans.
// Idempoten. Mengembalikan ErrNotFound bila ref tidak dikenal, dan ValidationError bila
// nominal tidak cocok atau pesanan sudah dibatalkan (perlu penanganan/refund manual).
func (s *OrderStore) MarkMidtransPaid(ctx context.Context, ref string, amount int64) (*Order, error) {
	orderID, _, err := s.FindMidtransPayment(ctx, ref)
	if err != nil {
		return nil, err
	}
	o, err := s.MarkQRISPaid(ctx, orderID, amount)
	if err != nil {
		return nil, err
	}
	if _, err := s.DB.ExecContext(ctx,
		`UPDATE payments SET status = 'paid', paid_at = COALESCE(paid_at, NOW())
		  WHERE provider = ? AND provider_ref = ? AND status IN ('pending', 'expired')`,
		midtransProvider, ref); err != nil {
		return nil, err
	}
	return o, nil
}

// PendingMidtransRefs mengembalikan ref tagihan Midtrans yang masih menunggu pembayaran:
// pesanan belum lunas, belum dibatalkan, dan tagihan dibuat dalam beberapa jam terakhir
// (lebih lama dari itu pasti sudah kedaluwarsa di Midtrans). Dipakai pemeriksa status berkala.
func (s *OrderStore) PendingMidtransRefs(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.DB.QueryContext(ctx,
		`SELECT p.provider_ref
		   FROM payments p JOIN orders o ON o.id = p.order_id
		  WHERE p.provider = ? AND p.status = 'pending' AND p.provider_ref IS NOT NULL
		    AND o.paid = 0 AND o.status <> 'dibatalkan'
		    AND p.created_at > DATE_SUB(NOW(), INTERVAL 3 HOUR)
		  ORDER BY p.id DESC LIMIT ?`, midtransProvider, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refs []string
	for rows.Next() {
		var ref string
		if err := rows.Scan(&ref); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}
