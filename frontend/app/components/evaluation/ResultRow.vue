<script setup lang="ts">
// Satu CP di daftar Evaluasi. "Rinci" membuka kendali mutu & biaya di tempat.
import type { EvaluationRow } from '~/types/api'

const props = defineProps<{ row: EvaluationRow, open: boolean }>()
const emit = defineEmits<{ toggle: [] }>()

const { detail } = useEvaluationApi()
const { data, status, error } = await useAsyncData(
  () => `evaluation:detail:${props.open ? props.row.cbg_code : 'tutup'}`,
  () => (props.open ? detail(props.row.cbg_code) : Promise.resolve(null)),
)

const isActive = computed(() => props.row.stage === 'AKTIF')
const state = computed(() => {
  if (!isActive.value) return { label: 'Belum bisa dievaluasi', color: 'neutral' as const, icon: 'i-lucide-clock' }
  if (props.row.needs_review) return { label: 'Perlu ditinjau', color: 'warning' as const, icon: 'i-lucide-triangle-alert' }
  if (!props.row.evaluable) return { label: 'Belum ada klaim', color: 'neutral' as const, icon: 'i-lucide-hourglass' }
  return { label: 'Sesuai target', color: 'success' as const, icon: 'i-lucide-circle-check' }
})
</script>

<template>
  <li class="overflow-hidden rounded-lg border bg-default" :class="row.needs_review ? 'border-warning/50' : 'border-default'">
    <div class="grid gap-3 p-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
      <div class="min-w-0 space-y-1.5">
        <p class="flex flex-wrap items-center gap-2 text-xs text-muted">
          <span class="font-mono font-medium text-default">{{ row.cbg_code }}</span>
          <PathwayStageBadge v-if="!isActive" :stage="row.stage" />
          <span v-else>Aktif sejak {{ formatDate(row.activated_at) }} · {{ formatNumber(row.post_episodes) }} episode sesudah aktif</span>
        </p>
        <p class="font-medium text-highlighted">{{ row.name }}</p>
        <ul v-if="row.flags.length" class="space-y-0.5">
          <li v-for="f in row.flags" :key="`${f.code}-${f.severity}`" class="flex items-start gap-1.5 text-xs text-warning">
            <UIcon name="i-lucide-triangle-alert" class="mt-0.5 size-3.5 shrink-0" />
            <span>
              <span class="font-medium">{{ FLAG_LABEL[f.code] ?? f.code }}</span>
              <template v-if="f.severity"> (severity {{ severityLabel(f.severity) }})</template>
              — {{ f.detail }}
            </span>
          </li>
        </ul>
      </div>
      <div class="flex flex-wrap items-center gap-2 md:justify-end">
        <UBadge :label="state.label" :icon="state.icon" :color="state.color" variant="subtle" />
        <UButton
          v-if="isActive && row.evaluable"
          :label="open ? 'Tutup' : 'Rinci'"
          :trailing-icon="open ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
          size="sm"
          color="neutral"
          variant="outline"
          :aria-expanded="open"
          @click="emit('toggle')"
        />
        <UButton :to="`/library/${row.cbg_code}`" label="Buka" size="sm" color="neutral" variant="ghost" trailing-icon="i-lucide-arrow-right" />
      </div>
    </div>

    <!-- Rincian kendali mutu & biaya -->
    <div v-if="open" class="space-y-5 border-t border-default bg-elevated/40 p-4">
      <USkeleton v-if="status === 'pending'" class="h-40 w-full" />
      <UAlert v-else-if="error" color="error" variant="subtle" :title="apiErrorMessage(error)" />
      <template v-else-if="data">
        <section class="space-y-2">
          <h4 class="flex items-center gap-2 text-sm font-semibold"><UIcon name="i-lucide-activity" class="size-4 text-primary" /> Kendali mutu</h4>
          <EvaluationQualityTable :rows="data.kendali_mutu" />
        </section>
        <section class="space-y-2">
          <h4 class="flex items-center gap-2 text-sm font-semibold"><UIcon name="i-lucide-wallet" class="size-4 text-primary" /> Kendali biaya</h4>
          <EvaluationCostPanel :data="data.kendali_biaya" :note="data.catatan" />
        </section>
        <UAlert
          v-if="row.needs_review"
          color="warning"
          variant="subtle"
          icon="i-lucide-undo-2"
          title="CP menyimpang dari target"
          description="KSM/Komite Medik atau Direktur dapat mengembalikan CP ini ke Revisi dari halaman detail CP untuk diperbaiki."
          :actions="[{ label: 'Buka detail CP', to: `/library/${row.cbg_code}`, color: 'warning', variant: 'outline' }]"
        />
      </template>
    </div>
  </li>
</template>
