package domain

import (
	"errors"
	"strings"
)

// Stage adalah tahap pengesahan CP.
//
//	DRAF ──→ REVIEW_TIM_CP ──→ REVIEW_KOMITE ──→ MENUNGGU_DIREKTUR ──→ AKTIF
//	  ↑            │                 │                   │               │
//	  └─ REVISI ←──┴─────────────────┴───────────────────┴───────────────┘
type Stage string

const (
	StageDraf             Stage = "DRAF"
	StageRevisi           Stage = "REVISI"
	StageReviewTimCP      Stage = "REVIEW_TIM_CP"
	StageReviewKomite     Stage = "REVIEW_KOMITE"
	StageMenungguDirektur Stage = "MENUNGGU_DIREKTUR"
	StageAktif            Stage = "AKTIF"
)

// AllStages dalam urutan alur.
func AllStages() []Stage {
	return []Stage{StageDraf, StageRevisi, StageReviewTimCP, StageReviewKomite, StageMenungguDirektur, StageAktif}
}

var stageLabels = map[Stage]string{
	StageDraf:             "Draf",
	StageRevisi:           "Perlu Revisi",
	StageReviewTimCP:      "Review Tim CP",
	StageReviewKomite:     "Review KSM/Komite Medik",
	StageMenungguDirektur: "Menunggu Direktur",
	StageAktif:            "Aktif",
}

var advanceLabels = map[Stage]string{
	StageDraf:             "Ajukan ke Tim CP",
	StageRevisi:           "Ajukan ulang ke Tim CP",
	StageReviewTimCP:      "Setujui & teruskan ke Komite",
	StageReviewKomite:     "Setujui & teruskan ke Direktur",
	StageMenungguDirektur: "Sahkan menjadi CP Aktif",
}

var nextStage = map[Stage]Stage{
	StageDraf:             StageReviewTimCP,
	StageRevisi:           StageReviewTimCP,
	StageReviewTimCP:      StageReviewKomite,
	StageReviewKomite:     StageMenungguDirektur,
	StageMenungguDirektur: StageAktif,
}

func (s Stage) Valid() bool {
	_, ok := stageLabels[s]
	return ok
}

func (s Stage) Label() string {
	if l, ok := stageLabels[s]; ok {
		return l
	}
	return string(s)
}

// Next mengembalikan tahap berikutnya bila CP dimajukan.
func (s Stage) Next() (Stage, bool) {
	n, ok := nextStage[s]
	return n, ok
}

// AdvanceLabel adalah teks tombol "majukan" pada tahap s.
func (s Stage) AdvanceLabel() string { return advanceLabels[s] }

// InApprovalQueue: CP yang sudah diajukan dan belum Aktif. CP Draf yang
// belum pernah diajukan tidak muncul di halaman Approval.
func (s Stage) InApprovalQueue() bool {
	return s == StageRevisi || s == StageReviewTimCP || s == StageReviewKomite || s == StageMenungguDirektur
}

// Action adalah tindakan pada panel pengesahan.
type Action string

const (
	ActionAdvance Action = "MAJUKAN"    // maju satu tahap
	ActionReturn  Action = "KEMBALIKAN" // kembalikan untuk revisi (tahap review)
	ActionRevoke  Action = "CABUT"      // cabut keberlakuan CP Aktif → Revisi
)

var (
	ErrInvalidAction  = errors.New("tindakan tidak berlaku pada tahap ini")
	ErrNotAuthorized  = errors.New("peran Anda tidak berwenang pada tahap ini")
	ErrReasonRequired = errors.New("alasan wajib diisi")
)

// Transition menghitung tahap tujuan dari tindakan action oleh peran role.
// Syarat aktif (lihat Readiness) tidak diperiksa di sini — pemanggil wajib
// memanggil CheckActivation bila tahap tujuan adalah AKTIF.
func Transition(from Stage, action Action, role Role, reason string) (Stage, error) {
	switch action {
	case ActionAdvance:
		to, ok := from.Next()
		if !ok {
			return "", ErrInvalidAction
		}
		if !CanAct(from, role) {
			return "", ErrNotAuthorized
		}
		return to, nil

	case ActionReturn:
		if from != StageReviewTimCP && from != StageReviewKomite && from != StageMenungguDirektur {
			return "", ErrInvalidAction
		}
		if !CanAct(from, role) {
			return "", ErrNotAuthorized
		}
		if strings.TrimSpace(reason) == "" {
			return "", ErrReasonRequired
		}
		return StageRevisi, nil

	case ActionRevoke:
		if from != StageAktif {
			return "", ErrInvalidAction
		}
		if !CanAct(from, role) {
			return "", ErrNotAuthorized
		}
		if strings.TrimSpace(reason) == "" {
			return "", ErrReasonRequired
		}
		return StageRevisi, nil
	}
	return "", ErrInvalidAction
}

// AvailableActions mengembalikan tindakan yang boleh dilakukan role pada
// tahap s — dipakai frontend untuk menampilkan tombol.
func AvailableActions(s Stage, role Role) []Action {
	if !CanAct(s, role) {
		return nil
	}
	switch s {
	case StageDraf, StageRevisi:
		return []Action{ActionAdvance}
	case StageReviewTimCP, StageReviewKomite, StageMenungguDirektur:
		return []Action{ActionAdvance, ActionReturn}
	case StageAktif:
		return []Action{ActionRevoke}
	}
	return nil
}
