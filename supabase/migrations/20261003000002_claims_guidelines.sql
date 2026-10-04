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
