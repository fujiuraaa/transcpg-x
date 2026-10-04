<script setup lang="ts">
// Pemilih periode laporan: pintasan (7/30/90 hari, tahun ini, semua) atau
// rentang kustom. Tanggal dibaca server sebagai WIB, keduanya inklusif.
import type { Period } from '~/composables/useAuditApi'

const model = defineModel<Period>({ required: true })
const preset = defineModel<PeriodPreset>('preset', { default: '30' })

const from = ref(model.value.from ?? '')
const to = ref(model.value.to ?? '')

function choose(p: PeriodPreset) {
  preset.value = p
  if (p === 'kustom') {
    from.value = model.value.from ?? ''
    to.value = model.value.to ?? isoDate(new Date())
    return
  }
  model.value = presetRange(p)
}

const invalid = computed(() => !!from.value && !!to.value && from.value > to.value)
watch([from, to], () => {
  if (preset.value !== 'kustom' || invalid.value) return
  model.value = { from: from.value || undefined, to: to.value || undefined }
})
</script>

<template>
  <div class="flex max-w-full flex-wrap items-center gap-2">
    <UFieldGroup class="max-w-full overflow-x-auto">
      <UButton
        v-for="p in PERIOD_PRESETS"
        :key="p.value"
        :label="p.label"
        size="sm"
        color="neutral"
        :variant="preset === p.value ? 'solid' : 'outline'"
        :aria-pressed="preset === p.value"
        @click="choose(p.value)"
      />
    </UFieldGroup>
    <template v-if="preset === 'kustom'">
      <UInput v-model="from" type="date" size="sm" aria-label="Dari tanggal" :max="to || undefined" class="w-40" />
      <span class="text-sm text-muted">s.d.</span>
      <UInput v-model="to" type="date" size="sm" aria-label="Sampai tanggal" :min="from || undefined" class="w-40" />
      <span v-if="invalid" class="text-xs text-error">Tanggal awal melewati tanggal akhir</span>
    </template>
  </div>
</template>
