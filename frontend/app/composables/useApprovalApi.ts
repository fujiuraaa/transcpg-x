import type { ApprovalRow } from '~/types/api'
import type { Stage } from '~/types/domain'

export type ApprovalTab = 'saya' | 'proses' | 'aktif'

export function useApprovalApi() {
  const api = useApi()
  return {
    queue: (tab: ApprovalTab) => api<ApprovalRow[]>('/approval', { query: { tab } }),
    counts: () => api<{ by_stage: Record<Stage, number>, mine: number }>('/approval/counts'),
  }
}
