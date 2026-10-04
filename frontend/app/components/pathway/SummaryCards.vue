<script setup lang="ts">
// 5 kartu ringkas di Halaman Detail CP.
import type { PathwayDetail } from '~/types/api'

const props = defineProps<{ cards: PathwayDetail['cards'], hospitalName: string }>()

const items = computed(() => [
  { label: 'Kelompok diagnosis', value: props.cards.mdc.name, hint: `MDC ${props.cards.mdc.code}`, icon: 'i-lucide-folder-tree', text: true },
  { label: 'Severity dominan', value: severityLabel(props.cards.dominant_severity), hint: 'terbanyak di data klaim', icon: 'i-lucide-gauge' },
  { label: 'Target LOS', value: props.cards.target_los != null ? `${formatNumber(props.cards.target_los, 1)} hari` : '—', hint: 'severity dominan', icon: 'i-lucide-calendar-days' },
  { label: 'Tarif INA-CBG', value: formatRupiah(props.cards.tariff), hint: `kelas 3 · ${props.hospitalName}`, icon: 'i-lucide-wallet' },
  { label: 'Rujukan panduan', value: `${props.cards.candidate_count} dokumen`, hint: 'kandidat acuan', icon: 'i-lucide-book-open' },
])
</script>

<template>
  <dl class="grid grid-cols-2 gap-3 sm:grid-cols-3 2xl:grid-cols-5">
    <div v-for="c in items" :key="c.label" class="rounded-lg border border-default bg-default p-3">
      <dt class="flex items-center gap-1.5 text-xs text-muted">
        <UIcon :name="c.icon" class="size-3.5" />
        {{ c.label }}
      </dt>
      <dd
        class="mt-1 font-semibold text-highlighted"
        :class="c.text ? 'line-clamp-2 text-sm leading-snug' : 'truncate text-lg tabular-nums'"
        :title="c.value"
      >
        {{ c.value }}
      </dd>
      <dd class="truncate text-xs text-muted" :title="c.hint">{{ c.hint }}</dd>
    </div>
  </dl>
</template>
