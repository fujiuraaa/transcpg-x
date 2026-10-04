package handlers

import (
	"net/http"
	"strings"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Tab Kriteria inklusi/eksklusi. Aturan wewenang sama dengan Rencana Klinis.
func (a *App) routeCriteria(m *http.ServeMux) {
	m.HandleFunc("GET /api/pathways/{code}/criteria", h(a.listCriteria))
	m.HandleFunc("POST /api/pathways/{code}/criteria", h(a.addCriterion))
	m.HandleFunc("DELETE /api/pathways/{code}/criteria/{id}", h(a.deleteCriterion))
}

func (a *App) listCriteria(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	list, err := a.store.ListCriteria(r.Context(), p.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, list)
}

func (a *App) addCriterion(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	var in store.CriterionInput
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	in.Description = strings.TrimSpace(in.Description)
	if in.Kind != "INKLUSI" && in.Kind != "EKSKLUSI" {
		return httpx.Unprocessable("jenis kriteria harus INKLUSI atau EKSKLUSI")
	}
	if in.Description == "" {
		return httpx.Unprocessable("uraian kriteria wajib diisi")
	}
	c, err := a.store.AddCriterion(r.Context(), p.ID, in, auth.CurrentUser(r))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, c)
}

func (a *App) deleteCriterion(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	if err := a.store.DeleteCriterion(r.Context(), p.ID, id, auth.CurrentUser(r)); err != nil {
		return err
	}
	return httpx.NoContent(w)
}
