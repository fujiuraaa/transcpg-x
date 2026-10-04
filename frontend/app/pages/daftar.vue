<script setup lang="ts">
// Halaman Daftar — pendaftaran mandiri; akun menunggu persetujuan Admin RS.
definePageMeta({ layout: 'auth', public: true })

const { loginMethods } = useAuth()
const { data: methods } = await useAsyncData('auth:methods', loginMethods)

const sentTo = ref<string | null>(null)
</script>

<template>
  <!-- Selesai: menunggu persetujuan -->
  <AuthCard
    v-if="sentTo"
    icon="i-lucide-mail-check"
    title="Pendaftaran terkirim"
    :subtitle="`Akun ${sentTo} menunggu persetujuan Admin RS. Anda bisa masuk setelah disetujui.`"
  >
    <ol class="mb-6 w-full space-y-2 text-sm">
      <li class="flex items-center gap-2"><UIcon name="i-lucide-circle-check" class="size-4 text-success" /> Data pendaftaran diterima</li>
      <li class="flex items-center gap-2"><UIcon name="i-lucide-hourglass" class="size-4 text-warning" /> Admin RS memeriksa & menetapkan peran</li>
      <li class="flex items-center gap-2 text-muted"><UIcon name="i-lucide-circle" class="size-4" /> Akun aktif — masuk seperti biasa</li>
    </ol>
    <UButton to="/login" label="Kembali ke halaman masuk" icon="i-lucide-arrow-left" size="lg" block class="rounded-xl" />
  </AuthCard>

  <!-- Formulir -->
  <AuthCard
    v-else
    icon="i-lucide-user-plus"
    title="Buat akun"
    subtitle="Daftar untuk menyusun, meninjau, atau mengesahkan Clinical Pathway."
    wide
  >
    <AuthRegisterForm :google-client-id="methods?.google.enabled ? methods.google.client_id : null" @success="sentTo = $event" />
    <p class="mt-6 text-center text-sm text-muted">
      Sudah punya akun?
      <NuxtLink to="/login" class="font-medium text-primary hover:underline">Masuk</NuxtLink>
    </p>
  </AuthCard>
</template>
