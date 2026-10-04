<script setup lang="ts">
// Kendali Mutu per severity: LOS realisasi (sesudah CP aktif) vs target.
// Batang = median LOS realisasi (abu-abu bila sesuai, kuning bila melewati);
// garis tegak = p75 target (ambang).
// Melewati ambang ditandai warna peringatan + ikon + teks, tidak warna saja.
import type { QualityRow } from '~/types/api'

const props = defineProps<{ rows: QualityRow[] }>()

const scaleMax = computed(() => Math.max(1, ...props.rows.flatMap(r => [r.los_median ?? 0, r.los_p75 ?? 0, r.target_p75 ?? 0])) * 1.15)
const pct = (v: number | null) => `${((v ?? 0) / scaleMax.value) * 100}%`
const over = (r: QualityRow) => r.los_median != null && r.target_p75 != null && r.los_median > r.target_p75
</script>

<template>
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th class="w-24">Severity</th>
          <th class="num w-20">Episode</th>
          <th class="min-w-64">Median LOS realisasi vs p75 target</th>
          <th class="num w-28">p75 realisasi</th>
          <th class="num w-28">Target LOS</th>
          <th class="num w-16">ICU</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.severity">
          <td class="font-medium">Severity {{ severityLabel(r.severity) }}</td>
          <td class="num">{{ formatNumber(r.episodes) }}</td>
          <td>
            <template v-if="r.episodes">
              <div class="relative h-3 rounded-full bg-elevated" role="img" :aria-label="`Median ${r.los_median} hari, p75 target ${r.target_p75} hari`">
                <div
                  class="absolute inset-y-0 left-0 rounded-full"
                  :class="over(r) ? 'bg-(--ui-color-warning-500)' : 'bg-(--ui-color-neutral-400)'"
                  :style="{ width: pct(r.los_median) }"
                />
                <div class="absolute -inset-y-1 w-0.5 rounded bg-(--ui-text-highlighted)" :style="{ left: pct(r.target_p75) }" />
              </div>
              <p class="mt-1 flex items-center gap-1 text-xs" :class="over(r) ? 'font-medium text-warning' : 'text-muted'">
                <UIcon v-if="over(r)" name="i-lucide-trending-up" class="size-3.5" />
                Median {{ formatNumber(r.los_median, 1) }} hr · p75 target {{ formatNumber(r.target_p75, 1) }} hr
                <template v-if="over(r)">· {{ formatNumber(r.los_median! - r.target_p75!, 1) }} hr di atas</template>
              </p>
            </template>
            <span v-else class="text-xs text-muted">Belum ada klaim sesudah CP aktif</span>
          </td>
          <td class="num">{{ r.episodes ? `${formatNumber(r.los_p75, 1)} hr` : '—' }}</td>
          <td class="num">{{ formatNumber(r.target_los, 1) }} hr</td>
          <td class="num">{{ r.episodes ? r.icu_episodes : '—' }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
