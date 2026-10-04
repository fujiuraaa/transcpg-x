package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// EvaluationRow adalah satu baris daftar Evaluasi CP.
type EvaluationRow struct {
	ID           int64         `json:"id"`
	CBGCode      string        `json:"cbg_code"`
	Name         string        `json:"name"`
	Stage        domain.Stage  `json:"stage"`
	ActivatedAt  *time.Time    `json:"activated_at"`
	PostEpisodes int           `json:"post_episodes"`
	Evaluable    bool          `json:"evaluable"` // sudah ada klaim sesudah aktif
	NeedsReview  bool          `json:"needs_review"`
	Flags        []domain.Flag `json:"flags"`
}

type EvaluationSummary struct {
	CPAktif     int        `json:"cp_aktif"`
	Evaluable   int        `json:"dapat_dievaluasi"`
	NeedsReview int        `json:"perlu_ditinjau"`
	ClaimsUpTo  *time.Time `json:"data_klaim_sampai"`
}

// Evaluations mengembalikan ringkasan + daftar CP. active=false → CP yang
// belum Aktif (belum bisa dievaluasi).
func (s *Store) Evaluations(ctx context.Context, active bool) (EvaluationSummary, []EvaluationRow, error) {
	var sum EvaluationSummary
	if err := s.pool.QueryRow(ctx, `select max(discharge_date)::timestamptz from claim_episodes`).Scan(&sum.ClaimsUpTo); err != nil {
		return sum, nil, err
	}

	type base struct {
		ID          int64        `db:"id"`
		CBGCode     string       `db:"cbg_code"`
		Name        string       `db:"name"`
		Stage       domain.Stage `db:"stage"`
		ActivatedAt *time.Time   `db:"activated_at"`
	}
	rows, _ := s.pool.Query(ctx, `
		select id, cbg_code, name, stage, activated_at from clinical_pathways
		where (stage = 'AKTIF') = $1 order by cbg_code`, active)
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[base])
	if err != nil {
		return sum, nil, err
	}

	out := make([]EvaluationRow, 0, len(list))
	for _, b := range list {
		r := EvaluationRow{ID: b.ID, CBGCode: b.CBGCode, Name: b.Name, Stage: b.Stage, ActivatedAt: b.ActivatedAt, Flags: []domain.Flag{}}
		if active && b.ActivatedAt != nil {
			in, err := s.evaluationInput(ctx, b.ID, b.CBGCode, *b.ActivatedAt)
			if err != nil {
				return sum, nil, err
			}
			r.PostEpisodes = in.PostEpisodes
			r.Evaluable = in.PostEpisodes > 0
			if f := domain.Evaluate(in); f != nil {
				r.Flags = f
			}
			r.NeedsReview = len(r.Flags) > 0
			sum.CPAktif++
			if r.Evaluable {
				sum.Evaluable++
			}
			if r.NeedsReview {
				sum.NeedsReview++
			}
		}
		out = append(out, r)
	}
	if !active {
		if err := s.pool.QueryRow(ctx, `select count(*) from clinical_pathways where stage = 'AKTIF'`).Scan(&sum.CPAktif); err != nil {
			return sum, nil, err
		}
	}
	return sum, out, nil
}

