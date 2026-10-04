-- TransCPG-X · 008 · Metadata lampiran PDF dokumen panduan.
-- File disimpan di object storage (Supabase Storage, bucket privat);
-- attachment_path (migrasi 002) menyimpan path objeknya.

alter table guideline_documents
  add column attachment_name        text,         -- nama file asli (untuk ditampilkan)
  add column attachment_size        bigint check (attachment_size between 1 and 52428800),
  add column attachment_uploaded_at timestamptz,
  add column attachment_uploaded_by bigint references app_users(id);

-- Produksi (Supabase): buat bucket privat sekali saja, mis. lewat dashboard
-- atau SQL berikut (jalankan bila skema storage tersedia):
--   insert into storage.buckets (id, name, public, file_size_limit, allowed_mime_types)
--   values ('guideline-attachments', 'guideline-attachments', false, 52428800, '{application/pdf}')
--   on conflict (id) do nothing;
