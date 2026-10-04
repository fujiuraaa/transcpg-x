<script setup lang="ts">
// Tab "Verifikasi Kode": 7 syarat kelengkapan. Hanya syarat bertanda Wajib
// yang menghambat pengesahan ke Aktif (dipagari server).
import type { Requirement } from '~/types/api'
import type { RequirementCode } from '~/types/domain'

const props = defineProps<{ code: string }>()
const emit = defineEmits<{ openTab: [tab: string] }>()
const api = usePathwayApi(() => props.code)
const { data, status } = await useAsyncData(pathwayKey(props.code, 'readiness'), () => api.readiness())

// Tombol "Lengkapi →": ke tab terkait atau ke meja kerja padanan.
const FIX: Record<RequirementCode, { tab?: string, to?: string, label: string }> = {
  ACUAN_CP: { tab: 'acuan', label: 'Tetapkan acuan' },
  ACUAN_SUB_CP: { tab: 'alur', label: 'Acuan sub-CP' },
  RENCANA_SEVERITY: { tab: 'rencana', label: 'Susun rencana' },
  PADANAN_KPTL: { to: '/padanan/kptl', label: 'Padanan KPTL' },
  TARIF_INACBG: { tab: 'tarif', label: 'Lihat tarif' },
  PADANAN_SNOMED: { to: '/padanan/snomed', label: 'Padanan SNOMED-CT' },
  ICD10_MASTER: { tab: 'diagnosis', label: 'Lihat diagnosis' },
}

const blocking = computed(() => (data.value?.requirements ?? []).filter(r => r.blocking))
const advisory = computed(() => (data.value?.requirements ?? []).filter(r => !r.blocking))
const metCount = computed(() => (data.value?.requirements ?? []).filter(r => r.met).length)

function go(r: Requirement) {
  const f = FIX[r.code]
  if (f.tab) emit('openTab', f.tab)
  else if (f.to) navigateTo(f.to)
}
</script>

<template>
  <div v-if="status === 'pending' && !data" class="space-y-2">
    <USkeleton v-for="i in 4" :key="i" class="h-14 w-full" />
  </div>

  <div v-else-if="data" class="space-y-6">
    <UAlert
      :color="data.ready ? 'success' : 'warning'"
      variant="subtle"
      :icon="data.ready ? 'i-lucide-shield-check' : 'i-lucide-shield-alert'"
      :title="data.ready ? 'Syarat wajib terpenuhi — CP dapat disahkan menjadi Aktif' : 'CP belum dapat disahkan menjadi Aktif'"
      :description="`${metCount} dari ${data.requirements.length} syarat terpenuhi. Syarat anjuran tidak menghambat, tetapi sebaiknya dilengkapi.`"
    />

    <section v-for="group in [{ title: 'Wajib — menghambat pengesahan', items: blocking }, { title: 'Anjuran', items: advisory }]" :key="group.title" class="space-y-2">
      <h3 class="text-xs font-medium uppercase tracking-wide text-muted">{{ group.title }}</h3>
      <ul class="divide-y divide-default rounded-lg border border-default bg-default">
        <li v-for="r in group.items" :key="r.code" class="flex items-start gap-3 p-4">
          <UIcon
            :name="r.met ? 'i-lucide-circle-check' : r.blocking ? 'i-lucide-circle-x' : 'i-lucide-circle-alert'"
            class="mt-0.5 size-5 shrink-0"
            :class="r.met ? 'text-success' : r.blocking ? 'text-error' : 'text-warning'"
          />
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium">
              <span class="text-muted">{{ r.no }}.</span> {{ r.title }}
            </p>
            <p class="text-xs text-muted">Penanggung jawab: {{ r.owner }}</p>
            <p v-if="!r.met" class="mt-1.5 flex flex-wrap gap-1">
              <UBadge v-for="m in r.missing" :key="m" :label="m" size="sm" color="neutral" variant="outline" />
            </p>
          </div>
          <span v-if="r.met" class="text-xs font-medium text-success">Terpenuhi</span>
          <UButton v-else :label="`${FIX[r.code].label} →`" size="xs" color="neutral" variant="outline" @click="go(r)" />
        </li>
      </ul>
    </section>
  </div>
</template>
