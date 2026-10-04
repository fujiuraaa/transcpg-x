package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Hospital adalah profil rumah sakit; kombinasi regional + tipe + kepemilikan
// + kelas rawat menentukan tarif INA-CBG yang ditampilkan.
type Hospital struct {
	ID               int64     `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	HospitalType     string    `json:"hospital_type" db:"hospital_type"`
	Ownership        string    `json:"ownership" db:"ownership"`
	BPJSRegional     int16     `json:"bpjs_regional" db:"bpjs_regional"`
	MaxCareClass     int16     `json:"max_care_class" db:"max_care_class"`
	BPJSProviderCode *string   `json:"bpjs_provider_code" db:"bpjs_provider_code"`
	Accreditation    *string   `json:"accreditation" db:"accreditation"`
	City             *string   `json:"city" db:"city"`
	Province         *string   `json:"province" db:"province"`
	Address          *string   `json:"address" db:"address"`
	Phone            *string   `json:"phone" db:"phone"`
	Email            *string   `json:"email" db:"email"`
	BedCapacity      *int32    `json:"bed_capacity" db:"bed_capacity"`
	ICUCapacity      *int32    `json:"icu_capacity" db:"icu_capacity"`
	IsDefault        bool      `json:"is_default" db:"is_default"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

func (s *Store) ListHospitals(ctx context.Context) ([]Hospital, error) {
	rows, _ := s.pool.Query(ctx, `select * from hospitals order by is_default desc, name`)
	return pgx.CollectRows(rows, pgx.RowToStructByName[Hospital])
}

func (s *Store) GetHospital(ctx context.Context, id int64) (Hospital, error) {
	rows, _ := s.pool.Query(ctx, `select * from hospitals where id = $1`, id)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Hospital])
}

// ResolveHospital mengembalikan profil id, atau profil bawaan bila id = 0.
func (s *Store) ResolveHospital(ctx context.Context, id int64) (Hospital, error) {
	rows, _ := s.pool.Query(ctx, `
		select * from hospitals where ($1 = 0 and is_default) or id = $1
		order by is_default desc limit 1`, id)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Hospital])
}

// UpdateHospital menyimpan Setup RS. ID dan is_default tidak berubah.
func (s *Store) UpdateHospital(ctx context.Context, h Hospital) (Hospital, error) {
	rows, _ := s.pool.Query(ctx, `
		update hospitals set
		  name = $2, hospital_type = $3, ownership = $4, bpjs_regional = $5, max_care_class = $6,
		  bpjs_provider_code = $7, accreditation = $8, city = $9, province = $10, address = $11,
		  phone = $12, email = $13, bed_capacity = $14, icu_capacity = $15
		where id = $1
		returning *`,
		h.ID, h.Name, h.HospitalType, h.Ownership, h.BPJSRegional, h.MaxCareClass,
		h.BPJSProviderCode, h.Accreditation, h.City, h.Province, h.Address,
		h.Phone, h.Email, h.BedCapacity, h.ICUCapacity)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Hospital])
}
