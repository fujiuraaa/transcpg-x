<script setup lang="ts">
// Pratinjau JSON mentah dari API — alat bantu selama UI belum dirancang.
withDefaults(defineProps<{
  data: unknown
  label?: string
  status?: 'idle' | 'pending' | 'success' | 'error'
  error?: unknown
  open?: boolean
}>(), { label: 'Data API', status: 'success', error: undefined, open: false })
</script>

<template>
  <div>
    <p v-if="status === 'pending'" class="text-sm text-muted">Memuat…</p>
    <UAlert v-else-if="error" color="error" variant="subtle" :title="apiErrorMessage(error)" />
    <details v-else class="rounded-md border border-default" :open="open">
      <summary class="cursor-pointer px-3 py-2 text-sm text-muted">{{ label }}</summary>
      <pre class="max-h-96 overflow-auto px-3 pb-3 text-xs">{{ JSON.stringify(data, null, 2) }}</pre>
    </details>
  </div>
</template>
