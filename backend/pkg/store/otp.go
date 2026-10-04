package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// OTP adalah kode login telepon yang masih berlaku untuk satu pengguna.
type OTP struct {
	ID        int64     `db:"id"`
	CodeHash  string    `db:"code_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	Attempts  int16     `db:"attempts"`
	CreatedAt time.Time `db:"created_at"`
}

// LastOTPAt: waktu kode terakhir dibuat (untuk jeda kirim ulang); nol bila belum ada.
func (s *Store) LastOTPAt(ctx context.Context, userID int64) (time.Time, error) {
	var t *time.Time
	err := s.pool.QueryRow(ctx, `select max(created_at) from login_otps where user_id = $1`, userID).Scan(&t)
	if err != nil || t == nil {
		return time.Time{}, err
	}
	return *t, nil
}

// CreateOTP membatalkan kode lama yang belum dipakai lalu menyimpan kode baru.
func (s *Store) CreateOTP(ctx context.Context, userID int64, codeHash string, ttl time.Duration) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			update login_otps set consumed_at = now()
			where user_id = $1 and consumed_at is null`, userID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			insert into login_otps (user_id, code_hash, expires_at)
			values ($1, $2, now() + make_interval(secs => $3))`, userID, codeHash, ttl.Seconds())
		return err
	})
}

// ActiveOTP: kode terbaru yang belum dipakai dan belum kedaluwarsa.
func (s *Store) ActiveOTP(ctx context.Context, userID int64) (OTP, error) {
	rows, _ := s.pool.Query(ctx, `
		select id, code_hash, expires_at, attempts, created_at from login_otps
		where user_id = $1 and consumed_at is null and expires_at > now()
		order by created_at desc limit 1`, userID)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OTP])
}

// ClaimOTPAttempt memakai satu jatah percobaan secara atomik SEBELUM kode
// diperiksa, sehingga request paralel tidak bisa melewati batas. false =
// jatah habis / kode sudah dipakai / kedaluwarsa.
func (s *Store) ClaimOTPAttempt(ctx context.Context, id int64, max int) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		update login_otps set attempts = attempts + 1
		where id = $1 and attempts < $2 and consumed_at is null and expires_at > now()`, id, max)
	return tag.RowsAffected() == 1, err
}

// ConsumeOTP menandai kode sudah dipakai (sekali pakai). false bila sudah
// dipakai request lain lebih dulu.
func (s *Store) ConsumeOTP(ctx context.Context, id int64) (bool, error) {
	tag, err := s.pool.Exec(ctx, `update login_otps set consumed_at = now() where id = $1 and consumed_at is null`, id)
	return tag.RowsAffected() == 1, err
}

// OTPsSince: jumlah kode yang dibuat untuk pengguna dalam jendela waktu (batas kirim per jam).
func (s *Store) OTPsSince(ctx context.Context, userID int64, window time.Duration) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		select count(*) from login_otps where user_id = $1 and created_at > now() - make_interval(secs => $2)`,
		userID, window.Seconds()).Scan(&n)
	return n, err
}
