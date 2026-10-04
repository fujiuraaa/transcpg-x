-- TransCPG-X · DATA CONTOH UNTUK PENGEMBANGAN LOKAL
-- Bukan data nyata. Kode KFA/KPTL, tarif, dan episode klaim di sini fiktif;
-- kode ICD-10, ICD-9-CM, LOINC, dan SNOMED-CT memakai contoh kode umum.
--
-- Akun uji (semua peran memakai password yang sama): Dev-TransCPG-2026
--   superadmin@transcpg.test  · admin@transcpg.test    · dpjp@transcpg.test
--   timcp@transcpg.test       · komite@transcpg.test   · direktur@transcpg.test
--   apoteker@transcpg.test    · perawat@transcpg.test
-- Login telepon (OTP): nomor uji +628120000000X — lihat update phone di bawah.
-- Saat dev (OTP_SENDER=log), kode OTP tercetak di log server API.

-- Profil RS ------------------------------------------------------------------------
insert into hospitals (name, hospital_type, ownership, bpjs_regional, max_care_class, city, province, is_default)
values ('RSUPN Dr. CIPTO MANGUNKUSUMO (RSCM)', 'A', 'PEMERINTAH', 1, 1, 'Jakarta Pusat', 'DKI Jakarta', true),
       ('RS Contoh Tipe C Swasta', 'C', 'SWASTA', 3, 1, 'Balikpapan', 'Kalimantan Timur', false);

-- Pengguna -------------------------------------------------------------------------
insert into app_users (hospital_id, full_name, email, password_hash, role)
select 1, u.full_name, u.email, crypt('Dev-TransCPG-2026', gen_salt('bf', 10)), u.role
from (values
  ('Super Admin',         'superadmin@transcpg.test', 'SUPER_ADMIN'),
  ('Admin RS',            'admin@transcpg.test',      'ADMIN_RS'),
  ('dr. Contoh DPJP',     'dpjp@transcpg.test',       'DPJP'),
  ('Koordinator Tim CP',  'timcp@transcpg.test',      'TIM_CP'),
  ('Ketua Komite Medik',  'komite@transcpg.test',     'KOMITE_MEDIK'),
  ('Direktur Medik',      'direktur@transcpg.test',   'DIREKTUR'),
  ('Apoteker Klinis',     'apoteker@transcpg.test',   'APOTEKER'),
  ('Perawat Pelaksana',   'perawat@transcpg.test',    'PERAWAT_PELAKSANA')
) as u(full_name, email, role);

update app_users u set phone = p.phone
from (values
  ('dpjp@transcpg.test',     '+6281200000001'),
  ('timcp@transcpg.test',    '+6281200000002'),
  ('komite@transcpg.test',   '+6281200000003'),
  ('direktur@transcpg.test', '+6281200000004')
) as p(email, phone)
where u.email = p.email;

-- Master ---------------------------------------------------------------------------
insert into mdc_groups (code, name) values
  ('A', 'Penyakit infeksi & parasit'),
  ('I', 'Sistem sirkulasi'),
  ('G', 'Sistem saraf pusat'),
  ('K', 'Sistem pencernaan');

insert into icd10_codes (code, name, is_valid) values
  ('A01.0', 'Typhoid fever', true),
  ('A01.1', 'Paratyphoid fever A', true),
  ('I50.0', 'Congestive heart failure', true),
  ('I63.9', 'Cerebral infarction, unspecified', true),
  ('K29.7', 'Gastritis, unspecified', true);

insert into icd9cm_codes (code, name) values
  ('99.21', 'Injection of antibiotic'),
  ('89.52', 'Electrocardiogram'),
  ('38.93', 'Venous catheterization, not elsewhere classified'),
  ('87.03', 'Computerized axial tomography of head');

insert into loinc_codes (code, name, component, system) values
  ('58410-2', 'CBC panel - Blood by Automated count', 'CBC panel', 'Bld'),
  ('2160-0',  'Creatinine [Mass/volume] in Serum or Plasma', 'Creatinine', 'Ser/Plas'),
  ('33762-6', 'NT-proBNP [Mass/volume] in Serum or Plasma', 'NT-proBNP', 'Ser/Plas');

