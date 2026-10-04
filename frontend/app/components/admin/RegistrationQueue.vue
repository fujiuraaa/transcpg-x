<script setup lang="ts">
// Antrean permintaan pendaftaran mandiri. Admin menetapkan peran final
// (boleh berbeda dari yang diajukan) lalu menyetujui, atau menolak dengan alasan.
import type { SessionUser } from '~/types/api'
import type { Role } from '~/types/domain'

const props = defineProps<{ rows: SessionUser[], roleOptions: { value: Role, label: string }[] }>()
const emit = defineEmits<{ changed: [] }>()

const { approveRegistration, rejectRegistration } = useAdminApi()
const { refreshMe } = useAuth()
const toast = useToast()

// Peran final per pendaftar — bawaan: peran yang diajukan.
const chosen = reactive<Record<number, Role>>({})
watch(() => props.rows, (rows) => {
  for (const r of rows) chosen[r.id] ??= r.role
}, { immediate: true })

const busyId = ref<number | null>(null)
const rejecting = ref<SessionUser | null>(null)
const reason = ref('')

const roleLabel = (r: Role) => props.roleOptions.find(o => o.value === r)?.label ?? r

async function approve(u: SessionUser) {
  busyId.value = u.id
  try {
    const res = await approveRegistration(u.id, chosen[u.id]!)
    toast.add({ title: `${res.full_name} disetujui`, description: `Peran: ${res.role_label}`, color: 'success', icon: 'i-lucide-user-check' })
    await refreshMe()
    emit('changed')
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busyId.value = null
  }
}

async function reject() {
  if (!rejecting.value) return
  busyId.value = rejecting.value.id
  try {
    await rejectRegistration(rejecting.value.id, reason.value.trim())
    toast.add({ title: `Pendaftaran ${rejecting.value.full_name} ditolak`, color: 'neutral' })
    rejecting.value = null
    await refreshMe()
    emit('changed')
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busyId.value = null
  }
}
</script>

<template>
  <ul class="space-y-3">
    <li v-for="u in rows" :key="u.id" class="grid gap-4 rounded-lg border border-default bg-default p-4 lg:grid-cols-[minmax(0,1fr)_16rem_auto] lg:items-center">
      <div class="min-w-0 space-y-1">
        <p class="font-medium text-highlighted">{{ u.full_name }}</p>
        <p class="flex flex-wrap gap-x-3 gap-y-0.5 text-sm text-muted">
          <span class="flex items-center gap-1"><UIcon name="i-lucide-mail" class="size-3.5" />{{ u.email }}</span>
          <span v-if="u.phone" class="flex items-center gap-1"><UIcon name="i-lucide-smartphone" class="size-3.5" />{{ u.phone }}</span>
        </p>
        <p class="flex flex-wrap items-center gap-2 text-xs text-muted">
          <UBadge :label="`Mengajukan: ${roleLabel(u.role)}`" size="sm" color="secondary" variant="subtle" />
          <span v-if="u.registration_note"><UIcon name="i-lucide-building-2" class="inline size-3.5" /> {{ u.registration_note }}</span>
          <span>· mendaftar {{ sinceLabel(u.created_at) === 'hari ini' ? 'hari ini' : `${sinceLabel(u.created_at)} lalu` }}</span>
        </p>
      </div>

      <UFormField label="Peran yang diberikan" size="sm">
        <USelect v-model="chosen[u.id]" :items="roleOptions" class="w-full" />
      </UFormField>

      <div class="flex gap-2 lg:justify-end">
        <UButton label="Setujui" icon="i-lucide-check" :loading="busyId === u.id && !rejecting" @click="approve(u)" />
        <UButton label="Tolak" icon="i-lucide-x" color="neutral" variant="outline" @click="rejecting = u; reason = ''" />
      </div>
    </li>
  </ul>

  <UModal
    :open="!!rejecting"
    title="Tolak pendaftaran?"
    :description="rejecting ? `${rejecting.full_name} · ${rejecting.email}` : ''"
    @update:open="(v) => { if (!v) rejecting = null }"
  >
    <template #body>
      <UFormField label="Alasan" description="Ditampilkan kepada pendaftar saat mencoba masuk. Pendaftar boleh mendaftar ulang." required>
        <UTextarea v-model="reason" :rows="3" autofocus class="w-full" placeholder="Mis. bukan staf RS ini, gunakan email dinas…" />
      </UFormField>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton label="Batal" color="neutral" variant="ghost" @click="rejecting = null" />
        <UButton label="Tolak pendaftaran" color="error" :disabled="!reason.trim()" :loading="busyId === rejecting?.id" @click="reject" />
      </div>
    </template>
  </UModal>
</template>
