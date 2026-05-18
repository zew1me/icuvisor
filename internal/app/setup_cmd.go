package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/cli/prompt"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
)

type setupArgs struct {
	configPath string
	offline    bool
	force      bool
	help       bool
}

func runSetupCommand(ctx context.Context, opts Options, args []string) error {
	parsed, err := parseSetupArgs(args)
	if err != nil {
		return err
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	if parsed.help {
		return writeSetupHelp(stdout)
	}
	path, err := resolveSetupConfigPath(parsed.configPath)
	if err != nil {
		return err
	}
	runner := opts.SetupRunner
	if runner == nil {
		runner = RunSetup
	}
	store := opts.SetupCredentialStore
	if store == nil {
		store = credstore.OSKeychain()
	}
	prompter := opts.SetupPrompter
	if prompter == nil {
		prompter = prompt.NewTerminal(opts.Stdin, stdout)
	}
	return runner(ctx, SetupOptions{
		ConfigPath:       path,
		Offline:          parsed.offline,
		Force:            parsed.force,
		Stdout:           stdout,
		Stderr:           stderr,
		CredentialStore:  store,
		Prompter:         prompter,
		ConfigExists:     opts.SetupConfigExists,
		ConfigWriter:     opts.SetupConfigWriter,
		ProfileFetcher:   opts.SetupProfileFetcher,
		TimezoneDetector: opts.SetupTimezoneDetector,
	})
}

func parseSetupArgs(args []string) (setupArgs, error) {
	var parsed setupArgs
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h" || arg == "help":
			parsed.help = true
		case arg == "--config":
			value, next, err := requireFlagValue(args, i, "--config", "/path/to/icuvisor.json")
			if err != nil {
				return setupArgs{}, setupUsageError(err)
			}
			parsed.configPath = value
			i = next
		case strings.HasPrefix(arg, "--config="):
			value, err := requireInlineFlagValue(arg, "--config", "/path/to/icuvisor.json")
			if err != nil {
				return setupArgs{}, setupUsageError(err)
			}
			parsed.configPath = value
		case arg == "--offline":
			parsed.offline = true
		case arg == "--force":
			parsed.force = true
		default:
			return setupArgs{}, newSetupUsageError("unknown setup flag %q", unknownSetupFlagName(arg))
		}
	}
	return parsed, nil
}

func unknownSetupFlagName(arg string) string {
	if strings.HasPrefix(arg, "--") {
		name, _, hasValue := strings.Cut(arg, "=")
		if hasValue {
			return name
		}
	}
	return arg
}

func setupUsageError(err error) error {
	var usageErr UsageError
	if errors.As(err, &usageErr) {
		msg := strings.TrimSuffix(usageErr.message, "\nRun 'icuvisor --help' for usage.")
		return newSetupUsageError("%s", msg)
	}
	return err
}

func newSetupUsageError(format string, args ...any) UsageError {
	return UsageError{message: fmt.Sprintf(format, args...) + "\nRun 'icuvisor setup --help' for usage."}
}

func resolveSetupConfigPath(path string) (string, error) {
	if trimmed := strings.TrimSpace(path); trimmed != "" {
		return trimmed, nil
	}
	if envPath := strings.TrimSpace(os.Getenv(config.EnvConfigPath)); envPath != "" {
		return envPath, nil
	}
	resolved, err := config.DefaultPath()
	if err != nil {
		return "", fmt.Errorf("resolve default config path: %w", err)
	}
	return resolved, nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
