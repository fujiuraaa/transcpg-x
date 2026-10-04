<script setup lang="ts">
// Tab "Acuan Klinis": telaah dokumen kandidat dan butir-butirnya.
// Penetapan acuan dilakukan dari panel Acuan Standar.
import type { CandidateRow, GuidelineItemRow, PathwayPermissions } from '~/types/api'
import type { GuidelineCategory } from '~/types/domain'

const props = defineProps<{ code: string, permissions: PathwayPermissions }>()
const api = usePathwayApi(() => props.code)
const guidelines = useGuidelineApi()

const search = ref('')
const debounced = useDebounced(search, 300)
const { data: candidates, status, refresh: refreshCandidates } = await useAsyncData(
  () => `${pathwayKey(props.code, 'candidates')}:${debounced.value}`,
  () => api.candidates(debounced.value),
)
const { data: references } = await useAsyncData(pathwayKey(props.code, 'references'), () => api.references())
const assignedIds = computed(() => new Set((references.value ?? []).map(r => r.document_id)))

const selected = ref<CandidateRow | null>(null)
watch(candidates, (list) => {
  if (!list?.length) return
  // Pertahankan pilihan, tetapi pakai baris terbaru (mis. status lampiran).
  selected.value = list.find(c => c.id === selected.value?.id) ?? selected.value ?? list[0]!
}, { immediate: true })

const { data: items, status: itemsStatus, refresh: refreshItems } = await useAsyncData(
  () => `guideline:${selected.value?.id ?? 0}:items`,
  () => (selected.value ? guidelines.items(selected.value.id) : Promise.resolve([] as GuidelineItemRow[])),
)
const grouped = computed(() => {
  const map = new Map<GuidelineCategory, GuidelineItemRow[]>()
  for (const it of items.value ?? []) {
    if (!map.has(it.category)) map.set(it.category, [])
    map.get(it.category)!.push(it)
  }
  return [...map.entries()]
})

// Pendaftaran dokumen & butir mengikuti wewenang menetapkan acuan saat Draf.
const canRegister = computed(() => props.permissions.can_edit.acuan)
const createOpen = ref(false)
const itemOpen = ref(false)
</script>

<template>
  <div class="grid gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
    <!-- Daftar kandidat -->
    <section class="space-y-3">
      <div class="flex gap-2">
        <UInput v-model="search" icon="i-lucide-search" placeholder="Cari dokumen lain…" class="flex-1" />
        <UButton v-if="canRegister" icon="i-lucide-file-plus" color="neutral" variant="outline" aria-label="Daftarkan dokumen panduan sendiri" title="Daftarkan dokumen panduan sendiri" @click="createOpen = true" />
      </div>
      <div v-if="status === 'pending' && !candidates" class="space-y-2">
        <USkeleton v-for="i in 3" :key="i" class="h-16 w-full" />
      </div>
      <p v-else-if="!candidates?.length" class="text-sm text-muted">Tidak ada dokumen yang cocok.</p>
      <ul v-else class="space-y-2">
        <li v-for="c in candidates" :key="c.id">
          <button
            type="button"
            class="w-full rounded-md border p-3 text-left transition-colors"
            :class="selected?.id === c.id ? 'border-primary bg-primary/5 ring-1 ring-primary' : 'border-default hover:bg-elevated'"
            @click="selected = c"
          >
            <div class="flex items-start justify-between gap-2">
              <p class="text-sm font-medium">{{ c.title }}</p>
              <UIcon v-if="assignedIds.has(c.id)" name="i-lucide-badge-check" class="size-4 shrink-0 text-success" title="Sudah ditetapkan sebagai acuan" />
            </div>
            <p class="mt-1 text-xs text-muted">
              {{ GUIDELINE_SOURCE[c.source_type] }}<template v-if="c.year"> · {{ c.year }}</template> · {{ c.item_count }} butir
              <template v-if="c.has_attachment"> · <span class="inline-flex items-center gap-0.5 text-toned"><UIcon name="i-lucide-paperclip" class="size-3" />PDF</span></template>
              <template v-if="!c.matches_diagnosis"> · <span class="text-warning">tidak cocok diagnosis utama</span></template>
            </p>
          </button>
        </li>
      </ul>
    </section>

    <!-- Butir dokumen terpilih -->
    <section class="rounded-lg border border-default bg-default">
      <template v-if="selected">
        <header class="flex items-start justify-between gap-3 border-b border-default p-4">
          <div>
            <h3 class="font-medium">{{ selected.title }}</h3>
            <p class="text-xs text-muted">
              {{ GUIDELINE_SOURCE[selected.source_type] }}
              <template v-if="selected.issuer"> · {{ selected.issuer }}</template>
              <template v-if="selected.regulation_ref"> · {{ selected.regulation_ref }}</template>
              <template v-if="selected.icd10_codes.length"> · ICD-10 {{ selected.icd10_codes.join(', ') }}</template>
            </p>
          </div>
          <UButton v-if="canRegister" label="Tambah butir" icon="i-lucide-plus" size="sm" color="neutral" variant="outline" @click="itemOpen = true" />
        </header>
        <div class="border-b border-default p-4">
          <h4 class="mb-2 text-xs font-medium uppercase tracking-wide text-muted">Lampiran PDF</h4>
          <PathwayGuidelineAttachment :key="selected.id" :document="selected" :can-edit="canRegister" @changed="refreshCandidates" />
        </div>
        <div class="max-h-[32rem] space-y-5 overflow-y-auto p-4">
          <USkeleton v-if="itemsStatus === 'pending'" class="h-24 w-full" />
          <UEmpty
            v-else-if="!grouped.length"
            icon="i-lucide-file-search"
            title="Belum ada butir acuan"
            description="Butir panduan dari dokumen ini belum diekstrak."
            variant="naked"
            size="sm"
          />
          <section v-for="[cat, rows] in grouped" :key="cat" class="space-y-2">
            <h4 class="text-xs font-medium uppercase tracking-wide text-muted">{{ GUIDELINE_CATEGORY[cat] }}</h4>
            <article v-for="it in rows" :key="it.id" class="rounded-md bg-elevated/60 p-3">
              <div class="flex items-baseline justify-between gap-2">
                <p class="text-sm font-medium">{{ it.title }}</p>
                <span class="shrink-0 text-xs text-muted">hlm. {{ it.page }}</span>
              </div>
              <blockquote class="mt-1 border-l-2 border-primary/40 pl-3 text-sm text-toned">{{ it.quote }}</blockquote>
            </article>
          </section>
        </div>
      </template>
      <UEmpty v-else icon="i-lucide-book-open" title="Pilih dokumen" description="Pilih dokumen di kiri untuk menelaah butir acuannya." variant="naked" />
    </section>

    <PathwayGuidelineCreateModal v-model:open="createOpen" @created="search = $event" />
    <PathwayGuidelineItemModal v-if="selected" v-model:open="itemOpen" :document="selected" @saved="refreshItems" />
  </div>
</template>
