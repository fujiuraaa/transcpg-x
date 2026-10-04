<script setup lang="ts">
// Panel padanan SNOMED-CT: pencarian manual di master + saringan jenis
// konsep. Tanpa usulan mesin. Konsep pensiun tampil tetapi tidak bisa dipilih.
import type { SnomedRow } from '~/types/api'
import type { MasterOption } from '~/components/common/MasterSearch.vue'

const props = defineProps<{ row: SnomedRow | null }>()
const open = defineModel<boolean>('open', { default: false })
const emit = defineEmits<{ saved: [] }>()

const { snomed } = useMappingApi()
const toast = useToast()

// Bawaan jenis konsep: diagnosis → disorder, prosedur → procedure.
// 'semua' = tanpa saringan (reka Select menolak nilai string kosong).
const tag = ref<string>('semua')
const picked = ref<MasterOption | undefined>()
const busy = ref(false)
watch(open, (v) => {
  if (!v) return
  picked.value = undefined
  tag.value = props.row?.vocabulary === 'ICD9CM' ? 'procedure' : 'disorder'
})

const TAGS = [
  { label: 'Semua jenis', value: 'semua' },
  { label: 'Gangguan (disorder)', value: 'disorder' },
  { label: 'Temuan (finding)', value: 'finding' },
  { label: 'Prosedur (procedure)', value: 'procedure' },
  { label: 'Entitas teramati (observable entity)', value: 'observable entity' },
  { label: 'Struktur tubuh (body structure)', value: 'body structure' },
  { label: 'Zat (substance)', value: 'substance' },
  { label: 'Situasi (situation)', value: 'situation' },
]

async function assign() {
  if (!props.row || !picked.value) return
  busy.value = true
  try {
    await snomed.set(props.row.vocabulary, props.row.code, picked.value.code)
    toast.add({ title: 'Padanan SNOMED-CT disimpan', description: `${props.row.code} → ${picked.value.name}`, color: 'success' })
    open.value = false
    emit('saved')
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    busy.value = false
  }
}
</script>

<template>
  <USlideover
    v-model:open="open"
    :title="row ? `Padankan ${row.code} ke SNOMED-CT` : 'Padankan'"
    :description="row?.name ?? undefined"
    :ui="{ content: 'max-w-xl' }"
  >
    <template #body>
      <div v-if="row" class="space-y-5">
        <div v-if="row.concept_id" class="rounded-lg border px-3 py-2 text-sm" :class="row.status === 'RAGU' ? 'border-warning/50 bg-warning/5' : 'border-default bg-elevated/50'">
          <p class="text-xs" :class="row.status === 'RAGU' ? 'font-medium text-warning' : 'text-muted'">
            {{ row.status === 'RAGU' ? 'Padanan otomatis bertentangan dengan peta resmi — perlu ditinjau' : 'Padanan saat ini' }}
          </p>
          <p>{{ row.preferred_term }} <span class="font-mono text-xs text-muted">{{ row.concept_id }}</span></p>
        </div>

        <UFormField label="Jenis konsep">
          <USelect v-model="tag" :items="TAGS" class="w-full" />
        </UFormField>
        <UFormField label="Cari konsep SNOMED-CT" description="Ketik istilah dalam bahasa Inggris, mis. typhoid fever. Konsep pensiun tidak dapat dipilih.">
          <CommonMasterSearch v-model="picked" kind="snomed" :semantic-tag="tag === 'semua' ? undefined : tag" placeholder="Ketik istilah atau concept ID" />
        </UFormField>

        <div v-if="picked" class="rounded-lg border border-primary/40 bg-primary/5 px-3 py-2 text-sm">
          <p class="text-xs text-muted">Akan ditetapkan</p>
          <p class="font-medium">{{ picked.name }}</p>
          <p class="text-xs text-muted"><span class="font-mono">{{ picked.code }}</span> · {{ picked.extra }}</p>
        </div>
      </div>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton label="Batal" color="neutral" variant="ghost" @click="open = false" />
        <UButton
          :label="row?.status === 'RAGU' && picked?.code === row?.concept_id ? 'Konfirmasi padanan ini' : 'Tetapkan padanan'"
          icon="i-lucide-check"
          :disabled="!picked"
          :loading="busy"
          @click="assign"
        />
      </div>
    </template>
  </USlideover>
</template>
