package store

import "context"

// ActivePathways: daftar CP berstatus AKTIF (dropdown halaman CP Aktif).
func (s *Store) ActivePathways(ctx context.Context) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select id, cbg_code, name, mdc_name, episodes, activated_at
		from v_pathway_overview where stage = 'AKTIF' order by name`)
}

// ActiveSummary: kartu ringkasan halaman CP Aktif.
func (s *Store) ActiveSummary(ctx context.Context) (Row, error) {
	return queryRow(ctx, s.pool, `
		select
		  count(*) filter (where stage = 'AKTIF')::int                      as cp_aktif,
		  coalesce(sum(episodes) filter (where stage = 'AKTIF'), 0)::int    as episode_tercakup,
		  count(distinct mdc_code) filter (where stage = 'AKTIF')::int      as kelompok_diagnosis,
		  count(*) filter (where stage = 'MENUNGGU_DIREKTUR')::int          as menunggu_direktur
		from v_pathway_overview`)
}

// ClaimPattern: diagnosis primer dan prosedur tersering per severity
// (tab "Pola Klaim" — konteks, bukan instruksi klinis).
func (s *Store) ClaimPattern(ctx context.Context, cbgCode string, severity int) (Row, error) {
	dx, err := queryRows(ctx, s.pool, `
		select d.icd10_code, i.name, count(*)::int as episodes
		from claim_episodes e
		join claim_episode_diagnoses d on d.episode_id = e.id and d.is_primary
		left join icd10_codes i on i.code = d.icd10_code
		where e.cbg_code = $1 and ($2 = 0 or e.severity = $2)
		group by d.icd10_code, i.name order by episodes desc limit 10`, cbgCode, severity)
	if err != nil {
		return nil, err
	}
	px, err := queryRows(ctx, s.pool, `
		select p.icd9cm_code, i.name, count(*)::int as episodes
		from claim_episodes e
		join claim_episode_procedures p on p.episode_id = e.id
		left join icd9cm_codes i on i.code = p.icd9cm_code
		where e.cbg_code = $1 and ($2 = 0 or e.severity = $2)
		group by p.icd9cm_code, i.name order by episodes desc limit 10`, cbgCode, severity)
	if err != nil {
		return nil, err
	}
	return Row{"diagnoses": dx, "procedures": px}, nil
}
