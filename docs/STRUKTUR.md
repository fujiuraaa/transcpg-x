# Struktur & Rangka TransCPG-X

Rangka aplikasi disusun dari dokumen di `Trans SPKD/TransCPG-X/`:
`alur-aplikasi.pdf` (28 Sep 2026, acuan utama), `panduan-penggunaan-aplikasi-transcpg.pdf`
(22 Sep 2026), `TransCPG-X — Dokumentasi Aplikasi.pdf` (v1.0, Juni 2026), `rujukan-regulasi.pdf`,
dan spreadsheet `List Features CPG` / `Sprint QA SPKD CPG`.

Status: **backend & database berfungsi penuh untuk alur inti.** UI/UX dirancang bertahap:
**Dashboard**, **Detail CP**, **CP Library**, **Approval**, **CP Aktif**, **Evaluasi CP**, **Padanan KPTL & SNOMED-CT**, **Setup Rumah Sakit**, **Masuk/Daftar**, **Manajemen User** (antrean pendaftaran + tambah/ubah/nonaktifkan/hapus pengguna), **lampiran PDF panduan**, dan **Laporan & Audit**, **Profil Saya**, dan halaman galat/404 selesai — semua halaman sudah ber-UI; kerangka responsif dengan laci navigasi di layar kecil. Tema: gradasi merah (`app/app.config.ts`,
token `--sev-*` & `.bg-brand-*` di `main.css`), tanpa biru.

## Arsitektur

```
Nuxt 4 SPA (frontend/)  ──HTTP JSON + JWT──▶  Go API (backend/)  ──pgx──▶  PostgreSQL / Supabase (supabase/)
        Vercel project A                         Vercel project B                 RLS aktif tanpa policy
```

- **Frontend dan backend dipisah** — dua folder, dua project Vercel. Saat dev, Nuxt meneruskan `/api` ke Go.
- **Login custom JWT** (tabel `app_users`, bukan Supabase Auth). Token hanya berisi ID pengguna; peran
  dibaca ulang tiap request sehingga perubahan peran/nonaktif berlaku seketika.
- **Aturan bisnis ada di satu tempat**: `backend/pkg/domain` (tanpa DB/HTTP, diuji unit).
  Frontend tidak menduplikasi aturan — tombol & formulir mengikuti `permissions` dari server.
- **Database ikut menjaga**: obat wajib lengkap, kode harus ada di master, item tidak boleh ganda,
  RLS aktif agar REST/anon Supabase tidak bisa membaca data.

## Peta folder

