package handler

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"coffeepaste/internal/httpx"
)

const maxUploadBytes = 5 << 20 // 5 MB

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type UploadHandler struct {
	dir       string
	publicURL string
}

func NewUploadHandler(dir, publicURL string) *UploadHandler {
	return &UploadHandler{dir: dir, publicURL: publicURL}
}

// Menu: POST /api/admin/upload (multipart/form-data, field "file")
func (h *UploadHandler) Menu(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+(1<<20))
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		unprocessable(w, "Gagal membaca unggahan. Ukuran foto maksimal 5 MB")
		return
	}
	defer r.MultipartForm.RemoveAll()

	var file multipart.File
	var hdr *multipart.FileHeader
	for _, field := range []string{"file", "image", "photo", "foto"} {
		f, hd, err := r.FormFile(field)
		if err == nil {
			file, hdr = f, hd
			break
		}
	}
	if file == nil {
		unprocessable(w, "File foto tidak ditemukan (gunakan field \"file\")")
		return
	}
	defer file.Close()

	if hdr.Size > maxUploadBytes {
		unprocessable(w, "Ukuran foto maksimal 5 MB")
		return
	}

	// Jenis file ditentukan dari isinya, bukan dari nama atau header kiriman klien.
	head := make([]byte, 512)
	n, _ := io.ReadFull(file, head)
	ext, ok := allowedImageTypes[http.DetectContentType(head[:n])]
	if !ok {
		unprocessable(w, "Format foto harus JPG, PNG, atau WebP")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		httpx.ServerError(w, err)
		return
	}

	rnd := make([]byte, 16)
	if _, err := rand.Read(rnd); err != nil {
		httpx.ServerError(w, err)
		return
	}
	name := hex.EncodeToString(rnd) + ext

	dir := filepath.Join(h.dir, "menu")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		httpx.ServerError(w, err)
		return
	}
	dst, err := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		httpx.ServerError(w, err)
		return
	}
	if err := dst.Close(); err != nil {
		httpx.ServerError(w, err)
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]string{"url": h.publicURL + "/uploads/menu/" + name})
}

// noListFS menyembunyikan isi folder agar daftar file upload tidak bisa ditelusuri.
type noListFS struct{ fs http.FileSystem }

func (n noListFS) Open(name string) (http.File, error) {
	f, err := n.fs.Open(name)
	if err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if st.IsDir() {
		f.Close()
		return nil, os.ErrNotExist
	}
	return f, nil
}

// StaticUploads melayani file di folder upload pada path /uploads/...
func StaticUploads(dir string) http.Handler {
	fs := http.StripPrefix("/uploads/", http.FileServer(noListFS{http.Dir(dir)}))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Nama file berupa hash acak, isinya tidak berubah: boleh disimpan lama
		// di cache HP supaya gambar menu tidak diunduh ulang.
		w.Header().Set("Cache-Control", "public, max-age=2592000, immutable")
		fs.ServeHTTP(w, r)
	})
}
