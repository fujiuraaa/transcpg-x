package handlers

import (
	"net/http"
	"strings"

	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

func (a *App) routeLibrary(m *http.ServeMux) {
	m.HandleFunc("GET /api/pathways", h(a.listPathways))
}

// listPathways: CP Library — 25 CP per halaman dengan saringan & urutan.
func (a *App) listPathways(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	f := store.LibraryFilter{
		Query: strings.TrimSpace(q.Get("q")),
		MDC:   q.Get("mdc"),
		Acuan: q.Get("acuan"),
		Sort:  q.Get("sort"),
		Page:  httpx.QueryInt(r, "page", 1),
	}
	if f.Acuan != "" && f.Acuan != "sudah" && f.Acuan != "belum" {
		return httpx.BadRequest("acuan harus 'sudah' atau 'belum'")
	}
	hosp, err := a.hospitalFromQuery(r)
	if err != nil {
		return err
	}
	f.Hospital = hosp
	rows, total, err := a.store.ListPathways(r.Context(), f)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if s, ok := row["stage"].(string); ok {
			row["stage_label"] = domain.Stage(s).Label()
		}
	}
	return httpx.OK(w, map[string]any{
		"items":     rows,
		"total":     total,
		"page":      max(f.Page, 1),
		"page_size": domain.LibraryPageSize,
	})
}
