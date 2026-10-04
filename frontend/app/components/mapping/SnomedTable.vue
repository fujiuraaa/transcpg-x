<script setup lang="ts">
// Daftar kode ICD-10 / ICD-9-CM dan padanan SNOMED-CT-nya.
import type { SnomedRow } from '~/types/api'

defineProps<{ rows: SnomedRow[], canManage: boolean }>()
const emit = defineEmits<{ map: [row: SnomedRow] }>()

const STATUS = {
  OTOMATIS: { label: 'Otomatis', color: 'success', icon: 'i-lucide-bot' },
  MANUAL: { label: 'Manual', color: 'success', icon: 'i-lucide-user-check' },
  RAGU: { label: 'Perlu ditinjau', color: 'warning', icon: 'i-lucide-circle-help' },
} as const
</script>

<template>
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th class="w-24">Vokabuler</th>
          <th class="w-24">Kode</th>
          <th>Nama</th>
          <th class="min-w-72">Konsep SNOMED-CT</th>
          <th class="w-36">Status</th>
          <th v-if="canManage" class="w-28"><span class="sr-only">Aksi</span></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="`${r.vocabulary}-${r.code}`" :class="r.status === 'RAGU' ? 'bg-warning/5' : ''">
          <td class="text-xs text-muted">{{ r.vocabulary === 'ICD10' ? 'ICD-10' : 'ICD-9-CM' }}</td>
          <td class="font-mono">{{ r.code }}</td>
          <td>{{ r.name ?? '—' }}</td>
          <td>
            <template v-if="r.concept_id">
              <p>{{ r.preferred_term }}</p>
              <p class="text-xs text-muted"><span class="font-mono">{{ r.concept_id }}</span> · {{ r.semantic_tag }}</p>
            </template>
            <span v-else class="text-muted">—</span>
          </td>
          <td>
            <UBadge v-if="r.status" :label="STATUS[r.status].label" :icon="STATUS[r.status].icon" :color="STATUS[r.status].color" variant="subtle" size="sm" />
            <UBadge v-else label="Belum dipadankan" icon="i-lucide-circle-dashed" color="neutral" variant="outline" size="sm" />
          </td>
          <td v-if="canManage" class="text-right">
            <UButton
              :label="r.concept_id ? (r.status === 'RAGU' ? 'Tinjau' : 'Ganti') : 'Padankan'"
              size="xs"
              :color="r.concept_id && r.status !== 'RAGU' ? 'neutral' : 'primary'"
              :variant="r.concept_id && r.status !== 'RAGU' ? 'outline' : 'solid'"
              @click="emit('map', r)"
            />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
