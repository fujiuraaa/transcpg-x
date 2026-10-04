package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// GuidelineCandidates: dokumen panduan yang cocok dengan diagnosis primer
// grouper, ditambah hasil pencarian bebas bila q diisi.
func (s *Store) GuidelineCandidates(ctx context.Context, cbgCode, q string) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		with dx as (
		  select array_agg(distinct d.icd10_code) as codes
		  from claim_episodes e
		  join claim_episode_diagnoses d on d.episode_id = e.id and d.is_primary
		  where e.cbg_code = $1
		)
		select g.id, g.title, g.source_type, g.issuer, g.regulation_ref, g.year, g.icd10_codes,
		       g.icd10_codes && dx.codes as matches_diagnosis,
		       g.attachment_path is not null as has_attachment, g.attachment_name, g.attachment_size,
		       (select count(*) from guideline_items i where i.document_id = g.id)::int as item_count
		from guideline_documents g, dx
		where g.icd10_codes && dx.codes or ($2 <> '' and g.title ilike '%' || $2 || '%')
		order by matches_diagnosis desc, g.source_type, g.year desc nulls last
		limit 100`, cbgCode, q)
}

func (s *Store) GuidelineItems(ctx context.Context, documentID int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select id, category, title, page, quote, icd10_codes
		from guideline_items where document_id = $1 order by category, id`, documentID)
}

type NewGuideline struct {
	Title         string   `json:"title"`
	SourceType    string   `json:"source_type"`
	Issuer        *string  `json:"issuer"`
	RegulationRef *string  `json:"regulation_ref"`
	Year          *int16   `json:"year"`
	ICD10Codes    []string `json:"icd10_codes"`
}

// CreateGuideline: "Masukkan dokumen panduan sendiri".
func (s *Store) CreateGuideline(ctx context.Context, n NewGuideline, u domain.User) (int64, error) {
	if n.ICD10Codes == nil {
		n.ICD10Codes = []string{}
	}
	var id int64
	err := s.pool.QueryRow(ctx, `
		insert into guideline_documents (title, source_type, issuer, regulation_ref, year, icd10_codes, created_by)
		values ($1, $2, $3, $4, $5, $6, $7) returning id`,
		n.Title, n.SourceType, n.Issuer, n.RegulationRef, n.Year, n.ICD10Codes, u.ID).Scan(&id)
	return id, err
}

type NewGuidelineItem struct {
	Category   string   `json:"category"`
	Title      string   `json:"title"`
	Page       string   `json:"page"`
	Quote      string   `json:"quote"`
	ICD10Codes []string `json:"icd10_codes"`
}

func (s *Store) AddGuidelineItem(ctx context.Context, documentID int64, n NewGuidelineItem) (int64, error) {
	if n.ICD10Codes == nil {
		n.ICD10Codes = []string{}
	}
	var id int64
	err := s.pool.QueryRow(ctx, `
		insert into guideline_items (document_id, category, title, page, quote, icd10_codes)
		values ($1, $2, $3, $4, $5, $6) returning id`,
		documentID, n.Category, n.Title, n.Page, n.Quote, n.ICD10Codes).Scan(&id)
	return id, err
}

// GuidelineExists dipakai handler sebelum menambah butir.
func (s *Store) GuidelineExists(ctx context.Context, id int64) error {
	var ok bool
	if err := s.pool.QueryRow(ctx, `select exists (select 1 from guideline_documents where id = $1)`, id).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return pgx.ErrNoRows
	}
	return nil
}

// Attachment adalah metadata lampiran PDF satu dokumen panduan.
type Attachment struct {
	Path       *string    `json:"-" db:"attachment_path"`
	Name       *string    `json:"name" db:"attachment_name"`
	Size       *int64     `json:"size" db:"attachment_size"`
	UploadedAt *time.Time `json:"uploaded_at" db:"attachment_uploaded_at"`
}

func (s *Store) GetAttachment(ctx context.Context, documentID int64) (Attachment, error) {
	rows, _ := s.pool.Query(ctx, `
		select attachment_path, attachment_name, attachment_size, attachment_uploaded_at
		from guideline_documents where id = $1`, documentID)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Attachment])
}

// SetAttachment mencatat lampiran baru; mengembalikan path lama (untuk dihapus).
func (s *Store) SetAttachment(ctx context.Context, documentID int64, path, name string, size int64, u domain.User) (*string, error) {
	var old *string
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `select attachment_path from guideline_documents where id = $1 for update`, documentID).Scan(&old); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			update guideline_documents set attachment_path = $2, attachment_name = $3, attachment_size = $4,
			  attachment_uploaded_at = now(), attachment_uploaded_by = $5
			where id = $1`, documentID, path, name, size, u.ID)
		return err
	})
	return old, err
}

// ClearAttachment melepas lampiran; mengembalikan path lama.
func (s *Store) ClearAttachment(ctx context.Context, documentID int64) (*string, error) {
	var old *string
	err := s.pool.QueryRow(ctx, `
		update guideline_documents g set attachment_path = null, attachment_name = null, attachment_size = null,
		  attachment_uploaded_at = null, attachment_uploaded_by = null
		from (select id, attachment_path from guideline_documents where id = $1 for update) prev
		where g.id = prev.id
		returning prev.attachment_path`, documentID).Scan(&old)
	return old, err
}

// GuidelineTitle untuk label jejak audit; "" bila tidak ditemukan.
func (s *Store) GuidelineTitle(ctx context.Context, id int64) string {
	var t string
	_ = s.pool.QueryRow(ctx, `select title from guideline_documents where id = $1`, id).Scan(&t)
	return t
}
