<script setup lang="ts">
// Formulir tambah/ubah satu baris rencana isi klinis.
// Urutan isian mengikuti panduan §9: Jenis → Severity → Hari → Sifat → Item.
import type { GuidelineItemRow, PlanItem, PlanItemInput } from '~/types/api'
import type { PlanItemType, PlanNature, Severity } from '~/types/domain'
import type { MasterOption } from '~/components/common/MasterSearch.vue'

const props = defineProps<{ code: string, item?: PlanItem | null, defaultSeverity?: number }>()
const open = defineModel<boolean>('open', { default: false })

const api = usePathwayApi(() => props.code)
const guidelines = useGuidelineApi()
const toast = useToast()

// --- Butir acuan dari dokumen yang sudah ditetapkan ------------------------------
const { data: references } = await useAsyncData(pathwayKey(props.code, 'references'), () => api.references())
const { data: refItems } = await useAsyncData(
  () => `${pathwayKey(props.code, 'reference-items')}:${(references.value ?? []).map(r => r.document_id).join(',')}`,
  async () => {
    const docs = [...new Map((references.value ?? []).map(r => [r.document_id, r.title])).entries()]
    const lists = await Promise.all(docs.map(async ([id, title]) =>
      (await guidelines.items(id)).map(i => ({ ...i, docTitle: title }))))
    return lists.flat()
  },
)
type RefItem = GuidelineItemRow & { docTitle: string }
const guidelineOptions = computed(() => [
  { label: 'Tidak dari panduan (DI LUAR PANDUAN)', value: 0 },
  ...(refItems.value ?? []).map((i: RefItem) => ({
    label: `${i.docTitle} · ${GUIDELINE_CATEGORY[i.category]} · ${i.title} (hlm. ${i.page})`,
    value: i.id,
  })),
])

// --- State formulir --------------------------------------------------------------
const blank = () => ({
  item_type: 'OBAT' as PlanItemType,
  severity: (props.defaultSeverity || 1) as Severity,
  day: 0,
  nature: 'WAJIB' as PlanNature,
  master: undefined as MasterOption | undefined,
  dose: '', route: '', frequency: '', duration: '',
  guideline_item_id: 0,
  source_note: '', note: '',
})
const form = reactive(blank())
const error = ref('')
const busy = ref(false)

watch(open, (v) => {
  if (!v) return
  error.value = ''
  Object.assign(form, blank())
  const it = props.item
  if (it) {
    Object.assign(form, {
      item_type: it.item_type, severity: it.severity, day: it.day, nature: it.nature,
      dose: it.dose ?? '', route: it.route ?? '', frequency: it.frequency ?? '', duration: it.duration ?? '',
      guideline_item_id: it.guideline_item_id ?? 0, source_note: it.source_note ?? '', note: it.note ?? '',
    })
    // Set setelah item_type agar watcher MasterSearch tidak mengosongkannya.
    nextTick(() => { form.master = { label: `${it.item_code} · ${it.item_name}`, code: it.item_code, name: it.item_name } })
  }
}, { immediate: true })

// Pilih butir acuan → isi sumber dokumen otomatis bila masih kosong.
watch(() => form.guideline_item_id, (id) => {
  const it = (refItems.value ?? []).find((i: RefItem) => i.id === id)
  if (it && !form.source_note) form.source_note = `${it.docTitle}, hlm. ${it.page}`
})

const isDrug = computed(() => form.item_type === 'OBAT')
const DAYS = Array.from({ length: 15 }, (_, d) => ({ label: dayLabel(d), value: d }))
const ROUTES = ['Oral', 'Intravena', 'Intramuskular', 'Subkutan', 'Inhalasi', 'Topikal', 'Rektal', 'Sublingual']

function clean(s: string) {
  return s.trim() || null
}

