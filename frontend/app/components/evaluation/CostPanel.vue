<script setup lang="ts">
// Kendali Biaya: bauran severity sebelum vs sesudah CP aktif (sisi tarif
// INA-CBG), bauran kelas rawat, tarif rata-rata, dan total klaim.
import type { EvaluationDetail } from '~/types/api'

const props = defineProps<{ data: EvaluationDetail['kendali_biaya'], note: string }>()

const before = computed(() => Object.fromEntries(props.data.bauran_severity.map(s => [s.severity, s.sebelum])))
const after = computed(() => Object.fromEntries(props.data.bauran_severity.map(s => [s.severity, s.sesudah])))
const share3 = (m: Record<number, number>) => {
  const t = Object.values(m).reduce((a, b) => a + b, 0)
  return t ? (m[3] ?? 0) / t : 0
}
const shift = computed(() => share3(after.value) - share3(before.value))

const totalEpisodes = computed(() => props.data.bauran_kelas.reduce((s, k) => s + k.episodes, 0))
const totalClaim = computed(() => props.data.bauran_kelas.reduce((s, k) => s + (k.total_claim ?? 0), 0))
</script>

<template>
  <div class="grid gap-4 xl:grid-cols-2">
    <section class="space-y-4 rounded-lg border border-default bg-default p-4">
      <h5 class="text-sm font-medium">Bauran severity</h5>
      <div class="space-y-1">
        <p class="text-xs text-muted">Sebelum CP aktif (pembentuk target)</p>
        <CommonSeverityBar :counts="before" />
      </div>
      <div class="space-y-1">
        <p class="text-xs text-muted">Sesudah CP aktif</p>
        <CommonSeverityBar :counts="after" />
      </div>
      <p class="flex items-center gap-1.5 text-xs" :class="shift >= 0.1 ? 'font-medium text-warning' : 'text-muted'">
        <UIcon :name="shift >= 0.1 ? 'i-lucide-triangle-alert' : 'i-lucide-arrow-right-left'" class="size-3.5" />
        Porsi severity III {{ shift >= 0 ? 'naik' : 'turun' }} {{ formatNumber(Math.abs(shift) * 100, 1) }} poin persen
        <template v-if="shift >= 0.1"> — periksa kesesuaian pengkodean</template>
      </p>
    </section>

    <section class="space-y-3 rounded-lg border border-default bg-default p-4">
      <h5 class="text-sm font-medium">Bauran kelas rawat & nilai klaim (sesudah aktif)</h5>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>Kelas</th>
              <th class="num">Episode</th>
              <th class="num">Tarif rata-rata</th>
              <th class="num">Total klaim</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="k in data.bauran_kelas" :key="k.care_class">
              <td>Kelas {{ k.care_class }}</td>
              <td class="num">{{ formatNumber(k.episodes) }}</td>
              <td class="num">{{ formatRupiah(k.avg_tariff) }}</td>
              <td class="num">{{ formatRupiah(k.total_claim) }}</td>
            </tr>
            <tr v-if="data.bauran_kelas.length" class="font-semibold">
              <td>Total</td>
              <td class="num">{{ formatNumber(totalEpisodes) }}</td>
              <td class="num">{{ formatRupiah(totalEpisodes ? totalClaim / totalEpisodes : null) }}</td>
              <td class="num">{{ formatRupiah(totalClaim) }}</td>
            </tr>
            <tr v-else>
              <td colspan="4" class="text-center text-muted">Belum ada klaim sesudah CP aktif.</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="flex items-start gap-1.5 text-xs text-muted">
        <UIcon name="i-lucide-info" class="mt-0.5 size-3.5 shrink-0" />{{ note }}
      </p>
    </section>
  </div>
</template>
