package store

import (
	"context"
	"encoding/json"
	"time"

	"transcpg-x/backend/pkg/domain"
)

// Activity adalah satu kejadian untuk activity_log (masuk, akun, RS, dokumen).
type Activity struct {
	Category    string // MASUK | AKUN | RS | DOKUMEN
	Action      string
	TargetType  string // USER | HOSPITAL | GUIDELINE | ""
	TargetID    int64
	TargetLabel string
	Summary     string
	Detail      map[string]any
	Actor       domain.User // nol = tanpa pelaku (mis. pendaftar baru)
	HospitalID  *int64      // lingkup Admin RS; default = RS pelaku
	IP          string
}

func (s *Store) LogActivity(ctx context.Context, a Activity) error {
	detail, err := json.Marshal(a.Detail)
	if err != nil || a.Detail == nil {
		detail = []byte("{}")
	}
	hospital := a.HospitalID
	if hospital == nil {
		hospital = a.Actor.HospitalID
	}
	_, err = s.pool.Exec(ctx, `
		insert into activity_log (category, action, target_type, target_id, target_label, summary, detail,
		  actor_id, actor_name, actor_role, hospital_id, ip)
		values ($1, $2, nullif($3, ''), nullif($4, 0), nullif($5, ''), $6, $7,
		  nullif($8, 0), nullif($9, ''), nullif($10, ''), $11, nullif($12, ''))`,
		a.Category, a.Action, a.TargetType, a.TargetID, a.TargetLabel, a.Summary, detail,
		a.Actor.ID, a.Actor.FullName, string(a.Actor.Role), hospital, a.IP)
	return err
}

// AuditFilter menyaring jejak audit. HospitalID = lingkup (Admin RS hanya RS-nya).
type AuditFilter struct {
	From, To   *time.Time
	Categories []string
	ActorID    *int64
	Q          string
	HospitalID *int64
	Limit      int
	Offset     int
}

func (f AuditFilter) args() []any {
	cats := f.Categories
	if cats == nil {
		cats = []string{}
	}
	return []any{f.From, f.To, cats, f.ActorID, f.Q, f.HospitalID}
}

const auditWhere = `
	where ($1::timestamptz is null or e.created_at >= $1)
	  and ($2::timestamptz is null or e.created_at < $2)
	  and (cardinality($3::text[]) = 0 or e.category = any($3))
	  and ($4::bigint is null or e.actor_id = $4)
	  and ($5 = '' or e.summary ilike '%' || $5 || '%' or e.target_label ilike '%' || $5 || '%'
	       or e.target_code ilike '%' || $5 || '%' or e.actor_name ilike '%' || $5 || '%')
	  and ($6::bigint is null or e.hospital_id = $6)`

// AuditEvents mengembalikan satu halaman jejak audit (terbaru dulu) beserta totalnya.
func (s *Store) AuditEvents(ctx context.Context, f AuditFilter) ([]Row, int, error) {
	args := append(f.args(), f.Limit, f.Offset)
	rows, err := queryRows(ctx, s.pool, `
		select e.uid, e.created_at, e.category, e.action, e.actor_id, e.actor_name, e.actor_role,
		       e.target_type, e.target_id, e.target_code, e.target_label, e.summary, e.detail, e.ip,
		       h.name as hospital_name, count(*) over ()::int as total
		from audit_events e left join hospitals h on h.id = e.hospital_id`+auditWhere+`
		order by e.created_at desc, e.uid desc
		limit $7 offset $8`, args...)
	if err != nil {
		return nil, 0, err
	}
	total := 0
	for _, r := range rows {
		total = int(r["total"].(int32))
		delete(r, "total")
	}
	if len(rows) == 0 && f.Offset > 0 {
		// Halaman di luar jangkauan: hitung ulang total agar paginasi tetap benar.
		if err := s.pool.QueryRow(ctx, `select count(*) from audit_events e`+auditWhere, f.args()...).Scan(&total); err != nil {
			return nil, 0, err
		}
	}
	return rows, total, nil
}

// AuditCategoryCounts: jumlah kejadian per kategori dengan filter yang sama
// (kecuali kategori) — untuk angka pada tombol saring.
func (s *Store) AuditCategoryCounts(ctx context.Context, f AuditFilter) (map[string]int, error) {
	f.Categories = nil
	rows, err := s.pool.Query(ctx, `select e.category, count(*)::int from audit_events e`+auditWhere+` group by 1`, f.args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var c string
		var n int
		if err := rows.Scan(&c, &n); err != nil {
			return nil, err
		}
		out[c] = n
	}
	return out, rows.Err()
}

