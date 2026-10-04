<script setup lang="ts">
// §12 Evaluasi CP — kendali mutu & kendali biaya untuk KSM/Komite Medik dan
// Komite Mutu. Hanya klaim sesudah CP disahkan yang dibandingkan.
definePageMeta({ title: 'Evaluasi CP' })

const route = useRoute()
const router = useRouter()
const { list } = useEvaluationApi()

type Tab = 'aktif' | 'belum'
const tab = computed<Tab>(() => (route.query.tab as Tab) ?? 'aktif')
const openCode = computed(() => (route.query.rinci as string) ?? null)

const { data, status, error } = await useAsyncData(() => `evaluation:${tab.value}`, () => list(tab.value))
// Ringkasan selalu dari tab aktif (tab "belum" hanya membawa cp_aktif).
const { data: activeData } = await useAsyncData('evaluation:summary', () => list('aktif'))

const filter = ref<'semua' | 'tinjau'>('semua')
// Saringan "Perlu ditinjau" hanya berlaku di tab CP Aktif (tombolnya hanya ada di sana).
const rows = computed(() => (data.value?.items ?? []).filter(r => tab.value !== 'aktif' || filter.value === 'semua' || r.needs_review))

function toggle(code: string) {
  router.replace({ query: { ...route.query, rinci: openCode.value === code ? undefined : code } })
}

const tabs = computed(() => [
  { label: 'CP Aktif — evaluasi kinerja', value: 'aktif', icon: 'i-lucide-activity' },
  { label: 'Belum Aktif — belum bisa dievaluasi', value: 'belum', icon: 'i-lucide-clock' },
])
</script>

<template>
  <div class="mx-auto max-w-[1400px] space-y-6">
    <header class="bg-brand-soft flex flex-wrap items-end justify-between gap-3 rounded-xl border border-primary/15 px-5 py-4">
      <div>
        <h2 class="text-2xl font-semibold text-highlighted">Evaluasi CP</h2>
        <p class="text-sm text-muted">Kendali mutu dan kendali biaya CP yang sudah aktif — untuk KSM/Komite Medik dan Komite Mutu.</p>
      </div>
      <UBadge label="Alat deteksi dini, bukan audit menyeluruh" icon="i-lucide-radar" color="neutral" variant="outline" />
    </header>

    <EvaluationStatCards :summary="activeData?.summary ?? null" />
    <EvaluationThresholdInfo />

    <div class="space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <UTabs
          :model-value="tab"
          :items="tabs"
          :content="false"
          variant="link"
          :ui="{ list: 'border-b border-default overflow-x-auto', trigger: 'shrink-0' }"
          @update:model-value="(v) => router.replace({ query: v === 'aktif' ? {} : { tab: v } })"
        />
        <div v-if="tab === 'aktif'" class="inline-flex rounded-md border border-default p-0.5" role="radiogroup" aria-label="Saring hasil">
          <button
            v-for="opt in [{ v: 'semua', l: 'Semua' }, { v: 'tinjau', l: 'Perlu ditinjau' }] as const"
            :key="opt.v"
            type="button"
            role="radio"
            :aria-checked="filter === opt.v"
            class="rounded px-3 py-1 text-xs font-medium"
            :class="filter === opt.v ? 'bg-primary text-inverted' : 'text-muted hover:text-default'"
            @click="filter = opt.v"
          >
            {{ opt.l }}
          </button>
        </div>
      </div>

      <div v-if="status === 'pending' && !data" class="space-y-3">
        <USkeleton v-for="i in 3" :key="i" class="h-20 w-full" />
      </div>
      <UAlert v-else-if="error" color="error" variant="subtle" :title="apiErrorMessage(error)" />
      <UEmpty
        v-else-if="!rows.length"
        :icon="tab === 'aktif' ? 'i-lucide-circle-check' : 'i-lucide-clock'"
        :title="tab === 'belum' ? 'Semua CP sudah Aktif' : filter === 'tinjau' ? 'Tidak ada CP yang perlu ditinjau' : 'Belum ada CP Aktif'"
        :description="tab === 'aktif' && filter === 'semua' ? 'Evaluasi tersedia setelah CP disahkan Direktur dan klaim sesudahnya masuk.' : undefined"
        variant="outline"
      />
      <ul v-else class="space-y-3">
        <EvaluationResultRow
          v-for="r in rows"
          :key="r.id"
          :row="r"
          :open="openCode === r.cbg_code"
          @toggle="toggle(r.cbg_code)"
        />
      </ul>
    </div>
  </div>
</template>
