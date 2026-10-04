// Package handlers berisi handler HTTP per modul dan routing API.
package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/config"
	"transcpg-x/backend/pkg/domain"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/storage"
	"transcpg-x/backend/pkg/store"
)

type App struct {
	cfg    config.Config
	store  *store.Store
	google *auth.GoogleVerifier
	sms    auth.SMSSender  // nil = login telepon nonaktif
	files  storage.Storage // nil = unggah lampiran nonaktif
	otpKey []byte          // kunci HMAC kode OTP (turunan JWT_SECRET)
}

func New(cfg config.Config, st *store.Store) *App {
	a := &App{cfg: cfg, store: st, google: auth.NewGoogleVerifier(cfg.GoogleClientID),
		otpKey: auth.DeriveKey(cfg.JWTSecret, "otp")}
	if cfg.OTPSender == "log" {
		a.sms = auth.LogSender{}
	}
	switch cfg.StorageDriver {
	case "supabase":
		a.files = storage.NewSupabase(cfg.SupabaseURL, cfg.SupabaseKey, cfg.StorageBucket)
	case "local":
		a.files = &storage.Local{Dir: cfg.LocalStorageDir, Secret: auth.DeriveKey(cfg.JWTSecret, "storage-local"), BaseURL: "/api/storage/local"}
	}
	return a
}

// Routes mengembalikan handler lengkap: CORS → recover → routing.
func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()

	// Publik
	mux.HandleFunc("GET /api/health", h(a.health))
	mux.HandleFunc("POST /api/auth/login", h(a.login))
	mux.HandleFunc("GET /api/auth/methods", h(a.loginMethods))
	mux.HandleFunc("POST /api/auth/google", h(a.loginGoogle))
	mux.HandleFunc("POST /api/auth/otp/request", h(a.requestOTP))
	mux.HandleFunc("POST /api/auth/otp/verify", h(a.verifyOTP))
	mux.HandleFunc("POST /api/auth/register", h(a.register))
	mux.HandleFunc("GET /api/auth/register/roles", h(a.registrationRoles))
	a.routeLocalStorage(mux)

	// Wajib login
	p := http.NewServeMux()
	a.routeAuth(p)
	a.routeProfile(p)
	a.routeMeta(p)
	a.routeHospitals(p)
	a.routeAdminUsers(p)
	a.routeAdminRegistrations(p)
	a.routeDashboard(p)
	a.routeLibrary(p)
	a.routePathway(p)
	a.routeReferences(p)
	a.routePlan(p)
	a.routeCriteria(p)
	a.routeGuidelines(p)
	a.routeAttachments(p)
	a.routeApproval(p)
	a.routeActive(p)
	a.routeEvaluation(p)
	a.routeKPTL(p)
	a.routeSnomed(p)
	a.routeMaster(p)
	a.routeAudit(p)
	mux.Handle("/api/", auth.Require(a.cfg.JWTSecret, a.store.GetUser)(p))

	return a.cors(recoverer(mux))
}

func (a *App) health(w http.ResponseWriter, r *http.Request) error {
	return httpx.OK(w, map[string]string{"status": "ok"})
}

// h membungkus handler dan menerjemahkan galat domain menjadi galat HTTP.
func h(fn httpx.HandlerFunc) http.HandlerFunc {
	return httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		return translate(fn(w, r))
	})
}

// admin = h + hanya untuk Admin.
func admin(fn httpx.HandlerFunc) http.Handler {
	return auth.RequireAdmin(h(fn))
}

func translate(err error) error {
	if err == nil {
		return nil
	}
	var nre *domain.NotReadyError
	switch {
	case errors.As(err, &nre):
		return &httpx.Error{Status: http.StatusConflict, Code: "SYARAT_BELUM_TERPENUHI", Message: "CP belum memenuhi syarat untuk disahkan", Details: nre.Unmet}
	case errors.Is(err, domain.ErrNotAuthorized):
		return httpx.Forbidden(err.Error())
	case errors.Is(err, domain.ErrContentLocked):
		return &httpx.Error{Status: http.StatusConflict, Code: "TERKUNCI", Message: err.Error()}
	case errors.Is(err, domain.ErrInvalidAction):
		return &httpx.Error{Status: http.StatusConflict, Code: "TINDAKAN_TIDAK_BERLAKU", Message: err.Error()}
	case errors.Is(err, domain.ErrReasonRequired):
		return httpx.Unprocessable(err.Error())
	case errors.Is(err, store.ErrRetiredConcept), errors.Is(err, store.ErrGuidelineNotReference):
		return httpx.Unprocessable(err.Error())
	case errors.Is(err, auth.ErrPasswordTooLong):
		return httpx.Unprocessable(err.Error())
	case errors.Is(err, store.ErrEmailTaken), errors.Is(err, store.ErrNotPending):
		return httpx.Conflict(err.Error())
	case errors.Is(err, storage.ErrNotFound):
		return httpx.Unprocessable("file belum terunggah ke penyimpanan — ulangi unggahan")
	case errors.Is(err, storage.ErrTooLarge):
		return httpx.NewError(http.StatusRequestEntityTooLarge, "TERLALU_BESAR", "ukuran file melebihi 50 MB")
	case errors.Is(err, storage.ErrNotPDF):
		return httpx.Unprocessable("file bukan PDF")
	case errors.Is(err, storage.ErrBadToken):
		return httpx.Forbidden("tautan unggah/unduh tidak valid atau kedaluwarsa — ulangi dari aplikasi")
	}
	return err
}

func (a *App) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && slices.Contains(a.cfg.AllowedOrigins, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				slog.Error("panic", "path", r.URL.Path, "value", v, "stack", string(debug.Stack()))
				httpx.WriteError(w, r, httpx.NewError(500, "INTERNAL", "terjadi kesalahan pada server"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
