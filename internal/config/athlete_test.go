package config

import (
	"testing"
)

func TestNormalizeAthleteIDForDisplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "numeric", input: "12345", want: "i12345"},
		{name: "prefixed", input: "i12345", want: "i12345"},
		{name: "invalid", input: " athlete ", want: "athlete"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeAthleteIDForDisplay(tc.input); got != tc.want {
				t.Fatalf("NormalizeAthleteIDForDisplay(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeAthleteID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "digits", input: "12345", want: "i12345"},
		{name: "prefixed", input: "i12345", want: "i12345"},
		{name: "uppercase prefix", input: "I12345", want: "i12345"},
		{name: "trim spaces", input: " 12345 ", want: "i12345"},
		{name: "empty", input: "", wantErr: true},
		{name: "prefix only", input: "i", wantErr: true},
		{name: "letters", input: "abc", wantErr: true},
		{name: "mixed", input: "i12x", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeAthleteID(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("NormalizeAthleteID() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeAthleteID() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeAthleteID() = %q, want %q", got, tc.want)
			}
		})
	}
}
