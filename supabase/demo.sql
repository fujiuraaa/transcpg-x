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
