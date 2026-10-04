// Package store berisi seluruh query database, dipisah per modul.
package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Row adalah satu baris hasil query baca (laporan, daftar, ringkasan).
type Row = map[string]any

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// querier dipenuhi oleh *pgxpool.Pool dan pgx.Tx.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func queryRows(ctx context.Context, q querier, sql string, args ...any) ([]Row, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	out, err := pgx.CollectRows(rows, pgx.RowToMap)
	if out == nil {
		out = []Row{}
	}
	return out, err
}

func queryRow(ctx context.Context, q querier, sql string, args ...any) (Row, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectExactlyOneRow(rows, pgx.RowToMap)
}

func (s *Store) inTx(ctx context.Context, fn func(pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, s.pool, fn)
}

// IsNotFound melaporkan apakah err berarti data tidak ditemukan.
func IsNotFound(err error) bool { return errors.Is(err, pgx.ErrNoRows) }
