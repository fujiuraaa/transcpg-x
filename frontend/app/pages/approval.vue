<script setup lang="ts">
// §10 Kotak Masuk Pengesahan — siapa yang sedang ditunggu, CP mana yang
// sudah selesai. Setuju/kembalikan dilakukan di Halaman Detail CP.
import type { ApprovalTab } from '~/composables/useApprovalApi'
import type { Stage } from '~/types/domain'

definePageMeta({ title: 'Approval' })

const route = useRoute()
const router = useRouter()
const { queue, counts } = useApprovalApi()
const { user } = useAuth()

const tab = computed<ApprovalTab>(() => (route.query.tab as ApprovalTab) ?? 'saya')
const stageFilter = computed<Stage | null>(() => (route.query.tahap as Stage) ?? null)

function go(next: { tab?: ApprovalTab, tahap?: Stage | null }) {
  const q: Record<string, string> = {}
  const t = next.tab ?? tab.value
  if (t !== 'saya') q.tab = t
  const s = next.tahap === undefined ? stageFilter.value : next.tahap
  if (s) q.tahap = s
  router.replace({ query: q })
}

// Kartu tahap: Aktif → tab CP Aktif; tahap lain → Sedang Diproses tersaring.
function selectStage(s: Stage | null) {
  // Melepas kartu Aktif (yang menyala karena tab CP Aktif) → kembali ke kotak masuk.
  if (!s) return go(tab.value === 'aktif' && !stageFilter.value ? { tab: 'saya', tahap: null } : { tahap: null })
  go({ tab: s === 'AKTIF' ? 'aktif' : 'proses', tahap: s === 'AKTIF' ? null : s })
}

const { data: countData } = await useAsyncData('approval:counts', counts)
const { data, status, error } = await useAsyncData(() => `approval:${tab.value}`, () => queue(tab.value))

const rows = computed(() => (data.value ?? []).filter(r => !stageFilter.value || r.stage === stageFilter.value))

const tabs = computed(() => [
  {
    label: 'Menunggu Tindakan Saya',
    value: 'saya',
    icon: 'i-lucide-inbox',
    badge: countData.value?.mine ? { label: String(countData.value.mine), color: 'primary' as const, variant: 'solid' as const } : undefined,
  },
  { label: 'Sedang Diproses', value: 'proses', icon: 'i-lucide-list-todo' },
  { label: 'CP Aktif', value: 'aktif', icon: 'i-lucide-badge-check' },
])

const EMPTY: Record<ApprovalTab, { icon: string, title: string, description: string }> = {
  saya: { icon: 'i-lucide-inbox', title: 'Tidak ada CP yang menunggu tindakan Anda', description: 'Antrean Anda kosong. CP yang perlu ditinjau oleh peran Anda akan muncul di sini.' },
  proses: { icon: 'i-lucide-list-todo', title: 'Tidak ada CP yang sedang diproses', description: 'CP Draf yang belum diajukan tidak muncul di sini — ajukan dulu dari CP Library.' },
  aktif: { icon: 'i-lucide-badge-check', title: 'Belum ada CP Aktif', description: 'CP menjadi Aktif setelah disahkan Direktur.' },
}
</script>

<template>
  <div class="mx-auto max-w-[1400px] space-y-6">
    <header class="bg-brand-soft flex flex-wrap items-end justify-between gap-3 rounded-xl border border-primary/15 px-5 py-4">
      <div>
        <h2 class="text-2xl font-semibold text-highlighted">Approval</h2>
        <p class="text-sm text-muted">Kotak masuk pengesahan — CP mana yang menunggu siapa, dan mana yang sudah selesai.</p>
      </div>
      <p v-if="user" class="flex items-center gap-1.5 text-xs text-muted">
        <UIcon name="i-lucide-info" class="size-3.5" />
        Setuju/kembalikan dilakukan di halaman detail CP, agar isi CP ditinjau dulu.
      </p>
    </header>

    <ApprovalStageCounts :counts="countData?.by_stage ?? null" :active="tab === 'aktif' && !stageFilter ? 'AKTIF' : stageFilter" @select="selectStage" />

    <div class="space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <UTabs
          :model-value="tab"
          :items="tabs"
          :content="false"
          variant="link"
          :ui="{ list: 'border-b border-default', trigger: 'shrink-0' }"
          @update:model-value="(v) => go({ tab: v as ApprovalTab, tahap: null })"
        />
        <UBadge
          v-if="stageFilter"
          :label="`Tahap: ${STAGE_LABEL[stageFilter]}`"
          color="neutral"
          variant="outline"
          trailing-icon="i-lucide-x"
          class="cursor-pointer"
          @click="go({ tahap: null })"
        />
      </div>

      <div v-if="status === 'pending' && !data" class="space-y-2">
        <USkeleton v-for="i in 3" :key="i" class="h-16 w-full" />
      </div>
      <UAlert v-else-if="error" color="error" variant="subtle" icon="i-lucide-circle-alert" :title="apiErrorMessage(error)" />
      <UEmpty
        v-else-if="!rows.length"
        :icon="EMPTY[tab].icon"
        :title="stageFilter ? `Tidak ada CP di tahap ${STAGE_LABEL[stageFilter]}` : EMPTY[tab].title"
        :description="EMPTY[tab].description"
        variant="outline"
      />
      <ApprovalQueueTable v-else :rows="rows" :show-readiness="tab !== 'aktif'" />
    </div>
  </div>
</template>
