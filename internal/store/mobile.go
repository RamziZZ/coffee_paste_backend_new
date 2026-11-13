package store

import (
	"context"
	"database/sql"
	"errors"
)

// ---- Review

// Create menyimpan review umum customer (tanpa produk). orderID opsional; bila diisi,
// pesanan harus milik customer dan berstatus selesai, dan pesanan itu belum punya review umum.
// Untuk ulasan per produk gunakan CreateForOrder.
func (s *ReviewStore) Create(ctx context.Context, userID int64, orderID *int64, rating int, comment string) (int64, error) {
	if orderID != nil {
		if err := s.checkCompletedOrder(ctx, s.DB, userID, *orderID); err != nil {
			return 0, err
		}
		// Indeks unik (order_id, menu_item_id) tidak menolak NULL ganda, jadi dicek manual.
		var n int
		if err := s.DB.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM reviews WHERE order_id = ? AND menu_item_id IS NULL`, *orderID).Scan(&n); err != nil {
			return 0, err
		}
		if n > 0 {
			return 0, ValidationError("Pesanan ini sudah pernah diulas")
		}
	}
	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO reviews (user_id, order_id, rating, comment) VALUES (?, ?, ?, ?)`,
		userID, orderID, rating, comment)
	if isDuplicate(err) {
		return 0, ValidationError("Pesanan ini sudah pernah diulas")
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type sqlQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// checkCompletedOrder memastikan pesanan milik customer dan sudah selesai.
func (s *ReviewStore) checkCompletedOrder(ctx context.Context, q sqlQueryer, userID, orderID int64) error {
	var owner sql.NullInt64
	var status string
	err := q.QueryRowContext(ctx,
		`SELECT customer_id, status FROM orders WHERE id = ?`, orderID).Scan(&owner, &status)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && (!owner.Valid || owner.Int64 != userID)) {
		return ValidationError("Pesanan tidak ditemukan")
	}
	if err != nil {
		return err
	}
	if status != "selesai" {
		return ValidationError("Review hanya bisa diberikan untuk pesanan yang sudah selesai")
	}
	return nil
}

// ProductReviewInput: satu ulasan produk dalam sebuah pesanan.
type ProductReviewInput struct {
	MenuItemID int64
	Rating     int
	Comment    string // boleh kosong (rating saja)
}

// CreateForOrder menyimpan ulasan beberapa produk dari satu pesanan sekaligus (atomik).
// Aturan: pesanan milik customer dan berstatus selesai; setiap produk harus ada di pesanan
// itu; satu produk hanya boleh diulas sekali per pesanan. Jika salah satu gagal, tidak ada
// yang tersimpan. Mengembalikan id review yang dibuat, sesuai urutan input.
func (s *ReviewStore) CreateForOrder(ctx context.Context, userID, orderID int64, in []ProductReviewInput) ([]int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.checkCompletedOrder(ctx, tx, userID, orderID); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx,
		`SELECT DISTINCT menu_item_id FROM order_items WHERE order_id = ? AND menu_item_id IS NOT NULL`, orderID)
	if err != nil {
		return nil, err
	}
	inOrder := make(map[int64]bool)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		inOrder[id] = true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	ids := make([]int64, 0, len(in))
	for _, rv := range in {
		if !inOrder[rv.MenuItemID] {
			return nil, ValidationError("Produk tidak ada di pesanan ini")
		}
		res, err := tx.ExecContext(ctx,
			`INSERT INTO reviews (user_id, order_id, menu_item_id, rating, comment) VALUES (?, ?, ?, ?, ?)`,
			userID, orderID, rv.MenuItemID, rv.Rating, rv.Comment)
		if isDuplicate(err) {
			return nil, ValidationError("Produk ini sudah pernah diulas pada pesanan ini")
		}
		if err != nil {
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}

// ---- Favorit

// Favorites: menu favorit milik customer (yang belum dihapus), terbaru dulu.
func (s *MenuStore) Favorites(ctx context.Context, userID int64) ([]AdminMenuItem, error) {
	rows, err := s.DB.QueryContext(ctx,
		adminMenuSelect+` AND m.id IN (SELECT menu_item_id FROM favorite_menu WHERE user_id = ?)
		 ORDER BY (SELECT f.created_at FROM favorite_menu f WHERE f.user_id = ? AND f.menu_item_id = m.id) DESC, m.id DESC`,
		userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AdminMenuItem, 0)
	for rows.Next() {
		it, err := scanAdminMenu(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// AddFavorite idempoten: menambah favorit yang sudah ada tidak error.
func (s *MenuStore) AddFavorite(ctx context.Context, userID, menuItemID int64) (*AdminMenuItem, error) {
	it, err := s.GetAdmin(ctx, menuItemID)
	if err != nil {
		return nil, err
	}
	if _, err := s.DB.ExecContext(ctx,
		`INSERT INTO favorite_menu (user_id, menu_item_id) VALUES (?, ?)
		 ON DUPLICATE KEY UPDATE created_at = created_at`, userID, menuItemID); err != nil {
		return nil, err
	}
	return it, nil
}

func (s *MenuStore) RemoveFavorite(ctx context.Context, userID, menuItemID int64) error {
	_, err := s.DB.ExecContext(ctx,
		`DELETE FROM favorite_menu WHERE user_id = ? AND menu_item_id = ?`, userID, menuItemID)
	return err
}

// ---- Token perangkat

// SaveDeviceToken menyimpan token notifikasi. Token yang sama otomatis berpindah ke
// user terbaru (mis. ganti akun di perangkat yang sama).
func (s *UserStore) SaveDeviceToken(ctx context.Context, userID int64, token, platform string) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO device_tokens (user_id, token, platform) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE user_id = VALUES(user_id), platform = VALUES(platform)`,
		userID, token, platform)
	return err
}

// DeviceTokens mengembalikan semua token perangkat milik user.
func (s *UserStore) DeviceTokens(ctx context.Context, userID int64) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT token FROM device_tokens WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// DeleteDeviceToken menghapus token yang sudah tidak berlaku (ditolak FCM).
func (s *UserStore) DeleteDeviceToken(ctx context.Context, token string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM device_tokens WHERE token = ?`, token)
	return err
}

// RemoveDeviceToken menghapus token milik user itu sendiri (saat logout).
func (s *UserStore) RemoveDeviceToken(ctx context.Context, userID int64, token string) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM device_tokens WHERE user_id = ? AND token = ?`, userID, token)
	return err
}
