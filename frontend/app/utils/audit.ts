import type { AuditCategory } from '~/types/api'

export const AUDIT_CATEGORY: Record<AuditCategory, { label: string, icon: string }> = {
  PENGESAHAN: { label: 'Pengesahan CP', icon: 'i-lucide-stamp' },
  ISI_CP: { label: 'Isi CP', icon: 'i-lucide-file-pen-line' },
  PADANAN: { label: 'Padanan', icon: 'i-lucide-link-2' },
  DOKUMEN: { label: 'Dokumen panduan', icon: 'i-lucide-book-open' },
  AKUN: { label: 'Akun pengguna', icon: 'i-lucide-user-cog' },
  RS: { label: 'Profil RS', icon: 'i-lucide-hospital' },
  MASUK: { label: 'Masuk aplikasi', icon: 'i-lucide-log-in' },
}
export const AUDIT_CATEGORIES = Object.keys(AUDIT_CATEGORY) as AuditCategory[]

/** Aksi yang perlu perhatian: ditandai ikon + warna, bukan warna saja. */
const ATTENTION = new Set(['KEMBALIKAN', 'CABUT', 'HAPUS', 'TOLAK', 'NONAKTIFKAN', 'MASUK_GAGAL', 'HAPUS_LAMPIRAN', 'GANTI_SANDI_GAGAL', 'KELUAR_SEMUA'])
export function isAttention(action: string) {
  return ATTENTION.has(action)
}

const DETAIL_LABEL: Record<string, string> = {
  dari: 'Dari tahap', ke: 'Ke tahap', alasan: 'Alasan', tahap: 'Tahap saat diubah',
  lama: 'Nilai lama', baru: 'Nilai baru', metode: 'Metode', email: 'Email', google: 'Lewat Google',
  peran_lama: 'Peran lama', peran_baru: 'Peran baru', nama_lama: 'Nama lama',
  item_name: 'Butir', item_code: 'Kode', item_type: 'Jenis', day: 'Hari', severity: 'Severity',
  dose: 'Dosis', route: 'Rute', frequency: 'Frekuensi', duration: 'Durasi', nature: 'Sifat', note: 'Catatan',
  source_note: 'Catatan sumber', kind: 'Jenis', description: 'Uraian', document_id: 'ID dokumen',
  sub_cp_id: 'ID sub-CP', reference_id: 'ID acuan', sumber: 'Sumber', kategori: 'Kategori', halaman: 'Halaman',
  ukuran_byte: 'Ukuran', tipe: 'Tipe RS', regional_bpjs: 'Regional BPJS', kelas_tertinggi: 'Kelas rawat tertinggi',
}
const HIDDEN = new Set(['id', 'updated_at', 'guideline_item_id', 'outside_guidance', 'created_at'])

/** Pasangan label–nilai yang layak ditampilkan dari detail kejadian. */
export function detailEntries(detail: Record<string, unknown>): { label: string, value: string }[] {
  return Object.entries(detail ?? {})
    .filter(([k, v]) => !HIDDEN.has(k) && v !== null && v !== '' && v !== undefined)
    .map(([k, v]) => ({
      label: DETAIL_LABEL[k] ?? k.replace(/_/g, ' '),
      value: k === 'ukuran_byte' ? formatBytes(Number(v))
        : typeof v === 'boolean' ? (v ? 'Ya' : 'Tidak')
          : typeof v === 'object' ? JSON.stringify(v) : String(v),
    }))
}

// --- Periode -------------------------------------------------------------------------
export function isoDate(d: Date) {
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

export type PeriodPreset = '7' | '30' | '90' | 'tahun' | 'semua' | 'kustom'
export const PERIOD_PRESETS: { label: string, value: PeriodPreset }[] = [
  { label: '7 hari', value: '7' },
  { label: '30 hari', value: '30' },
  { label: '90 hari', value: '90' },
  { label: 'Tahun ini', value: 'tahun' },
  { label: 'Semua', value: 'semua' },
  { label: 'Kustom', value: 'kustom' },
]

export function presetRange(p: PeriodPreset, now = new Date()): { from?: string, to?: string } {
  const to = isoDate(now)
  if (p === 'semua' || p === 'kustom') return {}
  if (p === 'tahun') return { from: `${now.getFullYear()}-01-01`, to }
  const from = new Date(now)
  from.setDate(from.getDate() - Number(p) + 1)
  return { from: isoDate(from), to }
}

export function periodLabel(from?: string, to?: string) {
  if (!from && !to) return 'seluruh waktu'
  const f = from ? formatDate(from) : 'awal'
  const t = to ? formatDate(to) : 'sekarang'
  return `${f} – ${t}`
}

// --- CSV di sisi klien (laporan kecil) ---------------------------------------------
export function todayStamp() {
  return isoDate(new Date()).replace(/-/g, '')
}

export function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}

/** CSV aman untuk Excel: BOM UTF-8 dan sel berawalan = + - @ diberi kutip tunggal. */
export function downloadCsv<T>(filename: string, columns: { label: string, value: (row: T) => unknown }[], rows: T[]) {
  const cell = (v: unknown) => {
    let s = v === null || v === undefined ? '' : String(v)
    if (/^[=+\-@\t\r]/.test(s)) s = `'${s}`
    return /[",\n\r]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s
  }
  const lines = [columns.map(c => cell(c.label)).join(','), ...rows.map(r => columns.map(c => cell(c.value(r))).join(','))]
  saveBlob(new Blob([String.fromCharCode(0xFEFF) + lines.join('\r\n')], { type: 'text/csv;charset=utf-8' }), filename)
}

/** Tanggal/waktu dalam WIB untuk CSV: "2026-10-04" atau "2026-10-04 07:15". */
export function wibStamp(iso: string | null | undefined, withTime = false): string {
  if (!iso) return ''
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Asia/Jakarta', year: 'numeric', month: '2-digit', day: '2-digit',
    ...(withTime ? { hour: '2-digit', minute: '2-digit', hourCycle: 'h23' as const } : {}),
  }).formatToParts(new Date(iso))
  const v = (t: string) => parts.find(p => p.type === t)?.value ?? ''
  const date = `${v('year')}-${v('month')}-${v('day')}`
  return withTime ? `${date} ${v('hour')}:${v('minute')}` : date
}
