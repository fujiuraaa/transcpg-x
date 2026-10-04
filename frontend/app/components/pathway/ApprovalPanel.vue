<script setup lang="ts">
// Panel Tahap Pengesahan: status syarat aktif, tombol sesuai wewenang, dan
// riwayat. Server tetap memeriksa ulang wewenang & syarat.
import type { PathwayDetail, Requirement } from '~/types/api'
import type { ApprovalAction } from '~/types/domain'

const props = defineProps<{ code: string, detail: PathwayDetail }>()
const emit = defineEmits<{ openTab: [tab: string] }>()

const api = usePathwayApi(() => props.code)
const { refreshMe } = useAuth()
const toast = useToast()

const stage = computed(() => props.detail.pathway.stage)
const perms = computed(() => props.detail.permissions)

const { data: readiness } = await useAsyncData(pathwayKey(props.code, 'readiness'), () => api.readiness())
const { data: history } = await useAsyncData(pathwayKey(props.code, 'history'), () => api.history())

const unmetBlocking = computed(() => readiness.value?.requirements.filter(r => r.blocking && !r.met) ?? [])
const recent = computed(() => (history.value?.approvals ?? []).slice(0, 4))

// --- Tindakan ------------------------------------------------------------------
const busy = ref(false)
const dialog = ref<{ action: ApprovalAction, label: string, requireReason: boolean } | null>(null)
const reason = ref('')
const unmet = ref<Requirement[]>([])
const historyOpen = ref(false)

const advance = computed(() => perms.value.actions.find(a => a.action === 'MAJUKAN'))
const secondary = computed(() => perms.value.actions.filter(a => a.action !== 'MAJUKAN'))

function start(a: { action: ApprovalAction, label: string, require_reason: boolean }) {
  unmet.value = []
  // Alasan wajib untuk kembalikan/cabut; pengesahan menjadi Aktif dikonfirmasi dulu.
  if (a.require_reason || stage.value === 'MENUNGGU_DIREKTUR') {
    reason.value = ''
    dialog.value = { action: a.action, label: a.label, requireReason: a.require_reason }
    return
  }
  run(a.action)
}

async function run(action: ApprovalAction) {
  busy.value = true
  try {
    const res = await api.transition(action, reason.value.trim())
    toast.add({ title: `Status berubah menjadi ${res.to_label}`, color: 'success', icon: 'i-lucide-circle-check' })
    dialog.value = null
    await Promise.all([refreshNuxtData(), refreshMe()])
  }
  catch (err) {
    const data = (err as { data?: { code?: string, details?: unknown } }).data
    if (data?.code === 'SYARAT_BELUM_TERPENUHI') {
      unmet.value = data.details as Requirement[]
      dialog.value = null
    }
    toast.add({ title: apiErrorMessage(err), color: 'error', icon: 'i-lucide-circle-alert' })
  }
  finally {
    busy.value = false
  }
}

const ACTION_VERB: Record<ApprovalAction, string> = { MAJUKAN: 'Memajukan', KEMBALIKAN: 'Mengembalikan', CABUT: 'Mencabut' }
</script>

