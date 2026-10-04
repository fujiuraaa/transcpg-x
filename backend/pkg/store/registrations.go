package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// ErrEmailTaken: email sudah dipakai akun aktif atau pendaftaran yang masih menunggu.
var ErrEmailTaken = errors.New("email sudah terdaftar — silakan masuk, atau tunggu persetujuan bila baru mendaftar")

// ErrNotPending: permintaan sudah diproses Admin lain.
var ErrNotPending = errors.New("permintaan pendaftaran ini sudah diproses")

type NewRegistration struct {
	FullName      string
	Email         string
	Phone         *string
	RequestedRole domain.Role
	PasswordHash  string
	Note          *string
}

// Register membuat akun berstatus MENUNGGU (nonaktif) di RS bawaan. Pendaftar
// yang pernah ditolak boleh mengajukan ulang dengan email yang sama.
func (s *Store) Register(ctx context.Context, n NewRegistration) (domain.User, error) {
	var u domain.User
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		var status domain.RegistrationStatus
		var existingID int64
		err := tx.QueryRow(ctx, `select id, registration_status from app_users where lower(email) = lower($1) for update`, n.Email).
			Scan(&existingID, &status)
		var rows pgx.Rows
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			rows, _ = tx.Query(ctx, `
				insert into app_users (hospital_id, full_name, email, phone, password_hash, role, is_active,
				                       registration_status, self_registered, registration_note)
				values ((select id from hospitals where is_default), $1, $2, $3, $4, $5, false, 'MENUNGGU', true, $6)
				returning `+userColumns,
				n.FullName, n.Email, n.Phone, n.PasswordHash, n.RequestedRole, n.Note)
		case err != nil:
			return err
		case status == domain.RegRejected:
			rows, _ = tx.Query(ctx, `
				update app_users set full_name = $2, phone = $3, password_hash = $4, role = $5, is_active = false,
				  registration_status = 'MENUNGGU', registration_note = $6, rejection_reason = null,
				  reviewed_by = null, reviewed_at = null, created_at = now()
				where id = $1
				returning `+userColumns,
				existingID, n.FullName, n.Phone, n.PasswordHash, n.RequestedRole, n.Note)
		default:
			return ErrEmailTaken
		}
		u, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.User])
		return err
	})
	return u, err
}

// ListRegistrations: permintaan pendaftaran menurut status (bawaan MENUNGGU).
func (s *Store) ListRegistrations(ctx context.Context, status domain.RegistrationStatus, hospitalID *int64) ([]domain.User, error) {
	rows, _ := s.pool.Query(ctx, `
		select `+userColumns+` from app_users
		where self_registered and registration_status = $1 and ($2::bigint is null or hospital_id = $2)
		order by created_at`, status, hospitalID)
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
}

func (s *Store) CountPendingRegistrations(ctx context.Context, hospitalID *int64) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `select count(*) from app_users
		where registration_status = 'MENUNGGU' and ($1::bigint is null or hospital_id = $1)`, hospitalID).Scan(&n)
	return n, err
}

// ApproveRegistration mengaktifkan akun dengan peran yang ditetapkan Admin.
func (s *Store) ApproveRegistration(ctx context.Context, id int64, role domain.Role, reviewer domain.User) (domain.User, error) {
	rows, _ := s.pool.Query(ctx, `
		update app_users set role = $2, is_active = true, registration_status = 'DISETUJUI',
		  reviewed_by = $3, reviewed_at = now(), rejection_reason = null
		where id = $1 and registration_status = 'MENUNGGU'
		returning `+userColumns, id, role, reviewer.ID)
	u, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotPending
	}
	return u, err
}

// RejectRegistration menolak dengan alasan (ditampilkan ke pendaftar saat mencoba masuk).
func (s *Store) RejectRegistration(ctx context.Context, id int64, reason string, reviewer domain.User) (domain.User, error) {
	rows, _ := s.pool.Query(ctx, `
		update app_users set is_active = false, registration_status = 'DITOLAK',
		  rejection_reason = $2, reviewed_by = $3, reviewed_at = now()
		where id = $1 and registration_status = 'MENUNGGU'
		returning `+userColumns, id, reason, reviewer.ID)
	u, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotPending
	}
	return u, err
}
