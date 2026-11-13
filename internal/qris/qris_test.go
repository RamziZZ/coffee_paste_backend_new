package qris

import (
	"fmt"
	"strings"
	"testing"
)

// staticCode merakit QRIS statis contoh (bukan merchant sungguhan) lengkap dengan CRC.
func staticCode() string {
	body := "000201" + "010211" +
		"2644" + "0014ID.CO.QRIS.WWW" + "0215ID1234567890123" + "0303UMI" + // tag 26 (44 karakter)
		"5204" + "5812" + // MCC
		"5303" + "360" + // mata uang IDR
		"5802" + "ID" +
		"5911" + "COFFEEPASTE" +
		"6006" + "JEMBER" +
		"6304"
	return fmt.Sprintf("%s%04X", body, CRC16(body))
}

func TestCRC16KnownVector(t *testing.T) {
	// Vektor uji standar CRC-16/CCITT-FALSE.
	if got := CRC16("123456789"); got != 0x29B1 {
		t.Fatalf("CRC16 = %04X, want 29B1", got)
	}
}

func TestStaticFixtureIsValid(t *testing.T) {
	if !Valid(staticCode()) {
		t.Fatalf("fixture statis tidak valid: %s", staticCode())
	}
}

func TestDynamicSetsAmountAndInitiation(t *testing.T) {
	code, err := Dynamic(staticCode(), 55500)
	if err != nil {
		t.Fatal(err)
	}
	if !Valid(code) {
		t.Fatalf("hasil tidak valid: %s", code)
	}
	if !strings.Contains(code, "010212") {
		t.Errorf("tag 01 harus 12 (dinamis): %s", code)
	}
	if strings.Contains(code, "010211") {
		t.Errorf("tag 01 statis masih ada: %s", code)
	}
	if !strings.Contains(code, "540555500") {
		t.Errorf("nominal 55500 tidak ada: %s", code)
	}
	// Tag 54 harus muncul setelah mata uang (53) dan sebelum negara (58).
	i53, i54, i58 := strings.Index(code, "5303360"), strings.Index(code, "540555500"), strings.Index(code, "5802ID")
	if !(i53 < i54 && i54 < i58) {
		t.Errorf("urutan tag salah: 53@%d 54@%d 58@%d", i53, i54, i58)
	}
}

func TestDynamicReplacesExistingAmount(t *testing.T) {
	first, err := Dynamic(staticCode(), 10000)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Dynamic(first, 25000)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(second, "540510000") || !strings.Contains(second, "540525000") {
		t.Errorf("nominal tidak diganti: %s", second)
	}
	if strings.Count(second, "6304") != 1 {
		t.Errorf("harus tepat satu CRC: %s", second)
	}
}

func TestDynamicRejectsBadInput(t *testing.T) {
	if _, err := Dynamic(staticCode(), 0); err == nil {
		t.Error("nominal 0 harus ditolak")
	}
	if _, err := Dynamic("", 1000); err == nil {
		t.Error("kode kosong harus ditolak")
	}
	if _, err := Dynamic("0002", 1000); err == nil {
		t.Error("kode terpotong harus ditolak")
	}
	if _, err := Dynamic("5802ID6304ABCD", 1000); err == nil {
		t.Error("kode tanpa tag 01 harus ditolak")
	}
}

func TestValidDetectsCorruption(t *testing.T) {
	code := staticCode()
	bad := code[:20] + "X" + code[21:]
	if Valid(bad) {
		t.Error("kode rusak dianggap valid")
	}
}
