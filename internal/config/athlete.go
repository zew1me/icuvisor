package config

import (
	"errors"
	"strings"
	"unicode"
)

// NormalizeAthleteID accepts intervals.icu athlete IDs with or without the i prefix.
func NormalizeAthleteID(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("missing athlete ID; set INTERVALS_ICU_ATHLETE_ID or athlete_id")
	}

	if trimmed[0] == 'i' || trimmed[0] == 'I' {
		trimmed = trimmed[1:]
	}
	if trimmed == "" {
		return "", errors.New("invalid athlete ID; use 12345 or i12345")
	}
	for _, r := range trimmed {
		if !unicode.IsDigit(r) {
			return "", errors.New("invalid athlete ID; use 12345 or i12345")
		}
	}
	return "i" + trimmed, nil
}

// NormalizeAthleteIDForDisplay returns the canonical public athlete ID when possible.
func NormalizeAthleteIDForDisplay(value string) string {
	normalized, err := NormalizeAthleteID(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return normalized
}
