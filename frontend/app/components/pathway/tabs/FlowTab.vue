<script setup lang="ts">
// Tab "Alur Pelaksanaan": putusan satu CP / perlu dipecah, sub-CP beserta
// acuannya (bisa disunting), dan posisi CP Library vs CP Aktif.
import type { PathwayPermissions, SubCP } from '~/types/api'

const props = defineProps<{ code: string, permissions: PathwayPermissions }>()
const api = usePathwayApi(() => props.code)
const toast = useToast()

const { data, status } = await useAsyncData(pathwayKey(props.code, 'flow'), () => api.flow())
const canEdit = computed(() => props.permissions.can_edit.acuan)

const pickerFor = ref<SubCP | null>(null)
const pickerOpen = computed({
  get: () => !!pickerFor.value,
  set: (v) => { if (!v) pickerFor.value = null },
})

// Cabut acuan sub-CP: lewat konfirmasi, dengan penjaga klik ganda.
const revoking = ref<{ id: number, title: string, sub: string } | null>(null)
const revokeBusy = ref(false)
async function revoke() {
  if (!revoking.value || revokeBusy.value) return
  revokeBusy.value = true
  try {
    await api.removeReference(revoking.value.id)
    toast.add({ title: 'Acuan sub-CP dicabut', description: revoking.value.title, color: 'success' })
    revoking.value = null
    await refreshNuxtData()
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    revokeBusy.value = false
  }
}
</script>

<template>
  <div v-if="status === 'pending' && !data" class="space-y-2">
    <USkeleton class="h-20 w-full" />
    <USkeleton class="h-40 w-full" />
  </div>

  <div v-else-if="data" class="space-y-6">
    <!-- Putusan -->
    <div
      class="flex gap-3 rounded-lg border p-4"
      :class="data.split_decision === 'PECAH' ? 'border-warning/40 bg-warning/5' : 'border-default bg-default'"
    >
      <UIcon
        :name="data.split_decision === 'PECAH' ? 'i-lucide-split' : 'i-lucide-circle-dot'"
        class="mt-0.5 size-6 shrink-0"
        :class="data.split_decision === 'PECAH' ? 'text-warning' : 'text-primary'"
      />
      <div>
        <h3 class="font-medium">
          {{ data.split_decision === 'PECAH' ? 'Perlu dipecah menjadi beberapa sub-CP'
            : data.split_decision === 'SATU' ? 'Dapat disusun sebagai satu CP' : 'Putusan belum tersedia' }}
        </h3>
        <p class="text-sm text-muted">
          <template v-if="data.split_decision === 'PECAH'">
            Grouper ini tercampur beberapa diagnosis yang signifikan. Tetapkan acuan untuk tiap sub-diagnosis di bawah — jumlah dokumen per sub-CP tidak dibatasi.
          </template>
          <template v-else-if="data.split_decision === 'SATU'">
            Grouper didominasi satu diagnosis, sehingga cukup satu acuan standar di tingkat CP.
          </template>
          <template v-else>Putusan diisi otomatis oleh analisis data klaim.</template>
        </p>
      </div>
    </div>

    <!-- Sub-CP -->
    <section v-if="data.sub_cps.length" class="space-y-2">
      <h3 class="text-xs font-medium uppercase tracking-wide text-muted">Sub-CP per diagnosis penyusun</h3>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th class="w-24">ICD-10</th>
              <th>Diagnosis</th>
              <th class="w-48">Porsi episode</th>
              <th>Acuan sub-CP</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in data.sub_cps" :key="s.id">
              <td class="font-mono">{{ s.icd10_code }}</td>
              <td class="font-medium">{{ s.label }}</td>
              <td>
                <div class="flex items-center gap-2">
                  <UProgress :model-value="(s.episode_share ?? 0) * 100" size="sm" class="flex-1" />
                  <span class="w-10 text-right text-xs tabular-nums text-muted">{{ formatPercent(s.episode_share) }}</span>
                </div>
              </td>
              <td>
                <div class="flex flex-wrap items-center gap-1.5">
                  <span
                    v-for="r in s.references"
                    :key="r.id"
                    class="inline-flex items-center gap-0.5 rounded-md bg-primary/10 py-0.5 pl-2 text-xs font-medium text-primary"
                    :class="canEdit ? 'pr-0.5' : 'pr-2'"
                  >
                    {{ r.title }}
                    <UButton
                      v-if="canEdit"
                      icon="i-lucide-x"
                      size="xs"
                      color="primary"
                      variant="link"
                      class="p-0.5"
                      :aria-label="`Cabut acuan ${r.title}`"
                      @click="revoking = { id: r.id, title: r.title, sub: s.label }"
                    />
                  </span>
                  <span v-if="!s.references.length" class="text-xs text-warning">Belum ada acuan</span>
                  <UButton v-if="canEdit" icon="i-lucide-plus" label="Tetapkan" size="xs" color="neutral" variant="ghost" @click="pickerFor = s" />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Siklus hidup -->
    <section class="grid gap-3 md:grid-cols-2">
      <div class="rounded-lg border border-default p-4">
        <h4 class="flex items-center gap-2 text-sm font-medium"><UIcon name="i-lucide-library" class="size-4 text-primary" /> CP Library (template)</h4>
        <p class="mt-1 text-sm text-muted">Disusun dan disahkan di TransCPG-X. Tidak diedit langsung untuk pasien.</p>
      </div>
      <div class="rounded-lg border border-default p-4">
        <h4 class="flex items-center gap-2 text-sm font-medium"><UIcon name="i-lucide-badge-check" class="size-4 text-primary" /> CP Aktif → TransCPR-X</h4>
        <p class="mt-1 text-sm text-muted">Setelah disahkan Direktur, dipakai melayani pasien di TransCPR-X (checklist harian, variance).</p>
      </div>
    </section>

    <UModal
      :open="!!revoking"
      title="Cabut acuan sub-CP?"
      :description="revoking ? `${revoking.title} · ${revoking.sub}` : ''"
      @update:open="(v) => { if (!v) revoking = null }"
    >
      <template #body>
        <p class="text-sm">Pencabutan tercatat di riwayat perubahan CP.</p>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton label="Batal" color="neutral" variant="ghost" @click="revoking = null" />
          <UButton label="Cabut acuan" color="error" :loading="revokeBusy" @click="revoke" />
        </div>
      </template>
    </UModal>

    <PathwayReferencePicker
      v-model:open="pickerOpen"
      :code="code"
      :sub-cp="pickerFor"
      :assigned="pickerFor?.references.map(r => r.document_id) ?? []"
    />
  </div>
</template>
