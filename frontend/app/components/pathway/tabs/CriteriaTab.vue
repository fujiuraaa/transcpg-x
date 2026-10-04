<script setup lang="ts">
// Tab "Kriteria": pasien mana yang cocok (inklusi) dan tidak boleh masuk
// (eksklusi). Aturan wewenang sama dengan Rencana Klinis.
import type { Criterion, PathwayPermissions } from '~/types/api'
import type { CriterionKind } from '~/types/domain'

const props = defineProps<{ code: string, permissions: PathwayPermissions }>()
const api = usePathwayApi(() => props.code)
const toast = useToast()

const { data, status } = await useAsyncData(pathwayKey(props.code, 'criteria'), () => api.criteria())
const canEdit = computed(() => props.permissions.can_edit.kriteria)

const COLUMNS: { kind: CriterionKind, title: string, hint: string, icon: string, color: string, example: string }[] = [
  { kind: 'INKLUSI', title: 'Inklusi', hint: 'Pasien yang cocok untuk CP ini', icon: 'i-lucide-circle-check', color: 'text-success', example: 'Mis. Usia ≥ 18 tahun' },
  { kind: 'EKSKLUSI', title: 'Eksklusi', hint: 'Pasien yang tidak boleh masuk CP ini', icon: 'i-lucide-circle-minus', color: 'text-error', example: 'Mis. Kehamilan' },
]

const drafts = reactive<Record<CriterionKind, string>>({ INKLUSI: '', EKSKLUSI: '' })
const busy = ref<CriterionKind | number | null>(null)

const list = (kind: CriterionKind) => (data.value ?? []).filter(c => c.kind === kind)

async function add(kind: CriterionKind) {
  const text = drafts[kind].trim()
  if (!text) return
  busy.value = kind
  try {
    await api.addCriterion(kind, text)
    drafts[kind] = ''
    await refreshNuxtData()
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busy.value = null
  }
}

async function remove(c: Criterion) {
  busy.value = c.id
  try {
    await api.deleteCriterion(c.id)
    toast.add({ title: 'Kriteria dihapus', description: c.description, color: 'success' })
    await refreshNuxtData()
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busy.value = null
  }
}
</script>

<template>
  <div class="space-y-4">
    <p v-if="permissions.content_locked" class="flex items-center gap-1 text-sm text-muted">
      <UIcon name="i-lucide-lock" class="size-4" /> Dikunci selama ditinjau Komite/Direktur.
    </p>

    <div class="grid gap-4 md:grid-cols-2">
      <section v-for="col in COLUMNS" :key="col.kind" class="rounded-lg border border-default bg-default">
        <header class="flex items-center gap-2 border-b border-default px-4 py-3">
          <UIcon :name="col.icon" class="size-5" :class="col.color" />
          <div class="flex-1">
            <h3 class="font-medium">{{ col.title }}</h3>
            <p class="text-xs text-muted">{{ col.hint }}</p>
          </div>
          <UBadge :label="String(list(col.kind).length)" color="neutral" variant="subtle" />
        </header>

        <USkeleton v-if="status === 'pending' && !data" class="m-4 h-16" />
        <p v-else-if="!list(col.kind).length" class="px-4 py-6 text-center text-sm text-muted">Belum ada kriteria {{ col.title.toLowerCase() }}.</p>
        <ul v-else class="divide-y divide-default">
          <li v-for="c in list(col.kind)" :key="c.id" class="flex items-start gap-2 px-4 py-2.5">
            <span class="mt-2 size-1.5 shrink-0 rounded-full bg-current" :class="col.color" aria-hidden="true" />
            <span class="flex-1 text-sm">{{ c.description }}</span>
            <UButton
              v-if="canEdit"
              icon="i-lucide-trash-2"
              size="xs"
              color="error"
              variant="ghost"
              :loading="busy === c.id"
              :aria-label="`Hapus kriteria ${c.description}`"
              @click="remove(c)"
            />
          </li>
        </ul>

        <form v-if="canEdit" class="flex gap-2 border-t border-default p-3" @submit.prevent="add(col.kind)">
          <UInput v-model="drafts[col.kind]" :placeholder="col.example" class="flex-1" :aria-label="`Kriteria ${col.title} baru`" />
          <UButton type="submit" icon="i-lucide-plus" label="Tambah" color="neutral" variant="outline" :disabled="!drafts[col.kind].trim()" :loading="busy === col.kind" />
        </form>
      </section>
    </div>
  </div>
</template>
