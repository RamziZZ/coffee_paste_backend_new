package store

import (
	"context"
	"database/sql"
	"time"

	"coffeepaste/internal/format"
)

type ReviewStore struct {
	DB *sql.DB
}

// PublicReview: satu ulasan yang tampil di landing page.
type PublicReview struct {
	ID        int64
	Name      string
	Comment   string
	Rating    int
	Product   string // nama produk yang diulas; kosong untuk ulasan umum
	CreatedAt time.Time
}

// Published: review yang dipublikasikan, memiliki komentar, rating minimal minRating,
// terbaru dulu. Ulasan yang hanya berisi bintang (tanpa komentar) tidak ditampilkan
// sebagai kartu, tetapi tetap dihitung di rating produk dan ringkasan.
func (s *ReviewStore) Published(ctx context.Context, minRating, limit int) ([]PublicReview, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT r.id, u.name, r.comment, r.rating, COALESCE(m.name, ''), r.created_at
		   FROM reviews r
		   JOIN users u ON u.id = r.user_id
		   LEFT JOIN menu_items m ON m.id = r.menu_item_id
		  WHERE r.is_published = 1 AND r.rating >= ? AND TRIM(r.comment) <> ''
		  ORDER BY r.created_at DESC, r.id DESC
		  LIMIT ?`, minRating, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]PublicReview, 0)
	for rows.Next() {
		var rv PublicReview
		if err := rows.Scan(&rv.ID, &rv.Name, &rv.Comment, &rv.Rating, &rv.Product, &rv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

// RatingSummary: rata-rata dan jumlah dari SEMUA ulasan yang dipublikasikan.
type RatingSummary struct {
	Average float64
	Total   int
}

func (s *ReviewStore) Summary(ctx context.Context) (RatingSummary, error) {
	var sum RatingSummary
	err := s.DB.QueryRowContext(ctx,
		`SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM reviews WHERE is_published = 1`,
	).Scan(&sum.Average, &sum.Total)
	return sum, err
}

// ProductRating: rating rata-rata satu produk.
type ProductRating struct {
	Average float64
	Count   int
}

// ProductRatings: rating semua produk yang pernah diulas (kunci = id menu).
func (s *ReviewStore) ProductRatings(ctx context.Context) (map[int64]ProductRating, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT menu_item_id, AVG(rating), COUNT(*)
		   FROM reviews
		  WHERE is_published = 1 AND menu_item_id IS NOT NULL
		  GROUP BY menu_item_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64]ProductRating)
	for rows.Next() {
		var id int64
		var pr ProductRating
		if err := rows.Scan(&id, &pr.Average, &pr.Count); err != nil {
			return nil, err
		}
		out[id] = pr
	}
	return out, rows.Err()
}

// ReviewedItems: produk yang sudah diulas customer, dikelompokkan per id pesanan (numerik).
func (s *ReviewStore) ReviewedItems(ctx context.Context, userID int64) (map[int64][]int64, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT order_id, menu_item_id
		   FROM reviews
		  WHERE user_id = ? AND order_id IS NOT NULL AND menu_item_id IS NOT NULL`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int64][]int64)
	for rows.Next() {
		var orderID, menuID int64
		if err := rows.Scan(&orderID, &menuID); err != nil {
			return nil, err
		}
		out[orderID] = append(out[orderID], menuID)
	}
	return out, rows.Err()
}

type AdminReview struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Rating      int    `json:"rating"`
	Comment     string `json:"comment"`
	Product     string `json:"product"`
	OrderID     *int64 `json:"orderId"`
	IsPublished bool   `json:"isPublished"`
	CreatedAt   string `json:"createdAt"`
}

// AdminList: semua review (termasuk yang disembunyikan), terbaru dulu.
func (s *ReviewStore) AdminList(ctx context.Context) ([]AdminReview, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT r.id, u.name, r.rating, r.comment, COALESCE(m.name, ''), r.order_id, r.is_published, r.created_at
		   FROM reviews r
		   JOIN users u ON u.id = r.user_id
		   LEFT JOIN menu_items m ON m.id = r.menu_item_id
		  ORDER BY r.created_at DESC, r.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AdminReview, 0)
	for rows.Next() {
		var rv AdminReview
		var orderID sql.NullInt64
		var created time.Time
		if err := rows.Scan(&rv.ID, &rv.Name, &rv.Rating, &rv.Comment, &rv.Product, &orderID, &rv.IsPublished, &created); err != nil {
			return nil, err
		}
		if orderID.Valid {
			v := orderID.Int64
			rv.OrderID = &v
		}
		rv.CreatedAt = created.Format(format.ISO)
		out = append(out, rv)
	}
	return out, rows.Err()
}

func (s *ReviewStore) SetPublished(ctx context.Context, id int64, published bool) error {
	res, err := s.DB.ExecContext(ctx, `UPDATE reviews SET is_published = ? WHERE id = ?`, published, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
