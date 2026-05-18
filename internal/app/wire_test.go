package app

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/clients"
	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/diagnostics"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	mcpserver "github.com/ricardocabral/icuvisor/internal/mcp"
	"github.com/ricardocabral/icuvisor/internal/prompts"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

type nopToolRegistry struct{}

func (nopToolRegistry) Register(context.Context, tools.Registrar) error { return nil }

type nopResourceRegistry struct{}

func (nopResourceRegistry) Register(context.Context, resources.Registrar) error { return nil }

type nopPromptRegistry struct{}

func (nopPromptRegistry) Register(context.Context, prompts.Registrar) error { return nil }

type fakeRecentToolCallRecorder struct{}

func (fakeRecentToolCallRecorder) RecordToolCall(context.Context, string, time.Time) error {
	return nil
}

func TestWireServerUsesInjectedCollaborators(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fakeClient := &intervals.Client{}
	toolRegistry := nopToolRegistry{}
	resourceRegistry := nopResourceRegistry{}
	promptRegistry := nopPromptRegistry{}
	recorder := fakeRecentToolCallRecorder{}

	var gotClientOptions intervals.Options
	var gotResourceClient clients.ProfileClient
	var gotResourceOptions resources.ResourceOptions
	var gotToolClient *intervals.Client
	var gotToolOptions tools.RegistryOptions
	var gotDefaultAthleteID string
	var gotServerOptions mcpserver.Options

	server, cleanup, err := wireServer(ctx, ServerInfo{
		Version:       "  ",
		DebugMetadata: true,
		DeleteMode:    safety.ModeFull,
		Toolset:       safety.ToolsetFull,
		Config: config.Config{
			Timezone: "America/Sao_Paulo",
			Coach:    coach.Config{DefaultAthleteID: "i42"},
		},
	}, deps{
		Logger: func() *slog.Logger { return logger },
		NewClient: func(opts intervals.Options) (*intervals.Client, error) {
			gotClientOptions = opts
			return fakeClient, nil
		},
		NewSelectionStore: func(defaultAthleteID string) *coach.SelectionStore {
			gotDefaultAthleteID = defaultAthleteID
			return coach.NewSelectionStore(defaultAthleteID)
		},
		RecentToolCallRecorder: func() (diagnostics.RecentToolCallRecorder, error) {
			return recorder, nil
		},
		NewPromptRegistry: func() prompts.Registry { return promptRegistry },
		NewResourceRegistry: func(client clients.ProfileClient, opts resources.ResourceOptions) resources.Registry {
			gotResourceClient = client
			gotResourceOptions = opts
			return resourceRegistry
		},
		NewToolRegistry: func(client *intervals.Client, opts tools.RegistryOptions) tools.Registry {
			gotToolClient = client
			gotToolOptions = opts
			return toolRegistry
		},
		NewServer: func(_ context.Context, opts mcpserver.Options) (*mcpserver.Server, error) {
			gotServerOptions = opts
			return &mcpserver.Server{}, nil
		},
	})
	if err != nil {
		t.Fatalf("wireServer() error = %v", err)
	}
	if server == nil {
		t.Fatal("wireServer() server = nil")
	}
	if cleanup == nil {
		t.Fatal("wireServer() cleanup = nil")
	}
	cleanup()

	if gotClientOptions.Version != "dev" {
		t.Fatalf("client version = %q, want dev", gotClientOptions.Version)
	}
	if gotClientOptions.Config.Timezone != "America/Sao_Paulo" {
		t.Fatalf("client timezone = %q", gotClientOptions.Config.Timezone)
	}
	if gotDefaultAthleteID != "i42" {
		t.Fatalf("selection store default athlete = %q, want i42", gotDefaultAthleteID)
	}
	if gotResourceClient != fakeClient {
		t.Fatal("resource registry did not receive injected client")
	}
	if gotResourceOptions.Version != "dev" || gotResourceOptions.TimezoneFallback != "America/Sao_Paulo" || !gotResourceOptions.DebugMetadata {
		t.Fatalf("resource options = %+v", gotResourceOptions)
	}
	if gotResourceOptions.DeleteMode != safety.ModeFull || gotResourceOptions.Toolset != safety.ToolsetFull {
		t.Fatalf("resource safety options = %+v", gotResourceOptions)
	}
	if gotToolClient != fakeClient {
		t.Fatal("tool registry did not receive injected client")
	}
	if gotToolOptions.Version != "dev" || gotToolOptions.TimezoneFallback != "America/Sao_Paulo" || !gotToolOptions.DebugMetadata {
		t.Fatalf("tool options = %+v", gotToolOptions)
	}
	if gotToolOptions.Toolset != safety.ToolsetFull || gotToolOptions.CoachConfig.DefaultAthleteID != "i42" {
		t.Fatalf("tool safety/coach options = %+v", gotToolOptions)
	}
	if gotServerOptions.Version != "dev" || gotServerOptions.Logger != logger {
		t.Fatalf("server options version/logger = %q/%v", gotServerOptions.Version, gotServerOptions.Logger)
	}
	if gotServerOptions.Registry != toolRegistry || gotServerOptions.ResourceRegistry != resourceRegistry || gotServerOptions.PromptRegistry != promptRegistry {
		t.Fatalf("server registries not wired from injected factories")
	}
	if gotServerOptions.RecentToolCallRecorder == nil {
		t.Fatal("server recent-tool-call recorder = nil")
	}
	if gotServerOptions.SelectionStore == nil || gotServerOptions.SelectionStore.Selected("") != "i42" {
		t.Fatalf("server selection store default = %q", gotServerOptions.SelectionStore.Selected(""))
	}
	if gotServerOptions.Capability == nil || gotServerOptions.Capability.Mode() != string(safety.ModeFull) || gotToolOptions.Capability == nil {
		t.Fatalf("capability not propagated")
	}
}
