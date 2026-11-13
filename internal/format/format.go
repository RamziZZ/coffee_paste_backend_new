package format

import (
	"strconv"
	"strings"
)

// Rupiah memformat bilangan bulat menjadi "Rp 18.000".
func Rupiah(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)

	var b strings.Builder
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(ch)
	}

	if neg {
		return "-Rp " + b.String()
	}
	return "Rp " + b.String()
}

// Slug mengubah teks menjadi slug huruf kecil bertanda hubung.
func Slug(s string) string {
	var b strings.Builder
	lastDash := true
	for _, ch := range strings.ToLower(s) {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			b.WriteRune(ch)
			lastDash = false
		} else if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// ISO adalah layout tanggal-waktu respons API, contoh 2026-10-03T14:30:00.
const ISO = "2006-01-02T15:04:05"
