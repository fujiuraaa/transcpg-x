<script setup lang="ts">
// 6 kartu angka Dashboard. "CP tanpa acuan" membuka Library tersaring.
import type { DashboardSummary } from '~/types/api'

const NuxtLink = resolveComponent('NuxtLink')
const props = defineProps<{ cards: DashboardSummary['cards'] | null }>()

const items = computed(() => {
  const c = props.cards
  return [
    { label: 'Clinical Pathway', value: c?.clinical_pathways, hint: 'grouper layak disusun (≥ 5 episode)', icon: 'i-lucide-route', to: '/library', warn: false },
    { label: 'CP tanpa acuan', value: c?.cp_tanpa_acuan, hint: 'acuan standar belum ditetapkan', icon: 'i-lucide-file-question', to: '/library?acuan=belum', warn: !!c?.cp_tanpa_acuan },
    { label: 'Dokumen panduan', value: c?.dokumen_panduan, hint: 'PPK RSCM, PNPK, PPK Asosiasi', icon: 'i-lucide-library-big', to: null, warn: false },
    { label: 'Butir acuan klinis', value: c?.butir_acuan, hint: 'diekstrak dari dokumen', icon: 'i-lucide-list-tree', to: null, warn: false },
    { label: 'Topik TKMKB', value: c?.topik_tkmkb, hint: 'rekomendasi audit medis BPJS', icon: 'i-lucide-gavel', to: null, warn: false },
    { label: 'Sistem skoring', value: c?.sistem_skoring, hint: 'skoring klinis internasional', icon: 'i-lucide-calculator', to: null, warn: false },
  ]
})
</script>

<template>
  <div class="grid grid-cols-2 gap-3 md:grid-cols-3 2xl:grid-cols-6">
    <component
      :is="it.to ? NuxtLink : 'div'"
      v-for="it in items"
      :key="it.label"
      :to="it.to ?? undefined"
      class="rounded-lg border p-4"
      :class="[
        it.warn ? 'border-warning/50 bg-warning/5' : 'border-default bg-default',
        it.to ? 'transition hover:border-primary/40 hover:shadow-sm' : '',
      ]"
    >
      <div class="flex items-center justify-between text-xs" :class="it.warn ? 'text-warning' : 'text-muted'">
        <span>{{ it.label }}</span>
        <UIcon :name="it.icon" class="size-4" />
      </div>
      <p class="mt-1 text-2xl font-semibold tabular-nums text-highlighted">
        <template v-if="it.value != null">{{ formatNumber(it.value) }}</template>
        <USkeleton v-else class="h-7 w-12" />
      </p>
      <p class="truncate text-xs text-muted" :title="it.hint">
        {{ it.hint }}<UIcon v-if="it.to" name="i-lucide-arrow-right" class="ml-1 inline size-3" />
      </p>
    </component>
  </div>
</template>
