package domain

import (
	"fmt"
	"math"
	"slices"
)

// Ambang "Perlu Ditinjau" pada halaman Evaluasi CP.
const (
	EvalShareAboveP75Threshold = 0.25 // ≥25% episode melebihi p75 target
	EvalSev3ShiftThreshold     = 0.10 // porsi severity III naik ≥10 poin persen
)

// FlagCode adalah alasan sebuah CP ditandai "Perlu Ditinjau".
type FlagCode string

const (
	FlagMedianAboveP75 FlagCode = "MEDIAN_LOS_DI_ATAS_P75"
	FlagManyAboveP75   FlagCode = "BANYAK_EPISODE_DI_ATAS_P75"
	FlagSev3Shift      FlagCode = "PERGESERAN_SEVERITY_III"
)

// Flag adalah satu temuan evaluasi.
type Flag struct {
	Code     FlagCode `json:"code"`
	Severity int      `json:"severity,omitempty"`
	Detail   string   `json:"detail"`
}

// LOSGroup membandingkan LOS sesudah CP aktif dengan target pada satu severity.
type LOSGroup struct {
	Severity    int
	BaselineP75 float64   // p75 LOS dari klaim sebelum CP aktif (target)
	PostLOS     []float64 // LOS episode sesudah CP aktif
}

// EvaluationInput: hanya klaim sesudah tanggal CP disahkan yang dibandingkan.
type EvaluationInput struct {
	Groups            []LOSGroup
	BaselineSev3Share float64 // 0..1
	PostSev3Share     float64 // 0..1
	PostEpisodes      int
}

// Evaluate menerapkan ketiga ambang dan mengembalikan temuan.
// Ini alat deteksi dini, bukan audit menyeluruh.
func Evaluate(in EvaluationInput) []Flag {
	var flags []Flag
	for _, g := range in.Groups {
		if len(g.PostLOS) == 0 || g.BaselineP75 <= 0 {
			continue
		}
		if m := Median(g.PostLOS); m > g.BaselineP75 {
			flags = append(flags, Flag{
				Code: FlagMedianAboveP75, Severity: g.Severity,
				Detail: fmt.Sprintf("Median LOS %.1f hari > p75 target %.1f hari", m, g.BaselineP75),
			})
		}
		above := 0
		for _, v := range g.PostLOS {
			if v > g.BaselineP75 {
				above++
			}
		}
		if share := float64(above) / float64(len(g.PostLOS)); share >= EvalShareAboveP75Threshold {
			flags = append(flags, Flag{
				Code: FlagManyAboveP75, Severity: g.Severity,
				Detail: fmt.Sprintf("%.0f%% episode melebihi p75 target %.1f hari", share*100, g.BaselineP75),
			})
		}
	}
	// Dibulatkan agar galat floating point tidak menggeser ambang tepat 10 poin.
	shift := math.Round((in.PostSev3Share-in.BaselineSev3Share)*10000) / 10000
	if in.PostEpisodes > 0 && shift >= EvalSev3ShiftThreshold {
		flags = append(flags, Flag{
			Code:   FlagSev3Shift,
			Detail: fmt.Sprintf("Porsi severity III naik %.1f poin persen — periksa kesesuaian pengkodean", shift*100),
		})
	}
	return flags
}

// Median dari xs (xs tidak diubah). Mengembalikan 0 untuk slice kosong.
func Median(xs []float64) float64 { return Percentile(xs, 0.5) }

// Percentile dengan interpolasi linear (sama dengan percentile_cont di PostgreSQL).
func Percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := slices.Clone(xs)
	slices.Sort(s)
	pos := p * float64(len(s)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	return s[lo] + (s[hi]-s[lo])*(pos-float64(lo))
}
