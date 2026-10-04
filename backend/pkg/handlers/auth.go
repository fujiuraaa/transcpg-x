package handlers

import (
	"net/http"
	"strings"
	"time"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

func (a *App) routeAuth(m *http.ServeMux) {
	m.HandleFunc("GET /api/auth/me", h(a.me))
	m.HandleFunc("POST /api/auth/logout", h(a.logout))
	m.HandleFunc("POST /api/auth/logout-all", h(a.logoutAll))
}

type sessionUser struct {
	domain.User
	RoleLabel string `json:"role_label"`
	IsAdmin   bool   `json:"is_admin"`
}

func toSessionUser(u domain.User) sessionUser {
	return sessionUser{User: u, RoleLabel: u.Role.Label(), IsAdmin: u.Role.IsAdmin()}
}

func (a *App) login(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	invalid := httpx.Unauthorized("email atau kata sandi salah")
	email := strings.ToLower(strings.TrimSpace(in.Email))
	key := "email:" + email
	// Batas percobaan berlaku juga untuk email yang tidak terdaftar, supaya
	// respons tidak membedakan keduanya.
	if n, err := a.store.LoginFailures(r.Context(), key, loginLockWindow); err != nil {
		return err
	} else if n >= loginMaxFailures {
		return tooManyAttempts()
	}
	u, hash, err := a.store.GetUserForLogin(r.Context(), email)
	if store.IsNotFound(err) {
		auth.BurnPasswordCheck(in.Password) // samakan waktu respons
		_ = a.store.RecordLoginFailure(r.Context(), key)
		return invalid
	}
	if err != nil {
		return err
	}
	if !auth.CheckPassword(hash, in.Password) {
		_ = a.store.RecordLoginFailure(r.Context(), key)
		a.record(r, store.Activity{Category: "MASUK", Action: "MASUK_GAGAL", Actor: u,
			TargetType: "USER", TargetID: u.ID, TargetLabel: u.FullName,
			Summary: "Gagal masuk: kata sandi salah", Detail: map[string]any{"metode": "Email"}})
		return invalid
	}
	return a.issueSession(w, r, u, "Email")
}

// issueSession menerbitkan JWT untuk pengguna yang sudah terautentikasi
// (lewat password, Google, atau OTP).
func (a *App) issueSession(w http.ResponseWriter, r *http.Request, u domain.User, method string) error {
	switch u.RegistrationStatus {
	case domain.RegPending:
		return &httpx.Error{Status: http.StatusForbidden, Code: "MENUNGGU_PERSETUJUAN",
			Message: "pendaftaran Anda masih menunggu persetujuan Admin RS"}
	case domain.RegRejected:
		msg := "pendaftaran Anda ditolak Admin RS"
		if u.RejectionReason != nil {
			msg += ": " + *u.RejectionReason
		}
		return &httpx.Error{Status: http.StatusForbidden, Code: "PENDAFTARAN_DITOLAK",
			Message: msg + ". Anda boleh mendaftar ulang."}
	}
	if !u.IsActive {
		return httpx.Unauthorized("akun Anda nonaktif — hubungi Admin RS")
	}
	token, exp, err := auth.IssueToken(a.cfg.JWTSecret, u.ID, u.TokenVersion, a.cfg.JWTTTL)
	if err != nil {
		return err
	}
	if err := a.store.TouchLogin(r.Context(), u.ID); err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "MASUK", Action: "MASUK", Actor: u,
		TargetType: "USER", TargetID: u.ID, TargetLabel: u.FullName,
		Summary: "Masuk lewat " + method, Detail: map[string]any{"metode": method}})
	return httpx.OK(w, map[string]any{"token": token, "expires_at": exp.Format(time.RFC3339), "user": toSessionUser(u)})
}

// me dipanggil saat aplikasi dibuka: memastikan sesi masih sah dan
// mengembalikan angka antrean untuk lencana menu Approval.
func (a *App) me(w http.ResponseWriter, r *http.Request) error {
	u := auth.CurrentUser(r)
	n, err := a.store.CountActionable(r.Context(), u.Role)
	if err != nil {
		return err
	}
	pending := 0
	if u.Role.IsAdmin() {
		if pending, err = a.store.CountPendingRegistrations(r.Context(), adminScope(u)); err != nil {
			return err
		}
	}
	return httpx.OK(w, map[string]any{"user": toSessionUser(u), "approval_badge": n, "registration_badge": pending})
}

// logout: token bersifat stateless, jadi cukup dihapus di sisi klien.
func (a *App) logout(w http.ResponseWriter, r *http.Request) error {
	return httpx.NoContent(w)
}

// logoutAll mencabut semua sesi akun ini di semua perangkat.
func (a *App) logoutAll(w http.ResponseWriter, r *http.Request) error {
	u := auth.CurrentUser(r)
	if _, err := a.store.RevokeSessions(r.Context(), u.ID); err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "MASUK", Action: "KELUAR_SEMUA", TargetType: "USER", TargetID: u.ID,
		TargetLabel: u.FullName, Summary: "Keluar dari semua perangkat"})
	return httpx.NoContent(w)
}

const (
	loginMaxFailures = 5
	loginLockWindow  = 15 * time.Minute
)

func tooManyAttempts() error {
	return httpx.NewError(http.StatusTooManyRequests, "TERLALU_BANYAK_PERCOBAAN",
		"terlalu banyak percobaan gagal — coba lagi dalam 15 menit")
}
