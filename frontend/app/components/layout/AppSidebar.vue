<script setup lang="ts">
// Bilah samping: menu navigasi (bisa dilipat jadi ikon), lencana antrean
// Approval, identitas pengguna, dan tombol Keluar.
// drawer = tampil di laci navigasi layar kecil (lebar penuh, tanpa tombol lipat).
const props = withDefaults(defineProps<{ drawer?: boolean }>(), { drawer: false })
const emit = defineEmits<{ navigate: [] }>()
const collapsedModel = defineModel<boolean>('collapsed', { default: false })
const collapsed = computed(() => !props.drawer && collapsedModel.value)
const { user, isAdmin, approvalBadge, registrationBadge, logout } = useAuth()
const badgeCount = (b?: 'approval' | 'registrations') => (b === 'approval' ? approvalBadge.value : b === 'registrations' ? registrationBadge.value : 0)

const route = useRoute()
const items = computed(() => MENU.filter(m => !m.adminOnly || isAdmin.value))
// Halaman turunan (mis. /library/A-4-13) ikut menyorot menu induknya.
const isActive = (to: string) => route.path === to || route.path.startsWith(`${to}/`)
</script>

<template>
  <aside
    class="flex flex-col bg-elevated"
    :class="drawer ? 'h-full w-full' : ['sticky top-0 h-screen border-r border-default transition-[width] duration-200 ease-out', collapsed ? 'w-16' : 'w-64']"
  >
    <div class="flex h-14 items-center justify-between gap-2 px-3">
      <NuxtLink to="/dashboard" class="flex items-center gap-2" aria-label="TransCPG-X — Dashboard">
        <span class="bg-brand-gradient grid size-8 shrink-0 place-items-center rounded-lg text-xs font-bold text-white shadow-sm">CP</span>
        <span v-if="!collapsed" class="font-semibold tracking-tight">TransCPG-X</span>
      </NuxtLink>
      <UButton
        v-if="!drawer"
        :icon="collapsed ? 'i-lucide-chevrons-right' : 'i-lucide-chevrons-left'"
        color="neutral"
        variant="ghost"
        size="sm"
        :aria-label="collapsed ? 'Buka bilah samping' : 'Lipat bilah samping'"
        @click="collapsedModel = !collapsedModel"
      />
    </div>

    <nav class="flex-1 space-y-1 overflow-y-auto px-2">
      <NuxtLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        class="relative flex items-center gap-3 rounded-md px-2 py-2 text-sm transition-colors duration-150 hover:bg-accented/50"
        :class="isActive(item.to) ? 'bg-brand-soft font-medium text-primary before:absolute before:inset-y-1.5 before:left-0 before:w-0.5 before:rounded-full before:bg-primary' : 'text-toned'"
        :aria-current="isActive(item.to) ? 'page' : undefined"
        :title="collapsed ? item.label : undefined"
        @click="emit('navigate')"
      >
        <UIcon :name="item.icon" class="size-5 shrink-0" />
        <span v-if="!collapsed" class="flex-1 truncate">{{ item.label }}</span>
        <UBadge
          v-if="badgeCount(item.badge) > 0 && !collapsed"
          :label="String(badgeCount(item.badge))"
          color="error"
          size="sm"
        />
      </NuxtLink>
    </nav>

    <div class="flex items-center gap-2 border-t border-default p-3" :class="collapsed && 'flex-col'">
      <NuxtLink
        to="/profil"
        class="flex min-w-0 flex-1 items-center gap-2 rounded-md p-1 transition-colors hover:bg-accented/50"
        :class="isActive('/profil') && 'bg-brand-soft'"
        :title="collapsed ? `${user?.full_name} — Profil Saya` : 'Profil Saya'"
        :aria-current="isActive('/profil') ? 'page' : undefined"
        @click="emit('navigate')"
      >
        <span class="bg-brand-gradient grid size-8 shrink-0 place-items-center rounded-full text-xs font-semibold text-white">{{ initials(user?.full_name) }}</span>
        <span v-if="!collapsed" class="min-w-0 flex-1">
          <span class="block truncate text-sm font-medium">{{ user?.full_name }}</span>
          <span class="block truncate text-xs text-muted">{{ user?.role_label }}</span>
        </span>
      </NuxtLink>
      <UButton icon="i-lucide-log-out" color="neutral" variant="ghost" size="sm" aria-label="Keluar" @click="logout" />
    </div>
  </aside>
</template>
