<script setup lang="ts">
// Pengaturan › Laporan & Audit — jejak audit gabungan dan laporan berkala.
// Admin RS melihat kejadian di RS-nya; System/Super Admin seluruh RS.
import type { Period } from '~/composables/useAuditApi'

definePageMeta({ title: 'Laporan & Audit', middleware: 'admin' })

const route = useRoute()
const router = useRouter()

type Tab = 'jejak' | 'laporan'
const tab = computed<Tab>(() => (route.query.tab as Tab) ?? 'jejak')
const tabs = [
  { label: 'Jejak audit', value: 'jejak', icon: 'i-lucide-scroll-text' },
  { label: 'Laporan', value: 'laporan', icon: 'i-lucide-file-chart-column' },
]

// Periode disimpan di URL agar tautan bisa dibagikan apa adanya.
const preset = ref<PeriodPreset>((route.query.periode as PeriodPreset) ?? '30')
const period = ref<Period>(preset.value === 'kustom'
  ? { from: route.query.from as string | undefined, to: route.query.to as string | undefined }
  : presetRange(preset.value))
watch(period, (p) => {
  router.replace({
    query: {
      ...route.query,
      periode: preset.value === '30' ? undefined : preset.value,
      from: preset.value === 'kustom' ? p.from : undefined,
      to: preset.value === 'kustom' ? p.to : undefined,
    },
  })
})
</script>

<template>
  <div class="mx-auto max-w-[1400px] space-y-6">
    <header class="bg-brand-soft flex flex-wrap items-end justify-between gap-3 rounded-xl border border-primary/15 px-5 py-4">
      <div>
        <NuxtLink to="/pengaturan" class="mb-1 inline-flex items-center gap-1 text-sm text-muted hover:text-default">
          <UIcon name="i-lucide-arrow-left" class="size-4" /> Pengaturan
        </NuxtLink>
        <h2 class="text-2xl font-semibold text-highlighted">Laporan &amp; Audit</h2>
        <p class="text-sm text-muted">Siapa melakukan apa dan kapan — pengesahan, isi CP, padanan, dokumen, akun, dan akses aplikasi.</p>
      </div>
      <UBadge label="Hanya baca · tidak dapat diubah" icon="i-lucide-lock" color="neutral" variant="outline" />
    </header>

    <div class="flex min-w-0 flex-wrap items-end justify-between gap-3 border-b border-default">
      <UTabs
        :model-value="tab"
        :items="tabs"
        :content="false"
        variant="link"
        :ui="{ list: 'border-0' }"
        @update:model-value="(v) => router.replace({ query: { ...route.query, tab: v === 'jejak' ? undefined : v } })"
      />
      <AuditPeriodPicker v-model="period" v-model:preset="preset" class="min-w-0 pb-2" />
    </div>

    <AuditEventLog v-if="tab === 'jejak'" :period="period" />
    <div v-else class="space-y-10">
      <AuditApprovalReport :period="period" />
      <AuditUsersReport :period="period" />
    </div>
  </div>
</template>
