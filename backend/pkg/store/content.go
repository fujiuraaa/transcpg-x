package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// editContent menjalankan fn di dalam transaksi setelah mengunci baris CP
// (SELECT … FOR UPDATE) dan memeriksa wewenang mengubah isi. Penguncian
// mencegah isi berubah bersamaan dengan perpindahan tahap.
func (s *Store) editContent(ctx context.Context, pathwayID int64, kind domain.ContentKind, u domain.User,
	fn func(tx pgx.Tx, stage domain.Stage) error) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		var stage domain.Stage
		if err := tx.QueryRow(ctx, `select stage from clinical_pathways where id = $1 for update`, pathwayID).
			Scan(&stage); err != nil {
			return err
		}
		if err := domain.CheckEditContent(stage, kind, u.Role); err != nil {
			return err
		}
		return fn(tx, stage)
	})
}

// logChange mencatat perubahan isi CP ke pathway_change_log.
func logChange(ctx context.Context, tx pgx.Tx, pathwayID int64, kind domain.ContentKind, action string,
	stage domain.Stage, u domain.User, payload any) error {
	_, err := tx.Exec(ctx, `
		insert into pathway_change_log (pathway_id, entity, action, stage, payload, actor_id, actor_role)
		values ($1, $2, $3, $4, $5, $6, $7)`,
		pathwayID, kind, action, stage, payload, u.ID, u.Role)
	return err
}
