<script setup lang="ts">
// Lampiran PDF satu dokumen panduan. File diunggah langsung ke storage
// lewat URL bertanda tangan; API hanya mencatat metadata.
import type { CandidateRow } from '~/types/api'

const props = defineProps<{ document: CandidateRow, canEdit: boolean }>()
const emit = defineEmits<{ changed: [] }>()

const guidelines = useGuidelineApi()
const toast = useToast()
const { upload, cancel, progress, uploading } = useAttachmentUpload()

const { data, status, refresh } = await useAsyncData(
  () => `guideline:${props.document.id}:attachment`,
  () => guidelines.attachment(props.document.id),
)
const attachment = computed(() => data.value?.attachment ?? null)

// URL unduh berlaku ± 10 menit; ambil ulang bila sudah lewat 8 menit.
const fetchedAt = ref(Date.now())
watch(data, () => { fetchedAt.value = Date.now() })
async function open(e: MouseEvent) {
  if (Date.now() - fetchedAt.value < 8 * 60_000) return
  e.preventDefault()
  const win = window.open('', '_blank')
  await refresh()
  if (win && attachment.value) win.location.href = attachment.value.url
}

const input = ref<HTMLInputElement | null>(null)
const dragging = ref(false)
const error = ref<string | null>(null)
const pendingName = ref('')

function pick() {
  input.value?.click()
}

function onInput(e: Event) {
  const el = e.target as HTMLInputElement
  const file = el.files?.[0]
  el.value = ''
  if (file) send(file)
}

function onDrop(e: DragEvent) {
  dragging.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file && props.canEdit && !uploading.value) send(file)
}

async function send(file: File) {
  error.value = checkPdf(file)
  if (error.value) return
  pendingName.value = file.name
  try {
    await upload(props.document.id, file)
    toast.add({ title: 'Lampiran PDF tersimpan', description: file.name, color: 'success' })
    await refresh()
    emit('changed')
  }
  catch (err) {
    error.value = (err as { data?: unknown }).data ? apiErrorMessage(err) : (err as Error).message
  }
}

const confirmRemove = ref(false)
const removing = ref(false)
async function remove() {
  removing.value = true
  try {
    await guidelines.removeAttachment(props.document.id)
    toast.add({ title: 'Lampiran dihapus', color: 'success' })
    confirmRemove.value = false
    await refresh()
    emit('changed')
  }
  catch (err) {
    toast.add({ title: apiErrorMessage(err), color: 'error' })
  }
  finally {
    removing.value = false
  }
}
</script>

<template>
  <div class="space-y-2">
    <input ref="input" type="file" accept="application/pdf,.pdf" class="sr-only" tabindex="-1" aria-hidden="true" @change="onInput">

    <USkeleton v-if="status === 'pending' && !data" class="h-14 w-full" />

    <!-- Sedang mengunggah -->
    <div v-else-if="uploading" class="rounded-md border border-default p-3" role="status" aria-live="polite">
      <div class="flex items-center justify-between gap-3 text-sm">
        <span class="flex min-w-0 items-center gap-2">
          <UIcon name="i-lucide-upload" class="size-4 shrink-0 text-primary" />
          <span class="truncate">Mengunggah {{ pendingName }}</span>
        </span>
        <span class="shrink-0 tabular-nums text-muted">{{ progress }}%</span>
      </div>
      <UProgress :model-value="progress" size="sm" class="mt-2" />
      <div class="mt-2 flex justify-end">
        <UButton label="Batal" size="xs" color="neutral" variant="ghost" @click="cancel" />
      </div>
    </div>

    <!-- Sudah ada lampiran -->
    <div
      v-else-if="attachment"
      class="flex flex-wrap items-center gap-3 rounded-md border border-default p-3"
      :class="dragging && 'ring-2 ring-primary'"
      @dragover.prevent="canEdit && (dragging = true)"
      @dragleave="dragging = false"
      @drop.prevent="onDrop"
    >
      <div class="flex size-9 shrink-0 items-center justify-center rounded-md bg-brand-soft">
        <UIcon name="i-lucide-file-text" class="size-5 text-primary" />
      </div>
      <div class="min-w-40 flex-1">
        <p class="truncate text-sm font-medium" :title="attachment.name">{{ attachment.name }}</p>
        <p class="text-xs text-muted">{{ formatBytes(attachment.size) }} · diunggah {{ formatDate(attachment.uploaded_at, true) }}</p>
      </div>
      <div class="flex shrink-0 gap-1">
        <UButton
          :to="attachment.url"
          target="_blank"
          rel="noopener"
          label="Buka PDF"
          icon="i-lucide-external-link"
          size="sm"
          color="neutral"
          variant="outline"
          @click="open"
        />
        <template v-if="canEdit">
          <UButton icon="i-lucide-file-up" size="sm" color="neutral" variant="ghost" aria-label="Ganti PDF" title="Ganti PDF" @click="pick" />
          <UButton icon="i-lucide-trash-2" size="sm" color="error" variant="ghost" aria-label="Hapus PDF" title="Hapus PDF" @click="confirmRemove = true" />
        </template>
      </div>
    </div>

    <!-- Belum ada lampiran -->
    <button
      v-else-if="canEdit"
      type="button"
      class="flex w-full flex-col items-center gap-1 rounded-md border border-dashed p-4 text-center transition-colors"
      :class="dragging ? 'border-primary bg-primary/5' : 'border-accented hover:bg-elevated'"
      @click="pick"
      @dragover.prevent="dragging = true"
      @dragleave="dragging = false"
      @drop.prevent="onDrop"
    >
      <UIcon name="i-lucide-file-up" class="size-6 text-primary" />
      <span class="text-sm font-medium">Unggah PDF dokumen</span>
      <span class="text-xs text-muted">Seret file ke sini atau klik untuk memilih · maks. 50 MB</span>
    </button>
    <p v-else class="flex items-center gap-2 text-xs text-muted">
      <UIcon name="i-lucide-file-x" class="size-4" /> Belum ada lampiran PDF.
    </p>

    <UAlert
      v-if="error"
      :title="error"
      icon="i-lucide-circle-alert"
      color="error"
      variant="subtle"
      close
      @update:open="error = null"
    />

    <UModal v-model:open="confirmRemove" title="Hapus lampiran PDF?" :description="attachment?.name">
      <template #body>
        <p class="text-sm">File PDF dihapus dari penyimpanan. Butir acuan yang sudah diekstrak tetap ada.</p>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton label="Batal" color="neutral" variant="ghost" @click="confirmRemove = false" />
          <UButton label="Hapus PDF" color="error" :loading="removing" @click="remove" />
        </div>
      </template>
    </UModal>
  </div>
</template>
