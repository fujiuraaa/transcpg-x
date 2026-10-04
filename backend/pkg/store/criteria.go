package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// Criterion adalah satu kriteria inklusi/eksklusi pasien.
type Criterion struct {
	ID          int64     `json:"id" db:"id"`
	Kind        string    `json:"kind" db:"kind"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type CriterionInput struct {
	Kind        string `json:"kind"` // INKLUSI | EKSKLUSI
	Description string `json:"description"`
}

func (s *Store) ListCriteria(ctx context.Context, pathwayID int64) ([]Criterion, error) {
	rows, _ := s.pool.Query(ctx, `
		select id, kind, description, created_at from pathway_criteria
		where pathway_id = $1 order by kind desc, id`, pathwayID)
	return pgx.CollectRows(rows, pgx.RowToStructByName[Criterion])
}

func (s *Store) AddCriterion(ctx context.Context, pathwayID int64, in CriterionInput, u domain.User) (Criterion, error) {
	var c Criterion
	err := s.editContent(ctx, pathwayID, domain.ContentKriteria, u, func(tx pgx.Tx, stage domain.Stage) error {
		rows, _ := tx.Query(ctx, `
			insert into pathway_criteria (pathway_id, kind, description, created_by)
			values ($1, $2, $3, $4) returning id, kind, description, created_at`,
			pathwayID, in.Kind, in.Description, u.ID)
		var err error
		if c, err = pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Criterion]); err != nil {
			return err
		}
		return logChange(ctx, tx, pathwayID, domain.ContentKriteria, "TAMBAH", stage, u, c)
	})
	return c, err
}

func (s *Store) DeleteCriterion(ctx context.Context, pathwayID, id int64, u domain.User) error {
	return s.editContent(ctx, pathwayID, domain.ContentKriteria, u, func(tx pgx.Tx, stage domain.Stage) error {
		rows, _ := tx.Query(ctx, `
			delete from pathway_criteria where id = $1 and pathway_id = $2
			returning id, kind, description, created_at`, id, pathwayID)
		c, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Criterion])
		if err != nil {
			return err
		}
		return logChange(ctx, tx, pathwayID, domain.ContentKriteria, "HAPUS", stage, u, c)
	})
}
