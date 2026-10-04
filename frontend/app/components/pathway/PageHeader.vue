<script setup lang="ts">
// Kepala Halaman Detail CP: identitas grouper, ciri data, tahap, pratinjau.
import type { PathwayDetail } from '~/types/api'

const props = defineProps<{ detail: PathwayDetail }>()
const p = computed(() => props.detail.pathway)
</script>

<template>
  <header class="space-y-3">
    <NuxtLink to="/library" class="inline-flex items-center gap-1 text-sm text-muted hover:text-default">
      <UIcon name="i-lucide-arrow-left" class="size-4" />
      CP Library
    </NuxtLink>

    <div class="flex flex-wrap items-start justify-between gap-4">
      <div class="min-w-0 space-y-1">
        <div class="flex flex-wrap items-center gap-2 text-sm text-muted">
          <span class="font-mono font-medium text-default">{{ p.cbg_code }}</span>
          <span aria-hidden="true">·</span>
          <span>{{ p.mdc_code }} — {{ p.mdc_name }}</span>
        </div>
        <h2 class="text-2xl font-semibold text-highlighted">{{ p.name }}</h2>
        <div class="flex flex-wrap items-center gap-2 pt-1">
          <UBadge color="neutral" variant="outline" icon="i-lucide-bed-double" :label="`${formatNumber(p.episodes)} episode rawat inap`" />
          <UBadge
            v-if="p.split_decision"
            :color="p.split_decision === 'PECAH' ? 'warning' : 'neutral'"
            variant="subtle"
            :icon="p.split_decision === 'PECAH' ? 'i-lucide-split' : 'i-lucide-circle-dot'"
            :label="p.split_decision === 'PECAH' ? 'Perlu dipecah' : 'Dapat jadi satu CP'"
          />
          <UBadge v-if="p.has_pediatric_cohort" color="neutral" variant="subtle" icon="i-lucide-baby" label="Ada kohort anak" />
        </div>
      </div>

      <div class="flex items-center gap-2">
        <UButton
          v-if="p.stage === 'AKTIF'"
          :to="{ path: '/cp-aktif', query: { code: p.cbg_code } }"
          icon="i-lucide-eye"
          label="Pratinjau CP Aktif"
          color="neutral"
          variant="outline"
        />
      </div>
    </div>

    <PathwayStageTrack :stage="p.stage" :since="p.stage_changed_at" />
  </header>
</template>
