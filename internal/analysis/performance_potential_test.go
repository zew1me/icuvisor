package analysis

import "testing"

func TestPerformancePotentialSportFamilyCapabilities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		sport  string
		family string
		power  bool
		pace   bool
		hr     bool
	}{
		{sport: "Ride", family: "cycling", power: true, hr: true},
		{sport: "Trail Run", family: "running", pace: true, hr: true},
		{sport: "Open Water Swim", family: "swimming", pace: true, hr: true},
		{sport: "Rowing", family: "rowing", pace: true, hr: true},
		{sport: "Strength", family: "other"},
	}
	for _, tc := range tests {
		t.Run(tc.sport, func(t *testing.T) {
			t.Parallel()
			if got := PerformancePotentialSportFamily(tc.sport); got != tc.family {
				t.Fatalf("PerformancePotentialSportFamily(%q) = %q, want %q", tc.sport, got, tc.family)
			}
			if got := PerformancePotentialSupportsPower(tc.sport); got != tc.power {
				t.Fatalf("PerformancePotentialSupportsPower(%q) = %v, want %v", tc.sport, got, tc.power)
			}
			if got := PerformancePotentialSupportsPace(tc.sport); got != tc.pace {
				t.Fatalf("PerformancePotentialSupportsPace(%q) = %v, want %v", tc.sport, got, tc.pace)
			}
			if got := PerformancePotentialSupportsHeartRate(tc.sport); got != tc.hr {
				t.Fatalf("PerformancePotentialSupportsHeartRate(%q) = %v, want %v", tc.sport, got, tc.hr)
			}
		})
	}
}
