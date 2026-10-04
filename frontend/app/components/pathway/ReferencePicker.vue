<script setup lang="ts">
// Slideover "Tetapkan acuan": pilih dokumen panduan untuk CP atau satu
// sub-CP. Kandidat dikelompokkan menurut kecocokan dengan diagnosis utama.
import type { CandidateRow, SubCP } from '~/types/api'
import type { GuidelineSource } from '~/types/domain'

const props = defineProps<{ code: string, assigned: number[], subCp?: SubCP | null }>()
const open = defineModel<boolean>('open', { default: false })

const api = usePathwayApi(() => props.code)
const toast = useToast()

const search = ref('')
const debounced = useDebounced(search, 300)
// Kunci terpisah dari tab Acuan Klinis: kotak cari keduanya bisa berbeda.
const { data: candidates, status, execute } = await useAsyncData(
  () => `${pathwayKey(props.code, 'picker-candidates')}:${debounced.value}`,
  () => api.candidates(debounced.value),
  { immediate: false },
)

const groups = computed(() => {
  const list = candidates.value ?? []
  return [
    { title: 'Cocok dengan diagnosis utama', items: list.filter(c => c.matches_diagnosis) },
    { title: 'Hasil pencarian lain', items: list.filter(c => !c.matches_diagnosis) },
  ].filter(g => g.items.length)
})

const selected = ref<CandidateRow | null>(null)
const note = ref('')
const busy = ref(false)

// Setiap kali dibuka (bisa untuk sub-CP lain): mulai bersih, lalu muat kandidat.
watch(open, (v) => {
  if (!v) return
  selected.value = null
  note.value = ''
  search.value = ''
  execute()
})

async function assign() {
  if (!selected.value) return
  busy.value = true
  try {
    await api.addReference({ document_id: selected.value.id, sub_cp_id: props.subCp?.id ?? null, note: note.value.trim() || null })
    toast.add({ title: 'Acuan ditetapkan', description: selected.value.title, color: 'success' })
    selected.value = null
    note.value = ''
    open.value = false
    await refreshNuxtData()
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busy.value = false
  }
}

const SOURCE_COLOR: Partial<Record<GuidelineSource, 'primary' | 'secondary' | 'neutral'>> = {
  PNPK: 'primary', PPK_RSCM: 'secondary', PPK_ASOSIASI: 'neutral',
}
</script>

<template>
  <USlideover
    v-model:open="open"
    :title="subCp ? `Tetapkan acuan sub-CP ${subCp.icd10_code}` : 'Tetapkan acuan standar'"
    :description="subCp ? subCp.label : 'Pilih dokumen panduan yang menjadi dasar isi CP.'"
    :ui="{ content: 'max-w-xl' }"
  >
    <template #body>
      <div class="space-y-4">
        <UInput v-model="search" icon="i-lucide-search" placeholder="Cari judul dokumen lain…" class="w-full" />

        <div v-if="status === 'pending'" class="space-y-2">
          <USkeleton v-for="i in 3" :key="i" class="h-16 w-full" />
        </div>

        <UEmpty
          v-else-if="!groups.length"
          icon="i-lucide-search-x"
          title="Tidak ada dokumen"
          description="Coba kata kunci lain, atau daftarkan dokumen panduan sendiri."
          variant="naked"
        />

        <section v-for="g in groups" :key="g.title" class="space-y-2">
          <h4 class="text-xs font-medium uppercase tracking-wide text-muted">{{ g.title }}</h4>
          <ul class="space-y-2">
            <li v-for="c in g.items" :key="c.id">
              <button
                type="button"
                class="w-full rounded-md border p-3 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-60"
                :class="selected?.id === c.id ? 'border-primary bg-primary/5 ring-1 ring-primary' : 'border-default hover:bg-elevated'"
                :disabled="assigned.includes(c.id)"
                @click="selected = c"
              >
                <div class="flex items-start justify-between gap-2">
                  <p class="text-sm font-medium">{{ c.title }}</p>
                  <UBadge v-if="assigned.includes(c.id)" label="Sudah ditetapkan" size="sm" color="success" variant="subtle" />
                </div>
                <div class="mt-1 flex flex-wrap items-center gap-1.5 text-xs text-muted">
                  <UBadge :label="GUIDELINE_SOURCE[c.source_type]" size="sm" :color="SOURCE_COLOR[c.source_type] ?? 'neutral'" variant="subtle" />
                  <span v-if="c.year">{{ c.year }}</span>
                  <span v-if="c.regulation_ref">· {{ c.regulation_ref }}</span>
                  <span>· {{ c.item_count }} butir</span>
                  <span v-if="c.has_attachment" class="inline-flex items-center gap-0.5">· <UIcon name="i-lucide-paperclip" class="size-3" />PDF</span>
                  <span v-if="c.icd10_codes.length">· ICD-10 {{ c.icd10_codes.join(', ') }}</span>
                </div>
              </button>
            </li>
          </ul>
        </section>
      </div>
    </template>

    <template #footer>
      <div class="w-full space-y-3">
        <UFormField v-if="selected" label="Catatan (opsional)">
          <UInput v-model="note" placeholder="Mis. edisi terbaru, berlaku untuk dewasa" class="w-full" />
        </UFormField>
        <div class="flex justify-end gap-2">
          <UButton label="Batal" color="neutral" variant="ghost" @click="open = false" />
          <UButton label="Tetapkan sebagai acuan" icon="i-lucide-check" :disabled="!selected" :loading="busy" @click="assign" />
        </div>
      </div>
    </template>
  </USlideover>
</template>
