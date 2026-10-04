package handlers

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Tab Rencana Klinis.
func (a *App) routePlan(m *http.ServeMux) {
	m.HandleFunc("GET /api/pathways/{code}/plan", h(a.listPlan))
	m.HandleFunc("POST /api/pathways/{code}/plan", h(a.addPlanItem))
	m.HandleFunc("PUT /api/pathways/{code}/plan/{id}", h(a.updatePlanItem))
	m.HandleFunc("DELETE /api/pathways/{code}/plan/{id}", h(a.deletePlanItem))
}

var (
	planItemTypes = []string{"OBAT", "LAB", "PROSEDUR", "TINDAKAN"}
	planNatures   = []string{"WAJIB", "KONDISIONAL"}
)

// validatePlanItem memberi pesan yang ramah sebelum constraint database.
func validatePlanItem(in *store.PlanItemInput) error {
	in.ItemCode = strings.TrimSpace(in.ItemCode)
	in.ItemName = strings.TrimSpace(in.ItemName)
	switch {
	case in.Severity < 1 || in.Severity > 3:
		return httpx.Unprocessable("severity harus I, II, atau III")
	case in.Day < 0 || in.Day > domain.MaxPlanDay:
		return httpx.Unprocessable(fmt.Sprintf("hari rawat harus H0 sampai H%d", domain.MaxPlanDay))
	case !slices.Contains(planItemTypes, in.ItemType):
		return httpx.Unprocessable("jenis harus OBAT, LAB, PROSEDUR, atau TINDAKAN")
	case !slices.Contains(planNatures, in.Nature):
		return httpx.Unprocessable("sifat harus WAJIB atau KONDISIONAL")
	case in.ItemCode == "" || in.ItemName == "":
		return httpx.Unprocessable("pilih item dari master data")
	}
	if in.ItemType == "OBAT" && (blank(in.Dose) || blank(in.Route) || blank(in.Frequency)) {
		return httpx.Unprocessable("obat wajib lengkap: dosis, rute, dan frekuensi harus diisi")
	}
	return nil
}

func blank(s *string) bool { return s == nil || strings.TrimSpace(*s) == "" }

func (a *App) listPlan(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	items, err := a.store.ListPlan(r.Context(), p.ID, httpx.QueryInt(r, "severity", 0))
	if err != nil {
		return err
	}
	return httpx.OK(w, items)
}

func (a *App) addPlanItem(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	var in store.PlanItemInput
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if err := validatePlanItem(&in); err != nil {
		return err
	}
	item, err := a.store.AddPlanItem(r.Context(), p.ID, in, auth.CurrentUser(r))
	if err != nil {
		return err
	}
	return httpx.JSON(w, http.StatusCreated, item)
}

func (a *App) updatePlanItem(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var in store.PlanItemInput
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	if err := validatePlanItem(&in); err != nil {
		return err
	}
	item, err := a.store.UpdatePlanItem(r.Context(), p.ID, id, in, auth.CurrentUser(r))
	if err != nil {
		return err
	}
	return httpx.OK(w, item)
}

func (a *App) deletePlanItem(w http.ResponseWriter, r *http.Request) error {
	p, err := a.pathway(r)
	if err != nil {
		return err
	}
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	if err := a.store.DeletePlanItem(r.Context(), p.ID, id, auth.CurrentUser(r)); err != nil {
		return err
	}
	return httpx.NoContent(w)
}
