package handler

import (
	"context"
	"errors"
	"strings"
	"testing"

	"coffeepaste/internal/store"
)

func TestChargeErrorMessage(t *testing.T) {
	cases := map[string]string{
		"midtrans: Access denied due to unauthorized transaction (401)": "MIDTRANS_SERVER_KEY",
		"midtrans: Payment channel is not activated. (402)":             "belum aktif",
		"midtrans: Post https://api.sandbox.midtrans.com: i/o timeout":  "internet",
		"midtrans: sesuatu yang lain (500)":                             "sesuatu yang lain (500)",
	}
	for in, want := range cases {
		if got := chargeErrorMessage(errors.New(in)); !strings.Contains(got, want) {
			t.Errorf("chargeErrorMessage(%q) = %q, harus memuat %q", in, got, want)
		}
	}
}

func TestMobileOrderCtxQRISError(t *testing.T) {
	o := store.Order{
		ID: "#CP00001", Status: "menunggu", PaymentMethod: "QRIS",
		Items: []store.OrderItem{{Price: 50000, Qty: 1}},
	}

	// Tanpa gateway dan tanpa QRIS statis: pesanan tetap terkirim, disertai alasan.
	m := NewMobileHandler(nil, nil, nil, nil).mobileOrderCtx(context.Background(), o)
	if m.QRISPayload != "" || m.QRISError == "" {
		t.Errorf("payload=%q error=%q: harus tanpa QR dan ada alasan", m.QRISPayload, m.QRISError)
	}

	// Pesanan tunai tidak perlu QR maupun alasan.
	cash := o
	cash.PaymentMethod = "Tunai (Cash)"
	if m := NewMobileHandler(nil, nil, nil, nil).mobileOrderCtx(context.Background(), cash); m.QRISError != "" {
		t.Errorf("pesanan tunai tidak boleh punya qrisError, dapat %q", m.QRISError)
	}

	// Dengan QRIS statis: QR dibuat dan tidak ada error.
	withStatic := NewMobileHandler(nil, nil, nil, nil).WithQRIS(testStaticQRIS())
	if m := withStatic.mobileOrderCtx(context.Background(), o); m.QRISPayload == "" || m.QRISError != "" {
		t.Errorf("payload=%q error=%q: harus ada QR tanpa error", m.QRISPayload, m.QRISError)
	}
}
