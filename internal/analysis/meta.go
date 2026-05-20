package analysis

import (
	"sort"
	"strings"
)

const (
	// MissingActionSkip is the default missing-data policy for analyzer windows.
	MissingActionSkip = "skip"
	// MinBaselineSamples is the default minimum sample count for baseline analyzers.
	MinBaselineSamples = 7
	// MinCorrelationSamples is the default minimum sample count for correlation analyzers.
	MinCorrelationSamples = 14
	// SourceToolGetActivityIntervals is the read tool used for activity interval rows.
	SourceToolGetActivityIntervals = "get_activity_intervals"
	// IntervalExecutionDeclineAutoLapSuspected is the reason for declining structured interval-execution claims.
	IntervalExecutionDeclineAutoLapSuspected = "auto_lap_suspected"
)

// AnalyzerMeta is the mandatory analyzer _meta contract emitted by analyzer tools.
type AnalyzerMeta struct {
	Method             string         `json:"method"`
	SourceTools        []string       `json:"source_tools"`
	N                  int            `json:"n"`
	MissingDays        int            `json:"missing_days"`
	MissingAction      string         `json:"missing_action"`
	InsufficientSample bool           `json:"insufficient_sample"`
	FormulaRef         string         `json:"formula_ref,omitempty"`
	IntervalSource     IntervalSource `json:"interval_source,omitempty"`
	AutoLapSuspected   *bool          `json:"auto_lap_suspected,omitempty"`
}

// AnalyzerMetaInput describes analyzer metadata before normalization.
type AnalyzerMetaInput struct {
	Method           string
	SourceTools      []string
	N                int
	MissingDays      int
	MissingAction    string
	MinSamples       int
	FormulaRef       string
	IntervalSource   IntervalSource
	AutoLapSuspected *bool
}

// IntervalExecutionClaimDecision tells analyzers whether interval-execution claims are safe.
type IntervalExecutionClaimDecision struct {
	Decline bool
	Reason  string
}

// ApplyIntervalSourceEvidence attaches interval-source evidence to analyzer metadata input.
func ApplyIntervalSourceEvidence(input AnalyzerMetaInput, evidence IntervalSourceResult) AnalyzerMetaInput {
	input.SourceTools = append(input.SourceTools, SourceToolGetActivityIntervals)
	input.IntervalSource = evidence.Source
	input.AutoLapSuspected = boolPointer(evidence.AutoLapSuspected)
	return input
}

// IntervalExecutionClaimPolicy returns the interval-execution claim policy for source evidence.
func IntervalExecutionClaimPolicy(evidence IntervalSourceResult) IntervalExecutionClaimDecision {
	if evidence.AutoLapSuspected {
		return IntervalExecutionClaimDecision{Decline: true, Reason: IntervalExecutionDeclineAutoLapSuspected}
	}
	return IntervalExecutionClaimDecision{}
}

// NewAnalyzerMeta normalizes analyzer metadata while preserving mandatory fields.
func NewAnalyzerMeta(input AnalyzerMetaInput) AnalyzerMeta {
	n := input.N
	if n < 0 {
		n = 0
	}
	missingDays := input.MissingDays
	if missingDays < 0 {
		missingDays = 0
	}
	missingAction := strings.TrimSpace(input.MissingAction)
	if missingAction == "" {
		missingAction = MissingActionSkip
	}
	return AnalyzerMeta{
		Method:             strings.TrimSpace(input.Method),
		SourceTools:        NormalizeSourceTools(input.SourceTools),
		N:                  n,
		MissingDays:        missingDays,
		MissingAction:      missingAction,
		InsufficientSample: InsufficientSample(n, input.MinSamples),
		FormulaRef:         strings.TrimSpace(input.FormulaRef),
		IntervalSource:     input.IntervalSource,
		AutoLapSuspected:   input.AutoLapSuspected,
	}
}

// NormalizeSourceTools trims, deduplicates, and sorts source tool names.
func NormalizeSourceTools(sourceTools []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(sourceTools))
	for _, sourceTool := range sourceTools {
		trimmed := strings.TrimSpace(sourceTool)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

// InsufficientSample reports whether a usable sample count misses a minimum.
func InsufficientSample(n int, minN int) bool {
	return minN > 0 && n < minN
}

func boolPointer(value bool) *bool {
	return &value
}
