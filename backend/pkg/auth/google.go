package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrGoogleToken = errors.New("token Google tidak valid")

// GoogleVerifier memverifikasi access token dari popup Google Identity
// Services lewat endpoint tokeninfo, lalu mengembalikan email terverifikasi.
type GoogleVerifier struct {
	ClientID     string
	TokenInfoURL string // bawaan https://oauth2.googleapis.com/tokeninfo
	HTTP         *http.Client
}

func NewGoogleVerifier(clientID string) *GoogleVerifier {
	return &GoogleVerifier{
		ClientID:     clientID,
		TokenInfoURL: "https://oauth2.googleapis.com/tokeninfo",
		HTTP:         &http.Client{Timeout: 10 * time.Second},
	}
}

type tokenInfo struct {
	Aud           string `json:"aud"`
	Azp           string `json:"azp"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	ExpiresIn     string `json:"expires_in"`
}

// Verify memastikan token diterbitkan untuk aplikasi ini (aud/azp = client
// ID kita — mencegah token dari aplikasi lain dipakai di sini) dan email-nya
// sudah diverifikasi Google.
func (g *GoogleVerifier) Verify(ctx context.Context, accessToken string) (string, error) {
	if g == nil || g.ClientID == "" {
		return "", errors.New("login Google belum dikonfigurasi")
	}
	if strings.TrimSpace(accessToken) == "" {
		return "", ErrGoogleToken
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		g.TokenInfoURL+"?access_token="+url.QueryEscape(accessToken), nil)
	if err != nil {
		return "", err
	}
	res, err := g.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("menghubungi Google: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", ErrGoogleToken
	}
	var ti tokenInfo
	if err := json.NewDecoder(res.Body).Decode(&ti); err != nil {
		return "", ErrGoogleToken
	}
	if ti.Aud != g.ClientID && ti.Azp != g.ClientID {
		return "", ErrGoogleToken
	}
	if ti.Email == "" || ti.EmailVerified != "true" {
		return "", ErrGoogleToken
	}
	return strings.ToLower(ti.Email), nil
}
