import type { ApprovalReport, AuditActor, AuditFilter, AuditPage, UserReportRow } from '~/types/api'

export interface Period { from?: string, to?: string }

function toQuery(f: AuditFilter) {
  return {
    from: f.from || undefined,
    to: f.to || undefined,
    category: f.category?.length ? f.category.join(',') : undefined,
    actor: f.actor || undefined,
    q: f.q?.trim() || undefined,
    page: f.page && f.page > 1 ? f.page : undefined,
  }
}

/** Pengaturan › Laporan & Audit (khusus Admin). */
export function useAuditApi() {
  const api = useApi()
  return {
    events: (f: AuditFilter) => api<AuditPage>('/admin/audit/events', { query: toQuery(f) }),
    /** Unduh jejak audit (maks. 10.000 baris) sebagai CSV. */
    async exportEvents(f: AuditFilter) {
      const blob = await api<Blob>('/admin/audit/events.csv', { query: { ...toQuery(f), page: undefined }, responseType: 'blob' })
      saveBlob(blob, `jejak-audit-${todayStamp()}.csv`)
    },
    actors: () => api<AuditActor[]>('/admin/audit/actors'),
    approvalReport: (p: Period) => api<ApprovalReport>('/admin/audit/reports/approval', { query: p }),
    usersReport: (p: Period) => api<UserReportRow[]>('/admin/audit/reports/users', { query: p }),
  }
}
