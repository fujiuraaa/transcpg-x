<script setup lang="ts">
// Tab "Ringkasan": statistik per severity dan struktur biaya klaim aktual.
const props = defineProps<{ code: string }>()
const api = usePathwayApi(() => props.code)
const { data, status } = await useAsyncData(pathwayKey(props.code, 'summary'), () => api.summary())

const counts = computed(() => Object.fromEntries((data.value?.severity_stats ?? []).map(s => [s.severity, s.episodes])))
const total = computed(() => (data.value?.severity_stats ?? []).reduce((a, s) => a + s.episodes, 0))
const MIN_EPISODES = 10

// Matriks severity × kelas rawat dari distribusi klaim.
const costCell = (sev: number, cls: number) => data.value?.cost_distribution.find(c => c.severity === sev && c.care_class === cls)
</script>

<template>
  <div v-if="status === 'pending' && !data" class="space-y-2">
    <USkeleton class="h-16 w-full" />
    <USkeleton class="h-40 w-full" />
  </div>

  <div v-else-if="data" class="space-y-6">
    <section class="space-y-3 rounded-lg border border-default bg-default p-4">
      <h3 class="text-sm font-medium">Distribusi episode per severity <span class="font-normal text-muted">· {{ formatNumber(total) }} episode</span></h3>
      <CommonSeverityBar :counts="counts" />
    </section>

    <section class="space-y-2">
      <h3 class="text-xs font-medium uppercase tracking-wide text-muted">Statistik per severity</h3>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>Severity</th>
              <th class="num">Episode</th>
              <th class="num">Porsi</th>
              <th class="num">LOS median</th>
              <th class="num">LOS p75</th>
              <th class="num">Target LOS</th>
              <th class="num">Usia median</th>
              <th>Kecukupan data</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in data.severity_stats" :key="s.severity">
              <td class="font-medium">Severity {{ severityLabel(s.severity) }}</td>
              <td class="num">{{ formatNumber(s.episodes) }}</td>
              <td class="num">{{ formatPercent(total ? s.episodes / total : null) }}</td>
              <td class="num">{{ formatNumber(s.los_median, 1) }} hr</td>
              <td class="num">{{ formatNumber(s.los_p75, 1) }} hr</td>
              <td class="num font-medium">{{ formatNumber(s.target_los, 1) }} hr</td>
              <td class="num">{{ formatNumber(s.age_median) }} th</td>
              <td>
                <span v-if="s.episodes >= MIN_EPISODES" class="flex items-center gap-1 text-xs text-success"><UIcon name="i-lucide-check" class="size-3.5" /> Memadai</span>
                <span v-else class="flex items-center gap-1 text-xs text-warning"><UIcon name="i-lucide-alert-triangle" class="size-3.5" /> &lt; {{ MIN_EPISODES }} episode</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="space-y-2">
      <h3 class="text-xs font-medium uppercase tracking-wide text-muted">Struktur klaim per kelas rawat</h3>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>Severity</th>
              <th v-for="cls in [1, 2, 3]" :key="cls" class="num">Kelas {{ cls }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sev in [1, 2, 3]" :key="sev">
              <td class="font-medium">Severity {{ severityLabel(sev) }}</td>
              <td v-for="cls in [1, 2, 3]" :key="cls" class="num">
                <template v-if="costCell(sev, cls)">
                  <span class="block">{{ formatRupiah(costCell(sev, cls)!.avg_claim_tariff) }}</span>
                  <span class="text-xs text-muted">{{ costCell(sev, cls)!.episodes }} episode</span>
                </template>
                <span v-else class="text-muted">—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="text-xs text-muted">Rata-rata tarif klaim INA-CBG. Biaya riil rumah sakit tidak dimuat aplikasi ini.</p>
    </section>
  </div>
</template>
