package handlers

import (
	"net/http"

	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
)

func (a *App) routeMeta(m *http.ServeMux) {
	m.HandleFunc("GET /api/meta", h(a.meta))
	m.HandleFunc("GET /api/mdc", h(a.listMDC))
}

// meta: label peran & tahap serta konstanta, supaya frontend tidak
// menduplikasi aturan.
func (a *App) meta(w http.ResponseWriter, r *http.Request) error {
	type opt struct {
		Value string `json:"value"`
		Label string `json:"label"`
	}
	var roles []opt
	for _, ro := range domain.AllRoles() {
		roles = append(roles, opt{string(ro), ro.Label()})
	}
	var stages []map[string]string
	for _, s := range domain.AllStages() {
		stages = append(stages, map[string]string{
			"value": string(s), "label": s.Label(), "advance_label": s.AdvanceLabel(), "waiting_on": domain.WaitingOn(s),
		})
	}
	return httpx.OK(w, map[string]any{
		"roles":  roles,
		"stages": stages,
		"constants": map[string]int{
			"min_episodes_eligible":     domain.MinEpisodesEligible,
			"min_episodes_per_severity": domain.MinEpisodesPerSeverity,
			"max_plan_day":              domain.MaxPlanDay,
			"library_page_size":         domain.LibraryPageSize,
		},
	})
}

func (a *App) listMDC(w http.ResponseWriter, r *http.Request) error {
	rows, err := a.store.ListMDC(r.Context())
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}
