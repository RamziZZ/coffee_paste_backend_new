package store

import (
	"context"
	"database/sql"
	"errors"
)

type AdminMenuItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       int64   `json:"price"`      // harga Regular
	PriceLarge  *int64  `json:"priceLarge"` // harga Large; null untuk makanan/pastry
	Status      string  `json:"status"`
	Image       *string `json:"image"`
	Description string  `json:"description"`
}

type MenuInput struct {
	Name        string
	Category    string
	Price       int64
	PriceLarge  *int64 // diabaikan untuk kategori non-minuman; kosong pada minuman = Price + DefaultLargeSurcharge
	Status      string
	Image       *string
	Description string
}

const adminMenuSelect = `SELECT m.id, m.name, c.name, m.price, ` + largePriceExpr + `, m.status, m.image_url, COALESCE(m.description, '')
	  FROM menu_items m
	  JOIN menu_categories c ON c.id = m.category_id
	 WHERE m.deleted_at IS NULL`

func scanAdminMenu(r rowScanner) (AdminMenuItem, error) {
	var it AdminMenuItem
	var img sql.NullString
	var large sql.NullInt64
	err := r.Scan(&it.ID, &it.Name, &it.Category, &it.Price, &large, &it.Status, &img, &it.Description)
	if large.Valid {
		v := large.Int64
		it.PriceLarge = &v
	}
	it.Image = nullStr(img)
	if it.Image != nil && *it.Image == "" {
		it.Image = nil
	}
	return it, err
}

// CategoryNames: nama kategori untuk filter dan dropdown form.
func (s *MenuStore) CategoryNames(ctx context.Context) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT name FROM menu_categories ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0)
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// ListAdmin: semua menu yang belum dihapus. Dipakai juga oleh kasir.
func (s *MenuStore) ListAdmin(ctx context.Context) ([]AdminMenuItem, error) {
	rows, err := s.DB.QueryContext(ctx, adminMenuSelect+` ORDER BY m.created_at DESC, m.id DESC`)
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

func (s *MenuStore) GetAdmin(ctx context.Context, id int64) (*AdminMenuItem, error) {
	it, err := scanAdminMenu(s.DB.QueryRowContext(ctx, adminMenuSelect+` AND m.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &it, nil
}

// category mengembalikan id kategori dan apakah kategori itu minuman (is_customizable).
func (s *MenuStore) category(ctx context.Context, name string) (id int64, drink bool, err error) {
	err = s.DB.QueryRowContext(ctx,
		`SELECT id, is_customizable FROM menu_categories WHERE name = ?`, name).Scan(&id, &drink)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, ErrInvalidCategory
	}
	return id, drink, err
}

// largeForSave menentukan nilai price_large yang disimpan: NULL untuk non-minuman; untuk minuman
// memakai input admin, atau Regular + selisih bawaan bila kosong. Large tidak boleh lebih murah dari Regular.
func largeForSave(in MenuInput, drink bool) (*int64, error) {
	if !drink {
		return nil, nil
	}
	if in.PriceLarge == nil {
		v := in.Price + DefaultLargeSurcharge
		return &v, nil
	}
	if *in.PriceLarge < in.Price {
		return nil, ValidationError("Harga Large tidak boleh lebih rendah dari harga Regular")
	}
	return in.PriceLarge, nil
}

func (s *MenuStore) Create(ctx context.Context, in MenuInput) (*AdminMenuItem, error) {
	catID, drink, err := s.category(ctx, in.Category)
	if err != nil {
		return nil, err
	}
	large, err := largeForSave(in, drink)
	if err != nil {
		return nil, err
	}
	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO menu_items (category_id, name, description, price, price_large, status, image_url)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		catID, in.Name, in.Description, in.Price, large, in.Status, in.Image)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetAdmin(ctx, id)
}

func (s *MenuStore) Update(ctx context.Context, id int64, in MenuInput) (*AdminMenuItem, error) {
	catID, drink, err := s.category(ctx, in.Category)
	if err != nil {
		return nil, err
	}
	large, err := largeForSave(in, drink)
	if err != nil {
		return nil, err
	}
	res, err := s.DB.ExecContext(ctx,
		`UPDATE menu_items
		    SET category_id = ?, name = ?, description = ?, price = ?, price_large = ?, status = ?, image_url = ?
		  WHERE id = ? AND deleted_at IS NULL`,
		catID, in.Name, in.Description, in.Price, large, in.Status, in.Image, id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return s.GetAdmin(ctx, id)
}

// SoftDelete mengisi deleted_at; baris tetap ada agar riwayat pesanan utuh.
func (s *MenuStore) SoftDelete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE menu_items SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// SetStatus hanya mengubah ketersediaan menu (dipakai kasir).
func (s *MenuStore) SetStatus(ctx context.Context, id int64, status string) (*AdminMenuItem, error) {
	if _, err := s.DB.ExecContext(ctx,
		`UPDATE menu_items SET status = ? WHERE id = ? AND deleted_at IS NULL`, status, id); err != nil {
		return nil, err
	}
	// RowsAffected = 0 juga terjadi bila status sudah sama, jadi keberadaan menu dicek lewat GetAdmin.
	return s.GetAdmin(ctx, id)
}
