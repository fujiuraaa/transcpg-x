package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// KPTLSummary: jumlah prosedur, sudah/belum dipadankan, dan cakupan episode.
func (s *Store) KPTLSummary(ctx context.Context) (Row, error) {
	return queryRow(ctx, s.pool, `
		with proc as (
		  select p.icd9cm_code, count(distinct p.episode_id) as episodes,
		         exists (select 1 from kptl_mappings k where k.icd9cm_code = p.icd9cm_code) as mapped
		  from claim_episode_procedures p group by p.icd9cm_code
		)
		select count(*)::int as prosedur,
		       count(*) filter (where mapped)::int as sudah,
		       count(*) filter (where not mapped)::int as belum,
		       coalesce(round(100.0 * sum(episodes) filter (where mapped) / nullif(sum(episodes), 0), 1), 0)::float8 as cakupan_episode_pct
		from proc`)
}

// KPTLWorklist: tab "Belum Dipadankan" (onlyUnmapped) atau "Semua Prosedur".
func (s *Store) KPTLWorklist(ctx context.Context, onlyUnmapped bool, q string) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select p.icd9cm_code, i.name, count(distinct p.episode_id)::int as episodes,
		       k.kptl_code, kc.name as kptl_name, u.full_name as mapped_by_name, k.mapped_at
		from claim_episode_procedures p
		left join icd9cm_codes i on i.code = p.icd9cm_code
		left join kptl_mappings k on k.icd9cm_code = p.icd9cm_code
		left join kptl_codes kc on kc.code = k.kptl_code
		left join app_users u on u.id = k.mapped_by
		where (not $1 or k.icd9cm_code is null)
		  and ($2 = '' or p.icd9cm_code ilike $2 || '%' or i.name ilike '%' || $2 || '%')
		group by p.icd9cm_code, i.name, k.kptl_code, kc.name, u.full_name, k.mapped_at
		order by episodes desc limit 200`, onlyUnmapped, q)
}

// KPTLSuggestions: usulan mesin berdasarkan kemiripan nama (pg_trgm).
// Akurasinya rendah (±30% tepat di peringkat 1) — hanya petunjuk.
func (s *Store) KPTLSuggestions(ctx context.Context, icd9 string) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select k.code, k.name, similarity(k.name, i.name)::float8 as score
		from icd9cm_codes i, kptl_codes k
		where i.code = $1
		order by score desc limit 10`, icd9)
}

func (s *Store) SetKPTLMapping(ctx context.Context, icd9, kptl string, note *string, u domain.User) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		var old *string
		_ = tx.QueryRow(ctx, `select kptl_code from kptl_mappings where icd9cm_code = $1 for update`, icd9).Scan(&old)
		if _, err := tx.Exec(ctx, `
			insert into kptl_mappings (icd9cm_code, kptl_code, note, mapped_by) values ($1, $2, $3, $4)
			on conflict (icd9cm_code) do update set kptl_code = excluded.kptl_code, note = excluded.note,
			  mapped_by = excluded.mapped_by, mapped_at = now()`, icd9, kptl, note, u.ID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			insert into mapping_audit_log (target, source_code, old_value, new_value, actor_id)
			values ('KPTL', $1, $2, $3, $4)`, icd9, old, kptl, u.ID)
		return err
	})
}

func (s *Store) DeleteKPTLMapping(ctx context.Context, icd9 string, u domain.User) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		var old string
		if err := tx.QueryRow(ctx, `delete from kptl_mappings where icd9cm_code = $1 returning kptl_code`, icd9).Scan(&old); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			insert into mapping_audit_log (target, source_code, old_value, new_value, actor_id)
			values ('KPTL', $1, $2, null, $3)`, icd9, old, u.ID)
		return err
	})
}
