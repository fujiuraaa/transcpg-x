<script setup lang="ts">
// Pengaturan › Manajemen User: permintaan pendaftaran mandiri dan daftar akun.
import type { SessionUser } from '~/types/api'
import type { Role } from '~/types/domain'

definePageMeta({ title: 'Manajemen User', middleware: 'admin' })

const route = useRoute()
const router = useRouter()
const { users, registrations, deleteUser } = useAdminApi()
const { load: loadHospitals } = useHospitalProfile()
const toast = useToast()
const { meta } = useDashboardApi()
const { user, refreshMe } = useAuth()

type Tab = 'permintaan' | 'pengguna' | 'ditolak'
const tab = computed<Tab>(() => (route.query.tab as Tab) ?? 'permintaan')

const { data: pending, status: pendingStatus, refresh: refreshPending } = await useAsyncData('admin:registrations', () => registrations('MENUNGGU'))
const { data: active, refresh: refreshUsers } = await useAsyncData('admin:users', users)
const { data: rejected, refresh: refreshRejected } = await useAsyncData('admin:registrations:rejected', () => registrations('DITOLAK'))
const { data: metaData } = await useAsyncData('meta', meta)

// Peran yang boleh diberikan admin ini (server tetap memeriksa ulang).
const roleOptions = computed(() => (metaData.value?.roles ?? []).filter(r => canAssignRole(user.value?.role, r.value)) as { value: Role, label: string }[])

// --- Tab Pengguna: saringan, tambah/ubah, hapus ------------------------------------
const q = ref('')
const roleFilter = ref<Role | 'semua'>('semua')
const statusFilter = ref<'semua' | 'aktif' | 'nonaktif'>('semua')
const filteredUsers = computed(() => {
  const t = q.value.trim().toLowerCase()
  return (active.value ?? []).filter(u =>
    (!t || u.full_name.toLowerCase().includes(t) || u.email.toLowerCase().includes(t) || (u.phone ?? '').includes(t))
    && (roleFilter.value === 'semua' || u.role === roleFilter.value)
    && (statusFilter.value === 'semua' || (statusFilter.value === 'aktif') === u.is_active))
})
const roleFilterItems = computed(() => [
  { label: 'Semua peran', value: 'semua' },
  ...(metaData.value?.roles ?? []).map(r => ({ label: r.label, value: r.value })),
])

onMounted(loadHospitals)
const formOpen = ref(false)
const editing = ref<SessionUser | null>(null)
function openCreate() {
  editing.value = null
  formOpen.value = true
}
function openEdit(u: SessionUser) {
  editing.value = u
  formOpen.value = true
}

const removing = ref<SessionUser | null>(null)
const removeError = ref('')
const busy = ref(false)
function askRemove(u: SessionUser) {
  removeError.value = ''
  removing.value = u
}
async function confirmRemove() {
  if (!removing.value) return
  busy.value = true
  removeError.value = ''
  try {
    await deleteUser(removing.value.id)
    toast.add({ title: `${removing.value.full_name} dihapus`, color: 'success' })
    removing.value = null
    await refreshUsers()
  }
  catch (err) {
    removeError.value = apiErrorMessage(err)
  }
  finally {
    busy.value = false
  }
}
async function deactivateInstead() {
  if (!removing.value) return
  const u = removing.value
  removing.value = null
  openEdit(u)
}

const tabs = computed(() => [
  {
    label: 'Permintaan pendaftaran',
    value: 'permintaan',
    icon: 'i-lucide-user-plus',
    badge: pending.value?.length ? { label: String(pending.value.length), color: 'primary' as const, variant: 'solid' as const } : undefined,
  },
  { label: 'Pengguna', value: 'pengguna', icon: 'i-lucide-users', badge: active.value ? { label: String(active.value.length), color: 'neutral' as const, variant: 'subtle' as const } : undefined },
  { label: 'Ditolak', value: 'ditolak', icon: 'i-lucide-user-x' },
])

// Mengubah akun sendiri (mis. nama) → perbarui juga data diri di sidebar.
async function onSaved() {
  await refreshUsers()
  if (editing.value?.id === user.value?.id) await refreshMe()
}

