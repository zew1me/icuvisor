package config

import (
	"errors"
	"strings"
	"unicode"
)

// NormalizeAthleteID validates intervals.icu athlete IDs in their canonical `i<digits>` form.
// intervals.icu always displays athlete IDs with a leading `i` in URLs; the bare-numeric form
// is rejected so misconfigured callers fail loudly instead of silently fanning out to a
// nonexistent athlete.
func NormalizeAthleteID(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("missing athlete ID; set INTERVALS_ICU_ATHLETE_ID or athlete_id (find it in the intervals.icu URL, e.g. i12345)")
	}

	if trimmed[0] != 'i' && trimmed[0] != 'I' {
		return "", errors.New("invalid athlete ID; intervals.icu IDs start with 'i' followed by digits, e.g. i12345 (find yours in the intervals.icu URL)")
	}
	digits := trimmed[1:]
	if digits == "" {
		return "", errors.New("invalid athlete ID; intervals.icu IDs start with 'i' followed by digits, e.g. i12345")
	}
	for _, r := range digits {
		if !unicode.IsDigit(r) {
			return "", errors.New("invalid athlete ID; intervals.icu IDs start with 'i' followed by digits, e.g. i12345")
		}
	}
	return "i" + digits, nil
}

// NormalizeAthleteIDForDisplay returns the canonical public athlete ID when possible.
func NormalizeAthleteIDForDisplay(value string) string {
	normalized, err := NormalizeAthleteID(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return normalized
}
