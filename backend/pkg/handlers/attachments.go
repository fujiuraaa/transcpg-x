package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"transcpg-x/backend/pkg/auth"
	"transcpg-x/backend/pkg/httpx"
	"transcpg-x/backend/pkg/storage"
	"transcpg-x/backend/pkg/store"
)

// Lampiran PDF dokumen panduan (maks. 50 MB). Alur:
//  1. POST …/attachment/upload-url → URL unggah bertanda tangan
//  2. browser mengunggah LANGSUNG ke storage (tidak lewat fungsi Vercel)
//  3. POST …/attachment/complete   → server memverifikasi lalu mencatat
//
// Membuka PDF memakai URL unduh sementara (bucket privat).
func (a *App) routeAttachments(m *http.ServeMux) {
	m.HandleFunc("GET /api/guidelines/{id}/attachment", h(a.getAttachment))
	m.HandleFunc("POST /api/guidelines/{id}/attachment/upload-url", h(a.attachmentUploadURL))
	m.HandleFunc("POST /api/guidelines/{id}/attachment/complete", h(a.attachmentComplete))
	m.HandleFunc("DELETE /api/guidelines/{id}/attachment", h(a.deleteAttachment))
}

// routeLocalStorage: endpoint publik ber-token untuk driver lokal (dev).
// Tidak memakai JWT karena tautan unduh dibuka di tab baru; keamanannya dari
// token HMAC berumur pendek yang diterbitkan endpoint di atas.
func (a *App) routeLocalStorage(m *http.ServeMux) {
	local, ok := a.files.(*storage.Local)
	if !ok {
		return
	}
	m.HandleFunc("PUT /api/storage/local/{token}", h(func(w http.ResponseWriter, r *http.Request) error {
		if err := local.ServeUpload(w, r, r.PathValue("token")); err != nil {
			return err
		}
		return httpx.NoContent(w)
	}))
	m.HandleFunc("GET /api/storage/local/{token}", h(func(w http.ResponseWriter, r *http.Request) error {
		return local.ServeDownload(w, r, r.PathValue("token"))
	}))
}

func (a *App) requireFiles() error {
	if a.files == nil {
		return httpx.NewError(http.StatusServiceUnavailable, "TIDAK_TERSEDIA", "penyimpanan lampiran belum dikonfigurasi (STORAGE_DRIVER)")
	}
	return nil
}

// guidelineForAttachment memastikan dokumen ada dan peran boleh mengelola lampiran.
func (a *App) guidelineForAttachment(r *http.Request, write bool) (int64, error) {
	if err := a.requireFiles(); err != nil {
		return 0, err
	}
	if write && !canRegisterGuideline(auth.CurrentUser(r).Role) {
		return 0, httpx.Forbidden("peran Anda tidak dapat mengelola lampiran dokumen panduan")
	}
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return 0, err
	}
	if err := a.store.GuidelineExists(r.Context(), id); err != nil {
		return 0, err
	}
	if write {
		return id, a.guidelineWritable(r, id)
	}
	return id, nil
}

func (a *App) getAttachment(w http.ResponseWriter, r *http.Request) error {
	id, err := a.guidelineForAttachment(r, false)
	if err != nil {
		return err
	}
	att, err := a.store.GetAttachment(r.Context(), id)
	if err != nil {
		return err
	}
	out := map[string]any{"attachment": nil, "max_size": storage.MaxAttachmentSize}
	if att.Path != nil {
		url, err := a.files.PresignDownload(r.Context(), *att.Path, deref(att.Name), 10*time.Minute)
		if err != nil {
			return err
		}
		out["attachment"] = map[string]any{"name": att.Name, "size": att.Size, "uploaded_at": att.UploadedAt, "url": url}
	}
	return httpx.OK(w, out)
}

