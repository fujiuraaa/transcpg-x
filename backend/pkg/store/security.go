package store

import (
	"context"
	"time"
)

// RecordLoginFailure mencatat satu percobaan gagal untuk key (mis. "email:x@y").
// Baris lebih dari sehari dibersihkan sekalian agar tabel tetap kecil.
func (s *Store) RecordLoginFailure(ctx context.Context, key string) error {
	_, err := s.pool.Exec(ctx, `
		with gc as (delete from login_failures where created_at < now() - interval '1 day')
		insert into login_failures (key) values ($1)`, key)
	return err
}

// LoginFailures: jumlah percobaan gagal untuk key dalam jendela waktu terakhir.
func (s *Store) LoginFailures(ctx context.Context, key string, window time.Duration) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `
		select count(*) from login_failures
		where key = $1 and created_at > now() - make_interval(secs => $2)`, key, window.Seconds()).Scan(&n)
	return n, err
}

// RevokeSessions menaikkan versi sesi: semua token yang sudah terbit ditolak.
func (s *Store) RevokeSessions(ctx context.Context, userID int64) (int, error) {
	var v int
	err := s.pool.QueryRow(ctx, `update app_users set token_version = token_version + 1 where id = $1 returning token_version`, userID).Scan(&v)
	return v, err
}

// GuidelineLocked melaporkan apakah dokumen panduan menjadi acuan CP yang
// isinya sedang terkunci (Review KSM/Komite Medik atau Menunggu Direktur).
func (s *Store) GuidelineLocked(ctx context.Context, documentID int64) (bool, error) {
	var locked bool
	err := s.pool.QueryRow(ctx, `
		select exists (
		  select 1 from pathway_references r join clinical_pathways p on p.id = r.pathway_id
		  where r.document_id = $1 and p.stage in ('REVIEW_KOMITE', 'MENUNGGU_DIREKTUR'))`, documentID).Scan(&locked)
	return locked, err
}
