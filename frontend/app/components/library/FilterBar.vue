<script setup lang="ts">
// Saringan CP Library: pencarian, kelompok diagnosis (MDC), status acuan,
// dan urutan. Nilainya disimpan di URL oleh halaman induk.
import type { LibraryQuery } from '~/types/api'

const props = defineProps<{ filter: LibraryQuery, total: number | null }>()
const emit = defineEmits<{ change: [patch: Partial<LibraryQuery>] }>()

const { mdc } = useDashboardApi()
const { data: mdcList } = await useAsyncData('mdc', mdc)
const mdcItems = computed(() => [
  { label: 'Semua kelompok diagnosis', value: 'semua' },
  ...(mdcList.value ?? []).map(m => ({ label: `${m.code} — ${m.name}`, value: String(m.code) })),
])

// Pencarian ditunda sejenak agar tidak memanggil API di setiap ketukan.
const q = ref(props.filter.q ?? '')
const debounced = useDebounced(q, 300)
watch(debounced, v => emit('change', { q: v }))
watch(() => props.filter.q, (v) => { if ((v ?? '') !== q.value) q.value = v ?? '' })

const hasFilter = computed(() => !!(props.filter.q || props.filter.mdc || props.filter.acuan || (props.filter.sort && props.filter.sort !== 'volume')))
</script>

<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-center gap-2">
      <UInput
        v-model="q"
        icon="i-lucide-search"
        placeholder="Cari kode, nama CP, atau kelompok diagnosis"
        class="min-w-64 flex-1"
        aria-label="Cari CP"
      >
        <template v-if="q" #trailing>
          <UButton icon="i-lucide-x" size="xs" color="neutral" variant="link" aria-label="Hapus pencarian" @click="q = ''" />
        </template>
      </UInput>
      <USelect
        :model-value="filter.mdc || 'semua'"
        :items="mdcItems"
        icon="i-lucide-folder-tree"
        class="w-64"
        aria-label="Kelompok diagnosis"
        @update:model-value="(v) => emit('change', { mdc: v === 'semua' ? '' : String(v) })"
      />
      <USelect
        :model-value="filter.acuan || 'semua'"
        :items="[{ label: 'Semua status acuan', value: 'semua' }, { label: 'Sudah punya acuan', value: 'sudah' }, { label: 'Belum punya acuan', value: 'belum' }]"
        icon="i-lucide-book-open"
        class="w-52"
        aria-label="Status acuan"
        @update:model-value="(v) => emit('change', { acuan: v === 'semua' ? '' : v as LibraryQuery['acuan'] })"
      />
      <USelect
        :model-value="filter.sort || 'volume'"
        :items="[{ label: 'Urut: volume kasus', value: 'volume' }, { label: 'Urut: target LOS', value: 'los' }, { label: 'Urut: kode grouper', value: 'kode' }, { label: 'Urut: belum punya acuan', value: 'tanpa_acuan' }]"
        icon="i-lucide-arrow-down-wide-narrow"
        class="w-56"
        aria-label="Urutan"
        @update:model-value="(v) => emit('change', { sort: v as LibraryQuery['sort'] })"
      />
    </div>
    <div class="flex items-center gap-3 text-sm text-muted">
      <span v-if="total != null"><span class="font-medium tabular-nums text-default">{{ formatNumber(total) }}</span> CP</span>
      <UButton
        v-if="hasFilter"
        label="Atur ulang saringan"
        icon="i-lucide-rotate-ccw"
        size="xs"
        color="neutral"
        variant="link"
        @click="q = ''; emit('change', { q: '', mdc: '', acuan: '', sort: 'volume' })"
      />
    </div>
  </div>
</template>
