package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// ErrRetiredConcept: konsep SNOMED-CT yang sudah pensiun tidak boleh dipilih.
var ErrRetiredConcept = errors.New("konsep SNOMED-CT sudah pensiun (retired) dan tidak dapat dipilih")

// SnomedSummary: ringkasan + cakupan per vokabuler.
func (s *Store) SnomedSummary(ctx context.Context) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		with src as (
		  select 'ICD10' as vocabulary, icd10_code as code from claim_episode_diagnoses
		  union
		  select 'ICD9CM', icd9cm_code from claim_episode_procedures
		)
		select src.vocabulary,
		       count(*)::int as kode,
		       count(m.id) filter (where m.status in ('OTOMATIS','MANUAL'))::int as dipadankan,
		       count(m.id) filter (where m.status = 'RAGU')::int as perlu_ditinjau,
		       count(*) filter (where m.id is null)::int as belum
		from src
		left join snomed_mappings m on m.vocabulary = src.vocabulary and m.source_code = src.code
		group by src.vocabulary order by src.vocabulary`)
}

// SnomedWorklist: filter "ragu" | "belum" | "" (semua), opsional per vokabuler.
func (s *Store) SnomedWorklist(ctx context.Context, filter, vocab string) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		with src as (
		  select distinct 'ICD10' as vocabulary, d.icd10_code as code, i.name
		  from claim_episode_diagnoses d left join icd10_codes i on i.code = d.icd10_code
		  union
		  select distinct 'ICD9CM', p.icd9cm_code, i.name
		  from claim_episode_procedures p left join icd9cm_codes i on i.code = p.icd9cm_code
		)
		select src.vocabulary, src.code, src.name, m.concept_id, c.preferred_term, c.semantic_tag, m.status, m.mapped_at
		from src
		left join snomed_mappings m on m.vocabulary = src.vocabulary and m.source_code = src.code
		left join snomed_concepts c on c.concept_id = m.concept_id
		where ($1 = '' or ($1 = 'ragu' and m.status = 'RAGU') or ($1 = 'belum' and m.id is null))
		  and ($2 = '' or src.vocabulary = $2)
		order by src.vocabulary, src.code limit 500`, filter, vocab)
}

// SetSnomedMapping menetapkan/mengganti padanan secara manual.
func (s *Store) SetSnomedMapping(ctx context.Context, vocab, code, conceptID string, u domain.User) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		var active bool
		if err := tx.QueryRow(ctx, `select active from snomed_concepts where concept_id = $1`, conceptID).Scan(&active); err != nil {
			return err
		}
		if !active {
			return ErrRetiredConcept
		}
		var old *string
		_ = tx.QueryRow(ctx, `select concept_id from snomed_mappings where vocabulary = $1 and source_code = $2 for update`,
			vocab, code).Scan(&old)
		if _, err := tx.Exec(ctx, `
			insert into snomed_mappings (vocabulary, source_code, concept_id, status, mapped_by)
			values ($1, $2, $3, 'MANUAL', $4)
			on conflict (vocabulary, source_code) do update set concept_id = excluded.concept_id,
			  status = 'MANUAL', mapped_by = excluded.mapped_by, mapped_at = now()`,
			vocab, code, conceptID, u.ID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			insert into mapping_audit_log (target, source_code, old_value, new_value, actor_id)
			values ('SNOMED', $1, $2, $3, $4)`, vocab+":"+code, old, conceptID, u.ID)
		return err
	})
}
