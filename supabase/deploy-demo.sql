-- TransCPG-X · DEPLOY DEMO — dibuat otomatis oleh gabung-demo.sh, jangan disunting langsung.
-- Isi: migrasi 001–010, bucket storage, dan data demo. Jalankan SEKALI pada database kosong.

-- ==================== migrations/20261003000001_master_data.sql ====================
-- TransCPG-X · 001 · Profil RS, pengguna, dan data master (koding & terminologi)
-- Kompatibel dengan Supabase CLI (supabase/migrations) dan PostgreSQL 15+.

create extension if not exists pg_trgm;
create extension if not exists pgcrypto;

-- updated_at otomatis ---------------------------------------------------------
create or replace function set_updated_at() returns trigger
language plpgsql as $$
begin
  new.updated_at := now();
  return new;
end $$;

-- Profil rumah sakit (menentukan tarif INA-CBG yang ditampilkan) ---------------
create table hospitals (
  id                 bigint generated always as identity primary key,
  name               text not null,
  hospital_type      text not null check (hospital_type in ('A','B','C','D','KLINIK_UTAMA','KLINIK_PRATAMA','PUSKESMAS')),
  ownership          text not null check (ownership in ('PEMERINTAH','SWASTA')),
  bpjs_regional      smallint not null check (bpjs_regional between 1 and 5),
  max_care_class     smallint not null default 1 check (max_care_class between 1 and 3),
  bpjs_provider_code text,
  accreditation      text,
  city               text,
  province           text,
  address            text,
  phone              text,
  email              text,
  bed_capacity       integer check (bed_capacity >= 0),
  icu_capacity       integer check (icu_capacity >= 0),
  is_default         boolean not null default false,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now()
);
create unique index hospitals_single_default on hospitals (is_default) where is_default;
create trigger hospitals_updated_at before update on hospitals for each row execute function set_updated_at();

