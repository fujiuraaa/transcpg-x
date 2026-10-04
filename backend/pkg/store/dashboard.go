package store

import (
	"context"

	"transcpg-x/backend/pkg/domain"
)

// DashboardSummary adalah isi halaman Dashboard.
type DashboardSummary struct {
	Cards             Row   `json:"cards"`              // 6 kartu angka
	CasesByMDC        []Row `json:"cases_by_mdc"`       // grafik batang per kelompok diagnosis
	Priority          []Row `json:"priority"`           // tabel prioritas (volume terbanyak)
	GuidelineCoverage []Row `json:"guideline_coverage"` // butir acuan per kategori
	Progress          Row   `json:"progress"`           // 3 kartu kemajuan
}

func (s *Store) DashboardSummary(ctx context.Context) (DashboardSummary, error) {
	var d DashboardSummary
	var err error

	if d.Cards, err = queryRow(ctx, s.pool, `
		select
		  (select count(*) from clinical_pathways)::int                           as clinical_pathways,
		  (select count(*) from v_pathway_overview where acuan_count = 0)::int    as cp_tanpa_acuan,
		  (select count(*) from guideline_documents)::int                         as dokumen_panduan,
		  (select count(*) from guideline_items)::int                             as butir_acuan,
		  (select count(*) from tkmkb_topics)::int                                as topik_tkmkb,
		  (select count(*) from clinical_scoring_tools)::int                      as sistem_skoring`); err != nil {
		return d, err
	}

	if d.CasesByMDC, err = queryRows(ctx, s.pool, `
		select g.mdc_code, m.name as mdc_name, count(*)::int as episodes,
		       round(100.0 * count(*) / sum(count(*)) over (), 1)::float8 as pct
		from claim_episodes e
		join ina_cbg_groupers g on g.code = e.cbg_code
		join mdc_groups m on m.code = g.mdc_code
		group by g.mdc_code, m.name
		order by episodes desc`); err != nil {
		return d, err
	}

	if d.Priority, err = queryRows(ctx, s.pool, `
		select row_number() over (order by episodes desc)::int as rank,
		       cbg_code, name, mdc_name, episodes, stage, acuan_count
		from v_pathway_overview
		order by episodes desc
		limit 20`); err != nil {
		return d, err
	}

	if d.GuidelineCoverage, err = queryRows(ctx, s.pool, `
		select category, count(*)::int as items
		from guideline_items group by category order by items desc`); err != nil {
		return d, err
	}

	d.Progress, err = queryRow(ctx, s.pool, `
		select
		  (select count(*) from v_pathway_overview where acuan_count > 0)::int as acuan_ditetapkan,
		  (select count(*) from v_pathway_overview where acuan_count = 0)::int as menunggu_penetapan,
		  (select count(*) from v_grouper_volume where episodes < $1)::int     as dilewati`,
		domain.MinEpisodesEligible)
	return d, err
}
