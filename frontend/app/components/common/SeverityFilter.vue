<script setup lang="ts">
// Pemilih severity: Semua · I · II · III (0 = semua).
const model = defineModel<number>({ default: 0 })
withDefaults(defineProps<{ allowAll?: boolean, compact?: boolean }>(), { allowAll: true, compact: false })
</script>

<template>
  <div class="inline-flex rounded-md border border-default p-0.5" role="radiogroup" aria-label="Severity">
    <button
      v-for="opt in (allowAll ? [0, 1, 2, 3] : [1, 2, 3])"
      :key="opt"
      type="button"
      role="radio"
      :aria-checked="model === opt"
      class="whitespace-nowrap rounded px-3 py-1 text-xs font-medium transition-colors"
      :class="model === opt ? 'bg-primary text-inverted' : 'text-muted hover:text-default'"
      @click="model = opt"
    >
      {{ opt === 0 ? 'Semua' : compact ? severityLabel(opt) : `Severity ${severityLabel(opt)}` }}
    </button>
  </div>
</template>
