<script setup lang="ts">
// Tab "Tarif & Biaya": tarif INA-CBG 3 severity × 3 kelas pada profil RS
// terpilih di bilah atas. Ikut berubah saat profil RS diganti.
const props = defineProps<{ code: string }>()
const api = usePathwayApi(() => props.code)
const { selected } = useHospitalProfile()
const { data, status } = await useAsyncData(
  () => `${pathwayKey(props.code, 'tariffs')}:${selected.value?.id ?? 0}`,
  () => api.tariffs(),
)

const tariff = (sev: number, cls: number) => data.value?.tariffs.find(t => t.severity === sev && t.care_class === cls)?.tariff ?? null
const OWNERSHIP: Record<string, string> = { PEMERINTAH: 'Pemerintah', SWASTA: 'Swasta' }
</script>

<template>
  <USkeleton v-if="status === 'pending' && !data" class="h-48 w-full" />

  <div v-else-if="data" class="space-y-4">
    <div class="flex flex-wrap items-center gap-2 rounded-lg border border-default bg-elevated/50 px-4 py-3 text-sm">
      <UIcon name="i-lucide-hospital" class="size-4 text-muted" />
      <span class="font-medium">{{ data.hospital.name }}</span>
      <span class="text-muted">· Regional {{ data.hospital.bpjs_regional }} · Tipe {{ data.hospital.hospital_type }} · {{ OWNERSHIP[data.hospital.ownership] }}</span>
      <span class="ml-auto text-xs text-muted">Ganti profil RS di bilah atas</span>
    </div>

    <UEmpty
      v-if="!data.tariffs.length"
      icon="i-lucide-wallet"
      title="Tarif tidak tersedia untuk profil RS ini"
      description="Periksa data tarif Permenkes untuk kombinasi regional, tipe RS, dan kepemilikan ini."
      variant="outline"
    />
    <div v-else class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>Severity</th>
            <th v-for="cls in [1, 2, 3]" :key="cls" class="num">Kelas {{ cls }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="sev in [1, 2, 3]" :key="sev">
            <td class="font-medium">
              Severity {{ severityLabel(sev) }}
              <span class="block text-xs font-normal text-muted">{{ sev === 1 ? 'tanpa komplikasi' : sev === 2 ? 'dengan komplikasi (CC)' : 'komplikasi berat (MCC)' }}</span>
            </td>
            <td v-for="cls in [1, 2, 3]" :key="cls" class="num text-base">{{ formatRupiah(tariff(sev, cls)) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <p class="text-xs text-muted">Sumber: {{ data.tariffs[0]?.regulation ?? 'Permenkes 3/2023' }}. Biaya riil rumah sakit tidak ditampilkan — hanya sisi pembayaran INA-CBG.</p>
  </div>
</template>
