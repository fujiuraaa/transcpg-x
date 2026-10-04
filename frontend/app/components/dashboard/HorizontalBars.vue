<script setup lang="ts">
// Grafik batang horizontal satu seri (magnitudo): satu warna, urut menurun (atau urutan asli bila keepOrder, mis. tahap),
// batang tipis berujung bulat di sisi data dan siku di garis dasar, nilai
// di ujung batang, tooltip saat disorot. Satu seri → tanpa kotak legenda.
export interface BarItem {
  key: string
  label: string
  sublabel?: string
  value: number
  display: string
  tooltip?: string
}

const props = withDefaults(defineProps<{ items: BarItem[], max?: number, label: string, keepOrder?: boolean }>(), { max: undefined, keepOrder: false })

const top = computed(() => props.max ?? Math.max(1, ...props.items.map(i => i.value)))
const sorted = computed(() => (props.keepOrder ? props.items : [...props.items].sort((a, b) => b.value - a.value)))
</script>

<template>
  <ul class="space-y-2.5" role="list" :aria-label="label">
    <li v-for="it in sorted" :key="it.key" class="grid grid-cols-[minmax(0,11rem)_minmax(0,1fr)] items-center gap-3 sm:grid-cols-[minmax(0,14rem)_minmax(0,1fr)]">
      <div class="min-w-0 text-right">
        <p class="truncate text-sm" :title="it.label">{{ it.label }}</p>
        <p v-if="it.sublabel" class="truncate text-xs text-muted">{{ it.sublabel }}</p>
      </div>
      <UTooltip :text="it.tooltip ?? `${it.label}: ${it.display}`" :content="{ side: 'top' }">
        <div class="group flex h-6 items-center gap-2" tabindex="0" :aria-label="`${it.label}: ${it.display}`">
          <div
            class="h-3.5 rounded-r-[4px] bg-(--sev-2) transition-opacity group-hover:opacity-80 group-focus-visible:ring-2 group-focus-visible:ring-primary"
            :style="{ width: `max(2px, ${(it.value / top) * 85}%)` }"
          />
          <span class="shrink-0 text-xs font-medium tabular-nums text-toned">{{ it.display }}</span>
        </div>
      </UTooltip>
    </li>
  </ul>
</template>