async function onChanged() {
  await Promise.all([refreshPending(), refreshUsers(), refreshRejected()])
}
</script>

<template>
  <div class="mx-auto max-w-[1200px] space-y-6">
    <header class="bg-brand-soft rounded-xl border border-primary/15 px-5 py-4">
      <NuxtLink to="/pengaturan" class="mb-1 inline-flex items-center gap-1 text-sm text-muted hover:text-default">
        <UIcon name="i-lucide-arrow-left" class="size-4" /> Pengaturan
      </NuxtLink>
      <h2 class="text-2xl font-semibold text-highlighted">Manajemen User</h2>
      <p class="text-sm text-muted">Setujui pendaftaran baru dan tetapkan peran. Peran menentukan siapa yang boleh menyusun, meninjau, dan mengesahkan CP.</p>
    </header>

    <UTabs
      :model-value="tab"
      :items="tabs"
      :content="false"
      variant="link"
      :ui="{ list: 'border-b border-default' }"
      @update:model-value="(v) => router.replace({ query: v === 'permintaan' ? {} : { tab: v } })"
    />

    <template v-if="tab === 'permintaan'">
      <USkeleton v-if="pendingStatus === 'pending' && !pending" class="h-24 w-full" />
      <UEmpty
        v-else-if="!pending?.length"
        icon="i-lucide-inbox"
        title="Tidak ada permintaan pendaftaran"
        description="Pendaftaran baru dari halaman Daftar akan muncul di sini untuk disetujui."
        variant="outline"
      />
      <AdminRegistrationQueue v-else :rows="pending" :role-options="roleOptions" @changed="onChanged" />
    </template>

    <template v-else-if="tab === 'pengguna'">
      <div class="flex flex-wrap items-center gap-2">
        <UInput v-model="q" icon="i-lucide-search" placeholder="Cari nama, email, atau telepon" class="min-w-56 flex-1" />
        <USelect v-model="roleFilter" :items="roleFilterItems" class="w-48" aria-label="Saring peran" />
        <USelect
          v-model="statusFilter"
          :items="[{ label: 'Semua status', value: 'semua' }, { label: 'Aktif', value: 'aktif' }, { label: 'Nonaktif', value: 'nonaktif' }]"
          class="w-40"
          aria-label="Saring status"
        />
        <UButton label="Tambah pengguna" icon="i-lucide-user-plus" @click="openCreate" />
      </div>
      <UEmpty
        v-if="!filteredUsers.length"
        icon="i-lucide-search-x"
        title="Tidak ada pengguna yang cocok"
        variant="outline"
      />
      <AdminUserTable v-else :rows="filteredUsers" manage @edit="openEdit" @remove="askRemove" />
      <p class="text-xs text-muted">{{ filteredUsers.length }} dari {{ active?.length ?? 0 }} pengguna</p>
    </template>

    <template v-else>
      <UEmpty v-if="!rejected?.length" icon="i-lucide-user-x" title="Tidak ada pendaftaran yang ditolak" variant="outline" />
      <AdminUserTable v-else :rows="rejected" rejected />
    </template>

    <AdminUserFormModal v-model:open="formOpen" :user="editing" :role-options="roleOptions" @saved="onSaved" />

    <UModal
      :open="!!removing"
      title="Hapus pengguna?"
      :description="removing ? `${removing.full_name} · ${removing.email}` : ''"
      @update:open="(v) => { if (!v) removing = null }"
    >
      <template #body>
        <p class="text-sm">Akun dihapus permanen dan tidak bisa dipulihkan. Bila pengguna hanya tidak lagi bertugas, sebaiknya <strong>nonaktifkan</strong> akunnya.</p>
        <UAlert v-if="removeError" color="warning" variant="subtle" icon="i-lucide-triangle-alert" :title="removeError" class="mt-3" />
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton label="Batal" color="neutral" variant="ghost" @click="removing = null" />
          <UButton label="Nonaktifkan saja" color="neutral" variant="outline" @click="deactivateInstead" />
          <UButton v-if="!removeError" label="Hapus permanen" color="error" :loading="busy" @click="confirmRemove" />
        </div>
      </template>
    </UModal>
  </div>
</template>
