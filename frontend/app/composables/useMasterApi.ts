import type { MasterKind } from '~/types/domain'

export interface MasterRow {
  code: string
  name: string
  extra: string | boolean | null
  active: boolean
}

/** Pencarian master data (minimal 2 karakter). */
export function useMasterApi() {
  const api = useApi()
  return {
    search: (kind: MasterKind, q: string, semanticTag?: string) =>
      api<MasterRow[]>(`/master/${kind}`, { query: { q, semantic_tag: semanticTag } }),
  }
}
