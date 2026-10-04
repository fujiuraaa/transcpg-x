package handlers

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Laporan & Audit (Pengaturan) — hanya admin. Admin RS melihat kejadian di
// RS-nya sendiri; System/Super Admin semua RS (boleh disaring ?hospital_id=).
func (a *App) routeAudit(m *http.ServeMux) {
	m.Handle("GET /api/admin/audit/events", admin(a.auditEvents))
	m.Handle("GET /api/admin/audit/events.csv", admin(a.auditEventsCSV))
	m.Handle("GET /api/admin/audit/actors", admin(a.auditActors))
	m.Handle("GET /api/admin/audit/reports/approval", admin(a.approvalReport))
	m.Handle("GET /api/admin/audit/reports/users", admin(a.usersReport))
}

var auditCategories = []string{"PENGESAHAN", "ISI_CP", "PADANAN", "DOKUMEN", "AKUN", "RS", "MASUK"}

const (
	auditPageSize  = 50
	auditCSVMaxRow = 10000
)

// Tanggal di filter dibaca sebagai WIB (UTC+7).
var wib = time.FixedZone("WIB", 7*3600)

// record mencatat kejadian ke activity_log. Gagal mencatat tidak
// membatalkan aksi yang sudah berhasil, tetapi dicatat ke log server.
func (a *App) record(r *http.Request, act store.Activity) {
	if act.Actor.ID == 0 {
		act.Actor = auth.CurrentUser(r)
	}
	act.IP = clientIP(r)
	if err := a.store.LogActivity(r.Context(), act); err != nil {
		slog.Error("gagal mencatat jejak audit", "category", act.Category, "action", act.Action, "err", err)
	}
}

// trustProxyHeaders: di Vercel (env VERCEL diset otomatis) header IP ditimpa
// oleh proxy sehingga bisa dipercaya; di tempat lain bisa dipalsukan klien,
// kecuali TRUST_PROXY_HEADERS=true diset sengaja di balik proxy tepercaya.
var trustProxyHeaders = os.Getenv("VERCEL") != "" || os.Getenv("TRUST_PROXY_HEADERS") == "true"

