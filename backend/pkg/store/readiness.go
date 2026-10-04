package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// Readiness mengumpulkan fakta 7 syarat untuk satu CP lalu menilainya.
func (s *Store) Readiness(ctx context.Context, pathwayID int64) (domain.Readiness, error) {
	return readiness(ctx, s.pool, pathwayID)
}

func readiness(ctx context.Context, q querier, pathwayID int64) (domain.Readiness, error) {
	in := domain.ReadinessInput{
		EpisodesBySeverity: map[int]int{},
		PlanRowsBySeverity: map[int]int{},
	}
	var cbg string
	var split *string
	if err := q.QueryRow(ctx, `select cbg_code, split_decision from clinical_pathways where id = $1`, pathwayID).
		Scan(&cbg, &split); err != nil {
		return domain.Readiness{}, err
	}
	in.NeedsSplit = split != nil && *split == "PECAH"

	err := q.QueryRow(ctx, `
		select count(*) from pathway_references where pathway_id = $1 and sub_cp_id is null`, pathwayID).
		Scan(&in.AcuanCount)
	if err != nil {
		return domain.Readiness{}, err
	}

	if err := scanStrings(ctx, q, &in.SubCPsWithoutAcuan, `
		select sc.icd10_code || ' ' || sc.label from pathway_sub_cps sc
		where sc.pathway_id = $1
		  and not exists (select 1 from pathway_references r where r.sub_cp_id = sc.id)
		order by sc.icd10_code`, pathwayID); err != nil {
		return domain.Readiness{}, err
	}

	if err := scanIntMap(ctx, q, in.EpisodesBySeverity, `
		select severity, episodes from pathway_severity_stats where pathway_id = $1`, pathwayID); err != nil {
		return domain.Readiness{}, err
	}
	if err := scanIntMap(ctx, q, in.PlanRowsBySeverity, `
		select severity, count(*) from clinical_plan_items where pathway_id = $1 group by severity`, pathwayID); err != nil {
		return domain.Readiness{}, err
	}

	if err := scanStrings(ctx, q, &in.ProceduresWithoutKPTL, `
		select distinct p.icd9cm_code from claim_episodes e
		join claim_episode_procedures p on p.episode_id = e.id
		where e.cbg_code = $1
		  and not exists (select 1 from kptl_mappings k where k.icd9cm_code = p.icd9cm_code)
		order by 1`, cbg); err != nil {
		return domain.Readiness{}, err
	}

	rows, _ := q.Query(ctx, `
		select s.sev from generate_series(1, 3) s(sev)
		where not exists (select 1 from ina_cbg_tariffs t where t.cbg_code = $1 and t.severity = s.sev)`, cbg)
	if in.SeveritiesWithoutTariff, err = pgx.CollectRows(rows, pgx.RowTo[int]); err != nil {
		return domain.Readiness{}, err
	}

	if err := scanStrings(ctx, q, &in.DiagnosesWithoutSnomed, `
		select distinct d.icd10_code from claim_episodes e
		join claim_episode_diagnoses d on d.episode_id = e.id
		where e.cbg_code = $1
		  and not exists (select 1 from snomed_mappings m where m.vocabulary = 'ICD10' and m.source_code = d.icd10_code)
		order by 1`, cbg); err != nil {
		return domain.Readiness{}, err
	}

	if err := scanStrings(ctx, q, &in.InvalidICD10, `
		select distinct d.icd10_code from claim_episodes e
		join claim_episode_diagnoses d on d.episode_id = e.id
		left join icd10_codes i on i.code = d.icd10_code
		where e.cbg_code = $1 and (i.code is null or not i.is_valid)
		order by 1`, cbg); err != nil {
		return domain.Readiness{}, err
	}

	return domain.EvaluateReadiness(in), nil
}

func scanStrings(ctx context.Context, q querier, dst *[]string, sql string, args ...any) error {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		*dst = append(*dst, v)
	}
	return rows.Err()
}

func scanIntMap(ctx context.Context, q querier, dst map[int]int, sql string, args ...any) error {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var k, v int
		if err := rows.Scan(&k, &v); err != nil {
			return err
		}
		dst[k] = v
	}
	return rows.Err()
}
