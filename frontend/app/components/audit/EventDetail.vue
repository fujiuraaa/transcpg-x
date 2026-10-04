<script setup lang="ts">
// Rincian satu kejadian jejak audit.
import type { AuditEvent } from '~/types/api'

const props = defineProps<{ event: AuditEvent | null }>()
const open = defineModel<boolean>('open', { default: false })

const entries = computed(() => (props.event ? detailEntries(props.event.detail) : []))
const objectLink = computed(() => {
  const e = props.event
  if (!e) return null
  if (e.target_type === 'CP' && e.target_code) return `/library/${e.target_code}`
  if (e.target_type === 'KPTL') return '/padanan/kptl'
  if (e.target_type === 'SNOMED') return '/padanan/snomed'
  if (e.target_type === 'USER') return '/pengaturan/users?tab=pengguna'
  if (e.target_type === 'HOSPITAL') return '/pengaturan/rumah-sakit'
  return null
})
</script>

<template>
  <USlideover v-model:open="open" title="Rincian kejadian" :description="event ? formatDate(event.created_at, true) : ''">
    <template #body>
      <div v-if="event" class="space-y-5">
        <div class="flex items-start gap-3">
          <div class="flex size-9 shrink-0 items-center justify-center rounded-md" :class="isAttention(event.action) ? 'bg-warning/10 text-warning' : 'bg-brand-soft text-primary'">
            <UIcon :name="isAttention(event.action) ? 'i-lucide-triangle-alert' : AUDIT_CATEGORY[event.category].icon" class="size-5" />
          </div>
          <div>
            <p class="font-medium text-highlighted">{{ event.summary }}</p>
            <p class="text-xs text-muted">{{ AUDIT_CATEGORY[event.category].label }} · <span class="font-mono">{{ event.action }}</span></p>
          </div>
        </div>

        <dl class="grid grid-cols-[8rem_minmax(0,1fr)] gap-x-3 gap-y-2 text-sm">
          <dt class="text-muted">Waktu</dt>
          <dd class="tabular-nums">{{ new Date(event.created_at).toLocaleString('id-ID', { dateStyle: 'full', timeStyle: 'medium' }) }}</dd>
          <dt class="text-muted">Pelaku</dt>
          <dd>
            {{ event.actor_name ?? '—' }}
            <span v-if="event.actor_role" class="text-muted"> · {{ event.actor_role_label }}</span>
          </dd>
          <template v-if="event.hospital_name">
            <dt class="text-muted">Rumah sakit</dt>
            <dd>{{ event.hospital_name }}</dd>
          </template>
          <template v-if="event.target_label || event.target_code">
            <dt class="text-muted">Objek</dt>
            <dd>
              <span v-if="event.target_code" class="font-mono text-xs">{{ event.target_code }}</span>
              {{ event.target_label }}
              <ULink v-if="objectLink" :to="objectLink" class="ml-1 inline-flex items-center gap-0.5 text-xs text-primary">
                buka <UIcon name="i-lucide-arrow-up-right" class="size-3" />
              </ULink>
            </dd>
          </template>
          <template v-if="event.ip">
            <dt class="text-muted">Alamat IP</dt>
            <dd class="font-mono text-xs">{{ event.ip }}</dd>
          </template>
        </dl>

        <section v-if="entries.length">
          <h3 class="mb-2 text-xs font-medium uppercase tracking-wide text-muted">Detail</h3>
          <dl class="grid grid-cols-[8rem_minmax(0,1fr)] gap-x-3 gap-y-1.5 rounded-md bg-elevated/60 p-3 text-sm">
            <template v-for="d in entries" :key="d.label">
              <dt class="text-muted first-letter:uppercase">{{ d.label }}</dt>
              <dd class="break-words">{{ d.value }}</dd>
            </template>
          </dl>
        </section>
        <p class="text-xs text-muted">Jejak audit hanya bisa dibaca — tidak dapat diubah atau dihapus dari aplikasi.</p>
      </div>
    </template>
  </USlideover>
</template>
