<script setup lang="ts">
// Tab "Diagnosis": diagnosis primer & sekunder ICD-10 per severity beserta
// padanan SNOMED-CT.
const props = defineProps<{ code: string }>()
const api = usePathwayApi(() => props.code)
const { data, status } = await useAsyncData(pathwayKey(props.code, 'diagnoses'), () => api.diagnoses())

const severity = ref(0)
const rows = computed(() => (data.value ?? []).filter(d => !severity.value || d.severity === severity.value))

const SNOMED: Record<string, { label: string, color: 'success' | 'warning', icon: string }> = {
  OTOMATIS: { label: 'Otomatis', color: 'success', icon: 'i-lucide-check' },
  MANUAL: { label: 'Manual', color: 'success', icon: 'i-lucide-user-check' },
  RAGU: { label: 'Perlu ditinjau', color: 'warning', icon: 'i-lucide-circle-help' },
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <CommonSeverityFilter v-model="severity" />
      <UButton to="/padanan/snomed" label="Meja kerja SNOMED-CT" trailing-icon="i-lucide-arrow-right" size="sm" color="neutral" variant="ghost" />
    </div>

    <USkeleton v-if="status === 'pending' && !data" class="h-40 w-full" />
    <UEmpty v-else-if="!rows.length" icon="i-lucide-stethoscope" title="Tidak ada diagnosis" variant="outline" />
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th v-if="!severity" class="w-16">Sev.</th>
            <th class="w-24">ICD-10</th>
            <th>Diagnosis</th>
            <th class="w-28">Peran</th>
            <th class="num w-24">Episode</th>
            <th class="w-56">SNOMED-CT</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="d in rows" :key="`${d.severity}-${d.icd10_code}-${d.is_primary}`">
            <td v-if="!severity" class="text-muted">{{ severityLabel(d.severity) }}</td>
            <td class="font-mono">{{ d.icd10_code }}</td>
            <td>
              {{ d.name ?? '—' }}
              <span v-if="!d.name" class="block text-xs text-error">Kode tidak ada di master ICD-10</span>
            </td>
            <td>
              <UBadge :label="d.is_primary ? 'Primer' : 'Sekunder'" size="sm" :color="d.is_primary ? 'primary' : 'neutral'" variant="subtle" />
            </td>
            <td class="num">{{ formatNumber(d.episodes) }}</td>
            <td>
              <template v-if="d.snomed_status">
                <UBadge :label="SNOMED[d.snomed_status]!.label" :icon="SNOMED[d.snomed_status]!.icon" size="sm" :color="SNOMED[d.snomed_status]!.color" variant="subtle" />
                <span class="ml-2 font-mono text-xs text-muted">{{ d.snomed_concept_id }}</span>
              </template>
              <UBadge v-else label="Belum dipadankan" icon="i-lucide-circle-dashed" size="sm" color="neutral" variant="outline" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
