import type {
  ApprovalHistoryRow, CandidateRow, CdssRule, ChangeLogRow, CostRow, Criterion, DiagnosisRow,
  Hospital, PathwayDetail, PlanItem, PlanItemInput, ProcedureRow, QualityIndicator, ReferenceRow,
  Requirement, ScoringTool, SeverityStat, SubCP, TariffRow, TransitionResult,
} from '~/types/api'
import type { ApprovalAction, CriterionKind } from '~/types/domain'

/** Seluruh endpoint Halaman Detail CP untuk satu kode grouper. */
export function usePathwayApi(code: MaybeRefOrGetter<string>) {
  const api = useApi()
  const { query: hospitalQuery } = useHospitalProfile()
  const base = () => `/pathways/${encodeURIComponent(toValue(code))}`

  return {
    // Kepala, panel, kartu
    detail: () => api<PathwayDetail>(base(), { query: hospitalQuery.value }),
    transition: (action: ApprovalAction, reason = '') =>
      api<TransitionResult>(`${base()}/transition`, { method: 'POST', body: { action, reason } }),
    history: () => api<{ approvals: ApprovalHistoryRow[], changes: ChangeLogRow[] }>(`${base()}/history`),

    // Tab baca-saja
    flow: () => api<{ split_decision: 'SATU' | 'PECAH' | null, sub_cps: SubCP[] }>(`${base()}/flow`),
    summary: () => api<{ severity_stats: SeverityStat[], cost_distribution: CostRow[] }>(`${base()}/summary`),
    diagnoses: () => api<DiagnosisRow[]>(`${base()}/diagnoses`),
    procedures: () => api<ProcedureRow[]>(`${base()}/procedures`),
    tariffs: () => api<{ hospital: Hospital, tariffs: TariffRow[], cost_distribution: CostRow[] }>(
      `${base()}/tariffs`, { query: hospitalQuery.value }),
    daily: (severity?: number) => api<{ plan: PlanItem[], claim_matrix: null }>(`${base()}/daily`, { query: { severity } }),
    cdssRules: () => api<{ rules: CdssRule[], note: string }>(`${base()}/cdss-rules`),
    scoring: () => api<ScoringTool[]>(`${base()}/scoring`),
    quality: () => api<QualityIndicator[]>(`${base()}/quality`),
    readiness: () => api<{ ready: boolean, requirements: Requirement[] }>(`${base()}/readiness`),

    // Acuan standar
    references: () => api<ReferenceRow[]>(`${base()}/references`),
    candidates: (q = '') => api<CandidateRow[]>(`${base()}/guideline-candidates`, { query: { q } }),
    addReference: (body: { document_id: number, sub_cp_id?: number | null, note?: string | null }) =>
      api<{ id: number }>(`${base()}/references`, { method: 'POST', body }),
    removeReference: (id: number) => api(`${base()}/references/${id}`, { method: 'DELETE' }),

    // Rencana isi klinis
    plan: (severity?: number) => api<PlanItem[]>(`${base()}/plan`, { query: { severity } }),
    addPlanItem: (body: PlanItemInput) => api<PlanItem>(`${base()}/plan`, { method: 'POST', body }),
    updatePlanItem: (id: number, body: PlanItemInput) =>
      api<PlanItem>(`${base()}/plan/${id}`, { method: 'PUT', body }),
    deletePlanItem: (id: number) => api(`${base()}/plan/${id}`, { method: 'DELETE' }),

    // Kriteria inklusi/eksklusi
    criteria: () => api<Criterion[]>(`${base()}/criteria`),
    addCriterion: (kind: CriterionKind, description: string) =>
      api<Criterion>(`${base()}/criteria`, { method: 'POST', body: { kind, description } }),
    deleteCriterion: (id: number) => api(`${base()}/criteria/${id}`, { method: 'DELETE' }),
  }
}

export type PathwayApi = ReturnType<typeof usePathwayApi>

/** Kunci cache useAsyncData per CP — dipakai bersama antar-komponen. */
export const pathwayKey = (code: string, part: string) => `pathway:${code}:${part}`