<template>
  <section class="rounded-lg border border-default bg-default" aria-labelledby="approval-title">
    <div class="flex items-center justify-between gap-2 border-b border-default px-4 py-3">
      <h3 id="approval-title" class="font-medium text-highlighted">Tahap Pengesahan</h3>
      <PathwayStageBadge :stage="stage" />
    </div>

    <div class="space-y-4 p-4">
      <!-- Syarat aktif -->
      <button
        v-if="stage !== 'AKTIF' && readiness"
        type="button"
        class="flex w-full items-start gap-2 rounded-md border px-3 py-2 text-left text-sm"
        :class="readiness.ready ? 'border-success/30 bg-success/5' : 'border-warning/40 bg-warning/5'"
        @click="emit('openTab', 'verifikasi')"
      >
        <UIcon
          :name="readiness.ready ? 'i-lucide-shield-check' : 'i-lucide-shield-alert'"
          class="mt-0.5 size-4 shrink-0"
          :class="readiness.ready ? 'text-success' : 'text-warning'"
        />
        <span class="flex-1">
          <span class="font-medium">{{ readiness.ready ? 'Syarat aktif lengkap' : `${unmetBlocking.length} syarat wajib belum terpenuhi` }}</span>
          <span class="block text-xs text-muted">
            {{ readiness.ready ? 'CP dapat disahkan menjadi Aktif.' : unmetBlocking.map(r => r.title.split(' (')[0]).join(' · ') }}
          </span>
        </span>
        <UIcon name="i-lucide-chevron-right" class="mt-0.5 size-4 text-muted" />
      </button>

      <!-- Tindakan -->
      <div v-if="perms.actions.length" class="space-y-2">
        <UButton
          v-if="advance"
          :label="advance.label"
          :icon="stage === 'MENUNGGU_DIREKTUR' ? 'i-lucide-stamp' : 'i-lucide-send'"
          block
          :loading="busy && !dialog"
          :disabled="busy"
          @click="start(advance)"
        />
        <UButton
          v-for="a in secondary"
          :key="a.action"
          :label="a.label"
          :icon="a.action === 'CABUT' ? 'i-lucide-ban' : 'i-lucide-undo-2'"
          color="warning"
          variant="outline"
          block
          :disabled="busy"
          @click="start(a)"
        />
      </div>
      <p v-else-if="perms.waiting_on" class="flex items-center gap-2 text-sm text-muted">
        <UIcon name="i-lucide-hourglass" class="size-4" />
        Menunggu tindakan {{ perms.waiting_on }}.
      </p>

      <UAlert v-if="unmet.length" color="error" variant="subtle" icon="i-lucide-circle-x" title="Belum dapat disahkan">
        <template #description>
          <ul class="list-disc space-y-1 pl-4">
            <li v-for="u in unmet" :key="u.code">{{ u.title }} — <span class="text-muted">{{ u.missing.join(', ') }}</span></li>
          </ul>
        </template>
      </UAlert>

      <!-- Riwayat singkat -->
      <div>
        <div class="mb-2 flex items-center justify-between">
          <h4 class="text-xs font-medium uppercase tracking-wide text-muted">Riwayat</h4>
          <UButton v-if="history" label="Lihat semua" size="xs" color="neutral" variant="link" @click="historyOpen = true" />
        </div>
        <p v-if="!recent.length" class="text-sm text-muted">Belum pernah diajukan.</p>
        <ol v-else class="space-y-3">
          <li v-for="h in recent" :key="h.id" class="flex gap-2 text-sm">
            <UIcon
              :name="h.action === 'MAJUKAN' ? 'i-lucide-arrow-right-circle' : 'i-lucide-undo-2'"
              class="mt-0.5 size-4 shrink-0"
              :class="h.action === 'MAJUKAN' ? 'text-primary' : 'text-warning'"
            />
            <div class="min-w-0">
              <p>
                <span class="font-medium">{{ STAGE_LABEL[h.to_stage] }}</span>
                <span class="text-muted"> · {{ h.actor_name }}</span>
              </p>
              <p v-if="h.reason" class="text-xs text-warning">“{{ h.reason }}”</p>
              <p class="text-xs text-muted">{{ formatDate(h.created_at, true) }}</p>
            </div>
          </li>
        </ol>
      </div>
    </div>

    <!-- Dialog alasan / konfirmasi -->
    <UModal
      :open="!!dialog"
      :title="dialog?.label"
      :description="dialog?.requireReason ? 'Alasan akan tercatat di riwayat pengesahan dan terlihat oleh penyusun.' : 'CP akan langsung berlaku dan dapat dipakai di TransCPR-X.'"
      @update:open="(v) => { if (!v) dialog = null }"
    >
      <template #body>
        <UFormField v-if="dialog?.requireReason" label="Alasan" required>
          <UTextarea v-model="reason" :rows="4" autofocus class="w-full" placeholder="Mis. dosis belum sesuai PNPK, rencana severity II belum lengkap…" />
        </UFormField>
        <p v-else class="text-sm">
          {{ detail.pathway.cbg_code }} · {{ detail.pathway.name }} akan disahkan menjadi <strong>CP Aktif</strong>.
        </p>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton label="Batal" color="neutral" variant="ghost" @click="dialog = null" />
          <UButton
            :label="dialog?.label"
            :color="dialog?.requireReason ? 'warning' : 'primary'"
            :disabled="dialog?.requireReason && !reason.trim()"
            :loading="busy"
            @click="dialog && run(dialog.action)"
          />
        </div>
      </template>
    </UModal>

    <!-- Riwayat lengkap -->
    <USlideover v-model:open="historyOpen" title="Riwayat CP" :description="`${detail.pathway.cbg_code} · ${detail.pathway.name}`">
      <template #body>
        <div class="space-y-6">
          <section>
            <h4 class="mb-2 text-sm font-medium">Riwayat pengesahan</h4>
            <ol class="space-y-3">
              <li v-for="h in history?.approvals ?? []" :key="h.id" class="rounded-md border border-default p-3 text-sm">
                <p>
                  <span class="font-medium">{{ ACTION_VERB[h.action] }}</span>
                  {{ STAGE_LABEL[h.from_stage] }} → {{ STAGE_LABEL[h.to_stage] }}
                </p>
                <p class="text-xs text-muted">{{ h.actor_name }} ({{ h.actor_role }}) · {{ formatDate(h.created_at, true) }}</p>
                <p v-if="h.reason" class="mt-1 text-xs">Alasan: {{ h.reason }}</p>
              </li>
            </ol>
          </section>
          <section>
            <h4 class="mb-2 text-sm font-medium">Riwayat perubahan isi</h4>
            <p v-if="!history?.changes.length" class="text-sm text-muted">Belum ada perubahan isi.</p>
            <ol v-else class="space-y-2">
              <li v-for="c in history.changes" :key="c.id" class="flex gap-2 text-sm">
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
    </USlideover>
  </section>
</template>
