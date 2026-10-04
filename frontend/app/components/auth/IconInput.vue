<script setup lang="ts">
// Kolom isian bergaya halaman login: ikon di kiri, sudut membulat.
defineOptions({ inheritAttrs: false })
defineProps<{ icon: string, label: string }>()
const model = defineModel<string>({ default: '' })
</script>

<template>
  <label class="relative block">
    <span class="sr-only">{{ label }}</span>
    <UIcon :name="icon" class="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-muted" />
    <input
      v-model="model"
      v-bind="$attrs"
      :placeholder="$attrs.placeholder as string ?? label"
      class="w-full rounded-xl border border-default bg-elevated/60 py-2.5 pr-10 pl-10 text-sm text-highlighted transition outline-none placeholder:text-dimmed focus:border-(--ui-color-primary-300) focus:bg-default focus:ring-3 focus:ring-(--ui-color-primary-200)/60"
    >
    <span class="absolute top-1/2 right-2 -translate-y-1/2">
      <slot name="trailing" />
    </span>
  </label>
</template>
