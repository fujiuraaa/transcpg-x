// Package storage menyimpan file (lampiran PDF panduan) di object storage.
// Browser mengunggah LANGSUNG ke storage memakai URL bertanda tangan, sehingga
// file besar (≤ 50 MB) tidak melewati fungsi Vercel yang dibatasi ±4,5 MB.
package storage

import (
	"context"
	"errors"
	"time"
)

// MaxAttachmentSize: batas ukuran lampiran PDF (50 MB).
const MaxAttachmentSize = 50 << 20

var (
	ErrNotFound = errors.New("file tidak ditemukan di penyimpanan")
	ErrTooLarge = errors.New("ukuran file melebihi 50 MB")
	ErrNotPDF   = errors.New("file bukan PDF")
)

// Upload menjelaskan cara browser mengunggah file secara langsung.
type Upload struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// Object adalah metadata file yang sudah tersimpan.
type Object struct {
	Size        int64
	ContentType string
}

// Storage diimplementasikan oleh driver Supabase (produksi) dan lokal (dev).
type Storage interface {
	// PresignUpload membuat URL unggah sekali pakai untuk objectPath.
	PresignUpload(ctx context.Context, objectPath, contentType string) (Upload, error)
	// PresignDownload membuat URL unduh sementara.
	PresignDownload(ctx context.Context, objectPath, filename string, ttl time.Duration) (string, error)
	// Stat memeriksa file setelah diunggah. ErrNotFound bila belum ada.
	Stat(ctx context.Context, objectPath string) (Object, error)
	// Delete menghapus file (abaikan bila sudah tidak ada).
	Delete(ctx context.Context, objectPath string) error
	// Head membaca n byte pertama file (memeriksa tanda "%PDF-").
	Head(ctx context.Context, objectPath string, n int) ([]byte, error)
}
