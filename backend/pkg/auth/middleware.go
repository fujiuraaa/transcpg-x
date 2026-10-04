package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
)

type ctxKey struct{}

// UserLoader memuat pengguna berdasarkan ID (diisi oleh store).
type UserLoader func(ctx context.Context, id int64) (domain.User, error)

// Require menolak request tanpa token valid atau dengan akun nonaktif.
func Require(secret []byte, load UserLoader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || raw == "" {
				httpx.WriteError(w, r, httpx.Unauthorized("silakan masuk terlebih dahulu"))
				return
			}
			id, ver, err := ParseToken(secret, raw)
			if err != nil {
				httpx.WriteError(w, r, httpx.Unauthorized("sesi tidak valid atau kedaluwarsa"))
				return
			}
			u, err := load(r.Context(), id)
			if errors.Is(err, pgx.ErrNoRows) || (err == nil && !u.IsActive) {
				httpx.WriteError(w, r, httpx.Unauthorized("akun tidak ditemukan atau nonaktif"))
				return
			}
			if err != nil {
				httpx.WriteError(w, r, err)
				return
			}
			if u.TokenVersion != ver {
				httpx.WriteError(w, r, httpx.Unauthorized("sesi berakhir karena kata sandi diganti atau Anda keluar dari semua perangkat — silakan masuk lagi"))
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
		})
	}
}

// RequireAdmin hanya meloloskan Admin RS, System Admin, dan Super Admin.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !CurrentUser(r).Role.IsAdmin() {
			httpx.WriteError(w, r, httpx.Forbidden("hanya Admin yang dapat membuka halaman ini"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CurrentUser mengembalikan pengguna yang dimuat Require.
func CurrentUser(r *http.Request) domain.User {
	u, _ := r.Context().Value(ctxKey{}).(domain.User)
	return u
}
