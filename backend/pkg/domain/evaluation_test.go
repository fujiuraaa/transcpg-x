package domain

import "testing"

func TestPercentile(t *testing.T) {
	xs := []float64{5, 1, 3, 2, 4}
	if m := Median(xs); m != 3 {
		t.Fatalf("median: %v", m)
	}
	if p := Percentile([]float64{1, 2, 3, 4}, 0.75); p != 3.25 {
		t.Fatalf("p75: %v", p)
	}
	if xs[0] != 5 {
		t.Fatal("input tidak boleh diurutkan di tempat")
	}
}

func TestEvaluateFlags(t *testing.T) {
	flags := Evaluate(EvaluationInput{
		Groups: []LOSGroup{
			{Severity: 1, BaselineP75: 4, PostLOS: []float64{3, 3, 4, 5}}, // 25% di atas p75 → tanda
			{Severity: 2, BaselineP75: 6, PostLOS: []float64{7, 8, 9}},    // median > p75 dan 100% di atas
			{Severity: 3, BaselineP75: 9, PostLOS: []float64{}},           // belum ada klaim
		},
		BaselineSev3Share: 0.10,
		PostSev3Share:     0.20,
		PostEpisodes:      7,
	})
	want := []struct {
		code FlagCode
		sev  int
	}{
		{FlagManyAboveP75, 1},
		{FlagMedianAboveP75, 2},
		{FlagManyAboveP75, 2},
		{FlagSev3Shift, 0},
	}
	if len(flags) != len(want) {
		t.Fatalf("dapat %+v", flags)
	}
	for i, w := range want {
		if flags[i].Code != w.code || flags[i].Severity != w.sev {
			t.Errorf("flag %d: dapat %+v, ingin %v/%d", i, flags[i], w.code, w.sev)
		}
	}
}

func TestEvaluateNoFlags(t *testing.T) {
	flags := Evaluate(EvaluationInput{
		Groups:            []LOSGroup{{Severity: 1, BaselineP75: 5, PostLOS: []float64{3, 4, 4, 5}}},
		BaselineSev3Share: 0.10,
		PostSev3Share:     0.19,
		PostEpisodes:      4,
	})
	if len(flags) != 0 {
		t.Fatalf("tidak ingin tanda, dapat %+v", flags)
	}
}
