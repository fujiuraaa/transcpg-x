package handler

import (
	"errors"
	"strings"
	"testing"
)

func TestSetupHint(t *testing.T) {
	cases := map[string]string{
		`tidak bisa terhubung ke database: failed to connect to user=postgres.abc database=postgres: FATAL: password authentication failed for user "postgres"`: "kata sandi database salah",
		"tidak bisa terhubung ke database: FATAL: Tenant or user not found":                                                                                    "nama pengguna pooler salah",
		"tidak bisa terhubung ke database: dial tcp: lookup db.x.supabase.co: no such host":                                                                    "host database tidak ditemukan",
	}
	for in, want := range cases {
		if got := setupHint(errors.New(in)); !strings.Contains(got, want) {
			t.Errorf("%q → %q, ingin memuat %q", in, got, want)
		}
	}
}
