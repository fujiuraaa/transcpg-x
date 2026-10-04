<script setup lang="ts">
// §13 Padanan KPTL — ICD-9-CM → kode billing BPJS. Peta resmi tidak pernah
// terbit lengkap, jadi padanan ditetapkan manusia (Tim CP / Admin).
import type { KptlRow } from '~/types/api'

definePageMeta({ title: 'Padanan KPTL' })

const route = useRoute()
const router = useRouter()
const { kptl } = useMappingApi()
const toast = useToast()

type Tab = 'belum' | 'semua'
const tab = computed<Tab>(() => (route.query.tab as Tab) ?? 'belum')
const q = ref((route.query.q as string) ?? '')
const debounced = useDebounced(q, 300)
watch(debounced, v => router.replace({ query: { ...route.query, q: v || undefined } }))

const { data: summary, refresh: refreshSummary } = await useAsyncData('kptl:summary', kptl.summary)
const { data, status, error, refresh } = await useAsyncData(
  () => `kptl:${tab.value}:${debounced.value}`,
  () => kptl.worklist(tab.value, debounced.value),
)
const canManage = computed(() => !!summary.value?.can_manage)

const pickerRow = ref<KptlRow | null>(null)
const pickerOpen = ref(false)
function openPicker(r: KptlRow) {
  pickerRow.value = r
  pickerOpen.value = true
}

const revoking = ref<KptlRow | null>(null)
const busy = ref(false)
async function revoke() {
  if (!revoking.value) return
  busy.value = true
  try {
    await kptl.remove(revoking.value.icd9cm_code)
    toast.add({ title: `Padanan ${revoking.value.icd9cm_code} dicabut`, color: 'success' })
    revoking.value = null
    await onSaved()
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busy.value = false
  }
}

async function onSaved() {
  await Promise.all([refresh(), refreshSummary()])
}

const tabs = computed(() => [
  {
    label: 'Belum Dipadankan',
    value: 'belum',
    icon: 'i-lucide-unlink',
    badge: summary.value?.belum ? { label: String(summary.value.belum), color: 'warning' as const, variant: 'solid' as const } : undefined,
  },
  { label: 'Semua Prosedur', value: 'semua', icon: 'i-lucide-list' },
])
</script>

<template>
  <div class="mx-auto max-w-[1400px] space-y-6">
    <header class="bg-brand-soft rounded-xl border border-primary/15 px-5 py-4">
      <h2 class="text-2xl font-semibold text-highlighted">Padanan KPTL</h2>
      <p class="text-sm text-muted">
        Prosedur di CP dicatat dalam ICD-9-CM, sedangkan BPJS menagih dengan kode KPTL. Peta resmi keduanya tidak
        pernah terbit lengkap, sehingga padanannya ditetapkan oleh Tim Koding.
      </p>
    </header>

    <MappingKptlSummary :summary="summary ?? null" />
    <MappingReadOnlyNotice :show="!!summary && !canManage" />

    <div class="space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <UTabs
          :model-value="tab"
          :items="tabs"
          :content="false"
          variant="link"
          :ui="{ list: 'border-b border-default' }"
          @update:model-value="(v) => router.replace({ query: { ...route.query, tab: v === 'belum' ? undefined : v } })"
        />
        <UInput v-model="q" icon="i-lucide-search" placeholder="Cari kode atau nama prosedur" class="w-72" />
      </div>

      <USkeleton v-if="status === 'pending' && !data" class="h-48 w-full" />
      <UAlert v-else-if="error" color="error" variant="subtle" :title="apiErrorMessage(error)" />
      <UEmpty
        v-else-if="!data?.length"
        :icon="tab === 'belum' && !q ? 'i-lucide-party-popper' : 'i-lucide-search-x'"
        :title="tab === 'belum' && !q ? 'Semua prosedur sudah dipadankan' : 'Tidak ada prosedur yang cocok'"
        variant="outline"
      />
      <MappingKptlTable v-else :rows="data" :can-manage="canManage" @map="openPicker" @revoke="revoking = $event" />
    </div>

    <MappingKptlPicker v-model:open="pickerOpen" :row="pickerRow" @saved="onSaved" />

    <UModal
      :open="!!revoking"
      title="Cabut padanan?"
      :description="revoking ? `${revoking.icd9cm_code} ${revoking.name ?? ''} → ${revoking.kptl_code} ${revoking.kptl_name ?? ''}` : ''"
      @update:open="(v) => { if (!v) revoking = null }"
    >
      <template #body>
        <p class="text-sm">Prosedur ini akan kembali berstatus belum dipadankan. Tindakan tercatat di log audit padanan.</p>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton label="Batal" color="neutral" variant="ghost" @click="revoking = null" />
          <UButton label="Cabut padanan" color="error" :loading="busy" @click="revoke" />
        </div>
      </template>
    </UModal>
  </div>
</template>
