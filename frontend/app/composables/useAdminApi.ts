import type { Hospital, SessionUser } from '~/types/api'
import type { Role } from '~/types/domain'

/** Pengaturan: Manajemen User & Setup Rumah Sakit (khusus Admin). */
export function useAdminApi() {
  const api = useApi()
  return {
    users: () => api<SessionUser[]>('/admin/users'),
    createUser: (body: { full_name: string, email: string, phone?: string | null, password: string, role: Role, hospital_id?: number }) =>
      api<SessionUser>('/admin/users', { method: 'POST', body }),
    updateUser: (id: number, body: Partial<{ full_name: string, phone: string, role: Role, is_active: boolean, password: string }>) =>
      api<SessionUser>(`/admin/users/${id}`, { method: 'PATCH', body }),
    deleteUser: (id: number) => api(`/admin/users/${id}`, { method: 'DELETE' }),
    // Permintaan pendaftaran mandiri
    registrations: (status: 'MENUNGGU' | 'DITOLAK' = 'MENUNGGU') =>
      api<SessionUser[]>('/admin/registrations', { query: { status } }),
    approveRegistration: (id: number, role: Role) =>
      api<SessionUser>(`/admin/registrations/${id}/approve`, { method: 'POST', body: { role } }),
    rejectRegistration: (id: number, reason: string) =>
      api<SessionUser>(`/admin/registrations/${id}/reject`, { method: 'POST', body: { reason } }),
    hospital: (id: number) => api<Hospital>(`/hospitals/${id}`),
    updateHospital: (id: number, body: Hospital) => api<Hospital>(`/hospitals/${id}`, { method: 'PUT', body }),
  }
}
