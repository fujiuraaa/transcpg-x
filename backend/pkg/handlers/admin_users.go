package handlers

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

const minPasswordLen = 8

func (a *App) routeAdminUsers(m *http.ServeMux) {
	m.Handle("GET /api/admin/users", admin(a.listUsers))
	m.Handle("POST /api/admin/users", admin(a.createUser))
	m.Handle("PATCH /api/admin/users/{id}", admin(a.updateUser))
	m.Handle("DELETE /api/admin/users/{id}", admin(a.deleteUser))
}

// adminScope: Admin RS hanya mengelola akun di RS-nya sendiri; System/Super
// Admin semua RS (nil).
func adminScope(actor domain.User) *int64 {
	if actor.Role != domain.RoleAdminRS {
		return nil
	}
	if actor.HospitalID == nil {
		none := int64(-1)
		return &none
	}
	return actor.HospitalID
}

// inScope melaporkan apakah akun target berada dalam lingkup admin.
func inScope(actor, target domain.User) bool {
	scope := adminScope(actor)
	return scope == nil || (target.HospitalID != nil && *target.HospitalID == *scope)
}

// manageTarget memuat akun target dan memastikan admin boleh mengelolanya.
// Akun di luar lingkup dilaporkan "tidak ditemukan" agar tidak bisa ditebak.
func (a *App) manageTarget(r *http.Request, id int64) (domain.User, error) {
	actor := auth.CurrentUser(r)
	target, err := a.store.GetUser(r.Context(), id)
	if err != nil {
		return target, err
	}
	if !inScope(actor, target) {
		return target, httpx.NotFound("pengguna tidak ditemukan")
	}
	if !target.Role.ManageableBy(actor.Role) {
		return target, httpx.Forbidden("Anda tidak dapat mengelola akun dengan peran " + target.Role.Label())
	}
	return target, nil
}

func (a *App) listUsers(w http.ResponseWriter, r *http.Request) error {
	users, err := a.store.ListUsers(r.Context(), adminScope(auth.CurrentUser(r)))
	if err != nil {
		return err
	}
	out := make([]sessionUser, len(users))
	for i, u := range users {
		out[i] = toSessionUser(u)
	}
	return httpx.OK(w, out)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		FullName   string      `json:"full_name"`
		Email      string      `json:"email"`
		Phone      *string     `json:"phone"`
		Password   string      `json:"password"`
		Role       domain.Role `json:"role"`
		HospitalID *int64      `json:"hospital_id"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	actor := auth.CurrentUser(r)
	switch {
	case strings.TrimSpace(in.FullName) == "":
		return httpx.Unprocessable("nama lengkap wajib diisi")
	case !validEmail(in.Email):
		return httpx.Unprocessable("format email tidak valid")
	case len(in.Password) < minPasswordLen:
		return httpx.Unprocessable("password awal minimal 8 karakter")
	case !in.Role.AssignableBy(actor.Role):
		return httpx.Forbidden("peran tersebut tidak dapat Anda berikan")
	}
	if in.HospitalID == nil {
		in.HospitalID = actor.HospitalID
	}
	// Admin RS hanya boleh membuat akun di RS-nya sendiri.
	if scope := adminScope(actor); scope != nil && (in.HospitalID == nil || *in.HospitalID != *scope) {
		return httpx.Forbidden("Admin RS hanya dapat membuat akun untuk rumah sakitnya sendiri")
	}
	phone, err := optionalPhone(in.Phone)
	if err != nil {
		return err
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return err
	}
	u, err := a.store.CreateUser(r.Context(), store.NewUser{
		HospitalID: in.HospitalID, FullName: strings.TrimSpace(in.FullName),
		Email: strings.TrimSpace(in.Email), Phone: phone, PasswordHash: hash, Role: in.Role,
	})
	if err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "AKUN", Action: "BUAT", TargetType: "USER", TargetID: u.ID,
		TargetLabel: u.FullName, HospitalID: u.HospitalID,
		Summary: "Membuat akun dengan peran " + u.Role.Label(), Detail: map[string]any{"email": u.Email}})
	return httpx.JSON(w, http.StatusCreated, toSessionUser(u))
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		FullName *string      `json:"full_name"`
		Phone    *string      `json:"phone"`
		Role     *domain.Role `json:"role"`
		IsActive *bool        `json:"is_active"`
		Password *string      `json:"password"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	actor := auth.CurrentUser(r)
	if id == actor.ID && ((in.IsActive != nil && !*in.IsActive) || in.Role != nil) {
		return httpx.Unprocessable("Anda tidak dapat menonaktifkan atau mengubah peran akun sendiri")
	}
	if id == actor.ID && in.Password != nil {
		return httpx.Unprocessable("ganti kata sandi Anda sendiri lewat Profil Saya (perlu kata sandi lama)")
	}
	var target domain.User
	if id == actor.ID {
		// Mengubah nama/telepon sendiri tetap boleh lewat sini.
		if target, err = a.store.GetUser(r.Context(), id); err != nil {
			return err
		}
	} else if target, err = a.manageTarget(r, id); err != nil {
		return err
	}
	if in.Role != nil && !in.Role.AssignableBy(actor.Role) {
		return httpx.Forbidden("peran tersebut tidak dapat Anda berikan")
	}
	patch := store.UserPatch{FullName: in.FullName, Role: in.Role, IsActive: in.IsActive}
	if in.Phone != nil {
		// "" = hapus nomor; selain itu harus nomor yang valid.
		empty := ""
		patch.Phone = &empty
		if strings.TrimSpace(*in.Phone) != "" {
			if patch.Phone, err = optionalPhone(in.Phone); err != nil {
				return err
			}
		}
	}
	if in.Password != nil {
		if len(*in.Password) < minPasswordLen {
			return httpx.Unprocessable("password minimal 8 karakter")
		}
		hash, err := auth.HashPassword(*in.Password)
		if err != nil {
			return err
		}
		patch.PasswordHash = &hash
	}
	u, err := a.store.UpdateUser(r.Context(), id, patch)
	if err != nil {
		return err
	}
	if changes, detail := describeUserChange(target, u, in.Password != nil); len(changes) > 0 {
		a.record(r, store.Activity{Category: "AKUN", Action: userChangeAction(target, u), TargetType: "USER",
			TargetID: u.ID, TargetLabel: u.FullName, HospitalID: u.HospitalID,
			Summary: strings.Join(changes, "; "), Detail: detail})
	}
	return httpx.OK(w, toSessionUser(u))
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	if id == auth.CurrentUser(r).ID {
		return httpx.Unprocessable("Anda tidak dapat menghapus akun sendiri")
	}
	target, err := a.manageTarget(r, id)
	if err != nil {
		return err
	}
	// Pengguna yang sudah tercatat di riwayat pengesahan tidak bisa dihapus
	// (FK) — galat 422 dikembalikan; nonaktifkan akun sebagai gantinya.
	if err := a.store.DeleteUser(r.Context(), id); err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23503" {
			return httpx.Conflict("pengguna ini sudah tercatat di riwayat CP (pengesahan, acuan, atau rencana) sehingga tidak dapat dihapus — nonaktifkan akunnya sebagai gantinya")
		}
		return err
	}
	a.record(r, store.Activity{Category: "AKUN", Action: "HAPUS", TargetType: "USER", TargetID: target.ID,
		TargetLabel: target.FullName, HospitalID: target.HospitalID,
		Summary: "Menghapus akun " + target.Role.Label(), Detail: map[string]any{"email": target.Email}})
	return httpx.NoContent(w)
}

