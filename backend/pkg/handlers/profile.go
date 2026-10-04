package handlers

import (
	"net/http"
	"strings"
	"time"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Profil Saya: setiap pengguna boleh mengubah nama, nomor telepon, dan kata
// sandinya sendiri. Peran & status akun tetap hanya lewat Admin.
func (a *App) routeProfile(m *http.ServeMux) {
	m.HandleFunc("PATCH /api/auth/me", h(a.updateProfile))
	m.HandleFunc("POST /api/auth/password", h(a.changePassword))
	m.HandleFunc("GET /api/auth/me/activity", h(a.myActivity))
}

func (a *App) updateProfile(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		FullName *string `json:"full_name"`
		Phone    *string `json:"phone"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	me := auth.CurrentUser(r)
	var patch store.UserPatch
	if in.FullName != nil {
		name := strings.TrimSpace(*in.FullName)
		if name == "" {
			return httpx.Unprocessable("nama lengkap wajib diisi")
		}
		if len(name) > 120 {
			return httpx.Unprocessable("nama lengkap maksimal 120 karakter")
		}
		patch.FullName = &name
	}
	if in.Phone != nil {
		empty := ""
		patch.Phone = &empty
		if strings.TrimSpace(*in.Phone) != "" {
			p, err := optionalPhone(in.Phone)
			if err != nil {
				return err
			}
			patch.Phone = p
		}
	}
	u, err := a.store.UpdateUser(r.Context(), me.ID, patch)
	if err != nil {
		return err
	}
	if changes, detail := describeUserChange(me, u, false); len(changes) > 0 {
		a.record(r, store.Activity{Category: "AKUN", Action: "UBAH_PROFIL", TargetType: "USER", TargetID: u.ID,
			TargetLabel: u.FullName, Summary: "Mengubah profil sendiri: " + strings.ToLower(strings.Join(changes, "; ")), Detail: detail})
	}
	return httpx.OK(w, toSessionUser(u))
}

func (a *App) changePassword(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Current string `json:"current_password"`
		New     string `json:"new_password"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	me := auth.CurrentUser(r)
	_, hash, err := a.store.GetUserForLogin(r.Context(), me.Email)
	if err != nil {
		return err
	}
	switch {
	case !auth.CheckPassword(hash, in.Current):
		a.record(r, store.Activity{Category: "AKUN", Action: "GANTI_SANDI_GAGAL", TargetType: "USER", TargetID: me.ID,
			TargetLabel: me.FullName, Summary: "Gagal mengganti kata sandi: kata sandi saat ini salah"})
		return httpx.Unprocessable("kata sandi saat ini salah")
	case len(in.New) < minPasswordLen:
		return httpx.Unprocessable("kata sandi baru minimal 8 karakter")
	case len(in.New) > auth.MaxPasswordBytes:
		return httpx.Unprocessable("kata sandi baru maksimal 72 karakter")
	case in.New == in.Current:
		return httpx.Unprocessable("kata sandi baru harus berbeda dari yang lama")
	}
	newHash, err := auth.HashPassword(in.New)
	if err != nil {
		return err
	}
	// UpdateUser menaikkan versi sesi → semua sesi lain otomatis keluar.
	u, err := a.store.UpdateUser(r.Context(), me.ID, store.UserPatch{PasswordHash: &newHash})
	if err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "AKUN", Action: "GANTI_SANDI", TargetType: "USER", TargetID: me.ID,
		TargetLabel: me.FullName, Summary: "Mengganti kata sandi sendiri (sesi di perangkat lain dikeluarkan)"})
	// Perangkat ini tetap masuk dengan token baru.
	token, exp, err := auth.IssueToken(a.cfg.JWTSecret, u.ID, u.TokenVersion, a.cfg.JWTTTL)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"token": token, "expires_at": exp.Format(time.RFC3339)})
}

// myActivity: 20 kejadian terakhir oleh pengguna ini (termasuk masuk aplikasi).
func (a *App) myActivity(w http.ResponseWriter, r *http.Request) error {
	me := auth.CurrentUser(r)
	items, _, err := a.store.AuditEvents(r.Context(), store.AuditFilter{ActorID: &me.ID, Limit: 20})
	if err != nil {
		return err
	}
	return httpx.OK(w, items)
}
