<script setup lang="ts">
// Tab "Acuan & Pengesahan": dokumen acuan, riwayat pengesahan, riwayat perubahan.
import type { ActiveDetail } from '~/types/api'

defineProps<{ data: ActiveDetail['references_approval'] }>()

const ACTION: Record<string, { icon: string, color: string, verb: string }> = {
  MAJUKAN: { icon: 'i-lucide-arrow-right-circle', color: 'text-primary', verb: 'Maju' },
  KEMBALIKAN: { icon: 'i-lucide-undo-2', color: 'text-warning', verb: 'Dikembalikan' },
  CABUT: { icon: 'i-lucide-ban', color: 'text-error', verb: 'Dicabut' },
}
</script>

<template>
  <div class="grid gap-4 lg:grid-cols-2">
    <section class="rounded-lg border border-default bg-default lg:col-span-2">
      <h4 class="flex items-center gap-2 border-b border-default px-4 py-2.5 text-sm font-medium">
        <UIcon name="i-lucide-book-open" class="size-4 text-muted" /> Dokumen acuan standar
      </h4>
      <ul class="divide-y divide-default">
        <li v-for="r in data.references" :key="r.id" class="flex gap-3 px-4 py-3">
          <UIcon name="i-lucide-file-text" class="mt-0.5 size-4 shrink-0 text-primary" />
          <div class="min-w-0">
            <p class="font-medium">{{ r.title }}</p>
            <p class="text-xs text-muted">
              {{ GUIDELINE_SOURCE[r.source_type] }}<template v-if="r.year"> · {{ r.year }}</template>
              <template v-if="r.regulation_ref"> · {{ r.regulation_ref }}</template>
              <template v-if="r.sub_cp_icd10"> · sub-CP {{ r.sub_cp_icd10 }}</template>
            </p>
          </div>
        </li>
        <li v-if="!data.references.length" class="px-4 py-6 text-center text-sm text-muted">Belum ada acuan.</li>
      </ul>
    </section>

    <section class="rounded-lg border border-default bg-default">
      <h4 class="flex items-center gap-2 border-b border-default px-4 py-2.5 text-sm font-medium">
        <UIcon name="i-lucide-stamp" class="size-4 text-muted" /> Riwayat pengesahan
      </h4>
      <ol class="space-y-4 p-4">
        <li v-for="h in data.approvals" :key="h.id" class="flex gap-3 text-sm">
          <UIcon :name="ACTION[h.action]?.icon ?? 'i-lucide-circle'" class="mt-0.5 size-4 shrink-0" :class="ACTION[h.action]?.color" />
          <div>
            <p><span class="font-medium">{{ STAGE_LABEL[h.to_stage] }}</span> <span class="text-muted">· {{ h.actor_name }}</span></p>
            <p v-if="h.reason" class="text-xs text-warning">“{{ h.reason }}”</p>
            <p class="text-xs text-muted">{{ formatDate(h.created_at, true) }}</p>
          </div>
        </li>
      </ol>
    </section>

    <section class="rounded-lg border border-default bg-default">
      <h4 class="flex items-center gap-2 border-b border-default px-4 py-2.5 text-sm font-medium">
        <UIcon name="i-lucide-history" class="size-4 text-muted" /> Riwayat perubahan isi
      </h4>
      <p v-if="!data.changes.length" class="px-4 py-6 text-center text-sm text-muted">Belum ada perubahan isi.</p>
      <ol v-else class="space-y-3 p-4">
        <li v-for="c in data.changes.slice(0, 15)" :key="c.id" class="flex gap-2 text-sm">
          <UBadge :label="c.action" size="sm" variant="subtle" :color="c.action === 'HAPUS' ? 'error' : c.action === 'UBAH' ? 'warning' : 'success'" />
          <div class="min-w-0">
            <p>{{ c.entity.toLowerCase() }} <span class="text-muted">· tahap {{ STAGE_LABEL[c.stage] }}</span></p>
            <p class="text-xs text-muted">{{ c.actor_name }} · {{ formatDate(c.created_at, true) }}</p>
          </div>
        </li>
      </ol>
    </section>
  </div>
</template>
