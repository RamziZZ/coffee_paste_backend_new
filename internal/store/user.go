package store

import (
	"context"
	"database/sql"
	"errors"

	"coffeepaste/internal/model"
)

var ErrNotFound = errors.New("data tidak ditemukan")

// ErrGoogleLinked: akun dengan email itu sudah terhubung ke akun Google yang berbeda.
var ErrGoogleLinked = errors.New("akun sudah terhubung dengan akun Google lain")

type UserStore struct {
	DB *sql.DB
}

// FindByEmail mengembalikan user aktif beserta hash password-nya.
func (s *UserStore) FindByEmail(ctx context.Context, email string) (*model.User, string, error) {
	var u model.User
	var hash string
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, name, email, role, password_hash, token_version
		   FROM users
		  WHERE email = ? AND is_active = 1 AND deleted_at IS NULL
		  LIMIT 1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &hash, &u.TokenVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	if err != nil {
		return nil, "", err
	}
	return &u, hash, nil
}

func (s *UserStore) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, name, email, role, token_version
		   FROM users
		  WHERE id = ? AND is_active = 1 AND deleted_at IS NULL
		  LIMIT 1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.TokenVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Phone mengambil nomor telepon user (kosong bila belum diisi).
func (s *UserStore) Phone(ctx context.Context, id int64) (*string, error) {
	var p sql.NullString
	err := s.DB.QueryRowContext(ctx, `SELECT phone FROM users WHERE id = ?`, id).Scan(&p)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return nullStr(p), nil
}

// CreateCustomer mendaftarkan akun customer baru (role customer, aktif).
// Mengembalikan ErrDuplicateEmail bila email sudah dipakai.
func (s *UserStore) CreateCustomer(ctx context.Context, name, email, phone, passwordHash string) (*model.User, error) {
	var phoneArg any
	if phone != "" {
		phoneArg = phone
	}
	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO users (role, name, email, password_hash, phone) VALUES ('customer', ?, ?, ?, ?)`,
		name, email, passwordHash, phoneArg)
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
	return s.FindByID(ctx, id)
}

// FindByGoogleSub mengembalikan user aktif yang terhubung ke akun Google (sub) tersebut.
func (s *UserStore) FindByGoogleSub(ctx context.Context, sub string) (*model.User, error) {
	var u model.User
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, name, email, role, token_version
		   FROM users
		  WHERE google_sub = ? AND is_active = 1 AND deleted_at IS NULL
		  LIMIT 1`, sub,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.TokenVersion)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// LinkGoogle menghubungkan akun yang sudah ada ke akun Google. Hanya berhasil bila akun
// itu belum terhubung ke Google mana pun (atau sudah terhubung ke sub yang sama).
func (s *UserStore) LinkGoogle(ctx context.Context, id int64, sub string) error {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE users SET google_sub = ?
		  WHERE id = ? AND (google_sub IS NULL OR google_sub = ?)`, sub, id, sub)
	if isDuplicate(err) {
		return ErrGoogleLinked // sub ini sudah dipakai akun lain
	}
	if err != nil {
		return err
	}
	// RowsAffected = 0 juga terjadi bila nilainya sudah sama (MySQL), jadi cek ulang.
	if n, _ := res.RowsAffected(); n == 0 {
		var cur sql.NullString
		if err := s.DB.QueryRowContext(ctx, `SELECT google_sub FROM users WHERE id = ?`, id).Scan(&cur); err != nil {
			return err
		}
		if !cur.Valid || cur.String != sub {
			return ErrGoogleLinked
		}
	}
	return nil
}

// CreateGoogleCustomer mendaftarkan customer baru dari login Google. passwordHash berisi
// hash acak yang tak pernah diketahui siapa pun (kolom password_hash wajib terisi).
func (s *UserStore) CreateGoogleCustomer(ctx context.Context, name, email, sub, passwordHash string) (*model.User, error) {
	res, err := s.DB.ExecContext(ctx,
		`INSERT INTO users (role, name, email, password_hash, google_sub) VALUES ('customer', ?, ?, ?, ?)`,
		name, email, passwordHash, sub)
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
	return s.FindByID(ctx, id)
}
