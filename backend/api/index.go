// Package handler adalah entry point Vercel Serverless Function.
// vercel.json mengarahkan seluruh /api/* ke fungsi ini; routing dilakukan
// oleh handlers.App.Routes.
package handler

import (
	"context"
	"log/slog"
	"net/http"
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
		httpx.WriteError(w, r, httpx.NewError(http.StatusServiceUnavailable, "UNAVAILABLE", "layanan belum siap"))
		return
	}
	h.ServeHTTP(w, r)
}
