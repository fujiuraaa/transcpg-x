package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

const userColumns = `id, hospital_id, full_name, email, phone, role, is_active, last_login_at, created_at, token_version,
	registration_status, self_registered, registration_note, rejection_reason, reviewed_at`

func (s *Store) GetUser(ctx context.Context, id int64) (domain.User, error) {
	rows, _ := s.pool.Query(ctx, `select `+userColumns+` from app_users where id = $1`, id)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.User])
}

// GetUserForLogin mengembalikan pengguna beserta hash password-nya.
func (s *Store) GetUserForLogin(ctx context.Context, email string) (domain.User, string, error) {
	type row struct {
		domain.User
		PasswordHash string `db:"password_hash"`
	}
	rows, _ := s.pool.Query(ctx, `select `+userColumns+`, password_hash from app_users where lower(email) = lower($1)`, email)
	r, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[row])
	return r.User, r.PasswordHash, err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	rows, _ := s.pool.Query(ctx, `select `+userColumns+` from app_users where lower(email) = lower($1)`, email)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.User])
}

// GetUserByPhone: phone harus sudah dinormalisasi ke E.164.
func (s *Store) GetUserByPhone(ctx context.Context, phone string) (domain.User, error) {
	rows, _ := s.pool.Query(ctx, `select `+userColumns+` from app_users where phone = $1`, phone)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.User])
}

func (s *Store) TouchLogin(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `update app_users set last_login_at = now() where id = $1`, id)
	return err
}

// ListUsers: akun yang sudah disetujui (permintaan pendaftaran ada di
// ListRegistrations). hospitalID = lingkup Admin RS; nil = semua RS.
func (s *Store) ListUsers(ctx context.Context, hospitalID *int64) ([]domain.User, error) {
	rows, _ := s.pool.Query(ctx, `select `+userColumns+` from app_users
		where registration_status = 'DISETUJUI' and ($1::bigint is null or hospital_id = $1)
		order by full_name`, hospitalID)
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
}

type NewUser struct {
	HospitalID   *int64
	FullName     string
	Email        string
	Phone        *string
	PasswordHash string
	Role         domain.Role
}

func (s *Store) CreateUser(ctx context.Context, n NewUser) (domain.User, error) {
	rows, _ := s.pool.Query(ctx, `
		insert into app_users (hospital_id, full_name, email, phone, password_hash, role)
		values ($1, $2, $3, $4, $5, $6)
		returning `+userColumns,
		n.HospitalID, n.FullName, n.Email, n.Phone, n.PasswordHash, n.Role)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.User])
}

// UserPatch: field nil tidak diubah.
type UserPatch struct {
	FullName *string
	// Phone: nil = tidak diubah; pointer ke "" = hapus nomor.
	Phone        *string
	Role         *domain.Role
	IsActive     *bool
	PasswordHash *string
}

func (s *Store) UpdateUser(ctx context.Context, id int64, p UserPatch) (domain.User, error) {
	rows, _ := s.pool.Query(ctx, `
		update app_users set
		  full_name     = coalesce($2, full_name),
		  role          = coalesce($3, role),
		  is_active     = coalesce($4, is_active),
		  password_hash = coalesce($5, password_hash),
		  token_version = token_version + case when $5::text is null then 0 else 1 end,
		  phone         = case when $6::text is null then phone else nullif($6, '') end
		where id = $1
		returning `+userColumns,
		id, p.FullName, p.Role, p.IsActive, p.PasswordHash, p.Phone)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.User])
}

func (s *Store) DeleteUser(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `delete from app_users where id = $1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return err
}
