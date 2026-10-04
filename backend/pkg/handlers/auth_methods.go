package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/store"
)

// Metode masuk selain email + kata sandi. Semuanya hanya untuk akun yang
// sudah dibuat Admin RS — tidak ada pendaftaran mandiri.

// loginMethods memberi tahu halaman login metode mana yang aktif.
func (a *App) loginMethods(w http.ResponseWriter, r *http.Request) error {
	return httpx.OK(w, map[string]any{
		"password": true,
		"google":   map[string]any{"enabled": a.cfg.GoogleClientID != "", "client_id": a.cfg.GoogleClientID},
		"phone":    map[string]any{"enabled": a.sms != nil},
	})
}

// loginGoogle: access token dari popup Google → email terverifikasi →
// akun dengan email yang sama.
func (a *App) loginGoogle(w http.ResponseWriter, r *http.Request) error {
	if a.cfg.GoogleClientID == "" {
		return httpx.NewError(http.StatusServiceUnavailable, "TIDAK_TERSEDIA", "login Google belum dikonfigurasi")
	}
	var in struct {
		AccessToken string `json:"access_token"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	email, err := a.google.Verify(r.Context(), in.AccessToken)
	if err != nil {
		slog.Info("login Google ditolak", "err", err)
		return httpx.Unauthorized("login Google gagal — coba lagi")
	}
	u, err := a.store.GetUserByEmail(r.Context(), email)
	if store.IsNotFound(err) {
		return &httpx.Error{Status: http.StatusUnauthorized, Code: "BELUM_TERDAFTAR",
			Message: fmt.Sprintf("akun Google %s belum terdaftar — silakan daftar terlebih dahulu", email)}
	}
	if err != nil {
		return err
	}
	return a.issueSession(w, r, u, "Google")
}

// requestOTP mengirim kode 6 digit ke nomor terdaftar. Respons sama baik
// nomor terdaftar maupun tidak, agar daftar nomor tidak bisa ditebak.
func (a *App) requestOTP(w http.ResponseWriter, r *http.Request) error {
	if a.sms == nil {
		return httpx.NewError(http.StatusServiceUnavailable, "TIDAK_TERSEDIA", "login dengan nomor telepon belum dikonfigurasi")
	}
	var in struct {
		Phone string `json:"phone"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	phone, err := auth.NormalizePhone(in.Phone)
	if err != nil {
		return httpx.Unprocessable("nomor telepon tidak valid — contoh: 0812 3456 7890")
	}
	resp := map[string]any{
		"masked":       auth.MaskPhone(phone),
		"expires_in":   int(auth.OTPTTL.Seconds()),
		"resend_after": int(auth.OTPResendInterval.Seconds()),
		"code_length":  auth.OTPLength,
	}

	u, err := a.store.GetUserByPhone(r.Context(), phone)
	if store.IsNotFound(err) || (err == nil && !u.IsActive) {
		return httpx.OK(w, resp)
	}
	if err != nil {
		return err
	}
	last, err := a.store.LastOTPAt(r.Context(), u.ID)
	if err != nil {
		return err
	}
	// Masih dalam jeda kirim ulang: diam-diam tidak mengirim. Respons tetap
	// sama supaya tidak membocorkan bahwa nomor ini terdaftar; hitung mundur
	// ditampilkan oleh halaman login dari resend_after.
	if !last.IsZero() && time.Since(last) < auth.OTPResendInterval {
		return httpx.OK(w, resp)
	}
	// Batas kirim per jam: kirim ulang tidak boleh jadi jalan menebak tanpa batas.
	if n, err := a.store.OTPsSince(r.Context(), u.ID, time.Hour); err != nil {
		return err
	} else if n >= otpMaxPerHour {
		return httpx.OK(w, resp)
	}
	code, err := auth.GenerateOTP()
	if err != nil {
		return err
	}
	if err := a.store.CreateOTP(r.Context(), u.ID, auth.HashOTP(a.otpKey, u.ID, code), auth.OTPTTL); err != nil {
		return err
	}
	msg := fmt.Sprintf("Kode masuk TransCPG-X: %s. Berlaku %d menit. Jangan berikan kode ini kepada siapa pun.", code, int(auth.OTPTTL.Minutes()))
	if err := a.sms.Send(r.Context(), phone, msg); err != nil {
		return err
	}
	return httpx.OK(w, resp)
}

// verifyOTP: kode benar → sesi. Kode sekali pakai; 5 kali salah → hangus.
func (a *App) verifyOTP(w http.ResponseWriter, r *http.Request) error {
	if a.sms == nil {
		return httpx.NewError(http.StatusServiceUnavailable, "TIDAK_TERSEDIA", "login dengan nomor telepon belum dikonfigurasi")
	}
	var in struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	invalid := httpx.Unauthorized("kode salah atau sudah kedaluwarsa")
	phone, err := auth.NormalizePhone(in.Phone)
	if err != nil {
		return invalid
	}
	u, err := a.store.GetUserByPhone(r.Context(), phone)
	if store.IsNotFound(err) {
		return invalid
	}
	if err != nil {
		return err
	}
	// Batas lintas kode: maksimal 10 kode salah per jam per akun.
	key := fmt.Sprintf("otp:%d", u.ID)
	if n, err := a.store.LoginFailures(r.Context(), key, time.Hour); err != nil {
		return err
	} else if n >= otpMaxFailuresPerHour {
		return invalid
	}
	otp, err := a.store.ActiveOTP(r.Context(), u.ID)
	if store.IsNotFound(err) {
		return invalid
	}
	if err != nil {
		return err
	}
	// Jatah percobaan diambil SEBELUM kode diperiksa (atomik), jadi request
	// paralel tidak bisa menebak lebih dari batas. Pesan galat selalu sama
	// supaya tidak membocorkan apakah nomor terdaftar.
	if ok, err := a.store.ClaimOTPAttempt(r.Context(), otp.ID, auth.OTPMaxAttempts); err != nil {
		return err
	} else if !ok {
		return invalid
	}
	if !auth.CheckOTP(a.otpKey, u.ID, in.Code, otp.CodeHash) {
		_ = a.store.RecordLoginFailure(r.Context(), key)
		return invalid
	}
	if ok, err := a.store.ConsumeOTP(r.Context(), otp.ID); err != nil {
		return err
	} else if !ok {
		return invalid
	}
	return a.issueSession(w, r, u, "nomor telepon")
}

const (
	otpMaxPerHour         = 5
	otpMaxFailuresPerHour = 10
)
