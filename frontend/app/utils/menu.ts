// Menu bilah samping — urutan sesuai alur-aplikasi §1.
export interface MenuItem {
  label: string
  to: string
  icon: string
  adminOnly?: boolean
  badge?: 'approval' | 'registrations'
}

export const MENU: MenuItem[] = [
  { label: 'Dashboard', to: '/dashboard', icon: 'i-lucide-layout-dashboard' },
  { label: 'CP Library', to: '/library', icon: 'i-lucide-library' },
  { label: 'Approval', to: '/approval', icon: 'i-lucide-stamp', badge: 'approval' },
  { label: 'CP Aktif', to: '/cp-aktif', icon: 'i-lucide-badge-check' },
  { label: 'Evaluasi CP', to: '/evaluasi', icon: 'i-lucide-chart-column' },
  { label: 'Padanan KPTL', to: '/padanan/kptl', icon: 'i-lucide-receipt' },
  { label: 'Padanan SNOMED-CT', to: '/padanan/snomed', icon: 'i-lucide-network' },
  { label: 'Pengaturan', to: '/pengaturan', icon: 'i-lucide-settings', adminOnly: true, badge: 'registrations' },
]
