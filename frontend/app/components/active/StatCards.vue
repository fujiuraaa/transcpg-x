<script setup lang="ts">
// 4 kartu ringkasan halaman CP Aktif.
import type { ActiveSummary } from '~/types/api'

const NuxtLink = resolveComponent('NuxtLink')
const props = defineProps<{ summary: ActiveSummary | null }>()

const cards = computed(() => [
  { label: 'CP Aktif', value: props.summary?.cp_aktif, hint: 'sudah disahkan Direktur', icon: 'i-lucide-badge-check', to: null },
  { label: 'Episode tercakup', value: props.summary?.episode_tercakup, hint: 'episode rawat inap dari CP aktif', icon: 'i-lucide-bed-double', to: null },
  { label: 'Kelompok diagnosis', value: props.summary?.kelompok_diagnosis, hint: 'MDC yang sudah punya CP aktif', icon: 'i-lucide-folder-tree', to: null },
  { label: 'Menunggu Direktur', value: props.summary?.menunggu_direktur, hint: 'segera menjadi CP aktif', icon: 'i-lucide-stamp', to: '/approval?tab=proses&tahap=MENUNGGU_DIREKTUR' },
])
</script>

<template>
  <div class="grid grid-cols-2 gap-3 xl:grid-cols-4">
    <component
      :is="c.to ? NuxtLink : 'div'"
      v-for="c in cards"
      :key="c.label"
      :to="c.to ?? undefined"
      class="rounded-lg border border-default bg-default p-4"
      :class="c.to ? 'transition hover:border-primary/40 hover:shadow-sm' : ''"
    >
      <div class="flex items-center justify-between text-xs text-muted">
        <span>{{ c.label }}</span>
        <UIcon :name="c.icon" class="size-4" />
      </div>
      <p class="mt-1 text-2xl font-semibold tabular-nums text-highlighted">
        <template v-if="c.value != null">{{ formatNumber(c.value) }}</template>
        <USkeleton v-else class="h-7 w-10" />
      </p>
      <p class="truncate text-xs text-muted">
        {{ c.hint }}<UIcon v-if="c.to" name="i-lucide-arrow-right" class="ml-1 inline size-3" />
      </p>
    </component>
  </div>
</template>
