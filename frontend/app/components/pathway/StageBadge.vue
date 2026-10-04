<script setup lang="ts">
import type { Stage } from '~/types/domain'

defineProps<{ stage: Stage }>()
</script>

<template>
  <UBadge :label="STAGE_LABEL[stage]" :color="STAGE_BADGE[stage].color" :variant="STAGE_BADGE[stage].variant" />
</template>
