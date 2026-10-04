package handlers

import (
	"net/http"

	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
)

// Halaman CP Aktif: pratinjau baca-saja. Penerapan ke pasien ada di TransCPR-X.
func (a *App) routeActive(m *http.ServeMux) {
	m.HandleFunc("GET /api/active-pathways", h(a.listActive))
	m.HandleFunc("GET /api/active-pathways/{code}", h(a.getActive))
}

func (a *App) listActive(w http.ResponseWriter, r *http.Request) error {
	list, err := a.store.ActivePathways(r.Context())
	if err != nil {
		return err
	}
	summary, err := a.store.ActiveSummary(r.Context())
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"summary": summary, "items": list})
}

func (a *App) getActive(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	if p.Stage != domain.StageAktif {
		return &httpx.Error{
			Status: http.StatusConflict, Code: "BELUM_AKTIF",
			Message: "CP ini belum Aktif (status saat ini: " + p.Stage.Label() + ")",
			Details: map[string]string{"stage": string(p.Stage)},
		}
	}
	severity := httpx.QueryInt(r, "severity", 0)
	ctx := r.Context()
	plan, err := a.store.ListPlan(ctx, p.ID, severity)
	if err != nil {
		return err
	}
	pattern, err := a.store.ClaimPattern(ctx, p.CBGCode, severity)
	if err != nil {
		return err
	}
	refs, err := a.store.References(ctx, p.ID)
	if err != nil {
		return err
	}
	approvals, err := a.store.ApprovalHistory(ctx, p.ID)
	if err != nil {
		return err
	}
	changes, err := a.store.ChangeLog(ctx, p.ID)
	if err != nil {
		return err
	}
	criteria, err := a.store.ListCriteria(ctx, p.ID)
	if err != nil {
		return err
	}
	// 4 tab: Rencana per Hari · Pola Klaim · Acuan & Pengesahan · Kriteria
	return httpx.OK(w, map[string]any{
		"pathway":       p,
		"plan":          plan,
		"claim_pattern": pattern,
		"references_approval": map[string]any{
			"references": refs, "approvals": approvals, "changes": changes,
		},
		"criteria": criteria,
	})
}
