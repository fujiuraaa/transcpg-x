<script setup lang="ts">
// Bilah atas: judul halaman, pemilih profil RS, dan peran pengguna.
const route = useRoute()
const { user } = useAuth()
const emit = defineEmits<{ openNav: [] }>()

const title = computed(() => {
  const base = route.meta.title ?? 'TransCPG-X'
  return route.params.code ? `${base} › ${route.params.code}` : base
})
</script>

<template>
  <header class="sticky top-0 z-10 flex h-14 items-center gap-3 border-b border-default bg-default/90 px-4 backdrop-blur sm:gap-4 sm:px-6">
    <UButton
      icon="i-lucide-menu"
      color="neutral"
      variant="ghost"
      class="lg:hidden"
      aria-label="Buka menu navigasi"
      @click="emit('openNav')"
    />
    <h1 class="min-w-0 flex-1 truncate font-medium">
      {{ title }}
    </h1>
    <LayoutHospitalSelector />
    <UBadge v-if="user" :label="user.role_label" color="neutral" variant="subtle" class="hidden sm:inline-flex" />
  </header>
</template>
