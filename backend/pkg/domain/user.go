package domain

import "time"

// User adalah pengguna yang sedang masuk atau dikelola Admin.
type User struct {
	ID          int64      `json:"id" db:"id"`
	HospitalID  *int64     `json:"hospital_id" db:"hospital_id"`
	FullName    string     `json:"full_name" db:"full_name"`
	Email       string     `json:"email" db:"email"`
	Phone       *string    `json:"phone" db:"phone"`
	Role        Role       `json:"role" db:"role"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
	// TokenVersion: versi sesi (lihat auth.IssueToken); tidak dikirim ke klien.
	TokenVersion int       `json:"-" db:"token_version"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// Pendaftaran mandiri
	RegistrationStatus RegistrationStatus `json:"registration_status" db:"registration_status"`
	SelfRegistered     bool               `json:"self_registered" db:"self_registered"`
	RegistrationNote   *string            `json:"registration_note" db:"registration_note"`
	RejectionReason    *string            `json:"rejection_reason" db:"rejection_reason"`
	ReviewedAt         *time.Time         `json:"reviewed_at" db:"reviewed_at"`
}
