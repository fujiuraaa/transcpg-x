<script setup lang="ts">
// §14 Padanan SNOMED-CT — terminologi klinis internasional untuk pertukaran
// data (mis. SATUSEHAT). Sebagian besar otomatis; manusia meninjau yang
// "ragu" dan mengisi yang belum dipadankan.
import type { SnomedRow, SnomedVocab } from '~/types/api'

definePageMeta({ title: 'Padanan SNOMED-CT' })

const route = useRoute()
const router = useRouter()
const { snomed } = useMappingApi()

type Filter = '' | 'ragu' | 'belum'
const filter = computed<Filter>(() => (route.query.filter as Filter | undefined) ?? 'ragu')
const vocab = computed<'' | SnomedVocab>(() => (route.query.vocab as SnomedVocab | undefined) ?? '')
const q = ref('')

const { data: summary, refresh: refreshSummary } = await useAsyncData('snomed:summary', snomed.summary)
const { data, status, error, refresh } = await useAsyncData(
  () => `snomed:${filter.value}:${vocab.value}`,
  () => snomed.worklist(filter.value, vocab.value),
)
const canManage = computed(() => !!summary.value?.can_manage)

// Pencarian lokal atas kode, nama, atau istilah SNOMED.
const rows = computed(() => {
  const t = q.value.trim().toLowerCase()
  const list = data.value ?? []
  if (!t) return list
  return list.filter(r => [r.code, r.name, r.preferred_term, r.concept_id].some(v => v?.toLowerCase().includes(t)))
})

const totals = computed(() => (summary.value?.vocabularies ?? []).reduce(
  (a, v) => ({ ragu: a.ragu + v.perlu_ditinjau, belum: a.belum + v.belum }),
  { ragu: 0, belum: 0 },
))

const tabs = computed(() => [
  { label: 'Perlu Ditinjau', value: 'ragu', icon: 'i-lucide-circle-help', badge: totals.value.ragu ? { label: String(totals.value.ragu), color: 'warning' as const, variant: 'solid' as const } : undefined },
  { label: 'Belum Dipadankan', value: 'belum', icon: 'i-lucide-circle-dashed', badge: totals.value.belum ? { label: String(totals.value.belum), color: 'neutral' as const, variant: 'subtle' as const } : undefined },
  { label: 'Semua Kode', value: 'semua', icon: 'i-lucide-list' },
])

function setQuery(patch: Record<string, string | undefined>) {
  router.replace({ query: { ...route.query, ...patch } })
}

const pickerRow = ref<SnomedRow | null>(null)
const pickerOpen = ref(false)
function openPicker(r: SnomedRow) {
  pickerRow.value = r
  pickerOpen.value = true
}
async function onSaved() {
  await Promise.all([refresh(), refreshSummary()])
}
</script>

<template>
  <div class="mx-auto max-w-[1400px] space-y-6">
    <header class="bg-brand-soft rounded-xl border border-primary/15 px-5 py-4">
      <h2 class="text-2xl font-semibold text-highlighted">Padanan SNOMED-CT</h2>
      <p class="text-sm text-muted">
        Terminologi klinis internasional untuk pertukaran data antarsistem, mis. SATUSEHAT. Sebagian besar kode sudah
        dipadankan otomatis — yang dikerjakan manusia: padanan <strong>ragu</strong> dan yang <strong>belum dipadankan</strong>.
      </p>
    </header>

    <USkeleton v-if="!summary" class="h-28 w-full" />
    <MappingSnomedCoverage v-else :rows="summary.vocabularies" />
    <MappingReadOnlyNotice :show="!!summary && !canManage" />

    <div class="space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <UTabs
          :model-value="filter || 'semua'"
          :items="tabs"
          :content="false"
          variant="link"
          :ui="{ list: 'border-b border-default overflow-x-auto', trigger: 'shrink-0' }"
          @update:model-value="(v) => setQuery({ filter: v === 'ragu' ? undefined : v === 'semua' ? '' : String(v) })"
        />
        <div class="flex flex-wrap gap-2">
          <USelect
            :model-value="vocab || 'semua'"
            :items="[{ label: 'Semua vokabuler', value: 'semua' }, { label: 'ICD-10 · diagnosis', value: 'ICD10' }, { label: 'ICD-9-CM · prosedur', value: 'ICD9CM' }]"
            class="w-48"
            aria-label="Vokabuler"
            @update:model-value="(v) => setQuery({ vocab: v === 'semua' ? undefined : String(v) })"
          />
          <UInput v-model="q" icon="i-lucide-search" placeholder="Cari kode, nama, atau istilah" class="w-64" />
        </div>
      </div>

      <USkeleton v-if="status === 'pending' && !data" class="h-48 w-full" />
      <UAlert v-else-if="error" color="error" variant="subtle" :title="apiErrorMessage(error)" />
      <UEmpty
        v-else-if="!rows.length"
        :icon="filter === 'ragu' && !q ? 'i-lucide-circle-check' : 'i-lucide-search-x'"
        :title="filter === 'ragu' && !q ? 'Tidak ada padanan yang perlu ditinjau' : filter === 'belum' && !q ? 'Semua kode sudah dipadankan' : 'Tidak ada kode yang cocok'"
        variant="outline"
      />
      <MappingSnomedTable v-else :rows="rows" :can-manage="canManage" @map="openPicker" />
    </div>

    <MappingSnomedPicker v-model:open="pickerOpen" :row="pickerRow" @saved="onSaved" />
  </div>
</template>