```
transcpg-x/
├── supabase/
│   ├── migrations/
│   │   ├── …001_master_data.sql        profil RS, pengguna, master koding (ICD-10, ICD-9-CM, LOINC, KFA, KPTL, SNOMED), tarif INA-CBG
│   │   ├── …002_claims_guidelines.sql  episode klaim agregat (tanpa identitas pasien), dokumen & butir panduan, TKMKB
│   │   ├── …003_pathways.sql           CP, statistik severity, sub-CP, acuan, rencana klinis, kriteria, riwayat pengesahan & perubahan
│   │   ├── …004_mappings.sql           padanan KPTL & SNOMED-CT + audit
│   │   ├── …005_views_functions.sql    v_pathway_overview, v_grouper_volume, generate_draft_pathways(), refresh_pathway_severity_stats(), RLS
│   │   ├── …006–008                    metode masuk (telepon/OTP), pendaftaran mandiri, lampiran PDF panduan
│   │   ├── …009_audit.sql              activity_log + view audit_events (jejak audit gabungan)
│   │   └── …010_hardening.sql          versi sesi, batas percobaan masuk, snapshot pelaku di riwayat, tanggal aktif WIB
│   └── seed.sql                        data contoh pengembangan + akun uji
├── backend/                            Go 1.24 · module transcpg-x/backend
│   ├── api/index.go                    entry Vercel (semua /api/* → satu fungsi)
│   ├── cmd/devserver/main.go           server lokal
│   └── pkg/
│       ├── domain/                     ★ aturan bisnis + tes: role, stage, permission, readiness, evaluation
│       ├── auth/                       JWT, bcrypt, middleware
│       ├── store/                      query SQL per modul (1 file per modul)
│       ├── handlers/                   handler HTTP per modul (1 file per modul) + routing di app.go
│       ├── httpx/                      respons JSON & terjemahan galat DB → HTTP
│       ├── config/, db/
└── frontend/                           Nuxt 4 · Nuxt UI 4 · Tailwind 4 · TypeScript
    └── app/
        ├── pages/                      1 file per halaman (lihat tabel di bawah)
        ├── components/
        │   ├── layout/                 AppSidebar, AppTopbar, HospitalSelector
        │   ├── common/                 PagePlaceholder, SectionPlaceholder, DataPreview
        │   └── pathway/                StageBadge, ApprovalPanel, ReferencePanel, SummaryCards
        │       └── tabs/               13 tab Detail CP, 1 file per tab
        ├── composables/                useApi, useAuth, useHospitalProfile + 1 klien API per modul
        ├── middleware/                 auth.global (wajib login), admin (Pengaturan)
        ├── types/                      domain.ts (cermin backend), api.ts (bentuk respons)
        └── utils/                      menu, label tahap/severity, format rupiah
```

## Peta modul: halaman ↔ API ↔ tabel

| § | Halaman (frontend) | Endpoint (backend) | Tabel/view utama |
|---|---|---|---|
| 2 | `/login` | `POST /api/auth/login`, `GET /api/auth/methods`, `POST /api/auth/google`, `POST /api/auth/otp/request`, `POST /api/auth/otp/verify`, `GET /api/auth/me`, `POST /api/auth/logout` | `app_users`, `login_otps` |
| 3 | `/dashboard` | `GET /api/dashboard` | `v_pathway_overview`, `v_grouper_volume`, `guideline_*`, `tkmkb_topics` |
| 4 | `/library` | `GET /api/pathways?q&mdc&acuan&sort&page`, `GET /api/mdc` | `v_pathway_overview` |
| 5 | `/library/[code]` | `GET /api/pathways/{code}` + tab: `/flow` `/summary` `/diagnoses` `/procedures` `/tariffs` `/daily` `/cdss-rules` `/scoring` `/quality` `/readiness` `/history` | `clinical_pathways`, `pathway_*`, `claim_*`, `ina_cbg_tariffs` |
| 6 | Panel Acuan + tab Acuan Klinis | `GET/POST /api/pathways/{code}/references`, `DELETE …/references/{id}`, `GET …/guideline-candidates`, `GET /api/guidelines/{id}/items`, `POST /api/guidelines`, `POST /api/guidelines/{id}/items` | `pathway_references`, `guideline_documents`, `guideline_items` |
| 7 | Tab Rencana Klinis | `GET/POST /api/pathways/{code}/plan`, `PUT/DELETE …/plan/{id}`, `GET /api/master/{kfa\|loinc\|icd9cm}` | `clinical_plan_items` |
| 8 | Tab Kriteria | `GET/POST /api/pathways/{code}/criteria`, `DELETE …/criteria/{id}` | `pathway_criteria` |
| 9 | Panel Pengesahan | `POST /api/pathways/{code}/transition` `{action: MAJUKAN\|KEMBALIKAN\|CABUT, reason}` | `clinical_pathways.stage`, `approval_history` |
| 10 | `/approval` | `GET /api/approval?tab=saya\|proses\|aktif`, `GET /api/approval/counts` | `v_pathway_overview` |
| 11 | `/cp-aktif` | `GET /api/active-pathways`, `GET /api/active-pathways/{code}?severity` | idem + `claim_*` |
| 12 | `/evaluasi` | `GET /api/evaluations?tab=aktif\|belum`, `GET /api/evaluations/{code}` | `claim_episodes` sesudah `activated_at` |
| 13 | `/padanan/kptl` | `GET /api/mappings/kptl/summary`, `GET /api/mappings/kptl?tab&q`, `GET …/{icd9}/suggestions`, `PUT/DELETE …/{icd9}` | `kptl_mappings`, `mapping_audit_log` |
| 14 | `/padanan/snomed` | `GET /api/mappings/snomed/summary`, `GET /api/mappings/snomed?filter&vocab`, `PUT …/{vocab}/{code}`, `GET /api/master/snomed?semantic_tag` | `snomed_mappings`, `snomed_concepts` |
| 15 | `/pengaturan`, `/pengaturan/users`, `/pengaturan/rumah-sakit` | `GET/POST /api/admin/users`, `PATCH/DELETE /api/admin/users/{id}`, `PUT /api/hospitals/{id}` | `app_users`, `hospitals` |
| 16 | Pemilih profil RS (bilah atas) | `GET /api/hospitals`; endpoint bertarif menerima `?hospital_id=` | `hospitals`, `ina_cbg_tariffs` |
| 17 | `/pengaturan/audit` | `GET /api/admin/audit/events?from&to&category&actor&q&page`, `GET …/events.csv`, `GET …/actors`, `GET …/reports/approval?from&to`, `GET …/reports/users?from&to` | `audit_events` (view), `activity_log`, `approval_history`, `app_users` |
| 18 | `/profil` (semua peran) | `PATCH /api/auth/me` `{full_name, phone}`, `POST /api/auth/password` `{current_password, new_password}`, `GET /api/auth/me/activity` | `app_users`, `audit_events` |
| — | (bersama) | `GET /api/meta` (label peran/tahap, konstanta), `GET /api/master/{kind}?q` | — |

