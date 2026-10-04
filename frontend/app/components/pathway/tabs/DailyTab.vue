<script setup lang="ts">
// Tab "Intervensi Harian": matriks hari rawat × jenis, menampilkan data
// rencana isi klinis untuk satu severity. Matriks dari pola klaim menyusul.
import type { PlanItem } from '~/types/api'
import type { PlanItemType } from '~/types/domain'

const props = defineProps<{ code: string }>()
const api = usePathwayApi(() => props.code)
const { data, status } = await useAsyncData(pathwayKey(props.code, 'plan'), () => api.plan())

const severity = ref(1)
const items = computed(() => (data.value ?? []).filter(i => i.severity === severity.value))
const days = computed(() => [...new Set(items.value.map(i => i.day))].sort((a, b) => a - b))
const TYPES = Object.keys(PLAN_TYPE) as PlanItemType[]
const cell = (type: PlanItemType, day: number): PlanItem[] => items.value.filter(i => i.item_type === type && i.day === day)
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <CommonSeverityFilter v-model="severity" :allow-all="false" />
      <span class="text-xs text-muted">Sumber: rencana isi klinis · <UIcon name="i-lucide-circle-dot" class="inline size-3 text-primary" /> wajib · <UIcon name="i-lucide-circle" class="inline size-3" /> kondisional</span>
    </div>

    <USkeleton v-if="status === 'pending' && !data" class="h-40 w-full" />
    <UEmpty
      v-else-if="!days.length"
      icon="i-lucide-calendar-range"
      :title="`Belum ada rencana untuk severity ${severityLabel(severity)}`"
      description="Susun rencana di tab Rencana Klinis; matriks harian terbentuk otomatis."
      variant="outline"
    />
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th class="w-36">Jenis</th>
            <th v-for="d in days" :key="d" class="min-w-40">{{ dayLabel(d) }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in TYPES" :key="t">
            <td class="font-medium">
              <span class="flex items-center gap-1.5"><UIcon :name="PLAN_TYPE[t].icon" class="size-4 text-muted" /> {{ PLAN_TYPE[t].label }}</span>
            </td>
            <td v-for="d in days" :key="d">
              <ul class="space-y-1">
                <li v-for="it in cell(t, d)" :key="it.id" class="flex items-start gap-1.5 text-xs">
                  <UIcon
                    :name="it.nature === 'WAJIB' ? 'i-lucide-circle-dot' : 'i-lucide-circle'"
                    class="mt-0.5 size-3 shrink-0"
                    :class="it.nature === 'WAJIB' ? 'text-primary' : 'text-muted'"
                    :title="PLAN_NATURE[it.nature]"
                  />
                  <span>
                    {{ it.item_name }}
                    <span v-if="it.dose" class="text-muted">· {{ it.dose }} {{ it.frequency }}</span>
                  </span>
                </li>
              </ul>
              <span v-if="!cell(t, d).length" class="text-dimmed">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
