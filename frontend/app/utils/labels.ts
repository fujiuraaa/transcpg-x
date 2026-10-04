import type { GuidelineCategory, GuidelineSource, MasterKind, PlanItemType, PlanNature, Severity, Stage } from '~/types/domain'

// Label tampilan. Sumber kebenaran tetap backend (GET /api/meta);
// salinan ini hanya agar badge bisa dirender tanpa menunggu meta.
export const STAGE_LABEL: Record<Stage, string> = {
  DRAF: 'Draf',
  REVISI: 'Perlu Revisi',
  REVIEW_TIM_CP: 'Review Tim CP',
  REVIEW_KOMITE: 'Review KSM/Komite Medik',
  MENUNGGU_DIREKTUR: 'Menunggu Direktur',
  AKTIF: 'Aktif',
}

// Makin dekat ke pengesahan, makin pekat merahnya; Aktif hijau, Revisi kuning.
export const STAGE_BADGE: Record<Stage, { color: 'neutral' | 'warning' | 'secondary' | 'primary' | 'success', variant: 'subtle' | 'solid' | 'outline' }> = {
  DRAF: { color: 'neutral', variant: 'outline' },
  REVISI: { color: 'warning', variant: 'subtle' },
  REVIEW_TIM_CP: { color: 'secondary', variant: 'subtle' },
  REVIEW_KOMITE: { color: 'primary', variant: 'subtle' },
  MENUNGGU_DIREKTUR: { color: 'primary', variant: 'solid' },
  AKTIF: { color: 'success', variant: 'subtle' },
}

export function severityLabel(s: Severity | number | null | undefined): string {
  return s === 1 ? 'I' : s === 2 ? 'II' : s === 3 ? 'III' : '—'
}

export function formatRupiah(v: number | null | undefined): string {
  if (v == null) return '—'
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(v)
}

// --- Rencana klinis -------------------------------------------------------------
export const PLAN_TYPE: Record<PlanItemType, { label: string, icon: string, master: MasterKind }> = {
  OBAT: { label: 'Obat', icon: 'i-lucide-pill', master: 'kfa' },
  LAB: { label: 'Laboratorium', icon: 'i-lucide-flask-conical', master: 'loinc' },
  PROSEDUR: { label: 'Prosedur', icon: 'i-lucide-scan-line', master: 'icd9cm' },
  TINDAKAN: { label: 'Tindakan', icon: 'i-lucide-hand-heart', master: 'icd9cm' },
}

export const PLAN_NATURE: Record<PlanNature, string> = {
  WAJIB: 'Wajib',
  KONDISIONAL: 'Kondisional',
}

export function dayLabel(day: number): string {
  return `H${day}`
}

// --- Dokumen panduan -----------------------------------------------------------
export const GUIDELINE_SOURCE: Record<GuidelineSource, string> = {
  PPK_RSCM: 'PPK RSCM',
  PNPK: 'PNPK Kemenkes',
  PPK_ASOSIASI: 'PPK Asosiasi',
  TIM_CP: 'Dokumen Tim CP',
  INTERNASIONAL: 'Internasional',
  TKMKB: 'TKMKB',
  LAIN: 'Lainnya',
}

export const GUIDELINE_CATEGORY: Record<GuidelineCategory, string> = {
  KRITERIA_DIAGNOSIS: 'Kriteria diagnosis',
  TATA_LAKSANA: 'Tata laksana',
  DOSIS_OBAT: 'Dosis obat',
  INDIKASI_TERAPI: 'Indikasi terapi',
  KONTRAINDIKASI: 'Kontraindikasi',
  MONITORING: 'Monitoring',
  KOMPLIKASI: 'Komplikasi',
  LAIN: 'Lain-lain',
}

// --- Lain-lain -------------------------------------------------------------------
export function formatDate(iso: string | null | undefined, withTime = false): string {
  if (!iso) return '—'
  return new Intl.DateTimeFormat('id-ID', {
    day: 'numeric', month: 'short', year: 'numeric',
    ...(withTime ? { hour: '2-digit', minute: '2-digit' } : {}),
  }).format(new Date(iso))
}

export function formatNumber(v: number | null | undefined, digits = 0): string {
  if (v == null) return '—'
  return new Intl.NumberFormat('id-ID', { maximumFractionDigits: digits }).format(v)
}

export function formatPercent(v: number | null | undefined, digits = 0): string {
  if (v == null) return '—'
  return `${formatNumber(v * 100, digits)}%`
}

/** "hari ini", "kemarin", "5 hari", "3 minggu", "2 bulan" sejak tanggal iso. */
export function sinceLabel(iso: string | null | undefined): string {
  if (!iso) return '—'
  const days = Math.floor((Date.now() - new Date(iso).getTime()) / 86_400_000)
  if (days <= 0) return 'hari ini'
  if (days === 1) return 'kemarin'
  if (days < 14) return `${days} hari`
  if (days < 60) return `${Math.floor(days / 7)} minggu`
  return `${Math.floor(days / 30)} bulan`
}

// --- Evaluasi CP -----------------------------------------------------------------
export const FLAG_LABEL: Record<string, string> = {
  MEDIAN_LOS_DI_ATAS_P75: 'Median LOS di atas p75 target',
  BANYAK_EPISODE_DI_ATAS_P75: '≥ 25% episode melebihi p75 target',
  PERGESERAN_SEVERITY_III: 'Porsi severity III naik ≥ 10 poin',
}

export function formatBytes(n: number | null | undefined): string {
  if (n == null) return '—'
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${formatNumber(n / 1024, 0)} KB`
  return `${formatNumber(n / 1024 / 1024, 1)} MB`
}

/** "baru saja", "5 menit yang lalu", "2 jam yang lalu"; lebih dari 7 hari → tanggal. */
export function relativeTime(iso: string | null | undefined): string {
  if (!iso) return '—'
  const sec = Math.round((new Date(iso).getTime() - Date.now()) / 1000)
  const abs = Math.abs(sec)
  if (abs < 45) return 'baru saja'
  const rtf = new Intl.RelativeTimeFormat('id', { numeric: 'auto' })
  if (abs < 3600) return rtf.format(Math.round(sec / 60), 'minute')
  if (abs < 86_400) return rtf.format(Math.round(sec / 3600), 'hour')
  if (abs < 7 * 86_400) return rtf.format(Math.round(sec / 86_400), 'day')
  return formatDate(iso, true)
}

/** Inisial untuk avatar: dua huruf pertama dari nama (tanpa gelar "dr."). */
export function initials(name: string | null | undefined): string {
  const words = (name ?? '').replace(/\b(dr|drg|ns|apt|prof)\.?\s+/gi, '').trim().split(/\s+/).filter(Boolean)
  return ((words[0]?.[0] ?? '') + (words[1]?.[0] ?? '')).toUpperCase() || '?'
}
