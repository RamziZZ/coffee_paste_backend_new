package store

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrDuplicateEmail  = errors.New("email sudah terdaftar")
	ErrInvalidCategory = errors.New("kategori tidak valid")
)

type rowScanner interface {
	Scan(dest ...any) error
}

// isDuplicate true untuk error MySQL 1062 (duplicate entry).
func isDuplicate(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == 1062
}

func nullStr(n sql.NullString) *string {
	if !n.Valid {
		return nil
	}
	v := n.String
	return &v
}

// ValidationError adalah pelanggaran aturan bisnis yang pesannya aman ditampilkan ke pengguna.
// Handler membalasnya dengan 422.
type ValidationError string

func (e ValidationError) Error() string { return string(e) }
