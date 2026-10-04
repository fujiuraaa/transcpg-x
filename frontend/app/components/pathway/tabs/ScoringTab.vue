<script setup lang="ts">
// Tab "Scoring Klinis": sistem skoring standar internasional yang relevan.
const props = defineProps<{ code: string }>()
const api = usePathwayApi(() => props.code)
const { data, status } = await useAsyncData(pathwayKey(props.code, 'scoring'), () => api.scoring())
</script>

<template>
  <USkeleton v-if="status === 'pending' && !data" class="h-32 w-full" />
  <UEmpty
    v-else-if="!data?.length"
    icon="i-lucide-calculator"
    title="Belum ada sistem skoring yang dipetakan"
    variant="outline"
  />
  <ul v-else class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
    <li v-for="t in data" :key="t.code" class="rounded-lg border border-default bg-default p-4">
      <div class="flex items-start justify-between gap-2">
        <span class="font-mono text-sm font-semibold">{{ t.code }}</span>
        <UBadge v-if="t.setting" :label="t.setting" size="sm" color="neutral" variant="subtle" />
      </div>
      <p class="mt-1 text-sm">{{ t.name }}</p>
      <p v-if="t.critical_threshold" class="mt-3 flex items-center gap-1.5 text-xs text-muted">
        <UIcon name="i-lucide-triangle-alert" class="size-3.5 text-warning" />
        Ambang kritis: <span class="font-medium text-default">{{ t.critical_threshold }}</span>
      </p>
    </li>
  </ul>
</template>
