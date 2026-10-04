package store

import (
	"context"
	"fmt"
)

// masterTables: whitelist tabel master yang dapat dicari lewat API.
var masterTables = map[string]struct {
	table, code, name, extra string
}{
	"icd10":  {"icd10_codes", "code", "name", "is_valid"},
	"icd9cm": {"icd9cm_codes", "code", "name", "null::text"},
	"loinc":  {"loinc_codes", "code", "name", "system"},
	"kfa":    {"kfa_drugs", "code", "name", "strength"},
	"kptl":   {"kptl_codes", "code", "name", "null::text"},
	"snomed": {"snomed_concepts", "concept_id", "preferred_term", "semantic_tag"},
}

// IsMasterKind melaporkan apakah kind adalah master yang dikenal.
func IsMasterKind(kind string) bool {
	_, ok := masterTables[kind]
	return ok
}

// SearchMaster mencari kode/nama di tabel master. Untuk SNOMED-CT,
// semanticTag menyaring jenis konsep dan kolom active ikut dikembalikan
// (konsep pensiun tampil tetapi tidak bisa dipilih).
func (s *Store) SearchMaster(ctx context.Context, kind, q, semanticTag string) ([]Row, error) {
	m := masterTables[kind]
	activeCol := "true"
	tagFilter := "true"
	args := []any{q}
	if kind == "snomed" {
		activeCol = "active"
		if semanticTag != "" {
			tagFilter = "semantic_tag = $2"
			args = append(args, semanticTag)
		}
	}
	sql := fmt.Sprintf(`
		select %[2]s as code, %[3]s as name, %[4]s as extra, %[5]s as active
		from %[1]s
		where (%[2]s ilike $1 || '%%' or %[3]s ilike '%%' || $1 || '%%') and %[6]s
		order by (%[2]s ilike $1 || '%%') desc, similarity(%[3]s, $1) desc
		limit 50`, m.table, m.code, m.name, m.extra, activeCol, tagFilter)
	return queryRows(ctx, s.pool, sql, args...)
}
