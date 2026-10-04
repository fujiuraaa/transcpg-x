package handlers

import (
	"net/http"

	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
)

// Halaman Evaluasi CP: kendali mutu & kendali biaya sesudah CP aktif.
func (a *App) routeEvaluation(m *http.ServeMux) {
	m.HandleFunc("GET /api/evaluations", h(a.listEvaluations))
	m.HandleFunc("GET /api/evaluations/{code}", h(a.getEvaluation))
}

func (a *App) listEvaluations(w http.ResponseWriter, r *http.Request) error {
	tab := r.URL.Query().Get("tab")
	if tab != "" && tab != "aktif" && tab != "belum" {
		return httpx.BadRequest("tab harus aktif atau belum")
	}
	summary, rows, err := a.store.Evaluations(r.Context(), tab != "belum")
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"summary": summary, "items": rows})
}

func (a *App) getEvaluation(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	if p.Stage != domain.StageAktif || p.ActivatedAt == nil {
		return &httpx.Error{Status: http.StatusConflict, Code: "BELUM_AKTIF", Message: "CP belum Aktif — belum bisa dievaluasi"}
	}
	detail, err := a.store.EvaluationDetail(r.Context(), p.ID, p.CBGCode, *p.ActivatedAt)
	if err != nil {
		return err
	}
	return httpx.OK(w, detail)
}
