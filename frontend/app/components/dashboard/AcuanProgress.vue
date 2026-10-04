<script setup lang="ts">
// Kemajuan penetapan acuan: bagian-dari-keseluruhan CP layak disusun, plus
// 3 kartu. "Dilewati" (< 5 episode) tidak bisa diklik — belum layak jadi CP.
import type { DashboardSummary } from '~/types/api'

const NuxtLink = resolveComponent('NuxtLink')
const props = defineProps<{ progress: DashboardSummary['progress'] }>()

const total = computed(() => props.progress.acuan_ditetapkan + props.progress.menunggu_penetapan)
const done = computed(() => (total.value ? props.progress.acuan_ditetapkan / total.value : 0))

const cards = computed(() => [
  { label: 'Acuan sudah ditetapkan', value: props.progress.acuan_ditetapkan, icon: 'i-lucide-file-check', iconClass: 'text-success', to: '/library?acuan=sudah', note: 'Lihat di CP Library' },
  { label: 'Menunggu penetapan', value: props.progress.menunggu_penetapan, icon: 'i-lucide-file-clock', iconClass: 'text-warning', to: '/library?acuan=belum', note: 'Tetapkan acuan' },
  { label: 'Dilewati', value: props.progress.dilewati, icon: 'i-lucide-file-x', iconClass: 'text-dimmed', to: null, note: '< 5 episode, belum layak disusun' },
])
</script>

<template>
  <div class="space-y-4">
    <div>
      <div class="mb-1.5 flex items-baseline justify-between text-sm">
        <span class="text-muted">CP dengan acuan standar</span>
        <span class="font-semibold tabular-nums">{{ formatPercent(done) }}</span>
      </div>
      <div
        class="flex h-3 gap-[2px] overflow-hidden rounded-full bg-elevated"
        role="img"
        :aria-label="`${progress.acuan_ditetapkan} dari ${total} CP sudah punya acuan`"
      >
        <div v-if="progress.acuan_ditetapkan" class="h-full rounded-l-full bg-(--ui-color-success-500)" :style="{ width: `${done * 100}%` }" />
        <div v-if="progress.menunggu_penetapan" class="h-full flex-1 rounded-r-full bg-(--ui-color-warning-400)" />
      </div>
      <p class="mt-1.5 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted">
        <span class="flex items-center gap-1.5"><span class="size-2.5 rounded-sm bg-(--ui-color-success-500)" aria-hidden="true" /> Sudah ditetapkan</span>
        <span class="flex items-center gap-1.5"><span class="size-2.5 rounded-sm bg-(--ui-color-warning-400)" aria-hidden="true" /> Menunggu penetapan</span>
      </p>
    </div>

    <div class="grid gap-2">
      <component
        :is="c.to ? NuxtLink : 'div'"
        v-for="c in cards"
        :key="c.label"
        :to="c.to ?? undefined"
        class="flex items-center gap-3 rounded-lg border border-default p-3"
        :class="c.to ? 'transition hover:border-primary/40 hover:bg-elevated/50' : 'opacity-75'"
        :aria-disabled="c.to ? undefined : 'true'"
      >
        <UIcon :name="c.icon" class="size-5 shrink-0" :class="c.iconClass" />
        <div class="min-w-0 flex-1">
          <p class="text-sm font-medium">{{ c.label }}</p>
          <p class="text-xs text-muted">{{ c.note }}</p>
        </div>
        <span class="text-xl font-semibold tabular-nums">{{ formatNumber(c.value) }}</span>
        <UIcon v-if="c.to" name="i-lucide-chevron-right" class="size-4 text-muted" />
      </component>
    </div>
  </div>
</template>
