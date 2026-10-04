<script setup lang="ts">
// Laporan Pengguna & Akses: status akun, kapan terakhir masuk, dan seberapa
// aktif tiap pengguna dalam periode (bahan tinjauan hak akses berkala).
import type { UserReportRow } from '~/types/api'
import type { Period } from '~/composables/useAuditApi'
import type { StatTile } from '~/components/audit/StatTiles.vue'

const props = defineProps<{ period: Period }>()
const { usersReport } = useAuditApi()

const { data, status, error } = await useAsyncData(
  () => `audit:report:users:${props.period.from ?? ''}:${props.period.to ?? ''}`,
  () => usersReport(props.period),
)

const IDLE_DAYS = 90
const idleSince = Date.now() - IDLE_DAYS * 86_400_000
function isIdle(u: UserReportRow) {
  return u.registration_status === 'DISETUJUI' && u.is_active && (!u.last_login_at || new Date(u.last_login_at).getTime() < idleSince)
}

const approved = computed(() => (data.value ?? []).filter(u => u.registration_status === 'DISETUJUI'))
const tiles = computed<StatTile[]>(() => {
  const rows = data.value
  const n = (f: (u: UserReportRow) => boolean) => (rows ? rows.filter(f).length : null)
  const idle = n(isIdle)
  const pending = n(u => u.registration_status === 'MENUNGGU')
  return [
    { label: 'Akun aktif', value: n(u => u.registration_status === 'DISETUJUI' && u.is_active), hint: 'dapat masuk aplikasi', icon: 'i-lucide-user-check' },
    { label: 'Nonaktif', value: n(u => u.registration_status === 'DISETUJUI' && !u.is_active), hint: 'dinonaktifkan Admin', icon: 'i-lucide-user-x' },
    { label: `Tidak masuk ≥ ${IDLE_DAYS} hari`, value: idle, hint: 'akun aktif — tinjau perlu tidaknya', icon: 'i-lucide-clock-alert', attention: !!idle },
    { label: 'Menunggu persetujuan', value: pending, hint: 'pendaftaran mandiri', icon: 'i-lucide-user-plus', attention: !!pending },
  ]
})

type Show = 'semua' | 'aktif' | 'diam' | 'nonaktif'
const show = ref<Show>('semua')
const rows = computed(() => approved.value.filter(u =>
  show.value === 'semua'
  || (show.value === 'aktif' && u.is_active)
  || (show.value === 'nonaktif' && !u.is_active)
  || (show.value === 'diam' && isIdle(u))))

function exportCsv() {
  downloadCsv<UserReportRow>(`pengguna-akses-${todayStamp()}.csv`, [
    { label: 'Nama', value: u => u.full_name },
    { label: 'Email', value: u => u.email },
    { label: 'Peran', value: u => u.role_label },
    { label: 'Rumah sakit', value: u => u.hospital_name ?? '' },
    { label: 'Status', value: u => (u.is_active ? 'Aktif' : 'Nonaktif') },
    { label: 'Terakhir masuk (WIB)', value: u => (u.last_login_at ? wibStamp(u.last_login_at, true) : 'Belum pernah') },
    { label: `Masuk (${periodLabel(props.period.from, props.period.to)})`, value: u => u.login_count },
    { label: 'Aksi tercatat', value: u => u.action_count },
    { label: 'Akun dibuat', value: u => wibStamp(u.created_at) },
  ], rows.value)
}
</script>

<template>
  <section class="space-y-4" aria-labelledby="rep-users">
    <header>
      <h3 id="rep-users" class="text-lg font-semibold text-highlighted">Pengguna &amp; akses</h3>
      <p class="text-sm text-muted">Status akun saat ini; jumlah masuk dan aksi dihitung pada {{ periodLabel(period.from, period.to) }}.</p>
    </header>

    <UAlert v-if="error" color="error" variant="subtle" icon="i-lucide-circle-alert" :title="apiErrorMessage(error)" />
    <template v-else>
      <AuditStatTiles :tiles="tiles" />

      <div class="flex flex-wrap items-center justify-between gap-2">
        <USelect
          v-model="show"
          :items="[
            { label: 'Semua akun', value: 'semua' },
            { label: 'Aktif', value: 'aktif' },
            { label: `Tidak masuk ≥ ${IDLE_DAYS} hari`, value: 'diam' },
            { label: 'Nonaktif', value: 'nonaktif' },
          ]"
          size="sm"
          class="w-56"
          aria-label="Saring akun"
        />
        <UButton label="CSV" icon="i-lucide-download" size="sm" color="neutral" variant="outline" :disabled="!rows.length" @click="exportCsv" />
      </div>

      <USkeleton v-if="status === 'pending' && !data" class="h-48 w-full" />
      <UEmpty v-else-if="!rows.length" icon="i-lucide-users" title="Tidak ada akun" variant="outline" />
      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>Pengguna</th>
              <th>Peran</th>
              <th>Status</th>
              <th>Terakhir masuk</th>
              <th class="num">Masuk</th>
              <th class="num">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in rows" :key="u.id">
              <td>
                <p class="font-medium">{{ u.full_name }}</p>
                <p class="text-xs text-muted">{{ u.email }}</p>
              </td>
              <td class="whitespace-nowrap">{{ u.role_label }}</td>
              <td>
                <UBadge v-if="u.is_active" label="Aktif" color="success" variant="subtle" icon="i-lucide-check" />
                <UBadge v-else label="Nonaktif" color="neutral" variant="subtle" icon="i-lucide-ban" />
              </td>
              <td class="whitespace-nowrap">
                <span v-if="isIdle(u)" class="inline-flex items-center gap-1 text-warning">
                  <UIcon name="i-lucide-clock-alert" class="size-4" />
                  {{ u.last_login_at ? sinceLabel(u.last_login_at) : 'Belum pernah' }}
                </span>
                <span v-else>{{ u.last_login_at ? formatDate(u.last_login_at, true) : 'Belum pernah' }}</span>
              </td>
              <td class="num">{{ formatNumber(u.login_count) }}</td>
              <td class="num">{{ formatNumber(u.action_count) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="text-xs text-muted">
        "Masuk" dihitung sejak pencatatan jejak audit diaktifkan. "Aksi" mencakup pengesahan, isi CP, padanan, dokumen, akun, dan profil RS.
      </p>
    </template>
  </section>
</template>
