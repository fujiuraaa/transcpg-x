<script setup lang="ts">
// Tombol "Masuk dengan Google". Memakai popup Google Identity Services
// (token client) — skrip Google hanya dimuat bila login Google aktif.
const props = withDefaults(defineProps<{ clientId: string | null, label?: string }>(), { label: 'Google' })
const emit = defineEmits<{ token: [accessToken: string], error: [message: string] }>()

const loading = ref(false)
const enabled = computed(() => !!props.clientId)

interface TokenResponse { access_token?: string, error?: string }
interface GoogleOAuth2 {
  initTokenClient: (cfg: {
    client_id: string
    scope: string
    callback: (r: TokenResponse) => void
    error_callback?: (e: { type: string }) => void
  }) => { requestAccessToken: () => void }
}
declare global {
  interface Window { google?: { accounts: { oauth2: GoogleOAuth2 } } }
}

let loader: Promise<void> | null = null
function loadGis() {
  loader ??= new Promise((resolve, reject) => {
    if (window.google?.accounts?.oauth2) return resolve()
    const s = document.createElement('script')
    s.src = 'https://accounts.google.com/gsi/client'
    s.async = true
    s.onload = () => resolve()
    s.onerror = () => {
      loader = null
      reject(new Error('Gagal memuat layanan Google. Periksa koneksi internet.'))
    }
    document.head.appendChild(s)
  })
  return loader
}

async function signIn() {
  if (!props.clientId) return
  loading.value = true
  try {
    await loadGis()
    window.google!.accounts.oauth2.initTokenClient({
      client_id: props.clientId,
      scope: 'openid email profile',
      callback: (r) => {
        loading.value = false
        if (r.access_token) emit('token', r.access_token)
        else emit('error', 'Login Google dibatalkan.')
      },
      error_callback: () => {
        loading.value = false
        emit('error', 'Jendela Google ditutup sebelum selesai.')
      },
    }).requestAccessToken()
  }
  catch (err) {
    loading.value = false
    emit('error', (err as Error).message)
  }
}
</script>

<template>
  <button
    type="button"
    class="flex h-12 grow items-center justify-center gap-2 rounded-xl border border-default bg-default text-sm font-medium transition hover:bg-elevated disabled:cursor-not-allowed disabled:opacity-50"
    :disabled="!enabled || loading"
    :title="enabled ? `${label} — akun Google` : 'Google belum diaktifkan: perlu Client ID Google di server'"
    @click="signIn"
  >
    <UIcon v-if="loading" name="i-lucide-loader-circle" class="size-5 animate-spin" />
    <!-- Logo resmi Google (warna merek dipertahankan) -->
    <svg v-else class="size-5" viewBox="0 0 48 48" aria-hidden="true">
      <path fill="#FFC107" d="M43.611 20.083H42V20H24v8h11.303c-1.649 4.657-6.08 8-11.303 8-6.627 0-12-5.373-12-12s5.373-12 12-12c3.059 0 5.842 1.154 7.961 3.039l5.657-5.657C34.046 6.053 29.268 4 24 4 12.955 4 4 12.955 4 24s8.955 20 20 20 20-8.955 20-20c0-1.341-.138-2.65-.389-3.917z" />
      <path fill="#FF3D00" d="m6.306 14.691 6.571 4.819C14.655 15.108 18.961 12 24 12c3.059 0 5.842 1.154 7.961 3.039l5.657-5.657C34.046 6.053 29.268 4 24 4 16.318 4 9.656 8.337 6.306 14.691z" />
      <path fill="#4CAF50" d="M24 44c5.166 0 9.86-1.977 13.409-5.192l-6.19-5.238A11.91 11.91 0 0 1 24 36c-5.202 0-9.619-3.317-11.283-7.946l-6.522 5.025C9.505 39.556 16.227 44 24 44z" />
      <path fill="#1976D2" d="M43.611 20.083H42V20H24v8h11.303a12.04 12.04 0 0 1-4.087 5.571l.003-.002 6.19 5.238C36.971 39.205 44 34 44 24c0-1.341-.138-2.65-.389-3.917z" />
    </svg>
    {{ label }}
  </button>
</template>
