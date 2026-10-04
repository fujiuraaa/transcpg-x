package auth

import (
	"context"
	"log/slog"
)

// SMSSender mengirim pesan singkat. Implementasi produksi (gateway SMS/WA)
// ditambahkan sesuai penyedia yang dipilih RS.
type SMSSender interface {
	Send(ctx context.Context, phone, message string) error
}

// LogSender hanya untuk pengembangan: pesan (termasuk kode OTP) ditulis ke log.
// Jangan dipakai di produksi.
type LogSender struct{}

func (LogSender) Send(_ context.Context, phone, message string) error {
	slog.Warn("SMS (mode pengembangan — tidak benar-benar dikirim)", "ke", phone, "pesan", message)
	return nil
}
