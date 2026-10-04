// Package domain berisi aturan bisnis TransCPG-X yang tidak bergantung pada
// database maupun HTTP: peran, tahap pengesahan, hak akses, syarat aktif,
// dan ambang evaluasi.
package domain

// Role adalah peran pengguna yang dikenal sistem.
type Role string

const (
	RoleSuperAdmin       Role = "SUPER_ADMIN"
	RoleSystemAdmin      Role = "SYSTEM_ADMIN"
	RoleAdminRS          Role = "ADMIN_RS"
	RoleDokter           Role = "DOKTER"
	RoleDPJP             Role = "DPJP"
	RoleDokterUmum       Role = "DOKTER_UMUM"
	RoleResiden          Role = "RESIDEN"
	RoleKoordinatorCP    Role = "KOORDINATOR_CP"
	RoleTimCP            Role = "TIM_CP"
	RolePerawatPelaksana Role = "PERAWAT_PELAKSANA"
	RoleKepalaPerawat    Role = "KEPALA_PERAWAT"
	RoleApoteker         Role = "APOTEKER"
	RoleDietisien        Role = "DIETISIEN"
	RoleKomiteMedik      Role = "KOMITE_MEDIK"
	RoleDirektur         Role = "DIREKTUR"
)

var roleLabels = map[Role]string{
	RoleSuperAdmin:       "Super Admin",
	RoleSystemAdmin:      "System Admin",
	RoleAdminRS:          "Admin RS",
	RoleDokter:           "Dokter",
	RoleDPJP:             "DPJP",
	RoleDokterUmum:       "Dokter Umum",
	RoleResiden:          "Residen (PPDS)",
	RoleKoordinatorCP:    "Koordinator CP",
	RoleTimCP:            "Tim CP",
	RolePerawatPelaksana: "Perawat Pelaksana",
	RoleKepalaPerawat:    "Kepala Perawat",
	RoleApoteker:         "Apoteker",
	RoleDietisien:        "Dietisien/Ahli Gizi",
	RoleKomiteMedik:      "KSM/Komite Medik",
	RoleDirektur:         "Direktur",
}

// AllRoles mengembalikan seluruh peran dalam urutan tampilan.
func AllRoles() []Role {
	return []Role{
		RoleSuperAdmin, RoleSystemAdmin, RoleAdminRS,
		RoleDokter, RoleDPJP, RoleDokterUmum, RoleResiden,
		RoleKoordinatorCP, RoleTimCP,
		RolePerawatPelaksana, RoleKepalaPerawat, RoleApoteker, RoleDietisien,
		RoleKomiteMedik, RoleDirektur,
	}
}

// Valid melaporkan apakah r adalah peran yang dikenal.
func (r Role) Valid() bool {
	_, ok := roleLabels[r]
	return ok
}

// Label mengembalikan nama peran untuk ditampilkan.
func (r Role) Label() string {
	if l, ok := roleLabels[r]; ok {
		return l
	}
	return string(r)
}

// IsAdmin melaporkan apakah peran boleh membuka menu Pengaturan.
func (r Role) IsAdmin() bool {
	return r == RoleAdminRS || r == RoleSystemAdmin || r == RoleSuperAdmin
}

// AssignableBy melaporkan apakah admin dengan peran actor boleh memberikan
// peran r kepada pengguna lain. Admin tidak bisa memberikan peran yang lebih
// tinggi dari dirinya: Super Admin hanya oleh Super Admin, System Admin hanya
// oleh System/Super Admin (System Admin bisa mengesahkan CP di tahap Direktur).
func (r Role) AssignableBy(actor Role) bool {
	if !r.Valid() || !actor.IsAdmin() {
		return false
	}
	switch r {
	case RoleSuperAdmin:
		return actor == RoleSuperAdmin
	case RoleSystemAdmin:
		return actor == RoleSuperAdmin || actor == RoleSystemAdmin
	}
	return true
}

// adminRank: tingkat kewenangan admin (non-admin = 0).
func (r Role) adminRank() int {
	switch r {
	case RoleSuperAdmin:
		return 3
	case RoleSystemAdmin:
		return 2
	case RoleAdminRS:
		return 1
	}
	return 0
}

// ManageableBy melaporkan apakah admin berperan actor boleh mengubah,
// mengatur ulang kata sandi, menonaktifkan, atau menghapus akun berperan r.
// Lebih ketat dari AssignableBy: akun admin setingkat tidak boleh saling
// mengambil alih (mis. Admin RS mengganti sandi Admin RS lain). Hanya Super
// Admin yang boleh mengelola sesama Super Admin.
func (r Role) ManageableBy(actor Role) bool {
	if !r.Valid() || !actor.IsAdmin() {
		return false
	}
	if actor == RoleSuperAdmin {
		return true
	}
	return r.adminRank() < actor.adminRank()
}
