package fusion

import "testing"

func TestComputeConfidence(t *testing.T) {
	tests := []struct {
		name        string
		successRate float64
		avgMs       int
		variance    float64
		wantMin     float64
		wantMax     float64
	}{
		{
			name:        "perfect success & fast response",
			successRate: 1.0,
			avgMs:       500,
			variance:    0.0,
			wantMin:     89.0,
			wantMax:     90.0,
		},
		{
			name:        "perfect success & slow response",
			successRate: 1.0,
			avgMs:       3000, // 1500 baseline + 1500 penalty = 10 penalty
			variance:    0.0,
			wantMin:     79.0,
			wantMax:     81.0,
		},
		{
			name:        "medium success rate",
			successRate: 0.5,
			avgMs:       1000,
			variance:    0.0,
			wantMin:     69.0,
			wantMax:     71.0,
		},
		{
			name:        "low success rate",
			successRate: 0.1,
			avgMs:       1000,
			variance:    0.0,
			wantMin:     53.0,
			wantMax:     55.0,
		},
		{
			name:        "zero success rate (floor)",
			successRate: 0.0,
			avgMs:       500,
			variance:    0.0,
			wantMin:     5.0, // floor
			wantMax:     50.0,
		},
		{
			name:        "high variance penalty",
			successRate: 1.0,
			avgMs:       500,
			variance:    100.0, // sqrt(100) = 10 penalty
			wantMin:     79.0,
			wantMax:     81.0,
		},
		{
			name:        "negative values (guard)",
			successRate: -1.0,
			avgMs:       500,
			variance:    0.0,
			wantMin:     5.0,
			wantMax:     50.0,
		},
		{
			name:        "overshoot success (guard)",
			successRate: 2.0,
			avgMs:       500,
			variance:    0.0,
			wantMin:     89.0,
			wantMax:     90.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeConfidence(tt.successRate, tt.avgMs, tt.variance)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("ComputeConfidence() = %v, want in [%v, %v]", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}
