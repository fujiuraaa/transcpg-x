<script setup lang="ts">
// Cakupan padanan SNOMED-CT per vokabuler: batang bertumpuk dipadankan /
// perlu ditinjau / belum, dengan legenda berlabel dan angka.
import type { SnomedVocabSummary } from '~/types/api'

defineProps<{ rows: SnomedVocabSummary[] }>()

const VOCAB: Record<string, string> = { ICD10: 'ICD-10 · diagnosis', ICD9CM: 'ICD-9-CM · prosedur' }
const SEG = [
  { key: 'dipadankan', label: 'Dipadankan', cls: 'bg-(--ui-color-success-500)' },
  { key: 'perlu_ditinjau', label: 'Perlu ditinjau', cls: 'bg-(--ui-color-warning-400)' },
  { key: 'belum', label: 'Belum', cls: 'bg-(--ui-color-neutral-300) dark:bg-(--ui-color-neutral-600)' },
] as const
</script>

<template>
  <div class="grid gap-3 md:grid-cols-2">
    <div v-for="v in rows" :key="v.vocabulary" class="rounded-lg border border-default bg-default p-4">
      <div class="flex items-baseline justify-between">
        <p class="text-sm font-medium">{{ VOCAB[v.vocabulary] ?? v.vocabulary }}</p>
        <p class="text-xs text-muted"><span class="font-semibold tabular-nums text-default">{{ formatPercent(v.kode ? v.dipadankan / v.kode : 0) }}</span> dari {{ v.kode }} kode</p>
      </div>
      <div class="mt-2 flex h-3 gap-[2px] overflow-hidden rounded-full" role="img" :aria-label="`${v.dipadankan} dipadankan, ${v.perlu_ditinjau} perlu ditinjau, ${v.belum} belum`">
        <div
          v-for="s in SEG"
          v-show="v[s.key] > 0"
          :key="s.key"
          class="h-full first:rounded-l-full last:rounded-r-full"
          :class="s.cls"
          :style="{ width: `${(v[s.key] / Math.max(1, v.kode)) * 100}%` }"
          :title="`${s.label}: ${v[s.key]}`"
        />
      </div>
      <p class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted">
        <span v-for="s in SEG" :key="s.key" class="flex items-center gap-1.5">
          <span class="size-2.5 rounded-sm" :class="s.cls" aria-hidden="true" />
          {{ s.label }} <span class="tabular-nums text-default">{{ v[s.key] }}</span>
        </span>
      </p>
    </div>
  </div>
</template>
