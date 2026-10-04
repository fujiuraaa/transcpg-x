# TransCPG-X

Sistem pengelolaan template Clinical Pathway (CP): menyusun acuan dan rencana klinis per grouper
INA-CBG, mengesahkannya lewat 4 tahap, lalu mengevaluasi mutu dan biaya sesudah CP aktif.

- **Struktur, peta modul, aturan bisnis, TODO, pertanyaan terbuka:** [docs/STRUKTUR.md](docs/STRUKTUR.md)
- Stack: Nuxt 4 SPA (`frontend/`) · Go serverless (`backend/`) · PostgreSQL/Supabase (`supabase/`) · Vercel

## Menjalankan lokal

Prasyarat: Go 1.24+, Node 22+, PostgreSQL 15+ (atau Supabase CLI).

### 1. Database

Dengan Supabase CLI (migrasi + seed otomatis):

```bash
supabase start
```

Atau PostgreSQL biasa:

```bash
createdb transcpg
for f in supabase/migrations/*.sql supabase/seed.sql; do psql -d transcpg -v ON_ERROR_STOP=1 -f "$f"; done
```

Akun uji ada di komentar atas `supabase/seed.sql`.

Data **demo** yang lebih lengkap (13 CP di semua tahap, 14 akun per peran) ada di `supabase/demo.sql`. Pakai file ini sebagai pengganti `seed.sql`.

### 2. Backend (API Go, port 8080)

```bash
cp backend/.env.example backend/.env
```

Isi `DATABASE_URL` dan `JWT_SECRET`, lalu:

```bash
cd backend && set -a && source .env && set +a && go run ./cmd/devserver
```

Tes aturan bisnis:

```bash
cd backend && go test ./...
```

### 3. Frontend (port 3000)

```bash
cd frontend && npm install && npm run dev
```

`/api` diteruskan ke `http://127.0.0.1:8080/api` (ubah lewat `NUXT_DEV_API_PROXY`).

## Deploy (Vercel, dua project)

| Project | Root directory | Env |
|---|---|---|
| backend | `backend/` | `DATABASE_URL`, `DB_SIMPLE_PROTOCOL=true` (pooler 6543), `JWT_SECRET`, `ALLOWED_ORIGINS=https://<frontend>` |
| frontend | `frontend/` | `NUXT_PUBLIC_API_BASE=https://<backend>/api` |

Migrasi: `supabase db push`.

**Demo online langkah demi langkah:** lihat [docs/DEPLOY-DEMO.md](docs/DEPLOY-DEMO.md). File siap tempel untuk Supabase SQL Editor: `supabase/deploy-demo.sql` (buat ulang dengan `sh supabase/gabung-demo.sh`).
