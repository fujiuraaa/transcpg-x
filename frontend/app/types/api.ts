// Bentuk respons API. Field laporan yang masih berubah diketik longgar
// (Record) dan dipersempit saat halaman terkait dikerjakan.
import type {
  ApprovalAction, CriterionKind, EvaluationFlagCode, GuidelineCategory, GuidelineSource,
  PlanItemType, PlanNature, RequirementCode, Role, Severity, Stage,
} from './domain'

export type Row = Record<string, unknown>

export interface ApiError {
  code: string
  message: string
  details?: unknown
}

export interface SessionUser {
  id: number
  hospital_id: number | null
  full_name: string
  email: string
  role: Role
  role_label: string
  is_admin: boolean
  is_active: boolean
  last_login_at: string | null
  created_at: string
  phone: string | null
  registration_status: 'MENUNGGU' | 'DISETUJUI' | 'DITOLAK'
  self_registered: boolean
  registration_note: string | null
  rejection_reason: string | null
  reviewed_at: string | null
}

export interface LoginResponse {
  token: string
  expires_at: string
  user: SessionUser
}

export interface MeResponse {
  user: SessionUser
  approval_badge: number
  /** Jumlah pendaftaran menunggu persetujuan (hanya untuk Admin). */
  registration_badge: number
}

export interface Hospital {
  id: number
  name: string
  hospital_type: string
  ownership: 'PEMERINTAH' | 'SWASTA'
  bpjs_regional: number
  max_care_class: number
  bpjs_provider_code: string | null
  accreditation: string | null
  city: string | null
  province: string | null
  address: string | null
  phone: string | null
  email: string | null
  bed_capacity: number | null
  icu_capacity: number | null
  is_default: boolean
}

export interface Meta {
  roles: { value: Role, label: string }[]
  stages: { value: Stage, label: string, advance_label: string, waiting_on: string }[]
  constants: Record<string, number>
}

export interface DashboardSummary {
  cards: {
    clinical_pathways: number
    cp_tanpa_acuan: number
    dokumen_panduan: number
    butir_acuan: number
    topik_tkmkb: number
    sistem_skoring: number
  }
  cases_by_mdc: { mdc_code: string, mdc_name: string, episodes: number, pct: number }[]
  priority: { rank: number, cbg_code: string, name: string, mdc_name: string, episodes: number, stage: Stage, acuan_count: number }[]
  guideline_coverage: { category: GuidelineCategory, items: number }[]
  progress: { acuan_ditetapkan: number, menunggu_penetapan: number, dilewati: number }
}

export interface LibraryQuery {
  q?: string
  mdc?: string
  acuan?: '' | 'sudah' | 'belum'
  sort?: 'volume' | 'los' | 'kode' | 'tanpa_acuan'
  page?: number
}

/** Satu baris CP Library (v_pathway_overview + peringkat & tarif dominan). */
export interface LibraryItem {
  id: number
  cbg_code: string
  name: string
  mdc_code: string
  mdc_name: string
  stage: Stage
  stage_label: string
  stage_changed_at: string
  activated_at: string | null
  split_decision: 'SATU' | 'PECAH' | null
  has_pediatric_cohort: boolean
  episodes: number
  dominant_severity: Severity | null
  target_los: number | null
  sev_episodes: Record<string, number>
  severities_adequate: boolean
  acuan_count: number
  acuan_titles: string | null
  candidate_count: number
  procedure_count: number
  procedure_kptl_mapped: number
  volume_rank: number
  dominant_tariff: number | null
}

export interface Paged<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface PathwayHeader {
  id: number
  cbg_code: string
  name: string
  mdc_code: string
  mdc_name: string
  stage: Stage
  stage_changed_at: string
  activated_at: string | null
  split_decision: 'SATU' | 'PECAH' | null
  has_pediatric_cohort: boolean
  episodes: number
  dominant_severity: Severity | null
  target_los: number | null
  candidate_count: number
}

export interface PathwayPermissions {
  actions: { action: ApprovalAction, label: string, require_reason: boolean }[]
  waiting_on: string
  content_locked: boolean
  can_edit: { acuan: boolean, rencana: boolean, kriteria: boolean }
}

export interface PathwayDetail {
  pathway: PathwayHeader
  stage_label: string
  permissions: PathwayPermissions
  cards: {
    mdc: { code: string, name: string }
    dominant_severity: Severity | null
    target_los: number | null
    tariff: number | null
    candidate_count: number
  }
  hospital: { id: number, name: string }
}

