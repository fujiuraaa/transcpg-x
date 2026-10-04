import type { GuidelineAttachment, GuidelineItemRow } from '~/types/api'
import type { GuidelineCategory, GuidelineSource } from '~/types/domain'

export interface UploadTarget {
  upload: { url: string, method: string, headers: Record<string, string> }
  object_path: string
}

/** Dokumen panduan: butir acuan, "Masukkan dokumen panduan sendiri", lampiran PDF. */
export function useGuidelineApi() {
  const api = useApi()
  return {
    items: (documentId: number) => api<GuidelineItemRow[]>(`/guidelines/${documentId}/items`),
    create: (body: {
      title: string
      source_type: GuidelineSource
      issuer?: string | null
      regulation_ref?: string | null
      year?: number | null
      icd10_codes?: string[]
    }) => api<{ id: number }>('/guidelines', { method: 'POST', body }),
    addItem: (documentId: number, body: {
      category: GuidelineCategory
      title: string
      page: string
      quote: string
      icd10_codes?: string[]
    }) => api<{ id: number }>(`/guidelines/${documentId}/items`, { method: 'POST', body }),

    // Lampiran PDF
    attachment: (documentId: number) =>
      api<{ attachment: GuidelineAttachment | null, max_size: number }>(`/guidelines/${documentId}/attachment`),
    attachmentUploadUrl: (documentId: number, file: File) =>
      api<UploadTarget>(`/guidelines/${documentId}/attachment/upload-url`, {
        method: 'POST',
        body: { filename: file.name, size: file.size, content_type: file.type || 'application/pdf' },
      }),
    attachmentComplete: (documentId: number, objectPath: string, filename: string) =>
      api<{ attachment: GuidelineAttachment | null }>(`/guidelines/${documentId}/attachment/complete`, {
        method: 'POST',
        body: { object_path: objectPath, filename },
      }),
    removeAttachment: (documentId: number) => api(`/guidelines/${documentId}/attachment`, { method: 'DELETE' }),
  }
}
