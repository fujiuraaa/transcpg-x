import type { AuditEvent, LoginResponse, MeResponse, SessionUser } from '~/types/api'
import type { Role } from '~/types/domain'

export interface RegisterInput {
  full_name: string
  email?: string
  phone?: string | null
  requested_role: Role
  password?: string
  google_access_token?: string
  note?: string | null
}

export interface LoginMethods {
  password: boolean
  google: { enabled: boolean, client_id: string }
  phone: { enabled: boolean }
}

export interface OtpRequestResult {
  masked: string
  expires_in: number
  resend_after: number
  code_length: number
}

/** Sesi pengguna: login (email, Google, OTP telepon), keluar, data diri. */
export function useAuth() {
  const token = useSessionToken()
  const user = useState<SessionUser | null>('auth:user', () => null)
  const approvalBadge = useState<number>('auth:approval-badge', () => 0)
  const registrationBadge = useState<number>('auth:registration-badge', () => 0)
  const api = useApi()

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => !!user.value?.is_admin)

  async function startSession(res: LoginResponse) {
    token.value = res.token
    user.value = res.user
    await refreshMe()
  }

  async function login(email: string, password: string) {
    await startSession(await api<LoginResponse>('/auth/login', { method: 'POST', body: { email, password } }))
  }

  /** Access token dari popup Google → server mencocokkan email terdaftar. */
  async function loginWithGoogle(accessToken: string) {
    await startSession(await api<LoginResponse>('/auth/google', { method: 'POST', body: { access_token: accessToken } }))
  }

  function requestOtp(phone: string) {
    return api<OtpRequestResult>('/auth/otp/request', { method: 'POST', body: { phone } })
  }

  async function verifyOtp(phone: string, code: string) {
    await startSession(await api<LoginResponse>('/auth/otp/verify', { method: 'POST', body: { phone, code } }))
  }

  function loginMethods() {
    return api<LoginMethods>('/auth/methods')
  }

  /** Dipanggil saat aplikasi dibuka dan setelah aksi pengesahan (angka lencana). */
  async function refreshMe() {
    const res = await api<MeResponse>('/auth/me')
    user.value = res.user
    approvalBadge.value = res.approval_badge
    registrationBadge.value = res.registration_badge ?? 0
  }

  /** Pendaftaran mandiri — akun menunggu persetujuan Admin RS. */
  function register(body: RegisterInput) {
    return api<{ status: string, email: string, message: string }>('/auth/register', { method: 'POST', body })
  }

  function registrationRoles() {
    return api<{ value: Role, label: string }[]>('/auth/register/roles')
  }

  /** Profil Saya: ubah nama/telepon sendiri. */
  async function updateProfile(body: { full_name?: string, phone?: string }) {
    user.value = await api<SessionUser>('/auth/me', { method: 'PATCH', body })
  }

  /** Ganti sandi: sesi di perangkat lain dikeluarkan; perangkat ini memakai token baru. */
  async function changePassword(current_password: string, new_password: string) {
    const res = await api<{ token: string }>('/auth/password', { method: 'POST', body: { current_password, new_password } })
    token.value = res.token
  }

  /** Keluar dari semua perangkat (termasuk perangkat ini). */
  async function logoutAll() {
    await api('/auth/logout-all', { method: 'POST' })
    token.value = null
    user.value = null
    approvalBadge.value = 0
    registrationBadge.value = 0
    await navigateTo('/login')
  }

  function myActivity() {
    return api<AuditEvent[]>('/auth/me/activity')
  }

  async function logout() {
    try {
      await api('/auth/logout', { method: 'POST' })
    }
    finally {
      token.value = null
      user.value = null
      approvalBadge.value = 0
      registrationBadge.value = 0
      await navigateTo('/login')
    }
  }

  return {
    user, approvalBadge, registrationBadge, isLoggedIn, isAdmin,
    login, loginWithGoogle, requestOtp, verifyOtp, loginMethods, register, registrationRoles, logout, refreshMe,
    updateProfile, changePassword, logoutAll, myActivity,
  }
}
