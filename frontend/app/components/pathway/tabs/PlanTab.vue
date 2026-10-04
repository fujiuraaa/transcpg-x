<script setup lang="ts">
// Tab "Rencana Klinis" — jantung CP: obat, lab, prosedur, tindakan per hari
// rawat per severity. Wewenang & kunci mengikuti permissions dari server.
import type { PathwayPermissions, PlanItem } from '~/types/api'

const props = defineProps<{ code: string, permissions: PathwayPermissions }>()
const api = usePathwayApi(() => props.code)
const toast = useToast()

const { data: items, status, error } = await useAsyncData(pathwayKey(props.code, 'plan'), () => api.plan())
const { data: summary } = await useAsyncData(pathwayKey(props.code, 'summary'), () => api.summary())

const severity = ref(0)
const canEdit = computed(() => props.permissions.can_edit.rencana)

// Cakupan per severity: wajib bila episode ≥10, dianjurkan bila data tipis.
const MIN_EPISODES = 10
const coverage = computed(() => [1, 2, 3].map((s) => {
  const episodes = summary.value?.severity_stats.find(x => x.severity === s)?.episodes ?? 0
  const rows = (items.value ?? []).filter(i => i.severity === s).length
  const required = episodes >= MIN_EPISODES
  return { severity: s, episodes, rows, required, ok: !required || rows > 0 }
}))

const visible = computed(() => (items.value ?? []).filter(i => !severity.value || i.severity === severity.value))
const byDay = computed(() => {
  const map = new Map<number, PlanItem[]>()
  for (const it of visible.value) {
    if (!map.has(it.day)) map.set(it.day, [])
    map.get(it.day)!.push(it)
  }
  return [...map.entries()].sort((a, b) => a[0] - b[0])
})

// --- Tambah / ubah / hapus ------------------------------------------------------
const modalOpen = ref(false)
const editing = ref<PlanItem | null>(null)
const deleting = ref<PlanItem | null>(null)
const busy = ref(false)

function add() {
  editing.value = null
  modalOpen.value = true
}
function edit(it: PlanItem) {
  editing.value = it
  modalOpen.value = true
}
async function remove() {
  if (!deleting.value) return
  busy.value = true
  try {
    await api.deletePlanItem(deleting.value.id)
    toast.add({ title: 'Baris rencana dihapus', color: 'success' })
    deleting.value = null
    await refreshNuxtData()
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busy.value = false
  }
}

function usage(it: PlanItem) {
  return [it.dose, it.route, it.frequency, it.duration].filter(Boolean).join(' · ')
}
</script>

