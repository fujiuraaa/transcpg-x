<script setup lang="ts">
// Prioritas penyusunan CP: grouper diurut dari volume kasus terbanyak.
// Klik baris → Halaman Detail CP.
import type { DashboardSummary } from '~/types/api'

const props = defineProps<{ rows: DashboardSummary['priority'], limit?: number }>()
const shown = computed(() => props.rows.slice(0, props.limit ?? 10))
const top = computed(() => Math.max(1, ...props.rows.map(r => r.episodes)))
</script>

<template>
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th class="w-12">#</th>
          <th>Grouper</th>
          <th class="w-44">Jumlah kasus</th>
          <th class="w-40">Tahap</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="r in shown"
          :key="r.cbg_code"
          class="cursor-pointer"
          tabindex="0"
          @click="navigateTo(`/library/${r.cbg_code}`)"
          @keydown.enter="navigateTo(`/library/${r.cbg_code}`)"
        >
          <td class="font-semibold tabular-nums text-muted">{{ r.rank }}</td>
          <td>
            <p class="flex items-center gap-2">
              <span class="font-mono text-xs text-muted">{{ r.cbg_code }}</span>
              <UIcon
                :name="r.acuan_count ? 'i-lucide-file-check' : 'i-lucide-file-question'"
                class="size-3.5"
                :class="r.acuan_count ? 'text-success' : 'text-warning'"
                :title="r.acuan_count ? 'Acuan sudah ditetapkan' : 'Acuan belum ditetapkan'"
              />
            </p>
            <p class="font-medium hover:text-primary">{{ r.name }}</p>
            <p class="text-xs text-muted">{{ r.mdc_name }}</p>
          </td>
          <td>
            <div class="flex items-center gap-2">
              <div class="h-2 flex-1 overflow-hidden rounded-full bg-elevated">
                <div class="h-full rounded-r-[4px] bg-(--sev-2)" :style="{ width: `${(r.episodes / top) * 100}%` }" />
              </div>
              <span class="w-10 text-right text-sm tabular-nums">{{ formatNumber(r.episodes) }}</span>
            </div>
          </td>
          <td><PathwayStageBadge :stage="r.stage" /></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
