package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// PathwayHeader adalah identitas dan tahap satu CP.
type PathwayHeader struct {
	ID                 int64        `json:"id" db:"id"`
	CBGCode            string       `json:"cbg_code" db:"cbg_code"`
	Name               string       `json:"name" db:"name"`
	MDCCode            string       `json:"mdc_code" db:"mdc_code"`
	MDCName            string       `json:"mdc_name" db:"mdc_name"`
	Stage              domain.Stage `json:"stage" db:"stage"`
	StageChangedAt     time.Time    `json:"stage_changed_at" db:"stage_changed_at"`
	ActivatedAt        *time.Time   `json:"activated_at" db:"activated_at"`
	SplitDecision      *string      `json:"split_decision" db:"split_decision"`
	HasPediatricCohort bool         `json:"has_pediatric_cohort" db:"has_pediatric_cohort"`
	Episodes           int32        `json:"episodes" db:"episodes"`
	DominantSeverity   *int16       `json:"dominant_severity" db:"dominant_severity"`
	TargetLOS          *float64     `json:"target_los" db:"target_los"`
	CandidateCount     int32        `json:"candidate_count" db:"candidate_count"`
}

func (s *Store) GetPathway(ctx context.Context, code string) (PathwayHeader, error) {
	rows, _ := s.pool.Query(ctx, `
		select id, cbg_code, name, mdc_code, mdc_name, stage, stage_changed_at, activated_at,
		       split_decision, has_pediatric_cohort, episodes, dominant_severity, target_los, candidate_count
		from v_pathway_overview where cbg_code = $1`, code)
	return pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[PathwayHeader])
}

// LibraryFilter adalah saringan halaman CP Library.
type LibraryFilter struct {
	Query string // kode, nama CP, atau kelompok diagnosis
	MDC   string
	Acuan string // "" | "sudah" | "belum"
	Sort  string // volume (bawaan) | los | kode | tanpa_acuan
	Page  int    // mulai 1
	// Profil RS untuk tarif INA-CBG dominan (kelas 3, severity dominan).
	Hospital Hospital
}

var librarySorts = map[string]string{
	"volume":      "episodes desc, cbg_code",
	"los":         "target_los desc nulls last, cbg_code",
	"kode":        "cbg_code",
	"tanpa_acuan": "(acuan_count = 0) desc, episodes desc",
}

// ListPathways mengembalikan satu halaman CP Library beserta total baris.
func (s *Store) ListPathways(ctx context.Context, f LibraryFilter) ([]Row, int, error) {
	order, ok := librarySorts[f.Sort]
	if !ok {
		order = librarySorts["volume"]
	}
	if f.Page < 1 {
		f.Page = 1
	}
	where := `
		where ($1 = '' or cbg_code ilike '%' || $1 || '%' or name ilike '%' || $1 || '%' or mdc_name ilike '%' || $1 || '%')
		  and ($2 = '' or mdc_code = $2)
		  and ($3 = '' or ($3 = 'sudah' and acuan_count > 0) or ($3 = 'belum' and acuan_count = 0))`
	args := []any{f.Query, f.MDC, f.Acuan}

	var total int
	if err := s.pool.QueryRow(ctx, `select count(*) from v_pathway_overview `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	// Peringkat volume dihitung atas seluruh CP, bukan hasil saringan.
	sql := fmt.Sprintf(`
		select v.*, tf.tariff as dominant_tariff from (
		  select rank() over (order by episodes desc)::int as volume_rank, * from v_pathway_overview
		) v
		left join lateral (
		  select t.tariff::float8 as tariff from ina_cbg_tariffs t
		  where t.cbg_code = v.cbg_code and t.severity = v.dominant_severity and t.care_class = 3
		    and t.bpjs_regional = $4 and t.hospital_type = $5 and t.ownership = $6
		) tf on true
		%s
		order by %s
		limit %d offset %d`, where, order, domain.LibraryPageSize, (f.Page-1)*domain.LibraryPageSize)
	args = append(args, f.Hospital.BPJSRegional, f.Hospital.HospitalType, f.Hospital.Ownership)
	rows, err := queryRows(ctx, s.pool, sql, args...)
	return rows, total, err
}

func (s *Store) ListMDC(ctx context.Context) ([]Row, error) {
	return queryRows(ctx, s.pool, `select code, name from mdc_groups order by code`)
}

// --- Tab-tab halaman detail CP (baca saja) ---------------------------------------

func (s *Store) SeverityStats(ctx context.Context, pathwayID int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select severity, episodes, los_median::float8, los_p75::float8, target_los::float8, age_median::float8
		from pathway_severity_stats where pathway_id = $1 order by severity`, pathwayID)
}

func (s *Store) SubCPs(ctx context.Context, pathwayID int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select sc.id, sc.icd10_code, sc.label, sc.episode_share::float8,
		       coalesce(json_agg(json_build_object('id', r.id, 'document_id', d.id, 'title', d.title))
		                filter (where r.id is not null), '[]') as references
		from pathway_sub_cps sc
		left join pathway_references r on r.sub_cp_id = sc.id
		left join guideline_documents d on d.id = r.document_id
		where sc.pathway_id = $1
		group by sc.id order by sc.episode_share desc nulls last`, pathwayID)
}

