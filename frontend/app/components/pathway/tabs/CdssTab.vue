<script setup lang="ts">
// Tab "Aturan CDSS": usulan aturan yang diturunkan otomatis dari pola klaim.
// Bukan aturan yang sudah ditinjau manusia; evaluasi real-time di TransCDSS-X.
const props = defineProps<{ code: string }>()
const api = usePathwayApi(() => props.code)
const { data, status } = await useAsyncData(pathwayKey(props.code, 'cdss'), () => api.cdssRules())

const RULE_TYPE: Record<string, string> = {
  ESCALATION: 'Eskalasi', STEP_ACTION: 'Tindakan wajib', DOCUMENTATION: 'Dokumentasi',
  PATHWAY_BRANCH: 'Percabangan pathway', CLINICAL_ALERT: 'Peringatan klinis', DRUG_SAFETY: 'Keamanan obat',
  LAB_ALERT: 'Nilai lab kritis', ALERT: 'Peringatan umum',
}
</script>

<template>
  <div class="space-y-4">
    <UAlert
      color="neutral"
      variant="subtle"
      icon="i-lucide-sparkles"
      title="Usulan, belum ditinjau manusia"
      description="Aturan diturunkan otomatis dari pola statistik klaim. Evaluasi keputusan klinis real-time dilakukan di TransCDSS-X."
    />
    <USkeleton v-if="status === 'pending' && !data" class="h-32 w-full" />
    <UEmpty
      v-else-if="!data?.rules.length"
      icon="i-lucide-git-branch"
      title="Belum ada usulan aturan"
      description="Usulan muncul setelah analisis pola klaim untuk grouper ini dijalankan."
      variant="outline"
    />
    <ul v-else class="space-y-2">
      <li v-for="r in data.rules" :key="r.id" class="rounded-lg border border-default p-4">
        <UBadge :label="RULE_TYPE[r.rule_type] ?? r.rule_type" size="sm" color="neutral" variant="subtle" />
        <dl class="mt-2 grid gap-2 text-sm sm:grid-cols-2">
          <div><dt class="text-xs text-muted">Bila</dt><dd class="font-mono text-xs">{{ JSON.stringify(r.condition) }}</dd></div>
          <div><dt class="text-xs text-muted">Maka</dt><dd class="font-mono text-xs">{{ JSON.stringify(r.action) }}</dd></div>
        </dl>
        <p v-if="r.rationale" class="mt-2 text-xs text-muted">{{ r.rationale }}</p>
      </li>
    </ul>
  </div>
</template>