insert into kfa_drugs (code, name, dosage_form, strength) values
  ('KFA-DEV-0001', 'Ceftriaxone 1 g Injeksi', 'Serbuk injeksi', '1 g'),
  ('KFA-DEV-0002', 'Parasetamol 500 mg Tablet', 'Tablet', '500 mg'),
  ('KFA-DEV-0003', 'Furosemid 10 mg/mL Injeksi', 'Larutan injeksi', '10 mg/mL');

insert into kptl_codes (code, name) values
  ('KPTL-DEV-001', 'Pemberian injeksi antibiotik'),
  ('KPTL-DEV-002', 'Elektrokardiografi (EKG)'),
  ('KPTL-DEV-003', 'Pemasangan kateter vena'),
  ('KPTL-DEV-004', 'CT scan kepala');

insert into snomed_concepts (concept_id, fsn, preferred_term, semantic_tag, active) values
  ('4834000',   'Typhoid fever (disorder)',              'Typhoid fever',              'disorder',  true),
  ('42343007',  'Congestive heart failure (disorder)',   'Congestive heart failure',   'disorder',  true),
  ('432504007', 'Cerebral infarction (disorder)',        'Cerebral infarction',        'disorder',  true),
  ('29303009',  'Electrocardiographic procedure (procedure)', 'Electrocardiographic procedure', 'procedure', true),
  ('900000001', 'Contoh konsep pensiun (disorder)',      'Contoh konsep pensiun',      'disorder',  false);

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
  ('A-4-13', 'Demam tifoid dan paratifoid (contoh)', 'A'),
  ('I-4-12', 'Gagal jantung (contoh)',               'I'),
  ('G-4-14', 'Infark serebral (contoh)',             'G'),
  ('K-4-17', 'Gastritis (contoh, volume rendah)',    'K');

insert into ina_cbg_tariffs (cbg_code, severity, bpjs_regional, hospital_type, ownership, care_class, tariff)
select g.code, s, r, t, o, c,
       (g.base * (1 + 0.45 * (s - 1)) * (1 + 0.2 * (3 - c)) * (1 - 0.03 * (r - 1)) * case t when 'A' then 1.2 when 'B' then 1.0 else 0.85 end)::numeric(14,2)
from (values ('A-4-13', 3200000), ('I-4-12', 5100000), ('G-4-14', 6400000)) as g(code, base)
cross join generate_series(1, 3) s
cross join generate_series(1, 5) r
cross join (values ('A'), ('B'), ('C')) as tt(t)
cross join (values ('PEMERINTAH'), ('SWASTA')) as oo(o)
cross join generate_series(1, 3) c;

-- Episode klaim (fiktif, deterministik) -------------------------------------------------
select setseed(0.42);
insert into claim_episodes (hospital_id, cbg_code, severity, care_class, admission_date, discharge_date, los_days, icu_days, age_years, source_batch)
select 1, g.cbg, g.sev, 1 + floor(random() * 3)::int, x.d, x.d + x.los, x.los,
       case when g.sev = 3 and random() < 0.3 then 1 else 0 end,
       18 + floor(random() * 60)::int, g.batch
from (values
  -- grouper,  sev, jumlah, LOS dasar, tanggal awal,  rentang hari, batch
  ('A-4-13', 1, 40, 4,  date '2025-01-01', 600, 'baseline'),
  ('A-4-13', 2, 15, 6,  date '2025-01-01', 600, 'baseline'),
  ('A-4-13', 3,  5, 9,  date '2025-01-01', 600, 'baseline'),
  ('I-4-12', 1, 30, 5,  date '2024-06-01', 600, 'baseline'),
  ('I-4-12', 2, 20, 7,  date '2024-06-01', 600, 'baseline'),
  ('I-4-12', 3, 12, 10, date '2024-06-01', 600, 'baseline'),
  ('I-4-12', 1, 10, 7,  date '2026-03-05', 180, 'pasca-aktif'),  -- LOS memanjang → "Perlu Ditinjau"
  ('I-4-12', 3,  6, 11, date '2026-03-05', 180, 'pasca-aktif'),
  ('G-4-14', 1, 25, 6,  date '2025-01-01', 600, 'baseline'),
  ('G-4-14', 2, 11, 8,  date '2025-01-01', 600, 'baseline'),
  ('K-4-17', 1,  3, 3,  date '2025-01-01', 600, 'baseline')
) as g(cbg, sev, n, base_los, start_date, span, batch)
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
         when 'I-4-12' then 'I50.0'
         when 'G-4-14' then 'I63.9'
         else 'K29.7' end,
       true
