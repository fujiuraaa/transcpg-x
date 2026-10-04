package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"transcpg-x/backend/pkg/domain"
)

// TransitionResult adalah hasil satu tindakan pengesahan.
type TransitionResult struct {
	From domain.Stage `json:"from"`
	To   domain.Stage `json:"to"`
}

// Transition memajukan/mengembalikan/mencabut CP. Menuju AKTIF, ke-7 syarat
// diperiksa di dalam transaksi yang sama dan seluruh kekurangan dikembalikan
// sekaligus sebagai *domain.NotReadyError.
func (s *Store) Transition(ctx context.Context, pathwayID int64, action domain.Action, reason string, u domain.User) (TransitionResult, error) {
	var res TransitionResult
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `select stage from clinical_pathways where id = $1 for update`, pathwayID).
			Scan(&res.From); err != nil {
			return err
		}
		to, err := domain.Transition(res.From, action, u.Role, reason)
		if err != nil {
			return err
		}
		if to == domain.StageAktif {
			r, err := readiness(ctx, tx, pathwayID)
			if err != nil {
				return err
			}
			if err := domain.CheckActivation(r); err != nil {
				return err
			}
		}
		res.To = to

		if _, err := tx.Exec(ctx, `
			update clinical_pathways set
			  stage = $2,
			  stage_changed_at = now(),
			  activated_at = case when $2 = 'AKTIF' then now() else activated_at end
			where id = $1`, pathwayID, to); err != nil {
			return err
		}
		var reasonArg *string
		if reason != "" {
			reasonArg = &reason
		}
		_, err = tx.Exec(ctx, `
			insert into approval_history (pathway_id, from_stage, to_stage, action, actor_id, actor_role, reason)
			values ($1, $2, $3, $4, $5, $6, $7)`,
			pathwayID, res.From, to, action, u.ID, u.Role, reasonArg)
		return err
	})
	return res, err
}

// ApprovalHistory: riwayat pengesahan satu CP (terbaru di atas).
func (s *Store) ApprovalHistory(ctx context.Context, pathwayID int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select h.id, h.from_stage, h.to_stage, h.action, h.reason, h.actor_role,
		       u.full_name as actor_name, h.created_at
		from approval_history h join app_users u on u.id = h.actor_id
		where h.pathway_id = $1 order by h.created_at desc, h.id desc`, pathwayID)
}

// ApprovalTab adalah tab halaman Approval.
type ApprovalTab string

const (
	TabMine       ApprovalTab = "saya"   // Menunggu Tindakan Saya
	TabProcessing ApprovalTab = "proses" // Sedang Diproses
	TabActive     ApprovalTab = "aktif"  // CP Aktif
)

// ApprovalQueue mengembalikan baris antrean untuk tab tertentu.
func (s *Store) ApprovalQueue(ctx context.Context, tab ApprovalTab, role domain.Role) ([]Row, error) {
	var stages []domain.Stage
	switch tab {
	case TabMine:
		stages = domain.StagesActionableBy(role)
	case TabActive:
		stages = []domain.Stage{domain.StageAktif}
	default:
		for _, st := range domain.AllStages() {
			if st.InApprovalQueue() {
				stages = append(stages, st)
			}
		}
	}
	if len(stages) == 0 {
		return []Row{}, nil
	}
	return queryRows(ctx, s.pool, `
		select id, cbg_code, name, mdc_name, episodes, stage, stage_changed_at, acuan_count
		from v_pathway_overview
		where stage = any($1)
		order by stage_changed_at`, stageStrings(stages))
}

// StageCounts: jumlah CP per tahap (kartu di atas halaman Approval).
func (s *Store) StageCounts(ctx context.Context) (map[domain.Stage]int, error) {
	type stageCount struct {
		Stage domain.Stage `db:"stage"`
		N     int          `db:"n"`
	}
	rows, _ := s.pool.Query(ctx, `select stage, count(*)::int as n from clinical_pathways group by stage`)
	list, err := pgx.CollectRows(rows, pgx.RowToStructByName[stageCount])
	if err != nil {
		return nil, err
	}
	out := map[domain.Stage]int{}
	for _, st := range domain.AllStages() {
		out[st] = 0
	}
	for _, x := range list {
		out[x.Stage] = x.N
	}
	return out, nil
}

func stageStrings(stages []domain.Stage) []string {
	out := make([]string, len(stages))
	for i, s := range stages {
		out[i] = string(s)
	}
	return out
}

// CountActionable: jumlah CP yang menunggu tindakan role (angka di menu Approval).
func (s *Store) CountActionable(ctx context.Context, role domain.Role) (int, error) {
	stages := domain.StagesActionableBy(role)
	if len(stages) == 0 {
		return 0, nil
	}
	var n int
	err := s.pool.QueryRow(ctx, `select count(*) from clinical_pathways where stage = any($1)`, stageStrings(stages)).Scan(&n)
	return n, err
}
