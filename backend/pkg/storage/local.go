package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Local menyimpan file di disk — HANYA untuk pengembangan (fungsi Vercel
// tidak punya disk permanen). URL unggah/unduh adalah endpoint API ini
// sendiri dengan token HMAC berumur pendek, meniru URL bertanda tangan.
type Local struct {
	Dir     string
	Secret  []byte
	BaseURL string // awalan URL endpoint, mis. "/api/storage/local"
}

type localToken struct {
	Path string `json:"p"`
	Op   string `json:"o"` // "put" | "get"
	Exp  int64  `json:"e"`
	Name string `json:"n,omitempty"`
}

var ErrBadToken = errors.New("token penyimpanan tidak valid atau kedaluwarsa")

func (l *Local) sign(t localToken) string {
	body, _ := json.Marshal(t)
	m := hmac.New(sha256.New, l.Secret)
	m.Write(body)
	return base64.RawURLEncoding.EncodeToString(body) + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func (l *Local) verify(token, op string) (localToken, error) {
	body64, sig64, ok := strings.Cut(token, ".")
	if !ok {
		return localToken{}, ErrBadToken
	}
	body, err1 := base64.RawURLEncoding.DecodeString(body64)
	sig, err2 := base64.RawURLEncoding.DecodeString(sig64)
	if err1 != nil || err2 != nil {
		return localToken{}, ErrBadToken
	}
	m := hmac.New(sha256.New, l.Secret)
	m.Write(body)
	if !hmac.Equal(sig, m.Sum(nil)) {
		return localToken{}, ErrBadToken
	}
	var t localToken
	if json.Unmarshal(body, &t) != nil || t.Op != op || time.Now().Unix() > t.Exp {
		return localToken{}, ErrBadToken
	}
	return t, nil
}

// file memetakan objectPath ke path disk, menolak keluar dari Dir.
func (l *Local) file(objectPath string) (string, error) {
	clean := filepath.Clean("/" + objectPath)
	if strings.Contains(clean, "..") {
		return "", ErrBadToken
	}
	return filepath.Join(l.Dir, clean), nil
}

func (l *Local) PresignUpload(_ context.Context, objectPath, contentType string) (Upload, error) {
	tok := l.sign(localToken{Path: objectPath, Op: "put", Exp: time.Now().Add(15 * time.Minute).Unix()})
	return Upload{URL: l.BaseURL + "/" + tok, Method: http.MethodPut, Headers: map[string]string{"Content-Type": contentType}}, nil
}

func (l *Local) PresignDownload(_ context.Context, objectPath, filename string, ttl time.Duration) (string, error) {
	if _, err := l.Stat(context.Background(), objectPath); err != nil {
		return "", err
	}
	return l.BaseURL + "/" + l.sign(localToken{Path: objectPath, Op: "get", Exp: time.Now().Add(ttl).Unix(), Name: filename}), nil
}

func (l *Local) Stat(_ context.Context, objectPath string) (Object, error) {
	f, err := l.file(objectPath)
	if err != nil {
		return Object{}, err
	}
	st, err := os.Stat(f)
	if errors.Is(err, os.ErrNotExist) {
		return Object{}, ErrNotFound
	}
	if err != nil {
		return Object{}, err
	}
	return Object{Size: st.Size(), ContentType: "application/pdf"}, nil
}

func (l *Local) Delete(_ context.Context, objectPath string) error {
	f, err := l.file(objectPath)
	if err != nil {
		return err
	}
	if err := os.Remove(f); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// ServeUpload menerima PUT dari browser: maks. 50 MB dan harus diawali "%PDF-".
func (l *Local) ServeUpload(w http.ResponseWriter, r *http.Request, token string) error {
	t, err := l.verify(token, "put")
	if err != nil {
		return err
	}
	f, err := l.file(t.Path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f), 0o750); err != nil {
		return err
	}
	// Token unggah sekali pakai: objek yang sudah ada tidak boleh ditimpa
	// (mis. setelah "complete" mencatat ukurannya).
	if _, err := os.Stat(f); err == nil {
		return ErrBadToken
	}
	body := http.MaxBytesReader(w, r.Body, MaxAttachmentSize)
	head := make([]byte, 5)
	if _, err := io.ReadFull(body, head); err != nil || !bytes.Equal(head, []byte("%PDF-")) {
		return ErrNotPDF
	}
	tmp := f + ".part"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, io.MultiReader(bytes.NewReader(head), body))
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(tmp)
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return ErrTooLarge
		}
		return err
	}
	return os.Rename(tmp, f)
}

// ServeDownload mengirim file untuk token GET yang sah.
func (l *Local) ServeDownload(w http.ResponseWriter, r *http.Request, token string) error {
	t, err := l.verify(token, "get")
	if err != nil {
		return err
	}
	f, err := l.file(t.Path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(f); err != nil {
		return ErrNotFound
	}
	w.Header().Set("Content-Type", "application/pdf")
	if t.Name != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", t.Name))
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, f)
	return nil
}

func (l *Local) Head(_ context.Context, objectPath string, n int) ([]byte, error) {
	f, err := l.file(objectPath)
	if err != nil {
		return nil, err
	}
	fh, err := os.Open(f)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	buf := make([]byte, n)
	k, err := io.ReadFull(fh, buf)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	return buf[:k], nil
}
