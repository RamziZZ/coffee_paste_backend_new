package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"coffeepaste/internal/qris"
)

type Config struct {
	Port           string
	PublicURL      string // URL publik backend, dipakai untuk tautan email dan URL foto
	DBHost         string
	DBPort         string
	DBUser         string
	DBPass         string
	DBName         string
	JWTSecret      []byte
	JWTTTL         time.Duration
	AllowedOrigins []string
	FrontendURL    string // dipakai untuk tautan reset password

	// GoogleClientIDs: OAuth client ID (tipe Web) yang diterima sebagai audience ID token
	// Google dari aplikasi mobile. Kosong = POST /api/auth/google nonaktif.
	GoogleClientIDs []string

	// Midtrans (QRIS dinamis + notifikasi otomatis). Server key kosong = nonaktif.
	MidtransServerKey  string
	MidtransProduction bool   // false = sandbox
	MidtransNotifyURL  string // opsional; menimpa Payment Notification URL di dashboard

	UploadDir            string
	DefaultKasirPassword string

	// QRISStatic: kode QRIS statis milik merchant (teks hasil baca QR). Dipakai untuk membuat
	// QRIS dinamis berisi nominal. Kosong = pembayaran QRIS dari aplikasi tidak menampilkan QR.
	QRISStatic string
	// QRISCallbackSecret: rahasia untuk endpoint callback pembayaran QRIS dari payment gateway.
	// Kosong = endpoint callback nonaktif.
	QRISCallbackSecret string

	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// BrevoAPIKey: API key Brevo untuk kirim email lewat HTTPS (hosting yang memblokir SMTP).
	// Bila diisi, dipakai menggantikan SMTP; alamat pengirim tetap dari SMTP_FROM.
	BrevoAPIKey string

	// FirebaseCredentials: isi JSON service account Firebase (atau base64-nya) untuk push
	// notification FCM. Kosong = push nonaktif.
	FirebaseCredentials string
}

// Load membaca .env (jika ada) lalu variabel environment.
// Variabel environment yang sudah di-set tidak ditimpa oleh .env.
func Load() (*Config, error) {
	loadDotEnv(".env")

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 16 {
		return nil, errors.New("JWT_SECRET wajib diisi (minimal 16 karakter)")
	}

	ttlHours, err := strconv.Atoi(get("JWT_TTL_HOURS", "24"))
	if err != nil || ttlHours <= 0 {
		return nil, errors.New("JWT_TTL_HOURS harus bilangan bulat positif")
	}

	var origins []string
	for _, o := range strings.Split(get("CORS_ORIGINS", "http://localhost:5173"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}

	if len(origins) == 0 {
		origins = []string{"http://localhost:5173"}
	}

	var googleIDs []string
	for _, id := range strings.Split(os.Getenv("GOOGLE_CLIENT_IDS"), ",") {
		if id = strings.TrimSpace(id); id != "" {
			googleIDs = append(googleIDs, id)
		}
	}

	midtransKey := strings.TrimSpace(os.Getenv("MIDTRANS_SERVER_KEY"))
	midtransProd := strings.EqualFold(strings.TrimSpace(os.Getenv("MIDTRANS_IS_PRODUCTION")), "true")
	port := get("PORT", "8000")

	qrisStatic := strings.TrimSpace(os.Getenv("QRIS_STATIC"))
	if qrisStatic != "" && !qris.Valid(qrisStatic) {
		return nil, errors.New("QRIS_STATIC tidak valid (struktur atau CRC salah); salin ulang teks hasil baca QRIS merchant")
	}

	firebase := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_JSON"))
	if path := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_FILE")); firebase == "" && path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("FIREBASE_CREDENTIALS_FILE tidak terbaca: %w", err)
		}
		firebase = string(b)
	}

	return &Config{
		Port:           port,
		PublicURL:      strings.TrimRight(get("PUBLIC_URL", "http://localhost:"+port), "/"),
		DBHost:         get("DB_HOST", "127.0.0.1"),
		DBPort:         get("DB_PORT", "3306"),
		DBUser:         get("DB_USER", "root"),
		DBPass:         os.Getenv("DB_PASS"),
		DBName:         get("DB_NAME", "coffee_paste"),
		JWTSecret:      []byte(secret),
		JWTTTL:         time.Duration(ttlHours) * time.Hour,
		AllowedOrigins: origins,
		FrontendURL:    strings.TrimRight(get("FRONTEND_URL", origins[0]), "/"),

		GoogleClientIDs: googleIDs,

		MidtransServerKey:  midtransKey,
		MidtransProduction: midtransProd,
		MidtransNotifyURL:  strings.TrimSpace(os.Getenv("MIDTRANS_NOTIFICATION_URL")),

		UploadDir:            get("UPLOAD_DIR", "./uploads"),
		DefaultKasirPassword: get("DEFAULT_KASIR_PASSWORD", "coffeepastekasir"),

		QRISStatic:         qrisStatic,
		QRISCallbackSecret: os.Getenv("QRIS_CALLBACK_SECRET"),

		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: get("SMTP_PORT", "587"),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		SMTPFrom: os.Getenv("SMTP_FROM"),

		BrevoAPIKey: strings.TrimSpace(os.Getenv("BREVO_API_KEY")),

		FirebaseCredentials: firebase,
	}, nil
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if key != "" && os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}