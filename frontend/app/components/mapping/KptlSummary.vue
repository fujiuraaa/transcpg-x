<script setup lang="ts">
// Ringkasan Padanan KPTL: jumlah prosedur, sudah/belum, cakupan episode.
import type { KptlSummary } from '~/types/api'

defineProps<{ summary: KptlSummary | null }>()
</script>

<template>
  <div class="grid grid-cols-2 gap-3 xl:grid-cols-4">
    <div class="rounded-lg border border-default bg-default p-4">
      <p class="flex items-center justify-between text-xs text-muted">Prosedur di data klaim <UIcon name="i-lucide-scan-line" class="size-4" /></p>
      <p class="mt-1 text-2xl font-semibold tabular-nums">{{ summary ? formatNumber(summary.prosedur) : '—' }}</p>
      <p class="text-xs text-muted">kode ICD-9-CM unik</p>
    </div>
    <div class="rounded-lg border border-default bg-default p-4">
      <p class="flex items-center justify-between text-xs text-muted">Sudah dipadankan <UIcon name="i-lucide-link" class="size-4 text-success" /></p>
      <p class="mt-1 text-2xl font-semibold tabular-nums">{{ summary ? formatNumber(summary.sudah) : '—' }}</p>
      <p class="text-xs text-muted">siap ditagihkan</p>
    </div>
    <div class="rounded-lg border p-4" :class="summary?.belum ? 'border-warning/50 bg-warning/5' : 'border-default bg-default'">
      <p class="flex items-center justify-between text-xs" :class="summary?.belum ? 'text-warning' : 'text-muted'">Belum dipadankan <UIcon name="i-lucide-unlink" class="size-4" /></p>
      <p class="mt-1 text-2xl font-semibold tabular-nums">{{ summary ? formatNumber(summary.belum) : '—' }}</p>
      <p class="text-xs text-muted">perlu ditetapkan Tim Koding</p>
    </div>
    <div class="rounded-lg border border-default bg-default p-4">
      <p class="flex items-center justify-between text-xs text-muted">Cakupan episode <UIcon name="i-lucide-gauge" class="size-4" /></p>
      <p class="mt-1 text-2xl font-semibold tabular-nums">{{ summary ? `${formatNumber(summary.cakupan_episode_pct, 1)}%` : '—' }}</p>
      <div class="mt-1.5 h-1.5 overflow-hidden rounded-full bg-elevated" role="img" :aria-label="`Cakupan episode ${summary?.cakupan_episode_pct ?? 0}%`">
        <div class="h-full rounded-full bg-(--ui-color-success-500)" :style="{ width: `${summary?.cakupan_episode_pct ?? 0}%` }" />
      </div>
    </div>
  </div>
</template>
