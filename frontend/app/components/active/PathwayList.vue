<script setup lang="ts">
// Daftar CP Aktif (kiri) dengan pencarian. Pada layar sempit tampil sebagai dropdown.
import type { ActiveListItem } from '~/types/api'

const props = defineProps<{ items: ActiveListItem[], selected: string | null }>()
const emit = defineEmits<{ select: [code: string] }>()

const q = ref('')
const filtered = computed(() => {
  const t = q.value.trim().toLowerCase()
  if (!t) return props.items
  return props.items.filter(i => [i.cbg_code, i.name, i.mdc_name].some(v => v.toLowerCase().includes(t)))
})

const selectItems = computed(() => props.items.map(i => ({ label: `${i.cbg_code} · ${i.name}`, value: i.cbg_code })))
</script>

<template>
  <!-- Layar sempit: dropdown -->
  <USelectMenu
    :model-value="selected ?? undefined"
    :items="selectItems"
    value-key="value"
    icon="i-lucide-badge-check"
    placeholder="Pilih CP Aktif"
    class="w-full lg:hidden"
    @update:model-value="(v) => v && emit('select', String(v))"
  />

  <!-- Layar lebar: daftar -->
  <nav class="hidden overflow-hidden rounded-lg border border-default bg-default lg:block" aria-label="Daftar CP Aktif">
    <div class="border-b border-default p-2">
      <UInput v-model="q" icon="i-lucide-search" placeholder="Cari CP aktif" size="sm" class="w-full" />
    </div>
    <ul class="max-h-[calc(100vh-22rem)] overflow-y-auto">
      <li v-for="i in filtered" :key="i.id">
        <button
          type="button"
          class="relative w-full border-b border-default px-3 py-2.5 text-left transition last:border-b-0 hover:bg-elevated"
          :class="i.cbg_code === selected ? 'bg-brand-soft before:absolute before:inset-y-2 before:left-0 before:w-0.5 before:rounded-full before:bg-primary' : ''"
          :aria-current="i.cbg_code === selected ? 'true' : undefined"
          @click="emit('select', i.cbg_code)"
        >
          <p class="font-mono text-xs text-muted">{{ i.cbg_code }}</p>
          <p class="truncate text-sm font-medium" :class="i.cbg_code === selected ? 'text-primary' : 'text-highlighted'">{{ i.name }}</p>
          <p class="text-xs text-muted">{{ formatNumber(i.episodes) }} episode · aktif {{ formatDate(i.activated_at) }}</p>
        </button>
      </li>
      <li v-if="!filtered.length" class="px-3 py-6 text-center text-sm text-muted">Tidak ada yang cocok.</li>
    </ul>
  </nav>
</template>