## Aturan bisnis yang sudah ditegakkan server

**Pengesahan 4 tahap** (`domain/stage.go`, `domain/permission.go`)

| Tahap | Majukan / kembalikan oleh | Isi CP boleh diubah oleh |
|---|---|---|
| DRAF / REVISI | Dokter, DPJP, Tim CP, Koordinator CP, Admin* | sama + Apoteker (rencana & kriteria saja) |
| REVIEW_TIM_CP | Tim CP, Koordinator CP, Admin* | Tim CP, Koordinator CP, Admin* |
| REVIEW_KOMITE | KSM/Komite Medik, Admin* | **dikunci** |
| MENUNGGU_DIREKTUR | Direktur, System Admin, Super Admin | **dikunci** |
| AKTIF | Cabut → Revisi: Komite Medik, Direktur, Super Admin | Tim CP, Koordinator CP, Admin* (tercatat) |

\*Admin = Admin RS, System Admin, Super Admin. Kembalikan & cabut **wajib alasan**.
Semua perpindahan tercatat di `approval_history`; semua perubahan isi di `pathway_change_log`.
Perubahan isi dan perpindahan tahap memakai `SELECT … FOR UPDATE` sehingga tidak bisa balapan.

**Syarat aktif** (`domain/readiness.go`). Ketujuh syarat dinilai sekaligus. Hanya #1 (acuan CP) dan
#3 (rencana tiap severity ≥10 episode) yang menghambat. Kekurangan dikembalikan semua dalam satu
respons 409 `SYARAT_BELUM_TERPENUHI`.

**Evaluasi** (`domain/evaluation.go`). Perlu Ditinjau bila median LOS > p75 target, ≥25% episode
> p75, atau porsi severity III naik ≥10 poin. Baseline = klaim sebelum `activated_at`.

**Lain-lain.** Grouper <5 episode dilewati. Library 25 per halaman. Rencana H0–H14. Konsep SNOMED
pensiun ditolak. Admin tidak bisa memberi atau mengubah peran yang lebih tinggi darinya.

## Metode masuk

