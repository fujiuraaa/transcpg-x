import type { Hospital } from '~/types/api'

/**
 * Profil RS terpilih di bilah atas. Menentukan tarif INA-CBG di Detail CP,
 * CP Aktif, dan Evaluasi. Tersimpan di cookie per peramban.
 */
export function useHospitalProfile() {
  const api = useApi()
  const hospitals = useState<Hospital[]>('hospital:list', () => [])
  const selectedId = useSharedCookie<number | null>('transcpg_hospital', { maxAge: 60 * 60 * 24 * 365 })

  const selected = computed(() =>
    hospitals.value.find(h => h.id === selectedId.value)
    ?? hospitals.value.find(h => h.is_default)
    ?? null,
  )

  /** Query string untuk endpoint bertarif: { hospital_id }. */
  const query = computed(() => (selected.value ? { hospital_id: selected.value.id } : {}))

  async function load() {
    hospitals.value = await api<Hospital[]>('/hospitals')
  }

  function select(id: number) {
    selectedId.value = id
  }

  return { hospitals, selected, query, load, select }
}
