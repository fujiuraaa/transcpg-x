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
