<script setup lang="ts">
// §3 Dashboard — gambaran kemajuan pengelolaan CP di rumah sakit.
import type { BarItem } from '~/components/dashboard/HorizontalBars.vue'

definePageMeta({ title: 'Dashboard' })

const { summary } = useDashboardApi()
const { user, approvalBadge, registrationBadge, isAdmin } = useAuth()
const { data, status, error } = await useAsyncData('dashboard', summary)

const greeting = computed(() => {
  const h = new Date().getHours()
  return h < 11 ? 'Selamat pagi' : h < 15 ? 'Selamat siang' : h < 19 ? 'Selamat sore' : 'Selamat malam'
})

const mdcBars = computed<BarItem[]>(() => (data.value?.cases_by_mdc ?? []).map(m => ({
  key: m.mdc_code,
  label: m.mdc_name,
  sublabel: `MDC ${m.mdc_code}`,
  value: m.pct,
  display: `${formatNumber(m.pct, 1)}%`,
  tooltip: `${m.mdc_name}: ${formatNumber(m.episodes)} episode (${formatNumber(m.pct, 1)}%)`,
})))

const coverageBars = computed<BarItem[]>(() => (data.value?.guideline_coverage ?? []).map(c => ({
  key: c.category,
  label: GUIDELINE_CATEGORY[c.category] ?? c.category,
  value: c.items,
  display: formatNumber(c.items),
  tooltip: `${GUIDELINE_CATEGORY[c.category] ?? c.category}: ${formatNumber(c.items)} butir`,
})))
const totalEpisodes = computed(() => (data.value?.cases_by_mdc ?? []).reduce((s, m) => s + m.episodes, 0))
</script>

<template>
  <div class="mx-auto max-w-[1500px] space-y-6">
    <!-- Sapaan & tindakan yang menunggu -->
    <header class="bg-brand-soft flex flex-wrap items-center justify-between gap-4 rounded-xl border border-primary/15 px-5 py-4">
      <div>
        <p class="text-sm text-muted">{{ greeting }},</p>
        <h2 class="text-2xl font-semibold text-highlighted">{{ user?.full_name }}</h2>
        <p class="text-sm text-muted">{{ user?.role_label }} · ringkasan kemajuan penyusunan Clinical Pathway</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton
          v-if="approvalBadge"
          to="/approval"
          :label="`${approvalBadge} CP menunggu tindakan Anda`"
          icon="i-lucide-stamp"
        />
        <UButton
          v-if="isAdmin && registrationBadge"
          to="/pengaturan/users"
          :label="`${registrationBadge} pendaftaran menunggu`"
          icon="i-lucide-user-plus"
          color="neutral"
          variant="outline"
        />
      </div>
    </header>

    <UAlert v-if="error" color="error" variant="subtle" :title="apiErrorMessage(error)" />

    <DashboardStatCards :cards="data?.cards ?? null" />

    <div class="grid gap-6 xl:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
      <section class="rounded-xl border border-default bg-default p-5">
        <header class="mb-4">
          <h3 class="font-semibold">Kasus rawat inap per kelompok diagnosis</h3>
          <p class="text-xs text-muted">Persentase dari {{ formatNumber(totalEpisodes) }} episode, per Major Diagnostic Category (MDC)</p>
        </header>
        <USkeleton v-if="status === 'pending' && !data" class="h-48 w-full" />
        <DashboardHorizontalBars v-else :items="mdcBars" label="Persentase kasus rawat inap per kelompok diagnosis" />
      </section>

      <section class="rounded-xl border border-default bg-default p-5">
        <header class="mb-4">
          <h3 class="font-semibold">Kemajuan penetapan acuan</h3>
          <p class="text-xs text-muted">CP tanpa acuan standar belum bisa disahkan</p>
        </header>
        <USkeleton v-if="!data && status === 'pending'" class="h-48 w-full" />
        <p v-else-if="!data" class="py-10 text-center text-sm text-muted">Data tidak dapat dimuat.</p>
        <DashboardAcuanProgress v-else :progress="data.progress" />
      </section>
    </div>

    <div class="grid gap-6 xl:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
      <section class="rounded-xl border border-default bg-default p-5">
        <header class="mb-4 flex flex-wrap items-end justify-between gap-2">
          <div>
            <h3 class="font-semibold">Prioritas penyusunan CP</h3>
            <p class="text-xs text-muted">Volume kasus terbanyak di atas — panduan Tim CP menentukan CP yang digarap lebih dulu</p>
          </div>
          <UButton to="/library" label="Lihat semua" trailing-icon="i-lucide-arrow-right" size="sm" color="neutral" variant="ghost" />
        </header>
        <USkeleton v-if="!data && status === 'pending'" class="h-48 w-full" />
        <p v-else-if="!data" class="py-10 text-center text-sm text-muted">Data tidak dapat dimuat.</p>
        <DashboardPriorityTable v-else :rows="data.priority" :limit="10" />
      </section>

      <section class="rounded-xl border border-default bg-default p-5">
        <header class="mb-4">
          <h3 class="font-semibold">Cakupan acuan klinis</h3>
          <p class="text-xs text-muted">Jumlah butir panduan yang sudah diekstrak, per jenis isi</p>
        </header>
        <USkeleton v-if="!data && status === 'pending'" class="h-48 w-full" />
        <p v-else-if="!data" class="py-10 text-center text-sm text-muted">Data tidak dapat dimuat.</p>
        <UEmpty v-else-if="!coverageBars.length" icon="i-lucide-list-tree" title="Belum ada butir acuan" variant="naked" />
        <DashboardHorizontalBars v-else :items="coverageBars" label="Jumlah butir acuan klinis per kategori" />
      </section>
    </div>
  </div>
</template>
