package domain

import (
	"errors"
	"slices"
)

// ErrContentLocked: isi CP sedang dikunci karena ditinjau Komite/Direktur.
var ErrContentLocked = errors.New("isi CP dikunci selama Review Komite dan Menunggu Direktur — kembalikan ke Revisi untuk mengubah")

// CheckEditContent mengembalikan ErrContentLocked atau ErrNotAuthorized bila
// role tidak boleh mengubah isi kind pada tahap s.
func CheckEditContent(s Stage, kind ContentKind, role Role) error {
	if IsContentLocked(s) {
		return ErrContentLocked
	}
	if !CanEditContent(s, kind, role) {
		return ErrNotAuthorized
	}
	return nil
}

// Matriks wewenang. Sumber: alur-aplikasi.pdf (28 Sep 2026) §6, §7, §9, §13
// dan panduan-penggunaan §4. Bila dokumen berbeda, dokumen terbaru dipakai —
// lihat docs/STRUKTUR.md bagian "Pertanyaan terbuka".
var (
	adminsAll = []Role{RoleSuperAdmin, RoleSystemAdmin, RoleAdminRS}

	authors   = append([]Role{RoleDokter, RoleDPJP, RoleTimCP, RoleKoordinatorCP}, adminsAll...)
	timCP     = append([]Role{RoleTimCP, RoleKoordinatorCP}, adminsAll...)
	komite    = append([]Role{RoleKomiteMedik}, adminsAll...)
	direktur  = []Role{RoleDirektur, RoleSystemAdmin, RoleSuperAdmin}
	revokers  = []Role{RoleKomiteMedik, RoleDirektur, RoleSuperAdmin}
	planExtra = []Role{RoleApoteker} // khusus Rencana Klinis & Kriteria saat Draf/Revisi
)

// actorsByStage: siapa yang boleh memajukan/mengembalikan pada tiap tahap.
var actorsByStage = map[Stage][]Role{
	StageDraf:             authors,
	StageRevisi:           authors,
	StageReviewTimCP:      timCP,
	StageReviewKomite:     komite,
	StageMenungguDirektur: direktur,
	StageAktif:            revokers,
}

var waitingOn = map[Stage]string{
	StageDraf:             "Penyusun (Dokter/DPJP/Tim CP)",
	StageRevisi:           "Penyusun (Dokter/DPJP/Tim CP)",
	StageReviewTimCP:      "Tim CP",
	StageReviewKomite:     "KSM/Komite Medik",
	StageMenungguDirektur: "Direktur",
}

// CanAct melaporkan apakah role berwenang bertindak pada tahap s.
func CanAct(s Stage, role Role) bool {
	return slices.Contains(actorsByStage[s], role)
}

// WaitingOn adalah pihak yang sedang ditunggu tindakannya.
func WaitingOn(s Stage) string { return waitingOn[s] }

// StagesActionableBy mengembalikan tahap antrean Approval di mana role
// berwenang bertindak — isi tab "Menunggu Tindakan Saya".
func StagesActionableBy(role Role) []Stage {
	var out []Stage
	for _, s := range AllStages() {
		if s.InApprovalQueue() && CanAct(s, role) {
			out = append(out, s)
		}
	}
	return out
}

// ContentKind adalah jenis isi CP yang dapat disunting.
type ContentKind string

const (
	ContentAcuan    ContentKind = "ACUAN"
	ContentRencana  ContentKind = "RENCANA"
	ContentKriteria ContentKind = "KRITERIA"
)

// IsContentLocked: isi CP dikunci total saat Review Komite dan Menunggu
// Direktur, supaya yang disahkan Direktur sama persis dengan yang ditinjau Komite.
func IsContentLocked(s Stage) bool {
	return s == StageReviewKomite || s == StageMenungguDirektur
}

// CanEditContent melaporkan apakah role boleh mengubah isi kind pada tahap s.
func CanEditContent(s Stage, kind ContentKind, role Role) bool {
	switch s {
	case StageDraf, StageRevisi:
		if slices.Contains(authors, role) {
			return true
		}
		return kind != ContentAcuan && slices.Contains(planExtra, role)
	case StageReviewTimCP:
		return slices.Contains(timCP, role)
	case StageAktif:
		// Perubahan pada CP Aktif langsung berlaku dan wajib tercatat di
		// pathway_change_log (dilakukan oleh store).
		return slices.Contains(timCP, role)
	}
	return false
}

// CanManageMappings: hanya Tim CP dan Admin yang menetapkan/mencabut padanan
// KPTL dan SNOMED-CT; peran lain hanya melihat.
func CanManageMappings(role Role) bool {
	return slices.Contains(timCP, role)
}
