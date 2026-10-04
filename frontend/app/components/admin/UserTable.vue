<script setup lang="ts">
// Tabel akun (yang sudah disetujui, atau yang ditolak). Bila `manage`,
// tampil aksi Ubah/Hapus — disembunyikan untuk akun yang lebih tinggi wewenangnya.
import type { SessionUser } from '~/types/api'

defineProps<{ rows: SessionUser[], rejected?: boolean, manage?: boolean }>()
const emit = defineEmits<{ edit: [u: SessionUser], remove: [u: SessionUser] }>()
const { user: me } = useAuth()
</script>

<template>
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th>Nama</th>
          <th>Kontak</th>
          <th class="w-44">Peran</th>
          <th class="w-36">{{ rejected ? 'Ditolak' : 'Status' }}</th>
          <th class="w-40">{{ rejected ? 'Alasan' : 'Masuk terakhir' }}</th>
          <th v-if="manage" class="w-24"><span class="sr-only">Aksi</span></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="u in rows" :key="u.id" :class="!rejected && !u.is_active ? 'opacity-60' : ''">
          <td>
            <p class="font-medium">
              {{ u.full_name }}
              <UBadge v-if="u.id === me?.id" label="Anda" size="sm" color="primary" variant="subtle" class="ml-1" />
            </p>
            <p v-if="u.self_registered" class="text-xs text-muted">Mendaftar sendiri</p>
          </td>
          <td class="text-sm">
            <p>{{ u.email }}</p>
            <p class="text-xs text-muted">{{ u.phone ?? 'Tanpa nomor telepon' }}</p>
          </td>
          <td><UBadge :label="u.role_label" :color="isAdminRole(u.role) ? 'primary' : 'neutral'" variant="subtle" size="sm" /></td>
          <td>
            <template v-if="rejected">{{ formatDate(u.reviewed_at) }}</template>
            <UBadge
              v-else
              :label="u.is_active ? 'Aktif' : 'Nonaktif'"
              :icon="u.is_active ? 'i-lucide-circle-check' : 'i-lucide-circle-pause'"
              :color="u.is_active ? 'success' : 'neutral'"
              variant="subtle"
              size="sm"
            />
          </td>
          <td class="text-sm text-muted">
            {{ rejected ? (u.rejection_reason ?? '—') : (u.last_login_at ? formatDate(u.last_login_at, true) : 'Belum pernah') }}
          </td>
          <td v-if="manage" class="whitespace-nowrap text-right">
            <template v-if="canAssignRole(me?.role, u.role)">
              <UButton icon="i-lucide-pencil" size="xs" color="neutral" variant="ghost" :aria-label="`Ubah ${u.full_name}`" @click="emit('edit', u)" />
              <UButton
                v-if="u.id !== me?.id"
                icon="i-lucide-trash-2"
                size="xs"
                color="error"
                variant="ghost"
                :aria-label="`Hapus ${u.full_name}`"
                @click="emit('remove', u)"
              />
            </template>
            <UTooltip v-else text="Akun dengan wewenang lebih tinggi">
              <UIcon name="i-lucide-lock" class="size-4 text-dimmed" />
            </UTooltip>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
