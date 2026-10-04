-- TransCPG-X · 010 · Pengerasan keamanan & jejak audit (hasil tinjauan)

-- 1. Sesi bisa dicabut: token membawa versi; naik saat kata sandi diganti/diatur ulang.
alter table app_users add column token_version integer not null default 0;

-- 2. Pembatasan percobaan masuk gagal (per email/nomor, juga untuk akun yang tidak ada).
create table login_failures (
  id         bigint generated always as identity primary key,
  key        text not null,          -- 'email:<lower>' | 'otp:<user_id>'
  created_at timestamptz not null default now()
);
create index login_failures_key_idx on login_failures (key, created_at desc);
alter table login_failures enable row level security;

-- 3. Snapshot pelaku di riwayat: nama/RS tetap seperti saat kejadian walau
--    pengguna kelak mengganti nama atau pindah RS.
alter table approval_history   add column actor_name text, add column actor_hospital_id bigint;
alter table pathway_change_log add column actor_name text, add column actor_hospital_id bigint;
alter table mapping_audit_log  add column actor_name text, add column actor_role text, add column actor_hospital_id bigint;

update approval_history h   set actor_name = u.full_name, actor_hospital_id = u.hospital_id from app_users u where u.id = h.actor_id;
update pathway_change_log c set actor_name = u.full_name, actor_hospital_id = u.hospital_id from app_users u where u.id = c.actor_id;
update mapping_audit_log m  set actor_name = u.full_name, actor_role = u.role, actor_hospital_id = u.hospital_id from app_users u where u.id = m.actor_id;

create function snapshot_actor() returns trigger language plpgsql as $$
begin
  select u.full_name, u.hospital_id into new.actor_name, new.actor_hospital_id
  from app_users u where u.id = new.actor_id;
  return new;
end $$;
create function snapshot_mapping_actor() returns trigger language plpgsql as $$
begin
  select u.full_name, u.role, u.hospital_id into new.actor_name, new.actor_role, new.actor_hospital_id
  from app_users u where u.id = new.actor_id;
  return new;
end $$;
create trigger approval_history_snapshot   before insert on approval_history   for each row execute function snapshot_actor();
create trigger pathway_change_log_snapshot before insert on pathway_change_log for each row execute function snapshot_actor();
create trigger mapping_audit_log_snapshot  before insert on mapping_audit_log  for each row execute function snapshot_mapping_actor();

