package handlers

import (
	"net/http"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Halaman Detail CP: kepala, panel, 5 kartu, dan tab baca-saja.
// Tab yang bisa disunting ada di references.go, plan.go, criteria.go.
func (a *App) routePathway(m *http.ServeMux) {
	m.HandleFunc("GET /api/pathways/{code}", h(a.getPathway))
	m.HandleFunc("GET /api/pathways/{code}/flow", h(a.pathwayFlow))              // Alur Pelaksanaan
	m.HandleFunc("GET /api/pathways/{code}/summary", h(a.pathwaySummary))        // Ringkasan
	m.HandleFunc("GET /api/pathways/{code}/diagnoses", h(a.pathwayDiagnoses))    // Diagnosis
	m.HandleFunc("GET /api/pathways/{code}/procedures", h(a.pathwayProcedures))  // Prosedur & KPTL
	m.HandleFunc("GET /api/pathways/{code}/tariffs", h(a.pathwayTariffs))        // Tarif & Biaya
	m.HandleFunc("GET /api/pathways/{code}/daily", h(a.pathwayDaily))            // Intervensi Harian
	m.HandleFunc("GET /api/pathways/{code}/cdss-rules", h(a.pathwayRules))       // Aturan CDSS
	m.HandleFunc("GET /api/pathways/{code}/scoring", h(a.pathwayScoring))        // Scoring Klinis
	m.HandleFunc("GET /api/pathways/{code}/quality", h(a.pathwayQuality))        // Indikator Mutu
	m.HandleFunc("GET /api/pathways/{code}/readiness", h(a.pathwayReadiness))    // Verifikasi Kode
	m.HandleFunc("GET /api/pathways/{code}/history", h(a.pathwayHistory))        // Riwayat pengesahan & perubahan
	m.HandleFunc("POST /api/pathways/{code}/transition", h(a.transitionPathway)) // Panel Pengesahan
}

// pathway memuat CP dari parameter {code}.
func (a *App) pathway(r *http.Request) (store.PathwayHeader, error) {
	p, err := a.store.GetPathway(r.Context(), r.PathValue("code"))
	if store.IsNotFound(err) {
		return p, httpx.NotFound("CP " + r.PathValue("code") + " tidak ditemukan")
	}
	return p, err
}

type actionView struct {
	Action        domain.Action `json:"action"`
	Label         string        `json:"label"`
	RequireReason bool          `json:"require_reason"`
}

// permissions: tombol & formulir yang boleh tampil untuk pengguna ini.
// Server tetap memeriksa ulang setiap aksi.
func permissions(stage domain.Stage, role domain.Role) map[string]any {
	var actions []actionView
	for _, act := range domain.AvailableActions(stage, role) {
		v := actionView{Action: act}
		switch act {
		case domain.ActionAdvance:
			v.Label = stage.AdvanceLabel()
		case domain.ActionReturn:
			v.Label, v.RequireReason = "Kembalikan untuk Revisi", true
		case domain.ActionRevoke:
			v.Label, v.RequireReason = "Cabut keberlakuan", true
		}
		actions = append(actions, v)
	}
	if actions == nil {
		actions = []actionView{}
	}
	return map[string]any{
		"actions":        actions,
		"waiting_on":     domain.WaitingOn(stage),
		"content_locked": domain.IsContentLocked(stage),
		"can_edit": map[string]bool{
			"acuan":    domain.CanEditContent(stage, domain.ContentAcuan, role),
			"rencana":  domain.CanEditContent(stage, domain.ContentRencana, role),
			"kriteria": domain.CanEditContent(stage, domain.ContentKriteria, role),
		},
	}
}

func (a *App) getPathway(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	hosp, err := a.hospitalFromQuery(r)
	if err != nil {
		return err
	}
	tariffs, err := a.store.Tariffs(r.Context(), p.CBGCode, hosp)
	if err != nil {
		return err
	}
	// Kartu tarif: kelas 3, severity dominan, pada profil RS terpilih.
	var cardTariff any
	for _, t := range tariffs {
		if p.DominantSeverity != nil && t["severity"] == *p.DominantSeverity && t["care_class"] == int16(3) {
			cardTariff = t["tariff"]
		}
	}
	return httpx.OK(w, map[string]any{
		"pathway":     p,
		"stage_label": p.Stage.Label(),
		"permissions": permissions(p.Stage, auth.CurrentUser(r).Role),
		"cards": map[string]any{
			"mdc":               map[string]string{"code": p.MDCCode, "name": p.MDCName},
			"dominant_severity": p.DominantSeverity,
			"target_los":        p.TargetLOS,
			"tariff":            cardTariff,
			"candidate_count":   p.CandidateCount,
		},
		"hospital": map[string]any{"id": hosp.ID, "name": hosp.Name},
	})
}

func (a *App) pathwayFlow(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	subs, err := a.store.SubCPs(r.Context(), p.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{
		"split_decision": p.SplitDecision,
		"sub_cps":        subs,
		// TODO: langkah siklus hidup & daftar kerja (lihat alur-aplikasi §5).
	})
}

func (a *App) pathwaySummary(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	stats, err := a.store.SeverityStats(r.Context(), p.ID)
	if err != nil {
		return err
	}
	cost, err := a.store.CostDistribution(r.Context(), p.CBGCode)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"severity_stats": stats, "cost_distribution": cost})
}

