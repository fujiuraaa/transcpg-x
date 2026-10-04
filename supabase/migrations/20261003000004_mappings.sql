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
