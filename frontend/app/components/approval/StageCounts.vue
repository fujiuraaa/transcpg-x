<script setup lang="ts">
// 5 kartu jumlah CP per tahap. Klik kartu → saring antrean ke tahap itu.
import type { Stage } from '~/types/domain'

const props = defineProps<{ counts: Record<Stage, number> | null, active: Stage | null }>()
const emit = defineEmits<{ select: [stage: Stage | null] }>()

const CARDS: { stage: Stage, label: string, who: string, icon: string }[] = [
  { stage: 'REVIEW_TIM_CP', label: 'Review Tim CP', who: 'menunggu Tim CP', icon: 'i-lucide-users' },
  { stage: 'REVIEW_KOMITE', label: 'Review Komite', who: 'menunggu KSM/Komite Medik', icon: 'i-lucide-gavel' },
  { stage: 'MENUNGGU_DIREKTUR', label: 'Menunggu Direktur', who: 'menunggu pengesahan', icon: 'i-lucide-stamp' },
  { stage: 'REVISI', label: 'Perlu Revisi', who: 'dikembalikan ke penyusun', icon: 'i-lucide-rotate-ccw' },
  { stage: 'AKTIF', label: 'Aktif', who: 'sudah disahkan', icon: 'i-lucide-badge-check' },
]

// Warna aksen kartu mengikuti badge tahap (makin pekat merah mendekati pengesahan).
const ACCENT: Record<Stage, string> = {
  DRAF: 'before:bg-(--ui-border-accented)',
  REVIEW_TIM_CP: 'before:bg-(--ui-color-secondary-500)',
  REVIEW_KOMITE: 'before:bg-(--ui-color-primary-400)',
  MENUNGGU_DIREKTUR: 'before:bg-(--ui-color-primary-700)',
  REVISI: 'before:bg-(--ui-color-warning-500)',
  AKTIF: 'before:bg-(--ui-color-success-500)',
}
</script>

<template>
  <div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-5" role="group" aria-label="Jumlah CP per tahap">
    <button
      v-for="c in CARDS"
      :key="c.stage"
      type="button"
      class="relative overflow-hidden rounded-lg border bg-default p-4 pl-5 text-left transition before:absolute before:inset-y-0 before:left-0 before:w-1 hover:shadow-sm"
      :class="[ACCENT[c.stage], props.active === c.stage ? 'border-primary ring-1 ring-primary' : 'border-default']"
      :aria-pressed="props.active === c.stage"
      @click="emit('select', props.active === c.stage ? null : c.stage)"
    >
      <div class="flex items-center justify-between text-xs text-muted">
        <span>{{ c.label }}</span>
        <UIcon :name="c.icon" class="size-4" />
      </div>
      <p class="mt-1 text-2xl font-semibold tabular-nums text-highlighted">
        <template v-if="counts">{{ counts[c.stage] ?? 0 }}</template>
        <USkeleton v-else class="h-7 w-8" />
      </p>
      <p class="truncate text-xs text-muted">{{ c.who }}</p>
    </button>
  </div>
</template>
