package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// References: acuan standar CP (tingkat CP dan sub-CP).
func (s *Store) References(ctx context.Context, pathwayID int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select r.id, r.sub_cp_id, sc.icd10_code as sub_cp_icd10, r.document_id, d.title, d.source_type,
		       d.issuer, d.regulation_ref, d.year, r.note, u.full_name as assigned_by_name, r.assigned_at
		from pathway_references r
		join guideline_documents d on d.id = r.document_id
		join app_users u on u.id = r.assigned_by
		left join pathway_sub_cps sc on sc.id = r.sub_cp_id
		where r.pathway_id = $1
		order by r.sub_cp_id nulls first, r.assigned_at`, pathwayID)
}

type NewReference struct {
	DocumentID int64   `json:"document_id"`
	SubCPID    *int64  `json:"sub_cp_id"`
	Note       *string `json:"note"`
}

// AddReference menetapkan dokumen sebagai acuan (klik "Tetapkan").
func (s *Store) AddReference(ctx context.Context, pathwayID int64, n NewReference, u domain.User) (int64, error) {
	var id int64
	err := s.editContent(ctx, pathwayID, domain.ContentAcuan, u, func(tx pgx.Tx, stage domain.Stage) error {
		if n.SubCPID != nil {
			var ok bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from pathway_sub_cps where id = $1 and pathway_id = $2)`,
				*n.SubCPID, pathwayID).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return pgx.ErrNoRows
			}
		}
		if err := tx.QueryRow(ctx, `
			insert into pathway_references (pathway_id, sub_cp_id, document_id, note, assigned_by)
			values ($1, $2, $3, $4, $5) returning id`,
			pathwayID, n.SubCPID, n.DocumentID, n.Note, u.ID).Scan(&id); err != nil {
			return err
		}
		return logChange(ctx, tx, pathwayID, domain.ContentAcuan, "TAMBAH", stage, u,
			map[string]any{"reference_id": id, "document_id": n.DocumentID, "sub_cp_id": n.SubCPID})
	})
	return id, err
}

// RemoveReference mencabut satu acuan (klik "Cabut").
func (s *Store) RemoveReference(ctx context.Context, pathwayID, refID int64, u domain.User) error {
	return s.editContent(ctx, pathwayID, domain.ContentAcuan, u, func(tx pgx.Tx, stage domain.Stage) error {
		var docID int64
		var subID *int64
		if err := tx.QueryRow(ctx, `
			delete from pathway_references where id = $1 and pathway_id = $2
			returning document_id, sub_cp_id`, refID, pathwayID).Scan(&docID, &subID); err != nil {
			return err
		}
		return logChange(ctx, tx, pathwayID, domain.ContentAcuan, "HAPUS", stage, u,
			map[string]any{"reference_id": refID, "document_id": docID, "sub_cp_id": subID})
	})
}