-- Pengguna aplikasi (login custom JWT; tidak memakai Supabase Auth) -------------
create table app_users (
  id            bigint generated always as identity primary key,
  hospital_id   bigint references hospitals(id) on delete set null,
  full_name     text not null,
  email         text not null,
  password_hash text not null,
  role          text not null check (role in (
                  'SUPER_ADMIN','SYSTEM_ADMIN','ADMIN_RS',
                  'DOKTER','DPJP','DOKTER_UMUM','RESIDEN',
                  'KOORDINATOR_CP','TIM_CP',
                  'PERAWAT_PELAKSANA','KEPALA_PERAWAT','APOTEKER','DIETISIEN',
                  'KOMITE_MEDIK','DIREKTUR')),
  is_active     boolean not null default true,
  last_login_at timestamptz,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
create unique index app_users_email_key on app_users (lower(email));
create trigger app_users_updated_at before update on app_users for each row execute function set_updated_at();

-- Master koding & terminologi --------------------------------------------------
create table mdc_groups (
  code text primary key,           -- Major Diagnostic Category INA-CBG
  name text not null
);

create table icd10_codes (
  code     text primary key,
  name     text not null,
  is_valid boolean not null default true  -- sesuai master Kemenkes 2010
);
create index icd10_codes_name_trgm on icd10_codes using gin (name gin_trgm_ops);

create table icd9cm_codes (
  code text primary key,
  name text not null
);
create index icd9cm_codes_name_trgm on icd9cm_codes using gin (name gin_trgm_ops);

create table loinc_codes (
  code      text primary key,
  name      text not null,
  component text,
  system    text
);
create index loinc_codes_name_trgm on loinc_codes using gin (name gin_trgm_ops);

create table kfa_drugs (
  code         text primary key,   -- Kamus Farmasi & Alat Kesehatan Kemenkes
  name         text not null,
  dosage_form  text,
  strength     text
);
create index kfa_drugs_name_trgm on kfa_drugs using gin (name gin_trgm_ops);

create table fornas_drugs (
  id           bigint generated always as identity primary key,
  kfa_code     text references kfa_drugs(code),
  name         text not null,
  restrictions text,
  edition      text not null default '2025'
);

create table kptl_codes (
  code text primary key,           -- kode billing BPJS
  name text not null
);
create index kptl_codes_name_trgm on kptl_codes using gin (name gin_trgm_ops);

create table snomed_concepts (
  concept_id     text primary key,
  fsn            text not null,     -- fully specified name
  preferred_term text not null,
  semantic_tag   text not null,     -- disorder, finding, procedure, ...
  active         boolean not null default true  -- konsep retired tampil tapi tidak bisa dipilih
);
create index snomed_concepts_term_trgm on snomed_concepts using gin (preferred_term gin_trgm_ops);

-- Grouper & tarif INA-CBG --------------------------------------------------------
create table ina_cbg_groupers (
  code      text primary key,      -- mis. A-4-13 (tanpa akhiran severity)
  name      text not null,
  mdc_code  text not null references mdc_groups(code),
  case_type text not null default 'RAWAT_INAP'
);

create table ina_cbg_tariffs (
  cbg_code      text not null references ina_cbg_groupers(code),
  severity      smallint not null check (severity between 1 and 3),
  bpjs_regional smallint not null check (bpjs_regional between 1 and 5),
  hospital_type text not null,
  ownership     text not null check (ownership in ('PEMERINTAH','SWASTA')),
  care_class    smallint not null check (care_class between 1 and 3),
  tariff        numeric(14,2) not null check (tariff >= 0),
  regulation    text not null default 'Permenkes 3/2023',
  primary key (cbg_code, severity, bpjs_regional, hospital_type, ownership, care_class)
);

-- Sistem skoring klinis -----------------------------------------------------------
create table clinical_scoring_tools (
  code               text primary key,  -- SOFA, NEWS2, GCS, ...
  name               text not null,
  setting            text,
  critical_threshold text
);

-- ==================== migrations/20261003000002_claims_guidelines.sql ====================
-- TransCPG-X · 002 · Data klaim agregat dan dokumen panduan klinis
-- Tidak ada identitas pasien (nama, no. RM) yang disimpan — hanya episode klaim.

create table claim_episodes (
  id              bigint generated always as identity primary key,
  hospital_id     bigint references hospitals(id),
  cbg_code        text not null references ina_cbg_groupers(code),
  severity        smallint not null check (severity between 1 and 3),
  care_class      smallint not null check (care_class between 1 and 3),
  admission_date  date not null,
  discharge_date  date not null,
  los_days        integer not null check (los_days >= 0),
  icu_days        integer not null default 0 check (icu_days >= 0),
  age_years       integer check (age_years >= 0),
  tariff          numeric(14,2),
  source_batch    text,
  check (discharge_date >= admission_date)
);
create index claim_episodes_cbg_idx on claim_episodes (cbg_code, discharge_date);

create table claim_episode_diagnoses (
  episode_id  bigint not null references claim_episodes(id) on delete cascade,
  icd10_code  text not null,
  is_primary  boolean not null default false,
  primary key (episode_id, icd10_code)
);
create index claim_episode_diagnoses_code_idx on claim_episode_diagnoses (icd10_code);

create table claim_episode_procedures (
  episode_id   bigint not null references claim_episodes(id) on delete cascade,
  icd9cm_code  text not null,
  primary key (episode_id, icd9cm_code)
);
create index claim_episode_procedures_code_idx on claim_episode_procedures (icd9cm_code);

-- Dokumen panduan (PPK RSCM, PNPK, PPK Asosiasi, dokumen Tim CP) -----------------
create table guideline_documents (
  id              bigint generated always as identity primary key,
  title           text not null,
  source_type     text not null check (source_type in ('PPK_RSCM','PNPK','PPK_ASOSIASI','TIM_CP','INTERNASIONAL','TKMKB','LAIN')),
  issuer          text,             -- mis. Kemenkes, PERKI, PDPI
  regulation_ref  text,             -- mis. HK.01.07/MENKES/1/2018
  year            smallint,
  icd10_codes     text[] not null default '{}',
  attachment_path text,             -- Supabase Storage, maks. 50 MB
  created_by      bigint references app_users(id),
  created_at      timestamptz not null default now()
);
create index guideline_documents_icd10_idx on guideline_documents using gin (icd10_codes);
create index guideline_documents_title_trgm on guideline_documents using gin (title gin_trgm_ops);

-- Butir acuan yang diekstrak dari dokumen ----------------------------------------
create table guideline_items (
  id           bigint generated always as identity primary key,
  document_id  bigint not null references guideline_documents(id) on delete cascade,
  category     text not null check (category in ('KRITERIA_DIAGNOSIS','TATA_LAKSANA','DOSIS_OBAT','INDIKASI_TERAPI','KONTRAINDIKASI','MONITORING','KOMPLIKASI','LAIN')),
  title        text not null,
  page         text not null,
  quote        text not null,
  icd10_codes  text[] not null default '{}'
);
create index guideline_items_document_idx on guideline_items (document_id);

-- Topik TKMKB (rekomendasi audit medis BPJS) -------------------------------------
create table tkmkb_topics (
  id           bigint generated always as identity primary key,
  title        text not null,
  year         smallint not null,
  document_id  bigint references guideline_documents(id)
);

-- ==================== migrations/20261003000003_pathways.sql ====================
-- TransCPG-X · 003 · Clinical Pathway, isi CP, dan alur pengesahan

create table clinical_pathways (
  id                   bigint generated always as identity primary key,
  cbg_code             text not null unique references ina_cbg_groupers(code),
  name                 text not null,
  stage                text not null default 'DRAF'
                       check (stage in ('DRAF','REVISI','REVIEW_TIM_CP','REVIEW_KOMITE','MENUNGGU_DIREKTUR','AKTIF')),
  split_decision       text check (split_decision in ('SATU','PECAH')),  -- hasil analisis klaim
  has_pediatric_cohort boolean not null default false,
  stage_changed_at     timestamptz not null default now(),
  activated_at         timestamptz,   -- pembanding evaluasi: hanya klaim sesudah tanggal ini
  created_at           timestamptz not null default now(),
  updated_at           timestamptz not null default now()
);
create trigger clinical_pathways_updated_at before update on clinical_pathways for each row execute function set_updated_at();

-- Statistik baseline per severity (dari klaim sebelum CP aktif; diisi job analisis)
create table pathway_severity_stats (
  pathway_id  bigint not null references clinical_pathways(id) on delete cascade,
  severity    smallint not null check (severity between 1 and 3),
  episodes    integer not null default 0,
  los_median  numeric(6,2),
  los_p75     numeric(6,2),
  target_los  numeric(6,2),
  age_median  numeric(5,1),
  primary key (pathway_id, severity)
);

-- Sub-CP: diagnosis penyusun bila grouper perlu dipecah
create table pathway_sub_cps (
  id             bigint generated always as identity primary key,
  pathway_id     bigint not null references clinical_pathways(id) on delete cascade,
  icd10_code     text not null,
  label          text not null,
  episode_share  numeric(5,4),
  unique (pathway_id, icd10_code)
);

-- Acuan standar (tingkat CP bila sub_cp_id null, tingkat sub-CP bila terisi)
create table pathway_references (
  id           bigint generated always as identity primary key,
  pathway_id   bigint not null references clinical_pathways(id) on delete cascade,
  sub_cp_id    bigint references pathway_sub_cps(id) on delete cascade,
  document_id  bigint not null references guideline_documents(id),
  note         text,
  assigned_by  bigint not null references app_users(id),
  assigned_at  timestamptz not null default now(),
  unique nulls not distinct (pathway_id, sub_cp_id, document_id)
);

-- Rencana isi klinis: obat/lab/prosedur/tindakan per severity per hari
create table clinical_plan_items (
  id                 bigint generated always as identity primary key,
  pathway_id         bigint not null references clinical_pathways(id) on delete cascade,
  severity           smallint not null check (severity between 1 and 3),
  day                smallint not null check (day between 0 and 14),
  item_type          text not null check (item_type in ('OBAT','LAB','PROSEDUR','TINDAKAN')),
  item_code          text not null,   -- KFA (obat) · LOINC (lab) · ICD-9-CM (prosedur/tindakan)
  item_name          text not null,
  dose               text,
  route              text,
  frequency          text,
  duration           text,
  nature             text not null check (nature in ('WAJIB','KONDISIONAL')),
  guideline_item_id  bigint references guideline_items(id),  -- null = "DI LUAR PANDUAN"
  source_note        text,            -- mis. "PPK PD RSCM 2021, hlm. 234"
  note               text,
  created_by         bigint not null references app_users(id),
  created_at         timestamptz not null default now(),
  updated_by         bigint references app_users(id),
  updated_at         timestamptz not null default now(),
  -- Tidak boleh ganda: item yang sama pada severity & hari yang sama
  unique (pathway_id, severity, day, item_type, item_code),
  -- Obat wajib lengkap: dosis, rute, frekuensi
  constraint plan_obat_lengkap check (
    item_type <> 'OBAT'
    or (coalesce(btrim(dose), '') <> '' and coalesce(btrim(route), '') <> '' and coalesce(btrim(frequency), '') <> '')
  )
);
create index clinical_plan_items_pathway_idx on clinical_plan_items (pathway_id, severity, day);
create trigger clinical_plan_items_updated_at before update on clinical_plan_items for each row execute function set_updated_at();

-- Kode harus sah: item_code wajib ada di master sesuai jenisnya
create or replace function validate_plan_item_code() returns trigger
language plpgsql as $$
declare
  found boolean;
begin
  if new.item_type = 'OBAT' then
    select exists (select 1 from kfa_drugs where code = new.item_code) into found;
  elsif new.item_type = 'LAB' then
    select exists (select 1 from loinc_codes where code = new.item_code) into found;
  else
    select exists (select 1 from icd9cm_codes where code = new.item_code) into found;
  end if;
  if not found then
    raise exception 'Kode % tidak ditemukan di master untuk jenis %', new.item_code, new.item_type
      using errcode = 'P0001', hint = 'KODE_TIDAK_SAH';
  end if;
  return new;
end $$;
create trigger clinical_plan_items_validate_code
  before insert or update of item_type, item_code on clinical_plan_items
  for each row execute function validate_plan_item_code();

-- Kriteria inklusi/eksklusi pasien
create table pathway_criteria (
  id           bigint generated always as identity primary key,
  pathway_id   bigint not null references clinical_pathways(id) on delete cascade,
  kind         text not null check (kind in ('INKLUSI','EKSKLUSI')),
  description  text not null check (btrim(description) <> ''),
  created_by   bigint not null references app_users(id),
  created_at   timestamptz not null default now()
);

-- Riwayat pengesahan: setiap perpindahan tahap
create table approval_history (
  id          bigint generated always as identity primary key,
  pathway_id  bigint not null references clinical_pathways(id) on delete cascade,
  from_stage  text not null,
  to_stage    text not null,
  action      text not null check (action in ('MAJUKAN','KEMBALIKAN','CABUT')),
  actor_id    bigint not null references app_users(id),
  actor_role  text not null,
  reason      text,
  created_at  timestamptz not null default now()
);
create index approval_history_pathway_idx on approval_history (pathway_id, created_at desc);

-- Riwayat perubahan isi CP (acuan/rencana/kriteria)
create table pathway_change_log (
  id          bigint generated always as identity primary key,
  pathway_id  bigint not null references clinical_pathways(id) on delete cascade,
  entity      text not null check (entity in ('ACUAN','RENCANA','KRITERIA')),
  action      text not null check (action in ('TAMBAH','UBAH','HAPUS')),
  stage       text not null,
  payload     jsonb not null,
  actor_id    bigint not null references app_users(id),
  actor_role  text not null,
  created_at  timestamptz not null default now()
);
create index pathway_change_log_pathway_idx on pathway_change_log (pathway_id, created_at desc);

-- Skoring, usulan aturan CDSS, dan indikator mutu (diturunkan dari data klaim)
create table pathway_scoring_tools (
  pathway_id  bigint not null references clinical_pathways(id) on delete cascade,
  tool_code   text not null references clinical_scoring_tools(code),
  primary key (pathway_id, tool_code)
);

create table pathway_rule_suggestions (
  id          bigint generated always as identity primary key,
  pathway_id  bigint not null references clinical_pathways(id) on delete cascade,
  rule_type   text not null check (rule_type in ('ESCALATION','STEP_ACTION','DOCUMENTATION','PATHWAY_BRANCH','CLINICAL_ALERT','DRUG_SAFETY','LAB_ALERT','ALERT')),
  condition   jsonb not null,
  action      jsonb not null,
  rationale   text
);

create table pathway_quality_indicators (
  id          bigint generated always as identity primary key,
  pathway_id  bigint not null references clinical_pathways(id) on delete cascade,
  code        text not null,
  name        text not null,
  target      numeric(8,2),
  baseline    numeric(8,2),
  unit        text,
  unique (pathway_id, code)
);

-- ==================== migrations/20261003000004_mappings.sql ====================
-- TransCPG-X · 004 · Padanan KPTL dan SNOMED-CT

-- ICD-9-CM → KPTL (billing BPJS). Ditetapkan manusia; usulan mesin hanya petunjuk.
create table kptl_mappings (
  icd9cm_code  text primary key references icd9cm_codes(code),
  kptl_code    text not null references kptl_codes(code),
  note         text,
  mapped_by    bigint not null references app_users(id),
  mapped_at    timestamptz not null default now()
);

-- ICD-10 / ICD-9-CM → SNOMED-CT. Sebagian besar otomatis; RAGU = bertentangan
-- dengan peta resmi dan perlu ditinjau manusia.
create table snomed_mappings (
  id           bigint generated always as identity primary key,
  vocabulary   text not null check (vocabulary in ('ICD10','ICD9CM')),
  source_code  text not null,
  concept_id   text not null references snomed_concepts(concept_id),
  status       text not null check (status in ('OTOMATIS','MANUAL','RAGU')),
  mapped_by    bigint references app_users(id),
  mapped_at    timestamptz not null default now(),
  unique (vocabulary, source_code)
);

-- Riwayat padanan (untuk audit; ditulis store setiap tetapkan/cabut)
create table mapping_audit_log (
  id          bigint generated always as identity primary key,
  target      text not null check (target in ('KPTL','SNOMED')),
  source_code text not null,
  old_value   text,
  new_value   text,
  actor_id    bigint not null references app_users(id),
  created_at  timestamptz not null default now()
);

-- ==================== migrations/20261003000005_views_functions.sql ====================
-- TransCPG-X · 005 · View baca, fungsi analisis klaim, dan pengamanan akses

-- Volume episode per grouper (dasar "layak disusun" vs "dilewati") -------------
create view v_grouper_volume with (security_invoker = true) as
select g.code as cbg_code, g.name, g.mdc_code, count(e.id)::int as episodes
from ina_cbg_groupers g
left join claim_episodes e on e.cbg_code = g.code
group by g.code;

-- Ringkasan satu CP untuk CP Library, Dashboard, dan Approval -----------------
create view v_pathway_overview with (security_invoker = true) as
select
  p.id, p.cbg_code, p.name, g.mdc_code, m.name as mdc_name,
  p.stage, p.stage_changed_at, p.activated_at, p.split_decision, p.has_pediatric_cohort,
  coalesce(s.episodes, 0)               as episodes,
  s.dominant_severity,
  s.target_los::float8                  as target_los,
  coalesce(s.sev_episodes, '{}'::jsonb) as sev_episodes,
  coalesce(s.severities_adequate, false) as severities_adequate,
  coalesce(r.acuan_count, 0)            as acuan_count,
  r.acuan_titles,
  coalesce(c.candidate_count, 0)        as candidate_count,
  coalesce(k.procedures, 0)             as procedure_count,
  coalesce(k.mapped, 0)                 as procedure_kptl_mapped
from clinical_pathways p
join ina_cbg_groupers g on g.code = p.cbg_code
join mdc_groups m on m.code = g.mdc_code
left join lateral (
  select sum(ss.episodes)::int                                   as episodes,
         (array_agg(ss.severity   order by ss.episodes desc))[1] as dominant_severity,
         (array_agg(ss.target_los order by ss.episodes desc))[1] as target_los,
         jsonb_object_agg(ss.severity, ss.episodes)              as sev_episodes,
         count(*) = 3 and bool_and(ss.episodes >= 10)            as severities_adequate
  from pathway_severity_stats ss
  where ss.pathway_id = p.id
) s on true
left join lateral (
  select count(*)::int as acuan_count,
         string_agg(d.title, '; ' order by pr.assigned_at) as acuan_titles
  from pathway_references pr
  join guideline_documents d on d.id = pr.document_id
  where pr.pathway_id = p.id and pr.sub_cp_id is null
) r on true
left join lateral (
  select count(*)::int as candidate_count
  from guideline_documents d
  where d.icd10_codes && array(
    select distinct cd.icd10_code
    from claim_episodes e
    join claim_episode_diagnoses cd on cd.episode_id = e.id and cd.is_primary
    where e.cbg_code = p.cbg_code)
) c on true
left join lateral (
  select count(distinct cp.icd9cm_code)::int as procedures,
         count(distinct km.icd9cm_code)::int as mapped
  from claim_episodes e
  join claim_episode_procedures cp on cp.episode_id = e.id
  left join kptl_mappings km on km.icd9cm_code = cp.icd9cm_code
  where e.cbg_code = p.cbg_code
) k on true;

-- Langkah 1 alur utama: grouper dengan episode ≥5 menjadi CP Draf ---------------
create or replace function generate_draft_pathways(min_episodes int default 5) returns integer
language sql as $$
  with inserted as (
    insert into clinical_pathways (cbg_code, name)
    select v.cbg_code, v.name
    from v_grouper_volume v
    where v.episodes >= min_episodes
    on conflict (cbg_code) do nothing
    returning 1
  )
  select count(*)::int from inserted;
$$;

-- Statistik baseline per severity dari klaim SEBELUM CP aktif -------------------
-- Catatan: target_los sementara = median LOS. Aturan final ditetapkan Tim CP.
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
  where p.activated_at is null or e.discharge_date < p.activated_at::date
  group by p.id, e.severity
  on conflict (pathway_id, severity) do update set
    episodes   = excluded.episodes,
    los_median = excluded.los_median,
    los_p75    = excluded.los_p75,
    target_los = excluded.target_los,
    age_median = excluded.age_median;
$$;

-- Pengamanan: seluruh akses data lewat backend Go (koneksi pemilik tabel). ---------
-- RLS aktif tanpa policy → API REST/anon Supabase tidak bisa membaca apa pun.
do $$
declare t record;
begin
  for t in select tablename from pg_tables where schemaname = 'public' loop
    execute format('alter table public.%I enable row level security', t.tablename);
  end loop;
  if exists (select 1 from pg_roles where rolname = 'anon') then
    execute 'revoke all on all tables in schema public from anon, authenticated';
    execute 'revoke all on all functions in schema public from anon, authenticated';
  end if;
end $$;

-- ==================== migrations/20261003000006_login_methods.sql ====================
-- TransCPG-X · 006 · Metode masuk tambahan: Google (berdasarkan email) dan
-- nomor telepon + kode OTP. Hanya untuk akun yang sudah dibuat Admin RS.

alter table app_users add column phone text;
alter table app_users add constraint app_users_phone_e164
  check (phone is null or phone ~ '^\+[1-9][0-9]{7,14}$');
create unique index app_users_phone_key on app_users (phone) where phone is not null;

-- Kode OTP disimpan sebagai hash (HMAC), tidak pernah dalam bentuk asli.
create table login_otps (
  id           bigint generated always as identity primary key,
  user_id      bigint not null references app_users(id) on delete cascade,
  code_hash    text not null,
  expires_at   timestamptz not null,
  attempts     smallint not null default 0,
  consumed_at  timestamptz,
  created_at   timestamptz not null default now()
);
create index login_otps_user_idx on login_otps (user_id, created_at desc);

alter table login_otps enable row level security;
do $$
begin
  if exists (select 1 from pg_roles where rolname = 'anon') then
    execute 'revoke all on login_otps from anon, authenticated';
  end if;
end $$;

-- ==================== migrations/20261003000007_self_registration.sql ====================
-- TransCPG-X · 007 · Pendaftaran mandiri dengan persetujuan Admin RS.
-- Pendaftar memilih peran yang DIAJUKAN; akun belum aktif sampai Admin
-- menyetujui (dan boleh mengubah perannya) atau menolak dengan alasan.

alter table app_users
  add column registration_status text not null default 'DISETUJUI'
    check (registration_status in ('MENUNGGU', 'DISETUJUI', 'DITOLAK')),
  add column self_registered   boolean not null default false,
  add column registration_note text,          -- mis. unit kerja / SMF, diisi pendaftar
  add column reviewed_by       bigint references app_users(id),
  add column reviewed_at       timestamptz,
  add column rejection_reason  text;

-- Akun yang belum disetujui tidak boleh aktif.
alter table app_users add constraint app_users_pending_inactive
  check (registration_status = 'DISETUJUI' or not is_active);

create index app_users_pending_idx on app_users (created_at) where registration_status = 'MENUNGGU';

-- ==================== migrations/20261003000008_guideline_attachments.sql ====================
-- TransCPG-X · 008 · Metadata lampiran PDF dokumen panduan.
-- File disimpan di object storage (Supabase Storage, bucket privat);
-- attachment_path (migrasi 002) menyimpan path objeknya.

alter table guideline_documents
  add column attachment_name        text,         -- nama file asli (untuk ditampilkan)
  add column attachment_size        bigint check (attachment_size between 1 and 52428800),
  add column attachment_uploaded_at timestamptz,
  add column attachment_uploaded_by bigint references app_users(id);

-- Produksi (Supabase): buat bucket privat sekali saja, mis. lewat dashboard
-- atau SQL berikut (jalankan bila skema storage tersedia):
--   insert into storage.buckets (id, name, public, file_size_limit, allowed_mime_types)
--   values ('guideline-attachments', 'guideline-attachments', false, 52428800, '{application/pdf}')
--   on conflict (id) do nothing;

-- ==================== migrations/20261003000009_audit.sql ====================
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

-- ==================== migrations/20261003000010_hardening.sql ====================
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

-- ==================== storage.sql ====================
-- TransCPG-X · Bucket penyimpanan lampiran PDF (khusus Supabase)
-- Privat: file hanya bisa diunduh lewat URL bertanda tangan dari backend.
insert into storage.buckets (id, name, public, file_size_limit, allowed_mime_types)
values ('guideline-attachments', 'guideline-attachments', false, 52428800, array['application/pdf'])
on conflict (id) do nothing;

-- ==================== demo.sql ====================
-- TransCPG-X · DATA DEMO (untuk presentasi)
-- Jalankan SETELAH semua migrasi, pada database kosong. Seluruh data fiktif:
-- nama orang, kode KFA/KPTL, tarif, episode klaim, dan kutipan panduan adalah
-- contoh — bukan teks/angka resmi. Kode ICD-10, ICD-9-CM, LOINC, dan SNOMED-CT
-- memakai kode umum agar terasa nyata.
--
-- Akun demo — SEMUA memakai kata sandi: Demo-TransCPG-2026
-- (segera ganti kata sandi akun admin lewat Profil Saya setelah demo online)
--   superadmin@transcpg.test   Super Admin
--   sysadmin@transcpg.test     System Admin
--   admin@transcpg.test        Admin RS (RSCM)
--   admin.rsc@transcpg.test    Admin RS (RS Contoh Tipe C)
--   dpjp@transcpg.test         DPJP — penyusun CP
--   dpjp2@transcpg.test        DPJP — penyusun CP
--   koordinator@transcpg.test  Koordinator CP
--   timcp@transcpg.test        Tim CP
--   komite@transcpg.test       KSM/Komite Medik
--   direktur@transcpg.test     Direktur
--   apoteker@transcpg.test     Apoteker
--   perawat@transcpg.test      Perawat Pelaksana
--   kepala.perawat@transcpg.test  Kepala Perawat
--   residen@transcpg.test      Residen (PPDS)

-- Profil RS ------------------------------------------------------------------------
insert into hospitals (name, hospital_type, ownership, bpjs_regional, max_care_class, bpjs_provider_code,
                       accreditation, city, province, address, phone, email, bed_capacity, icu_capacity, is_default)
values ('RSUPN Dr. CIPTO MANGUNKUSUMO (RSCM)', 'A', 'PEMERINTAH', 1, 1, '0901R001', 'Paripurna', 'Jakarta Pusat',
        'DKI Jakarta', 'Jl. Diponegoro No. 71, Salemba', '(021) 1500135', 'info@rscm.demo', 1000, 120, true),
       ('RS Contoh Tipe C Swasta', 'C', 'SWASTA', 3, 1, null, 'Madya', 'Balikpapan',
        'Kalimantan Timur', 'Jl. Contoh No. 1', null, null, 150, 10, false);

-- Pengguna -------------------------------------------------------------------------
insert into app_users (hospital_id, full_name, email, password_hash, role, last_login_at, created_at)
select u.hosp, u.full_name, u.email, crypt('Demo-TransCPG-2026', gen_salt('bf', 10)), u.role,
       now() - make_interval(hours => u.last_h), timestamptz '2026-01-05 09:00+07'
from (values
  (1, 'Super Admin',                        'superadmin@transcpg.test',     'SUPER_ADMIN',       30),
  (1, 'System Admin',                       'sysadmin@transcpg.test',       'SYSTEM_ADMIN',      200),
  (1, 'Admin RS RSCM',                      'admin@transcpg.test',          'ADMIN_RS',          5),
  (2, 'Admin RS Contoh',                    'admin.rsc@transcpg.test',      'ADMIN_RS',          400),
  (1, 'dr. Andi Pratama, Sp.PD',            'dpjp@transcpg.test',           'DPJP',              3),
  (1, 'dr. Sari Wulandari, Sp.P',           'dpjp2@transcpg.test',          'DPJP',              20),
  (1, 'dr. Hendra Wijaya, MARS',            'koordinator@transcpg.test',    'KOORDINATOR_CP',    8),
  (1, 'Ns. Rina Kusuma, M.Kep',             'timcp@transcpg.test',          'TIM_CP',            2),
  (1, 'dr. Ratna Dewi, Sp.JP(K)',           'komite@transcpg.test',         'KOMITE_MEDIK',      26),
  (1, 'dr. Agus Salim, MPH',                'direktur@transcpg.test',       'DIREKTUR',          50),
  (1, 'apt. Maya Lestari, S.Farm',          'apoteker@transcpg.test',       'APOTEKER',          70),
  (1, 'Ns. Dimas Saputra, S.Kep',           'perawat@transcpg.test',        'PERAWAT_PELAKSANA', 30),
  (1, 'Ns. Lilis Handayani, S.Kep',         'kepala.perawat@transcpg.test', 'KEPALA_PERAWAT',    2600),
  (1, 'dr. Fajar Nugroho (PPDS Penyakit Dalam)', 'residen@transcpg.test', 'RESIDEN',           96)
) as u(hosp, full_name, email, role, last_h);

update app_users u set phone = p.phone
from (values
  ('dpjp@transcpg.test',     '+6281200000001'),
  ('timcp@transcpg.test',    '+6281200000002'),
  ('komite@transcpg.test',   '+6281200000003'),
  ('direktur@transcpg.test', '+6281200000004')
) as p(email, phone)
where u.email = p.email;

-- Pendaftaran mandiri: dua menunggu persetujuan, satu ditolak
insert into app_users (hospital_id, full_name, email, password_hash, role, is_active, registration_status,
                       self_registered, registration_note, rejection_reason, created_at)
values
  (1, 'dr. Yoga Permana, Sp.A', 'yoga.permana@transcpg.test', crypt('Demo-TransCPG-2026', gen_salt('bf', 10)),
   'DPJP', false, 'MENUNGGU', true, 'SMF Ilmu Kesehatan Anak — ingin ikut menyusun CP pediatri', null, now() - interval '2 days'),
  (1, 'Ns. Fitri Amalia, S.Kep', 'fitri.amalia@transcpg.test', crypt('Demo-TransCPG-2026', gen_salt('bf', 10)),
   'PERAWAT_PELAKSANA', false, 'MENUNGGU', true, 'Ruang rawat inap Gedung A lt. 5', null, now() - interval '6 hours'),
  (1, 'Budi (tamu)', 'budi.tamu@transcpg.test', crypt('Demo-TransCPG-2026', gen_salt('bf', 10)),
   'DIREKTUR', false, 'DITOLAK', true, null, 'Bukan pegawai RS — peran Direktur tidak dapat diajukan pihak luar', now() - interval '9 days');
update app_users set reviewed_by = (select id from app_users where email = 'admin@transcpg.test'),
                     reviewed_at = now() - interval '8 days'
where email = 'budi.tamu@transcpg.test';

-- Master ---------------------------------------------------------------------------
insert into mdc_groups (code, name) values
  ('A', 'Penyakit infeksi & parasit'),
  ('D', 'Darah & organ pembentuk darah'),
  ('E', 'Endokrin, nutrisi & metabolik'),
  ('G', 'Sistem saraf pusat'),
  ('I', 'Sistem sirkulasi'),
  ('J', 'Sistem pernapasan'),
  ('K', 'Sistem pencernaan'),
  ('L', 'Kulit & jaringan subkutan'),
  ('N', 'Ginjal & saluran kemih'),
  ('O', 'Kehamilan & persalinan');

insert into icd10_codes (code, name, is_valid) values
  ('A01.0', 'Typhoid fever', true),
  ('A01.1', 'Paratyphoid fever A', true),
  ('A90',   'Dengue fever [classical dengue]', true),
  ('A91',   'Dengue haemorrhagic fever', true),
  ('D64.9', 'Anaemia, unspecified', true),
  ('E11.5', 'Type 2 diabetes mellitus with peripheral circulatory complications', true),
  ('E11.6', 'Type 2 diabetes mellitus with other specified complications', true),
  ('I21.9', 'Acute myocardial infarction, unspecified', true),
  ('I50.0', 'Congestive heart failure', true),
  ('I63.9', 'Cerebral infarction, unspecified', true),
  ('J18.9', 'Pneumonia, unspecified', true),
  ('J44.1', 'Chronic obstructive pulmonary disease with acute exacerbation, unspecified', true),
  ('K29.7', 'Gastritis, unspecified', true),
  ('K52.9', 'Noninfective gastroenteritis and colitis, unspecified', true),
  ('L03.1', 'Cellulitis of other parts of limb', true),
  ('N39.0', 'Urinary tract infection, site not specified', true),
  ('O82',   'Single delivery by caesarean section', true);

insert into icd9cm_codes (code, name) values
  ('99.21', 'Injection of antibiotic'),
  ('99.18', 'Injection or infusion of electrolytes'),
  ('99.04', 'Transfusion of packed cells'),
  ('89.52', 'Electrocardiogram'),
  ('88.72', 'Diagnostic ultrasound of heart'),
  ('38.93', 'Venous catheterization, not elsewhere classified'),
  ('87.03', 'Computerized axial tomography of head'),
  ('87.44', 'Routine chest x-ray, so described'),
  ('93.96', 'Other oxygen enrichment'),
  ('93.94', 'Respiratory medication administered by nebulizer'),
  ('57.94', 'Insertion of indwelling urinary catheter'),
  ('74.1',  'Low cervical cesarean section'),
  ('86.04', 'Other incision with drainage of skin and subcutaneous tissue'),
  ('00.66', 'Percutaneous transluminal coronary angioplasty');

insert into loinc_codes (code, name, component, system) values
  ('58410-2', 'CBC panel - Blood by Automated count', 'CBC panel', 'Bld'),
  ('718-7',   'Hemoglobin [Mass/volume] in Blood', 'Hemoglobin', 'Bld'),
  ('777-3',   'Platelets [#/volume] in Blood by Automated count', 'Platelets', 'Bld'),
  ('4544-3',  'Hematocrit [Volume Fraction] of Blood by Automated count', 'Hematocrit', 'Bld'),
  ('2160-0',  'Creatinine [Mass/volume] in Serum or Plasma', 'Creatinine', 'Ser/Plas'),
  ('2951-2',  'Sodium [Moles/volume] in Serum or Plasma', 'Sodium', 'Ser/Plas'),
  ('2345-7',  'Glucose [Mass/volume] in Serum or Plasma', 'Glucose', 'Ser/Plas'),
  ('4548-4',  'Hemoglobin A1c/Hemoglobin.total in Blood', 'Hemoglobin A1c', 'Bld'),
  ('1988-5',  'C reactive protein [Mass/volume] in Serum or Plasma', 'CRP', 'Ser/Plas'),
  ('33762-6', 'NT-proBNP [Mass/volume] in Serum or Plasma', 'NT-proBNP', 'Ser/Plas'),
  ('10839-9', 'Troponin I.cardiac [Mass/volume] in Serum or Plasma', 'Troponin I', 'Ser/Plas'),
  ('600-7',   'Bacteria identified in Blood by Culture', 'Bacteria identified', 'Bld'),
  ('630-4',   'Bacteria identified in Urine by Culture', 'Bacteria identified', 'Urine'),
  ('24356-8', 'Urinalysis complete panel - Urine', 'Urinalysis panel', 'Urine'),
  ('24336-0', 'Gas panel - Arterial blood', 'Gas panel', 'BldA');

insert into kfa_drugs (code, name, dosage_form, strength) values
  ('KFA-DEV-0001', 'Ceftriaxone 1 g Injeksi',              'Serbuk injeksi',  '1 g'),
  ('KFA-DEV-0002', 'Parasetamol 500 mg Tablet',            'Tablet',          '500 mg'),
  ('KFA-DEV-0003', 'Furosemid 10 mg/mL Injeksi',           'Larutan injeksi', '10 mg/mL'),
  ('KFA-DEV-0004', 'Ringer Laktat Infus 500 mL',           'Infus',           '500 mL'),
  ('KFA-DEV-0005', 'Asam Asetilsalisilat 80 mg Tablet',    'Tablet',          '80 mg'),
  ('KFA-DEV-0006', 'Klopidogrel 75 mg Tablet',             'Tablet',          '75 mg'),
  ('KFA-DEV-0007', 'Atorvastatin 20 mg Tablet',            'Tablet',          '20 mg'),
  ('KFA-DEV-0008', 'Insulin Aspart 100 IU/mL Injeksi',     'Larutan injeksi', '100 IU/mL'),
  ('KFA-DEV-0009', 'Metformin 500 mg Tablet',              'Tablet',          '500 mg'),
  ('KFA-DEV-0010', 'Levofloksasin 750 mg/150 mL Infus',    'Infus',           '750 mg'),
  ('KFA-DEV-0011', 'Salbutamol 2,5 mg/2,5 mL Nebul',       'Larutan inhalasi','2,5 mg'),
  ('KFA-DEV-0012', 'Metilprednisolon 125 mg Injeksi',      'Serbuk injeksi',  '125 mg'),
  ('KFA-DEV-0013', 'Ondansetron 4 mg/2 mL Injeksi',        'Larutan injeksi', '4 mg'),
  ('KFA-DEV-0014', 'Omeprazol 40 mg Injeksi',              'Serbuk injeksi',  '40 mg'),
  ('KFA-DEV-0015', 'Sefiksim 100 mg Kapsul',               'Kapsul',          '100 mg'),
  ('KFA-DEV-0016', 'Siprofloksasin 500 mg Tablet',         'Tablet',          '500 mg'),
  ('KFA-DEV-0017', 'Enoksaparin 60 mg/0,6 mL Injeksi',     'Larutan injeksi', '60 mg'),
  ('KFA-DEV-0018', 'Spironolakton 25 mg Tablet',           'Tablet',          '25 mg'),
  ('KFA-DEV-0019', 'Oksitosin 10 IU/mL Injeksi',           'Larutan injeksi', '10 IU/mL'),
  ('KFA-DEV-0020', 'Sefazolin 1 g Injeksi',                'Serbuk injeksi',  '1 g');

insert into kptl_codes (code, name) values
  ('KPTL-DEV-001', 'Pemberian injeksi antibiotik'),
  ('KPTL-DEV-002', 'Elektrokardiografi (EKG)'),
  ('KPTL-DEV-003', 'Pemasangan kateter vena'),
  ('KPTL-DEV-004', 'CT scan kepala tanpa kontras'),
  ('KPTL-DEV-005', 'Foto toraks PA'),
  ('KPTL-DEV-006', 'Terapi oksigen'),
  ('KPTL-DEV-007', 'Nebulisasi'),
  ('KPTL-DEV-008', 'Pemasangan kateter urin'),
  ('KPTL-DEV-009', 'Seksio sesarea'),
  ('KPTL-DEV-010', 'Ekokardiografi transtorakal'),
  ('KPTL-DEV-011', 'Transfusi darah PRC'),
  ('KPTL-DEV-012', 'Insisi dan drainase abses'),
  ('KPTL-DEV-013', 'Pemberian cairan elektrolit infus'),
  ('KPTL-DEV-014', 'Angioplasti koroner perkutan');

insert into snomed_concepts (concept_id, fsn, preferred_term, semantic_tag, active) values
  ('4834000',   'Typhoid fever (disorder)',                         'Typhoid fever',                         'disorder',  true),
  ('85904008',  'Paratyphoid fever (disorder)',                     'Paratyphoid fever',                     'disorder',  true),
  ('38362002',  'Dengue (disorder)',                                'Dengue',                                'disorder',  true),
  ('20927009',  'Dengue hemorrhagic fever (disorder)',              'Dengue hemorrhagic fever',              'disorder',  true),
  ('271737000', 'Anemia (disorder)',                                'Anemia',                                'disorder',  true),
  ('44054006',  'Diabetes mellitus type 2 (disorder)',              'Diabetes mellitus type 2',              'disorder',  true),
  ('22298006',  'Myocardial infarction (disorder)',                 'Myocardial infarction',                 'disorder',  true),
  ('42343007',  'Congestive heart failure (disorder)',              'Congestive heart failure',              'disorder',  true),
  ('432504007', 'Cerebral infarction (disorder)',                   'Cerebral infarction',                   'disorder',  true),
  ('233604007', 'Pneumonia (disorder)',                             'Pneumonia',                             'disorder',  true),
  ('195951007', 'Acute exacerbation of chronic obstructive airways disease (disorder)', 'Acute exacerbation of COPD', 'disorder', true),
  ('4556007',   'Gastritis (disorder)',                             'Gastritis',                             'disorder',  true),
  ('25374005',  'Gastroenteritis (disorder)',                       'Gastroenteritis',                       'disorder',  true),
  ('128045006', 'Cellulitis (disorder)',                            'Cellulitis',                            'disorder',  true),
  ('68566005',  'Urinary tract infectious disease (disorder)',      'Urinary tract infection',               'disorder',  true),
  ('11466000',  'Cesarean section (procedure)',                     'Cesarean section',                      'procedure', true),
  ('29303009',  'Electrocardiographic procedure (procedure)',       'Electrocardiographic procedure',        'procedure', true),
  ('303653007', 'Computed tomography of head (procedure)',          'CT of head',                            'procedure', true),
  ('399208008', 'Plain chest X-ray (procedure)',                    'Plain chest X-ray',                     'procedure', true),
  ('45211000',  'Catheterization of urinary bladder (procedure)',   'Urinary catheterization',               'procedure', true),
  ('116859006', 'Transfusion of blood product (procedure)',         'Blood transfusion',                     'procedure', true),
  ('40701008',  'Echocardiography (procedure)',                     'Echocardiography',                      'procedure', true),
  ('900000001', 'Contoh konsep pensiun (disorder)',                 'Contoh konsep pensiun',                 'disorder',  false);

insert into clinical_scoring_tools (code, name, setting, critical_threshold) values
  ('SOFA',    'Sequential Organ Failure Assessment',        'ICU',        '≥2 = Sepsis'),
  ('QSOFA',   'Quick SOFA',                                 'IGD',        '≥2 = High risk'),
  ('SIRS',    'Systemic Inflammatory Response Syndrome',    'Umum',       '≥2 kriteria'),
  ('GCS',     'Glasgow Coma Scale',                         'Semua',      '<8 = Severe'),
  ('NEWS2',   'National Early Warning Score 2',             'RI',         '≥5 = Urgent'),
  ('CURB65',  'CURB-65',                                    'IGD/RI',     '≥3 = Severe'),
  ('NIHSS',   'NIH Stroke Scale',                           'Stroke',     '≥16 = Severe'),
  ('APACHE2', 'APACHE II',                                  'ICU',        '>25 = High mortality'),
  ('GRACE',   'Global Registry of Acute Coronary Events',   'Kardiologi', '>140 = High risk'),
  ('HEART',   'History ECG Age Risk Troponin',              'IGD',        '≥7 = High risk'),
  ('PELOD2',  'Pediatric Logistic Organ Dysfunction-2',     'PICU',       '≥10 = High mortality');

-- Grouper & tarif (fiktif) ------------------------------------------------------------
insert into ina_cbg_groupers (code, name, mdc_code) values
  ('A-4-13', 'Demam tifoid dan paratifoid',                 'A'),
  ('A-4-10', 'Demam berdarah dengue',                       'A'),
  ('D-4-13', 'Anemia',                                      'D'),
  ('E-4-10', 'Diabetes melitus dengan komplikasi',          'E'),
  ('G-4-14', 'Infark serebral',                             'G'),
  ('I-4-11', 'Infark miokard akut',                         'I'),
  ('I-4-12', 'Gagal jantung',                               'I'),
  ('J-4-15', 'Penyakit paru obstruktif kronik eksaserbasi', 'J'),
  ('J-4-16', 'Pneumonia',                                   'J'),
  ('K-4-12', 'Gastroenteritis akut',                        'K'),
  ('K-4-17', 'Gastritis',                                   'K'),
  ('L-4-13', 'Selulitis',                                   'L'),
  ('N-4-10', 'Infeksi saluran kemih',                       'N'),
  ('O-6-10', 'Persalinan dengan seksio sesarea',            'O');

insert into ina_cbg_tariffs (cbg_code, severity, bpjs_regional, hospital_type, ownership, care_class, tariff)
select g.code, s, r, t, o, c,
       (round(g.base * (1 + 0.45 * (s - 1)) * (1 + 0.2 * (3 - c)) * (1 - 0.03 * (r - 1))
              * case t when 'A' then 1.2 when 'B' then 1.0 else 0.85 end
              * case o when 'SWASTA' then 1.03 else 1 end / 1000) * 1000)::numeric(14,2)
from (values ('A-4-13', 3200000), ('A-4-10', 3600000), ('D-4-13', 2900000), ('E-4-10', 4800000),
             ('G-4-14', 6400000), ('I-4-11', 9800000), ('I-4-12', 5100000), ('J-4-15', 4600000),
             ('J-4-16', 4300000), ('K-4-12', 2500000), ('K-4-17', 2200000), ('L-4-13', 3300000),
             ('N-4-10', 2800000), ('O-6-10', 7200000)) as g(code, base)
cross join generate_series(1, 3) s
cross join generate_series(1, 5) r
cross join (values ('A'), ('B'), ('C'), ('D')) as tt(t)
cross join (values ('PEMERINTAH'), ('SWASTA')) as oo(o)
cross join generate_series(1, 3) c;

-- Episode klaim (fiktif, deterministik) -------------------------------------------------
select setseed(0.42);
insert into claim_episodes (hospital_id, cbg_code, severity, care_class, admission_date, discharge_date, los_days, icu_days, age_years, source_batch)
select 1, g.cbg, g.sev, 1 + floor(random() * 3)::int, x.d, x.d + x.los, x.los,
       case when g.sev = 3 and random() < 0.3 then 1 + floor(random() * 3)::int else 0 end,
       g.age0 + floor(random() * g.age_span)::int, g.batch
from (values
  -- grouper, sev, jumlah, LOS dasar, tanggal awal, rentang hari, batch, usia min, rentang usia
  ('A-4-13', 1, 40, 4,  date '2024-09-01', 540, 'baseline',     12, 50),
  ('A-4-13', 2, 15, 6,  date '2024-09-01', 540, 'baseline',     12, 50),
  ('A-4-13', 3,  5, 9,  date '2024-09-01', 540, 'baseline',     12, 50),
  ('A-4-10', 1, 60, 4,  date '2024-09-01', 700, 'baseline',     10, 45),
  ('A-4-10', 2, 20, 5,  date '2024-09-01', 700, 'baseline',     10, 45),
  ('A-4-10', 3,  6, 8,  date '2024-09-01', 700, 'baseline',     10, 45),
  ('A-4-10', 1,  4, 3,  date '2026-09-16',  12, 'pasca-aktif',  10, 45),
  ('D-4-13', 1, 14, 3,  date '2024-09-01', 700, 'baseline',     20, 60),
  ('D-4-13', 2,  8, 5,  date '2024-09-01', 700, 'baseline',     20, 60),
  ('D-4-13', 3,  3, 7,  date '2024-09-01', 700, 'baseline',     20, 60),
  ('E-4-10', 1, 45, 5,  date '2024-06-01', 700, 'baseline',     40, 40),
  ('E-4-10', 2, 25, 7,  date '2024-06-01', 700, 'baseline',     40, 40),
  ('E-4-10', 3, 10, 10, date '2024-06-01', 700, 'baseline',     40, 40),
  ('E-4-10', 1, 14, 5,  date '2026-05-21', 125, 'pasca-aktif',  40, 40),  -- porsi severity III naik → "Perlu Ditinjau"
  ('E-4-10', 3, 10, 10, date '2026-05-21', 125, 'pasca-aktif',  40, 40),
  ('G-4-14', 1, 25, 6,  date '2024-09-01', 700, 'baseline',     45, 40),
  ('G-4-14', 2, 11, 8,  date '2024-09-01', 700, 'baseline',     45, 40),
  ('G-4-14', 3,  6, 12, date '2024-09-01', 700, 'baseline',     45, 40),
  ('I-4-11', 1, 15, 4,  date '2024-09-01', 700, 'baseline',     40, 40),
  ('I-4-11', 2, 12, 6,  date '2024-09-01', 700, 'baseline',     40, 40),
  ('I-4-11', 3,  8, 9,  date '2024-09-01', 700, 'baseline',     40, 40),
  ('I-4-12', 1, 30, 5,  date '2024-06-01', 600, 'baseline',     45, 40),
  ('I-4-12', 2, 20, 7,  date '2024-06-01', 600, 'baseline',     45, 40),
  ('I-4-12', 3, 12, 10, date '2024-06-01', 600, 'baseline',     45, 40),
  ('I-4-12', 1, 10, 7,  date '2026-03-05', 200, 'pasca-aktif',  45, 40),  -- LOS memanjang → "Perlu Ditinjau"
  ('I-4-12', 3,  6, 11, date '2026-03-05', 200, 'pasca-aktif',  45, 40),
  ('J-4-15', 1, 20, 5,  date '2024-09-01', 700, 'baseline',     50, 35),
  ('J-4-15', 2, 12, 7,  date '2024-09-01', 700, 'baseline',     50, 35),
  ('J-4-15', 3,  6, 10, date '2024-09-01', 700, 'baseline',     50, 35),
  ('J-4-16', 1, 80, 5,  date '2024-06-01', 650, 'baseline',     18, 65),
  ('J-4-16', 2, 40, 7,  date '2024-06-01', 650, 'baseline',     18, 65),
  ('J-4-16', 3, 15, 10, date '2024-06-01', 650, 'baseline',     18, 65),
  ('J-4-16', 1, 30, 4,  date '2026-04-16', 160, 'pasca-aktif',  18, 65),  -- membaik: LOS lebih pendek
  ('J-4-16', 2, 12, 6,  date '2026-04-16', 160, 'pasca-aktif',  18, 65),
  ('K-4-12', 1, 70, 3,  date '2024-09-01', 700, 'baseline',     15, 55),
  ('K-4-12', 2, 20, 4,  date '2024-09-01', 700, 'baseline',     15, 55),
  ('K-4-12', 3,  4, 6,  date '2024-09-01', 700, 'baseline',     15, 55),
  ('K-4-17', 1,  3, 3,  date '2025-01-01', 600, 'baseline',     20, 50),
  ('L-4-13', 1, 18, 5,  date '2024-09-01', 700, 'baseline',     20, 55),
  ('L-4-13', 2,  6, 7,  date '2024-09-01', 700, 'baseline',     20, 55),
  ('L-4-13', 3,  2, 10, date '2024-09-01', 700, 'baseline',     20, 55),
  ('N-4-10', 1, 35, 3,  date '2024-09-01', 700, 'baseline',     18, 60),
  ('N-4-10', 2, 10, 5,  date '2024-09-01', 700, 'baseline',     18, 60),
  ('N-4-10', 3,  3, 8,  date '2024-09-01', 700, 'baseline',     18, 60),
  ('O-6-10', 1, 90, 3,  date '2024-09-01', 700, 'baseline',     20, 20),
  ('O-6-10', 2, 15, 4,  date '2024-09-01', 700, 'baseline',     20, 20),
  ('O-6-10', 3,  3, 6,  date '2024-09-01', 700, 'baseline',     20, 20)
) as g(cbg, sev, n, base_los, start_date, span, batch, age0, age_span)
cross join lateral generate_series(1, g.n) i
cross join lateral (
  select g.start_date + floor(random() * g.span)::int + 0 * i as d,
         greatest(1, g.base_los + floor(random() * 5)::int - 2 + 0 * i) as los
) x;

-- Tarif klaim = tarif INA-CBG profil RSCM (regional 1, tipe A, pemerintah) sesuai kelas rawat
update claim_episodes e set tariff = t.tariff
from ina_cbg_tariffs t
where t.cbg_code = e.cbg_code and t.severity = e.severity and t.care_class = e.care_class
  and t.bpjs_regional = 1 and t.hospital_type = 'A' and t.ownership = 'PEMERINTAH';

insert into claim_episode_diagnoses (episode_id, icd10_code, is_primary)
select e.id,
       case e.cbg_code
         when 'A-4-13' then case when e.id % 5 = 0 then 'A01.1' else 'A01.0' end
         when 'A-4-10' then case when e.id % 3 = 0 then 'A90' else 'A91' end
         when 'D-4-13' then 'D64.9'
         when 'E-4-10' then case when e.id % 4 = 0 then 'E11.6' else 'E11.5' end
         when 'G-4-14' then 'I63.9'
         when 'I-4-11' then 'I21.9'
         when 'I-4-12' then 'I50.0'
         when 'J-4-15' then 'J44.1'
         when 'J-4-16' then 'J18.9'
         when 'K-4-12' then 'K52.9'
         when 'K-4-17' then 'K29.7'
         when 'L-4-13' then 'L03.1'
         when 'N-4-10' then 'N39.0'
         else 'O82' end,
       true
from claim_episodes e;

insert into claim_episode_procedures (episode_id, icd9cm_code)
select e.id, p.code
from claim_episodes e
join (values
  ('A-4-13', '99.21'), ('A-4-10', '99.18'), ('D-4-13', '99.04'), ('E-4-10', '38.93'),
  ('G-4-14', '87.03'), ('I-4-11', '89.52'), ('I-4-11', '00.66'), ('I-4-12', '89.52'),
  ('I-4-12', '88.72'), ('J-4-15', '93.94'), ('J-4-16', '87.44'), ('J-4-16', '93.96'),
  ('K-4-12', '99.18'), ('L-4-13', '86.04'), ('N-4-10', '57.94'), ('O-6-10', '74.1')
) as p(cbg, code) on p.cbg = e.cbg_code
where p.code not in ('00.66', '86.04', '99.04') or e.id % 3 = 0;  -- prosedur tertentu hanya sebagian episode

-- Dokumen panduan ------------------------------------------------------------------------
insert into guideline_documents (title, source_type, issuer, regulation_ref, year, icd10_codes, created_by) values
  ('PNPK Tata Laksana Demam Tifoid',                'PNPK',          'Kemenkes', 'HK.01.07/MENKES/1/2018',   2018, '{A01.0,A01.1}', null),
  ('PNPK Tata Laksana Gagal Jantung',               'PNPK',          'Kemenkes', 'HK.01.07/MENKES/717/2021', 2021, '{I50.0}',       null),
  ('PNPK Tata Laksana Stroke',                      'PNPK',          'Kemenkes', 'HK.01.07/MENKES/394/2019', 2019, '{I63.9}',       null),
  ('PNPK Tata Laksana Infeksi Dengue Anak dan Remaja','PNPK',        'Kemenkes', 'HK.01.07/MENKES/9845/2020',2020, '{A90,A91}',     null),
  ('PNPK Tata Laksana Sindrom Koroner Akut',        'PNPK',          'Kemenkes', 'HK.01.07/MENKES/1032/2019',2019, '{I21.9}',       null),
  ('Pedoman Pneumonia Komuniti (PDPI)',             'PPK_ASOSIASI',  'PDPI',     null,                       2022, '{J18.9}',       null),
  ('Pedoman Pengelolaan Diabetes Melitus Tipe 2 (PERKENI)', 'PPK_ASOSIASI', 'PERKENI', null,               2021, '{E11.5,E11.6}', null),
  ('Panduan Infeksi Saluran Kemih (IAUI)',          'PPK_ASOSIASI',  'IAUI',     null,                       2020, '{N39.0}',       null),
  ('GOLD Report: Strategi Global PPOK',             'INTERNASIONAL', 'GOLD',     null,                       2025, '{J44.1}',       null),
  ('PPK Penyakit Dalam RSCM',                       'PPK_RSCM',      'RSCM',     null,                       2021, '{A01.0,I50.0,E11.5,K52.9}', null),
  ('Panduan Gastroenteritis Akut Dewasa — Tim CP',  'TIM_CP',        'Tim CP RSCM', null,                    2024, '{K52.9}',
   (select id from app_users where email = 'timcp@transcpg.test'));

-- Butir acuan (ringkasan contoh — bukan kutipan teks asli)
insert into guideline_items (document_id, category, title, page, quote, icd10_codes)
select d.id, v.cat, v.title, v.page, v.quote || ' (contoh, bukan teks asli)', d.icd10_codes
from (values
  ('PNPK Tata Laksana Demam Tifoid', 'KRITERIA_DIAGNOSIS', 'Kriteria diagnosis klinis',       '12', 'Demam ≥ 7 hari disertai gejala saluran cerna, dikonfirmasi pemeriksaan penunjang.'),
  ('PNPK Tata Laksana Demam Tifoid', 'TATA_LAKSANA',       'Antibiotik lini pertama',         '23', 'Seftriakson intravena selama 5–7 hari untuk kasus rawat inap.'),
  ('PNPK Tata Laksana Demam Tifoid', 'DOSIS_OBAT',         'Antibiotik oral lanjutan',        '25', 'Sefiksim oral dapat dipakai sebagai terapi lanjutan setelah perbaikan klinis.'),
  ('PNPK Tata Laksana Demam Tifoid', 'KOMPLIKASI',         'Tanda perforasi usus',            '31', 'Nyeri perut hebat dan defans muskular memerlukan konsultasi bedah segera.'),
  ('PNPK Tata Laksana Gagal Jantung', 'DOSIS_OBAT',        'Diuretik intravena',              '41', 'Furosemid intravena untuk kongesti, dosis disesuaikan respons diuresis.'),
  ('PNPK Tata Laksana Gagal Jantung', 'MONITORING',        'Pemantauan fungsi ginjal',        '44', 'Kreatinin dan elektrolit dipantau berkala selama terapi diuretik.'),
  ('PNPK Tata Laksana Gagal Jantung', 'KRITERIA_DIAGNOSIS','Peptida natriuretik',             '18', 'NT-proBNP membantu menegakkan diagnosis gagal jantung akut.'),
  ('PNPK Tata Laksana Gagal Jantung', 'INDIKASI_TERAPI',   'Ekokardiografi',                  '20', 'Ekokardiografi dilakukan untuk menilai fraksi ejeksi.'),
  ('PNPK Tata Laksana Stroke',        'TATA_LAKSANA',      'CT scan kepala segera',           '15', 'CT kepala tanpa kontras dilakukan sesegera mungkin setelah tiba.'),
  ('PNPK Tata Laksana Stroke',        'MONITORING',        'Skala NIHSS',                     '17', 'Derajat defisit neurologis dinilai dengan NIHSS saat masuk dan berkala.'),
  ('PNPK Tata Laksana Stroke',        'DOSIS_OBAT',        'Antiplatelet',                    '29', 'Aspirin diberikan dalam 24–48 jam bila tidak ada kontraindikasi.'),
  ('PNPK Tata Laksana Infeksi Dengue Anak dan Remaja', 'MONITORING', 'Hematokrit dan trombosit', '22', 'Hematokrit dan trombosit dipantau tiap 12–24 jam pada fase kritis.'),
  ('PNPK Tata Laksana Infeksi Dengue Anak dan Remaja', 'TATA_LAKSANA', 'Cairan kristaloid',     '27', 'Cairan kristaloid isotonik diberikan sesuai derajat kebocoran plasma.'),
  ('PNPK Tata Laksana Sindrom Koroner Akut', 'TATA_LAKSANA', 'EKG 10 menit',                   '9', 'EKG 12 sadapan dilakukan dalam 10 menit pertama kontak medis.'),
  ('PNPK Tata Laksana Sindrom Koroner Akut', 'KRITERIA_DIAGNOSIS', 'Troponin',                 '11', 'Troponin jantung diperiksa saat masuk dan diulang sesuai protokol.'),
  ('PNPK Tata Laksana Sindrom Koroner Akut', 'DOSIS_OBAT',  'Antiplatelet ganda',              '24', 'Aspirin dan inhibitor P2Y12 diberikan sebagai terapi awal.'),
  ('Pedoman Pneumonia Komuniti (PDPI)', 'KRITERIA_DIAGNOSIS', 'Skor CURB-65',                  '14', 'Derajat keparahan dinilai dengan CURB-65 untuk menentukan tempat perawatan.'),
  ('Pedoman Pneumonia Komuniti (PDPI)', 'TATA_LAKSANA',     'Antibiotik empiris rawat inap',  '21', 'Fluorokuinolon respirasi atau beta-laktam dengan makrolid untuk rawat inap.'),
  ('Pedoman Pneumonia Komuniti (PDPI)', 'MONITORING',       'Kultur darah',                    '19', 'Kultur darah diambil sebelum antibiotik pertama pada kasus berat.'),
  ('Pedoman Pneumonia Komuniti (PDPI)', 'TATA_LAKSANA',     'Terapi oksigen',                  '23', 'Oksigen diberikan untuk mempertahankan saturasi target.'),
  ('Pedoman Pengelolaan Diabetes Melitus Tipe 2 (PERKENI)', 'MONITORING', 'HbA1c',             '16', 'HbA1c diperiksa untuk menilai kendali glikemik jangka panjang.'),
  ('Pedoman Pengelolaan Diabetes Melitus Tipe 2 (PERKENI)', 'TATA_LAKSANA', 'Insulin basal-bolus', '38', 'Pasien rawat inap dengan hiperglikemia berat dikelola dengan insulin.'),
  ('Pedoman Pengelolaan Diabetes Melitus Tipe 2 (PERKENI)', 'MONITORING', 'Glukosa darah berkala', '40', 'Glukosa darah dipantau sebelum makan dan sebelum tidur.'),
  ('Panduan Infeksi Saluran Kemih (IAUI)', 'KRITERIA_DIAGNOSIS', 'Urinalisis dan kultur',     '8', 'Urinalisis dan kultur urin dilakukan sebelum antibiotik.'),
  ('Panduan Infeksi Saluran Kemih (IAUI)', 'TATA_LAKSANA',  'Antibiotik empiris',              '15', 'Antibiotik empiris disesuaikan pola kuman lokal, lalu hasil kultur.'),
  ('GOLD Report: Strategi Global PPOK', 'TATA_LAKSANA',     'Bronkodilator kerja singkat',     '112', 'Bronkodilator kerja singkat inhalasi adalah terapi awal eksaserbasi.'),
  ('GOLD Report: Strategi Global PPOK', 'DOSIS_OBAT',       'Kortikosteroid sistemik',         '114', 'Kortikosteroid sistemik jangka pendek mempercepat pemulihan.'),
  ('PPK Penyakit Dalam RSCM',          'TATA_LAKSANA',      'Rehidrasi',                       '102', 'Rehidrasi cairan intravena pada dehidrasi sedang–berat.'),
  ('PPK Penyakit Dalam RSCM',          'MONITORING',        'Balans cairan',                   '104', 'Balans cairan dicatat setiap 8 jam.'),
  ('Panduan Gastroenteritis Akut Dewasa — Tim CP', 'TATA_LAKSANA', 'Antiemetik',               '4', 'Ondansetron intravena bila muntah persisten.'),
  ('Panduan Gastroenteritis Akut Dewasa — Tim CP', 'MONITORING',   'Elektrolit',               '5', 'Natrium dan kalium diperiksa pada dehidrasi berat.')
) as v(doc, cat, title, page, quote)
join guideline_documents d on d.title = v.doc;

insert into tkmkb_topics (title, year, document_id) values
  ('Stroke', 2025, (select id from guideline_documents where title = 'PNPK Tata Laksana Stroke')),
  ('Gastroenteritis akut', 2025, null),
  ('Pneumonia', 2025, (select id from guideline_documents where title = 'Pedoman Pneumonia Komuniti (PDPI)'));

-- Clinical Pathway -------------------------------------------------------------------------
-- Langkah 1: grouper ≥ 5 episode → CP Draf (K-4-17 dengan 3 episode dilewati)
select generate_draft_pathways(5);
update clinical_pathways set created_at = timestamptz '2026-01-15 08:00+07', stage_changed_at = timestamptz '2026-01-15 08:00+07',
                             split_decision = 'SATU';
update clinical_pathways set split_decision = 'PECAH' where cbg_code in ('A-4-13', 'A-4-10');
update clinical_pathways set has_pediatric_cohort = true where cbg_code in ('A-4-10', 'A-4-13');

insert into pathway_sub_cps (pathway_id, icd10_code, label, episode_share)
select p.id, v.icd, v.label, v.share
from clinical_pathways p
join (values ('A-4-13', 'A01.0', 'Demam tifoid', 0.80), ('A-4-13', 'A01.1', 'Paratifoid A', 0.20),
             ('A-4-10', 'A91', 'Demam berdarah dengue', 0.67), ('A-4-10', 'A90', 'Demam dengue', 0.33)) as v(cbg, icd, label, share)
  on v.cbg = p.cbg_code;

-- Riwayat pengesahan (menentukan tahap setiap CP)
insert into approval_history (pathway_id, from_stage, to_stage, action, actor_id, actor_role, reason, created_at)
select p.id, h.f, h.t, h.act, u.id, u.role, h.reason, h.at::timestamptz
from (values
  -- Aktif: dievaluasi
  ('I-4-12', 'DRAF',              'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp@transcpg.test',        null, '2026-02-10 10:00+07'),
  ('I-4-12', 'REVIEW_TIM_CP',     'REVIEW_KOMITE',     'MAJUKAN',    'timcp@transcpg.test',       null, '2026-02-17 14:00+07'),
  ('I-4-12', 'REVIEW_KOMITE',     'MENUNGGU_DIREKTUR', 'MAJUKAN',    'komite@transcpg.test',      null, '2026-02-24 09:30+07'),
  ('I-4-12', 'MENUNGGU_DIREKTUR', 'AKTIF',             'MAJUKAN',    'direktur@transcpg.test',    null, '2026-03-01 08:00+07'),
  ('J-4-16', 'DRAF',              'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp2@transcpg.test',       null, '2026-03-20 11:00+07'),
  ('J-4-16', 'REVIEW_TIM_CP',     'REVISI',            'KEMBALIKAN', 'timcp@transcpg.test',
   'Dosis levofloksasin belum sesuai Pedoman Pneumonia PDPI; mohon cantumkan kultur darah sebelum antibiotik.', '2026-03-25 15:00+07'),
  ('J-4-16', 'REVISI',            'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp2@transcpg.test',       null, '2026-03-30 09:00+07'),
  ('J-4-16', 'REVIEW_TIM_CP',     'REVIEW_KOMITE',     'MAJUKAN',    'koordinator@transcpg.test', null, '2026-04-03 13:00+07'),
  ('J-4-16', 'REVIEW_KOMITE',     'MENUNGGU_DIREKTUR', 'MAJUKAN',    'komite@transcpg.test',      null, '2026-04-10 10:00+07'),
  ('J-4-16', 'MENUNGGU_DIREKTUR', 'AKTIF',             'MAJUKAN',    'direktur@transcpg.test',    null, '2026-04-15 08:00+07'),
  ('E-4-10', 'DRAF',              'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp@transcpg.test',        null, '2026-05-02 10:00+07'),
  ('E-4-10', 'REVIEW_TIM_CP',     'REVIEW_KOMITE',     'MAJUKAN',    'timcp@transcpg.test',       null, '2026-05-07 16:00+07'),
  ('E-4-10', 'REVIEW_KOMITE',     'MENUNGGU_DIREKTUR', 'MAJUKAN',    'komite@transcpg.test',      null, '2026-05-14 09:00+07'),
  ('E-4-10', 'MENUNGGU_DIREKTUR', 'AKTIF',             'MAJUKAN',    'direktur@transcpg.test',    null, '2026-05-20 08:00+07'),
  ('A-4-10', 'DRAF',              'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp@transcpg.test',        null, '2026-08-20 10:00+07'),
  ('A-4-10', 'REVIEW_TIM_CP',     'REVIEW_KOMITE',     'MAJUKAN',    'timcp@transcpg.test',       null, '2026-08-28 14:00+07'),
  ('A-4-10', 'REVIEW_KOMITE',     'MENUNGGU_DIREKTUR', 'MAJUKAN',    'komite@transcpg.test',      null, '2026-09-08 09:00+07'),
  ('A-4-10', 'MENUNGGU_DIREKTUR', 'AKTIF',             'MAJUKAN',    'direktur@transcpg.test',    null, '2026-09-15 08:00+07'),
  -- Menunggu Direktur
  ('K-4-12', 'DRAF',              'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp@transcpg.test',        null, '2026-09-05 10:00+07'),
  ('K-4-12', 'REVIEW_TIM_CP',     'REVIEW_KOMITE',     'MAJUKAN',    'koordinator@transcpg.test', null, '2026-09-15 11:00+07'),
  ('K-4-12', 'REVIEW_KOMITE',     'MENUNGGU_DIREKTUR', 'MAJUKAN',    'komite@transcpg.test',      null, '2026-09-28 15:00+07'),
  -- Review Komite
  ('G-4-14', 'DRAF',              'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp@transcpg.test',        null, '2026-09-10 10:00+07'),
  ('G-4-14', 'REVIEW_TIM_CP',     'REVIEW_KOMITE',     'MAJUKAN',    'timcp@transcpg.test',       null, '2026-09-20 14:00+07'),
  ('I-4-11', 'DRAF',              'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp@transcpg.test',        null, '2026-09-22 10:00+07'),
  ('I-4-11', 'REVIEW_TIM_CP',     'REVIEW_KOMITE',     'MAJUKAN',    'koordinator@transcpg.test', null, '2026-10-01 09:00+07'),
  -- Review Tim CP
  ('N-4-10', 'DRAF',              'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp2@transcpg.test',       null, '2026-10-02 13:00+07'),
  -- Perlu Revisi
  ('A-4-13', 'DRAF',              'REVIEW_TIM_CP',     'MAJUKAN',    'dpjp@transcpg.test',        null, '2026-09-25 10:00+07'),
  ('A-4-13', 'REVIEW_TIM_CP',     'REVISI',            'KEMBALIKAN', 'koordinator@transcpg.test',
   'Rencana isi klinis severity II dan III belum disusun; acuan sub-CP Paratifoid A belum ditetapkan.', '2026-09-30 16:00+07')
) as h(cbg, f, t, act, email, reason, at)
join clinical_pathways p on p.cbg_code = h.cbg
join app_users u on u.email = h.email;

-- Tahap terakhir setiap CP mengikuti riwayatnya
update clinical_pathways p set stage = h.to_stage, stage_changed_at = h.created_at,
       activated_at = case when h.to_stage = 'AKTIF' then h.created_at end
from (select distinct on (pathway_id) pathway_id, to_stage, created_at
      from approval_history order by pathway_id, created_at desc) h
where h.pathway_id = p.id;

-- Statistik baseline (klaim sebelum CP aktif)
select refresh_pathway_severity_stats();

-- Acuan standar ---------------------------------------------------------------------------
insert into pathway_references (pathway_id, sub_cp_id, document_id, note, assigned_by, assigned_at)
select p.id, sc.id, d.id, v.note, u.id, v.at::timestamptz
from (values
  ('I-4-12', null,    'PNPK Tata Laksana Gagal Jantung',           null,                          'timcp@transcpg.test',  '2026-02-05 09:00+07'),
  ('I-4-12', null,    'PPK Penyakit Dalam RSCM',                   'Pelengkap untuk komorbid',    'dpjp@transcpg.test',   '2026-02-06 09:00+07'),
  ('J-4-16', null,    'Pedoman Pneumonia Komuniti (PDPI)',         null,                          'dpjp2@transcpg.test',  '2026-03-15 09:00+07'),
  ('E-4-10', null,    'Pedoman Pengelolaan Diabetes Melitus Tipe 2 (PERKENI)', null,              'dpjp@transcpg.test',   '2026-04-28 09:00+07'),
  ('A-4-10', 'A91',   'PNPK Tata Laksana Infeksi Dengue Anak dan Remaja', null,                   'dpjp@transcpg.test',   '2026-08-15 09:00+07'),
  ('A-4-10', 'A90',   'PNPK Tata Laksana Infeksi Dengue Anak dan Remaja', null,                   'dpjp@transcpg.test',   '2026-08-15 09:05+07'),
  ('K-4-12', null,    'Panduan Gastroenteritis Akut Dewasa — Tim CP', null,                       'timcp@transcpg.test',  '2026-09-01 09:00+07'),
  ('K-4-12', null,    'PPK Penyakit Dalam RSCM',                   null,                          'dpjp@transcpg.test',   '2026-09-01 09:10+07'),
  ('G-4-14', null,    'PNPK Tata Laksana Stroke',                  null,                          'timcp@transcpg.test',  '2026-09-05 09:00+07'),
  ('I-4-11', null,    'PNPK Tata Laksana Sindrom Koroner Akut',    null,                          'dpjp@transcpg.test',   '2026-09-18 09:00+07'),
  ('N-4-10', null,    'Panduan Infeksi Saluran Kemih (IAUI)',      null,                          'dpjp2@transcpg.test',  '2026-09-28 09:00+07'),
  ('A-4-13', 'A01.0', 'PNPK Tata Laksana Demam Tifoid',            null,                          'dpjp@transcpg.test',   '2026-09-20 09:00+07'),
  ('J-4-15', null,    'GOLD Report: Strategi Global PPOK',         'Belum ada PNPK nasional terbaru', 'dpjp2@transcpg.test', '2026-09-29 09:00+07')
) as v(cbg, sub, doc, note, email, at)
join clinical_pathways p on p.cbg_code = v.cbg
join guideline_documents d on d.title = v.doc
join app_users u on u.email = v.email
left join pathway_sub_cps sc on sc.pathway_id = p.id and sc.icd10_code = v.sub;

-- Rencana isi klinis ----------------------------------------------------------------------
-- gi = judul butir panduan (harus dari dokumen acuan CP itu); null = DI LUAR PANDUAN
insert into clinical_plan_items (pathway_id, severity, day, item_type, item_code, item_name, dose, route, frequency, duration,
                                 nature, guideline_item_id, source_note, created_by, created_at)
select p.id, v.sev, v.day, v.type, v.code, v.name, v.dose, v.route, v.freq, v.dur, v.nature,
       gi.id, case when gi.id is not null then d.title || ', hlm. ' || gi.page end,
       u.id, v.at::timestamptz
from (values
  -- Gagal jantung (Aktif)
  ('I-4-12', 1, 0, 'LAB',      '33762-6',      'NT-proBNP',                     null,    null,        null,      null,     'WAJIB',       'Peptida natriuretik',      'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 1, 0, 'PROSEDUR', '89.52',        'Elektrokardiogram',             null,    null,        null,      null,     'WAJIB',       null,                       'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 1, 0, 'OBAT',     'KFA-DEV-0003', 'Furosemid 10 mg/mL Injeksi',    '40 mg', 'Intravena', '2×/hari', '3 hari', 'WAJIB',       'Diuretik intravena',       'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 1, 1, 'LAB',      '2160-0',       'Kreatinin serum',               null,    null,        null,      null,     'WAJIB',       'Pemantauan fungsi ginjal', 'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 1, 1, 'PROSEDUR', '88.72',        'Ekokardiografi',                null,    null,        null,      null,     'KONDISIONAL', 'Ekokardiografi',           'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 1, 2, 'OBAT',     'KFA-DEV-0018', 'Spironolakton 25 mg Tablet',    '25 mg', 'Oral',      '1×/hari', 'lanjut', 'KONDISIONAL', null,                       'apoteker@transcpg.test', '2026-02-07'),
  ('I-4-12', 2, 0, 'LAB',      '33762-6',      'NT-proBNP',                     null,    null,        null,      null,     'WAJIB',       'Peptida natriuretik',      'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 2, 0, 'OBAT',     'KFA-DEV-0003', 'Furosemid 10 mg/mL Injeksi',    '40 mg', 'Intravena', '2×/hari', '5 hari', 'WAJIB',       'Diuretik intravena',       'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 2, 1, 'LAB',      '2951-2',       'Natrium serum',                 null,    null,        null,      null,     'WAJIB',       'Pemantauan fungsi ginjal', 'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 3, 0, 'PROSEDUR', '89.52',        'Elektrokardiogram',             null,    null,        null,      null,     'WAJIB',       null,                       'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 3, 0, 'OBAT',     'KFA-DEV-0003', 'Furosemid 10 mg/mL Injeksi',    '80 mg', 'Intravena', '2×/hari', '5 hari', 'WAJIB',       'Diuretik intravena',       'dpjp@transcpg.test', '2026-02-06'),
  ('I-4-12', 3, 0, 'LAB',      '24336-0',      'Analisis gas darah arteri',     null,    null,        null,      null,     'KONDISIONAL', null,                       'dpjp@transcpg.test', '2026-02-06'),
  -- Pneumonia (Aktif)
  ('J-4-16', 1, 0, 'PROSEDUR', '87.44',        'Foto toraks',                   null,    null,        null,      null,     'WAJIB',       null,                       'dpjp2@transcpg.test', '2026-03-16'),
  ('J-4-16', 1, 0, 'LAB',      '600-7',        'Kultur darah',                  null,    null,        null,      null,     'WAJIB',       'Kultur darah',             'dpjp2@transcpg.test', '2026-03-28'),
  ('J-4-16', 1, 0, 'OBAT',     'KFA-DEV-0010', 'Levofloksasin 750 mg Infus',    '750 mg','Intravena', '1×/hari', '5 hari', 'WAJIB',       'Antibiotik empiris rawat inap', 'dpjp2@transcpg.test', '2026-03-28'),
  ('J-4-16', 1, 0, 'LAB',      '1988-5',       'CRP',                           null,    null,        null,      null,     'KONDISIONAL', null,                       'dpjp2@transcpg.test', '2026-03-16'),
  ('J-4-16', 2, 0, 'PROSEDUR', '93.96',        'Terapi oksigen',                null,    null,        null,      null,     'WAJIB',       'Terapi oksigen',           'dpjp2@transcpg.test', '2026-03-16'),
  ('J-4-16', 2, 0, 'OBAT',     'KFA-DEV-0010', 'Levofloksasin 750 mg Infus',    '750 mg','Intravena', '1×/hari', '7 hari', 'WAJIB',       'Antibiotik empiris rawat inap', 'dpjp2@transcpg.test', '2026-03-28'),
  ('J-4-16', 3, 0, 'LAB',      '24336-0',      'Analisis gas darah arteri',     null,    null,        null,      null,     'WAJIB',       null,                       'dpjp2@transcpg.test', '2026-03-16'),
  -- Diabetes (Aktif)
  ('E-4-10', 1, 0, 'LAB',      '4548-4',       'HbA1c',                         null,    null,        null,      null,     'WAJIB',       'HbA1c',                    'dpjp@transcpg.test', '2026-04-29'),
  ('E-4-10', 1, 0, 'LAB',      '2345-7',       'Glukosa darah',                 null,    null,        null,      null,     'WAJIB',       'Glukosa darah berkala',    'dpjp@transcpg.test', '2026-04-29'),
  ('E-4-10', 1, 0, 'OBAT',     'KFA-DEV-0008', 'Insulin Aspart Injeksi',        '4 IU',  'Subkutan',  '3×/hari', 'rawat',  'WAJIB',       'Insulin basal-bolus',      'dpjp@transcpg.test', '2026-04-29'),
  ('E-4-10', 2, 0, 'LAB',      '2160-0',       'Kreatinin serum',               null,    null,        null,      null,     'WAJIB',       null,                       'dpjp@transcpg.test', '2026-04-29'),
  ('E-4-10', 3, 0, 'PROSEDUR', '38.93',        'Pemasangan kateter vena',       null,    null,        null,      null,     'WAJIB',       null,                       'dpjp@transcpg.test', '2026-04-29'),
  -- DBD (Aktif)
  ('A-4-10', 1, 0, 'LAB',      '777-3',        'Trombosit',                     null,    null,        null,      null,     'WAJIB',       'Hematokrit dan trombosit', 'dpjp@transcpg.test', '2026-08-16'),
  ('A-4-10', 1, 0, 'LAB',      '4544-3',       'Hematokrit',                    null,    null,        null,      null,     'WAJIB',       'Hematokrit dan trombosit', 'dpjp@transcpg.test', '2026-08-16'),
  ('A-4-10', 1, 0, 'OBAT',     'KFA-DEV-0004', 'Ringer Laktat Infus',           '500 mL','Intravena', 'tiap 8 jam', '2 hari', 'WAJIB',  'Cairan kristaloid',        'dpjp@transcpg.test', '2026-08-16'),
  ('A-4-10', 1, 0, 'OBAT',     'KFA-DEV-0002', 'Parasetamol 500 mg Tablet',     '500 mg','Oral',      '3×/hari', 'k/p',    'KONDISIONAL', null,                       'dpjp@transcpg.test', '2026-08-16'),
  ('A-4-10', 2, 1, 'LAB',      '777-3',        'Trombosit',                     null,    null,        null,      null,     'WAJIB',       'Hematokrit dan trombosit', 'dpjp@transcpg.test', '2026-08-16'),
  -- Gastroenteritis (Menunggu Direktur)
  ('K-4-12', 1, 0, 'OBAT',     'KFA-DEV-0004', 'Ringer Laktat Infus',           '500 mL','Intravena', 'tiap 6 jam', '1 hari', 'WAJIB',  'Rehidrasi',                'dpjp@transcpg.test', '2026-09-02'),
  ('K-4-12', 1, 0, 'OBAT',     'KFA-DEV-0013', 'Ondansetron 4 mg Injeksi',      '4 mg',  'Intravena', '3×/hari', 'k/p',    'KONDISIONAL', 'Antiemetik',               'dpjp@transcpg.test', '2026-09-02'),
  ('K-4-12', 1, 0, 'LAB',      '2951-2',       'Natrium serum',                 null,    null,        null,      null,     'KONDISIONAL', 'Elektrolit',               'dpjp@transcpg.test', '2026-09-02'),
  ('K-4-12', 2, 0, 'LAB',      '58410-2',      'Darah lengkap',                 null,    null,        null,      null,     'WAJIB',       null,                       'dpjp@transcpg.test', '2026-09-02'),
  -- Stroke (Review Komite)
  ('G-4-14', 1, 0, 'PROSEDUR', '87.03',        'CT scan kepala',                null,    null,        null,      null,     'WAJIB',       'CT scan kepala segera',    'dpjp@transcpg.test', '2026-09-06'),
  ('G-4-14', 1, 1, 'OBAT',     'KFA-DEV-0005', 'Asam Asetilsalisilat 80 mg',    '160 mg','Oral',      '1×/hari', 'lanjut', 'WAJIB',       'Antiplatelet',             'dpjp@transcpg.test', '2026-09-06'),
  ('G-4-14', 2, 0, 'PROSEDUR', '87.03',        'CT scan kepala',                null,    null,        null,      null,     'WAJIB',       'CT scan kepala segera',    'dpjp@transcpg.test', '2026-09-06'),
  ('G-4-14', 2, 1, 'OBAT',     'KFA-DEV-0007', 'Atorvastatin 20 mg Tablet',     '20 mg', 'Oral',      '1×/hari', 'lanjut', 'KONDISIONAL', null,                       'apoteker@transcpg.test', '2026-09-07'),
  -- Infark miokard (Review Komite)
  ('I-4-11', 1, 0, 'PROSEDUR', '89.52',        'Elektrokardiogram',             null,    null,        null,      null,     'WAJIB',       'EKG 10 menit',             'dpjp@transcpg.test', '2026-09-19'),
  ('I-4-11', 1, 0, 'LAB',      '10839-9',      'Troponin I',                    null,    null,        null,      null,     'WAJIB',       'Troponin',                 'dpjp@transcpg.test', '2026-09-19'),
  ('I-4-11', 1, 0, 'OBAT',     'KFA-DEV-0005', 'Asam Asetilsalisilat 80 mg',    '320 mg','Oral',      'dosis awal', '1 kali', 'WAJIB',  'Antiplatelet ganda',       'dpjp@transcpg.test', '2026-09-19'),
  ('I-4-11', 1, 0, 'OBAT',     'KFA-DEV-0006', 'Klopidogrel 75 mg Tablet',      '300 mg','Oral',      'dosis awal', '1 kali', 'WAJIB',  'Antiplatelet ganda',       'dpjp@transcpg.test', '2026-09-19'),
  ('I-4-11', 3, 0, 'PROSEDUR', '00.66',        'Angioplasti koroner (PCI)',     null,    null,        null,      null,     'KONDISIONAL', null,                       'dpjp@transcpg.test', '2026-09-19'),
  -- ISK (Review Tim CP)
  ('N-4-10', 1, 0, 'LAB',      '24356-8',      'Urinalisis',                    null,    null,        null,      null,     'WAJIB',       'Urinalisis dan kultur',    'dpjp2@transcpg.test', '2026-09-29'),
  ('N-4-10', 1, 0, 'LAB',      '630-4',        'Kultur urin',                   null,    null,        null,      null,     'WAJIB',       'Urinalisis dan kultur',    'dpjp2@transcpg.test', '2026-09-29'),
  ('N-4-10', 1, 0, 'OBAT',     'KFA-DEV-0001', 'Ceftriaxone 1 g Injeksi',       '1 g',   'Intravena', '1×/hari', '3 hari', 'WAJIB',       'Antibiotik empiris',       'dpjp2@transcpg.test', '2026-09-29'),
  -- Tifoid (Perlu Revisi — baru severity I)
  ('A-4-13', 1, 0, 'OBAT',     'KFA-DEV-0001', 'Ceftriaxone 1 g Injeksi',       '2 g',   'Intravena', '1×/hari', '5 hari', 'WAJIB',       'Antibiotik lini pertama',  'dpjp@transcpg.test', '2026-09-21'),
  ('A-4-13', 1, 0, 'LAB',      '58410-2',      'Darah lengkap',                 null,    null,        null,      null,     'WAJIB',       null,                       'dpjp@transcpg.test', '2026-09-21'),
  ('A-4-13', 1, 3, 'OBAT',     'KFA-DEV-0015', 'Sefiksim 100 mg Kapsul',        '200 mg','Oral',      '2×/hari', '7 hari', 'KONDISIONAL', 'Antibiotik oral lanjutan', 'dpjp@transcpg.test', '2026-09-21'),
  -- PPOK (Draf)
  ('J-4-15', 1, 0, 'PROSEDUR', '93.94',        'Nebulisasi',                    null,    null,        null,      null,     'WAJIB',       'Bronkodilator kerja singkat', 'dpjp2@transcpg.test', '2026-09-30'),
  ('J-4-15', 1, 0, 'OBAT',     'KFA-DEV-0011', 'Salbutamol 2,5 mg Nebul',       '2,5 mg','Inhalasi',  '4×/hari', '3 hari', 'WAJIB',       'Bronkodilator kerja singkat', 'dpjp2@transcpg.test', '2026-09-30')
) as v(cbg, sev, day, type, code, name, dose, route, freq, dur, nature, gi_title, email, at)
join clinical_pathways p on p.cbg_code = v.cbg
join app_users u on u.email = v.email
left join lateral (
  select i.id, i.page, i.document_id from guideline_items i
  join pathway_references r on r.document_id = i.document_id and r.pathway_id = p.id
  where i.title = v.gi_title limit 1
) gi on true
left join guideline_documents d on d.id = gi.document_id;

-- Kriteria inklusi/eksklusi
insert into pathway_criteria (pathway_id, kind, description, created_by)
select p.id, c.kind, c.descr, (select id from app_users where email = 'dpjp@transcpg.test')
from (values
  ('I-4-12', 'INKLUSI',  'Usia ≥ 18 tahun'),
  ('I-4-12', 'INKLUSI',  'Diagnosis primer gagal jantung kongestif'),
  ('I-4-12', 'EKSKLUSI', 'Kehamilan'),
  ('I-4-12', 'EKSKLUSI', 'Syok kardiogenik saat masuk'),
  ('J-4-16', 'INKLUSI',  'Pneumonia komuniti dengan CURB-65 ≥ 2'),
  ('J-4-16', 'EKSKLUSI', 'Pneumonia nosokomial atau terkait ventilator'),
  ('E-4-10', 'INKLUSI',  'DM tipe 2 dengan komplikasi kronik'),
  ('E-4-10', 'EKSKLUSI', 'Ketoasidosis diabetik'),
  ('A-4-10', 'INKLUSI',  'Infeksi dengue terkonfirmasi NS1/serologi'),
  ('K-4-12', 'INKLUSI',  'Diare akut < 14 hari dengan dehidrasi sedang–berat'),
  ('G-4-14', 'INKLUSI',  'Stroke iskemik akut terkonfirmasi CT'),
  ('G-4-14', 'EKSKLUSI', 'Perdarahan intraserebral'),
  ('I-4-11', 'INKLUSI',  'Infark miokard akut STEMI/NSTEMI')
) as c(cbg, kind, descr)
join clinical_pathways p on p.cbg_code = c.cbg;

-- Riwayat perubahan isi (ringkas): acuan & rencana yang ditambahkan
insert into pathway_change_log (pathway_id, entity, action, stage, payload, actor_id, actor_role, created_at)
select r.pathway_id, 'ACUAN', 'TAMBAH', 'DRAF',
       jsonb_build_object('reference_id', r.id, 'document_id', r.document_id, 'sub_cp_id', r.sub_cp_id),
       u.id, u.role, r.assigned_at
from pathway_references r join app_users u on u.id = r.assigned_by
union all
select i.pathway_id, 'RENCANA', 'TAMBAH', 'DRAF', to_jsonb(i) - 'created_by' - 'updated_by', u.id, u.role, i.created_at
from clinical_plan_items i join app_users u on u.id = i.created_by;

-- Skor klinis, usulan aturan CDSS, indikator mutu ----------------------------------------
insert into pathway_scoring_tools (pathway_id, tool_code)
select p.id, t.tool from clinical_pathways p
join (values ('I-4-12', 'NEWS2'), ('I-4-12', 'GRACE'), ('J-4-16', 'CURB65'), ('J-4-16', 'NEWS2'), ('J-4-16', 'QSOFA'),
             ('G-4-14', 'NIHSS'), ('G-4-14', 'GCS'), ('I-4-11', 'GRACE'), ('I-4-11', 'HEART'),
             ('A-4-10', 'NEWS2'), ('E-4-10', 'NEWS2'), ('J-4-15', 'NEWS2'), ('N-4-10', 'QSOFA')) as t(cbg, tool)
  on t.cbg = p.cbg_code;

insert into pathway_rule_suggestions (pathway_id, rule_type, condition, action, rationale)
select p.id, r.type, r.cond::jsonb, r.act::jsonb, r.why
from clinical_pathways p
join (values
  ('I-4-12', 'ESCALATION',     '{"jika": "NEWS2 ≥ 5"}',                       '{"maka": "Konsultasi DPJP dalam 30 menit"}',               'Usulan dari pola klaim: 18% severity III dirawat ICU'),
  ('I-4-12', 'LAB_ALERT',      '{"jika": "Kreatinin naik ≥ 0,3 mg/dL / 48 jam"}', '{"maka": "Tinjau dosis diuretik"}',                    'Butir PNPK Gagal Jantung: pemantauan fungsi ginjal'),
  ('I-4-12', 'DOCUMENTATION',  '{"jika": "Hari rawat ke-5"}',                  '{"maka": "Catat alasan bila belum pulang (variance)"}',    'LOS p75 baseline = 7 hari'),
  ('J-4-16', 'ESCALATION',     '{"jika": "CURB-65 ≥ 3"}',                      '{"maka": "Pertimbangkan rawat ICU"}',                       'Pedoman PDPI'),
  ('J-4-16', 'STEP_ACTION',    '{"jika": "Hari ke-3 afebris & stabil"}',       '{"maka": "Pertimbangkan switch antibiotik oral"}',          'Mempersingkat LOS'),
  ('E-4-10', 'LAB_ALERT',      '{"jika": "Glukosa < 70 mg/dL"}',               '{"maka": "Protokol hipoglikemia"}',                         'Keselamatan pasien insulin'),
  ('A-4-10', 'CLINICAL_ALERT', '{"jika": "Hematokrit naik ≥ 20%"}',            '{"maka": "Evaluasi kebocoran plasma, naikkan cairan"}',     'PNPK Dengue'),
  ('G-4-14', 'PATHWAY_BRANCH', '{"jika": "Onset < 4,5 jam"}',                  '{"maka": "Evaluasi kelayakan trombolisis"}',                'PNPK Stroke'),
  ('I-4-11', 'DRUG_SAFETY',    '{"jika": "Riwayat perdarahan aktif"}',         '{"maka": "Tunda antiplatelet ganda, konsultasi"}',          'Kontraindikasi')
) as r(cbg, type, cond, act, why) on r.cbg = p.cbg_code;

insert into pathway_quality_indicators (pathway_id, code, name, target, baseline, unit)
select p.id, q.code, q.name, q.target, q.baseline, q.unit
from clinical_pathways p
join (values
  ('I-4-12', 'LOS',   'Median lama rawat severity I',            5.0,  6.0, 'hari'),
  ('I-4-12', 'EKG',   'Kepatuhan EKG hari ke-0',                 95.0, 81.0, '%'),
  ('I-4-12', 'VAR',   'Dokumentasi variance',                    90.0, 40.0, '%'),
  ('J-4-16', 'LOS',   'Median lama rawat severity I',            5.0,  5.5, 'hari'),
  ('J-4-16', 'KULT',  'Kultur darah sebelum antibiotik',         90.0, 55.0, '%'),
  ('E-4-10', 'HBA1C', 'HbA1c diperiksa saat rawat',              95.0, 70.0, '%'),
  ('A-4-10', 'LOS',   'Median lama rawat severity I',            4.0,  4.0, 'hari'),
  ('G-4-14', 'CT',    'CT kepala < 60 menit',                    90.0, 62.0, '%')
) as q(cbg, code, name, target, baseline, unit) on q.cbg = p.cbg_code;

-- Padanan ----------------------------------------------------------------------------------
insert into kptl_mappings (icd9cm_code, kptl_code, note, mapped_by, mapped_at)
select v.icd9, v.kptl, v.note, (select id from app_users where email = 'timcp@transcpg.test'), v.at::timestamptz
from (values
  ('89.52', 'KPTL-DEV-002', null,                         '2026-02-12'),
  ('99.21', 'KPTL-DEV-001', null,                         '2026-02-12'),
  ('38.93', 'KPTL-DEV-003', null,                         '2026-02-13'),
  ('87.03', 'KPTL-DEV-004', 'Sesuai juknis BPJS 2024',    '2026-03-02'),
  ('87.44', 'KPTL-DEV-005', null,                         '2026-03-18'),
  ('93.96', 'KPTL-DEV-006', null,                         '2026-03-18'),
  ('74.1',  'KPTL-DEV-009', null,                         '2026-05-05'),
  ('88.72', 'KPTL-DEV-010', null,                         '2026-05-05'),
  ('99.18', 'KPTL-DEV-013', null,                         '2026-09-03')
) as v(icd9, kptl, note, at);

insert into mapping_audit_log (target, source_code, old_value, new_value, actor_id, created_at)
select 'KPTL', m.icd9cm_code, null, m.kptl_code, m.mapped_by, m.mapped_at from kptl_mappings m;

insert into snomed_mappings (vocabulary, source_code, concept_id, status) values
  ('ICD10',  'A01.0', '4834000',   'OTOMATIS'),
  ('ICD10',  'A01.1', '85904008',  'OTOMATIS'),
  ('ICD10',  'A90',   '38362002',  'OTOMATIS'),
  ('ICD10',  'A91',   '20927009',  'OTOMATIS'),
  ('ICD10',  'D64.9', '271737000', 'OTOMATIS'),
  ('ICD10',  'E11.5', '44054006',  'RAGU'),
  ('ICD10',  'I21.9', '22298006',  'OTOMATIS'),
  ('ICD10',  'I50.0', '42343007',  'OTOMATIS'),
  ('ICD10',  'I63.9', '432504007', 'RAGU'),
  ('ICD10',  'J18.9', '233604007', 'OTOMATIS'),
  ('ICD10',  'J44.1', '195951007', 'OTOMATIS'),
  ('ICD10',  'K29.7', '4556007',   'OTOMATIS'),
  ('ICD10',  'K52.9', '25374005',  'OTOMATIS'),
  ('ICD10',  'N39.0', '68566005',  'OTOMATIS'),
  ('ICD9CM', '89.52', '29303009',  'OTOMATIS'),
  ('ICD9CM', '87.03', '303653007', 'OTOMATIS'),
  ('ICD9CM', '87.44', '399208008', 'OTOMATIS'),
  ('ICD9CM', '74.1',  '11466000',  'OTOMATIS'),
  ('ICD9CM', '57.94', '45211000',  'OTOMATIS'),
  ('ICD9CM', '88.72', '40701008',  'RAGU');
update snomed_mappings set status = 'MANUAL', mapped_by = (select id from app_users where email = 'timcp@transcpg.test'),
                           mapped_at = timestamptz '2026-06-02 10:00+07'
where source_code in ('K52.9', 'J44.1');

-- Jejak aktivitas (masuk, akun, RS, dokumen) — agar Laporan & Audit tampak hidup ---------
insert into activity_log (category, action, target_type, target_id, target_label, summary, detail,
                          actor_id, actor_name, actor_role, hospital_id, ip, created_at)
select 'MASUK', 'MASUK', 'USER', u.id, u.full_name, 'Masuk lewat Email', '{"metode": "Email"}', u.id, u.full_name, u.role,
       u.hospital_id, '10.10.' || (u.id % 20) || '.' || (10 + d.n), now() - make_interval(days => d.n, hours => ((u.id * 3) % 9)::int)
from app_users u
cross join generate_series(0, 13) d(n)
where u.registration_status = 'DISETUJUI' and (u.id + d.n) % 3 <> 0
  and u.email not in ('kepala.perawat@transcpg.test', 'admin.rsc@transcpg.test');

insert into activity_log (category, action, target_type, target_id, target_label, summary, detail,
                          actor_id, actor_name, actor_role, hospital_id, ip, created_at)
select v.cat, v.act, v.ttype, t.id, coalesce(t.full_name, v.label), v.summary, v.detail::jsonb,
       a.id, a.full_name, a.role, coalesce(t.hospital_id, a.hospital_id), '10.10.1.5', now() - v.ago::interval
from (values
  ('MASUK', 'MASUK_GAGAL', 'USER', 'perawat@transcpg.test',  null, 'perawat@transcpg.test', 'Gagal masuk: kata sandi salah', '{"metode": "Email"}', '1 day 3 hours'),
  ('MASUK', 'MASUK_GAGAL', 'USER', 'perawat@transcpg.test',  null, 'perawat@transcpg.test', 'Gagal masuk: kata sandi salah', '{"metode": "Email"}', '1 day 2 hours 58 minutes'),
  ('AKUN',  'TOLAK',       'USER', 'budi.tamu@transcpg.test', null, 'admin@transcpg.test',  'Menolak pendaftaran', '{"alasan": "Bukan pegawai RS"}', '8 days'),
  ('AKUN',  'BUAT',        'USER', 'residen@transcpg.test',  null, 'admin@transcpg.test',   'Membuat akun dengan peran Residen (PPDS)', '{}', '20 days'),
  ('AKUN',  'UBAH_PERAN',  'USER', 'koordinator@transcpg.test', null, 'admin@transcpg.test', 'Peran Tim CP → Koordinator CP', '{"peran_lama": "Tim CP", "peran_baru": "Koordinator CP"}', '30 days'),
  ('AKUN',  'GANTI_SANDI', 'USER', 'direktur@transcpg.test', null, 'direktur@transcpg.test', 'Mengganti kata sandi sendiri (sesi di perangkat lain dikeluarkan)', '{}', '12 days'),
  ('RS',    'UBAH',        'HOSPITAL', null, 'RSUPN Dr. CIPTO MANGUNKUSUMO (RSCM)', 'admin@transcpg.test', 'Mengubah profil rumah sakit', '{"tipe": "A", "regional_bpjs": 1}', '25 days'),
  ('DOKUMEN', 'DAFTARKAN', 'GUIDELINE', null, 'Panduan Gastroenteritis Akut Dewasa — Tim CP', 'timcp@transcpg.test', 'Mendaftarkan dokumen panduan', '{"sumber": "TIM_CP"}', '33 days'),
  ('DOKUMEN', 'TAMBAH_BUTIR', 'GUIDELINE', null, 'Panduan Gastroenteritis Akut Dewasa — Tim CP', 'timcp@transcpg.test', 'Menambah butir acuan: Antiemetik', '{"kategori": "TATA_LAKSANA", "halaman": "4"}', '33 days')
) as v(cat, act, ttype, target_email, label, actor_email, summary, detail, ago)
join app_users a on a.email = v.actor_email
left join app_users t on t.email = v.target_email;

update activity_log l set target_id = d.id
from guideline_documents d where l.target_type = 'GUIDELINE' and l.target_label = d.title;
update activity_log set target_id = (select id from hospitals where is_default) where target_type = 'HOSPITAL';

