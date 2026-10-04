package handlers

import (
	"net/http"
	"strings"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Panel Acuan Standar + tab Acuan Klinis.
func (a *App) routeReferences(m *http.ServeMux) {
	m.HandleFunc("GET /api/pathways/{code}/references", h(a.listReferences))
	m.HandleFunc("POST /api/pathways/{code}/references", h(a.addReference))
	m.HandleFunc("DELETE /api/pathways/{code}/references/{id}", h(a.removeReference))
	m.HandleFunc("GET /api/pathways/{code}/guideline-candidates", h(a.guidelineCandidates))
}

func (a *App) listReferences(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	rows, err := a.store.References(r.Context(), p.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}

func (a *App) addReference(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	var in store.NewReference
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if in.DocumentID <= 0 {
		return httpx.Unprocessable("pilih dokumen panduan")
	}
	id, err := a.store.AddReference(r.Context(), p.ID, in, auth.CurrentUser(r))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (a *App) removeReference(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	if err := a.store.RemoveReference(r.Context(), p.ID, id, auth.CurrentUser(r)); err != nil {
		return err
	}
	return httpx.NoContent(w)
}

func (a *App) guidelineCandidates(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	rows, err := a.store.GuidelineCandidates(r.Context(), p.CBGCode, strings.TrimSpace(r.URL.Query().Get("q")))
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}
