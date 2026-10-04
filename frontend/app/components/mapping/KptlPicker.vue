<script setup lang="ts">
// Panel "Padankan": (1) cari di master KPTL, (2) usulan mesin berlabel
// peringatan. Keputusan tetap di tangan Tim Koding.
import type { KptlRow } from '~/types/api'
import type { MasterOption } from '~/components/common/MasterSearch.vue'

const props = defineProps<{ row: KptlRow | null }>()
const open = defineModel<boolean>('open', { default: false })
const emit = defineEmits<{ saved: [] }>()

const { kptl } = useMappingApi()
const toast = useToast()

const picked = ref<MasterOption | undefined>()
const note = ref('')
const busy = ref<string | null>(null)

const { data: sugg, status: suggStatus } = await useAsyncData(
  () => `kptl:suggest:${props.row?.icd9cm_code ?? ''}`,
  () => (props.row ? kptl.suggestions(props.row.icd9cm_code) : Promise.resolve(null)),
)
watch(open, (v) => { if (v) { picked.value = undefined; note.value = '' } })

async function assign(code: string, name: string) {
  if (!props.row || busy.value) return
  busy.value = code
  try {
    await kptl.set(props.row.icd9cm_code, code, note.value.trim() || null)
    toast.add({ title: 'Padanan ditetapkan', description: `${props.row.icd9cm_code} → ${code} · ${name}`, color: 'success' })
    open.value = false
    emit('saved')
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busy.value = null
  }
}
</script>

<template>
  <USlideover
    v-model:open="open"
    :title="row ? `Padankan ${row.icd9cm_code}` : 'Padankan'"
    :description="row?.name ?? undefined"
    :ui="{ content: 'max-w-xl' }"
  >
    <template #body>
      <div v-if="row" class="space-y-6">
        <div v-if="row.kptl_code" class="rounded-lg border border-default bg-elevated/50 px-3 py-2 text-sm">
          <p class="text-xs text-muted">Padanan saat ini</p>
          <p><span class="font-mono text-xs">{{ row.kptl_code }}</span> · {{ row.kptl_name }}</p>
        </div>

        <!-- 1. Cari di master -->
        <section class="space-y-2">
          <h4 class="flex items-center gap-2 text-sm font-semibold">
            <span class="grid size-5 place-items-center rounded-full bg-primary/10 text-xs text-primary">1</span>
            Cari di master KPTL
          </h4>
          <CommonMasterSearch v-model="picked" kind="kptl" placeholder="Ketik nama tindakan, mis. injeksi antibiotik" />
          <div v-if="picked" class="flex items-center justify-between gap-2 rounded-lg border border-primary/40 bg-primary/5 px-3 py-2 text-sm">
            <span class="min-w-0 truncate"><span class="font-mono text-xs">{{ picked.code }}</span> · {{ picked.name }}</span>
            <UButton label="Tetapkan" icon="i-lucide-check" size="sm" :loading="busy === picked.code" :disabled="!!busy && busy !== picked.code" @click="assign(picked.code, picked.name)" />
          </div>
        </section>

        <!-- 2. Usulan mesin -->
        <section class="space-y-2">
          <h4 class="flex items-center gap-2 text-sm font-semibold">
            <span class="grid size-5 place-items-center rounded-full bg-primary/10 text-xs text-primary">2</span>
            Usulan mesin
          </h4>
          <UAlert
            color="warning"
            variant="subtle"
            icon="i-lucide-triangle-alert"
            title="Sering salah — hanya petunjuk"
            description="Berdasarkan kemiripan nama: hanya ±30% tepat di peringkat 1 dan ±48% masuk 10 besar. Periksa kesesuaian tindakan sebelum menetapkan."
          />
          <USkeleton v-if="suggStatus === 'pending'" class="h-24 w-full" />
          <ul v-else class="divide-y divide-default rounded-lg border border-default">
            <li v-for="(s, i) in sugg?.suggestions ?? []" :key="s.code" class="flex items-center gap-3 px-3 py-2">
              <span class="w-5 text-xs tabular-nums text-muted">{{ i + 1 }}</span>
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm"><span class="font-mono text-xs text-muted">{{ s.code }}</span> {{ s.name }}</p>
                <div class="mt-1 flex items-center gap-2">
                  <div class="h-1 w-24 overflow-hidden rounded-full bg-elevated">
                    <div class="h-full rounded-full bg-(--ui-color-neutral-400)" :style="{ width: `${Math.round(s.score * 100)}%` }" />
                  </div>
                  <span class="text-xs tabular-nums text-muted">kemiripan {{ Math.round(s.score * 100) }}%</span>
                </div>
              </div>
              <UButton label="Tetapkan" size="xs" color="neutral" variant="outline" :loading="busy === s.code" :disabled="!!busy && busy !== s.code" @click="assign(s.code, s.name)" />
            </li>
            <li v-if="!sugg?.suggestions.length" class="px-3 py-4 text-center text-sm text-muted">Tidak ada usulan.</li>
          </ul>
        </section>

        <UFormField label="Catatan (opsional)" description="Disimpan bersama padanan, mis. dasar keputusan koding.">
          <UInput v-model="note" class="w-full" placeholder="Mis. sesuai juknis BPJS 2024" />
        </UFormField>
      </div>
    </template>
  </USlideover>
</template>
