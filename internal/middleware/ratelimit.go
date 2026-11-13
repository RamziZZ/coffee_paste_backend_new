package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"coffeepaste/internal/httpx"
)

type bucket struct {
	count int
	reset time.Time
}

// RateLimit membatasi jumlah request per IP: maksimal `max` dalam jendela `window`.
// Setiap pemanggilan membuat penghitung terpisah, jadi pasang satu instance per rute.
// Penghitung disimpan di memori (hilang saat server restart; cukup untuk satu instance).
func RateLimit(max int, window time.Duration) Middleware {
	var mu sync.Mutex
	hits := make(map[string]*bucket)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			now := time.Now()

			mu.Lock()
			if len(hits) > 10000 { // bersihkan entri kedaluwarsa agar memori tidak membengkak
				for k, b := range hits {
					if now.After(b.reset) {
						delete(hits, k)
					}
				}
			}
			b := hits[ip]
			if b == nil || now.After(b.reset) {
				b = &bucket{reset: now.Add(window)}
				hits[ip] = b
			}
			b.count++
			over := b.count > max
			retry := int(time.Until(b.reset).Seconds()) + 1
			mu.Unlock()

			if over {
				w.Header().Set("Retry-After", strconv.Itoa(retry))
				httpx.Error(w, http.StatusTooManyRequests, "Terlalu banyak percobaan. Coba lagi beberapa saat lagi")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP memakai entri paling kanan X-Forwarded-For (ditambahkan oleh proxy tepercaya,
// mis. Railway), bila tidak ada memakai alamat koneksi langsung.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
