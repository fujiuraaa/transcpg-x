<script setup lang="ts">
// Kotak distribusi severity I/II/III — persen pasien di setiap tingkat.
// Severity berdata tipis (< 10 episode) ditandai, karena target LOS-nya
// belum bisa ditetapkan.
const props = defineProps<{ counts: Record<string, number>, total: number }>()
const MIN_EPISODES = 10

const boxes = computed(() => [1, 2, 3].map((s) => {
  const n = props.counts[String(s)] ?? 0
  return { s, n, share: props.total ? n / props.total : 0, thin: n < MIN_EPISODES }
}))
</script>

<template>
  <div class="grid grid-cols-3 gap-1" role="group" aria-label="Distribusi severity">
    <div
      v-for="b in boxes"
      :key="b.s"
      class="rounded-md border border-default px-2 py-1.5 text-center"
      :title="`Severity ${severityLabel(b.s)}: ${b.n} episode${b.thin ? ' — data tipis' : ''}`"
    >
      <div class="mx-auto mb-1 h-1 w-6 rounded-full" :style="{ backgroundColor: `var(--sev-${b.s})` }" aria-hidden="true" />
      <p class="text-[10px] font-medium uppercase text-muted">Sev {{ severityLabel(b.s) }}</p>
      <p class="text-sm font-semibold tabular-nums" :class="b.thin ? 'text-muted' : 'text-highlighted'">{{ formatPercent(b.share) }}</p>
      <p class="text-[10px] tabular-nums" :class="b.thin ? 'text-warning' : 'text-muted'">{{ b.n }} ep</p>
    </div>
  </div>
</template>
