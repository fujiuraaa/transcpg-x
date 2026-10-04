<script setup lang="ts">
// Kolom pencarian master data (ICD-10, ICD-9-CM, LOINC, KFA, KPTL, SNOMED).
// Hasil dicari di server (minimal 2 huruf); konsep nonaktif tidak bisa dipilih.
import type { MasterKind } from '~/types/domain'

export interface MasterOption {
  label: string
  code: string
  name: string
  extra?: string | boolean | null
  disabled?: boolean
}

const props = defineProps<{ kind: MasterKind, placeholder?: string, semanticTag?: string }>()
const model = defineModel<MasterOption | undefined>()

const { search } = useMasterApi()
const term = ref('')
const debounced = useDebounced(term, 250)
const options = ref<MasterOption[]>([])
const loading = ref(false)
const failed = ref(false)
// Hanya respons pencarian terakhir yang dipakai (yang lebih lambat dibuang).
let seq = 0

watch([debounced, () => props.kind, () => props.semanticTag], async ([q]) => {
  // Setelah memilih, kolom berisi label item — jangan dicari ulang.
  if (model.value && q === model.value.label) return
  const mine = ++seq
  if (q.trim().length < 2) {
    options.value = []
    loading.value = false
    return
  }
  loading.value = true
  failed.value = false
  try {
    const rows = await search(props.kind, q.trim(), props.semanticTag)
    if (mine !== seq) return
    options.value = rows.map(r => ({
      label: `${r.code} · ${r.name}`,
      code: r.code,
      name: r.name,
      extra: r.extra,
      disabled: !r.active,
    }))
  }
  catch {
    if (mine !== seq) return
    options.value = []
    failed.value = true
  }
  finally {
    if (mine === seq) loading.value = false
  }
})

// Pilihan dikosongkan dari luar (mis. panel dibuka untuk baris lain) → kosongkan teks cari.
watch(model, (v) => {
  if (!v) term.value = ''
})

// Ganti jenis master → kosongkan pilihan lama.
watch(() => props.kind, () => {
  model.value = undefined
  term.value = ''
})
</script>

<template>
  <UInputMenu
    v-model="model"
    v-model:search-term="term"
    :items="options"
    :loading="loading"
    ignore-filter
    icon="i-lucide-search"
    :placeholder="placeholder ?? 'Ketik kode atau nama (min. 2 huruf)'"
    class="w-full"
  >
    <template #item-label="{ item }">
      <span class="font-mono text-xs text-muted">{{ item.code }}</span>
      <span class="ml-2">{{ item.name }}</span>
      <span v-if="typeof item.extra === 'string' && item.extra" class="ml-2 text-xs text-muted">({{ item.extra }})</span>
      <span v-if="item.disabled" class="ml-2 text-xs text-error">pensiun</span>
    </template>
    <template #empty>
      <span v-if="failed" class="text-sm text-error">Gagal mencari — periksa koneksi lalu ketik ulang</span>
      <span v-else class="text-sm text-muted">{{ term.trim().length < 2 ? 'Ketik minimal 2 huruf' : 'Tidak ditemukan di master' }}</span>
    </template>
  </UInputMenu>
</template>
