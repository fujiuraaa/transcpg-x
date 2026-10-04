package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// register: pendaftaran mandiri. Akun dibuat berstatus MENUNGGU (nonaktif)
// dan baru bisa dipakai setelah Admin RS menyetujui & menetapkan peran.
// Bisa dengan email + kata sandi, atau dengan akun Google (email diambil
// dari Google yang sudah terverifikasi).
func (a *App) register(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		FullName          string      `json:"full_name"`
		Email             string      `json:"email"`
		Phone             *string     `json:"phone"`
		RequestedRole     domain.Role `json:"requested_role"`
		Password          string      `json:"password"`
		GoogleAccessToken string      `json:"google_access_token"`
		Note              *string     `json:"note"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	in.FullName = strings.TrimSpace(in.FullName)
	email := strings.ToLower(strings.TrimSpace(in.Email))

	var password string
	if in.GoogleAccessToken != "" {
		gEmail, err := a.google.Verify(r.Context(), in.GoogleAccessToken)
		if err != nil {
			slog.Info("daftar Google ditolak", "err", err)
			return httpx.Unauthorized("verifikasi akun Google gagal — coba lagi")
		}
		email = gEmail
		// Akun Google tidak punya kata sandi; isi acak yang tidak bisa ditebak.
		password = randomSecret()
	} else {
		password = in.Password
		if len(password) < minPasswordLen {
			return httpx.Unprocessable("kata sandi minimal 8 karakter")
		}
	}

	switch {
	case in.FullName == "":
		return httpx.Unprocessable("nama lengkap wajib diisi")
	case !validEmail(email):
		return httpx.Unprocessable("format email tidak valid")
	case !in.RequestedRole.SelfRegistrable():
		return httpx.Unprocessable("pilih peran yang diajukan")
	}
	phone, err := optionalPhone(in.Phone)
	if err != nil {
		return err
	}
	if in.Note != nil {
		if n := strings.TrimSpace(*in.Note); n == "" {
			in.Note = nil
		} else if len(n) > 500 {
			return httpx.Unprocessable("keterangan maksimal 500 karakter")
		} else {
			in.Note = &n
		}
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	u, err := a.store.Register(r.Context(), store.NewRegistration{
		FullName: in.FullName, Email: email, Phone: phone, RequestedRole: in.RequestedRole,
		PasswordHash: hash, Note: in.Note,
	})
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23505" {
		// Jangan sebut kolom/constraint yang bentrok (mencegah menebak data).
		return httpx.Conflict("pendaftaran tidak dapat diproses — periksa kembali email dan nomor telepon, atau kosongkan nomor telepon")
	}
	if err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "AKUN", Action: "DAFTAR", Actor: u,
		TargetType: "USER", TargetID: u.ID, TargetLabel: u.FullName,
		Summary: "Mendaftar mandiri, mengajukan peran " + u.Role.Label(),
		Detail:  map[string]any{"email": u.Email, "google": in.GoogleAccessToken != ""}})
	return httpx.JSON(w, http.StatusCreated, map[string]any{
		"status":         u.RegistrationStatus,
		"email":          u.Email,
		"requested_role": u.Role,
		"message":        "Pendaftaran terkirim. Anda bisa masuk setelah Admin RS menyetujui.",
	})
}

// registrationRoles: pilihan peran di formulir daftar.
func (a *App) registrationRoles(w http.ResponseWriter, r *http.Request) error {
	type opt struct {
		Value domain.Role `json:"value"`
		Label string      `json:"label"`
	}
	var out []opt
	for _, ro := range domain.SelfRegistrableRoles() {
		out = append(out, opt{ro, ro.Label()})
	}
	return httpx.OK(w, out)
}

func randomSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