from claim_episodes e;

insert into claim_episode_procedures (episode_id, icd9cm_code)
select e.id, p.code
from claim_episodes e
join (values ('A-4-13', '99.21'), ('I-4-12', '89.52'), ('I-4-12', '38.93'), ('G-4-14', '87.03')) as p(cbg, code)
  on p.cbg = e.cbg_code;

-- Dokumen panduan ------------------------------------------------------------------------
insert into guideline_documents (title, source_type, issuer, regulation_ref, year, icd10_codes) values
  ('PNPK Tata Laksana Demam Tifoid', 'PNPK',        'Kemenkes', 'HK.01.07/MENKES/1/2018',    2018, '{A01.0,A01.1}'),
  ('PNPK Tata Laksana Gagal Jantung', 'PNPK',       'Kemenkes', 'HK.01.07/MENKES/717/2021',  2021, '{I50.0}'),
  ('PNPK Tata Laksana Stroke',        'PNPK',       'Kemenkes', 'HK.01.07/MENKES/394/2019',  2019, '{I63.9}'),
  ('PPK Penyakit Dalam RSCM (contoh)', 'PPK_RSCM',  'RSCM',     null,                        2021, '{A01.0,I50.0}');

insert into guideline_items (document_id, category, title, page, quote, icd10_codes) values
  (1, 'TATA_LAKSANA',       'Antibiotik lini pertama',     '23', 'Contoh kutipan butir acuan (bukan teks asli).', '{A01.0}'),
  (1, 'KRITERIA_DIAGNOSIS', 'Kriteria diagnosis klinis',   '12', 'Contoh kutipan butir acuan (bukan teks asli).', '{A01.0}'),
  (2, 'DOSIS_OBAT',         'Diuretik intravena',          '41', 'Contoh kutipan butir acuan (bukan teks asli).', '{I50.0}'),
  (2, 'MONITORING',         'Pemantauan fungsi ginjal',    '44', 'Contoh kutipan butir acuan (bukan teks asli).', '{I50.0}');

insert into tkmkb_topics (title, year) values ('Stroke (contoh)', 2025), ('Gastroenteritis akut (contoh)', 2025);

-- Langkah 1: grouper ≥5 episode → CP Draf; K-4-17 (3 episode) dilewati ----------------
select generate_draft_pathways(5);
update clinical_pathways set split_decision = 'SATU' where cbg_code in ('I-4-12', 'G-4-14');
update clinical_pathways set split_decision = 'PECAH' where cbg_code = 'A-4-13';
insert into pathway_sub_cps (pathway_id, icd10_code, label, episode_share)
select id, 'A01.0', 'Demam tifoid', 0.8 from clinical_pathways where cbg_code = 'A-4-13'
union all
select id, 'A01.1', 'Paratifoid A', 0.2 from clinical_pathways where cbg_code = 'A-4-13';

-- I-4-12 sudah AKTIF sejak 1 Maret 2026 (baseline = klaim sebelum tanggal itu)
update clinical_pathways set stage = 'AKTIF', activated_at = '2026-03-01', stage_changed_at = '2026-03-01'
where cbg_code = 'I-4-12';
-- G-4-14 sedang ditinjau Komite
update clinical_pathways set stage = 'REVIEW_KOMITE', stage_changed_at = '2026-09-20'
where cbg_code = 'G-4-14';

select refresh_pathway_severity_stats();

insert into pathway_references (pathway_id, document_id, assigned_by)
select p.id, d.id, (select id from app_users where email = 'timcp@transcpg.test')
from clinical_pathways p
join guideline_documents d on (p.cbg_code, d.regulation_ref) in (('I-4-12', 'HK.01.07/MENKES/717/2021'), ('G-4-14', 'HK.01.07/MENKES/394/2019'));

