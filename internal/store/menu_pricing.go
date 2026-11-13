package store

import (
	"context"
	"fmt"
)

// Harga minuman punya dua macam: Regular (kolom menu_items.price) dan Large (menu_items.price_large).
// Hanya kategori dengan is_customizable = 1 (Coffee, Non-Coffee) yang punya harga Large;
// untuk makanan dan pastry price_large selalu NULL.

// DefaultLargeSurcharge: selisih bawaan Large terhadap Regular (rupiah). Dipakai sebagai nilai awal
// di form admin (bisa diganti) dan sebagai cadangan bila price_large kosong.
// Harus sama dengan DEFAULT_LARGE_SURCHARGE di web (itemOptions.js) dan mobile (product_model.dart).
const DefaultLargeSurcharge int64 = 4000

// defaultLargeSurchargeSQL: nilai yang sama sebagai teks, untuk disisipkan ke query SQL konstan.
const defaultLargeSurchargeSQL = "4000"

// largePriceExpr: harga Large efektif per baris menu (NULL untuk kategori non-minuman).
// Membutuhkan alias m (menu_items) dan c (menu_categories).
const largePriceExpr = `CASE WHEN c.is_customizable = 1 THEN COALESCE(m.price_large, m.price + ` + defaultLargeSurchargeSQL + `) END`

// EffectiveLargePrice mengembalikan harga Large; bila belum diisi memakai harga Regular + selisih bawaan.
func EffectiveLargePrice(regular int64, large *int64) int64 {
	if large != nil && *large > 0 {
		return *large
	}
	return regular + DefaultLargeSurcharge
}

// UnitPrice mengembalikan harga satuan sesuai ukuran. Ukuran selain "Large" dihargai Regular.
func UnitPrice(regular int64, large *int64, size *string) int64 {
	if size != nil && *size == "Large" {
		return EffectiveLargePrice(regular, large)
	}
	return regular
}

// EnsurePriceSchema menambahkan kolom price_large bila belum ada, lalu mengisi minuman lama dengan
// harga Regular + DefaultLargeSurcharge. Aman dipanggil berulang (idempoten) dan bekerja di MySQL maupun MariaDB.
func (s *MenuStore) EnsurePriceSchema(ctx context.Context) error {
	var n int
	if err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM information_schema.COLUMNS
		  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'menu_items' AND COLUMN_NAME = 'price_large'`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		if _, err := s.DB.ExecContext(ctx,
			`ALTER TABLE menu_items ADD COLUMN price_large INT UNSIGNED NULL DEFAULT NULL AFTER price`); err != nil {
			return fmt.Errorf("tambah kolom price_large: %w", err)
		}
	}
	_, err := s.DB.ExecContext(ctx,
		`UPDATE menu_items m
		   JOIN menu_categories c ON c.id = m.category_id
		    SET m.price_large = m.price + ?
		  WHERE c.is_customizable = 1 AND m.price_large IS NULL`, DefaultLargeSurcharge)
	return err
}
