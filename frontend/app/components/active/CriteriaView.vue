<script setup lang="ts">
// Tab "Kriteria" (baca saja): inklusi & eksklusi pasien.
import type { Criterion } from '~/types/api'

const props = defineProps<{ criteria: Criterion[] }>()
const COLUMNS = [
  { kind: 'INKLUSI', title: 'Inklusi', hint: 'Pasien yang cocok untuk CP ini', icon: 'i-lucide-circle-check', color: 'text-success' },
  { kind: 'EKSKLUSI', title: 'Eksklusi', hint: 'Pasien yang tidak boleh masuk CP ini', icon: 'i-lucide-circle-minus', color: 'text-error' },
] as const
const list = (k: string) => props.criteria.filter(c => c.kind === k)
</script>

<template>
  <div class="grid gap-4 md:grid-cols-2">
    <section v-for="col in COLUMNS" :key="col.kind" class="rounded-lg border border-default bg-default">
      <header class="flex items-center gap-2 border-b border-default px-4 py-3">
        <UIcon :name="col.icon" class="size-5" :class="col.color" />
        <div>
          <h4 class="font-medium">{{ col.title }}</h4>
          <p class="text-xs text-muted">{{ col.hint }}</p>
        </div>
      </header>
      <p v-if="!list(col.kind).length" class="px-4 py-6 text-center text-sm text-muted">Tidak ada kriteria {{ col.title.toLowerCase() }}.</p>
      <ul v-else class="divide-y divide-default">
        <li v-for="c in list(col.kind)" :key="c.id" class="flex gap-2 px-4 py-2.5 text-sm">
          <span class="mt-2 size-1.5 shrink-0 rounded-full bg-current" :class="col.color" aria-hidden="true" />
          {{ c.description }}
        </li>
      </ul>
    </section>
  </div>
</template>
