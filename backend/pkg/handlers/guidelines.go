package handlers

import (
	"net/http"
	"slices"
	"strings"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Dokumen panduan: lihat butir, dan "Masukkan dokumen panduan sendiri".
func (a *App) routeGuidelines(m *http.ServeMux) {
	m.HandleFunc("GET /api/guidelines/{id}/items", h(a.guidelineItems))
	m.HandleFunc("POST /api/guidelines", h(a.createGuideline))
	m.HandleFunc("POST /api/guidelines/{id}/items", h(a.addGuidelineItem))
}

var (
	guidelineSources    = []string{"PPK_RSCM", "PNPK", "PPK_ASOSIASI", "TIM_CP", "INTERNASIONAL", "TKMKB", "LAIN"}
	guidelineCategories = []string{"KRITERIA_DIAGNOSIS", "TATA_LAKSANA", "DOSIS_OBAT", "INDIKASI_TERAPI", "KONTRAINDIKASI", "MONITORING", "KOMPLIKASI", "LAIN"}
)

// Pendaftaran dokumen mengikuti wewenang menetapkan acuan saat Draf.
func canRegisterGuideline(role domain.Role) bool {
	return domain.CanEditContent(domain.StageDraf, domain.ContentAcuan, role)
}

func (a *App) guidelineItems(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	rows, err := a.store.GuidelineItems(r.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(w, rows)
}

func (a *App) createGuideline(w http.ResponseWriter, r *http.Request) error {
	u := auth.CurrentUser(r)
	if !canRegisterGuideline(u.Role) {
		return httpx.Forbidden("peran Anda tidak dapat mendaftarkan dokumen panduan")
	}
	var in store.NewGuideline
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return httpx.Unprocessable("nama dokumen wajib diisi")
	}
	if !slices.Contains(guidelineSources, in.SourceType) {
		return httpx.Unprocessable("sumber dokumen tidak dikenal")
	}
	id, err := a.store.CreateGuideline(r.Context(), in, u)
	if err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "DOKUMEN", Action: "DAFTARKAN", TargetType: "GUIDELINE", TargetID: id,
		TargetLabel: in.Title, Summary: "Mendaftarkan dokumen panduan", Detail: map[string]any{"sumber": in.SourceType}})
	return httpx.JSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (a *App) addGuidelineItem(w http.ResponseWriter, r *http.Request) error {
	if !canRegisterGuideline(auth.CurrentUser(r).Role) {
		return httpx.Forbidden("peran Anda tidak dapat mengisi butir acuan")
	}
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	if err := a.store.GuidelineExists(r.Context(), id); err != nil {
		return err
	}
	if err := a.guidelineWritable(r, id); err != nil {
		return err
	}
	var in store.NewGuidelineItem
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	switch {
	case !slices.Contains(guidelineCategories, in.Category):
		return httpx.Unprocessable("kategori butir tidak dikenal")
	case strings.TrimSpace(in.Title) == "":
		return httpx.Unprocessable("judul butir wajib diisi")
	case strings.TrimSpace(in.Page) == "" || strings.TrimSpace(in.Quote) == "":
		return httpx.Unprocessable("halaman dan kutipan keduanya wajib diisi")
	}
	itemID, err := a.store.AddGuidelineItem(r.Context(), id, in)
	if err != nil {
		return err
	}
	a.record(r, store.Activity{Category: "DOKUMEN", Action: "TAMBAH_BUTIR", TargetType: "GUIDELINE", TargetID: id,
		TargetLabel: a.store.GuidelineTitle(r.Context(), id), Summary: "Menambah butir acuan: " + strings.TrimSpace(in.Title),
		Detail: map[string]any{"kategori": in.Category, "halaman": in.Page}})
	return httpx.JSON(w, http.StatusCreated, map[string]int64{"id": itemID})
}

// guidelineWritable: dokumen yang menjadi acuan CP berisi terkunci (Review
// KSM/Komite Medik, Menunggu Direktur) tidak boleh diubah butir/lampirannya —
// isi acuan yang sedang ditinjau tidak boleh bergeser.
func (a *App) guidelineWritable(r *http.Request, id int64) error {
	locked, err := a.store.GuidelineLocked(r.Context(), id)
	if err != nil {
		return err
	}
	if locked {
		return &httpx.Error{Status: http.StatusConflict, Code: "TERKUNCI",
			Message: "dokumen ini menjadi acuan CP yang sedang ditinjau KSM/Komite Medik atau Direktur — butir dan lampirannya terkunci sampai tahap itu selesai"}
	}
	return nil
}
