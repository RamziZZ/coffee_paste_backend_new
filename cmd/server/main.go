package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"coffeepaste/internal/config"
	"coffeepaste/internal/database"
	"coffeepaste/internal/googleauth"
	"coffeepaste/internal/handler"
	"coffeepaste/internal/httpx"
	"coffeepaste/internal/mailer"
	"coffeepaste/internal/middleware"
	"coffeepaste/internal/midtrans"
	"coffeepaste/internal/push"
	"coffeepaste/internal/realtime"
	"coffeepaste/internal/store"

	_ "github.com/go-sql-driver/mysql"
)

// midtransReconcileEvery = jeda pengecekan status tagihan QRIS pending ke Midtrans.
const midtransReconcileEvery = 3 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("konfigurasi tidak valid: %v", err)
	}

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("gagal terhubung ke database: %v", err)
	}
	defer db.Close()

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("PERHATIAN : gagal membuat folder upload: %v", err)
	}

	// Store
	users := &store.UserStore{DB: db}
	menu := &store.MenuStore{DB: db}
	reviews := &store.ReviewStore{DB: db}
	kasir := &store.KasirStore{DB: db}
	customers := &store.CustomerStore{DB: db}
	expenses := &store.ExpenseStore{DB: db}
	orders := &store.OrderStore{DB: db}
	stock := &store.StockStore{DB: db}

	mail := mailer.Choose(cfg.BrevoAPIKey, cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	switch mail.(type) {
	case *mailer.BrevoMailer:
		log.Println("Email: Brevo API aktif")
	case *mailer.SMTPMailer:
		log.Println("Email: SMTP aktif")
	default:
		log.Println("Email: BREVO_API_KEY dan SMTP_HOST kosong, email hanya dicetak ke log")
	}

	// Harga Large per menu (kolom menu_items.price_large). Dibuat otomatis bila belum ada; minuman lama
	// diisi Regular + 4.000. Wajib berhasil karena semua query menu membaca kolom ini.
	if err := menu.EnsurePriceSchema(context.Background()); err != nil {
		log.Fatalf("gagal menyiapkan kolom price_large (jalankan deploy/migrations/2026-11-01_menu_price_large.sql): %v", err)
	}

	// Tabel OTP lupa password (aplikasi mobile). Gagal membuat tabel tidak mematikan server;
	// fitur OTP saja yang error sampai tabel dibuat (lihat deploy/migrations).
	if err := users.EnsurePasswordOTPSchema(context.Background()); err != nil {
		log.Printf("PERHATIAN: gagal menyiapkan tabel password_reset_otps: %v", err)
	}

	// Tabel laporan stok habis (kasir melapor, admin memantau). Dibuat otomatis bila belum ada.
	if err := stock.EnsureSchema(context.Background()); err != nil {
		log.Printf("PERHATIAN: gagal menyiapkan tabel stock_shortages: %v", err)
	}

	// Siaran realtime (SSE): pesanan baru/berubah langsung sampai ke web kasir dan aplikasi mobile.
	hub := realtime.NewHub()

	// Push notification (FCM): sampai ke HP walau aplikasi tertutup.
	var orderPush *handler.OrderPusher
	if cfg.FirebaseCredentials == "" {
		log.Println("Push FCM nonaktif: FIREBASE_CREDENTIALS_JSON kosong (notifikasi hanya lewat SSE saat aplikasi terbuka)")
	} else if sender, err := push.ParseCredentials(cfg.FirebaseCredentials); err != nil {
		log.Printf("PERHATIAN: push FCM nonaktif, kredensial Firebase tidak valid: %v", err)
	} else {
		orderPush = handler.NewOrderPusher(push.NewService(sender, users), orders)
		log.Println("Push FCM aktif")
	}

	// Handler
	authH := handler.NewAuthHandler(users, cfg.JWTSecret, cfg.JWTTTL)
	if len(cfg.GoogleClientIDs) > 0 {
		authH.WithGoogle(googleauth.New(cfg.GoogleClientIDs))
	}
	publicH := handler.NewPublicHandler(menu, reviews)
	adminMenuH := handler.NewAdminMenuHandler(menu)
	adminKasirH := handler.NewAdminKasirHandler(kasir, cfg.DefaultKasirPassword)
	adminCustomerH := handler.NewAdminCustomerHandler(customers)
	adminExpenseH := handler.NewAdminExpenseHandler(expenses)
	adminReviewH := handler.NewAdminReviewHandler(reviews)
	adminProfileH := handler.NewAdminProfileHandler(users, mail, cfg.PublicURL).WithTokens(cfg.JWTSecret, cfg.JWTTTL)
	uploadH := handler.NewUploadHandler(cfg.UploadDir, cfg.PublicURL)
	kasirMenuH := handler.NewKasirMenuHandler(menu)
	stockH := handler.NewStockHandler(stock)
	kasirOrderH := handler.NewKasirOrderHandler(orders).WithHub(hub).WithPush(orderPush)
	kasirProfileH := handler.NewKasirProfileHandler(users, mail).WithTokens(cfg.JWTSecret, cfg.JWTTTL)
	adminOrderH := handler.NewAdminOrderHandler(orders)
	resetH := handler.NewPasswordResetHandler(users, mail, cfg.FrontendURL)
	otpH := handler.NewPasswordOTPHandler(users, mail) // lupa password via OTP (aplikasi mobile)
	mobileH := handler.NewMobileHandler(reviews, menu, orders, users).WithQRIS(cfg.QRISStatic).WithHub(hub)

	// Klien Midtrans (nil jika MIDTRANS_SERVER_KEY kosong).
	// Catatan: sesuaikan tipe di bawah dengan nilai kembalian midtrans.New di internal/midtrans.
	var midtransClient *midtrans.Client
	if cfg.MidtransServerKey != "" {
		midtransClient = midtrans.New(cfg.MidtransServerKey, cfg.MidtransProduction, cfg.MidtransNotifyURL)
		mobileH.WithGateway(midtransClient)
		mode := "sandbox"
		if cfg.MidtransProduction {
			mode = "PRODUKSI"
		}
		log.Printf("Midtrans QRIS aktif (%s)", mode)
		// Kunci sandbox berawalan "SB-"; kunci produksi tanpa awalan itu. Salah pasangan
		// membuat Midtrans menolak (401) sehingga QR tidak pernah muncul.
		isSandboxKey := strings.HasPrefix(cfg.MidtransServerKey, "SB-")
		if !cfg.MidtransProduction && !isSandboxKey {
			log.Printf("PERHATIAN: MIDTRANS_IS_PRODUCTION=false tetapi MIDTRANS_SERVER_KEY bukan kunci sandbox (tidak berawalan SB-). Midtrans akan menolak dengan 401 dan QR tidak tampil.")
		}
		if cfg.MidtransProduction && isSandboxKey {
			log.Printf("PERHATIAN: MIDTRANS_IS_PRODUCTION=true tetapi MIDTRANS_SERVER_KEY adalah kunci sandbox (SB-). Midtrans akan menolak dengan 401.")
		}
	}
	midtransH := handler.NewMidtransHandler(orders, cfg.MidtransServerKey).WithHub(hub).WithPush(orderPush)
	if midtransClient != nil {
		// Cadangan webhook: status tagihan ditanyakan langsung ke Midtrans, jadi pesanan tetap
		// menjadi lunas walau Payment Notification URL belum diisi atau server di localhost.
		midtransH.WithStatusChecker(midtransClient)
	}
	paymentCallbackH := handler.NewPaymentCallbackHandler(orders, cfg.QRISCallbackSecret).WithHub(hub).WithPush(orderPush)
	customerProfileH := handler.NewCustomerProfileHandler(users, cfg.JWTSecret, cfg.JWTTTL).WithMailer(mail)

	// Middleware
	authenticate := middleware.Authenticate(cfg.JWTSecret, users)
	requireAdmin := middleware.RequireRole("admin")
	requireKasir := middleware.RequireRole("kasir")
	requireCustomer := middleware.RequireRole("customer")

	mux := http.NewServeMux()

	// admin mendaftarkan rute yang hanya boleh diakses role admin.
	admin := func(pattern string, h http.HandlerFunc) {
		mux.Handle(pattern, middleware.Chain(h, authenticate, requireAdmin))
	}

	// kasir mendaftarkan rute yang hanya boleh diakses role kasir.
	kasirRoute := func(pattern string, h http.HandlerFunc) {
		mux.Handle(pattern, middleware.Chain(h, authenticate, requireKasir))
	}

	// customer mendaftarkan rute untuk aplikasi mobile (role customer).
	customer := func(pattern string, h http.HandlerFunc) {
		mux.Handle(pattern, middleware.Chain(h, authenticate, requireCustomer))
	}

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			httpx.Error(w, http.StatusServiceUnavailable, "Database tidak dapat dijangkau")
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// File upload (foto menu)
	mux.Handle("GET /uploads/", handler.StaticUploads(cfg.UploadDir))

	// Publik
	mux.HandleFunc("GET /api/menu/featured", publicH.Featured)
	mux.HandleFunc("GET /api/menu/categories", publicH.Categories)
	mux.HandleFunc("GET /api/menu/filters", publicH.Filters)
	mux.HandleFunc("GET /api/reviews", publicH.Reviews)
	mux.HandleFunc("GET /api/menu", publicH.Menu) // daftar lengkap untuk aplikasi mobile

	// Autentikasi
	// Pembatasan per IP untuk endpoint yang rawan ditebak/disalahgunakan.
	loginLimit := middleware.RateLimit(10, 5*time.Minute)
	forgotLimit := middleware.RateLimit(5, 15*time.Minute)
	resetLimit := middleware.RateLimit(10, 15*time.Minute)
	registerLimit := middleware.RateLimit(10, 15*time.Minute)
	googleLimit := middleware.RateLimit(10, 5*time.Minute)

	mux.Handle("POST /api/auth/login", middleware.Chain(http.HandlerFunc(authH.Login), loginLimit))
	mux.Handle("POST /api/auth/google", middleware.Chain(http.HandlerFunc(authH.Google), googleLimit))
	mux.Handle("POST /api/auth/register", middleware.Chain(http.HandlerFunc(authH.Register), registerLimit))
	mux.Handle("POST /api/auth/logout", middleware.Chain(http.HandlerFunc(authH.Logout), authenticate))
	mux.Handle("GET /api/auth/me", middleware.Chain(http.HandlerFunc(authH.Me), authenticate))
	mux.Handle("POST /api/auth/forgot-password", middleware.Chain(http.HandlerFunc(resetH.Forgot), forgotLimit))
	mux.Handle("POST /api/auth/reset-password", middleware.Chain(http.HandlerFunc(resetH.Reset), resetLimit))
	mux.HandleFunc("GET /api/auth/email-change/verify", adminProfileH.VerifyEmailChange)

	// Lupa password via OTP 6 digit (aplikasi mobile). Endpoint web di atas tidak berubah.
	otpRequestLimit := middleware.RateLimit(5, 15*time.Minute)
	otpVerifyLimit := middleware.RateLimit(10, 15*time.Minute)
	otpResetLimit := middleware.RateLimit(10, 15*time.Minute)
	mux.Handle("POST /api/auth/forgot-password/otp", middleware.Chain(http.HandlerFunc(otpH.Request), otpRequestLimit))
	mux.Handle("POST /api/auth/forgot-password/otp/verify", middleware.Chain(http.HandlerFunc(otpH.Verify), otpVerifyLimit))
	mux.Handle("POST /api/auth/reset-password/otp", middleware.Chain(http.HandlerFunc(otpH.Reset), otpResetLimit))

	// Admin: menu
	admin("GET /api/admin/menu/categories", adminMenuH.Categories)
	admin("GET /api/admin/menu", adminMenuH.List)
	admin("POST /api/admin/menu", adminMenuH.Create)
	admin("PUT /api/admin/menu/{id}", adminMenuH.Update)
	admin("DELETE /api/admin/menu/{id}", adminMenuH.Delete)
	admin("POST /api/admin/upload", uploadH.Menu)

	// Admin: kasir, customer, pengeluaran
	admin("GET /api/admin/kasir", adminKasirH.List)
	admin("POST /api/admin/kasir", adminKasirH.Create)
	admin("PUT /api/admin/kasir/{id}", adminKasirH.Update)
	admin("DELETE /api/admin/kasir/{id}", adminKasirH.Delete)
	admin("GET /api/admin/customer", adminCustomerH.List)
	admin("GET /api/admin/pengeluaran", adminExpenseH.List)
	admin("POST /api/admin/pengeluaran", adminExpenseH.Create)
	admin("PUT /api/admin/pengeluaran/{id}", adminExpenseH.Update)
	admin("DELETE /api/admin/pengeluaran/{id}", adminExpenseH.Delete)

	// Admin: stok habis (laporan dari kasir)
	admin("GET /api/admin/stok-habis", stockH.AdminList)
	admin("PUT /api/admin/stok-habis/{id}/status", stockH.AdminSetStatus)

	// Admin: pesanan
	admin("GET /api/admin/pesanan", adminOrderH.List)

	// Admin: profil
	admin("GET /api/admin/profile", adminProfileH.Get)
	admin("PUT /api/admin/profile", adminProfileH.UpdateName)
	admin("POST /api/admin/profile/email", adminProfileH.RequestEmailChange)
	admin("DELETE /api/admin/profile/email", adminProfileH.CancelEmailChange)
	admin("PUT /api/admin/profile/password", adminProfileH.ChangePassword)

	// Admin: moderasi review (opsional)
	admin("GET /api/admin/reviews", adminReviewH.List)
	admin("PUT /api/admin/reviews/{id}", adminReviewH.SetPublished)

	// Kasir: menu
	kasirRoute("GET /api/kasir/menu", kasirMenuH.List)
	kasirRoute("PUT /api/kasir/menu/{id}/status", kasirMenuH.SetStatus)

	// Kasir: lapor stok habis
	kasirRoute("GET /api/kasir/stok-habis", stockH.KasirList)
	kasirRoute("POST /api/kasir/stok-habis", stockH.KasirCreate)
	kasirRoute("DELETE /api/kasir/stok-habis/{id}", stockH.KasirDelete)

	// Kasir: pesanan
	kasirRoute("GET /api/kasir/pesanan", kasirOrderH.List)
	kasirRoute("GET /api/kasir/pesanan/stream", kasirOrderH.Stream) // SSE: pesanan baru/berubah
	kasirRoute("POST /api/kasir/pesanan", kasirOrderH.Create)
	kasirRoute("PUT /api/kasir/pesanan/{id}", kasirOrderH.Update)

	// Kasir: profil
	kasirRoute("GET /api/kasir/profile", kasirProfileH.GetProfile)
	kasirRoute("PUT /api/kasir/profile", kasirProfileH.UpdateProfile)
	kasirRoute("POST /api/kasir/profile/email", kasirProfileH.RequestEmailChange)
	kasirRoute("POST /api/kasir/profile/email/verify", kasirProfileH.VerifyEmailChange)
	kasirRoute("PUT /api/kasir/profile/password", kasirProfileH.ChangePassword)

	// Mobile (customer)
	customer("POST /api/reviews", mobileH.CreateReview)
	customer("GET /api/favorites", mobileH.Favorites)
	customer("POST /api/favorites", mobileH.AddFavorite)
	customer("DELETE /api/favorites/{menuItemId}", mobileH.RemoveFavorite)
	customer("GET /api/profile", customerProfileH.Get)
	customer("PUT /api/profile", customerProfileH.Update)
	customer("PUT /api/profile/password", customerProfileH.ChangePassword)
	// Ganti email: kode 6 digit dikirim ke email baru. Permintaan dibatasi per IP agar tidak dipakai mengirim spam.
	emailChangeLimit := middleware.RateLimit(5, 15*time.Minute)
	mux.Handle("POST /api/profile/email", middleware.Chain(http.HandlerFunc(customerProfileH.RequestEmailChange), authenticate, requireCustomer, emailChangeLimit))
	customer("POST /api/profile/email/verify", customerProfileH.VerifyEmailChange)
	customer("DELETE /api/profile/email", customerProfileH.CancelEmailChange)
	customer("GET /api/orders", mobileH.Orders)
	customer("GET /api/orders/stream", mobileH.Stream) // SSE: perubahan pesanan milik customer
	customer("POST /api/orders", mobileH.CreateOrder)
	customer("POST /api/orders/{id}/cancel", mobileH.CancelOrder)
	customer("POST /api/orders/{id}/reviews", mobileH.ReviewOrder)
	customer("POST /api/devices", mobileH.SaveDevice)
	customer("DELETE /api/devices", mobileH.RemoveDevice)

	// Callback payment gateway (publik, dilindungi header rahasia)
	mux.HandleFunc("POST /api/payments/qris/callback", paymentCallbackH.QRIS)
	mux.Handle("POST /api/payments/midtrans/notification", middleware.Chain(http.HandlerFunc(midtransH.Notification), middleware.RateLimit(120, time.Minute)))

	// Rute tak dikenal -> JSON 404
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusNotFound, "Endpoint tidak ditemukan")
	})

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: middleware.Chain(mux,
			middleware.Recover,
			middleware.Logger,
			middleware.CORS(cfg.AllowedOrigins),
		),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second, // cukup untuk upload foto 5 MB
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if midtransClient != nil {
		go midtransH.RunReconciler(ctx, midtransReconcileEvery)
	}

	go func() {
		log.Printf("Server berjalan di http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server berhenti: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("mematikan server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown tidak bersih: %v", err)
	}
}