// describeUserChange merangkum perubahan akun untuk jejak audit
// (kata sandi tidak pernah dicatat, hanya fakta bahwa ia diganti).
func describeUserChange(before, after domain.User, passwordReset bool) ([]string, map[string]any) {
	var out []string
	detail := map[string]any{}
	if before.Role != after.Role {
		out = append(out, "Peran "+before.Role.Label()+" → "+after.Role.Label())
		detail["peran_lama"], detail["peran_baru"] = before.Role.Label(), after.Role.Label()
	}
	if before.IsActive != after.IsActive {
		if after.IsActive {
			out = append(out, "Mengaktifkan akun")
		} else {
			out = append(out, "Menonaktifkan akun")
		}
	}
	if before.FullName != after.FullName {
		out = append(out, "Nama diubah")
		detail["nama_lama"] = before.FullName
	}
	if deref(before.Phone) != deref(after.Phone) {
		out = append(out, "Nomor telepon diubah")
	}
	if passwordReset {
		out = append(out, "Kata sandi diatur ulang")
	}
	return out, detail
}

func userChangeAction(before, after domain.User) string {
	switch {
	case before.IsActive && !after.IsActive:
		return "NONAKTIFKAN"
	case !before.IsActive && after.IsActive:
		return "AKTIFKAN"
	case before.Role != after.Role:
		return "UBAH_PERAN"
	}
	return "UBAH"
}

func validEmail(s string) bool {
	addr, err := mail.ParseAddress(strings.TrimSpace(s))
	return err == nil && addr.Address == strings.TrimSpace(s)
}

// optionalPhone menormalkan nomor telepon (untuk login OTP); kosong = tanpa nomor.
func optionalPhone(p *string) (*string, error) {
	if p == nil || strings.TrimSpace(*p) == "" {
		return nil, nil
	}
	n, err := auth.NormalizePhone(*p)
	if err != nil {
		return nil, httpx.Unprocessable("nomor telepon tidak valid — contoh: 0812 3456 7890")
	}
	return &n, nil
}
