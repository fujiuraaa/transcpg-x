<script setup lang="ts">
// Tab "Prosedur & KPTL": prosedur ICD-9-CM per severity + status padanan
// KPTL (kode billing BPJS) yang ditetapkan Tim CP.
const props = defineProps<{ code: string }>()
const api = usePathwayApi(() => props.code)
const { data, status } = await useAsyncData(pathwayKey(props.code, 'procedures'), () => api.procedures())

const severity = ref(0)
const rows = computed(() => (data.value ?? []).filter(p => !severity.value || p.severity === severity.value))

const unique = computed(() => {
  const m = new Map<string, boolean>()
  for (const p of data.value ?? []) m.set(p.icd9cm_code, !!p.kptl_code)
  return { total: m.size, mapped: [...m.values()].filter(Boolean).length }
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <CommonSeverityFilter v-model="severity" />
      <div class="flex items-center gap-3">
        <span class="text-sm text-muted">
          <span class="font-medium tabular-nums text-default">{{ unique.mapped }}/{{ unique.total }}</span> prosedur sudah punya padanan KPTL
        </span>
        <UButton to="/padanan/kptl" label="Meja kerja KPTL" trailing-icon="i-lucide-arrow-right" size="sm" color="neutral" variant="ghost" />
      </div>
    </div>

    <USkeleton v-if="status === 'pending' && !data" class="h-40 w-full" />
    <UEmpty v-else-if="!rows.length" icon="i-lucide-scan-line" title="Tidak ada prosedur tercatat" variant="outline" />
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th v-if="!severity" class="w-16">Sev.</th>
            <th class="w-24">ICD-9-CM</th>
            <th>Prosedur</th>
            <th class="num w-24">Episode</th>
            <th class="w-72">Padanan KPTL</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in rows" :key="`${p.severity}-${p.icd9cm_code}`">
            <td v-if="!severity" class="text-muted">{{ severityLabel(p.severity) }}</td>
            <td class="font-mono">{{ p.icd9cm_code }}</td>
            <td>{{ p.name ?? '—' }}</td>
            <td class="num">{{ formatNumber(p.episodes) }}</td>
            <td>
              <span v-if="p.kptl_code" class="flex items-start gap-1.5">
                <UIcon name="i-lucide-check" class="mt-0.5 size-4 shrink-0 text-success" />
                <span><span class="font-mono text-xs">{{ p.kptl_code }}</span> · {{ p.kptl_name }}</span>
              </span>
              <UBadge v-else label="Belum dipadankan" icon="i-lucide-circle-dashed" size="sm" color="warning" variant="subtle" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