Email + kata sandi, **Google**, dan **nomor telepon (OTP)** — untuk akun yang sudah **disetujui**.

**Pendaftaran mandiri** (`/daftar`, `POST /api/auth/register`) — keputusan pemilik produk 3 Okt 2026, menggantikan
"akun hanya dibuat Admin RS" di dokumen lama. Pendaftar mengisi data + **peran yang diajukan** (peran admin
tidak bisa diajukan), boleh lewat Google. Akun berstatus `MENUNGGU` dan nonaktif sampai Admin RS
menyetujui (menetapkan peran final, boleh berbeda) atau menolak dengan alasan di Pengaturan › Manajemen User
(`/api/admin/registrations`). Yang ditolak boleh mendaftar ulang. Belum ada: verifikasi email otomatis
(Admin yang memverifikasi identitas) dan pembatasan laju pendaftaran per IP.

- **Google**: popup Google Identity Services → access token → server memverifikasi lewat `tokeninfo`
  (audience harus = `GOOGLE_CLIENT_ID`, email terverifikasi) → dicocokkan ke `app_users.email`.
  Aktif bila `GOOGLE_CLIENT_ID` diisi (OAuth Client ID tipe *Web*, origin = URL frontend).
- **Telepon**: kode 6 digit, disimpan sebagai HMAC, berlaku 5 menit, sekali pakai, maks. 5 percobaan,
  kirim ulang setelah 60 detik. Respons permintaan kode selalu sama agar nomor terdaftar tidak bisa ditebak.
  Nomor diisi Admin pada akun pengguna (`phone`, format E.164). Aktif bila `OTP_SENDER` diisi;
  `log` hanya untuk pengembangan — produksi perlu pengirim SMS/WA di `backend/pkg/auth/sms.go`.

## Lampiran PDF panduan

Tab **Acuan Klinis** › dokumen terpilih › *Lampiran PDF*: unggah (klik atau seret, maks. 50 MB, hanya PDF),
buka, ganti, hapus. Hanya peran yang boleh menetapkan acuan (Tim CP/Admin) yang bisa mengubah; semua bisa membuka.
Daftar dokumen memberi tanda 📎 PDF.

File **tidak lewat API** (batas body fungsi Vercel ±4.5 MB): `POST …/attachment/upload-url` → browser
`PUT` langsung ke storage dengan URL bertanda tangan (progres ditampilkan) → `POST …/attachment/complete`
(server memeriksa objek ada & ukurannya, lalu menghapus file lama). URL unduh berlaku ± 10 menit.

- `STORAGE_DRIVER=local` — pengembangan saja; file di `backend/data/uploads`, diperiksa tanda `%PDF-`.
- `STORAGE_DRIVER=supabase` — produksi. Buat bucket **privat** `guideline-attachments` (batas 50 MB,
  MIME `application/pdf`), isi `SUPABASE_URL` + `SUPABASE_SERVICE_ROLE_KEY` (hanya di backend).
  Driver ini **belum diuji** terhadap proyek Supabase sungguhan.

## Profil Saya & kenyamanan antarmuka

- `/profil` (klik nama di pojok kiri bawah): ubah nama & nomor telepon, ganti kata sandi sendiri (wajib
  sandi lama, minimal 8, beda dari lama), dan 20 aktivitas terakhir akun termasuk riwayat masuk + IP.
  Peran, RS, email, dan status tetap hanya lewat Admin. Semua perubahan tercatat di jejak audit.
- Gerak halus: transisi halaman/layout (pudar + geser 4px, 160 ms), garis muat merah di atas layar,
  layar pembuka bermerek saat aplikasi pertama dimuat (`app/spa-loading-template.html`), sidebar lipat
  beranimasi; semuanya mati otomatis bila pengguna memilih "kurangi gerakan".
