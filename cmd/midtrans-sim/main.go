// Command midtrans-sim mengirim notifikasi pembayaran Midtrans palsu (bertanda tangan sah)
// ke backend lokal, untuk menguji alur "bayar -> pesanan otomatis lunas" tanpa Midtrans.
//
//	go run ./cmd/midtrans-sim -ref CP00001-1760000000 -amount 55500
//
// ref dan amount tertera di log server saat tagihan QRIS dibuat.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"coffeepaste/internal/config"
	"coffeepaste/internal/midtrans"
)

func main() {
	ref := flag.String("ref", "", "order_id Midtrans (mis. CP00001-1760000000), lihat log server")
	amount := flag.Int64("amount", 0, "nominal tagihan dalam rupiah")
	status := flag.String("status", "settlement", "settlement | expire | cancel | pending")
	url := flag.String("url", "", "alamat backend (default http://localhost:<PORT>)")
	flag.Parse()

	if *ref == "" || *amount <= 0 {
		flag.Usage()
		log.Fatal("-ref dan -amount wajib diisi")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("konfigurasi: %v", err)
	}
	if cfg.MidtransServerKey == "" {
		log.Fatal("MIDTRANS_SERVER_KEY belum diisi di .env")
	}
	target := *url
	if target == "" {
		target = "http://localhost:" + cfg.Port
	}

	gross := fmt.Sprintf("%d.00", *amount)
	payload := map[string]string{
		"order_id":           *ref,
		"status_code":        "200",
		"gross_amount":       gross,
		"transaction_status": *status,
		"fraud_status":       "accept",
		"payment_type":       "qris",
		"transaction_time":   time.Now().Format("2006-01-02 15:04:05"),
		"signature_key":      midtrans.Signature(*ref, "200", gross, cfg.MidtransServerKey),
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(target+"/api/payments/midtrans/notification", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("gagal menghubungi backend: %v", err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	fmt.Printf("HTTP %d\n%s\n", resp.StatusCode, out)
}
