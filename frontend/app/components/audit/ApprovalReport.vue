<script setup lang="ts">
// Laporan Pengesahan CP: rekap perpindahan tahap dalam periode, lama rata-rata
// di tiap tahap, dan posisi seluruh CP saat ini (bahan akreditasi/rapat komite).
import type { ApprovalReport } from '~/types/api'
import type { Period } from '~/composables/useAuditApi'
import type { BarItem } from '~/components/dashboard/HorizontalBars.vue'
import type { StatTile } from '~/components/audit/StatTiles.vue'
import type { Stage } from '~/types/domain'

const props = defineProps<{ period: Period }>()
const { approvalReport } = useAuditApi()

const { data, status, error } = await useAsyncData(
  () => `audit:report:approval:${props.period.from ?? ''}:${props.period.to ?? ''}`,
  () => approvalReport(props.period),
)

const tiles = computed<StatTile[]>(() => {
  const t = data.value?.totals
  return [
    { label: 'Diajukan', value: t?.diajukan ?? null, hint: 'dari Draf/Revisi ke Review Tim CP', icon: 'i-lucide-send' },
    { label: 'Disahkan', value: t?.disahkan ?? null, hint: 'menjadi CP Aktif', icon: 'i-lucide-badge-check' },
    { label: 'Dikembalikan', value: t?.dikembalikan ?? null, hint: 'untuk revisi, dengan alasan', icon: 'i-lucide-undo-2', attention: !!t?.dikembalikan },
    { label: 'Dicabut', value: t?.dicabut ?? null, hint: 'status Aktif dicabut', icon: 'i-lucide-ban', attention: !!t?.dicabut },
  ]
})

const STAGE_ORDER: Stage[] = ['DRAF', 'REVISI', 'REVIEW_TIM_CP', 'REVIEW_KOMITE', 'MENUNGGU_DIREKTUR']
const durationBars = computed<BarItem[]>(() => {
  const rows = data.value?.stage_durations ?? []
  return STAGE_ORDER.flatMap((s) => {
    const r = rows.find(x => x.stage === s)
    if (!r) return []
    return [{
      key: s,
      label: STAGE_LABEL[s],
      sublabel: `${formatNumber(r.transitions)} perpindahan · terlama ${formatNumber(r.max_days, 1)} hari`,
      value: r.avg_days,
      display: `${formatNumber(r.avg_days, 1)} hari`,
      tooltip: `${STAGE_LABEL[s]}: rata-rata ${formatNumber(r.avg_days, 1)} hari, terlama ${formatNumber(r.max_days, 1)} hari (${r.transitions} perpindahan)`,
    }]
  })
})

const stageFilter = ref<Stage | 'semua'>('semua')
const stageCounts = computed(() => {
  const m = new Map<Stage, number>()
  for (const p of data.value?.pathways ?? []) m.set(p.stage, (m.get(p.stage) ?? 0) + 1)
  return m
})
const pathways = computed(() => (data.value?.pathways ?? []).filter(p => stageFilter.value === 'semua' || p.stage === stageFilter.value))
const stageItems = computed(() => [
  { label: `Semua tahap (${data.value?.pathways.length ?? 0})`, value: 'semua' },
  ...(['AKTIF', 'MENUNGGU_DIREKTUR', 'REVIEW_KOMITE', 'REVIEW_TIM_CP', 'REVISI', 'DRAF'] as Stage[])
    .filter(s => stageCounts.value.has(s))
    .map(s => ({ label: `${STAGE_LABEL[s]} (${stageCounts.value.get(s)})`, value: s })),
])

function exportCsv() {
  downloadCsv<ApprovalReport['pathways'][number]>(`posisi-cp-${todayStamp()}.csv`, [
    { label: 'Kode', value: p => p.code },
    { label: 'Nama CP', value: p => p.name },
    { label: 'Tahap', value: p => STAGE_LABEL[p.stage] },
    { label: 'Di tahap ini sejak', value: p => wibStamp(p.stage_changed_at) },
    { label: 'Tanggal aktif', value: p => wibStamp(p.activated_at) },
    { label: 'Disahkan oleh', value: p => p.approved_by ?? '' },
    { label: 'Jumlah dikembalikan', value: p => p.returned_count },
  ], pathways.value)
}
</script>

<template>
  <section class="space-y-4" aria-labelledby="rep-approval">
    <header class="flex flex-wrap items-end justify-between gap-2">
      <div>
        <h3 id="rep-approval" class="text-lg font-semibold text-highlighted">Pengesahan CP</h3>
        <p class="text-sm text-muted">Perpindahan tahap pada {{ periodLabel(period.from, period.to) }}. CP berlaku untuk semua RS.</p>
      </div>
    </header>

    <UAlert v-if="error" color="error" variant="subtle" icon="i-lucide-circle-alert" :title="apiErrorMessage(error)" />
    <template v-else>
      <AuditStatTiles :tiles="tiles" />

      <div class="grid gap-4 2xl:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
        <figure class="rounded-lg border border-default bg-default p-4">
          <figcaption class="mb-3">
            <p class="font-medium">Rata-rata lama di tiap tahap</p>
            <p class="text-xs text-muted">Hari sejak masuk tahap sampai diteruskan/dikembalikan, untuk perpindahan dalam periode.</p>
          </figcaption>
          <USkeleton v-if="status === 'pending' && !data" class="h-40 w-full" />
          <p v-else-if="!durationBars.length" class="py-6 text-center text-sm text-muted">Belum ada perpindahan tahap pada periode ini.</p>
          <DashboardHorizontalBars v-else :items="durationBars" label="Rata-rata lama di tiap tahap pengesahan" keep-order />
        </figure>

        <div class="space-y-2">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <p class="font-medium">Posisi CP saat ini</p>
            <div class="flex gap-2">
              <USelect v-model="stageFilter" :items="stageItems" size="sm" class="w-56" aria-label="Saring tahap" />
              <UButton label="CSV" icon="i-lucide-download" size="sm" color="neutral" variant="outline" :disabled="!pathways.length" @click="exportCsv" />
            </div>
          </div>
          <div class="table-wrap max-h-80 overflow-y-auto">
            <table class="data-table">
              <thead>
                <tr>
                  <th>CP</th>
                  <th>Tahap</th>
                  <th>Sejak</th>
                  <th class="num">Dikembalikan</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="p in pathways" :key="p.code">
                  <td>
                    <NuxtLink :to="`/library/${p.code}`" class="font-medium hover:text-primary">{{ p.name }}</NuxtLink>
                    <p class="font-mono text-xs text-muted">{{ p.code }}</p>
                  </td>
                  <td><PathwayStageBadge :stage="p.stage" /></td>
                  <td class="whitespace-nowrap">
                    {{ formatDate(p.stage_changed_at) }}
                    <p v-if="p.approved_by" class="text-xs text-muted">oleh {{ p.approved_by }}</p>
                  </td>
                  <td class="num">{{ p.returned_count || '—' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>
  </section>
</template>
