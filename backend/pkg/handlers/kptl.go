package handlers

import (
	"net/http"
	"strings"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
)

// Meja kerja Padanan KPTL (ICD-9-CM → kode billing BPJS).
func (a *App) routeKPTL(m *http.ServeMux) {
	m.HandleFunc("GET /api/mappings/kptl/summary", h(a.kptlSummary))
	m.HandleFunc("GET /api/mappings/kptl", h(a.kptlWorklist))
	m.HandleFunc("GET /api/mappings/kptl/{icd9}/suggestions", h(a.kptlSuggestions))
	m.HandleFunc("PUT /api/mappings/kptl/{icd9}", h(a.setKPTL))
	m.HandleFunc("DELETE /api/mappings/kptl/{icd9}", h(a.deleteKPTL))
}

func requireMappingRole(r *http.Request) error {
	if !domain.CanManageMappings(auth.CurrentUser(r).Role) {
		return httpx.Forbidden("peran Anda hanya dapat melihat padanan")
	}
	return nil
}

func (a *App) kptlSummary(w http.ResponseWriter, r *http.Request) error {
	row, err := a.store.KPTLSummary(r.Context())
	if err != nil {
		return err
	}
	row["can_manage"] = domain.CanManageMappings(auth.CurrentUser(r).Role)
	return httpx.OK(w, row)
}

func (a *App) kptlWorklist(w http.ResponseWriter, r *http.Request) error {
	tab := r.URL.Query().Get("tab")
	rows, err := a.store.KPTLWorklist(r.Context(), tab != "semua", strings.TrimSpace(r.URL.Query().Get("q")))
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}

func (a *App) kptlSuggestions(w http.ResponseWriter, r *http.Request) error {
	rows, err := a.store.KPTLSuggestions(r.Context(), r.PathValue("icd9"))
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{
		"suggestions": rows,
		"warning":     "Usulan mesin berdasarkan kemiripan nama — hanya ±30% tepat di peringkat 1. Keputusan tetap di tangan Tim Koding.",
	})
}

func (a *App) setKPTL(w http.ResponseWriter, r *http.Request) error {
	if err := requireMappingRole(r); err != nil {
		return err
	}
	var in struct {
		KPTLCode string  `json:"kptl_code"`
		Note     *string `json:"note"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if strings.TrimSpace(in.KPTLCode) == "" {
		return httpx.Unprocessable("pilih kode KPTL")
	}
	if err := a.store.SetKPTLMapping(r.Context(), r.PathValue("icd9"), in.KPTLCode, in.Note, auth.CurrentUser(r)); err != nil {
		return err
	}
	return httpx.NoContent(w)
}

func (a *App) deleteKPTL(w http.ResponseWriter, r *http.Request) error {
	if err := requireMappingRole(r); err != nil {
		return err
	}
	if err := a.store.DeleteKPTLMapping(r.Context(), r.PathValue("icd9"), auth.CurrentUser(r)); err != nil {
		return err
	}
	return httpx.NoContent(w)
}
