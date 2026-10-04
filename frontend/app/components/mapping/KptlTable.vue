<script setup lang="ts">
// Daftar prosedur ICD-9-CM dan padanan KPTL-nya (urut episode terbanyak).
import type { KptlRow } from '~/types/api'

defineProps<{ rows: KptlRow[], canManage: boolean }>()
const emit = defineEmits<{ map: [row: KptlRow], revoke: [row: KptlRow] }>()
</script>

<template>
  <div class="table-wrap">
    <table class="data-table">
      <thead>
        <tr>
          <th class="w-24">ICD-9-CM</th>
          <th>Prosedur</th>
          <th class="num w-24">Episode</th>
          <th class="min-w-64">Padanan KPTL</th>
          <th v-if="canManage" class="w-44"><span class="sr-only">Aksi</span></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.icd9cm_code">
          <td class="font-mono">{{ r.icd9cm_code }}</td>
          <td>{{ r.name ?? '—' }}</td>
          <td class="num">{{ formatNumber(r.episodes) }}</td>
          <td>
            <div v-if="r.kptl_code" class="flex items-start gap-1.5">
              <UIcon name="i-lucide-link" class="mt-0.5 size-4 shrink-0 text-success" />
              <div>
                <p><span class="font-mono text-xs">{{ r.kptl_code }}</span> · {{ r.kptl_name }}</p>
                <p class="text-xs text-muted">{{ r.mapped_by_name }} · {{ formatDate(r.mapped_at) }}</p>
              </div>
            </div>
            <UBadge v-else label="Belum dipadankan" icon="i-lucide-unlink" size="sm" color="warning" variant="subtle" />
          </td>
          <td v-if="canManage" class="whitespace-nowrap text-right">
            <UButton
              :label="r.kptl_code ? 'Ganti' : 'Padankan'"
              :icon="r.kptl_code ? 'i-lucide-replace' : 'i-lucide-link'"
              size="xs"
              :color="r.kptl_code ? 'neutral' : 'primary'"
              :variant="r.kptl_code ? 'outline' : 'solid'"
              @click="emit('map', r)"
            />
            <UButton v-if="r.kptl_code" label="Cabut" size="xs" color="error" variant="ghost" class="ml-1" @click="emit('revoke', r)" />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
