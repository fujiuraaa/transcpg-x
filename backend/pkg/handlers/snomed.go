package handlers

import (
	"net/http"
	"strings"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
)

// Meja kerja Padanan SNOMED-CT. Tidak ada usulan mesin — pencarian manual.
func (a *App) routeSnomed(m *http.ServeMux) {
	m.HandleFunc("GET /api/mappings/snomed/summary", h(a.snomedSummary))
	m.HandleFunc("GET /api/mappings/snomed", h(a.snomedWorklist))
	m.HandleFunc("PUT /api/mappings/snomed/{vocab}/{code}", h(a.setSnomed))
}

func (a *App) snomedSummary(w http.ResponseWriter, r *http.Request) error {
	rows, err := a.store.SnomedSummary(r.Context())
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{
		"vocabularies": rows,
		"can_manage":   domain.CanManageMappings(auth.CurrentUser(r).Role),
	})
}

func (a *App) snomedWorklist(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	filter, vocab := q.Get("filter"), q.Get("vocab")
	if filter != "" && filter != "ragu" && filter != "belum" {
		return httpx.BadRequest("filter harus ragu, belum, atau kosong")
	}
	if vocab != "" && vocab != "ICD10" && vocab != "ICD9CM" {
		return httpx.BadRequest("vocab harus ICD10 atau ICD9CM")
	}
	rows, err := a.store.SnomedWorklist(r.Context(), filter, vocab)
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}

func (a *App) setSnomed(w http.ResponseWriter, r *http.Request) error {
	if err := requireMappingRole(r); err != nil {
		return err
	}
	vocab := r.PathValue("vocab")
	if vocab != "ICD10" && vocab != "ICD9CM" {
		return httpx.BadRequest("vocab harus ICD10 atau ICD9CM")
	}
	var in struct {
		ConceptID string `json:"concept_id"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if strings.TrimSpace(in.ConceptID) == "" {
		return httpx.Unprocessable("pilih konsep SNOMED-CT")
	}
	if err := a.store.SetSnomedMapping(r.Context(), vocab, r.PathValue("code"), in.ConceptID, auth.CurrentUser(r)); err != nil {
		return err
	}
	return httpx.NoContent(w)
}
