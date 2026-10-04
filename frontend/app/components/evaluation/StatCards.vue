<script setup lang="ts">
// 4 kartu ringkasan Evaluasi CP.
import type { EvaluationSummary } from '~/types/api'

const props = defineProps<{ summary: EvaluationSummary | null }>()

const cards = computed(() => {
  const s = props.summary
  return [
    { label: 'CP Aktif', value: s ? formatNumber(s.cp_aktif) : null, hint: 'sudah disahkan Direktur', icon: 'i-lucide-badge-check', warn: false },
    { label: 'Sudah dapat dievaluasi', value: s ? formatNumber(s.dapat_dievaluasi) : null, hint: 'sudah ada klaim sesudah aktif', icon: 'i-lucide-chart-column', warn: false },
    { label: 'Perlu ditinjau', value: s ? formatNumber(s.perlu_ditinjau) : null, hint: 'melewati salah satu ambang', icon: 'i-lucide-triangle-alert', warn: !!s?.perlu_ditinjau },
    { label: 'Data klaim sampai', value: s ? formatDate(s.data_klaim_sampai) : null, hint: 'tanggal pulang terakhir', icon: 'i-lucide-calendar-check', warn: false },
  ]
})
</script>

<template>
  <div class="grid grid-cols-2 gap-3 xl:grid-cols-4">
    <div
      v-for="c in cards"
      :key="c.label"
      class="rounded-lg border p-4"
      :class="c.warn ? 'border-warning/50 bg-warning/5' : 'border-default bg-default'"
    >
      <div class="flex items-center justify-between text-xs" :class="c.warn ? 'text-warning' : 'text-muted'">
        <span>{{ c.label }}</span>
        <UIcon :name="c.icon" class="size-4" />
      </div>
      <p class="mt-1 text-2xl font-semibold tabular-nums text-highlighted">
        <template v-if="c.value != null">{{ c.value }}</template>
        <USkeleton v-else class="h-7 w-12" />
      </p>
      <p class="truncate text-xs text-muted">{{ c.hint }}</p>
    </div>
  </div>
</template>
