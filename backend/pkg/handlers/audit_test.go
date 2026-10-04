package handlers

import (
	"testing"

	"transcpg-x/backend/pkg/domain"
)

func TestCSVSafe(t *testing.T) {
	got := csvSafe("=HYPERLINK(\"x\")", "+62", "@a", "-1", "biasa", "")
	want := []string{"'=HYPERLINK(\"x\")", "'+62", "'@a", "'-1", "biasa", ""}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("sel %d: %q, ingin %q", i, got[i], want[i])
		}
	}
}

func TestDescribeUserChange(t *testing.T) {
	before := domain.User{FullName: "A", Role: domain.RoleTimCP, IsActive: true}
	after := before
	after.Role, after.IsActive = domain.RoleKoordinatorCP, false
	changes, detail := describeUserChange(before, after, true)
	if len(changes) != 3 || detail["peran_baru"] != "Koordinator CP" {
		t.Fatalf("ringkasan salah: %v %v", changes, detail)
	}
	if userChangeAction(before, after) != "NONAKTIFKAN" {
		t.Error("menonaktifkan harus didahulukan")
	}
	if c, _ := describeUserChange(before, before, false); len(c) != 0 {
		t.Error("tanpa perubahan tidak boleh dicatat")
	}
}
