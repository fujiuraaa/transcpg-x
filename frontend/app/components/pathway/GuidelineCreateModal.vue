<script setup lang="ts">
// "Masukkan dokumen panduan sendiri" — bila dokumen yang dibutuhkan belum ada.
// Unggah lampiran PDF menyusul (lihat docs/STRUKTUR.md › TODO).
import type { GuidelineSource } from '~/types/domain'

const open = defineModel<boolean>('open', { default: false })
const emit = defineEmits<{ created: [title: string] }>()

const guidelines = useGuidelineApi()
const toast = useToast()

const form = reactive({ title: '', source_type: 'TIM_CP' as GuidelineSource, issuer: '', regulation_ref: '', year: '', icd10: '' })
const error = ref('')
const busy = ref(false)

watch(open, (v) => {
  if (v) Object.assign(form, { title: '', source_type: 'TIM_CP', issuer: '', regulation_ref: '', year: '', icd10: '' })
  error.value = ''
})

const SOURCES = Object.entries(GUIDELINE_SOURCE).map(([value, label]) => ({ value, label }))

async function save() {
  error.value = ''
  if (!form.title.trim()) {
    error.value = 'Nama dokumen wajib diisi.'
    return
  }
  busy.value = true
  try {
    await guidelines.create({
      title: form.title.trim(),
      source_type: form.source_type,
      issuer: form.issuer.trim() || null,
      regulation_ref: form.regulation_ref.trim() || null,
      year: form.year ? Number(form.year) : null,
      icd10_codes: form.icd10.split(/[\s,;]+/).map(s => s.trim().toUpperCase()).filter(Boolean),
    })
    toast.add({ title: 'Dokumen didaftarkan', description: 'Isi butir acuannya dari tab Acuan Klinis.', color: 'success' })
    emit('created', form.title.trim())
    open.value = false
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
  <UModal v-model:open="open" title="Daftarkan dokumen panduan" description="Untuk dokumen yang belum tersedia di daftar kandidat.">
    <template #body>
      <form id="guideline-form" class="space-y-4" @submit.prevent="save">
        <UFormField label="Nama dokumen" required>
          <UInput v-model="form.title" placeholder="Mis. PPK Kardiologi RS 2025" class="w-full" autofocus />
        </UFormField>
        <div class="grid gap-4 sm:grid-cols-2">
          <UFormField label="Sumber" required>
            <USelect v-model="form.source_type" :items="SOURCES" class="w-full" />
          </UFormField>
          <UFormField label="Tahun">
            <UInput v-model="form.year" type="number" inputmode="numeric" placeholder="2025" class="w-full" />
          </UFormField>
          <UFormField label="Penerbit">
            <UInput v-model="form.issuer" placeholder="Mis. PERKI" class="w-full" />
          </UFormField>
          <UFormField label="Nomor regulasi">
            <UInput v-model="form.regulation_ref" placeholder="Mis. HK.01.07/MENKES/…" class="w-full" />
          </UFormField>
        </div>
        <UFormField label="Kode ICD-10 yang dicakup" description="Pisahkan dengan koma. Dipakai untuk mencocokkan dokumen dengan CP.">
          <UInput v-model="form.icd10" placeholder="I50.0, I50.1" class="w-full" />
        </UFormField>
        <UAlert v-if="error" color="error" variant="subtle" :title="error" />
      </form>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton label="Batal" color="neutral" variant="ghost" @click="open = false" />
        <UButton type="submit" form="guideline-form" label="Daftarkan" :loading="busy" />
      </div>
    </template>
  </UModal>
</template>