- `app/error.vue`: halaman 404 dan galat tak terduga dengan tombol Kembali / Ke Dashboard.
- Catatan: token sesi tidak dicabut saat kata sandi diganti (JWT stateless) — sesi lain tetap aktif
  sampai kedaluwarsa.

## Laporan & Audit

`/pengaturan/audit` (khusus admin), dua tab dengan periode bersama (7/30/90 hari, tahun ini, semua, kustom; tanggal WIB, inklusif):

- **Jejak audit** — satu daftar untuk 7 kategori: Pengesahan CP, Isi CP, Padanan (dari tabel riwayat yang
  sudah ada) serta Dokumen panduan, Akun pengguna, Profil RS, Masuk aplikasi (tabel baru `activity_log`).
  Saring kategori/pelaku/kata kunci, rincian per kejadian, unduh CSV (maks. 10.000 baris, aman dari
  injeksi formula Excel). Hanya baca — tidak ada endpoint ubah/hapus.
- **Laporan** — *Pengesahan CP* (diajukan/disahkan/dikembalikan/dicabut, rata-rata lama tiap tahap,
  posisi semua CP + CSV) dan *Pengguna & akses* (akun aktif/nonaktif, tidak masuk ≥ 90 hari, menunggu
  persetujuan, jumlah masuk & aksi per pengguna + CSV).

Lingkup: Admin RS hanya kejadian RS-nya (RS pelaku saat kejadian); System/Super Admin semua RS. Yang
dicatat di `activity_log`: masuk berhasil (metode) & gagal kata sandi (akun yang ada saja), daftar mandiri,
buat/ubah peran/aktif-nonaktif/atur ulang sandi/hapus akun, setujui/tolak pendaftaran, ubah profil RS,
daftarkan dokumen, tambah butir, unggah/ganti/hapus lampiran. Nama & peran pelaku disalin saat kejadian.
Pencatatan bersifat *best effort*: bila gagal, aksi tetap berhasil dan galat ditulis ke log server.

## Keamanan (tinjauan 4 Okt 2026)

- **Lingkup Admin RS**: hanya melihat/mengelola akun & pendaftar di RS-nya; tidak bisa membuat akun untuk RS lain.
- **Tidak saling ambil alih**: admin hanya mengelola akun berperingkat lebih rendah (`Role.ManageableBy`);
  hanya Super Admin yang boleh mengelola sesama Super Admin. Sandi sendiri hanya lewat Profil (perlu sandi lama).
- **Sesi bisa dicabut**: token membawa `token_version`; naik saat sandi diganti/diatur ulang atau
  "Keluar dari semua perangkat" (`POST /api/auth/logout-all`) → token lama ditolak.
- **Masuk**: 5 gagal / 15 menit per email (juga email fiktif) → 429; waktu respons disamakan.
  **OTP**: jatah percobaan diambil atomik sebelum dicek, maks. 5 kode/jam & 10 salah/jam, pesan galat seragam.
- Kunci turunan per keperluan (JWT, OTP, token storage). Sandi > 72 byte → 422.
- Dokumen panduan yang menjadi acuan CP di tahap Review Komite / Menunggu Direktur terkunci (butir & PDF).
- Butir panduan di Rencana Klinis harus dari dokumen acuan CP itu sendiri.
- PDF diperiksa tanda `%PDF-` juga untuk Supabase; token unggah lokal tidak bisa menimpa file.
- IP di jejak audit hanya membaca header proxy di Vercel (atau `TRUST_PROXY_HEADERS=true`).
- Riwayat pengesahan/isi/padanan menyimpan snapshot nama & RS pelaku (trigger), tanggal aktif dihitung WIB.

## Sudah diverifikasi

- `go test ./...` — tes unit domain (transisi, wewenang, kunci isi, syarat aktif, ambang evaluasi).
- Migrasi + seed dijalankan pada PostgreSQL 16.
- Smoke test API end-to-end: login, dashboard, library, detail, alur lengkap DRAF → AKTIF termasuk
  penolakan wewenang, kunci isi, alasan wajib, gerbang syarat aktif, kembalikan & ajukan ulang,
  CP Aktif, evaluasi (tanda muncul pada data contoh), padanan KPTL/SNOMED, master, admin user.
