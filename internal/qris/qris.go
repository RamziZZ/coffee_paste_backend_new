// Package qris mengubah kode QRIS statis milik merchant menjadi QRIS dinamis yang sudah
// berisi nominal tagihan, sehingga customer cukup memindai dan tidak perlu mengetik jumlah.
//
// Format mengikuti EMV QR Code (TLV): setiap elemen = ID 2 digit + panjang 2 digit + nilai.
// Perubahan yang dilakukan pada kode statis:
//   - tag 01 (point of initiation) dari "11" (statis) menjadi "12" (dinamis)
//   - tag 54 (nominal transaksi) disisipkan
//   - tag 63 (CRC16-CCITT) dihitung ulang
package qris

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type field struct {
	id    string
	value string
}

// Dynamic membuat QRIS dinamis dari kode statis merchant dan nominal rupiah (> 0).
func Dynamic(static string, amount int64) (string, error) {
	if amount <= 0 {
		return "", errors.New("nominal QRIS harus lebih dari 0")
	}
	fields, err := parse(strings.TrimSpace(static))
	if err != nil {
		return "", err
	}

	var out []field
	hasInitiation := false
	for _, f := range fields {
		switch f.id {
		case "63": // CRC lama dibuang, dihitung ulang di bawah
		case "54": // nominal lama (kalau ada) diganti
		case "01":
			hasInitiation = true
			out = append(out, field{"01", "12"})
		default:
			out = append(out, f)
		}
	}
	if !hasInitiation {
		return "", errors.New("kode QRIS tidak memiliki tag 01")
	}

	// Tag 54 harus berada setelah tag 53 (mata uang) dan sebelum tag 55 ke atas.
	amountField := field{"54", strconv.FormatInt(amount, 10)}
	inserted := false
	final := make([]field, 0, len(out)+1)
	for _, f := range out {
		if !inserted && f.id > "54" {
			final = append(final, amountField)
			inserted = true
		}
		final = append(final, f)
	}
	if !inserted {
		final = append(final, amountField)
	}

	var b strings.Builder
	for _, f := range final {
		if len(f.value) > 99 {
			return "", fmt.Errorf("nilai tag %s terlalu panjang", f.id)
		}
		fmt.Fprintf(&b, "%s%02d%s", f.id, len(f.value), f.value)
	}
	b.WriteString("6304")
	body := b.String()
	return fmt.Sprintf("%s%04X", body, CRC16(body)), nil
}

// Valid memeriksa struktur TLV dan CRC sebuah kode QRIS.
func Valid(code string) bool {
	code = strings.TrimSpace(code)
	if len(code) < 8 || code[len(code)-8:len(code)-4] != "6304" {
		return false
	}
	if _, err := parse(code); err != nil {
		return false
	}
	want := fmt.Sprintf("%04X", CRC16(code[:len(code)-4]))
	return strings.EqualFold(code[len(code)-4:], want)
}

func parse(code string) ([]field, error) {
	if code == "" {
		return nil, errors.New("kode QRIS kosong")
	}
	var fields []field
	for i := 0; i < len(code); {
		if i+4 > len(code) {
			return nil, errors.New("kode QRIS terpotong")
		}
		id := code[i : i+2]
		n, err := strconv.Atoi(code[i+2 : i+4])
		if err != nil || n < 0 {
			return nil, fmt.Errorf("panjang tag %s tidak valid", id)
		}
		if i+4+n > len(code) {
			return nil, fmt.Errorf("nilai tag %s terpotong", id)
		}
		fields = append(fields, field{id, code[i+4 : i+4+n]})
		i += 4 + n
	}
	return fields, nil
}

// CRC16 = CRC-16/CCITT-FALSE (poly 0x1021, init 0xFFFF) atas seluruh teks sampai "6304".
func CRC16(s string) uint16 {
	crc := uint16(0xFFFF)
	for i := 0; i < len(s); i++ {
		crc ^= uint16(s[i]) << 8
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
