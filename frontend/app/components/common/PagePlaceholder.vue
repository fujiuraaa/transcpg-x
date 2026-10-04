<script setup lang="ts">
// Penanda halaman yang tampilannya belum dirancang. Menampilkan rencana
// komponen (dari dokumen alur) supaya tahap UI/UX punya daftar kerja.
defineProps<{
  title: string
  description?: string
  planned?: string[]
}>()
</script>

<template>
  <section class="space-y-4">
    <div>
      <h2 class="text-xl font-semibold">{{ title }}</h2>
      <p v-if="description" class="text-sm text-muted">{{ description }}</p>
    </div>
    <UAlert
      v-if="planned?.length"
      color="neutral"
      variant="subtle"
      icon="i-lucide-pencil-ruler"
      title="Rencana komponen (UI/UX menyusul)"
    >
      <template #description>
        <ul class="list-disc pl-5">
          <li v-for="p in planned" :key="p">{{ p }}</li>
        </ul>
      </template>
    </UAlert>
    <slot />
  </section>
</template>
