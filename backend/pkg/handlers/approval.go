package handlers

import (
	"net/http"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Halaman Approval: kotak masuk pengesahan. Tindakan setuju/kembalikan
// TIDAK dilakukan di sini, melainkan di halaman Detail CP.
func (a *App) routeApproval(m *http.ServeMux) {
	m.HandleFunc("GET /api/approval", h(a.approvalQueue))
	m.HandleFunc("GET /api/approval/counts", h(a.approvalCounts))
}

func (a *App) approvalQueue(w http.ResponseWriter, r *http.Request) error {
	tab := store.ApprovalTab(r.URL.Query().Get("tab"))
	switch tab {
	case store.TabMine, store.TabProcessing, store.TabActive:
	case "":
		tab = store.TabMine
	default:
		return httpx.BadRequest("tab harus saya, proses, atau aktif")
	}
	u := auth.CurrentUser(r)
	rows, err := a.store.ApprovalQueue(r.Context(), tab, u.Role)
	if err != nil {
		return err
	}
	for _, row := range rows {
		stage := domain.Stage(row["stage"].(string))
		row["stage_label"] = stage.Label()
		row["waiting_on"] = domain.WaitingOn(stage)
		row["can_act"] = domain.CanAct(stage, u.Role) // tombol "Tinjau" vs "Buka"
		if stage != domain.StageAktif {
			rd, err := a.store.Readiness(r.Context(), row["id"].(int64))
			if err != nil {
				return err
			}
			missing := []string{}
			for _, q := range rd.UnmetBlocking() {
				missing = append(missing, q.Title)
			}
			row["syarat_aktif_lengkap"] = rd.Ready()
			row["syarat_aktif_kurang"] = missing
		}
	}
	return httpx.OK(w, rows)
}

func (a *App) approvalCounts(w http.ResponseWriter, r *http.Request) error {
	counts, err := a.store.StageCounts(r.Context())
	if err != nil {
		return err
	}
	mine, err := a.store.CountActionable(r.Context(), auth.CurrentUser(r).Role)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"by_stage": counts, "mine": mine})
}
