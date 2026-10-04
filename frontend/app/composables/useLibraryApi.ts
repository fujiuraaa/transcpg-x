import type { LibraryItem, LibraryQuery, Paged } from '~/types/api'

export function useLibraryApi() {
  const api = useApi()
  const { query: hospitalQuery } = useHospitalProfile()
  return {
    // Tarif dominan mengikuti profil RS terpilih di bilah atas.
    list: (query: LibraryQuery) => api<Paged<LibraryItem>>('/pathways', { query: { ...query, ...hospitalQuery.value } }),
  }
}
