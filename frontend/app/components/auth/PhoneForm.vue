<script setup lang="ts">
// Masuk dengan nomor telepon: (1) kirim kode OTP, (2) isi 6 digit kode.
// Nomor harus sudah didaftarkan Admin RS pada akun pengguna.
import type { OtpRequestResult } from '~/composables/useAuth'

const emit = defineEmits<{ success: [] }>()
const { requestOtp, verifyOtp } = useAuth()

const phone = ref('')
const sent = ref<OtpRequestResult | null>(null)
// Satu angka per kotak; digit nol di depan tetap terjaga karena digabung per kotak.
const code = ref<(number | undefined)[]>([])
const error = ref('')
const busy = ref(false)

// Hitung mundur kirim ulang.
const resendIn = ref(0)
let timer: ReturnType<typeof setInterval> | undefined
function startCountdown(seconds: number) {
  resendIn.value = seconds
  clearInterval(timer)
  timer = setInterval(() => {
    resendIn.value = Math.max(0, resendIn.value - 1)
    if (!resendIn.value) clearInterval(timer)
  }, 1000)
}
onBeforeUnmount(() => clearInterval(timer))

async function send() {
  error.value = ''
  if (phone.value.replace(/\D/g, '').length < 9) {
    error.value = 'Masukkan nomor telepon yang terdaftar, mis. 0812 3456 7890.'
    return
  }
  busy.value = true
  try {
    sent.value = await requestOtp(phone.value)
    code.value = []
    startCountdown(sent.value.resend_after)
  }
  catch (err) {
    error.value = apiErrorMessage(err)
  }
  finally {
    busy.value = false
  }
}

async function verify() {
  const c = code.value.map(d => (d == null ? '' : String(d))).join('')
  error.value = ''
  if (c.length !== (sent.value?.code_length ?? 6)) {
    error.value = 'Isi semua digit kode.'
    return
  }
  busy.value = true
  try {
    await verifyOtp(phone.value, c)
    emit('success')
  }
  catch (err) {
    error.value = apiErrorMessage(err)
    code.value = []
  }
  finally {
    busy.value = false
  }
}

function changeNumber() {
  sent.value = null
  code.value = []
  error.value = ''
}
</script>

<template>
  <!-- Langkah 1: nomor telepon -->
  <form v-if="!sent" class="flex w-full flex-col gap-3" novalidate @submit.prevent="send">
    <AuthIconInput
      v-model="phone"
      icon="i-lucide-smartphone"
      label="Nomor telepon"
      placeholder="Nomor telepon, mis. 0812 3456 7890"
      type="tel"
      inputmode="tel"
      autocomplete="tel"
      autofocus
    />
    <p class="min-h-4 text-xs text-error" role="alert" aria-live="polite">{{ error }}</p>
    <UButton type="submit" label="Kirim kode OTP" icon="i-lucide-message-square-text" size="lg" block :loading="busy" class="rounded-xl" />
  </form>

  <!-- Langkah 2: kode OTP -->
  <form v-else class="flex w-full flex-col items-center gap-3" novalidate @submit.prevent="verify">
    <p class="text-center text-sm text-muted">
      Masukkan {{ sent.code_length }} digit kode yang dikirim ke
      <span class="font-medium text-highlighted">{{ sent.masked }}</span>.
      Berlaku {{ Math.round(sent.expires_in / 60) }} menit.
    </p>
    <UPinInput
      v-model="code"
      :length="sent.code_length"
      type="number"
      otp
      size="xl"
      autofocus
      aria-label="Kode OTP"
      @complete="verify"
    />
    <p class="min-h-4 text-center text-xs text-error" role="alert" aria-live="polite">{{ error }}</p>
    <UButton type="submit" label="Verifikasi & masuk" size="lg" block :loading="busy" class="rounded-xl" />
    <div class="flex w-full items-center justify-between text-xs">
      <button type="button" class="font-medium text-toned hover:text-primary hover:underline" @click="changeNumber">
        Ganti nomor
      </button>
      <span v-if="resendIn" class="tabular-nums text-muted">Kirim ulang dalam {{ resendIn }} dtk</span>
      <button v-else type="button" class="font-medium text-primary hover:underline" :disabled="busy" @click="send">
        Kirim ulang kode
      </button>
    </div>
  </form>
</template>
