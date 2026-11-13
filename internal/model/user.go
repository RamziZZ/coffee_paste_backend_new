package model

// User adalah bentuk data pengguna yang aman dikirim ke klien (tanpa password).
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`

	// TokenVersion naik setiap password diganti/direset; token lama dengan versi berbeda ditolak.
	TokenVersion int `json:"-"`
}
