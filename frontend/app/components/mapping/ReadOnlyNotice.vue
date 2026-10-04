<script setup lang="ts">
// Keterangan untuk peran yang hanya boleh melihat padanan.
defineProps<{ show: boolean }>()
</script>

<template>
  <UAlert
    v-if="show"
    color="neutral"
    variant="subtle"
    icon="i-lucide-eye"
    title="Peran Anda hanya dapat melihat"
    description="Padanan ditetapkan dan dicabut oleh Tim CP atau Admin."
  />
</template>
