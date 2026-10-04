<script setup lang="ts">
// Kartu identitas di Profil Saya. Peran & status hanya bisa diubah Admin RS.
import type { SessionUser } from '~/types/api'

const props = defineProps<{ user: SessionUser }>()
const { hospitals } = useHospitalProfile()
const hospitalName = computed(() => hospitals.value.find(h => h.id === props.user.hospital_id)?.name ?? '—')
</script>

<template>
  <section class="overflow-hidden rounded-xl border border-default bg-default">
    <div class="bg-brand-gradient h-16" aria-hidden="true" />
    <div class="space-y-4 px-5 pb-5">
      <div class="-mt-8">
        <span class="grid size-16 place-items-center rounded-full bg-default text-xl font-semibold text-primary ring-4 ring-(--ui-bg) shadow-sm">
          {{ initials(user.full_name) }}
        </span>
        <p class="mt-2 truncate text-lg font-semibold text-highlighted">{{ user.full_name }}</p>
        <p class="truncate text-sm text-muted">{{ user.email }}</p>
      </div>
      <dl class="grid grid-cols-[7rem_minmax(0,1fr)] gap-x-3 gap-y-2 text-sm">
        <dt class="text-muted">Peran</dt>
        <dd><UBadge :label="user.role_label" color="primary" variant="subtle" /></dd>
        <dt class="text-muted">Rumah sakit</dt>
        <dd>{{ hospitalName }}</dd>
        <dt class="text-muted">Telepon</dt>
        <dd>{{ user.phone ?? '—' }}</dd>
        <dt class="text-muted">Akun sejak</dt>
        <dd>{{ formatDate(user.created_at) }}</dd>
      </dl>
      <p class="flex gap-2 rounded-md bg-elevated/60 p-3 text-xs text-muted">
        <UIcon name="i-lucide-info" class="mt-0.5 size-4 shrink-0" />
        Peran, rumah sakit, dan status akun diatur oleh Admin RS. Hubungi Admin bila ada yang keliru.
      </p>
    </div>
  </section>
</template>
