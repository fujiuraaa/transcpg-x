package domain

import (
	"fmt"
	"strings"
)

// RequirementCode adalah kode 7 syarat kelengkapan CP (tab Verifikasi Kode).
type RequirementCode string

const (
	ReqAcuanCP         RequirementCode = "ACUAN_CP"
	ReqAcuanSubCP      RequirementCode = "ACUAN_SUB_CP"
	ReqRencanaSeverity RequirementCode = "RENCANA_SEVERITY"
	ReqPadananKPTL     RequirementCode = "PADANAN_KPTL"
	ReqTarifINACBG     RequirementCode = "TARIF_INACBG"
	ReqPadananSnomed   RequirementCode = "PADANAN_SNOMED"
	ReqICD10Master     RequirementCode = "ICD10_MASTER"
)

// Requirement adalah satu syarat beserta status pemenuhannya.
type Requirement struct {
	No       int             `json:"no"`
	Code     RequirementCode `json:"code"`
	Title    string          `json:"title"`
	Owner    string          `json:"owner"`
	Blocking bool            `json:"blocking"`
	Met      bool            `json:"met"`
	Missing  []string        `json:"missing"`
}

// ReadinessInput adalah fakta tentang satu CP yang dikumpulkan store.
type ReadinessInput struct {
	AcuanCount              int
	NeedsSplit              bool
	SubCPsWithoutAcuan      []string
	EpisodesBySeverity      map[int]int
	PlanRowsBySeverity      map[int]int
	ProceduresWithoutKPTL   []string
	SeveritiesWithoutTariff []int
	DiagnosesWithoutSnomed  []string
	InvalidICD10            []string
}

// Readiness adalah hasil pemeriksaan 7 syarat.
type Readiness struct {
	Requirements []Requirement `json:"requirements"`
}

// Ready: seluruh syarat yang menghambat sudah terpenuhi.
func (r Readiness) Ready() bool { return len(r.UnmetBlocking()) == 0 }

// UnmetBlocking mengembalikan syarat wajib yang belum terpenuhi.
func (r Readiness) UnmetBlocking() []Requirement {
	var out []Requirement
	for _, q := range r.Requirements {
		if q.Blocking && !q.Met {
			out = append(out, q)
		}
	}
	return out
}

// EvaluateReadiness memeriksa ke-7 syarat sekaligus, supaya seluruh
// kekurangan bisa ditampilkan dalam satu kali percobaan.
func EvaluateReadiness(in ReadinessInput) Readiness {
	var missingPlan []string
	for _, s := range Severities {
		if in.EpisodesBySeverity[s] >= MinEpisodesPerSeverity && in.PlanRowsBySeverity[s] == 0 {
			missingPlan = append(missingPlan, fmt.Sprintf("Severity %s (%d episode)", SeverityLabel(s), in.EpisodesBySeverity[s]))
		}
	}
	var missingTariff []string
	for _, s := range in.SeveritiesWithoutTariff {
		missingTariff = append(missingTariff, "Severity "+SeverityLabel(s))
	}
	var missingAcuan []string
	if in.AcuanCount == 0 {
		missingAcuan = []string{"Belum ada dokumen panduan yang ditetapkan"}
	}
	var missingSub []string
	if in.NeedsSplit {
		missingSub = in.SubCPsWithoutAcuan
	}

	reqs := []Requirement{
		{No: 1, Code: ReqAcuanCP, Title: "Acuan standar CP (minimal 1 dokumen panduan)", Owner: "Tim CP · KSM", Blocking: true, Missing: missingAcuan},
		{No: 2, Code: ReqAcuanSubCP, Title: "Acuan tiap sub-CP (bila grouper perlu dipecah)", Owner: "Tim CP · KSM terkait", Missing: missingSub},
		{No: 3, Code: ReqRencanaSeverity, Title: fmt.Sprintf("Rencana isi klinis tiap severity yang cukup data (≥%d episode)", MinEpisodesPerSeverity), Owner: "Tim CP · DPJP · Farmasi", Blocking: true, Missing: missingPlan},
		{No: 4, Code: ReqPadananKPTL, Title: "Padanan KPTL tiap prosedur", Owner: "Tim CP (koding)", Missing: in.ProceduresWithoutKPTL},
		{No: 5, Code: ReqTarifINACBG, Title: "Tarif INA-CBG tiap severity", Owner: "Admin · data tarif Permenkes", Missing: missingTariff},
		{No: 6, Code: ReqPadananSnomed, Title: "Padanan SNOMED-CT tiap diagnosis", Owner: "Tim CP (koding)", Missing: in.DiagnosesWithoutSnomed},
		{No: 7, Code: ReqICD10Master, Title: "Kode ICD-10 sesuai master", Owner: "Tim CP (koding)", Missing: in.InvalidICD10},
	}
	for i := range reqs {
		reqs[i].Met = len(reqs[i].Missing) == 0
		if reqs[i].Missing == nil {
			reqs[i].Missing = []string{}
		}
	}
	return Readiness{Requirements: reqs}
}

// NotReadyError dikembalikan bila CP belum memenuhi syarat untuk Aktif.
type NotReadyError struct {
	Unmet []Requirement
}

func (e *NotReadyError) Error() string {
	titles := make([]string, len(e.Unmet))
	for i, q := range e.Unmet {
		titles[i] = q.Title
	}
	return "syarat aktif belum terpenuhi: " + strings.Join(titles, "; ")
}

// CheckActivation mengembalikan *NotReadyError bila ada syarat wajib yang
// belum terpenuhi.
func CheckActivation(r Readiness) error {
	if unmet := r.UnmetBlocking(); len(unmet) > 0 {
		return &NotReadyError{Unmet: unmet}
	}
	return nil
}
