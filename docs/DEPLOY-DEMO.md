# Panduan Deploy Demo Online — TransCPG-X

Hasil akhir: demo bisa dibuka siapa saja lewat link, dari login sampai seluruh data.

```
Browser ──► Frontend (Vercel, Nuxt) ──► API (Vercel, Go) ──► Database + Storage (Supabase)
```

Semua layanan di bawah punya paket **gratis** yang cukup untuk demo. Perkiraan waktu: 30–45 menit.

> **Rahasia** (kata sandi database, `service_role key`, `JWT_SECRET`) hanya diketik di dasbor
> Supabase/Vercel. Jangan masukkan ke Git, chat, atau file yang di-commit.

---

## 0. Persiapan

1. Repositori sudah di GitHub (lihat panduan Git). Vercel mengambil kode dari sana.
2. Siapkan akun **Supabase** (supabase.com) dan **Vercel** (vercel.com). Keduanya bisa masuk dengan akun GitHub.
3. Buat `JWT_SECRET` di Terminal Mac. Hasilnya disimpan sebentar di catatan pribadi:

   ```bash
   openssl rand -base64 48
   ```

---

## 1. Supabase: database & penyimpanan

1. **New project**
   - Name: `transcpg-demo`
   - Database Password: buat yang kuat, lalu **simpan** karena dipakai di langkah 1.4.
   - Region: **Southeast Asia (Singapore)**, yang terdekat ke Indonesia.
2. Tunggu proyek selesai dibuat (±2 menit).
3. **SQL Editor → New query.** Buka file `supabase/deploy-demo.sql` di VS Code, salin **seluruh isinya**,
   tempel, lalu klik **Run**. Hasil yang diharapkan: *Success. No rows returned*.
   - File ini membuat semua tabel (migrasi 001–010), bucket PDF privat, dan data demo.
   - Jalankan **sekali saja** pada proyek baru. Menjalankan dua kali akan galat karena tabel sudah ada.
4. **Connect** (tombol di atas) → tab **Connection string** → pilih **Transaction pooler** (port **6543**).
   Salin URI-nya, lalu ganti `[YOUR-PASSWORD]` dengan kata sandi dari langkah 1.1. Ini nilai `DATABASE_URL`.
5. **Project Settings → API** (atau *API Keys*). Catat dua nilai:
   - **Project URL** → untuk `SUPABASE_URL`, contoh `https://abcd1234.supabase.co`
   - **service_role** key (klik *Reveal*) → untuk `SUPABASE_SERVICE_ROLE_KEY`. **Rahasia**, jangan dibagikan.

---

## 2. Vercel: API backend (Go)

1. **Add New… → Project** → pilih repositori `transcpg-x` → **Import**.
2. Isi pengaturan:
   - Project Name: `transcpg-x-api`
   - **Root Directory: `backend`** (klik *Edit* lalu pilih folder backend)
   - Framework Preset: **Other**
3. Buka **Environment Variables** dan isi:

   | Nama | Nilai |
   |---|---|
   | `DATABASE_URL` | URI pooler dari langkah 1.4 |
   | `DB_SIMPLE_PROTOCOL` | `true` |
   | `JWT_SECRET` | hasil `openssl rand` di langkah 0.3 |
   | `STORAGE_DRIVER` | `supabase` |
   | `SUPABASE_URL` | Project URL dari langkah 1.5 |
   | `SUPABASE_SERVICE_ROLE_KEY` | service_role key dari langkah 1.5 |
   | `ALLOWED_ORIGINS` | sementara `http://localhost:3000`, diganti di langkah 4 |

   `GOOGLE_CLIENT_ID` dan `OTP_SENDER` **dikosongkan**, sehingga login Google dan OTP telepon tampil nonaktif di demo.
4. Klik **Deploy**. Setelah selesai, catat domainnya, misalnya `https://transcpg-x-api.vercel.app`.
5. Uji dengan membuka `https://transcpg-x-api.vercel.app/api/health` di browser. Hasil yang diharapkan: `{"status":"ok"…}`.

---

## 3. Vercel: frontend (Nuxt)

1. **Add New… → Project** → impor repositori yang **sama** sekali lagi.
2. Isi pengaturan:
   - Project Name: `transcpg-x`
   - **Root Directory: `frontend`**
   - Framework Preset: **Nuxt.js** (biasanya terdeteksi otomatis)
