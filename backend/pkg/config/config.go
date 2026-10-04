// Package config membaca konfigurasi backend dari environment variable.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL string
	// DBSimpleProtocol wajib true bila memakai pooler Supabase mode
	// "transaction" (port 6543) — pooler tidak mendukung prepared statement.
	DBSimpleProtocol bool
	DBMaxConns       int32
	JWTSecret        []byte
	JWTTTL           time.Duration
	AllowedOrigins   []string
	Addr             string
	// GoogleClientID: OAuth Client ID (Web) dari Google Cloud Console.
	// Kosong = tombol Google dinonaktifkan.
	GoogleClientID string
	// OTPSender: pengirim SMS untuk login telepon. "log" = tulis ke log
	// (hanya pengembangan). Kosong = login telepon dinonaktifkan.
	OTPSender string
	// Penyimpanan lampiran: "supabase" (produksi), "local" (dev), kosong = nonaktif.
	StorageDriver   string
	SupabaseURL     string
	SupabaseKey     string // service role key — hanya di server
	StorageBucket   string
	LocalStorageDir string
}

// Error adalah galat konfigurasi (variabel lingkungan kurang/salah). Pesannya
// hanya menyebut nama variabel, tidak pernah nilainya.
type Error struct{ msg string }

func (e *Error) Error() string { return e.msg }

func cfgErr(format string, a ...any) error { return &Error{msg: fmt.Sprintf(format, a...)} }

func Load() (Config, error) {
	c := Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		DBSimpleProtocol: os.Getenv("DB_SIMPLE_PROTOCOL") == "true",
		DBMaxConns:       4,
		JWTSecret:        []byte(os.Getenv("JWT_SECRET")),
		JWTTTL:           30 * 24 * time.Hour, // sesi bertahan sampai pengguna keluar
		AllowedOrigins:   splitList(os.Getenv("ALLOWED_ORIGINS")),
		Addr:             envOr("ADDR", ":8080"),
		GoogleClientID:   os.Getenv("GOOGLE_CLIENT_ID"),
		OTPSender:        os.Getenv("OTP_SENDER"),
		StorageDriver:    os.Getenv("STORAGE_DRIVER"),
		SupabaseURL:      os.Getenv("SUPABASE_URL"),
		SupabaseKey:      os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		StorageBucket:    envOr("STORAGE_BUCKET", "guideline-attachments"),
		LocalStorageDir:  envOr("LOCAL_STORAGE_DIR", "./data/uploads"),
	}
	switch c.StorageDriver {
	case "", "local":
	case "supabase":
		if c.SupabaseURL == "" || c.SupabaseKey == "" {
			return c, cfgErr("STORAGE_DRIVER=supabase memerlukan SUPABASE_URL dan SUPABASE_SERVICE_ROLE_KEY")
		}
	default:
		return c, cfgErr("STORAGE_DRIVER %q tidak dikenal (pilihan: supabase, local)", c.StorageDriver)
	}
	if c.OTPSender != "" && c.OTPSender != "log" {
		return c, cfgErr("OTP_SENDER %q tidak dikenal (pilihan: log)", c.OTPSender)
	}
	if v := os.Getenv("JWT_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return c, cfgErr("JWT_TTL tidak valid (contoh: 720h)")
		}
		c.JWTTTL = d
	}
	if c.DatabaseURL == "" {
		return c, cfgErr("DATABASE_URL wajib diisi")
	}
	if len(c.JWTSecret) < 32 {
		return c, cfgErr("JWT_SECRET wajib diisi, minimal 32 karakter")
	}
	return c, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
