<script setup lang="ts">
// Penanda posisi CP di alur pengesahan: Draf → Tim CP → Komite → Direktur → Aktif.
// REVISI ditampilkan pada langkah pertama dengan tanda "Perlu Revisi".
import type { Stage } from '~/types/domain'

const props = defineProps<{ stage: Stage, since?: string }>()

const STEPS: { stage: Stage, label: string, who: string }[] = [
  { stage: 'DRAF', label: 'Penyusunan', who: 'Dokter · DPJP · Tim CP' },
  { stage: 'REVIEW_TIM_CP', label: 'Review Tim CP', who: 'Tim CP' },
  { stage: 'REVIEW_KOMITE', label: 'Review Komite', who: 'KSM/Komite Medik' },
  { stage: 'MENUNGGU_DIREKTUR', label: 'Pengesahan', who: 'Direktur' },
  { stage: 'AKTIF', label: 'Aktif', who: 'Dipakai di TransCPR-X' },
]

const current = computed(() => (props.stage === 'REVISI' ? 0 : STEPS.findIndex(s => s.stage === props.stage)))

function state(i: number) {
  if (i < current.value) return 'done'
  if (i === current.value) return props.stage === 'REVISI' ? 'revisi' : 'current'
  return 'todo'
}
</script>

<template>
  <ol class="grid grid-cols-5 gap-2" aria-label="Tahap pengesahan">
    <li
      v-for="(s, i) in STEPS"
      :key="s.stage"
      class="relative rounded-md border px-3 py-2"
      :class="{
        'border-default bg-default': state(i) === 'todo',
        'border-primary/25 bg-primary/5': state(i) === 'done',
        'bg-brand-soft border-primary ring-1 ring-primary': state(i) === 'current',
        'border-warning bg-warning/10 ring-1 ring-warning': state(i) === 'revisi',
      }"
      :aria-current="state(i) === 'current' || state(i) === 'revisi' ? 'step' : undefined"
    >
      <div class="flex items-center gap-1.5 text-xs font-medium">
        <UIcon
          :name="state(i) === 'done' ? 'i-lucide-circle-check' : state(i) === 'revisi' ? 'i-lucide-rotate-ccw' : state(i) === 'current' ? 'i-lucide-circle-dot' : 'i-lucide-circle'"
          class="size-4 shrink-0"
          :class="{
            'text-primary': state(i) === 'done' || state(i) === 'current',
            'text-warning': state(i) === 'revisi',
            'text-dimmed': state(i) === 'todo',
          }"
        />
        <span :class="state(i) === 'todo' ? 'text-muted' : 'text-highlighted'">
          {{ state(i) === 'revisi' ? 'Perlu Revisi' : s.label }}
        </span>
      </div>
      <p class="mt-0.5 truncate text-xs text-muted">{{ s.who }}</p>
      <p v-if="(state(i) === 'current' || state(i) === 'revisi') && since" class="mt-0.5 text-xs text-muted">
        sejak {{ formatDate(since) }}
      </p>
    </li>
  </ol>
</template>
