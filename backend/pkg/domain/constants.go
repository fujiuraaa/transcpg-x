package domain

const (
	// MinEpisodesEligible: grouper dengan episode rawat inap di bawah angka ini
	// "dilewati" — belum layak disusun menjadi CP.
	MinEpisodesEligible = 5

	// MinEpisodesPerSeverity: severity dengan episode sebanyak ini dianggap
	// cukup data — target LOS bisa ditetapkan dan rencana klinisnya wajib diisi.
	MinEpisodesPerSeverity = 10

	// MaxPlanDay: hari rawat terakhir di rencana isi klinis (H0..H14).
	MaxPlanDay = 14

	// LibraryPageSize: jumlah CP per halaman di CP Library.
	LibraryPageSize = 25
)

// Severities adalah tingkat keparahan INA-CBG: I, II, III.
var Severities = []int{1, 2, 3}

// SeverityLabel mengubah 1/2/3 menjadi I/II/III.
func SeverityLabel(s int) string {
	switch s {
	case 1:
		return "I"
	case 2:
		return "II"
	case 3:
		return "III"
	}
	return "?"
}