insert into clinical_plan_items (pathway_id, severity, day, item_type, item_code, item_name, dose, route, frequency, duration, nature, guideline_item_id, source_note, created_by)
select p.id, v.sev, v.day, v.type, v.code, v.name, v.dose, v.route, v.freq, v.dur, v.nature, v.gi, v.src,
       (select id from app_users where email = 'timcp@transcpg.test')
from clinical_pathways p
cross join (values
  (1, 0, 'OBAT', 'KFA-DEV-0003', 'Furosemid 10 mg/mL Injeksi', '40 mg', 'Intravena', '2×/hari', '3 hari', 'WAJIB', 3::bigint, 'PNPK Gagal Jantung 2021, hlm. 41'),
  (1, 0, 'LAB',  '33762-6',      'NT-proBNP',                  null,    null,        null,      null,     'WAJIB', null::bigint, null),
  (2, 0, 'OBAT', 'KFA-DEV-0003', 'Furosemid 10 mg/mL Injeksi', '40 mg', 'Intravena', '2×/hari', '5 hari', 'WAJIB', 3::bigint, 'PNPK Gagal Jantung 2021, hlm. 41'),
  (3, 0, 'PROSEDUR', '89.52',    'Elektrokardiogram',          null,    null,        null,      null,     'WAJIB', null::bigint, null)
) as v(sev, day, type, code, name, dose, route, freq, dur, nature, gi, src)
where p.cbg_code = 'I-4-12';

insert into clinical_plan_items (pathway_id, severity, day, item_type, item_code, item_name, nature, created_by)
select p.id, s, 0, 'PROSEDUR', '87.03', 'CT scan kepala', 'WAJIB', (select id from app_users where email = 'dpjp@transcpg.test')
from clinical_pathways p cross join generate_series(1, 2) s
where p.cbg_code = 'G-4-14';

insert into pathway_criteria (pathway_id, kind, description, created_by)
select p.id, k, d, (select id from app_users where email = 'dpjp@transcpg.test')
from clinical_pathways p
cross join (values ('INKLUSI', 'Usia ≥ 18 tahun'), ('INKLUSI', 'Diagnosis primer gagal jantung kongestif'), ('EKSKLUSI', 'Kehamilan')) as c(k, d)
where p.cbg_code = 'I-4-12';

insert into approval_history (pathway_id, from_stage, to_stage, action, actor_id, actor_role, created_at)
select p.id, h.f, h.t, 'MAJUKAN', u.id, u.role, h.at::timestamptz
from clinical_pathways p
cross join (values
  ('DRAF',              'REVIEW_TIM_CP',     'dpjp@transcpg.test',     '2026-02-10'),
  ('REVIEW_TIM_CP',     'REVIEW_KOMITE',     'timcp@transcpg.test',    '2026-02-17'),
  ('REVIEW_KOMITE',     'MENUNGGU_DIREKTUR', 'komite@transcpg.test',   '2026-02-24'),
  ('MENUNGGU_DIREKTUR', 'AKTIF',             'direktur@transcpg.test', '2026-03-01')
) as h(f, t, email, at)
join app_users u on u.email = h.email
where p.cbg_code = 'I-4-12';

insert into pathway_scoring_tools (pathway_id, tool_code)
select p.id, t from clinical_pathways p cross join (values ('NEWS2'), ('GRACE')) as x(t) where p.cbg_code = 'I-4-12'
union all
select p.id, t from clinical_pathways p cross join (values ('NIHSS'), ('GCS')) as x(t) where p.cbg_code = 'G-4-14';

-- Padanan ----------------------------------------------------------------------------------
insert into kptl_mappings (icd9cm_code, kptl_code, mapped_by)
values ('89.52', 'KPTL-DEV-002', (select id from app_users where email = 'timcp@transcpg.test'));

insert into snomed_mappings (vocabulary, source_code, concept_id, status) values
  ('ICD10',  'A01.0', '4834000',   'OTOMATIS'),
  ('ICD10',  'I50.0', '42343007',  'OTOMATIS'),
  ('ICD10',  'I63.9', '432504007', 'RAGU'),
  ('ICD9CM', '89.52', '29303009',  'OTOMATIS');
