package store

import (
	"context"
	"database/sql"
)

type MenuStore struct {
	DB *sql.DB
}

type MenuItem struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	PriceLarge  *int64 // harga Large (hanya diisi oleh Categories; nil untuk non-minuman)
	Image       *string
}

type Category struct {
	ID          int64
	Name        string
	FilterLabel string
	Icon        string
	Items       []MenuItem
}

// BestSellers: menu terlaris dari pesanan berstatus selesai (view v_menu_best_sellers).
func (s *MenuStore) BestSellers(ctx context.Context, limit int) ([]MenuItem, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, name, COALESCE(description, ''), price, image_url
		   FROM v_menu_best_sellers
		  WHERE total_sold > 0
		  ORDER BY total_sold DESC, id ASC
		  LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	return scanItems(rows)
}

// FeaturedFlagged: cadangan bila belum ada penjualan, memakai kolom is_featured.
func (s *MenuStore) FeaturedFlagged(ctx context.Context, limit int) ([]MenuItem, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, name, COALESCE(description, ''), price, image_url
		   FROM menu_items
		  WHERE is_featured = 1 AND deleted_at IS NULL
		  ORDER BY id ASC
		  LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	return scanItems(rows)
}

// Categories: seluruh kategori beserta menu yang belum dihapus.
func (s *MenuStore) Categories(ctx context.Context) ([]Category, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, name, COALESCE(filter_label, ''), COALESCE(icon, '')
		   FROM menu_categories
		  ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cats := make([]Category, 0)
	index := make(map[int64]int)
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.FilterLabel, &c.Icon); err != nil {
			return nil, err
		}
		c.Items = make([]MenuItem, 0)
		index[c.ID] = len(cats)
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	itemRows, err := s.DB.QueryContext(ctx,
		`SELECT m.category_id, m.id, m.name, COALESCE(m.description, ''), m.price, `+largePriceExpr+`, m.image_url
		   FROM menu_items m
		   JOIN menu_categories c ON c.id = m.category_id
		  WHERE m.deleted_at IS NULL
		  ORDER BY m.category_id ASC, m.id ASC`)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var catID int64
		var it MenuItem
		var img sql.NullString
		var large sql.NullInt64
		if err := itemRows.Scan(&catID, &it.ID, &it.Name, &it.Description, &it.Price, &large, &img); err != nil {
			return nil, err
		}
		if large.Valid {
			v := large.Int64
			it.PriceLarge = &v
		}
		if img.Valid && img.String != "" {
			v := img.String
			it.Image = &v
		}
		if i, ok := index[catID]; ok {
			cats[i].Items = append(cats[i].Items, it)
		}
	}
	return cats, itemRows.Err()
}

func scanItems(rows *sql.Rows) ([]MenuItem, error) {
	defer rows.Close()

	items := make([]MenuItem, 0)
	for rows.Next() {
		var it MenuItem
		var img sql.NullString
		if err := rows.Scan(&it.ID, &it.Name, &it.Description, &it.Price, &img); err != nil {
			return nil, err
		}
		if img.Valid && img.String != "" {
			v := img.String
			it.Image = &v
		}
		items = append(items, it)
	}
	return items, rows.Err()
}
