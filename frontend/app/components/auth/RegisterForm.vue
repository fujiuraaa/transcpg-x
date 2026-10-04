<script setup lang="ts">
// Formulir pendaftaran mandiri. Peran yang dipilih adalah peran yang
// DIAJUKAN — Admin RS memutuskan peran final saat menyetujui.
import type { Role } from '~/types/domain'

const props = defineProps<{ googleClientId: string | null }>()
const emit = defineEmits<{ success: [email: string] }>()

const { register, registrationRoles } = useAuth()
const { data: roles } = await useAsyncData('auth:register-roles', registrationRoles)

const form = reactive({
  full_name: '',
  email: '',
  phone: '',
  requested_role: undefined as Role | undefined,
  note: '',
  password: '',
  password2: '',
})
const showPassword = ref(false)
const error = ref('')
const busy = ref(false)

// Daftar dengan Google: email diambil dari akun Google (sudah terverifikasi),
// tanpa kata sandi. Nama & email dari Google mengisi formulir otomatis.
const google = ref<{ token: string, email: string } | null>(null)
async function onGoogleToken(token: string) {
  error.value = ''
  try {
    const info = await $fetch<{ email?: string, name?: string }>('https://www.googleapis.com/oauth2/v3/userinfo', {
      headers: { Authorization: `Bearer ${token}` },
    })
    google.value = { token, email: info.email ?? '' }
    if (info.name && !form.full_name) form.full_name = info.name
  }
  catch {
    error.value = 'Gagal membaca data akun Google. Coba lagi.'
  }
}

const validEmail = (v: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)

async function submit() {
  error.value = ''
  if (!form.full_name.trim()) return (error.value = 'Isi nama lengkap.')
  if (!google.value && !validEmail(form.email.trim())) return (error.value = 'Format email tidak valid.')
  if (!form.requested_role) return (error.value = 'Pilih peran yang Anda ajukan.')
  if (!google.value) {
    if (form.password.length < 8) return (error.value = 'Kata sandi minimal 8 karakter.')
    if (form.password !== form.password2) return (error.value = 'Ulangi kata sandi belum sama.')
  }
  busy.value = true
  try {
    const res = await register({
      full_name: form.full_name.trim(),
      email: google.value ? undefined : form.email.trim(),
      phone: form.phone.trim() || null,
      requested_role: form.requested_role,
      password: google.value ? undefined : form.password,
      google_access_token: google.value?.token,
      note: form.note.trim() || null,
    })
    emit('success', res.email)
  }
  catch (err) {
    error.value = apiErrorMessage(err)
  }
  finally {
    busy.value = false
  }
}
</script>

<template>
  <form class="flex w-full flex-col gap-3" novalidate @submit.prevent="submit">
    <!-- Daftar dengan Google -->
    <div v-if="!google" class="flex w-full">
      <AuthGoogleButton :client-id="props.googleClientId" label="Daftar dengan Google" @token="onGoogleToken" @error="error = $event" />
    </div>
    <div v-else class="flex items-center justify-between gap-2 rounded-xl border border-success/40 bg-success/5 px-3 py-2 text-sm">
      <span class="flex min-w-0 items-center gap-2">
        <UIcon name="i-lucide-badge-check" class="size-4 shrink-0 text-success" />
        <span class="truncate">Akun Google: <span class="font-medium">{{ google.email }}</span></span>
      </span>
      <UButton label="Batal" size="xs" color="neutral" variant="link" @click="google = null" />
    </div>

    <div class="my-1 flex w-full items-center">
      <div class="grow border-t border-dashed border-default" />
      <span class="mx-2 text-xs text-dimmed">{{ google ? 'Lengkapi data' : 'Atau isi data berikut' }}</span>
      <div class="grow border-t border-dashed border-default" />
    </div>

    <AuthIconInput v-model="form.full_name" icon="i-lucide-user-round" label="Nama lengkap (beserta gelar)" autocomplete="name" />
    <AuthIconInput v-if="!google" v-model="form.email" icon="i-lucide-mail" label="Email" type="email" autocomplete="email" />
    <AuthIconInput
      v-model="form.phone"
      icon="i-lucide-smartphone"
      label="Nomor telepon (opsional, untuk masuk dengan OTP)"
      type="tel"
      inputmode="tel"
      autocomplete="tel"
    />

    <USelectMenu
      v-model="form.requested_role"
      :items="roles ?? []"
      value-key="value"
      placeholder="Peran yang diajukan"
      icon="i-lucide-id-card"
      size="lg"
      :search-input="false"
      class="w-full"
      :ui="{ base: 'rounded-xl bg-elevated/60 py-2.5' }"
      aria-label="Peran yang diajukan"
    />
    <AuthIconInput v-model="form.note" icon="i-lucide-building-2" label="Unit kerja / SMF (opsional)" />

    <template v-if="!google">
      <AuthIconInput
        v-model="form.password"
        icon="i-lucide-lock"
        label="Kata sandi (min. 8 karakter)"
        :type="showPassword ? 'text' : 'password'"
        autocomplete="new-password"
      >
        <template #trailing>
          <UButton
            :icon="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'"
            color="neutral"
            variant="link"
            size="sm"
            :aria-label="showPassword ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'"
            @click="showPassword = !showPassword"
          />
        </template>
      </AuthIconInput>
      <AuthIconInput
        v-model="form.password2"
        icon="i-lucide-lock-keyhole"
        label="Ulangi kata sandi"
        :type="showPassword ? 'text' : 'password'"
        autocomplete="new-password"
      />
    </template>

    <p class="flex items-start gap-1.5 rounded-lg bg-elevated px-3 py-2 text-xs text-muted">
      <UIcon name="i-lucide-shield-check" class="mt-0.5 size-3.5 shrink-0 text-primary" />
      Akun aktif setelah disetujui Admin RS. Admin dapat menyesuaikan peran sesuai tugas Anda.
    </p>
    <p class="min-h-4 text-xs text-error" role="alert" aria-live="polite">{{ error }}</p>

    <UButton type="submit" label="Kirim pendaftaran" size="lg" block :loading="busy" class="rounded-xl" />
  </form>
</template>
