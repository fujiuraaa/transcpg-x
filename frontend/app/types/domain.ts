// Tipe domain — cermin dari backend/pkg/domain. Ubah keduanya bersamaan.

export type Role =
  | 'SUPER_ADMIN' | 'SYSTEM_ADMIN' | 'ADMIN_RS'
  | 'DOKTER' | 'DPJP' | 'DOKTER_UMUM' | 'RESIDEN'
  | 'KOORDINATOR_CP' | 'TIM_CP'
  | 'PERAWAT_PELAKSANA' | 'KEPALA_PERAWAT' | 'APOTEKER' | 'DIETISIEN'
  | 'KOMITE_MEDIK' | 'DIREKTUR'

export type Stage = 'DRAF' | 'REVISI' | 'REVIEW_TIM_CP' | 'REVIEW_KOMITE' | 'MENUNGGU_DIREKTUR' | 'AKTIF'

export type ApprovalAction = 'MAJUKAN' | 'KEMBALIKAN' | 'CABUT'

export type Severity = 1 | 2 | 3

export type PlanItemType = 'OBAT' | 'LAB' | 'PROSEDUR' | 'TINDAKAN'
export type PlanNature = 'WAJIB' | 'KONDISIONAL'
export type CriterionKind = 'INKLUSI' | 'EKSKLUSI'

export type GuidelineSource = 'PPK_RSCM' | 'PNPK' | 'PPK_ASOSIASI' | 'TIM_CP' | 'INTERNASIONAL' | 'TKMKB' | 'LAIN'
export type GuidelineCategory =
  | 'KRITERIA_DIAGNOSIS' | 'TATA_LAKSANA' | 'DOSIS_OBAT' | 'INDIKASI_TERAPI'
  | 'KONTRAINDIKASI' | 'MONITORING' | 'KOMPLIKASI' | 'LAIN'

export type MasterKind = 'icd10' | 'icd9cm' | 'loinc' | 'kfa' | 'kptl' | 'snomed'

export type RequirementCode =
  | 'ACUAN_CP' | 'ACUAN_SUB_CP' | 'RENCANA_SEVERITY' | 'PADANAN_KPTL'
  | 'TARIF_INACBG' | 'PADANAN_SNOMED' | 'ICD10_MASTER'

export type EvaluationFlagCode = 'MEDIAN_LOS_DI_ATAS_P75' | 'BANYAK_EPISODE_DI_ATAS_P75' | 'PERGESERAN_SEVERITY_III'