// clientIP: alamat klien untuk jejak audit.
func clientIP(r *http.Request) string {
	if trustProxyHeaders {
		if ip := r.Header.Get("X-Real-Ip"); ip != "" {
			return ip
		}
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			return strings.TrimSpace(strings.Split(xff, ",")[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// auditScope: RS yang boleh dilihat pelaku.
func auditScope(r *http.Request) *int64 {
	u := auth.CurrentUser(r)
	if u.Role == domain.RoleAdminRS {
		if u.HospitalID == nil {
			none := int64(-1) // Admin RS tanpa RS tidak melihat apa pun
			return &none
		}
		return u.HospitalID
	}
	if id := int64(httpx.QueryInt(r, "hospital_id", 0)); id > 0 {
		return &id
	}
	return nil
}

// period membaca ?from=YYYY-MM-DD&to=YYYY-MM-DD (keduanya inklusif).
func period(r *http.Request) (from, to *time.Time, err error) {
	parse := func(name string) (*time.Time, error) {
		v := r.URL.Query().Get(name)
		if v == "" {
			return nil, nil
		}
		t, err := time.ParseInLocation("2006-01-02", v, wib)
		if err != nil {
			return nil, httpx.BadRequest("tanggal " + name + " harus berformat YYYY-MM-DD")
		}
		return &t, nil
	}
	if from, err = parse("from"); err != nil {
		return
	}
	if to, err = parse("to"); err != nil {
		return
	}
	if to != nil {
		next := to.AddDate(0, 0, 1)
		to = &next
	}
	if from != nil && to != nil && !from.Before(*to) {
		return nil, nil, httpx.Unprocessable("tanggal awal harus sebelum tanggal akhir")
	}
	return
}

func auditFilter(r *http.Request) (store.AuditFilter, error) {
	from, to, err := period(r)
	if err != nil {
		return store.AuditFilter{}, err
	}
	f := store.AuditFilter{From: from, To: to, Q: strings.TrimSpace(r.URL.Query().Get("q")), HospitalID: auditScope(r)}
	if c := r.URL.Query().Get("category"); c != "" {
		for _, v := range strings.Split(c, ",") {
			if !slices.Contains(auditCategories, v) {
				return f, httpx.BadRequest("kategori tidak dikenal: " + v)
			}
			f.Categories = append(f.Categories, v)
		}
	}
	if id := int64(httpx.QueryInt(r, "actor", 0)); id > 0 {
		f.ActorID = &id
	}
	return f, nil
}

func (a *App) auditEvents(w http.ResponseWriter, r *http.Request) error {
	f, err := auditFilter(r)
	if err != nil {
		return err
	}
	page := max(httpx.QueryInt(r, "page", 1), 1)
	f.Limit, f.Offset = auditPageSize, (page-1)*auditPageSize
	items, total, err := a.store.AuditEvents(r.Context(), f)
	if err != nil {
		return err
	}
	for _, e := range items {
		e["actor_role_label"] = domain.Role(str(e["actor_role"])).Label()
	}
	counts, err := a.store.AuditCategoryCounts(r.Context(), f)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{
		"items": items, "total": total, "page": page, "per_page": auditPageSize, "counts": counts,
	})
}

var categoryLabels = map[string]string{
	"PENGESAHAN": "Pengesahan CP", "ISI_CP": "Isi CP", "PADANAN": "Padanan", "DOKUMEN": "Dokumen panduan",
	"AKUN": "Akun pengguna", "RS": "Profil RS", "MASUK": "Masuk aplikasi",
}

func (a *App) auditEventsCSV(w http.ResponseWriter, r *http.Request) error {
	f, err := auditFilter(r)
	if err != nil {
		return err
	}
	f.Limit = auditCSVMaxRow
	items, total, err := a.store.AuditEvents(r.Context(), f)
	if err != nil {
		return err
	}
	name := "jejak-audit-" + time.Now().In(wib).Format("20060102-1504") + ".csv"
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("X-Total-Count", fmt.Sprint(total))
	_, _ = w.Write([]byte("\xef\xbb\xbf")) // BOM agar Excel membaca UTF-8
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Waktu (WIB)", "Kategori", "Aksi", "Pelaku", "Peran", "Rumah sakit", "Objek", "Ringkasan", "Detail", "IP"})
	for _, e := range items {
		at, _ := e["created_at"].(time.Time)
		object := strings.TrimSpace(str(e["target_code"]) + " " + str(e["target_label"]))
		role := domain.Role(str(e["actor_role"])).Label()
		detail := ""
		if d, ok := e["detail"].(map[string]any); ok && len(d) > 0 {
			detail = flattenDetail(d)
		}
		_ = cw.Write(csvSafe(at.In(wib).Format("2006-01-02 15:04:05"), categoryLabels[str(e["category"])], str(e["action"]),
			str(e["actor_name"]), role, str(e["hospital_name"]), object, str(e["summary"]), detail, str(e["ip"])))
	}
	cw.Flush()
	return cw.Error()
}

func (a *App) auditActors(w http.ResponseWriter, r *http.Request) error {
	rows, err := a.store.AuditActors(r.Context(), auditScope(r))
	if err != nil {
		return err
	}
	for _, row := range rows {
		row["role_label"] = domain.Role(str(row["role"])).Label()
	}
	return httpx.OK(w, rows)
}

func (a *App) approvalReport(w http.ResponseWriter, r *http.Request) error {
	from, to, err := period(r)
	if err != nil {
		return err
	}
	rep, err := a.store.ApprovalReport(r.Context(), from, to)
	if err != nil {
		return err
	}
	return httpx.OK(w, rep)
}

func (a *App) usersReport(w http.ResponseWriter, r *http.Request) error {
	from, to, err := period(r)
	if err != nil {
		return err
	}
	rows, err := a.store.UsersReport(r.Context(), from, to, auditScope(r))
	if err != nil {
		return err
	}
	for _, row := range rows {
		row["role_label"] = domain.Role(str(row["role"])).Label()
	}
	return httpx.OK(w, rows)
}

func str(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

// flattenDetail: "kunci: nilai; …" — nilai kosong dilewati.
func flattenDetail(d map[string]any) string {
	parts := make([]string, 0, len(d))
	for _, k := range slices.Sorted(maps.Keys(d)) {
		if v := str(d[k]); v != "" && v != "<nil>" {
			parts = append(parts, k+": "+v)
		}
	}
	return strings.Join(parts, "; ")
}

// csvSafe mencegah injeksi formula saat CSV dibuka di Excel/Sheets.
func csvSafe(cells ...string) []string {
	for i, c := range cells {
		if c != "" && strings.ContainsRune("=+-@\t\r", rune(c[0])) {
			cells[i] = "'" + c
		}
	}
	return cells
}
