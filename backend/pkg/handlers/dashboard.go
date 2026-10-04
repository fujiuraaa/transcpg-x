package handlers

import (
	"net/http"

	"transcpg-x/backend/pkg/httpx"
)

func (a *App) routeDashboard(m *http.ServeMux) {
	m.HandleFunc("GET /api/dashboard", h(a.dashboard))
}

func (a *App) dashboard(w http.ResponseWriter, r *http.Request) error {
	d, err := a.store.DashboardSummary(r.Context())
	if err != nil {
		return err
	}
	return httpx.OK(w, d)
}
