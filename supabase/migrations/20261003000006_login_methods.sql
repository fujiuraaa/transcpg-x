-- TransCPG-X · 006 · Metode masuk tambahan: Google (berdasarkan email) dan
-- nomor telepon + kode OTP. Hanya untuk akun yang sudah dibuat Admin RS.

alter table app_users add column phone text;
alter table app_users add constraint app_users_phone_e164
  check (phone is null or phone ~ '^\+[1-9][0-9]{7,14}$');
create unique index app_users_phone_key on app_users (phone) where phone is not null;

-- Kode OTP disimpan sebagai hash (HMAC), tidak pernah dalam bentuk asli.
create table login_otps (
  id           bigint generated always as identity primary key,
  user_id      bigint not null references app_users(id) on delete cascade,
  code_hash    text not null,
  expires_at   timestamptz not null,
  attempts     smallint not null default 0,
  consumed_at  timestamptz,
  created_at   timestamptz not null default now()
);
create index login_otps_user_idx on login_otps (user_id, created_at desc);

alter table login_otps enable row level security;
do $$
begin
  if exists (select 1 from pg_roles where rolname = 'anon') then
    execute 'revoke all on login_otps from anon, authenticated';
  end if;
end $$;
