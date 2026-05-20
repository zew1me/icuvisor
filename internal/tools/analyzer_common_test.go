package tools

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func assertAnalyzerGolden(t *testing.T, fixture string, got any) {
	t.Helper()
	gotJSON := canonicalAnalyzerJSON(t, got)
	path := filepath.Join("testdata", fixture)
	if os.Getenv("UPDATE_ANALYZER_GOLDENS") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create analyzer golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(gotJSON), 0o644); err != nil {
			t.Fatalf("update analyzer golden %s: %v", path, err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read analyzer golden %s: %v", path, err)
	}
	if gotJSON != string(bytes.TrimSpace(want))+"\n" {
		t.Fatalf("analyzer golden %s mismatch\n got: %s\nwant: %s", path, gotJSON, string(want))
	}
}

func canonicalAnalyzerJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal analyzer response: %v", err)
	}
	return string(data) + "\n"
}
