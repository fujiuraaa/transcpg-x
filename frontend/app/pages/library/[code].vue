<script setup lang="ts">
// §5 Halaman Detail CP — pusat kerja: acuan, rencana, kriteria, pengesahan.
definePageMeta({ title: 'CP Library' })

const route = useRoute()
const code = computed(() => String(route.params.code))
const api = usePathwayApi(code)
const { selected } = useHospitalProfile()

const { data: detail, status, error } = await useAsyncData(
  () => `${pathwayKey(code.value, 'detail')}:${selected.value?.id ?? 0}`,
  () => api.detail(),
)
const { data: readiness } = await useAsyncData(pathwayKey(code.value, 'readiness'), () => api.readiness())
const unmetCount = computed(() => readiness.value?.requirements.filter(r => r.blocking && !r.met).length ?? 0)

// 13 tab konten — urutan sesuai alur-aplikasi §5.
const tabs = computed(() => [
  { label: 'Alur Pelaksanaan', value: 'alur', icon: 'i-lucide-route' },
  { label: 'Acuan Klinis', value: 'acuan', icon: 'i-lucide-book-open' },
  { label: 'Ringkasan', value: 'ringkasan', icon: 'i-lucide-chart-bar' },
  { label: 'Diagnosis', value: 'diagnosis', icon: 'i-lucide-stethoscope' },
  { label: 'Prosedur & KPTL', value: 'prosedur', icon: 'i-lucide-scan-line' },
  { label: 'Tarif & Biaya', value: 'tarif', icon: 'i-lucide-wallet' },
  { label: 'Intervensi Harian', value: 'harian', icon: 'i-lucide-calendar-range' },
  { label: 'Aturan CDSS', value: 'cdss', icon: 'i-lucide-git-branch' },
  { label: 'Scoring Klinis', value: 'scoring', icon: 'i-lucide-calculator' },
  { label: 'Indikator Mutu', value: 'mutu', icon: 'i-lucide-target' },
  {
    label: 'Verifikasi Kode',
    value: 'verifikasi',
    icon: 'i-lucide-list-checks',
    badge: unmetCount.value ? { label: String(unmetCount.value), color: 'warning' as const, variant: 'solid' as const } : undefined,
  },
  { label: 'Rencana Klinis', value: 'rencana', icon: 'i-lucide-clipboard-list' },
  { label: 'Kriteria', value: 'kriteria', icon: 'i-lucide-filter' },
])
const tab = computed({
  get: () => (route.query.tab as string) ?? 'rencana',
  set: v => navigateTo({ query: { ...route.query, tab: v } }, { replace: true }),
})

// Baris tab bisa digeser: pastikan tab aktif selalu terlihat.
watch(tab, async () => {
  await nextTick()
  document.querySelector('[role="tab"][aria-selected="true"]')?.scrollIntoView({ block: 'nearest', inline: 'center', behavior: 'smooth' })
}, { flush: 'post' })
onMounted(() => document.querySelector('[role="tab"][aria-selected="true"]')?.scrollIntoView({ block: 'nearest', inline: 'center' }))
</script>

<template>
  <div v-if="status === 'pending' && !detail" class="space-y-4">
    <USkeleton class="h-24 w-full" />
    <USkeleton class="h-64 w-full" />
  </div>

  <UEmpty
    v-else-if="error"
    icon="i-lucide-file-x"
    :title="apiErrorMessage(error)"
    :actions="[{ label: 'Kembali ke CP Library', to: '/library', icon: 'i-lucide-arrow-left' }]"
  />

  <div v-else-if="detail" class="mx-auto max-w-[1600px] space-y-6">
    <PathwayPageHeader :detail="detail" />

    <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_22rem]">
      <!-- Panel kerja: di atas pada layar sempit, di kanan (menempel) pada layar lebar -->
      <aside class="grid gap-4 self-start md:grid-cols-2 xl:sticky xl:top-20 xl:order-last xl:grid-cols-1">
        <PathwayApprovalPanel :code="code" :detail="detail" @open-tab="tab = $event" />
        <PathwayReferencePanel :code="code" :detail="detail" />
      </aside>

      <div class="min-w-0 space-y-5">
        <PathwaySummaryCards :cards="detail.cards" :hospital-name="detail.hospital.name" />

        <UTabs
          v-model="tab"
          :items="tabs"
          :content="false"
          variant="link"
          :ui="{ list: 'overflow-x-auto border-b border-default', trigger: 'shrink-0', label: 'whitespace-nowrap overflow-visible' }"
        />

        <div>
          <PathwayTabsFlowTab v-if="tab === 'alur'" :code="code" :permissions="detail.permissions" />
          <PathwayTabsGuidelinesTab v-else-if="tab === 'acuan'" :code="code" :permissions="detail.permissions" />
          <PathwayTabsSummaryTab v-else-if="tab === 'ringkasan'" :code="code" />
          <PathwayTabsDiagnosesTab v-else-if="tab === 'diagnosis'" :code="code" />
          <PathwayTabsProceduresTab v-else-if="tab === 'prosedur'" :code="code" />
          <PathwayTabsTariffsTab v-else-if="tab === 'tarif'" :code="code" />
          <PathwayTabsDailyTab v-else-if="tab === 'harian'" :code="code" />
          <PathwayTabsCdssTab v-else-if="tab === 'cdss'" :code="code" />
          <PathwayTabsScoringTab v-else-if="tab === 'scoring'" :code="code" />
          <PathwayTabsQualityTab v-else-if="tab === 'mutu'" :code="code" />
          <PathwayTabsVerificationTab v-else-if="tab === 'verifikasi'" :code="code" @open-tab="tab = $event" />
          <PathwayTabsPlanTab v-else-if="tab === 'rencana'" :code="code" :permissions="detail.permissions" />
          <PathwayTabsCriteriaTab v-else-if="tab === 'kriteria'" :code="code" :permissions="detail.permissions" />
        </div>
      </div>
    </div>
  </div>
</template>
