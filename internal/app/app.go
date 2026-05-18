package app

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

// Options contains process-level dependencies for the icuvisor CLI.
type Options struct {
	Version string
	Args    []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer

	LoadConfig  func(context.Context, config.Options) (config.Config, error)
	StartServer func(context.Context, ServerInfo) error

	DiagnosticsRecentToolCallsPath string

	SetupRunner           SetupRunner
	SetupCredentialStore  credstore.Store
	SetupPrompter         SetupPrompter
	SetupConfigExists     func(string) (bool, error)
	SetupConfigWriter     SetupConfigWriter
	SetupProfileFetcher   SetupProfileFetcher
	SetupTimezoneDetector func() string
}

// ServerInfo carries process metadata needed by lower layers.
type ServerInfo struct {
	Version       string
	Config        config.Config
	DebugMetadata bool
	DeleteMode    safety.Mode
	Toolset       safety.Toolset
	Capability    safety.Capability
}

// RunCLI executes the icuvisor CLI, writes any error to opts.Stderr, and returns a process exit code.
func RunCLI(ctx context.Context, opts Options) int {
	stderr := opts.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	if err := Run(ctx, opts); err != nil {
		_, _ = fmt.Fprintf(stderr, "icuvisor: %v\n", err)
		return ExitCode(err)
	}
	return 0
}

// Run executes the icuvisor CLI.
func Run(ctx context.Context, opts Options) error {
	version := opts.Version
	if version == "" {
		version = "dev"
	}

	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}

	args := opts.Args
	if len(args) > 0 && args[0] == "setup" {
		return runSetupCommand(ctx, opts, args[1:])
	}
	if len(args) > 0 && args[0] == "diagnostics" {
		return runDiagnosticsCommand(ctx, opts, args[1:])
	}
	if helpRequested(args) {
		if hasCommand(args, "version") {
			return writeVersionHelp(stdout)
		}
		if hasCommand(args, "diagnostics") {
			return writeDiagnosticsHelp(stdout)
		}
		return writeTopLevelHelp(stdout)
	}
	if len(args) > 0 && args[0] == "version" {
		_, err := fmt.Fprintln(stdout, version)
		if err != nil {
			return fmt.Errorf("writing version: %w", err)
		}
		return nil
	}

	configOpts, err := parseDefaultArgs(args)
	if err != nil {
		return err
	}

	return startServer(ctx, opts.LoadConfig, opts.StartServer, ServerInfo{Version: version}, configOpts)
}

// UsageError reports invalid CLI input that should exit with code 2.
type UsageError struct {
	message string
}

func (e UsageError) Error() string {
	return e.message
}

// ExitCode maps Run errors to process exit codes.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var usageErr UsageError
	if errors.As(err, &usageErr) {
		return 2
	}
	return 1
}

func newUsageError(format string, args ...any) UsageError {
	return UsageError{message: fmt.Sprintf(format, args...) + "\nRun 'icuvisor --help' for usage."}
}

func helpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

func hasCommand(args []string, command string) bool {
	for _, arg := range args {
		if arg == command {
			return true
		}
	}
	return false
}

func startServer(ctx context.Context, loader func(context.Context, config.Options) (config.Config, error), starter func(context.Context, ServerInfo) error, info ServerInfo, configOpts config.Options) error {
	if loader == nil {
		loader = config.Load
		if configOpts.CredentialStore == nil {
			configOpts.CredentialStore = credstore.OSKeychain()
		}
	}
	cfg, err := loader(ctx, configOpts)
	if err != nil {
		return err
	}
	info.Config = cfg
	info.DebugMetadata = cfg.DebugMetadata
	info.DeleteMode = cfg.DeleteMode
	info.Toolset = cfg.Toolset
	info.Capability = safety.NewCapability(cfg.DeleteMode)

	if starter == nil {
		starter = defaultStartServer
	}
	if err := starter(ctx, info); err != nil {
		return err
	}
	return nil
}
