package handlers

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Pencarian master data: icd10, icd9cm, loinc, kfa, kptl, snomed.
func (a *App) routeMaster(m *http.ServeMux) {
	m.HandleFunc("GET /api/master/{kind}", h(a.searchMaster))
}

func (a *App) searchMaster(w http.ResponseWriter, r *http.Request) error {
	kind := r.PathValue("kind")
	if !store.IsMasterKind(kind) {
		return httpx.NotFound("master " + kind + " tidak dikenal")
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(q) < 2 {
		return httpx.OK(w, []store.Row{})
	}
	rows, err := a.store.SearchMaster(r.Context(), kind, q, r.URL.Query().Get("semantic_tag"))
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}
