package domain

import (
	"errors"
	"testing"
)

func TestTransitionHappyPath(t *testing.T) {
	steps := []struct {
		from Stage
		role Role
		want Stage
	}{
		{StageDraf, RoleDPJP, StageReviewTimCP},
		{StageReviewTimCP, RoleTimCP, StageReviewKomite},
		{StageReviewKomite, RoleKomiteMedik, StageMenungguDirektur},
		{StageMenungguDirektur, RoleDirektur, StageAktif},
		{StageRevisi, RoleDokter, StageReviewTimCP},
	}
	for _, s := range steps {
		got, err := Transition(s.from, ActionAdvance, s.role, "")
		if err != nil || got != s.want {
			t.Errorf("%s oleh %s: dapat (%s, %v), ingin %s", s.from, s.role, got, err, s.want)
		}
	}
}

func TestTransitionAuthorization(t *testing.T) {
	cases := []struct {
		from Stage
		role Role
	}{
		{StageDraf, RoleApoteker},            // Apoteker hanya menyunting rencana, tidak mengajukan
		{StageReviewTimCP, RoleDPJP},         // DPJP bukan peninjau Tim CP
		{StageReviewKomite, RoleTimCP},       // Tim CP tidak bisa melompati Komite
		{StageMenungguDirektur, RoleAdminRS}, // Admin RS tidak mengesahkan
		{StageMenungguDirektur, RoleKomiteMedik},
	}
	for _, c := range cases {
		if _, err := Transition(c.from, ActionAdvance, c.role, ""); !errors.Is(err, ErrNotAuthorized) {
			t.Errorf("%s oleh %s: ingin ErrNotAuthorized, dapat %v", c.from, c.role, err)
		}
	}
}

func TestReturnRequiresReason(t *testing.T) {
	if _, err := Transition(StageReviewKomite, ActionReturn, RoleKomiteMedik, "  "); !errors.Is(err, ErrReasonRequired) {
		t.Fatalf("ingin ErrReasonRequired, dapat %v", err)
	}
	got, err := Transition(StageReviewKomite, ActionReturn, RoleKomiteMedik, "Dosis belum sesuai PNPK")
	if err != nil || got != StageRevisi {
		t.Fatalf("dapat (%s, %v)", got, err)
	}
}

func TestReturnNotAllowedFromDraft(t *testing.T) {
	if _, err := Transition(StageDraf, ActionReturn, RoleSuperAdmin, "x"); !errors.Is(err, ErrInvalidAction) {
		t.Fatalf("ingin ErrInvalidAction, dapat %v", err)
	}
	if _, err := Transition(StageAktif, ActionAdvance, RoleSuperAdmin, ""); !errors.Is(err, ErrInvalidAction) {
		t.Fatalf("AKTIF tidak bisa dimajukan, dapat %v", err)
	}
}

func TestRevokeActive(t *testing.T) {
	if _, err := Transition(StageAktif, ActionRevoke, RoleTimCP, "x"); !errors.Is(err, ErrNotAuthorized) {
		t.Fatalf("Tim CP tidak boleh mencabut, dapat %v", err)
	}
	got, err := Transition(StageAktif, ActionRevoke, RoleDirektur, "LOS jauh di atas target")
	if err != nil || got != StageRevisi {
		t.Fatalf("dapat (%s, %v)", got, err)
	}
}

func TestContentLock(t *testing.T) {
	for _, s := range []Stage{StageReviewKomite, StageMenungguDirektur} {
		for _, r := range AllRoles() {
			for _, k := range []ContentKind{ContentAcuan, ContentRencana, ContentKriteria} {
				if CanEditContent(s, k, r) {
					t.Errorf("%s: %s tidak boleh mengubah %s", s, r, k)
				}
			}
		}
	}
	if !CanEditContent(StageDraf, ContentRencana, RoleApoteker) {
		t.Error("Apoteker boleh menyusun rencana saat Draf")
	}
	if CanEditContent(StageDraf, ContentAcuan, RoleApoteker) {
		t.Error("Apoteker tidak menetapkan acuan")
	}
	if CanEditContent(StageReviewTimCP, ContentRencana, RoleDPJP) {
		t.Error("saat Review Tim CP hanya Tim CP/Admin")
	}
	if !CanEditContent(StageAktif, ContentRencana, RoleTimCP) {
		t.Error("Tim CP boleh mengubah CP Aktif (tercatat)")
	}
}

func TestStagesActionableBy(t *testing.T) {
	got := StagesActionableBy(RoleKomiteMedik)
	if len(got) != 1 || got[0] != StageReviewKomite {
		t.Fatalf("Komite: dapat %v", got)
	}
	if len(StagesActionableBy(RolePerawatPelaksana)) != 0 {
		t.Fatal("Perawat tidak punya antrean pengesahan")
	}
}

func TestSelfRegistrableRoles(t *testing.T) {
	for _, r := range []Role{RoleSuperAdmin, RoleSystemAdmin, RoleAdminRS, Role("X")} {
		if r.SelfRegistrable() {
			t.Errorf("%s tidak boleh diajukan sendiri", r)
		}
	}
	if !RoleDPJP.SelfRegistrable() || !RoleDirektur.SelfRegistrable() {
		t.Error("peran klinis/manajerial boleh diajukan (tetap menunggu persetujuan)")
	}
	if n := len(SelfRegistrableRoles()); n != len(AllRoles())-3 {
		t.Errorf("ingin %d peran, dapat %d", len(AllRoles())-3, n)
	}
}