func (s *Store) evaluationInput(ctx context.Context, pathwayID int64, cbg string, activatedAt time.Time) (domain.EvaluationInput, error) {
	var in domain.EvaluationInput
	var baseTotal, baseSev3 float64
	err := s.pool.QueryRow(ctx, `
		select coalesce(sum(episodes), 0), coalesce(sum(episodes) filter (where severity = 3), 0)
		from pathway_severity_stats where pathway_id = $1`, pathwayID).Scan(&baseTotal, &baseSev3)
	if err != nil {
		return in, err
	}
	if baseTotal > 0 {
		in.BaselineSev3Share = baseSev3 / baseTotal
	}

	type group struct {
		Severity int       `db:"severity"`
		P75      *float64  `db:"p75"`
		LOS      []float64 `db:"los"`
	}
	rows, _ := s.pool.Query(ctx, `
		select st.severity, st.los_p75::float8 as p75,
		       coalesce(array_agg(e.los_days::float8) filter (where e.id is not null), '{}') as los
		from pathway_severity_stats st
		left join claim_episodes e
		  on e.cbg_code = $2 and e.severity = st.severity and e.discharge_date >= ($3::timestamptz at time zone 'Asia/Jakarta')::date
		where st.pathway_id = $1
		group by st.severity, st.los_p75 order by st.severity`, pathwayID, cbg, activatedAt)
	groups, err := pgx.CollectRows(rows, pgx.RowToStructByName[group])
	if err != nil {
		return in, err
	}
	for _, g := range groups {
		lg := domain.LOSGroup{Severity: g.Severity, PostLOS: g.LOS}
		if g.P75 != nil {
			lg.BaselineP75 = *g.P75
		}
		in.Groups = append(in.Groups, lg)
	}
	// Dihitung terpisah: episode sesudah aktif pada severity tanpa baseline tetap masuk bauran.
	var post, postSev3 int
	if err := s.pool.QueryRow(ctx, `
		select count(*), count(*) filter (where severity = 3) from claim_episodes
		where cbg_code = $1 and discharge_date >= ($2::timestamptz at time zone 'Asia/Jakarta')::date`, cbg, activatedAt).Scan(&post, &postSev3); err != nil {
		return in, err
	}
	in.PostEpisodes = post
	if post > 0 {
		in.PostSev3Share = float64(postSev3) / float64(post)
	}
	return in, nil
}

// EvaluationDetail: rincian kendali mutu & kendali biaya satu CP Aktif.
func (s *Store) EvaluationDetail(ctx context.Context, pathwayID int64, cbg string, activatedAt time.Time) (Row, error) {
	quality, err := queryRows(ctx, s.pool, `
		select st.severity, st.target_los::float8, st.los_p75::float8 as target_p75,
		       count(e.id)::int as episodes,
		       (percentile_cont(0.5)  within group (order by e.los_days))::float8 as los_median,
		       (percentile_cont(0.75) within group (order by e.los_days))::float8 as los_p75,
		       count(e.id) filter (where e.icu_days > 0)::int as icu_episodes
		from pathway_severity_stats st
		left join claim_episodes e
		  on e.cbg_code = $2 and e.severity = st.severity and e.discharge_date >= ($3::timestamptz at time zone 'Asia/Jakarta')::date
		where st.pathway_id = $1
		group by st.severity, st.target_los, st.los_p75 order by st.severity`, pathwayID, cbg, activatedAt)
	if err != nil {
		return nil, err
	}
	severityMix, err := queryRows(ctx, s.pool, `
		select severity,
		       count(*) filter (where discharge_date <  ($2::timestamptz at time zone 'Asia/Jakarta')::date)::int as sebelum,
		       count(*) filter (where discharge_date >= ($2::timestamptz at time zone 'Asia/Jakarta')::date)::int as sesudah
		from claim_episodes where cbg_code = $1 group by severity order by severity`, cbg, activatedAt)
	if err != nil {
		return nil, err
	}
	classMix, err := queryRows(ctx, s.pool, `
		select care_class, count(*)::int as episodes, avg(tariff)::float8 as avg_tariff, sum(tariff)::float8 as total_claim
		from claim_episodes where cbg_code = $1 and discharge_date >= ($2::timestamptz at time zone 'Asia/Jakarta')::date
		group by care_class order by care_class`, cbg, activatedAt)
	if err != nil {
		return nil, err
	}
	return Row{
		"kendali_mutu":  quality,
		"kendali_biaya": Row{"bauran_severity": severityMix, "bauran_kelas": classMix},
		"catatan":       "Biaya rumah sakit tidak dimuat; hanya sisi tarif INA-CBG yang ditampilkan.",
	}, nil
}
