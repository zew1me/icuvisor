package safety

import (
	"log/slog"
	"strings"
)

const EnvToolset = "ICUVISOR_TOOLSET"

// Toolset controls which tool catalog tier is registered.
type Toolset string

const (
	// ToolsetCompact exposes the smallest model-compatible tool catalog.
	ToolsetCompact Toolset = "compact"
	// ToolsetCore exposes the curated daily-use tool catalog.
	ToolsetCore Toolset = "core"
	// ToolsetFull exposes the full tool catalog.
	ToolsetFull Toolset = "full"
)

// ParseToolset resolves raw ICUVISOR_TOOLSET values. Empty or unknown values
// intentionally fall back to core so misconfiguration preserves the token-saving default.
func ParseToolset(value string) Toolset {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ToolsetCompact):
		return ToolsetCompact
	case string(ToolsetFull):
		return ToolsetFull
	case string(ToolsetCore), "":
		return ToolsetCore
	default:
		return ToolsetCore
	}
}

// NewToolsetFromEnv resolves ICUVISOR_TOOLSET once and returns the resulting catalog tier.
func NewToolsetFromEnv(getenv func(string) string) Toolset {
	if getenv == nil {
		return ToolsetCore
	}
	return ParseToolset(getenv(EnvToolset))
}

// LogResolvedToolset emits the single startup log entry for the process toolset tier.
func LogResolvedToolset(logger *slog.Logger, toolset Toolset) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("resolved toolset", "toolset", toolset.String())
}

func (t Toolset) String() string {
	switch t {
	case ToolsetCompact, ToolsetFull, ToolsetCore:
		return string(t)
	default:
		return string(ToolsetCore)
	}
}