export interface Requirement {
  no: number
  code: RequirementCode
  title: string
  owner: string
  blocking: boolean
  met: boolean
  missing: string[]
}

export interface PlanItem {
  id: number
  severity: Severity
  day: number
  item_type: PlanItemType
  item_code: string
  item_name: string
  dose: string | null
  route: string | null
  frequency: string | null
  duration: string | null
  nature: PlanNature
  guideline_item_id: number | null
  outside_guidance: boolean
  source_note: string | null
  note: string | null
  updated_at: string
}

export type PlanItemInput = Omit<PlanItem, 'id' | 'outside_guidance' | 'updated_at'>

export interface Criterion {
  id: number
  kind: CriterionKind
  description: string
  created_at: string
}

export interface TransitionResult {
  from: Stage
  to: Stage
  to_label: string
}

// --- Tab & panel Halaman Detail CP ----------------------------------------------

export interface SeverityStat {
  severity: Severity
  episodes: number
  los_median: number | null
  los_p75: number | null
  target_los: number | null
  age_median: number | null
}

export interface CostRow {
  severity: Severity
  care_class: number
  episodes: number
  avg_claim_tariff: number | null
}

export interface TariffRow {
  severity: Severity
  care_class: number
  tariff: number
  regulation: string
}

export interface DiagnosisRow {
  severity: Severity
  icd10_code: string
  name: string | null
  is_primary: boolean
  episodes: number
  snomed_concept_id: string | null
  snomed_status: 'OTOMATIS' | 'MANUAL' | 'RAGU' | null
}

export interface ProcedureRow {
  severity: Severity
  icd9cm_code: string
  name: string | null
  episodes: number
  kptl_code: string | null
  kptl_name: string | null
}

export interface ReferenceRow {
  id: number
  sub_cp_id: number | null
  sub_cp_icd10: string | null
  document_id: number
  title: string
  source_type: GuidelineSource
  issuer: string | null
  regulation_ref: string | null
  year: number | null
  note: string | null
  assigned_by_name: string
  assigned_at: string
}

export interface CandidateRow {
  id: number
  title: string
  source_type: GuidelineSource
  issuer: string | null
  regulation_ref: string | null
  year: number | null
  icd10_codes: string[]
  matches_diagnosis: boolean
  item_count: number
  has_attachment: boolean
  attachment_name: string | null
  attachment_size: number | null
}

export interface GuidelineAttachment {
  name: string
  size: number
  uploaded_at: string
  /** URL unduh sementara (± 10 menit). */
  url: string
}

export interface GuidelineItemRow {
  id: number
  category: GuidelineCategory
  title: string
  page: string
  quote: string
  icd10_codes: string[]
}

export interface SubCP {
  id: number
  icd10_code: string
  label: string
  episode_share: number | null
  references: { id: number, document_id: number, title: string }[]
}

export interface ApprovalHistoryRow {
  id: number
  from_stage: Stage
  to_stage: Stage
  action: ApprovalAction
  reason: string | null
  actor_role: Role
  actor_name: string
  created_at: string
}

export interface ChangeLogRow {
  id: number
  entity: 'ACUAN' | 'RENCANA' | 'KRITERIA'
  action: 'TAMBAH' | 'UBAH' | 'HAPUS'
  stage: Stage
  payload: Record<string, unknown>
  actor_role: Role
  actor_name: string
  created_at: string
}

export interface ScoringTool {
  code: string
  name: string
  setting: string | null
  critical_threshold: string | null
}

export interface QualityIndicator {
  id: number
  code: string
  name: string
  target: number | null
  baseline: number | null
  unit: string | null
}

export interface CdssRule {
  id: number
  rule_type: string
  condition: Record<string, unknown>
  action: Record<string, unknown>
  rationale: string | null
}

/** Satu baris antrean halaman Approval. */
export interface ApprovalRow {
  id: number
  cbg_code: string
  name: string
  mdc_name: string
  episodes: number
  stage: Stage
  stage_label: string
  stage_changed_at: string
  acuan_count: number
  waiting_on: string
  can_act: boolean
  /** Tidak ada untuk CP yang sudah Aktif. */
  syarat_aktif_lengkap?: boolean
  syarat_aktif_kurang?: string[]
}

// --- Halaman CP Aktif ------------------------------------------------------------

export interface ActiveSummary {
  cp_aktif: number
  episode_tercakup: number
  kelompok_diagnosis: number
  menunggu_direktur: number
}

export interface ActiveListItem {
  id: number
  cbg_code: string
  name: string
  mdc_name: string
  episodes: number
  activated_at: string
}