-- Jejak audit memakai snapshot (bukan nama/RS saat ini).
create or replace view audit_events with (security_invoker = true) as
  -- Pengesahan CP
  select 'ap-' || h.id as uid, h.created_at, 'PENGESAHAN'::text as category, h.action,
         h.actor_id, coalesce(h.actor_name, u.full_name) as actor_name, h.actor_role, coalesce(h.actor_hospital_id, u.hospital_id) as hospital_id,
         'CP'::text as target_type, p.id as target_id, p.cbg_code as target_code, p.name as target_label,
         case h.action
           when 'MAJUKAN' then case h.to_stage when 'AKTIF' then 'Mengesahkan CP menjadi Aktif'
                               else 'Meneruskan ke ' || stage_label(h.to_stage) end
           when 'KEMBALIKAN' then 'Mengembalikan untuk revisi (dari ' || stage_label(h.from_stage) || ')'
           else 'Mencabut status Aktif' end as summary,
         jsonb_build_object('dari', stage_label(h.from_stage), 'ke', stage_label(h.to_stage), 'alasan', h.reason) as detail,
         null::text as ip
  from approval_history h
  join clinical_pathways p on p.id = h.pathway_id
  join app_users u on u.id = h.actor_id

  union all
  -- Perubahan isi CP (acuan, rencana klinis, kriteria)
  select 'cp-' || c.id, c.created_at, 'ISI_CP', c.action,
         c.actor_id, coalesce(c.actor_name, u.full_name), c.actor_role, coalesce(c.actor_hospital_id, u.hospital_id),
         'CP', p.id, p.cbg_code, p.name,
         case c.action when 'TAMBAH' then 'Menambah ' when 'UBAH' then 'Mengubah ' else 'Menghapus ' end ||
         case c.entity
           when 'ACUAN' then 'acuan: ' || coalesce((select g.title from guideline_documents g where g.id = (c.payload->>'document_id')::bigint), 'dokumen #' || (c.payload->>'document_id'))
           when 'RENCANA' then 'rencana klinis: ' || coalesce(c.payload->>'item_name', '') ||
                               ' (hari ' || coalesce(c.payload->>'day', '?') || ', severity ' || coalesce(c.payload->>'severity', '?') || ')'
           else 'kriteria ' || lower(coalesce(c.payload->>'kind', '')) || ': ' || coalesce(c.payload->>'description', '') end,
         c.payload || jsonb_build_object('tahap', stage_label(c.stage)),
         null
  from pathway_change_log c
  join clinical_pathways p on p.id = c.pathway_id
  join app_users u on u.id = c.actor_id

  union all
  -- Padanan KPTL / SNOMED-CT
  select 'mp-' || m.id, m.created_at, 'PADANAN',
         case when m.new_value is null then 'CABUT' when m.old_value is null then 'TETAPKAN' else 'UBAH' end,
         m.actor_id, coalesce(m.actor_name, u.full_name), coalesce(m.actor_role, u.role), coalesce(m.actor_hospital_id, u.hospital_id),
         m.target, null, m.source_code,
         case when m.target = 'KPTL' then (select i.name from icd9cm_codes i where i.code = m.source_code) end,
         case when m.new_value is null then 'Mencabut padanan ' || m.target || ' ' || m.old_value
              when m.old_value is null then 'Menetapkan padanan ' || m.target || ' ' || m.new_value
              else 'Mengganti padanan ' || m.target || ' ' || m.old_value || ' → ' || m.new_value end,
         jsonb_build_object('lama', m.old_value, 'baru', m.new_value),
         null
  from mapping_audit_log m
  join app_users u on u.id = m.actor_id

  union all
  -- Masuk, akun, profil RS, dokumen panduan
  select 'ac-' || a.id, a.created_at, a.category, a.action,
         a.actor_id, coalesce(a.actor_name, u.full_name), a.actor_role, a.hospital_id,
         a.target_type, a.target_id, null, a.target_label, a.summary, a.detail, a.ip
  from activity_log a
  left join app_users u on u.id = a.actor_id;

-- 4. Tanggal aktif dibaca dalam WIB, bukan zona waktu sesi database (Supabase = UTC).
create or replace function refresh_pathway_severity_stats() returns void
language sql as $$
  insert into pathway_severity_stats (pathway_id, severity, episodes, los_median, los_p75, target_los, age_median)
  select p.id, e.severity, count(*),
         percentile_cont(0.5)  within group (order by e.los_days),
         percentile_cont(0.75) within group (order by e.los_days),
         percentile_cont(0.5)  within group (order by e.los_days),
         percentile_cont(0.5)  within group (order by e.age_years)
  from clinical_pathways p
  join claim_episodes e on e.cbg_code = p.cbg_code
  where p.activated_at is null or e.discharge_date < (p.activated_at at time zone 'Asia/Jakarta')::date
  group by p.id, e.severity
  on conflict (pathway_id, severity) do update set
    episodes   = excluded.episodes,
    los_median = excluded.los_median,
    los_p75    = excluded.los_p75,
    target_los = excluded.target_los,
    age_median = excluded.age_median;
$$;

do $$
begin
  if exists (select 1 from pg_roles where rolname = 'anon') then
    execute 'revoke all on login_failures, audit_events from anon, authenticated';
    execute 'revoke all on function snapshot_actor(), snapshot_mapping_actor() from anon, authenticated';
  end if;
end $$;
