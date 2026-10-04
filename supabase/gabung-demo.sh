#!/bin/sh
# Menggabungkan semua migrasi + bucket storage + data demo menjadi satu file
# untuk ditempel di Supabase › SQL Editor (database baru yang masih kosong).
# Jalankan dari folder supabase/:  sh gabung-demo.sh
set -e
cd "$(dirname "$0")"
out=deploy-demo.sql
{
  echo "-- TransCPG-X · DEPLOY DEMO — dibuat otomatis oleh gabung-demo.sh, jangan disunting langsung."
  echo "-- Isi: migrasi 001–010, bucket storage, dan data demo. Jalankan SEKALI pada database kosong."
  echo
  for f in migrations/*.sql storage.sql demo.sql; do
    echo "-- ==================== $f ===================="
    cat "$f"
    echo
  done
} > "$out"
echo "Selesai: supabase/$out ($(wc -l < "$out") baris)"
