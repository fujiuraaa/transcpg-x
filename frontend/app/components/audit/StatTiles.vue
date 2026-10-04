<script setup lang="ts">
// Deretan kartu angka ringkas untuk laporan. `attention` = perlu perhatian
// (ikon + warna, tidak hanya warna).
export interface StatTile {
  label: string
  value: number | null
  hint: string
  icon: string
  attention?: boolean
}

defineProps<{ tiles: StatTile[] }>()
</script>

<template>
  <div class="grid grid-cols-2 gap-3 xl:grid-cols-4">
    <div
      v-for="t in tiles"
      :key="t.label"
      class="rounded-lg border p-4"
      :class="t.attention ? 'border-warning/50 bg-warning/5' : 'border-default bg-default'"
    >
      <div class="flex items-center justify-between text-xs" :class="t.attention ? 'text-warning' : 'text-muted'">
        <span>{{ t.label }}</span>
        <UIcon :name="t.icon" class="size-4" />
      </div>
      <p class="mt-1 text-2xl font-semibold tabular-nums text-highlighted">
        <template v-if="t.value != null">{{ formatNumber(t.value) }}</template>
        <USkeleton v-else class="h-7 w-12" />
      </p>
      <p class="truncate text-xs text-muted">{{ t.hint }}</p>
    </div>
  </div>
</template>
