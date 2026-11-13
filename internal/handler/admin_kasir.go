package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/store"

	"golang.org/x/crypto/bcrypt"
)

type AdminKasirHandler struct {
	kasir           *store.KasirStore
	defaultPassword string
}

func NewAdminKasirHandler(kasir *store.KasirStore, defaultPassword string) *AdminKasirHandler {
	return &AdminKasirHandler{kasir: kasir, defaultPassword: defaultPassword}
}

type kasirRequest struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
	Address    string  `json:"address"`
	Gender     string  `json:"gender"`
	BirthPlace *string `json:"birthPlace"`
	BirthDate  *string `json:"birthDate"`
}

func (k kasirRequest) toInput() (store.KasirInput, string) {
	name := strings.TrimSpace(k.Name)
	email := strings.TrimSpace(k.Email)
	phone := strings.TrimSpace(k.Phone)
	address := strings.TrimSpace(k.Address)

	switch {
	case name == "":
		return store.KasirInput{}, "Nama wajib diisi"
	case utf8.RuneCountInString(name) > 100:
		return store.KasirInput{}, "Nama maksimal 100 karakter"
	case !validEmail(email):
		return store.KasirInput{}, "Format email tidak valid"
	case phone == "":
		return store.KasirInput{}, "Nomor telepon wajib diisi"
	case utf8.RuneCountInString(phone) > 30:
		return store.KasirInput{}, "Nomor telepon maksimal 30 karakter"
	case address == "":
		return store.KasirInput{}, "Alamat wajib diisi"
	case k.Gender != "Laki-laki" && k.Gender != "Perempuan":
		return store.KasirInput{}, "Jenis kelamin harus Laki-laki atau Perempuan"
	}

	in := store.KasirInput{Name: name, Email: email, Phone: phone, Address: address, Gender: k.Gender}

	if k.BirthPlace != nil {
		if v := strings.TrimSpace(*k.BirthPlace); v != "" {
			if utf8.RuneCountInString(v) > 100 {
				return store.KasirInput{}, "Tempat lahir maksimal 100 karakter"
			}
			in.BirthPlace = &v
		}
	}
	if k.BirthDate != nil {
		if v := strings.TrimSpace(*k.BirthDate); v != "" {
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return store.KasirInput{}, "Tanggal lahir harus berformat YYYY-MM-DD"
			}
			in.BirthDate = &v
		}
	}
	return in, ""
}

// List: GET /api/admin/kasir
func (h *AdminKasirHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.kasir.List(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

// Create: POST /api/admin/kasir
func (h *AdminKasirHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req kasirRequest
	if !decodeOrFail(w, r, &req) {
		return
	}
	in, msg := req.toInput()
	if msg != "" {
		unprocessable(w, msg)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(h.defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}

	k, err := h.kasir.Create(r.Context(), in, string(hash))
	if errors.Is(err, store.ErrDuplicateEmail) {
		unprocessable(w, "Email sudah terdaftar")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, k)
}

// Update: PUT /api/admin/kasir/{id}
func (h *AdminKasirHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Kasir tidak ditemukan")
		return
	}
	var req kasirRequest
	if !decodeOrFail(w, r, &req) {
		return
	}
	in, msg := req.toInput()
	if msg != "" {
		unprocessable(w, msg)
		return
	}

	k, err := h.kasir.Update(r.Context(), id, in)
	if errors.Is(err, store.ErrDuplicateEmail) {
		unprocessable(w, "Email sudah terdaftar")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		respondNotFound(w, "Kasir tidak ditemukan")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, k)
}

// Delete: DELETE /api/admin/kasir/{id}
func (h *AdminKasirHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Kasir tidak ditemukan")
		return
	}
	err := h.kasir.Delete(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respondNotFound(w, "Kasir tidak ditemukan")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "Kasir berhasil dihapus"})
}