// Diagnoses: diagnosis primer & sekunder per severity beserta padanan SNOMED-CT.
func (s *Store) Diagnoses(ctx context.Context, cbgCode string) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select e.severity, d.icd10_code, i.name, d.is_primary, count(*)::int as episodes,
		       sm.concept_id as snomed_concept_id, sm.status as snomed_status
		from claim_episodes e
		join claim_episode_diagnoses d on d.episode_id = e.id
		left join icd10_codes i on i.code = d.icd10_code
		left join snomed_mappings sm on sm.vocabulary = 'ICD10' and sm.source_code = d.icd10_code
		where e.cbg_code = $1
		group by e.severity, d.icd10_code, i.name, d.is_primary, sm.concept_id, sm.status
		order by e.severity, d.is_primary desc, episodes desc`, cbgCode)
}

// Procedures: prosedur ICD-9-CM per severity beserta status padanan KPTL.
func (s *Store) Procedures(ctx context.Context, cbgCode string) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select e.severity, p.icd9cm_code, i.name, count(*)::int as episodes,
		       km.kptl_code, k.name as kptl_name
		from claim_episodes e
		join claim_episode_procedures p on p.episode_id = e.id
		left join icd9cm_codes i on i.code = p.icd9cm_code
		left join kptl_mappings km on km.icd9cm_code = p.icd9cm_code
		left join kptl_codes k on k.code = km.kptl_code
		where e.cbg_code = $1
		group by e.severity, p.icd9cm_code, i.name, km.kptl_code, k.name
		order by e.severity, episodes desc`, cbgCode)
}

// Tariffs: tarif INA-CBG 3 severity × 3 kelas pada profil RS h.
func (s *Store) Tariffs(ctx context.Context, cbgCode string, h Hospital) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select severity, care_class, tariff::float8, regulation
		from ina_cbg_tariffs
		where cbg_code = $1 and bpjs_regional = $2 and hospital_type = $3 and ownership = $4
		order by severity, care_class`, cbgCode, h.BPJSRegional, h.HospitalType, h.Ownership)
}

// CostDistribution: sebaran kelas rawat dan tarif klaim aktual per severity.
func (s *Store) CostDistribution(ctx context.Context, cbgCode string) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select severity, care_class, count(*)::int as episodes, avg(tariff)::float8 as avg_claim_tariff
		from claim_episodes where cbg_code = $1
		group by severity, care_class order by severity, care_class`, cbgCode)
}

func (s *Store) ScoringTools(ctx context.Context, pathwayID int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select t.* from pathway_scoring_tools pt
		join clinical_scoring_tools t on t.code = pt.tool_code
		where pt.pathway_id = $1 order by t.code`, pathwayID)
}

func (s *Store) RuleSuggestions(ctx context.Context, pathwayID int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select id, rule_type, condition, action, rationale
		from pathway_rule_suggestions where pathway_id = $1 order by rule_type, id`, pathwayID)
}

func (s *Store) QualityIndicators(ctx context.Context, pathwayID int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select id, code, name, target::float8, baseline::float8, unit
		from pathway_quality_indicators where pathway_id = $1 order by code`, pathwayID)
}

// ChangeLog: riwayat perubahan isi CP.
func (s *Store) ChangeLog(ctx context.Context, pathwayID int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select c.id, c.entity, c.action, c.stage, c.payload, c.actor_role, u.full_name as actor_name, c.created_at
		from pathway_change_log c join app_users u on u.id = c.actor_id
		where c.pathway_id = $1 order by c.created_at desc limit 200`, pathwayID)
}