3. Isi Environment Variables:

   | Nama | Nilai |
   |---|---|
   | `NUXT_PUBLIC_API_BASE` | `https://transcpg-x-api.vercel.app/api` (domain langkah 2.4 + `/api`) |

4. Klik **Deploy**, lalu catat domainnya, misalnya `https://transcpg-x.vercel.app`. **Ini link demo Anda.**

---

## 4. Hubungkan frontend ↔ API

1. Kembali ke proyek **transcpg-x-api** → **Settings → Environment Variables**.
2. Ubah `ALLOWED_ORIGINS` menjadi domain frontend **persis**, tanpa garis miring di akhir:
   `https://transcpg-x.vercel.app`
3. **Deployments** → titik tiga pada deployment teratas → **Redeploy**.

---

## 5. Coba demo

Buka link frontend, lalu masuk dengan akun demo. Daftar lengkap ada di bagian atas `supabase/demo.sql`;
semua akun memakai kata sandi yang sama, yang juga tercantum di sana.

| Untuk menunjukkan | Masuk sebagai |
|---|---|
| Menyusun CP, Rencana Klinis, Acuan | `dpjp@transcpg.test` |
| Review Tim CP, Padanan KPTL/SNOMED | `timcp@transcpg.test` |
| Persetujuan Komite (2 CP menunggu) | `komite@transcpg.test` |
| Pengesahan Direktur (1 CP menunggu) | `direktur@transcpg.test` |
| Evaluasi CP (2 CP "Perlu ditinjau") | `komite@transcpg.test` atau `direktur@transcpg.test` |
| Manajemen User (2 pendaftar menunggu), Laporan & Audit | `admin@transcpg.test` |

**Isi data demo:**
- 13 CP di semua tahap: 4 Aktif, 1 Menunggu Direktur, 2 Review Komite, 1 Review Tim CP, 1 Perlu Revisi, dan 4 Draf.
- 928 episode klaim fiktif.
- 11 dokumen panduan berisi 31 butir.
- Padanan KPTL & SNOMED-CT dengan beberapa yang ragu atau belum dipadankan.
- Riwayat masuk dan aktivitas selama 2 minggu.

**Segera setelah online:** masuk sebagai `superadmin@transcpg.test` dan `admin@transcpg.test`,
lalu ganti kata sandinya lewat **Profil Saya**, karena kata sandi demo tertulis di repositori.
Akun peran klinis boleh tetap memakai kata sandi demo agar mudah dibagikan ke peserta.

---

## Bila ada masalah

| Gejala | Penyebab & solusi |
|---|---|
| Login gagal "Server tidak dapat dihubungi" | `NUXT_PUBLIC_API_BASE` salah atau API belum ter-deploy. Cek `…/api/health`. |
| Konsol browser: *CORS* / *blocked by CORS policy* | `ALLOWED_ORIGINS` belum sama persis dengan domain frontend. Ganti, lalu **Redeploy** API. |
| API menjawab `layanan belum siap` | `DATABASE_URL` atau `JWT_SECRET` salah/kosong. Lihat **Logs** proyek API di Vercel. |
| Galat *prepared statement* di log API | `DB_SIMPLE_PROTOCOL` belum `true` (wajib untuk pooler port 6543). |
| Build API gagal soal versi Go | Beri tahu saya. `go.mod` memakai Go 1.26, dan bila Vercel belum mendukung, versinya disesuaikan. |
| Unggah PDF gagal | Pastikan `STORAGE_DRIVER=supabase` dan bucket `guideline-attachments` ada (Supabase → Storage). |
| Link *preview* Vercel (domain acak) tidak bisa login | Wajar: hanya domain di `ALLOWED_ORIGINS` yang diizinkan. Pakai domain produksi. |

## Memperbarui demo

- **Kode:** cukup `git push`, dan Vercel otomatis men-deploy ulang kedua proyek.
- **Struktur database** (ada migrasi baru): jalankan file migrasi barunya saja di SQL Editor.
- **Mengulang data demo dari awal:** buat proyek Supabase baru (paling bersih), jalankan `deploy-demo.sql`,
  lalu ganti `DATABASE_URL`, `SUPABASE_URL`, dan `SUPABASE_SERVICE_ROLE_KEY` di proyek API.
- File `deploy-demo.sql` dibuat ulang dengan `sh supabase/gabung-demo.sh` setiap kali migrasi atau `demo.sql` berubah.
