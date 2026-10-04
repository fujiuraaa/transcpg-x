import type { ActiveDetail, ActiveListItem, ActiveSummary } from '~/types/api'

/** Halaman CP Aktif (pratinjau baca-saja). */
export function useActiveApi() {
  const api = useApi()
  return {
    list: () => api<{ summary: ActiveSummary, items: ActiveListItem[] }>('/active-pathways'),
    detail: (code: string, severity?: number) =>
      api<ActiveDetail>(`/active-pathways/${encodeURIComponent(code)}`, { query: { severity } }),
  }
}