<template>
  <div class="space-y-4">
    <!-- Cakupan per severity -->
    <div class="grid gap-3 sm:grid-cols-3">
      <button
        v-for="c in coverage"
        :key="c.severity"
        type="button"
        class="rounded-lg border p-3 text-left transition-colors"
        :class="[
          severity === c.severity ? 'ring-2 ring-primary' : '',
          c.ok ? 'border-default bg-default hover:bg-elevated' : 'border-warning/50 bg-warning/5 hover:bg-warning/10',
        ]"
        @click="severity = severity === c.severity ? 0 : c.severity"
      >
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Severity {{ severityLabel(c.severity) }}</span>
          <UBadge
            :label="c.required ? 'Wajib' : 'Dianjurkan'"
            size="sm"
            :color="c.required ? (c.ok ? 'success' : 'warning') : 'neutral'"
            variant="subtle"
          />
        </div>
        <p class="mt-1 text-2xl font-semibold tabular-nums">{{ c.rows }} <span class="text-sm font-normal text-muted">baris</span></p>
        <p class="flex items-center gap-1 text-xs" :class="c.ok ? 'text-muted' : 'text-warning'">
          <UIcon :name="c.ok ? 'i-lucide-check' : 'i-lucide-alert-triangle'" class="size-3.5" />
          {{ c.episodes }} episode{{ !c.ok ? ' — rencana belum ada' : '' }}
        </p>
      </button>
    </div>

    <!-- Bilah alat -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <CommonSeverityFilter v-model="severity" />
      <div class="flex items-center gap-2">
        <span v-if="permissions.content_locked" class="flex items-center gap-1 text-sm text-muted">
          <UIcon name="i-lucide-lock" class="size-4" /> Dikunci selama ditinjau
        </span>
        <span v-else-if="!canEdit" class="text-sm text-muted">Hanya dapat melihat pada tahap ini</span>
        <UButton v-if="canEdit" label="Tambah baris" icon="i-lucide-plus" @click="add" />
      </div>
    </div>

    <!-- Tabel per hari -->
    <div v-if="status === 'pending' && !items" class="space-y-2">
      <USkeleton v-for="i in 4" :key="i" class="h-10 w-full" />
    </div>
    <UAlert v-else-if="error" color="error" variant="subtle" :title="apiErrorMessage(error)" />
    <UEmpty
      v-else-if="!byDay.length"
      icon="i-lucide-clipboard-list"
      title="Belum ada rencana isi klinis"
      :description="severity ? `Belum ada baris untuk severity ${severityLabel(severity)}.` : 'Tambahkan obat, pemeriksaan lab, prosedur, atau tindakan per hari rawat.'"
      :actions="canEdit ? [{ label: 'Tambah baris', icon: 'i-lucide-plus', onClick: add }] : []"
      variant="outline"
    />
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th class="w-14">Hari</th>
            <th v-if="!severity" class="w-14">Sev.</th>
            <th class="min-w-56">Item</th>
            <th class="w-40">Aturan pakai</th>
            <th class="w-28">Sifat</th>
            <th class="w-44">Sumber</th>
            <th v-if="canEdit" class="w-20"><span class="sr-only">Aksi</span></th>
          </tr>
        </thead>
        <tbody>
          <template v-for="[day, rows] in byDay" :key="day">
            <tr v-for="(it, idx) in rows" :key="it.id">
              <td v-if="idx === 0" :rowspan="rows.length" class="font-mono text-sm font-medium">{{ dayLabel(day) }}</td>
              <td v-if="!severity" class="text-muted">{{ severityLabel(it.severity) }}</td>
              <td>
                <div class="flex items-start gap-2">
                  <UIcon :name="PLAN_TYPE[it.item_type].icon" class="mt-0.5 size-4 shrink-0 text-muted" :title="PLAN_TYPE[it.item_type].label" />
                  <div class="min-w-0">
                    <p class="font-medium">{{ it.item_name }}</p>
                    <p class="font-mono text-xs text-muted">{{ PLAN_TYPE[it.item_type].label }} · {{ it.item_code }}</p>
                    <p v-if="it.note" class="mt-0.5 text-xs italic text-muted">{{ it.note }}</p>
                  </div>
                </div>
              </td>
              <td class="text-sm">{{ usage(it) || '—' }}</td>
              <td>
                <UBadge
                  :label="PLAN_NATURE[it.nature]"
                  size="sm"
                  :color="it.nature === 'WAJIB' ? 'primary' : 'neutral'"
                  :variant="it.nature === 'WAJIB' ? 'subtle' : 'outline'"
                />
              </td>
              <td class="text-xs">
                <UBadge v-if="it.outside_guidance" label="Di luar panduan" size="sm" color="warning" variant="subtle" icon="i-lucide-info" />
                <span v-else class="text-muted">{{ it.source_note ?? 'Butir acuan' }}</span>
                <p v-if="it.outside_guidance && it.source_note" class="mt-1 text-muted">{{ it.source_note }}</p>
              </td>
              <td v-if="canEdit" class="whitespace-nowrap text-right">
                <UButton icon="i-lucide-pencil" size="xs" color="neutral" variant="ghost" :aria-label="`Ubah ${it.item_name}`" @click="edit(it)" />
                <UButton icon="i-lucide-trash-2" size="xs" color="error" variant="ghost" :aria-label="`Hapus ${it.item_name}`" @click="deleting = it" />
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <PathwayPlanItemModal v-model:open="modalOpen" :code="code" :item="editing" :default-severity="severity" />

    <UModal
      :open="!!deleting"
      title="Hapus baris rencana?"
      :description="deleting ? `${dayLabel(deleting.day)} · Severity ${severityLabel(deleting.severity)} · ${deleting.item_name}` : ''"
      @update:open="(v) => { if (!v) deleting = null }"
    >
      <template #body>
        <p class="text-sm">Penghapusan tercatat di riwayat perubahan CP.</p>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton label="Batal" color="neutral" variant="ghost" @click="deleting = null" />
          <UButton label="Hapus" color="error" :loading="busy" @click="remove" />
        </div>
      </template>
    </UModal>
  </div>
</template>
