package domain

import (
	"errors"
	"testing"
)

func TestReadinessBlocking(t *testing.T) {
	in := ReadinessInput{
		AcuanCount:         0,
		EpisodesBySeverity: map[int]int{1: 40, 2: 12, 3: 4},
		PlanRowsBySeverity: map[int]int{1: 5},
		// Syarat tidak menghambat yang belum terpenuhi tidak boleh menahan pengesahan.
		ProceduresWithoutKPTL: []string{"99.21"},
	}
	r := EvaluateReadiness(in)
	if r.Ready() {
		t.Fatal("tanpa acuan & rencana severity II, CP belum siap")
	}
	unmet := r.UnmetBlocking()
	if len(unmet) != 2 || unmet[0].Code != ReqAcuanCP || unmet[1].Code != ReqRencanaSeverity {
		t.Fatalf("dapat %+v", unmet)
	}
	if len(unmet[1].Missing) != 1 {
		t.Fatalf("hanya severity II yang wajib (III < %d episode): %v", MinEpisodesPerSeverity, unmet[1].Missing)
	}

	var nre *NotReadyError
	if err := CheckActivation(r); !errors.As(err, &nre) || len(nre.Unmet) != 2 {
		t.Fatalf("ingin NotReadyError dengan 2 syarat, dapat %v", err)
	}
}

func TestReadinessReady(t *testing.T) {
	r := EvaluateReadiness(ReadinessInput{
		AcuanCount:            1,
		EpisodesBySeverity:    map[int]int{1: 40, 2: 12},
		PlanRowsBySeverity:    map[int]int{1: 5, 2: 1},
		ProceduresWithoutKPTL: []string{"99.21"},
		InvalidICD10:          []string{"A01.9X"},
	})
	if !r.Ready() || CheckActivation(r) != nil {
		t.Fatal("syarat wajib terpenuhi — CP siap meski syarat anjuran belum lengkap")
	}
	if len(r.Requirements) != 7 {
		t.Fatalf("ingin 7 syarat, dapat %d", len(r.Requirements))
	}
}

func TestSubCPOnlyWhenSplit(t *testing.T) {
	r := EvaluateReadiness(ReadinessInput{AcuanCount: 1, SubCPsWithoutAcuan: []string{"A01.0"}})
	if !r.Requirements[1].Met {
		t.Fatal("acuan sub-CP hanya dinilai bila grouper perlu dipecah")
	}
	r = EvaluateReadiness(ReadinessInput{AcuanCount: 1, NeedsSplit: true, SubCPsWithoutAcuan: []string{"A01.0"}})
	if r.Requirements[1].Met || !r.Ready() {
		t.Fatal("sub-CP tanpa acuan tercatat, tetapi tidak menghambat")
	}
}