export interface ClaimPatternRow {
  name: string | null
  episodes: number
  icd10_code?: string
  icd9cm_code?: string
}

export interface ActiveDetail {
  pathway: PathwayHeader
  plan: PlanItem[]
  claim_pattern: { diagnoses: ClaimPatternRow[], procedures: ClaimPatternRow[] }
  references_approval: { references: ReferenceRow[], approvals: ApprovalHistoryRow[], changes: ChangeLogRow[] }
  criteria: Criterion[]
}

export interface EvaluationFlag {
  code: EvaluationFlagCode
  severity?: Severity
  detail: string
}

export interface EvaluationRow {
  id: number
  cbg_code: string
  name: string
  stage: Stage
  activated_at: string | null
  post_episodes: number
  evaluable: boolean
  needs_review: boolean
  flags: EvaluationFlag[]
}

// --- Halaman Evaluasi CP ------------------------------------------------------------

export interface EvaluationSummary {
  cp_aktif: number
  dapat_dievaluasi: number
  perlu_ditinjau: number
  data_klaim_sampai: string | null
}

export interface QualityRow {
  severity: Severity
  target_los: number | null
  target_p75: number | null
  episodes: number
  los_median: number | null
  los_p75: number | null
  icu_episodes: number
}

export interface EvaluationDetail {
  kendali_mutu: QualityRow[]
  kendali_biaya: {
    bauran_severity: { severity: Severity, sebelum: number, sesudah: number }[]
    bauran_kelas: { care_class: number, episodes: number, avg_tariff: number | null, total_claim: number | null }[]
  }
  catatan: string
}

// --- Padanan KPTL & SNOMED-CT ------------------------------------------------------------

export interface KptlSummary {
  prosedur: number
  sudah: number
  belum: number
  cakupan_episode_pct: number
  can_manage: boolean
}

export interface KptlRow {
  icd9cm_code: string
  name: string | null
  episodes: number
  kptl_code: string | null
  kptl_name: string | null
  mapped_by_name: string | null
  mapped_at: string | null
}

export interface KptlSuggestion {
  code: string
  name: string
  score: number
}

export type SnomedVocab = 'ICD10' | 'ICD9CM'
export type SnomedStatus = 'OTOMATIS' | 'MANUAL' | 'RAGU'

export interface SnomedVocabSummary {
  vocabulary: SnomedVocab
  kode: number
  dipadankan: number
  perlu_ditinjau: number
  belum: number
}

export interface SnomedRow {
  vocabulary: SnomedVocab
  code: string
  name: string | null
  concept_id: string | null
  preferred_term: string | null
  semantic_tag: string | null
  status: SnomedStatus | null
  mapped_at: string | null
}

// --- Laporan & Audit ---------------------------------------------------------------
export type AuditCategory = 'PENGESAHAN' | 'ISI_CP' | 'PADANAN' | 'DOKUMEN' | 'AKUN' | 'RS' | 'MASUK'

export interface AuditEvent {
  uid: string
  created_at: string
  category: AuditCategory
  action: string
  actor_id: number | null
  actor_name: string | null
  actor_role: Role | null
  actor_role_label: string
  target_type: 'CP' | 'KPTL' | 'SNOMED' | 'USER' | 'HOSPITAL' | 'GUIDELINE' | null
  target_id: number | null
  target_code: string | null
  target_label: string | null
  summary: string
  detail: Record<string, unknown>
  ip: string | null
  hospital_name: string | null
}

export interface AuditPage {
  items: AuditEvent[]
  total: number
  page: number
  per_page: number
  counts: Partial<Record<AuditCategory, number>>
}

export interface AuditFilter {
  from?: string
  to?: string
  category?: AuditCategory[]
  actor?: number | null
  q?: string
  page?: number
}

export interface AuditActor {
  id: number
  full_name: string
  role: Role
  role_label: string
}

export interface ApprovalReport {
  totals: { diajukan: number, disahkan: number, dikembalikan: number, dicabut: number }
  stage_durations: { stage: Stage, transitions: number, avg_days: number, max_days: number }[]
  pathways: {
    code: string
    name: string
    stage: Stage
    stage_changed_at: string
    activated_at: string | null
    returned_count: number
    approved_by: string | null
  }[]
}

export interface UserReportRow {
  id: number
  full_name: string
  email: string
  role: Role
  role_label: string
  hospital_name: string | null
  is_active: boolean
  registration_status: 'MENUNGGU' | 'DISETUJUI' | 'DITOLAK'
  last_login_at: string | null
  created_at: string
  login_count: number
  action_count: number
}