// AuditActors: daftar pengguna untuk saringan "Pelaku".
func (s *Store) AuditActors(ctx context.Context, hospitalID *int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select id, full_name, role from app_users
		where registration_status = 'DISETUJUI' and ($1::bigint is null or hospital_id = $1)
		order by full_name`, hospitalID)
}

// ApprovalReport: rekap pengesahan CP dalam periode, lama tiap tahap, dan
// posisi semua CP saat ini. CP berlaku lintas RS, jadi tidak dilingkupi RS.
func (s *Store) ApprovalReport(ctx context.Context, from, to *time.Time) (Row, error) {
	totals, err := queryRow(ctx, s.pool, `
		select count(*) filter (where action = 'MAJUKAN' and from_stage in ('DRAF','REVISI'))::int as diajukan,
		       count(*) filter (where action = 'MAJUKAN' and to_stage = 'AKTIF')::int as disahkan,
		       count(*) filter (where action = 'KEMBALIKAN')::int as dikembalikan,
		       count(*) filter (where action = 'CABUT')::int as dicabut
		from approval_history
		where ($1::timestamptz is null or created_at >= $1) and ($2::timestamptz is null or created_at < $2)`, from, to)
	if err != nil {
		return nil, err
	}
	// Lama tinggal di tahap = jarak dari perpindahan sebelumnya (atau dibuatnya CP).
	durations, err := queryRows(ctx, s.pool, `
		with t as (
		  select h.from_stage, h.created_at,
		         least(h.created_at, coalesce(lag(h.created_at) over (partition by h.pathway_id order by h.created_at, h.id), p.created_at)) as entered_at
		  from approval_history h join clinical_pathways p on p.id = h.pathway_id
		)
		select from_stage as stage, count(*)::int as transitions,
		       round(avg(extract(epoch from created_at - entered_at) / 86400)::numeric, 1)::float8 as avg_days,
		       round(max(extract(epoch from created_at - entered_at) / 86400)::numeric, 1)::float8 as max_days
		from t
		where ($1::timestamptz is null or created_at >= $1) and ($2::timestamptz is null or created_at < $2)
		  and from_stage in ('DRAF','REVISI','REVIEW_TIM_CP','REVIEW_KOMITE','MENUNGGU_DIREKTUR')
		group by from_stage`, from, to)
	if err != nil {
		return nil, err
	}
	pathways, err := queryRows(ctx, s.pool, `
		select p.cbg_code as code, p.name, p.stage, p.stage_changed_at, p.activated_at,
		       (select count(*) from approval_history h where h.pathway_id = p.id and h.action = 'KEMBALIKAN')::int as returned_count,
		       (select u.full_name from approval_history h join app_users u on u.id = h.actor_id
		        where h.pathway_id = p.id and h.to_stage = 'AKTIF' order by h.created_at desc limit 1) as approved_by
		from clinical_pathways p
		order by case p.stage when 'AKTIF' then 5 when 'MENUNGGU_DIREKTUR' then 4 when 'REVIEW_KOMITE' then 3
		         when 'REVIEW_TIM_CP' then 2 when 'REVISI' then 1 else 0 end desc, p.stage_changed_at desc`)
	if err != nil {
		return nil, err
	}
	return Row{"totals": totals, "stage_durations": durations, "pathways": pathways}, nil
}

// UsersReport: rekap akun dan aktivitas pengguna dalam periode.
func (s *Store) UsersReport(ctx context.Context, from, to *time.Time, hospitalID *int64) ([]Row, error) {
	return queryRows(ctx, s.pool, `
		select u.id, u.full_name, u.email, u.role, h.name as hospital_name, u.is_active, u.registration_status,
		       u.last_login_at, u.created_at,
		       (select count(*) from activity_log a where a.actor_id = u.id and a.category = 'MASUK' and a.action = 'MASUK'
		          and ($1::timestamptz is null or a.created_at >= $1) and ($2::timestamptz is null or a.created_at < $2))::int as login_count,
		       (select count(*) from audit_events e where e.actor_id = u.id and e.category <> 'MASUK'
		          and ($1::timestamptz is null or e.created_at >= $1) and ($2::timestamptz is null or e.created_at < $2))::int as action_count
		from app_users u left join hospitals h on h.id = u.hospital_id
		where ($3::bigint is null or u.hospital_id = $3)
		order by u.registration_status = 'DISETUJUI' desc, u.is_active desc, u.full_name`, from, to, hospitalID)
}
