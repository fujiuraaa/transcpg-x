package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Supabase memakai Supabase Storage REST API dengan service role key
// (hanya di server). Bucket harus privat.
type Supabase struct {
	BaseURL    string // https://<project>.supabase.co
	ServiceKey string
	Bucket     string
	HTTP       *http.Client
}

func NewSupabase(baseURL, serviceKey, bucket string) *Supabase {
	return &Supabase{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		ServiceKey: serviceKey,
		Bucket:     bucket,
		HTTP:       &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Supabase) objectURL(kind, objectPath string) string {
	return fmt.Sprintf("%s/storage/v1/object/%s%s/%s", s.BaseURL, kind, url.PathEscape(s.Bucket), escapePath(objectPath))
}

func (s *Supabase) do(ctx context.Context, method, u string, body any) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, u, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.ServiceKey)
	req.Header.Set("apikey", s.ServiceKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return s.HTTP.Do(req)
}

func (s *Supabase) PresignUpload(ctx context.Context, objectPath, contentType string) (Upload, error) {
	res, err := s.do(ctx, http.MethodPost, s.objectURL("upload/sign/", objectPath), map[string]any{})
	if err != nil {
		return Upload{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Upload{}, fmt.Errorf("supabase presign upload: status %d", res.StatusCode)
	}
	var out struct {
		URL string `json:"url"` // /object/upload/sign/<bucket>/<path>?token=...
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return Upload{}, err
	}
	return Upload{
		URL:     s.BaseURL + "/storage/v1" + out.URL,
		Method:  http.MethodPut,
		Headers: map[string]string{"Content-Type": contentType, "x-upsert": "false"},
	}, nil
}

func (s *Supabase) PresignDownload(ctx context.Context, objectPath, filename string, ttl time.Duration) (string, error) {
	res, err := s.do(ctx, http.MethodPost, s.objectURL("sign/", objectPath), map[string]any{"expiresIn": int(ttl.Seconds())})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusBadRequest {
		return "", ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("supabase presign download: status %d", res.StatusCode)
	}
	var out struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	u := s.BaseURL + "/storage/v1" + out.SignedURL
	if filename != "" {
		u += "&download=" + url.QueryEscape(filename)
	}
	return u, nil
}

func (s *Supabase) Stat(ctx context.Context, objectPath string) (Object, error) {
	res, err := s.do(ctx, http.MethodHead, s.objectURL("authenticated/", objectPath), nil)
	if err != nil {
		return Object{}, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusBadRequest {
		return Object{}, ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return Object{}, fmt.Errorf("supabase stat: status %d", res.StatusCode)
	}
	size, _ := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
	return Object{Size: size, ContentType: res.Header.Get("Content-Type")}, nil
}

func (s *Supabase) Delete(ctx context.Context, objectPath string) error {
	res, err := s.do(ctx, http.MethodDelete, s.objectURL("", objectPath), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotFound {
		return fmt.Errorf("supabase delete: status %d", res.StatusCode)
	}
	return nil
}

func escapePath(p string) string {
	parts := strings.Split(p, "/")
	for i, s := range parts {
		parts[i] = url.PathEscape(s)
	}
	return strings.Join(parts, "/")
}

// Head mengambil n byte pertama lewat permintaan Range.
func (s *Supabase) Head(ctx context.Context, objectPath string, n int) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.objectURL("authenticated/", objectPath), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.ServiceKey)
	req.Header.Set("apikey", s.ServiceKey)
	req.Header.Set("Range", fmt.Sprintf("bytes=0-%d", n-1))
	res, err := s.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusBadRequest {
		return nil, ErrNotFound
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("supabase head: status %d", res.StatusCode)
	}
	return io.ReadAll(io.LimitReader(res.Body, int64(n)))
}
