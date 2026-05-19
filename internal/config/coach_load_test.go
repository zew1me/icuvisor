package config

import (
	"context"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/coach"
)

func TestLoadCoachModeFromEnvAndDotEnv(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dotEnvPath := dir + "/.env"
	writeFile(t, dotEnvPath, strings.Join([]string{
		"INTERVALS_ICU_API_KEY=dotenv-key",
		"INTERVALS_ICU_ATHLETE_ID=i444",
		"ICUVISOR_COACH_MODE=auto",
	}, "\n"))

	cfg, err := Load(context.Background(), Options{DotEnvPath: dotEnvPath, Env: map[string]string{}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CoachMode != coach.ModeAuto || cfg.CoachModeEnabled() {
		t.Fatalf("CoachMode = %q enabled=%t, want auto disabled with empty roster", cfg.CoachMode, cfg.CoachModeEnabled())
	}

	cfg, err = Load(context.Background(), Options{DotEnvPath: dotEnvPath, Env: map[string]string{
		EnvAPIKey:     "env-key",
		EnvAthleteID:  "i12345",
		EnvCoachMode:  " ON ",
		EnvConfigPath: "",
	}})
	if err == nil {
		t.Fatal("Load() with coach mode on and empty roster error = nil, want error")
	}
	if !strings.Contains(err.Error(), "coach mode is on") {
		t.Fatalf("error = %q, want coach mode roster error", err)
	}

	_, err = Load(context.Background(), Options{Env: map[string]string{
		EnvAPIKey:     "env-key",
		EnvAthleteID:  "i12345",
		EnvCoachMode:  "maybe",
		EnvConfigPath: "",
	}})
	if err == nil || !strings.Contains(err.Error(), "invalid coach mode") {
		t.Fatalf("Load() invalid coach mode error = %v, want invalid coach mode", err)
	}
}

func TestLoadCoachConfigSchemaAndValidation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := dir + "/config.json"
	writeFile(t, configPath, `{
		"api_key": "json-key",
		"athlete_id": "i111",
		"coach": {
			"athletes": [
				{"id": "i222", "label": " Jane ", "allowed_tools": ["get_*", "get_*"], "denied_tools": ["delete_event"]},
				{"id": "i333", "allowed_tools": ["*"], "denied_tools": []}
			],
			"default_athlete_id": "i333"
		}
	}`)

	cfg, err := Load(context.Background(), Options{Path: configPath, Env: map[string]string{EnvCoachMode: "auto"}})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.CoachMode != coach.ModeAuto || !cfg.CoachModeEnabled() {
		t.Fatalf("CoachMode = %q enabled=%t, want auto enabled", cfg.CoachMode, cfg.CoachModeEnabled())
	}
	if cfg.Coach.DefaultAthleteID != "i333" {
		t.Fatalf("DefaultAthleteID = %q, want i333", cfg.Coach.DefaultAthleteID)
	}
	if cfg.AthleteID != "i333" {
		t.Fatalf("AthleteID = %q, want coach default i333 to override legacy top-level athlete_id", cfg.AthleteID)
	}
	if len(cfg.Coach.Athletes) != 2 || cfg.Coach.Athletes[0].ID != "i222" || cfg.Coach.Athletes[0].Label != "Jane" {
		t.Fatalf("Coach athletes = %#v, want normalized roster", cfg.Coach.Athletes)
	}
	if got := cfg.Coach.Athletes[0].AllowedTools; len(got) != 1 || got[0] != "get_*" {
		t.Fatalf("AllowedTools = %#v, want deduped get_*", got)
	}
}

func TestLoadCoachModeAllowsCoachOnlyConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode string
	}{
		{name: "on", mode: "on"},
		{name: "auto", mode: "auto"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := dir + "/config.json"
			writeFile(t, path, `{
				"api_key": "json-key",
				"coach": {
					"athletes": [{"id": "i222", "allowed_tools": ["*"]}],
					"default_athlete_id": "i222"
				}
			}`)
			cfg, err := Load(context.Background(), Options{Path: path, DotEnvPath: dir + "/missing.env", Env: map[string]string{EnvCoachMode: tc.mode}})
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if !cfg.CoachModeEnabled() || cfg.AthleteID != "i222" {
				t.Fatalf("Coach enabled=%t AthleteID=%q, want enabled with default i222", cfg.CoachModeEnabled(), cfg.AthleteID)
			}
		})
	}
}

func TestLoadCoachConfigValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		env     map[string]string
		wantErr string
	}{
		{name: "unknown json field", json: `{"api_key":"k","athlete_id":"i1","coach":{"athletes":[],"typo":true}}`, wantErr: "unknown field"},
		{name: "duplicate normalized athlete", json: `{"api_key":"k","athlete_id":"i1","coach":{"athletes":[{"id":"i2"},{"id":"I2"}]}}`, wantErr: "duplicate coach athlete id"},
		{name: "default outside roster", json: `{"api_key":"k","athlete_id":"i1","coach":{"athletes":[{"id":"i2"}],"default_athlete_id":"i3"}}`, wantErr: "coach.default_athlete_id"},
		{name: "unknown exact tool", json: `{"api_key":"k","athlete_id":"i1","coach":{"athletes":[{"id":"i2","allowed_tools":["get_athlete_profiel"]}]}}`, wantErr: "unknown athlete-scoped tool"},
		{name: "unknown wildcard", json: `{"api_key":"k","athlete_id":"i1","coach":{"athletes":[{"id":"i2","allowed_tools":["bogus_*"]}]}}`, wantErr: "matches no athlete-scoped tools"},
		{name: "off still validates stanza", json: `{"api_key":"k","athlete_id":"i1","coach":{"athletes":[{"id":"i2","denied_tools":["select_athlete"]}]}}`, env: map[string]string{EnvCoachMode: "off"}, wantErr: "unknown athlete-scoped tool"},
		{name: "on multiple athletes needs default", json: `{"api_key":"k","athlete_id":"i1","coach":{"athletes":[{"id":"i2"},{"id":"i3"}]}}`, env: map[string]string{EnvCoachMode: "on"}, wantErr: "default_athlete_id is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := dir + "/config.json"
			writeFile(t, path, tc.json)
			env := tc.env
			if env == nil {
				env = map[string]string{}
			}
			_, err := Load(context.Background(), Options{Path: path, DotEnvPath: dir + "/missing.env", Env: env})
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %q, want %q", err, tc.wantErr)
			}
		})
	}
}
