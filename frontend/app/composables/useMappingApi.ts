import type { KptlRow, KptlSuggestion, KptlSummary, SnomedRow, SnomedVocab, SnomedVocabSummary } from '~/types/api'

/** Meja kerja Padanan KPTL dan SNOMED-CT. */
export function useMappingApi() {
  const api = useApi()
  return {
    kptl: {
      summary: () => api<KptlSummary>('/mappings/kptl/summary'),
      worklist: (tab: 'belum' | 'semua', q = '') => api<KptlRow[]>('/mappings/kptl', { query: { tab, q } }),
      suggestions: (icd9: string) =>
        api<{ suggestions: KptlSuggestion[], warning: string }>(`/mappings/kptl/${encodeURIComponent(icd9)}/suggestions`),
      set: (icd9: string, kptlCode: string, note?: string | null) =>
        api(`/mappings/kptl/${encodeURIComponent(icd9)}`, { method: 'PUT', body: { kptl_code: kptlCode, note } }),
      remove: (icd9: string) => api(`/mappings/kptl/${encodeURIComponent(icd9)}`, { method: 'DELETE' }),
    },
    snomed: {
      summary: () => api<{ vocabularies: SnomedVocabSummary[], can_manage: boolean }>('/mappings/snomed/summary'),
      worklist: (filter: '' | 'ragu' | 'belum', vocab: '' | SnomedVocab = '') =>
        api<SnomedRow[]>('/mappings/snomed', { query: { filter, vocab } }),
      set: (vocab: SnomedVocab, code: string, conceptId: string) =>
        api(`/mappings/snomed/${vocab}/${encodeURIComponent(code)}`, { method: 'PUT', body: { concept_id: conceptId } }),
    },
  }
}
