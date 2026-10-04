<script setup lang="ts">
// Jejak audit gabungan: pengesahan, isi CP, padanan, dokumen panduan,
// akun, profil RS, dan masuk aplikasi. Hanya baca; terbaru di atas.
import type { AuditCategory, AuditEvent, AuditFilter } from '~/types/api'
import type { Period } from '~/composables/useAuditApi'

const props = defineProps<{ period: Period }>()
const { events, exportEvents, actors } = useAuditApi()
const toast = useToast()

const categories = ref<AuditCategory[]>([])
const actor = ref<number | null>(null)
const search = ref('')
const q = useDebounced(search, 300)
const page = ref(1)

const filter = computed<AuditFilter>(() => ({
  ...props.period, category: categories.value, actor: actor.value, q: q.value, page: page.value,
}))
// Saringan berubah → kembali ke halaman 1.
watch(() => [props.period, categories.value, actor.value, q.value], () => { page.value = 1 }, { deep: true })

const { data, status, error } = await useAsyncData(
  () => `audit:events:${JSON.stringify(filter.value)}`,
  () => events(filter.value),
)
const { data: actorList } = await useAsyncData('audit:actors', actors)

const actorItems = computed(() => [
  { label: 'Semua pelaku', value: null },
  ...(actorList.value ?? []).map(a => ({ label: a.full_name, description: a.role_label, value: a.id })),
])

function toggleCategory(c: AuditCategory) {
  categories.value = categories.value.includes(c) ? categories.value.filter(x => x !== c) : [...categories.value, c]
}
const allCount = computed(() => Object.values(data.value?.counts ?? {}).reduce((s, n) => s + (n ?? 0), 0))
const filtered = computed(() => categories.value.length > 0 || actor.value !== null || !!q.value.trim())
function reset() {
  categories.value = []
  actor.value = null
  search.value = ''
}

const selected = ref<AuditEvent | null>(null)
const detailOpen = ref(false)
function openDetail(e: AuditEvent) {
  selected.value = e
  detailOpen.value = true
}

const exporting = ref(false)
async function exportCsv() {
  exporting.value = true
  try {
    await exportEvents(filter.value)
    if ((data.value?.total ?? 0) > 10000) {
      toast.add({ title: 'Hanya 10.000 kejadian terbaru yang diunduh', description: 'Persempit periode atau saringan untuk data lengkap.', color: 'warning' })
    }
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    exporting.value = false
  }
}

function time(iso: string) {
  return new Date(iso).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<template>
  <section class="space-y-4">
    <!-- Saringan: satu baris di atas tabel -->
    <div class="flex flex-wrap items-center gap-2">
      <UInput v-model="search" icon="i-lucide-search" placeholder="Cari aktivitas, objek, atau pelaku" class="min-w-56 flex-1" />
      <USelectMenu
        v-model="actor"
        :items="actorItems"
        value-key="value"
        placeholder="Semua pelaku"
        class="w-56"
        aria-label="Saring pelaku"
      />
      <UButton v-if="filtered" label="Atur ulang" icon="i-lucide-x" color="neutral" variant="ghost" @click="reset" />
      <UButton
        label="Unduh CSV"
        icon="i-lucide-download"
        color="neutral"
        variant="outline"
        :loading="exporting"
        :disabled="!data?.total"
        @click="exportCsv"
      />
    </div>

    <div class="flex flex-wrap gap-1.5" role="group" aria-label="Saring kategori">
      <UButton
        size="xs"
        color="neutral"
        :variant="categories.length === 0 ? 'solid' : 'outline'"
        :aria-pressed="categories.length === 0"
        @click="categories = []"
      >
        Semua <span class="tabular-nums opacity-70">{{ formatNumber(allCount) }}</span>
      </UButton>
      <UButton
        v-for="c in AUDIT_CATEGORIES"
        :key="c"
        size="xs"
        color="neutral"
        :icon="AUDIT_CATEGORY[c].icon"
        :variant="categories.includes(c) ? 'solid' : 'outline'"
        :aria-pressed="categories.includes(c)"
        @click="toggleCategory(c)"
      >
        {{ AUDIT_CATEGORY[c].label }} <span class="tabular-nums opacity-70">{{ formatNumber(data?.counts[c] ?? 0) }}</span>
      </UButton>
    </div>

    <UAlert v-if="error" color="error" variant="subtle" icon="i-lucide-circle-alert" :title="apiErrorMessage(error)" />
    <USkeleton v-else-if="status === 'pending' && !data" class="h-64 w-full" />
    <UEmpty
      v-else-if="!data?.items.length"
      icon="i-lucide-file-search"
      title="Tidak ada kejadian"
      :description="filtered ? 'Tidak ada yang cocok dengan saringan pada periode ini.' : 'Belum ada aktivitas tercatat pada periode ini.'"
      variant="outline"
    />
    <div v-else class="table-wrap" :class="status === 'pending' && 'opacity-60'">
      <table class="data-table">
        <thead>
          <tr>
            <th class="w-32">Waktu</th>
            <th class="w-44">Kategori</th>
            <th class="w-48">Pelaku</th>
            <th>Aktivitas</th>
            <th class="w-10"><span class="sr-only">Rincian</span></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="e in data.items" :key="e.uid" class="cursor-pointer" @click="openDetail(e)">
            <td class="whitespace-nowrap">
              <p>{{ formatDate(e.created_at) }}</p>
              <p class="text-xs tabular-nums text-muted">{{ time(e.created_at) }}</p>
            </td>
            <td>
              <span class="inline-flex items-center gap-1.5 whitespace-nowrap text-toned">
                <UIcon :name="AUDIT_CATEGORY[e.category].icon" class="size-4 shrink-0 text-muted" />
                {{ AUDIT_CATEGORY[e.category].label }}
              </span>
            </td>
            <td>
              <p class="font-medium">{{ e.actor_name ?? '—' }}</p>
              <p class="text-xs text-muted">{{ e.actor_role_label }}</p>
            </td>
            <td>
              <p class="flex items-start gap-1.5" :class="isAttention(e.action) && 'text-warning'">
                <UIcon v-if="isAttention(e.action)" name="i-lucide-triangle-alert" class="mt-0.5 size-4 shrink-0" />
                <span>{{ e.summary }}</span>
              </p>
              <p v-if="e.target_label || e.target_code" class="text-xs text-muted">
                <span v-if="e.target_code" class="font-mono">{{ e.target_code }}</span>
                <template v-if="e.target_code && e.target_label"> · </template>{{ e.target_label }}
              </p>
            </td>
            <td>
              <UButton icon="i-lucide-chevron-right" size="xs" color="neutral" variant="ghost" :aria-label="`Rincian: ${e.summary}`" @click.stop="openDetail(e)" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="data && data.total > data.per_page" class="flex flex-wrap items-center justify-between gap-2">
      <p class="text-xs text-muted">
        {{ formatNumber((data.page - 1) * data.per_page + 1) }}–{{ formatNumber(Math.min(data.page * data.per_page, data.total)) }}
        dari {{ formatNumber(data.total) }} kejadian
      </p>
      <UPagination v-model:page="page" :total="data.total" :items-per-page="data.per_page" size="sm" />
    </div>
    <p v-else-if="data?.total" class="text-xs text-muted">{{ formatNumber(data.total) }} kejadian</p>

    <AuditEventDetail v-model:open="detailOpen" :event="selected" />
  </section>
</template>
