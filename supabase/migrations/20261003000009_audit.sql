-- TransCPG-X · 009 · Laporan & Audit
-- activity_log mencatat kejadian di luar isi CP: masuk aplikasi, pengelolaan
-- akun, profil RS, dan dokumen panduan. audit_events menggabungkannya dengan
-- riwayat yang sudah ada (pengesahan, perubahan isi CP, padanan) menjadi satu
-- jejak audit yang seragam.

create table activity_log (
  id           bigint generated always as identity primary key,
  category     text not null check (category in ('MASUK','AKUN','RS','DOKUMEN')),
  action       text not null,
  target_type  text check (target_type in ('USER','HOSPITAL','GUIDELINE')),
  target_id    bigint,
  target_label text,
  summary      text not null,
  detail       jsonb not null default '{}',
  -- Nama & peran disalin saat kejadian: tetap terbaca walau akun kelak dihapus.
  actor_id     bigint references app_users(id) on delete set null,
  actor_name   text,
  actor_role   text,
  hospital_id  bigint references hospitals(id) on delete set null,  -- lingkup Admin RS
  ip           text,
  created_at   timestamptz not null default now()
);
create index activity_log_created_idx on activity_log (created_at desc);
create index activity_log_hospital_idx on activity_log (hospital_id, created_at desc);
alter table activity_log enable row level security;

create function stage_label(s text) returns text language sql immutable as $$
  select case s
    when 'DRAF' then 'Draf' when 'REVISI' then 'Perlu Revisi'
    when 'REVIEW_TIM_CP' then 'Review Tim CP' when 'REVIEW_KOMITE' then 'Review KSM/Komite Medik'
    when 'MENUNGGU_DIREKTUR' then 'Menunggu Direktur' when 'AKTIF' then 'Aktif' else s end
$$;

create view audit_events with (security_invoker = true) as
  -- Pengesahan CP
  select 'ap-' || h.id as uid, h.created_at, 'PENGESAHAN'::text as category, h.action,
         h.actor_id, u.full_name as actor_name, h.actor_role, u.hospital_id,
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
         c.actor_id, u.full_name, c.actor_role, u.hospital_id,
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
         m.actor_id, u.full_name, u.role, u.hospital_id,
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
         a.actor_id, coalesce(u.full_name, a.actor_name), a.actor_role, a.hospital_id,
         a.target_type, a.target_id, null, a.target_label, a.summary, a.detail, a.ip
  from activity_log a
  left join app_users u on u.id = a.actor_id;

-- Hanya backend (koneksi langsung) yang membaca jejak audit; tutup akses Data API.
do $$
begin
  if exists (select 1 from pg_roles where rolname = 'anon') then
    execute 'revoke all on activity_log, audit_events from anon, authenticated';
    execute 'revoke all on function stage_label(text) from anon, authenticated';
  end if;
end $$;
