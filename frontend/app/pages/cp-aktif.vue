<script setup lang="ts">
// §11 CP Aktif — pratinjau baca-saja isi CP yang sudah disahkan Direktur.
// Penerapan ke pasien (checklist harian, variance) dilakukan di TransCPR-X.
definePageMeta({ title: 'CP Aktif' })

const route = useRoute()
const router = useRouter()
const { list, detail } = useActiveApi()

const { data: listData, status: listStatus, error: listError } = await useAsyncData('active:list', list)

// CP terpilih & severity disimpan di URL (bisa dibuka dari "Pratinjau CP Aktif").
const code = computed(() => (route.query.code as string) ?? listData.value?.items[0]?.cbg_code ?? null)
const severity = computed(() => Number(route.query.severity ?? 1))
function go(q: { code?: string, severity?: number, tab?: string }) {
  router.replace({ query: { ...route.query, ...q } })
}

const { data, status, error } = await useAsyncData(
  () => `active:${code.value}:${severity.value}`,
  () => (code.value ? detail(code.value, severity.value) : Promise.resolve(null)),
)
const notActive = computed(() => (error.value as { data?: { code?: string } } | undefined)?.data?.code === 'BELUM_AKTIF')

// Siapa & kapan mengesahkan (dari riwayat pengesahan).
const activation = computed(() => data.value?.references_approval.approvals.find(a => a.to_stage === 'AKTIF') ?? null)

const TABS = [
  { label: 'Rencana per Hari', value: 'rencana', icon: 'i-lucide-calendar-range' },
  { label: 'Pola Klaim', value: 'pola', icon: 'i-lucide-chart-bar' },
  { label: 'Acuan & Pengesahan', value: 'acuan', icon: 'i-lucide-book-open' },
  { label: 'Kriteria', value: 'kriteria', icon: 'i-lucide-filter' },
]
const tab = computed(() => (route.query.tab as string) ?? 'rencana')
</script>

<template>
  <div class="mx-auto max-w-[1500px] space-y-6">
    <header class="bg-brand-soft flex flex-wrap items-end justify-between gap-3 rounded-xl border border-primary/15 px-5 py-4">
      <div>
        <h2 class="text-2xl font-semibold text-highlighted">CP Aktif</h2>
        <p class="text-sm text-muted">Isi CP yang sudah disahkan Direktur — rencana per hari, pola klaim, acuan, dan kriteria pasien.</p>
      </div>
      <p class="flex items-center gap-1.5 text-xs text-muted">
        <UIcon name="i-lucide-eye" class="size-3.5" />
        Pratinjau baca-saja. Penerapan ke pasien dilakukan di TransCPR-X.
      </p>
    </header>

    <ActiveStatCards :summary="listData?.summary ?? null" />

    <USkeleton v-if="listStatus === 'pending' && !listData" class="h-64 w-full" />
    <UAlert v-else-if="listError" color="error" variant="subtle" :title="apiErrorMessage(listError)" />

    <!-- Kemungkinan 1: belum ada CP Aktif -->
    <UEmpty
      v-else-if="!listData?.items.length && !route.query.code"
      icon="i-lucide-badge-check"
      title="Belum ada CP Aktif"
      description="CP menjadi Aktif setelah melewati Review Tim CP, Review Komite, dan disahkan Direktur."
      :actions="[{ label: 'Lihat antrean pengesahan', to: '/approval', icon: 'i-lucide-stamp' }]"
      variant="outline"
    />

    <div v-else class="grid gap-6 lg:grid-cols-[18rem_minmax(0,1fr)]">
      <ActivePathwayList :items="listData?.items ?? []" :selected="code" @select="(c) => go({ code: c })" />

      <div class="min-w-0 space-y-5">
        <!-- Kemungkinan 2: CP yang diminta belum Aktif -->
        <UAlert
          v-if="notActive"
          color="warning"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          :title="apiErrorMessage(error)"
          description="Pratinjau hanya tersedia untuk CP yang sudah disahkan Direktur."
          :actions="[{ label: 'Buka halaman detail CP', to: `/library/${code}`, color: 'warning', variant: 'outline' }]"
        />
        <UAlert v-else-if="error" color="error" variant="subtle" :title="apiErrorMessage(error)" />

        <!-- Kemungkinan 3: tampilkan CP -->
        <template v-else-if="data">
          <section class="rounded-lg border border-default bg-default p-5">
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="min-w-0 space-y-1">
                <p class="flex flex-wrap items-center gap-2 text-sm text-muted">
                  <span class="font-mono font-medium text-default">{{ data.pathway.cbg_code }}</span>
                  <span aria-hidden="true">·</span>
                  <span>{{ data.pathway.mdc_name }}</span>
                </p>
                <h3 class="text-xl font-semibold text-highlighted">{{ data.pathway.name }}</h3>
                <p class="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted">
                  <PathwayStageBadge stage="AKTIF" />
                  <span v-if="activation"><UIcon name="i-lucide-stamp" class="inline size-3.5" /> Disahkan {{ activation.actor_name }} · {{ formatDate(activation.created_at) }}</span>
                  <span><UIcon name="i-lucide-bed-double" class="inline size-3.5" /> {{ formatNumber(data.pathway.episodes) }} episode</span>
                </p>
              </div>
              <UButton :to="`/library/${data.pathway.cbg_code}`" label="Buka detail CP" icon="i-lucide-external-link" color="neutral" variant="outline" size="sm" />
            </div>
            <div class="mt-4 flex flex-wrap items-center gap-3 border-t border-default pt-4">
              <span class="text-sm text-muted">Tampilkan untuk</span>
              <CommonSeverityFilter :model-value="severity" :allow-all="false" @update:model-value="(s: number) => go({ severity: s })" />
              <span v-if="status === 'pending'" class="text-xs text-muted">Memuat…</span>
            </div>
          </section>

          <UTabs
            :model-value="tab"
            :items="TABS"
            :content="false"
            variant="link"
            :ui="{ list: 'border-b border-default overflow-x-auto', trigger: 'shrink-0' }"
            @update:model-value="(v) => go({ tab: String(v) })"
          />

          <div :class="{ 'opacity-60 transition-opacity': status === 'pending' }">
            <ActiveDailyPlan v-if="tab === 'rencana'" :plan="data.plan" :severity="severity" />
            <ActiveClaimPattern v-else-if="tab === 'pola'" :diagnoses="data.claim_pattern.diagnoses" :procedures="data.claim_pattern.procedures" />
            <ActiveReferencesApproval v-else-if="tab === 'acuan'" :data="data.references_approval" />
            <ActiveCriteriaView v-else :criteria="data.criteria" />
          </div>
        </template>

        <USkeleton v-else class="h-64 w-full" />
      </div>
    </div>
  </div>
</template>
