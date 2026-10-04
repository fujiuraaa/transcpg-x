-- TransCPG-X · Bucket penyimpanan lampiran PDF (khusus Supabase)
-- Privat: file hanya bisa diunduh lewat URL bertanda tangan dari backend.
insert into storage.buckets (id, name, public, file_size_limit, allowed_mime_types)
values ('guideline-attachments', 'guideline-attachments', false, 52428800, array['application/pdf'])
on conflict (id) do nothing;
