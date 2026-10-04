<script setup lang="ts">
// Halaman galat global: 404 (halaman tidak ada) dan galat tak terduga.
import type { NuxtError } from '#app'

const props = defineProps<{ error: NuxtError }>()
const notFound = computed(() => props.error.statusCode === 404)
const offline = computed(() => props.error.statusCode === 503)

useHead({ title: notFound.value ? 'Halaman tidak ditemukan · TransCPG-X' : offline.value ? 'Server tidak dapat dihubungi · TransCPG-X' : 'Terjadi kesalahan · TransCPG-X' })

function home() {
  clearError({ redirect: '/dashboard' })
}
function reload() {
  window.location.reload()
}
function back() {
  if (window.history.length > 1) {
    clearError()
    window.history.back()
  }
  else {
    home()
  }
}
</script>

<template>
  <UApp>
    <main class="grid min-h-screen place-items-center bg-default px-4">
      <div class="animate-in w-full max-w-md text-center">
        <span class="bg-brand-gradient mx-auto grid size-12 place-items-center rounded-xl text-sm font-bold text-white shadow-sm">CP</span>
        <p class="mt-6 text-5xl font-semibold tabular-nums text-primary">{{ error.statusCode || 500 }}</p>
        <h1 class="mt-2 text-xl font-semibold text-highlighted">
          {{ notFound ? 'Halaman tidak ditemukan' : offline ? 'Server tidak dapat dihubungi' : 'Terjadi kesalahan' }}
        </h1>
        <p class="mt-2 text-sm text-muted">
          <template v-if="notFound">Alamat yang Anda buka tidak ada atau sudah dipindahkan.</template>
          <template v-else-if="offline">Periksa koneksi internet Anda, lalu coba lagi. Bila berulang, server mungkin sedang dalam perbaikan.</template>
          <template v-else>Maaf, ada yang tidak berjalan semestinya. Coba muat ulang; bila berulang, beri tahu Admin RS.</template>
        </p>
        <div class="mt-6 flex flex-wrap justify-center gap-2">
          <UButton v-if="offline" label="Coba lagi" icon="i-lucide-refresh-cw" color="neutral" variant="outline" @click="reload" />
          <UButton v-else label="Kembali" icon="i-lucide-arrow-left" color="neutral" variant="outline" @click="back" />
          <UButton label="Ke Dashboard" icon="i-lucide-layout-dashboard" @click="home" />
        </div>
        <details v-if="!notFound && error.message" class="mt-6 text-left text-xs text-muted">
          <summary class="cursor-pointer">Rincian teknis</summary>
          <pre class="mt-2 overflow-x-auto rounded-md bg-elevated p-3 whitespace-pre-wrap">{{ error.message }}</pre>
        </details>
      </div>
    </main>
  </UApp>
</template>
