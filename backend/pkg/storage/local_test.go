package storage

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newLocal(t *testing.T) *Local {
	return &Local{Dir: t.TempDir(), Secret: []byte("rahasia-uji"), BaseURL: "/api/storage/local"}
}

func put(l *Local, token string, body []byte) error {
	r := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body))
	return l.ServeUpload(httptest.NewRecorder(), r, token)
}

func tokenOf(u string) string { return u[strings.LastIndex(u, "/")+1:] }

func TestLocalUploadDownload(t *testing.T) {
	l := newLocal(t)
	ctx := context.Background()
	up, _ := l.PresignUpload(ctx, "guidelines/1/a.pdf", "application/pdf")
	pdf := []byte("%PDF-1.7\nisi uji")
	if err := put(l, tokenOf(up.URL), pdf); err != nil {
		t.Fatal(err)
	}
	obj, err := l.Stat(ctx, "guidelines/1/a.pdf")
	if err != nil || obj.Size != int64(len(pdf)) {
		t.Fatalf("stat: %+v %v", obj, err)
	}
	dl, err := l.PresignDownload(ctx, "guidelines/1/a.pdf", "panduan.pdf", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	if err := l.ServeDownload(rec, httptest.NewRequest(http.MethodGet, "/", nil), tokenOf(dl)); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rec.Body.Bytes(), pdf) || !strings.Contains(rec.Header().Get("Content-Disposition"), "panduan.pdf") {
		t.Fatalf("unduh salah: %q", rec.Body.String())
	}
}

func TestLocalRejects(t *testing.T) {
	l := newLocal(t)
	ctx := context.Background()
	up, _ := l.PresignUpload(ctx, "guidelines/1/b.pdf", "application/pdf")
	tok := tokenOf(up.URL)

	if err := put(l, tok, []byte("bukan pdf")); !errors.Is(err, ErrNotPDF) {
		t.Errorf("bukan PDF harus ditolak, dapat %v", err)
	}
	if err := put(l, tok+"x", []byte("%PDF-1.7")); !errors.Is(err, ErrBadToken) {
		t.Errorf("token rusak harus ditolak, dapat %v", err)
	}
	// Token unduh tidak boleh dipakai untuk unggah.
	_ = put(l, tok, []byte("%PDF-1.7"))
	dl, _ := l.PresignDownload(ctx, "guidelines/1/b.pdf", "", time.Minute)
	if err := put(l, tokenOf(dl), []byte("%PDF-1.7")); !errors.Is(err, ErrBadToken) {
		t.Errorf("token GET dipakai untuk PUT harus ditolak, dapat %v", err)
	}
	big := append([]byte("%PDF-"), make([]byte, MaxAttachmentSize)...)
	up2, _ := l.PresignUpload(ctx, "guidelines/1/c.pdf", "application/pdf")
	if err := put(l, tokenOf(up2.URL), big); !errors.Is(err, ErrTooLarge) {
		t.Errorf("> 50 MB harus ditolak, dapat %v", err)
	}
	if _, err := l.Stat(ctx, "guidelines/1/c.pdf"); !errors.Is(err, ErrNotFound) {
		t.Error("file kebesaran tidak boleh tersisa")
	}
	if _, err := l.file("../../etc/passwd"); err == nil {
		p, _ := l.file("../../etc/passwd")
		if !strings.HasPrefix(p, l.Dir) {
			t.Error("path traversal harus tetap di dalam Dir")
		}
	}
}

func TestLocalNoOverwriteAndHead(t *testing.T) {
	l := newLocal(t)
	ctx := context.Background()
	up, _ := l.PresignUpload(ctx, "guidelines/9/x.pdf", "application/pdf")
	if err := put(l, tokenOf(up.URL), []byte("%PDF-1.7 asli")); err != nil {
		t.Fatal(err)
	}
	if err := put(l, tokenOf(up.URL), []byte("%PDF-1.7 timpa")); !errors.Is(err, ErrBadToken) {
		t.Errorf("token unggah tidak boleh dipakai ulang untuk menimpa, dapat %v", err)
	}
	head, err := l.Head(ctx, "guidelines/9/x.pdf", 5)
	if err != nil || string(head) != "%PDF-" {
		t.Errorf("head = %q, %v", head, err)
	}
}
