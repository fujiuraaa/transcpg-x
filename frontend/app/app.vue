<script setup lang="ts">
// Bahasa Indonesia untuk komponen Nuxt UI (format angka 1.000, label bawaan).
import { id } from '@nuxt/ui/locale'
</script>

<template>
  <UApp :locale="id">
    <!-- Garis muat di atas layar saat berpindah halaman -->
    <NuxtLoadingIndicator color="var(--ui-primary)" :height="3" :throttle="150" />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
