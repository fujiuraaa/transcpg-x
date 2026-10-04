import type { Role } from '~/types/domain'

// Cermin dari backend/pkg/domain/role.go (AssignableBy). Hanya untuk
// menyembunyikan aksi yang pasti ditolak — server tetap memeriksa ulang.
export const ADMIN_ROLES: Role[] = ['SUPER_ADMIN', 'SYSTEM_ADMIN', 'ADMIN_RS']

export function isAdminRole(r: Role | undefined | null): boolean {
  return !!r && ADMIN_ROLES.includes(r)
}

/** Apakah admin berperan `actor` boleh memberikan / mengelola peran `target`. */
export function canAssignRole(actor: Role | undefined | null, target: Role): boolean {
  if (!isAdminRole(actor)) return false
  if (target === 'SUPER_ADMIN') return actor === 'SUPER_ADMIN'
  if (target === 'SYSTEM_ADMIN') return actor === 'SUPER_ADMIN' || actor === 'SYSTEM_ADMIN'
  return true
}

/** Kata sandi acak yang mudah dibaca (tanpa karakter mirip seperti 0/O, 1/l). */
export function generatePassword(length = 12): string {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789'
  const buf = new Uint32Array(length)
  crypto.getRandomValues(buf)
  return Array.from(buf, n => chars[n % chars.length]).join('')
}
