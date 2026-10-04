package handlers

import (
	"net/http"
	"slices"
	"strings"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

func (a *App) routeHospitals(m *http.ServeMux) {
	m.HandleFunc("GET /api/hospitals", h(a.listHospitals))
	m.HandleFunc("GET /api/hospitals/{id}", h(a.getHospital))
	m.Handle("PUT /api/hospitals/{id}", admin(a.updateHospital))
}

func (a *App) listHospitals(w http.ResponseWriter, r *http.Request) error {
	list, err := a.store.ListHospitals(r.Context())
	if err != nil {
		return err
	}
	return httpx.OK(w, list)
}

func (a *App) getHospital(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	hosp, err := a.store.GetHospital(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(w, hosp)
}

var hospitalTypes = []string{"A", "B", "C", "D", "KLINIK_UTAMA", "KLINIK_PRATAMA", "PUSKESMAS"}

func (a *App) updateHospital(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var in store.Hospital
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	in.ID = id
	// Admin RS hanya boleh mengubah profil RS-nya sendiri; System/Super Admin semua.
	actor := auth.CurrentUser(r)
	if actor.Role == domain.RoleAdminRS && (actor.HospitalID == nil || *actor.HospitalID != id) {
		return httpx.Forbidden("Admin RS hanya dapat mengubah profil rumah sakitnya sendiri")
	}
	for _, f := range []**string{&in.BPJSProviderCode, &in.Accreditation, &in.City, &in.Province, &in.Address, &in.Phone, &in.Email} {
		if *f != nil {
			if t := strings.TrimSpace(**f); t == "" {
				*f = nil
			} else {
				*f = &t
			}
		}
	}
	in.Name = strings.TrimSpace(in.Name)
	switch {
	case strings.TrimSpace(in.Name) == "":
		return httpx.Unprocessable("nama RS wajib diisi")
	case !slices.Contains(hospitalTypes, in.HospitalType):
		return httpx.Unprocessable("tipe RS tidak dikenal")
	case in.Ownership != "PEMERINTAH" && in.Ownership != "SWASTA":
		return httpx.Unprocessable("kepemilikan harus PEMERINTAH atau SWASTA")
	case in.BPJSRegional < 1 || in.BPJSRegional > 5:
		return httpx.Unprocessable("regional BPJS harus 1–5")
	case in.MaxCareClass < 1 || in.MaxCareClass > 3:
		return httpx.Unprocessable("kelas rawat tertinggi harus 1–3")
	case in.Email != nil && !validEmail(*in.Email):
		return httpx.Unprocessable("format email RS tidak valid")
	case (in.BedCapacity != nil && *in.BedCapacity < 0) || (in.ICUCapacity != nil && *in.ICUCapacity < 0):
		return httpx.Unprocessable("kapasitas tidak boleh negatif")
	}
	hosp, err := a.store.UpdateHospital(r.Context(), in)
	if err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "RS", Action: "UBAH", TargetType: "HOSPITAL", TargetID: hosp.ID,
		TargetLabel: hosp.Name, HospitalID: &hosp.ID, Summary: "Mengubah profil rumah sakit",
		Detail: map[string]any{"tipe": hosp.HospitalType, "regional_bpjs": hosp.BPJSRegional, "kelas_tertinggi": hosp.MaxCareClass}})
	return httpx.OK(w, hosp)
}

// hospitalFromQuery membaca ?hospital_id= (profil RS di bilah atas);
// kosong = profil bawaan.
func (a *App) hospitalFromQuery(r *http.Request) (store.Hospital, error) {
	return a.store.ResolveHospital(r.Context(), int64(httpx.QueryInt(r, "hospital_id", 0)))
}
