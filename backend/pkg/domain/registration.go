package domain

// RegistrationStatus: status pendaftaran akun.
type RegistrationStatus string

const (
	RegPending  RegistrationStatus = "MENUNGGU"
	RegApproved RegistrationStatus = "DISETUJUI"
	RegRejected RegistrationStatus = "DITOLAK"
)

// SelfRegistrable: peran yang boleh DIAJUKAN sendiri saat mendaftar.
// Peran admin hanya bisa diberikan oleh admin. Peran yang diajukan tetap
// baru berlaku setelah Admin RS menyetujui (dan boleh mengubahnya).
func (r Role) SelfRegistrable() bool {
	return r.Valid() && !r.IsAdmin()
}

// SelfRegistrableRoles dalam urutan tampilan formulir daftar.
func SelfRegistrableRoles() []Role {
	var out []Role
	for _, r := range AllRoles() {
		if r.SelfRegistrable() {
			out = append(out, r)
		}
	}
	return out
}
