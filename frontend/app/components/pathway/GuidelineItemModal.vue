<script setup lang="ts">
// Tambah satu butir acuan ke dokumen panduan. Halaman & kutipan wajib diisi
// agar setiap butir bisa ditelusuri ke sumbernya.
import type { CandidateRow } from '~/types/api'
import type { GuidelineCategory } from '~/types/domain'

const props = defineProps<{ document: CandidateRow }>()
const open = defineModel<boolean>('open', { default: false })
const emit = defineEmits<{ saved: [] }>()

const guidelines = useGuidelineApi()
const toast = useToast()

const form = reactive({ category: 'TATA_LAKSANA' as GuidelineCategory, title: '', page: '', quote: '', icd10: '' })
const error = ref('')
const busy = ref(false)

watch(open, (v) => {
  if (v) Object.assign(form, { category: 'TATA_LAKSANA', title: '', page: '', quote: '', icd10: props.document.icd10_codes.join(', ') })
  error.value = ''
})

const CATEGORIES = Object.entries(GUIDELINE_CATEGORY).map(([value, label]) => ({ value, label }))

async function save() {
  error.value = ''
  if (!form.title.trim() || !form.page.trim() || !form.quote.trim()) {
    error.value = 'Judul, halaman, dan kutipan wajib diisi.'
    return
  }
  busy.value = true
  try {
    await guidelines.addItem(props.document.id, {
      category: form.category,
      title: form.title.trim(),
      page: form.page.trim(),
      quote: form.quote.trim(),
      icd10_codes: form.icd10.split(/[\s,;]+/).map(s => s.trim().toUpperCase()).filter(Boolean),
    })
    toast.add({ title: 'Butir acuan ditambahkan', color: 'success' })
    open.value = false
    emit('saved')
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
  <UModal v-model:open="open" title="Tambah butir acuan" :description="document.title" :ui="{ content: 'max-w-xl' }">
    <template #body>
      <form id="guideline-item-form" class="space-y-4" @submit.prevent="save">
        <div class="grid gap-4 sm:grid-cols-[1fr_8rem]">
          <UFormField label="Kategori" required>
            <USelect v-model="form.category" :items="CATEGORIES" class="w-full" />
          </UFormField>
          <UFormField label="Halaman" required>
            <UInput v-model="form.page" placeholder="234" class="w-full" />
          </UFormField>
        </div>
        <UFormField label="Judul butir" required>
          <UInput v-model="form.title" placeholder="Mis. Antibiotik lini pertama" class="w-full" />
        </UFormField>
        <UFormField label="Kutipan" description="Salin teks dari dokumen apa adanya." required>
          <UTextarea v-model="form.quote" :rows="4" class="w-full" />
        </UFormField>
        <UFormField label="Kode ICD-10 terkait">
          <UInput v-model="form.icd10" class="w-full" />
        </UFormField>
        <UAlert v-if="error" color="error" variant="subtle" :title="error" />
      </form>
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton label="Batal" color="neutral" variant="ghost" @click="open = false" />
        <UButton type="submit" form="guideline-item-form" label="Simpan butir" :loading="busy" />
      </div>
    </template>
  </UModal>
</template>