func (a *App) attachmentUploadURL(w http.ResponseWriter, r *http.Request) error {
	id, err := a.guidelineForAttachment(r, true)
	if err != nil {
		return err
	}
	var in struct {
		Filename    string `json:"filename"`
		Size        int64  `json:"size"`
		ContentType string `json:"content_type"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	switch {
	case in.Size <= 0:
		return httpx.Unprocessable("file kosong")
	case in.Size > storage.MaxAttachmentSize:
		return httpx.Unprocessable("ukuran file melebihi 50 MB")
	case in.ContentType != "application/pdf" || !strings.EqualFold(path.Ext(in.Filename), ".pdf"):
		return httpx.Unprocessable("hanya file PDF yang dapat diunggah")
	}
	objectPath := fmt.Sprintf("guidelines/%d/%s.pdf", id, randomHex(12))
	up, err := a.files.PresignUpload(r.Context(), objectPath, "application/pdf")
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"upload": up, "object_path": objectPath})
}

func (a *App) attachmentComplete(w http.ResponseWriter, r *http.Request) error {
	id, err := a.guidelineForAttachment(r, true)
	if err != nil {
		return err
	}
	var in struct {
		ObjectPath string `json:"object_path"`
		Filename   string `json:"filename"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		return err
	}
	// Path harus milik dokumen ini (diterbitkan oleh upload-url).
	prefix := fmt.Sprintf("guidelines/%d/", id)
	if !strings.HasPrefix(in.ObjectPath, prefix) || strings.Contains(in.ObjectPath, "..") || path.Ext(in.ObjectPath) != ".pdf" {
		return httpx.BadRequest("object_path tidak valid")
	}
	obj, err := a.files.Stat(r.Context(), in.ObjectPath)
	if err != nil {
		return err
	}
	head, err := a.files.Head(r.Context(), in.ObjectPath, 5)
	if err != nil {
		return err
	}
	if obj.Size > storage.MaxAttachmentSize || string(head) != "%PDF-" ||
		(obj.ContentType != "" && !strings.HasPrefix(obj.ContentType, "application/pdf")) {
		_ = a.files.Delete(r.Context(), in.ObjectPath)
		return httpx.Unprocessable("file bukan PDF atau melebihi 50 MB")
	}
	old, err := a.store.SetAttachment(r.Context(), id, in.ObjectPath, cleanFilename(in.Filename), obj.Size, auth.CurrentUser(r))
	if err != nil {
		return err
	}
	if old != nil && *old != in.ObjectPath {
		_ = a.files.Delete(r.Context(), *old) // lampiran lama diganti
	}
	verb, action := "Mengunggah", "UNGGAH_LAMPIRAN"
	if old != nil {
		verb, action = "Mengganti", "GANTI_LAMPIRAN"
	}
	a.record(r, store.Activity{Category: "DOKUMEN", Action: action, TargetType: "GUIDELINE", TargetID: id,
		TargetLabel: a.store.GuidelineTitle(r.Context(), id), Summary: verb + " lampiran PDF " + cleanFilename(in.Filename),
		Detail: map[string]any{"ukuran_byte": obj.Size}})
	return a.getAttachment(w, r)
}

func (a *App) deleteAttachment(w http.ResponseWriter, r *http.Request) error {
	id, err := a.guidelineForAttachment(r, true)
	if err != nil {
		return err
	}
	old, err := a.store.ClearAttachment(r.Context(), id)
	if err != nil {
		return err
	}
	if old != nil {
		_ = a.files.Delete(r.Context(), *old)
		a.record(r, store.Activity{Category: "DOKUMEN", Action: "HAPUS_LAMPIRAN", TargetType: "GUIDELINE", TargetID: id,
			TargetLabel: a.store.GuidelineTitle(r.Context(), id), Summary: "Menghapus lampiran PDF"})
	}
	return httpx.NoContent(w)
}

// cleanFilename: nama tampilan aman — tanpa folder/karakter kontrol, maks. 200, berakhiran .pdf.
func cleanFilename(name string) string {
	name = path.Base(strings.ReplaceAll(strings.TrimSpace(name), "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == '"' {
			return -1
		}
		return r
	}, name)
	if name == "" || name == "." || name == "/" {
		name = "lampiran.pdf"
	}
	if len(name) > 200 {
		name = name[len(name)-200:]
	}
	if !strings.EqualFold(path.Ext(name), ".pdf") {
		name += ".pdf"
	}
	return name
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
