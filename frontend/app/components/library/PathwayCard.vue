<script setup lang="ts">
// Satu kartu CP di Library: peringkat, identitas, acuan, tanda-tanda
// kesiapan, distribusi severity, dan angka ringkas. Klik → Detail CP.
import type { LibraryItem } from '~/types/api'

const props = defineProps<{ cp: LibraryItem }>()
const MIN_EPISODES = 10

const thinSeverities = computed(() =>
  [1, 2, 3].filter(s => (props.cp.sev_episodes[String(s)] ?? 0) < MIN_EPISODES).map(severityLabel))

const kptlComplete = computed(() => props.cp.procedure_count > 0 && props.cp.procedure_kptl_mapped === props.cp.procedure_count)
</script>

<template>
  <NuxtLink
    :to="`/library/${cp.cbg_code}`"
    class="group grid gap-4 rounded-lg border border-default bg-default p-4 transition hover:border-primary/40 hover:shadow-sm focus-visible:outline-2 focus-visible:outline-primary lg:grid-cols-[3rem_minmax(0,1fr)_12rem_11rem]"
  >
    <!-- Peringkat volume -->
    <div class="flex items-center gap-2 lg:flex-col lg:items-start lg:gap-0">
      <span class="text-xs text-muted">Peringkat</span>
      <span class="text-lg font-semibold tabular-nums text-highlighted">#{{ cp.volume_rank }}</span>
    </div>

    <!-- Identitas, acuan, tanda -->
    <div class="min-w-0 space-y-2">
      <div>
        <p class="flex flex-wrap items-center gap-x-2 text-xs text-muted">
          <span class="font-mono font-medium text-default">{{ cp.cbg_code }}</span>
          <span aria-hidden="true">·</span>
          <span class="truncate">{{ cp.mdc_code }} — {{ cp.mdc_name }}</span>
        </p>
        <h3 class="truncate font-semibold text-highlighted group-hover:text-primary">{{ cp.name }}</h3>
      </div>

      <p class="flex items-start gap-1.5 text-sm">
        <UIcon
          :name="cp.acuan_count ? 'i-lucide-file-check' : 'i-lucide-file-question'"
          class="mt-0.5 size-4 shrink-0"
          :class="cp.acuan_count ? 'text-success' : 'text-warning'"
        />
        <span v-if="cp.acuan_count" class="line-clamp-1" :title="cp.acuan_titles ?? ''">{{ cp.acuan_titles }}</span>
        <span v-else class="text-muted">
          Acuan standar belum ditetapkan · <span class="text-default">{{ cp.candidate_count }} kandidat</span>
        </span>
      </p>

      <div class="flex flex-wrap gap-1.5">
        <PathwayStageBadge :stage="cp.stage" />
        <UBadge
          v-if="cp.split_decision"
          size="sm"
          variant="subtle"
          :color="cp.split_decision === 'PECAH' ? 'warning' : 'neutral'"
          :icon="cp.split_decision === 'PECAH' ? 'i-lucide-split' : 'i-lucide-circle-dot'"
          :label="cp.split_decision === 'PECAH' ? 'Perlu dipecah' : 'Dapat jadi satu CP'"
        />
        <UBadge
          size="sm"
          variant="subtle"
          :color="cp.severities_adequate ? 'success' : 'neutral'"
          :icon="cp.severities_adequate ? 'i-lucide-check' : 'i-lucide-triangle-alert'"
          :label="cp.severities_adequate ? '3 severity memadai' : `Pasien kurang di severity ${thinSeverities.join(', ')}`"
        />
        <UBadge v-if="cp.has_pediatric_cohort" size="sm" variant="subtle" color="neutral" icon="i-lucide-baby" label="Ada kohort anak" />
        <UBadge
          v-if="cp.procedure_count"
          size="sm"
          variant="outline"
          :color="kptlComplete ? 'success' : 'neutral'"
          icon="i-lucide-receipt"
          :label="`KPTL ${cp.procedure_kptl_mapped}/${cp.procedure_count}`"
          :title="`${cp.procedure_kptl_mapped} dari ${cp.procedure_count} prosedur sudah punya padanan KPTL`"
        />
      </div>
    </div>

    <!-- Distribusi severity -->
    <LibrarySeverityBoxes :counts="cp.sev_episodes" :total="cp.episodes" class="self-center" />

    <!-- Angka ringkas -->
    <dl class="grid grid-cols-3 gap-2 self-center text-sm lg:grid-cols-1 lg:gap-1 lg:text-right">
      <div>
        <dt class="text-xs text-muted">Episode rawat inap</dt>
        <dd class="font-semibold tabular-nums">{{ formatNumber(cp.episodes) }}</dd>
      </div>
      <div>
        <dt class="text-xs text-muted">Target LOS</dt>
        <dd class="tabular-nums">{{ cp.target_los != null ? `${formatNumber(cp.target_los, 1)} hari` : '—' }}</dd>
      </div>
      <div>
        <dt class="text-xs text-muted">Tarif dominan (kls 3)</dt>
        <dd class="tabular-nums">{{ formatRupiah(cp.dominant_tariff) }}</dd>
      </div>
    </dl>
  </NuxtLink>
</template>
