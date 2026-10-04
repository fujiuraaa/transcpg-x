<script setup lang="ts">
// Tab "Rencana per Hari": lini waktu H0, H1, … untuk satu severity —
// yang nantinya dijalankan TransCPR-X sebagai checklist harian pasien.
import type { PlanItem } from '~/types/api'

const props = defineProps<{ plan: PlanItem[], severity: number }>()

const days = computed(() => {
  const map = new Map<number, PlanItem[]>()
  for (const it of props.plan) {
    if (!map.has(it.day)) map.set(it.day, [])
    map.get(it.day)!.push(it)
  }
  return [...map.entries()].sort((a, b) => a[0] - b[0])
})

const usage = (it: PlanItem) => [it.dose, it.route, it.frequency, it.duration].filter(Boolean).join(' · ')
</script>

<template>
  <UEmpty
    v-if="!days.length"
    icon="i-lucide-calendar-x"
    :title="`Tidak ada rencana untuk severity ${severityLabel(severity)}`"
    description="Severity ini tidak wajib punya rencana bila datanya tipis (< 10 episode)."
    variant="outline"
  />
  <ol v-else class="relative space-y-6 before:absolute before:top-2 before:bottom-2 before:left-[1.4rem] before:w-px before:bg-(--ui-border-accented)">
    <li v-for="[day, items] in days" :key="day" class="relative grid grid-cols-[3rem_minmax(0,1fr)] gap-4">
      <div class="bg-brand-gradient z-10 grid size-11 place-items-center rounded-full font-mono text-sm font-semibold text-white shadow-sm ring-4 ring-(--ui-bg)">
        {{ dayLabel(day) }}
      </div>
      <div class="rounded-lg border border-default bg-default">
        <p class="border-b border-default px-4 py-2 text-xs font-medium text-muted">
          Hari rawat ke-{{ day }} · {{ items.length }} item
        </p>
        <ul class="divide-y divide-default">
          <li v-for="it in items" :key="it.id" class="grid gap-x-4 gap-y-1 px-4 py-3 sm:grid-cols-[minmax(0,1fr)_auto]">
            <div class="flex min-w-0 gap-2.5">
              <UIcon :name="PLAN_TYPE[it.item_type].icon" class="mt-0.5 size-4 shrink-0 text-muted" />
              <div class="min-w-0">
                <p class="font-medium">{{ it.item_name }}</p>
                <p class="text-xs text-muted">
                  {{ PLAN_TYPE[it.item_type].label }} · <span class="font-mono">{{ it.item_code }}</span>
                </p>
                <p v-if="usage(it)" class="mt-1 text-sm">{{ usage(it) }}</p>
                <p v-if="it.note" class="mt-0.5 text-xs italic text-muted">{{ it.note }}</p>
              </div>
            </div>
            <div class="flex flex-wrap items-start gap-1.5 sm:flex-col sm:items-end">
              <UBadge
                :label="PLAN_NATURE[it.nature]"
                size="sm"
                :color="it.nature === 'WAJIB' ? 'primary' : 'neutral'"
                :variant="it.nature === 'WAJIB' ? 'subtle' : 'outline'"
              />
              <UBadge v-if="it.outside_guidance" label="Di luar panduan" size="sm" color="warning" variant="subtle" icon="i-lucide-info" />
              <span v-if="it.source_note" class="text-xs text-muted sm:text-right">{{ it.source_note }}</span>
            </div>
          </li>
        </ul>
      </div>
    </li>
  </ol>
</template>
