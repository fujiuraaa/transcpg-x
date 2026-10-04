<script setup lang="ts">
// Tab "Indikator Mutu": target terukur yang dihitung dari baseline data klaim.
const props = defineProps<{ code: string }>()
const api = usePathwayApi(() => props.code)
const { data, status } = await useAsyncData(pathwayKey(props.code, 'quality'), () => api.quality())
</script>

<template>
  <USkeleton v-if="status === 'pending' && !data" class="h-32 w-full" />
  <UEmpty
    v-else-if="!data?.length"
    icon="i-lucide-target"
    title="Belum ada indikator mutu"
    description="Indikator (LOS, kepatuhan tindakan inti, dokumentasi variance) dihitung dari baseline klaim oleh analisis data."
    variant="outline"
  />
  <div v-else class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th class="w-32">Kode</th>
          <th>Indikator</th>
          <th class="num">Baseline</th>
          <th class="num">Target</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="q in data" :key="q.id">
          <td class="font-mono text-xs">{{ q.code }}</td>
          <td>{{ q.name }}</td>
          <td class="num">{{ formatNumber(q.baseline, 1) }} {{ q.unit }}</td>
          <td class="num font-medium">{{ formatNumber(q.target, 1) }} {{ q.unit }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
