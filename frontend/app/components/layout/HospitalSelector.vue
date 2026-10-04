<script setup lang="ts">
// Pemilih profil RS — mengubah tarif INA-CBG di seluruh halaman.
const { hospitals, selected, select } = useHospitalProfile()

const items = computed(() => hospitals.value.map(h => ({ label: h.name, value: h.id })))
const model = computed({
  get: () => selected.value?.id,
  set: (id) => { if (id != null) select(id) },
})
</script>

<template>
  <USelect
    v-if="items.length"
    v-model="model"
    :items="items"
    icon="i-lucide-hospital"
    class="w-40 sm:w-72"
    aria-label="Profil rumah sakit"
  />
</template>