func (a *App) pathwayDiagnoses(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	rows, err := a.store.Diagnoses(r.Context(), p.CBGCode)
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}

func (a *App) pathwayProcedures(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	rows, err := a.store.Procedures(r.Context(), p.CBGCode)
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}

func (a *App) pathwayTariffs(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	hosp, err := a.hospitalFromQuery(r)
	if err != nil {
		return err
	}
	tariffs, err := a.store.Tariffs(r.Context(), p.CBGCode, hosp)
	if err != nil {
		return err
	}
	cost, err := a.store.CostDistribution(r.Context(), p.CBGCode)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"hospital": hosp, "tariffs": tariffs, "cost_distribution": cost})
}

// pathwayDaily: tab Intervensi Harian — menampilkan data rencana isi klinis.
func (a *App) pathwayDaily(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	plan, err := a.store.ListPlan(r.Context(), p.ID, httpx.QueryInt(r, "severity", 0))
	if err != nil {
		return err
	}
	// TODO: matriks hari × kategori dari pola klaim (assessment, monitoring, lab, tindakan, asuhan).
	return httpx.OK(w, map[string]any{"plan": plan, "claim_matrix": nil})
}

func (a *App) pathwayRules(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	rows, err := a.store.RuleSuggestions(r.Context(), p.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{
		"rules": rows,
		"note":  "Usulan aturan diturunkan otomatis dari pola klaim, belum ditinjau manusia.",
	})
}

func (a *App) pathwayScoring(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	rows, err := a.store.ScoringTools(r.Context(), p.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}

func (a *App) pathwayQuality(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	rows, err := a.store.QualityIndicators(r.Context(), p.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}

func (a *App) pathwayReadiness(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	rd, err := a.store.Readiness(r.Context(), p.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"ready": rd.Ready(), "requirements": rd.Requirements})
}

func (a *App) pathwayHistory(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	approvals, err := a.store.ApprovalHistory(r.Context(), p.ID)
	if err != nil {
		return err
	}
	changes, err := a.store.ChangeLog(r.Context(), p.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"approvals": approvals, "changes": changes})
}

func (a *App) transitionPathway(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	var in struct {
		Action domain.Action `json:"action"`
		Reason string        `json:"reason"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	res, err := a.store.Transition(r.Context(), p.ID, in.Action, in.Reason, auth.CurrentUser(r))
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{
		"from": res.From, "to": res.To, "to_label": res.To.Label(),
	})
}
