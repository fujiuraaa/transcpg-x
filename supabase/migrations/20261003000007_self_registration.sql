-- TransCPG-X · 007 · Pendaftaran mandiri dengan persetujuan Admin RS.
-- Pendaftar memilih peran yang DIAJUKAN; akun belum aktif sampai Admin
-- menyetujui (dan boleh mengubah perannya) atau menolak dengan alasan.

alter table app_users
  add column registration_status text not null default 'DISETUJUI'
    check (registration_status in ('MENUNGGU', 'DISETUJUI', 'DITOLAK')),
  add column self_registered   boolean not null default false,
  add column registration_note text,          -- mis. unit kerja / SMF, diisi pendaftar
  add column reviewed_by       bigint references app_users(id),
  add column reviewed_at       timestamptz,
  add column rejection_reason  text;

-- Akun yang belum disetujui tidak boleh aktif.
alter table app_users add constraint app_users_pending_inactive
  check (registration_status = 'DISETUJUI' or not is_active);

create index app_users_pending_idx on app_users (created_at) where registration_status = 'MENUNGGU';
