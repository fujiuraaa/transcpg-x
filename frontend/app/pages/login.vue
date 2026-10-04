<script setup lang="ts">
// Halaman Masuk — email + kata sandi, Google, atau nomor telepon (OTP).
// Akun baru dibuat lewat halaman Daftar dan menunggu persetujuan Admin RS.
definePageMeta({ layout: 'auth', public: true })

const { loginMethods, loginWithGoogle } = useAuth()
const { data: methods } = await useAsyncData('auth:methods', loginMethods)

type Mode = 'email' | 'phone'
const mode = ref<Mode>('email')
const googleError = ref('')
const googleNotRegistered = ref(false)
const googleBusy = ref(false)

const COPY: Record<Mode, { title: string, subtitle: string, icon: string }> = {
  email: {
    title: 'Masuk dengan email',
    subtitle: 'Sistem pengelolaan template Clinical Pathway rumah sakit.',
    icon: 'i-lucide-log-in',
  },
  phone: {
    title: 'Masuk dengan nomor telepon',
    subtitle: 'Kami kirim kode sekali pakai ke nomor yang terdaftar di akun Anda.',
    icon: 'i-lucide-smartphone',
  },
}

async function onGoogleToken(accessToken: string) {
  googleError.value = ''
  googleNotRegistered.value = false
  googleBusy.value = true
  try {
    await loginWithGoogle(accessToken)
    await navigateTo('/dashboard')
  }
  catch (err) {
    googleError.value = apiErrorMessage(err)
    googleNotRegistered.value = (err as { data?: { code?: string } }).data?.code === 'BELUM_TERDAFTAR'
  }
  finally {
    googleBusy.value = false
  }
}

const done = () => navigateTo('/dashboard')
</script>

<template>
  <AuthCard :icon="COPY[mode].icon" :title="COPY[mode].title" :subtitle="COPY[mode].subtitle">
    <AuthEmailForm v-if="mode === 'email'" @success="done" />
    <AuthPhoneForm v-else @success="done" />

    <div class="my-5 flex w-full items-center">
      <div class="grow border-t border-dashed border-default" />
      <span class="mx-2 text-xs text-dimmed">Atau masuk dengan</span>
      <div class="grow border-t border-dashed border-default" />
    </div>

    <div class="flex w-full gap-3">
      <AuthGoogleButton
        :client-id="methods?.google.enabled ? methods.google.client_id : null"
        @token="onGoogleToken"
        @error="googleError = $event"
      />
      <button
        v-if="mode === 'email'"
        type="button"
        class="flex h-12 grow items-center justify-center gap-2 rounded-xl border border-default bg-default text-sm font-medium transition hover:bg-elevated disabled:cursor-not-allowed disabled:opacity-50"
        :disabled="!methods?.phone.enabled"
        :title="methods?.phone.enabled ? 'Masuk dengan kode OTP ke nomor telepon' : 'Login nomor telepon belum diaktifkan oleh Admin'"
        @click="mode = 'phone'"
      >
        <UIcon name="i-lucide-smartphone" class="size-5 text-toned" />
        Nomor telepon
      </button>
      <button
        v-else
        type="button"
        class="flex h-12 grow items-center justify-center gap-2 rounded-xl border border-default bg-default text-sm font-medium transition hover:bg-elevated"
        @click="mode = 'email'"
      >
        <UIcon name="i-lucide-mail" class="size-5 text-toned" />
        Email
      </button>
    </div>

    <p v-if="googleBusy" class="mt-3 text-xs text-muted">Memverifikasi akun Google…</p>
    <p v-else-if="googleError" class="mt-3 text-center text-xs text-error" role="alert">
      {{ googleError }}
      <NuxtLink v-if="googleNotRegistered" to="/daftar" class="font-medium underline">Daftar sekarang</NuxtLink>
    </p>
    <p v-else-if="methods && !methods.google.enabled" class="mt-3 text-center text-xs text-dimmed">
      Login Google belum diaktifkan — perlu Client ID Google di server.
    </p>

    <p class="mt-6 text-center text-sm text-muted">
      Belum punya akun?
      <NuxtLink to="/daftar" class="font-medium text-primary hover:underline">Daftar sekarang</NuxtLink>
    </p>
  </AuthCard>
</template>
