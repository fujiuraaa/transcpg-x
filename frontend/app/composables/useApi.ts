import type { ApiError } from '~/types/api'

const TOKEN_COOKIE = 'transcpg_token'

/** Cookie token sesi — bertahan sampai pengguna menekan Keluar. */
export function useSessionToken() {
  return useSharedCookie<string | null>(TOKEN_COOKIE, {
    maxAge: 60 * 60 * 24 * 30,
    sameSite: 'strict',
    secure: !import.meta.dev,
  })
}

/**
 * Klien HTTP ke API Go. Menyisipkan token, dan mengarahkan ke /login bila
 * sesi tidak sah (401).
 */
export function useApi() {
  const config = useRuntimeConfig()
  const token = useSessionToken()

  return $fetch.create({
    baseURL: config.public.apiBase,
    // Server tidak menjawab → galat setelah 20 detik, bukan menunggu selamanya.
    timeout: 20_000,
    onRequest({ options }) {
      if (token.value) {
        options.headers.set('Authorization', `Bearer ${token.value}`)
      }
    },
    async onResponseError({ response }) {
      if (response.status === 401 && token.value) {
        token.value = null
        await navigateTo('/login')
      }
    },
  })
}

/** Ambil pesan galat API yang ramah untuk ditampilkan. */
export function apiErrorMessage(err: unknown): string {
  const data = (err as { data?: ApiError })?.data
  const timedOut = (err as { name?: string, cause?: { name?: string } })?.cause?.name === 'TimeoutError'
  const msg = data?.message ?? (timedOut ? 'Server tidak menjawab. Periksa koneksi lalu coba lagi.' : 'Terjadi kesalahan. Coba lagi.')
  return msg.charAt(0).toUpperCase() + msg.slice(1)
}
