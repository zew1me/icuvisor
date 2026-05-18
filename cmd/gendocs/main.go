package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ricardocabral/icuvisor/internal/tools"
)

const defaultToolsOut = "web/data/tools.json"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "gendocs: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("gendocs", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	out := flags.String("out", defaultToolsOut, "path to write generated tool catalog JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	return writeToolsCatalog(*out, tools.Catalog())
}

func writeToolsCatalog(out string, catalog []tools.ToolDescriptor) error {
	if out == "" {
		return fmt.Errorf("missing --out path")
	}
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tool catalog: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(out)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(out)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, out); err != nil {
		return fmt.Errorf("replace output file: %w", err)
	}
	return nil
}
