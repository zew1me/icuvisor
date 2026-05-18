package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesToolsCatalogGolden(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "data", "tools.json")
	if err := run([]string{"--out", out}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading generated catalog: %v", err)
	}
	want, err := os.ReadFile(filepath.Join("testdata", "tools.golden.json"))
	if err != nil {
		t.Fatalf("reading golden catalog: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("generated catalog differs from golden\n got: %s\nwant: %s", got, want)
	}
}
