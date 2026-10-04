<script setup lang="ts">
// Tab "Pola Klaim": diagnosis primer & prosedur tersering dari data klaim.
// Ini konteks — bukan instruksi klinis. Batang satu warna dengan label angka
// langsung; porsi dihitung terhadap jumlah episode severity terpilih.
import type { ClaimPatternRow } from '~/types/api'

const props = defineProps<{ diagnoses: ClaimPatternRow[], procedures: ClaimPatternRow[] }>()

// Setiap episode punya tepat satu diagnosis primer → jumlahnya = total episode.
const total = computed(() => props.diagnoses.reduce((s, d) => s + d.episodes, 0))

const sections = computed(() => [
  { title: 'Diagnosis primer tersering', icon: 'i-lucide-stethoscope', vocab: 'ICD-10', rows: props.diagnoses.map(d => ({ code: d.icd10_code ?? '', name: d.name, n: d.episodes })) },
  { title: 'Prosedur tersering', icon: 'i-lucide-scan-line', vocab: 'ICD-9-CM', rows: props.procedures.map(p => ({ code: p.icd9cm_code ?? '', name: p.name, n: p.episodes })) },
])
</script>

<template>
  <div class="space-y-4">
    <p class="flex items-center gap-1.5 text-xs text-muted">
      <UIcon name="i-lucide-info" class="size-3.5" />
      Gambaran dari {{ formatNumber(total) }} episode klaim — informasi konteks, bukan instruksi klinis.
    </p>
    <div class="grid gap-4 lg:grid-cols-2">
      <section v-for="s in sections" :key="s.title" class="rounded-lg border border-default bg-default">
        <h4 class="flex items-center gap-2 border-b border-default px-4 py-2.5 text-sm font-medium">
          <UIcon :name="s.icon" class="size-4 text-muted" /> {{ s.title }}
        </h4>
        <p v-if="!s.rows.length" class="px-4 py-6 text-center text-sm text-muted">Tidak ada data.</p>
        <ul v-else class="space-y-3 p-4">
          <li v-for="r in s.rows" :key="r.code" :title="`${r.code} · ${r.n} episode`">
            <div class="flex items-baseline justify-between gap-3 text-sm">
              <span class="min-w-0 truncate"><span class="font-mono text-xs text-muted">{{ r.code }}</span> {{ r.name ?? '—' }}</span>
              <span class="shrink-0 tabular-nums">{{ formatNumber(r.n) }} <span class="text-xs text-muted">({{ formatPercent(total ? r.n / total : 0) }})</span></span>
            </div>
            <div class="mt-1 h-2 overflow-hidden rounded-full bg-elevated">
              <div class="h-full rounded-full bg-(--sev-2)" :style="{ width: `${total ? Math.min(100, (r.n / total) * 100) : 0}%` }" />
            </div>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>
