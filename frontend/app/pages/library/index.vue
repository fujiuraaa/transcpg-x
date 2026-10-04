<script setup lang="ts">
// §4 CP Library — seluruh CP yang layak disusun (grouper ≥5 episode).
// Saringan & halaman disimpan di URL agar bisa dibagikan dan tombol Kembali bekerja.
import type { LibraryQuery } from '~/types/api'

definePageMeta({ title: 'CP Library' })

const route = useRoute()
const router = useRouter()
const { list } = useLibraryApi()
const { selected } = useHospitalProfile()

const filter = computed<LibraryQuery>(() => ({
  q: (route.query.q as string) ?? '',
  mdc: (route.query.mdc as string) ?? '',
  acuan: (route.query.acuan as LibraryQuery['acuan']) ?? '',
  sort: (route.query.sort as LibraryQuery['sort']) ?? 'volume',
  page: Number(route.query.page ?? 1),
}))

/** Mengubah saringan selalu kembali ke halaman 1; nilai kosong dibuang dari URL. */
function setFilter(patch: Partial<LibraryQuery>) {
  const next: Record<string, string | number> = { ...(route.query as Record<string, string>), ...patch, page: patch.page ?? 1 }
  for (const k of Object.keys(next)) {
    if (next[k] === '' || (k === 'sort' && next[k] === 'volume') || (k === 'page' && next[k] === 1)) delete next[k]
  }
  router.replace({ query: next })
}

const { data, status, error } = await useAsyncData(
  () => `library:${JSON.stringify(filter.value)}:${selected.value?.id ?? 0}`,
  () => list(filter.value),
)

const page = computed({
  get: () => filter.value.page ?? 1,
  set: (p) => {
    setFilter({ page: p })
    window.scrollTo({ top: 0, behavior: 'smooth' })
  },
})
</script>

<template>
  <div class="mx-auto max-w-[1400px] space-y-6">
    <header class="bg-brand-soft rounded-xl border border-primary/15 px-5 py-4">
      <h2 class="text-2xl font-semibold text-highlighted">CP Library</h2>
      <p class="text-sm text-muted">
        Seluruh grouper INA-CBG yang layak disusun menjadi Clinical Pathway (≥ 5 episode rawat inap).
        Pilih satu CP untuk menyusun acuan, rencana klinis, dan mengajukan pengesahan.
      </p>
    </header>

    <LibraryFilterBar :filter="filter" :total="data?.total ?? null" @change="setFilter" />

    <div v-if="status === 'pending' && !data" class="space-y-3">
      <USkeleton v-for="i in 4" :key="i" class="h-32 w-full rounded-lg" />
    </div>

    <UAlert v-else-if="error" color="error" variant="subtle" icon="i-lucide-circle-alert" :title="apiErrorMessage(error)" />

    <UEmpty
      v-else-if="!data?.items.length"
      icon="i-lucide-search-x"
      title="Tidak ada CP yang cocok"
      description="Coba ubah kata kunci atau atur ulang saringan."
      :actions="[{ label: 'Atur ulang saringan', icon: 'i-lucide-rotate-ccw', color: 'neutral', variant: 'outline', onClick: () => setFilter({ q: '', mdc: '', acuan: '', sort: 'volume' }) }]"
      variant="outline"
    />

    <template v-else>
      <ul class="space-y-3" :class="{ 'opacity-60 transition-opacity': status === 'pending' }" :aria-busy="status === 'pending'">
        <li v-for="cp in data.items" :key="cp.id">
          <LibraryPathwayCard :cp="cp" />
        </li>
      </ul>

      <div class="flex flex-wrap items-center justify-between gap-3 border-t border-default pt-4">
        <p class="text-sm text-muted">
          Menampilkan
          <span class="tabular-nums text-default">{{ (data.page - 1) * data.page_size + 1 }}–{{ Math.min(data.page * data.page_size, data.total) }}</span>
          dari <span class="tabular-nums text-default">{{ formatNumber(data.total) }}</span> CP
        </p>
        <UPagination
          v-if="data.total > data.page_size"
          v-model:page="page"
          :total="data.total"
          :items-per-page="data.page_size"
        />
      </div>
    </template>
  </div>
</template>
