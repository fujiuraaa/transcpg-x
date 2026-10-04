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
