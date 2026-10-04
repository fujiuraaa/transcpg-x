<script setup lang="ts">
// Bingkai kartu halaman Masuk & Daftar: gradasi merah lembut, kotak ikon,
// judul, dan subjudul. Isi formulir lewat slot.
withDefaults(defineProps<{ icon: string, title: string, subtitle?: string, wide?: boolean }>(), { subtitle: undefined, wide: false })
</script>

<template>
  <div
    class="flex w-full flex-col items-center rounded-3xl border border-(--ui-color-primary-100) bg-linear-to-b from-(--ui-color-primary-50)/70 to-(--ui-bg) p-8 shadow-xl shadow-(--ui-color-primary-900)/10 dark:border-(--ui-color-primary-900)/60 dark:from-(--ui-color-primary-950)/40"
    :class="wide ? 'max-w-lg' : 'max-w-sm'"
  >
    <div class="mb-6 flex size-14 items-center justify-center rounded-2xl bg-default shadow-lg shadow-(--ui-color-primary-900)/10">
      <UIcon :name="icon" class="size-7 text-primary" />
    </div>
    <p class="mb-1 flex items-center gap-1.5 text-xs font-medium tracking-wide text-primary uppercase">
      <span class="bg-brand-gradient size-2 rounded-full" aria-hidden="true" /> TransCPG-X
    </p>
    <h1 class="mb-2 text-center text-2xl font-semibold text-highlighted">{{ title }}</h1>
    <p v-if="subtitle" class="mb-6 text-center text-sm text-muted">{{ subtitle }}</p>
    <slot />
  </div>
</template>
