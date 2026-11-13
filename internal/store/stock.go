package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"coffeepaste/internal/format"
)

type StockStore struct {
	DB *sql.DB
}

type StockShortage struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Note         string  `json:"note"`
	Status       string  `json:"status"` // "habis" (perlu dibeli) | "dibeli"
	ReportedBy   string  `json:"reportedBy"`
	ReportedByID string  `json:"reportedById"`
	ReportedAt   string  `json:"reportedAt"`
	PurchasedAt  *string `json:"purchasedAt"`
}

type StockInput struct {
	Name     string
	Category string
	Note     string
}

const stockSchema = `CREATE TABLE IF NOT EXISTS stock_shortages (
  id bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  name varchar(80) NOT NULL,
  category varchar(30) NOT NULL,
  note varchar(200) NOT NULL DEFAULT '',
  status enum('habis','dibeli') NOT NULL DEFAULT 'habis',
  reported_by bigint(20) UNSIGNED DEFAULT NULL,
  reported_by_name varchar(100) NOT NULL,
  reported_at datetime NOT NULL DEFAULT current_timestamp(),
  purchased_by bigint(20) UNSIGNED DEFAULT NULL,
  purchased_at datetime DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_stock_status (status, reported_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

// EnsureSchema membuat tabel laporan stok habis bila belum ada (aman dipanggil berulang).
func (s *StockStore) EnsureSchema(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, stockSchema)
	return err
}

const stockSelect = `SELECT s.id, s.name, s.category, s.note, s.status,
	       COALESCE(s.reported_by, 0), s.reported_by_name, s.reported_at, s.purchased_at
	  FROM stock_shortages s`

func scanStock(r rowScanner) (StockShortage, error) {
	var x StockShortage
	var reportedBy int64
	var reportedAt time.Time
	var purchasedAt sql.NullTime
	err := r.Scan(&x.ID, &x.Name, &x.Category, &x.Note, &x.Status,
		&reportedBy, &x.ReportedBy, &reportedAt, &purchasedAt)
	if err != nil {
		return x, err
	}
	if reportedBy > 0 {
		x.ReportedByID = strconv.FormatInt(reportedBy, 10)
	}
	x.ReportedAt = reportedAt.Format(format.ISO)
	if purchasedAt.Valid {
		v := purchasedAt.Time.Format(format.ISO)
		x.PurchasedAt = &v
	}
	return x, nil
}

// List mengembalikan semua laporan yang masih perlu dibeli, ditambah laporan yang sudah
// dibeli dalam `purchasedDays` hari terakhir.
func (s *StockStore) List(ctx context.Context, purchasedDays int) ([]StockShortage, error) {
	rows, err := s.DB.QueryContext(ctx,
		stockSelect+` WHERE s.status = 'habis' OR s.purchased_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		 ORDER BY s.reported_at DESC, s.id DESC LIMIT 500`, purchasedDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]StockShortage, 0)
	for rows.Next() {
		x, err := scanStock(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s *StockStore) Get(ctx context.Context, id int64) (*StockShortage, error) {
	x, err := scanStock(s.DB.QueryRowContext(ctx, stockSelect+` WHERE s.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &x, nil
}

// Create menyimpan laporan baru. Bahan yang sama (tanpa membedakan huruf besar/kecil) yang
// masih perlu dibeli tidak boleh dilaporkan dua kali.
func (s *StockStore) Create(ctx context.Context, in StockInput, userID int64, userName string) (*StockShortage, error) {
	var pending int
	err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM stock_shortages WHERE status = 'habis' AND name = ?`, in.Name).Scan(&pending)
	if err != nil {
		return nil, err
	}
	if pending > 0 {
		return nil, ValidationError("\"" + in.Name + "\" sudah dilaporkan habis dan belum dibeli")
	}

	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO stock_shortages (name, category, note, reported_by, reported_by_name)
		 VALUES (?, ?, ?, ?, ?)`,
		in.Name, in.Category, in.Note, userID, userName)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// DeleteOwnPending membatalkan laporan milik kasir yang belum ditandai dibeli.
func (s *StockStore) DeleteOwnPending(ctx context.Context, id, userID int64) error {
	res, err := s.DB.ExecContext(ctx,
		`DELETE FROM stock_shortages WHERE id = ? AND reported_by = ? AND status = 'habis'`, id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return nil
	}
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	return ValidationError("Hanya laporan milik Anda yang belum dibeli yang bisa dibatalkan")
}

// SetStatus dipakai admin untuk menandai bahan sudah dibeli, atau mengembalikannya ke daftar belanja.
func (s *StockStore) SetStatus(ctx context.Context, id int64, status string, adminID int64) (*StockShortage, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	_, err := s.DB.ExecContext(ctx,
		`UPDATE stock_shortages
		    SET status = ?,
		        purchased_at = IF(? = 'dibeli', NOW(), NULL),
		        purchased_by = IF(? = 'dibeli', ?, NULL)
		  WHERE id = ?`,
		status, status, status, adminID, id)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}