async function save() {
  error.value = ''
  if (!form.master) {
    error.value = 'Pilih item dari master data.'
    return
  }
  if (isDrug.value && (!form.dose.trim() || !form.route.trim() || !form.frequency.trim())) {
    error.value = 'Obat wajib lengkap: dosis, rute, dan frekuensi harus diisi.'
    return
  }
  const body: PlanItemInput = {
    severity: form.severity,
    day: form.day,
    item_type: form.item_type,
    item_code: form.master.code,
    item_name: form.master.name,
    dose: isDrug.value ? clean(form.dose) : null,
    route: isDrug.value ? clean(form.route) : null,
    frequency: isDrug.value ? clean(form.frequency) : null,
    duration: isDrug.value ? clean(form.duration) : null,
    nature: form.nature,
    guideline_item_id: form.guideline_item_id || null,
    source_note: clean(form.source_note),
    note: clean(form.note),
  }
  busy.value = true
  try {
    if (props.item) await api.updatePlanItem(props.item.id, body)
    else await api.addPlanItem(body)
    toast.add({ title: props.item ? 'Baris rencana diperbarui' : 'Baris rencana ditambahkan', color: 'success' })
    open.value = false
    await refreshNuxtData()
  }
  catch (err) {
    error.value = apiErrorMessage(err)
  }
  finally {
    busy.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="item ? 'Ubah baris rencana' : 'Tambah baris rencana'"
    description="Obat, pemeriksaan lab, prosedur, atau tindakan pada satu hari rawat."
    :ui="{ content: 'max-w-2xl' }"
  >
    <template #body>
      <form id="plan-form" class="space-y-5" @submit.prevent="save">
        <!-- Jenis -->
        <UFormField label="Jenis" required>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <button
              v-for="(meta, t) in PLAN_TYPE"
              :key="t"
              type="button"
              class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm transition-colors"
              :class="form.item_type === t ? 'border-primary bg-primary/10 font-medium' : 'border-default hover:bg-elevated'"
              :aria-pressed="form.item_type === t"
              @click="form.item_type = t"
            >
              <UIcon :name="meta.icon" class="size-4" />
              {{ meta.label }}
            </button>
          </div>
        </UFormField>

        <div class="grid gap-4 sm:grid-cols-3">
          <UFormField label="Severity" required>
            <CommonSeverityFilter v-model="form.severity" :allow-all="false" compact />
          </UFormField>
          <UFormField label="Hari rawat" required>
            <USelect v-model="form.day" :items="DAYS" class="w-full" />
          </UFormField>
          <UFormField label="Sifat" required>
            <URadioGroup
              v-model="form.nature"
              orientation="horizontal"
              :items="[{ label: 'Wajib', value: 'WAJIB' }, { label: 'Kondisional', value: 'KONDISIONAL' }]"
            />
          </UFormField>
        </div>

        <UFormField
          :label="PLAN_TYPE[form.item_type].label"
          :description="`Dari master ${form.item_type === 'OBAT' ? 'KFA' : form.item_type === 'LAB' ? 'LOINC' : 'ICD-9-CM'}`"
          required
        >
          <CommonMasterSearch v-model="form.master" :kind="PLAN_TYPE[form.item_type].master" />
        </UFormField>

        <!-- Aturan pakai obat -->
        <div v-if="isDrug" class="grid gap-4 rounded-md bg-elevated/60 p-3 sm:grid-cols-4">
          <UFormField label="Dosis" required>
            <UInput v-model="form.dose" placeholder="2 g" class="w-full" />
          </UFormField>
          <UFormField label="Rute" required>
            <UInputMenu v-model="form.route" :items="ROUTES" create-item placeholder="Intravena" class="w-full" @create="(v: string) => form.route = v" />
          </UFormField>
          <UFormField label="Frekuensi" required>
            <UInput v-model="form.frequency" placeholder="1×/hari" class="w-full" />
          </UFormField>
          <UFormField label="Durasi">
            <UInput v-model="form.duration" placeholder="7 hari" class="w-full" />
          </UFormField>
        </div>

        <UFormField label="Butir acuan" description="Dari dokumen acuan yang sudah ditetapkan untuk CP ini.">
          <USelectMenu v-model="form.guideline_item_id" :items="guidelineOptions" value-key="value" class="w-full" />
        </UFormField>
        <UAlert
          v-if="!form.guideline_item_id"
          color="warning"
          variant="subtle"
          icon="i-lucide-info"
          title="Akan ditandai DI LUAR PANDUAN"
          description="Tetap sah, tetapi sertakan alasan klinis di catatan karena akan dibaca KSM/Komite Medik."
        />

        <div class="grid gap-4 sm:grid-cols-2">
          <UFormField label="Sumber dokumen">
            <UInput v-model="form.source_note" placeholder="PPK PD RSCM 2021, hlm. 234" class="w-full" />
          </UFormField>
          <UFormField label="Catatan">
            <UInput v-model="form.note" placeholder="Alasan klinis, kondisi pemberian…" class="w-full" />
          </UFormField>
        </div>

        <UAlert v-if="error" color="error" variant="subtle" icon="i-lucide-circle-alert" :title="error" />
      </form>
    </template>

    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton label="Batal" color="neutral" variant="ghost" @click="open = false" />
        <UButton type="submit" form="plan-form" :label="item ? 'Simpan perubahan' : 'Tambah baris'" :loading="busy" />
      </div>
    </template>
  </UModal>
</template>
