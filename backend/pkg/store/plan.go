package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// PlanItem adalah satu baris rencana isi klinis.
type PlanItem struct {
	ID              int64     `json:"id" db:"id"`
	Severity        int16     `json:"severity" db:"severity"`
	Day             int16     `json:"day" db:"day"`
	ItemType        string    `json:"item_type" db:"item_type"`
	ItemCode        string    `json:"item_code" db:"item_code"`
	ItemName        string    `json:"item_name" db:"item_name"`
	Dose            *string   `json:"dose" db:"dose"`
	Route           *string   `json:"route" db:"route"`
	Frequency       *string   `json:"frequency" db:"frequency"`
	Duration        *string   `json:"duration" db:"duration"`
	Nature          string    `json:"nature" db:"nature"`
	GuidelineItemID *int64    `json:"guideline_item_id" db:"guideline_item_id"`
	OutsideGuidance bool      `json:"outside_guidance" db:"outside_guidance"` // tanda "DI LUAR PANDUAN"
	SourceNote      *string   `json:"source_note" db:"source_note"`
	Note            *string   `json:"note" db:"note"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

const planColumns = `id, severity, day, item_type, item_code, item_name, dose, route, frequency, duration,
	nature, guideline_item_id, guideline_item_id is null as outside_guidance, source_note, note, updated_at`

// PlanItemInput adalah isi formulir tambah/ubah baris rencana.
type PlanItemInput struct {
	Severity        int16   `json:"severity"`
	Day             int16   `json:"day"`
	ItemType        string  `json:"item_type"`
	ItemCode        string  `json:"item_code"`
	ItemName        string  `json:"item_name"`
	Dose            *string `json:"dose"`
	Route           *string `json:"route"`
	Frequency       *string `json:"frequency"`
	Duration        *string `json:"duration"`
	Nature          string  `json:"nature"`
	GuidelineItemID *int64  `json:"guideline_item_id"`
	SourceNote      *string `json:"source_note"`
	Note            *string `json:"note"`
}

// ListPlan mengembalikan rencana isi klinis; severity 0 = semua severity.
func (s *Store) ListPlan(ctx context.Context, pathwayID int64, severity int) ([]PlanItem, error) {
	rows, _ := s.pool.Query(ctx, `select `+planColumns+` from clinical_plan_items
		where pathway_id = $1 and ($2 = 0 or severity = $2)
		order by severity, day, item_type, item_name`, pathwayID, severity)
	return pgx.CollectRows(rows, pgx.RowToStructByName[PlanItem])
}

// Aturan "obat wajib lengkap", "kode harus sah", dan "tidak boleh ganda"
// ditegakkan oleh constraint & trigger database (migrasi 003).

func (s *Store) AddPlanItem(ctx context.Context, pathwayID int64, in PlanItemInput, u domain.User) (PlanItem, error) {
	var item PlanItem
	err := s.editContent(ctx, pathwayID, domain.ContentRencana, u, func(tx pgx.Tx, stage domain.Stage) error {
		if err := checkGuidelineItem(ctx, tx, pathwayID, in.GuidelineItemID); err != nil {
			return err
		}
		rows, _ := tx.Query(ctx, `
			insert into clinical_plan_items (pathway_id, severity, day, item_type, item_code, item_name, dose, route,
			  frequency, duration, nature, guideline_item_id, source_note, note, created_by, updated_by)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $15)
			returning `+planColumns,
			pathwayID, in.Severity, in.Day, in.ItemType, in.ItemCode, in.ItemName, in.Dose, in.Route,
			in.Frequency, in.Duration, in.Nature, in.GuidelineItemID, in.SourceNote, in.Note, u.ID)
		var err error
		if item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[PlanItem]); err != nil {
			return err
		}
		return logChange(ctx, tx, pathwayID, domain.ContentRencana, "TAMBAH", stage, u, item)
	})
	return item, err
}

func (s *Store) UpdatePlanItem(ctx context.Context, pathwayID, itemID int64, in PlanItemInput, u domain.User) (PlanItem, error) {
	var item PlanItem
	err := s.editContent(ctx, pathwayID, domain.ContentRencana, u, func(tx pgx.Tx, stage domain.Stage) error {
		if err := checkGuidelineItem(ctx, tx, pathwayID, in.GuidelineItemID); err != nil {
			return err
		}
		rows, _ := tx.Query(ctx, `
			update clinical_plan_items set
			  severity = $3, day = $4, item_type = $5, item_code = $6, item_name = $7, dose = $8, route = $9,
			  frequency = $10, duration = $11, nature = $12, guideline_item_id = $13, source_note = $14,
			  note = $15, updated_by = $16
			where id = $1 and pathway_id = $2
			returning `+planColumns,
			itemID, pathwayID, in.Severity, in.Day, in.ItemType, in.ItemCode, in.ItemName, in.Dose, in.Route,
			in.Frequency, in.Duration, in.Nature, in.GuidelineItemID, in.SourceNote, in.Note, u.ID)
		var err error
		if item, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[PlanItem]); err != nil {
			return err
		}
		return logChange(ctx, tx, pathwayID, domain.ContentRencana, "UBAH", stage, u, item)
	})
	return item, err
}

func (s *Store) DeletePlanItem(ctx context.Context, pathwayID, itemID int64, u domain.User) error {
	return s.editContent(ctx, pathwayID, domain.ContentRencana, u, func(tx pgx.Tx, stage domain.Stage) error {
		rows, _ := tx.Query(ctx, `delete from clinical_plan_items where id = $1 and pathway_id = $2 returning `+planColumns,
			itemID, pathwayID)
		item, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[PlanItem])
		if err != nil {
			return err
		}
		return logChange(ctx, tx, pathwayID, domain.ContentRencana, "HAPUS", stage, u, item)
	})
}

// ErrGuidelineNotReference: butir panduan bukan dari dokumen acuan CP ini.
var ErrGuidelineNotReference = errors.New("butir panduan harus berasal dari dokumen yang sudah ditetapkan sebagai acuan CP ini")

// checkGuidelineItem memastikan butir yang dirujuk berasal dari dokumen acuan
// CP ini; kalau tidak, tanda "DI LUAR PANDUAN" bisa dihapus secara keliru.
func checkGuidelineItem(ctx context.Context, tx pgx.Tx, pathwayID int64, itemID *int64) error {
	if itemID == nil {
		return nil
	}
	var ok bool
	if err := tx.QueryRow(ctx, `
		select exists (
		  select 1 from guideline_items i join pathway_references r on r.document_id = i.document_id
		  where i.id = $1 and r.pathway_id = $2)`, *itemID, pathwayID).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return ErrGuidelineNotReference
	}
	return nil
}
