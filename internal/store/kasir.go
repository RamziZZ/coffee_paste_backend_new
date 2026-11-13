package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"coffeepaste/internal/format"
)

type KasirStore struct {
	DB *sql.DB
}

type Kasir struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	Address    string  `json:"address"`
	Gender     *string `json:"gender"`
	BirthPlace *string `json:"birthPlace"`
	BirthDate  *string `json:"birthDate"`
	JoinedAt   string  `json:"joinedAt"`
}

type KasirInput struct {
	Name       string
	Email      string
	Phone      string
	Address    string
	Gender     string
	BirthPlace *string // nil = tidak diubah
	BirthDate  *string // nil = tidak diubah, format YYYY-MM-DD
}

const kasirSelect = `SELECT id, name, email, COALESCE(phone, ''), COALESCE(address, ''),
	       gender, birth_place, DATE_FORMAT(birth_date, '%Y-%m-%d'), created_at
	  FROM users
	 WHERE role = 'kasir' AND deleted_at IS NULL`

func scanKasir(r rowScanner) (Kasir, error) {
	var k Kasir
	var gender, bp, bd sql.NullString
	var created time.Time
	if err := r.Scan(&k.ID, &k.Name, &k.Email, &k.Phone, &k.Address, &gender, &bp, &bd, &created); err != nil {
		return k, err
	}
	k.Gender = nullStr(gender)
	k.BirthPlace = nullStr(bp)
	k.BirthDate = nullStr(bd)
	k.JoinedAt = created.Format(format.ISO)
	return k, nil
}

func (s *KasirStore) List(ctx context.Context) ([]Kasir, error) {
	rows, err := s.DB.QueryContext(ctx, kasirSelect+` ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Kasir, 0)
	for rows.Next() {
		k, err := scanKasir(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (s *KasirStore) Get(ctx context.Context, id int64) (*Kasir, error) {
	k, err := scanKasir(s.DB.QueryRowContext(ctx, kasirSelect+` AND id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (s *KasirStore) Create(ctx context.Context, in KasirInput, passwordHash string) (*Kasir, error) {
	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO users (role, name, email, password_hash, phone, address, gender, birth_place, birth_date)
		 VALUES ('kasir', ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.Name, in.Email, passwordHash, in.Phone, in.Address, in.Gender, in.BirthPlace, in.BirthDate)
	if isDuplicate(err) {
		return nil, ErrDuplicateEmail
	}
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *KasirStore) Update(ctx context.Context, id int64, in KasirInput) (*Kasir, error) {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE users
		    SET name = ?, email = ?, phone = ?, address = ?, gender = ?,
		        birth_place = COALESCE(?, birth_place), birth_date = COALESCE(?, birth_date)
		  WHERE id = ? AND role = 'kasir' AND deleted_at IS NULL`,
		in.Name, in.Email, in.Phone, in.Address, in.Gender, in.BirthPlace, in.BirthDate, id)
	if isDuplicate(err) {
		return nil, ErrDuplicateEmail
	}
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return s.Get(ctx, id)
}

// Delete: soft delete + nonaktif. Baris users TIDAK dihapus, jadi seluruh riwayat pesanan
// (orders.cashier_id), log status, dan nama kasir di laporan tetap utuh. Akun yang sudah
// dihapus tidak bisa login dan tidak muncul di daftar kasir (lihat deleted_at IS NULL).
func (s *KasirStore) Delete(ctx context.Context, id int64) error {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE users SET deleted_at = NOW(), is_active = 0
		  WHERE id = ? AND role = 'kasir' AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
