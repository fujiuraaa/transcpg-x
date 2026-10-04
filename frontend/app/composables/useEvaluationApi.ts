import type { EvaluationDetail, EvaluationRow, EvaluationSummary } from '~/types/api'

export function useEvaluationApi() {
  const api = useApi()
  return {
    list: (tab: 'aktif' | 'belum') =>
      api<{ summary: EvaluationSummary, items: EvaluationRow[] }>('/evaluations', { query: { tab } }),
    detail: (code: string) => api<EvaluationDetail>(`/evaluations/${encodeURIComponent(code)}`),
  }
}
