// Package handler adalah entry point Vercel Serverless Function.
// vercel.json mengarahkan seluruh /api/* ke fungsi ini; routing dilakukan
// oleh handlers.App.Routes.
package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"transcpg-x/backend/pkg/config"
	"transcpg-x/backend/pkg/db"
	"transcpg-x/backend/pkg/handlers"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

var (
	mu  sync.Mutex
	app http.Handler
)

// getApp membuat pool sekali per instance fungsi dan memakainya ulang
// antar-invocation. Bila gagal (mis. database sesaat tidak terjangkau),
// request berikutnya mencoba lagi.
func getApp(ctx context.Context) (http.Handler, error) {
	mu.Lock()
	defer mu.Unlock()
	if app != nil {
		return app, nil
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	cfg.DBMaxConns = 2
	pool, err := db.Open(ctx, cfg)
	if err != nil {
		return nil, err
	}
	app = handlers.New(cfg, store.New(pool)).Routes()
	return app, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	h, err := getApp(r.Context())
	if err != nil {
		slog.Error("inisialisasi gagal", "err", err)
		httpx.WriteError(w, r, httpx.NewError(http.StatusServiceUnavailable, "UNAVAILABLE", "layanan belum siap: "+setupHint(err)))
		return
	}
	h.ServeHTTP(w, r)
}

// setupHint menerjemahkan galat inisialisasi menjadi petunjuk pemasangan
// tanpa membocorkan rahasia (rincian lengkap tetap di log Vercel).
func setupHint(err error) string {
	var cfgErr *config.Error
	if errors.As(err, &cfgErr) {
		return cfgErr.Error() // hanya nama variabel yang kurang, aman ditampilkan
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "database_url tidak valid"):
		return "DATABASE_URL tidak valid — salin ulang URI Transaction pooler dari Supabase"
	case strings.Contains(msg, "password authentication failed"):
		return "kata sandi database salah — periksa bagian [YOUR-PASSWORD] di DATABASE_URL"
	case strings.Contains(msg, "tenant or user not found"), strings.Contains(msg, "tenant/user"):
		return "nama pengguna pooler salah — pakai URI Transaction pooler apa adanya (user postgres.<id-proyek>)"
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "lookup"):
		return "alamat host database tidak ditemukan — periksa DATABASE_URL"
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "i/o timeout"), strings.Contains(msg, "connection refused"):
		return "database tidak terjangkau — pakai Transaction pooler (port 6543), bukan Direct connection"
	case strings.Contains(msg, "prepared statement"):
		return "set DB_SIMPLE_PROTOCOL=true untuk pooler port 6543"
	}
	return "tidak dapat terhubung ke database — lihat Logs di Vercel"
}
