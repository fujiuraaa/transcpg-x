package domain

import "testing"

func TestManageableBy(t *testing.T) {
	cases := []struct {
		target, actor Role
		want          bool
	}{
		{RoleTimCP, RoleAdminRS, true},
		{RoleAdminRS, RoleAdminRS, false}, // sesama Admin RS tidak boleh saling ambil alih
		{RoleAdminRS, RoleSystemAdmin, true},
		{RoleSystemAdmin, RoleSystemAdmin, false},
		{RoleSystemAdmin, RoleSuperAdmin, true},
		{RoleSuperAdmin, RoleSuperAdmin, true},
		{RoleDokter, RoleTimCP, false}, // bukan admin
	}
	for _, c := range cases {
		if got := c.target.ManageableBy(c.actor); got != c.want {
			t.Errorf("%s oleh %s = %v, ingin %v", c.target, c.actor, got, c.want)
		}
	}
}
