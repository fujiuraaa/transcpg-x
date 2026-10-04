package handlers

import (
	"net/http"
	"strings"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Antrean permintaan pendaftaran (Pengaturan › Manajemen User).
func (a *App) routeAdminRegistrations(m *http.ServeMux) {
	m.Handle("GET /api/admin/registrations", admin(a.listRegistrations))
	m.Handle("POST /api/admin/registrations/{id}/approve", admin(a.approveRegistration))
	m.Handle("POST /api/admin/registrations/{id}/reject", admin(a.rejectRegistration))
}

func (a *App) listRegistrations(w http.ResponseWriter, r *http.Request) error {
	status := domain.RegistrationStatus(r.URL.Query().Get("status"))
	if status == "" {
		status = domain.RegPending
	}
	if status != domain.RegPending && status != domain.RegRejected {
		return httpx.BadRequest("status harus MENUNGGU atau DITOLAK")
	}
	list, err := a.store.ListRegistrations(r.Context(), status, adminScope(auth.CurrentUser(r)))
	if err != nil {
		return err
	}
	out := make([]sessionUser, len(list))
	for i, u := range list {
		out[i] = toSessionUser(u)
	}
	return httpx.OK(w, out)
}

// approveRegistration: Admin menetapkan peran final (boleh berbeda dari yang diajukan).
func (a *App) approveRegistration(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		Role domain.Role `json:"role"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	actor := auth.CurrentUser(r)
	if !in.Role.AssignableBy(actor.Role) {
		return httpx.Forbidden("peran tersebut tidak dapat Anda berikan")
	}
	if err := a.registrationInScope(r, id); err != nil {
		return err
	}
	u, err := a.store.ApproveRegistration(r.Context(), id, in.Role, actor)
	if err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "AKUN", Action: "SETUJUI", TargetType: "USER", TargetID: u.ID,
		TargetLabel: u.FullName, HospitalID: u.HospitalID, Summary: "Menyetujui pendaftaran dengan peran " + u.Role.Label(),
		Detail: map[string]any{"email": u.Email}})
	return httpx.OK(w, toSessionUser(u))
}

func (a *App) rejectRegistration(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var in struct {
		Reason string `json:"reason"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if strings.TrimSpace(in.Reason) == "" {
		return httpx.Unprocessable("alasan penolakan wajib diisi")
	}
	if err := a.registrationInScope(r, id); err != nil {
		return err
	}
	u, err := a.store.RejectRegistration(r.Context(), id, strings.TrimSpace(in.Reason), auth.CurrentUser(r))
	if err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "AKUN", Action: "TOLAK", TargetType: "USER", TargetID: u.ID,
		TargetLabel: u.FullName, HospitalID: u.HospitalID, Summary: "Menolak pendaftaran",
		Detail: map[string]any{"email": u.Email, "alasan": strings.TrimSpace(in.Reason)}})
	return httpx.OK(w, toSessionUser(u))
}

// registrationInScope: Admin RS hanya memproses pendaftar untuk RS-nya.
func (a *App) registrationInScope(r *http.Request, id int64) error {
	target, err := a.store.GetUser(r.Context(), id)
	if err != nil {
		return err
	}
	if !inScope(auth.CurrentUser(r), target) {
		return httpx.NotFound("pendaftaran tidak ditemukan")
	}
	return nil
}
