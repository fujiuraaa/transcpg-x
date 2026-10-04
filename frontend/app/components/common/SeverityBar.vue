<script setup lang="ts">
// Batang bertumpuk distribusi episode per severity. Severity bersifat urutan
// → satu hue (merah) dari terang ke gelap; langkah mode gelap dipilih
// tersendiri (lihat main.css --sev-*). Identitas tidak hanya dari warna:
// celah 2px antar-segmen, legenda berlabel, dan tabel angka selalu ada.
const props = defineProps<{ counts: Record<number, number> }>()

const STEPS: Record<number, string> = { 1: 'var(--sev-1)', 2: 'var(--sev-2)', 3: 'var(--sev-3)' }

const total = computed(() => [1, 2, 3].reduce((s, k) => s + (props.counts[k] ?? 0), 0))
const segments = computed(() => [1, 2, 3]
  .map(s => ({ severity: s, n: props.counts[s] ?? 0, share: total.value ? (props.counts[s] ?? 0) / total.value : 0 }))
  .filter(s => s.n > 0))
</script>

<template>
  <figure class="space-y-2">
    <div class="flex h-4 w-full gap-[2px] overflow-hidden rounded" role="img" :aria-label="segments.map(s => `Severity ${severityLabel(s.severity)} ${formatPercent(s.share)}`).join(', ')">
      <div
        v-for="s in segments"
        :key="s.severity"
        class="h-full first:rounded-l last:rounded-r"
        :style="{ width: `${s.share * 100}%`, backgroundColor: STEPS[s.severity] }"
        :title="`Severity ${severityLabel(s.severity)}: ${formatNumber(s.n)} episode (${formatPercent(s.share)})`"
      />
    </div>
    <figcaption class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted">
      <span v-for="s in [1, 2, 3]" :key="s" class="flex items-center gap-1.5">
        <span class="size-2.5 rounded-sm" :style="{ backgroundColor: STEPS[s] }" aria-hidden="true" />
        Severity {{ severityLabel(s) }} · <span class="tabular-nums text-default">{{ formatNumber(counts[s] ?? 0) }}</span>
        <span v-if="total">({{ formatPercent((counts[s] ?? 0) / total) }})</span>
      </span>
    </figcaption>
  </figure>
</template>
