import type { DashboardSummary, Meta, Row } from '~/types/api'

export function useDashboardApi() {
  const api = useApi()
  return {
    summary: () => api<DashboardSummary>('/dashboard'),
    meta: () => api<Meta>('/meta'),
    mdc: () => api<Row[]>('/mdc'),
  }
}