- Frontend: `nuxt typecheck` bersih, `nuxt build` sukses.
- Lampiran PDF (driver lokal): unggah, ganti (file lama terhapus), buka, hapus; tolak non-PDF, isi palsu,
  > 50 MB, peran tanpa wewenang, path objek dokumen lain.
- Laporan & Audit: non-admin 403; masuk/gagal masuk/aktif-nonaktif tercatat; saring kategori, pelaku,
  kata kunci, tanggal; validasi tanggal & kategori; CSV; kedua laporan; tampilan seluler tanpa geser samping.

## Belum dikerjakan (TODO)

- Uji driver storage Supabase pada proyek sungguhan; ekstraksi butir acuan dari PDF masih manual.
- Job analisis klaim: impor klaim, `split_decision`, usulan aturan CDSS, indikator mutu, matriks
  Intervensi Harian, langkah siklus hidup di tab Alur Pelaksanaan.
- API untuk TransCPR-X mengambil template AKTIF (disebut di dokumentasi v1.0).
- Audit: retensi/arsip `activity_log`, peringatan otomatis (mis. gagal masuk berulang).
- Verifikasi email pendaftaran mandiri (sekarang Admin yang memverifikasi identitas); pendaftar DITOLAK bisa mendaftar ulang dengan email yang sama.
- Keputusan: pemisahan tugas pengesahan (saat ini System/Super Admin boleh bertindak di semua tahap, Admin RS di tahap Tim CP & Komite).
- Tes integrasi store/handler terhadap database.

## Pertanyaan terbuka (perbedaan antar-dokumen)

Bila dokumen berbeda, rangka ini mengikuti **alur-aplikasi.pdf (28 Sep)** sebagai yang terbaru.
Mohon dikonfirmasi:

1. **Ubah isi CP Aktif.** Alur §6–7: Tim CP/Admin boleh mengubah langsung (tercatat). Panduan §4:
   terkunci total, harus dikembalikan ke Revisi. → *Diikuti: boleh, tercatat.*
2. **Siapa yang mencabut CP Aktif.** Alur §9: Komite, Direktur, Superadmin. Panduan §4: Komite,
   Direktur, Admin. → *Diikuti: Komite, Direktur, Super Admin.*
3. **"Tim Koding"** disebut sebagai pemeta KPTL/SNOMED, tetapi tidak ada sebagai peran sistem; alur
   §13 menyebut hanya Tim CP & Admin. → *Diikuti: Tim CP, Koordinator CP, Admin.*
4. **Regional BPJS DKI Jakarta.** Alur §16: Regional 1. Dokumentasi v1.0: Regional 5. → *Seed: 1.*
5. **Admin RS di tahap review.** Panduan menyebut "Admin" boleh memajukan di Draf/Review Tim CP/Review
   Komite. Apakah termasuk Admin RS, atau hanya System/Super Admin?
6. **Admin RS memberi peran System Admin.** Panduan §14 mengizinkan. Karena System Admin bisa
   mengesahkan di tahap Direktur, rangka ini **membatasinya** ke System/Super Admin.
7. **Peran "Dokter".** Apakah Dokter Umum & Residen (PPDS) juga boleh menyusun/mengajukan? → *Saat ini
   hanya DOKTER & DPJP.*
8. **Model approval per setting** (IGD/ICU provisional, Rawat Jalan simplified) di dokumentasi v1.0
   tidak muncul lagi di dokumen terbaru. → *Hanya rawat inap 4 tahap.*
9. **Aturan target LOS** dan **aturan pecah/satu CP** belum didefinisikan. → *Sementara target LOS =
   median LOS; `split_decision` diisi job analisis.*
