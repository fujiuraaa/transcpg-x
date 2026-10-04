<script setup lang="ts">
// Masuk dengan email + kata sandi. Tidak ada reset mandiri — akun dan
// kata sandi dikelola Admin RS.
const emit = defineEmits<{ success: [] }>()
const { login } = useAuth()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const error = ref('')
const busy = ref(false)
const forgotOpen = ref(false)

const validEmail = (v: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)

async function submit() {
  error.value = ''
  if (!email.value.trim() || !password.value) {
    error.value = 'Isi email dan kata sandi.'
    return
  }
  if (!validEmail(email.value.trim())) {
    error.value = 'Format email tidak valid.'
    return
  }
  busy.value = true
  try {
    await login(email.value.trim(), password.value)
    emit('success')
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
    <AuthIconInput v-model="email" icon="i-lucide-mail" label="Email" type="email" autocomplete="username" autofocus />
    <AuthIconInput
      v-model="password"
      icon="i-lucide-lock"
      label="Kata sandi"
      :type="showPassword ? 'text' : 'password'"
      autocomplete="current-password"
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

    <div class="flex items-start justify-between gap-3 text-xs">
      <p class="text-error" role="alert" aria-live="polite">{{ error }}</p>
      <button type="button" class="shrink-0 font-medium text-toned hover:text-primary hover:underline" @click="forgotOpen = !forgotOpen">
        Lupa kata sandi?
      </button>
    </div>
    <p v-if="forgotOpen" class="rounded-lg bg-elevated px-3 py-2 text-xs text-muted">
      Kata sandi diatur oleh Admin RS. Hubungi Admin RS di rumah sakit Anda untuk mengatur ulang.
    </p>

    <UButton type="submit" label="Masuk" size="lg" block :loading="busy" class="mt-1 rounded-xl" />
  </form>
</template>
