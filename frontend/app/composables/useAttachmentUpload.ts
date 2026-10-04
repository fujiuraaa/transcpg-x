import type { UploadTarget } from '~/composables/useGuidelineApi'

export const MAX_ATTACHMENT_SIZE = 50 * 1024 * 1024

/** Periksa file sebelum diunggah; kembalikan pesan galat atau null. */
export function checkPdf(file: File): string | null {
  const isPdf = file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf')
  if (!isPdf) return 'Hanya file PDF yang dapat diunggah.'
  if (file.size === 0) return 'File kosong.'
  if (file.size > MAX_ATTACHMENT_SIZE) return `Ukuran file ${formatBytes(file.size)} melebihi batas 50 MB.`
  return null
}

/**
 * Unggah lampiran PDF: minta URL bertanda tangan → unggah langsung ke
 * storage (dengan progres) → konfirmasi ke API. File tidak lewat API.
 */
export function useAttachmentUpload() {
  const guidelines = useGuidelineApi()
  const progress = ref(0)
  const uploading = ref(false)
  let xhr: XMLHttpRequest | null = null

  function put(target: UploadTarget, file: File) {
    return new Promise<void>((resolve, reject) => {
      xhr = new XMLHttpRequest()
      xhr.open(target.upload.method, target.upload.url)
      for (const [k, v] of Object.entries(target.upload.headers)) xhr.setRequestHeader(k, v)
      xhr.upload.onprogress = (e) => { if (e.lengthComputable) progress.value = Math.round((e.loaded / e.total) * 100) }
      xhr.onload = () => {
        if (xhr!.status >= 200 && xhr!.status < 300) return resolve()
        let msg = `Unggah gagal (status ${xhr!.status}).`
        try { msg = JSON.parse(xhr!.responseText).message ?? msg } catch { /* bukan JSON */ }
        reject(new Error(msg.charAt(0).toUpperCase() + msg.slice(1)))
      }
      xhr.onerror = () => reject(new Error('Koneksi terputus saat mengunggah.'))
      xhr.onabort = () => reject(new Error('Unggahan dibatalkan.'))
      xhr.send(file)
    })
  }

  async function upload(documentId: number, file: File) {
    const bad = checkPdf(file)
    if (bad) throw new Error(bad)
    uploading.value = true
    progress.value = 0
    try {
      const target = await guidelines.attachmentUploadUrl(documentId, file)
      await put(target, file)
      return await guidelines.attachmentComplete(documentId, target.object_path, file.name)
    }
    finally {
      uploading.value = false
      xhr = null
    }
  }

  function cancel() {
    xhr?.abort()
  }

  return { upload, cancel, progress, uploading }
}
