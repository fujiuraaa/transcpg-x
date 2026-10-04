<script setup lang="ts">
// §15 Pengaturan — hanya Admin RS, System Admin, Super Admin.
definePageMeta({ title: 'Pengaturan', middleware: 'admin' })

const { registrationBadge } = useAuth()
const NuxtLink = resolveComponent('NuxtLink')

const CARDS = [
  { title: 'Manajemen User', description: 'Setujui pendaftaran, tetapkan peran, kelola akun.', to: '/pengaturan/users', icon: 'i-lucide-users', badge: true },
  { title: 'Setup Rumah Sakit', description: 'Profil RS: nama, tipe, regional BPJS, kepemilikan.', to: '/pengaturan/rumah-sakit', icon: 'i-lucide-hospital' },
  { title: 'Laporan & Audit', description: 'Jejak audit aktivitas dan laporan pengesahan & akses.', to: '/pengaturan/audit', icon: 'i-lucide-file-search' },
]
</script>

<template>
  <div class="grid gap-4 md:grid-cols-3">
    <component
      :is="card.to ? NuxtLink : 'div'"
      v-for="card in CARDS"
      :key="card.title"
      :to="card.to ?? undefined"
      :class="card.to ? 'hover:ring-primary' : 'opacity-60'"
      class="rounded-lg p-4 ring ring-default"
    >
      <div class="flex items-center justify-between">
        <UIcon :name="card.icon" class="size-6" />
        <UBadge v-if="card.badge && registrationBadge" :label="`${registrationBadge} menunggu`" color="primary" size="sm" />
      </div>
      <p class="mt-2 font-medium">{{ card.title }}</p>
      <p class="text-sm text-muted">{{ card.description }}</p>
    </component>
  </div>
</template>
