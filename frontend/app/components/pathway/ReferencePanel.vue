<script setup lang="ts">
// Panel Acuan Standar CP (tingkat CP). Acuan sub-CP dikelola di tab
// Alur Pelaksanaan. Tanpa acuan, CP tidak dapat disahkan menjadi Aktif.
import type { PathwayDetail, ReferenceRow } from '~/types/api'

const props = defineProps<{ code: string, detail: PathwayDetail }>()
const api = usePathwayApi(() => props.code)
const toast = useToast()

const { data: references, status } = await useAsyncData(pathwayKey(props.code, 'references'), () => api.references())
const cpRefs = computed(() => (references.value ?? []).filter(r => r.sub_cp_id == null))

const perms = computed(() => props.detail.permissions)
const pickerOpen = ref(false)
const revoking = ref<ReferenceRow | null>(null)
const busy = ref(false)

async function revoke() {
  if (!revoking.value) return
  busy.value = true
  try {
    await api.removeReference(revoking.value.id)
    toast.add({ title: 'Acuan dicabut', color: 'success' })
    revoking.value = null
    await refreshNuxtData()
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="rounded-lg border border-default bg-default" aria-labelledby="ref-title">
    <div class="flex items-center justify-between gap-2 border-b border-default px-4 py-3">
      <h3 id="ref-title" class="font-medium text-highlighted">Acuan Standar</h3>
      <UBadge
        :label="cpRefs.length ? `${cpRefs.length} dokumen` : 'Belum ada'"
        :color="cpRefs.length ? 'success' : 'warning'"
        variant="subtle"
      />
    </div>

    <div class="space-y-3 p-4">
      <UAlert
        v-if="perms.content_locked"
        color="neutral"
        variant="subtle"
        icon="i-lucide-lock"
        title="Dikunci selama ditinjau"
        description="Agar yang disahkan Direktur sama dengan yang ditinjau Komite."
      />

      <USkeleton v-if="status === 'pending'" class="h-16 w-full" />

      <p v-else-if="!cpRefs.length" class="text-sm text-muted">
        Belum ditetapkan · {{ detail.cards.candidate_count }} kandidat cocok dengan diagnosis utama.
        <span class="block text-xs">Wajib minimal 1 dokumen sebelum CP dapat disahkan.</span>
      </p>

      <ul v-else class="space-y-2">
        <li v-for="r in cpRefs" :key="r.id" class="group rounded-md border border-default p-3">
          <div class="flex items-start gap-2">
            <UIcon name="i-lucide-file-text" class="mt-0.5 size-4 shrink-0 text-primary" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium leading-snug">{{ r.title }}</p>
              <p class="text-xs text-muted">
                {{ GUIDELINE_SOURCE[r.source_type] }}<template v-if="r.year"> · {{ r.year }}</template>
                <template v-if="r.regulation_ref"> · {{ r.regulation_ref }}</template>
              </p>
              <p v-if="r.note" class="mt-1 text-xs italic text-muted">{{ r.note }}</p>
              <p class="mt-1 text-xs text-dimmed">Ditetapkan {{ r.assigned_by_name }} · {{ formatDate(r.assigned_at) }}</p>
            </div>
            <UButton
              v-if="perms.can_edit.acuan"
              icon="i-lucide-x"
              size="xs"
              color="error"
              variant="ghost"
              :aria-label="`Cabut ${r.title}`"
              @click="revoking = r"
            />
          </div>
        </li>
      </ul>

      <UButton
        v-if="perms.can_edit.acuan"
        label="Tetapkan acuan"
        icon="i-lucide-plus"
        color="neutral"
        variant="outline"
        block
        @click="pickerOpen = true"
      />
    </div>

    <PathwayReferencePicker v-model:open="pickerOpen" :code="code" :assigned="cpRefs.map(r => r.document_id)" />

    <UModal
      :open="!!revoking"
      title="Cabut acuan?"
      :description="revoking?.title"
      @update:open="(v) => { if (!v) revoking = null }"
    >
      <template #body>
        <p class="text-sm">Pencabutan tercatat di riwayat perubahan CP.</p>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton label="Batal" color="neutral" variant="ghost" @click="revoking = null" />
          <UButton label="Cabut acuan" color="error" :loading="busy" @click="revoke" />
        </div>
      </template>
    </UModal>
  </section>
</template>
