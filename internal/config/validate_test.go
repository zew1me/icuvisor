package config

import (
	"context"
	"strings"
	"testing"
)

func TestParseDebugMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "true", in: "true", want: true},
		{name: "mixed case", in: " TRUE ", want: true},
		{name: "false", in: "false", want: false},
		{name: "invalid", in: "yes", want: false},
		{name: "empty", in: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ParseDebugMetadata(tt.in); got != tt.want {
				t.Fatalf("ParseDebugMetadata(%q) = %t, want %t", tt.in, got, tt.want)
			}
		})
	}
}

func TestLoadConfigFileErrorsAreActionableAndRedacted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	testCredential := strings.Repeat("x", 12)

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		_, err := Load(context.Background(), Options{
			Path:       dir + "/missing.json",
			DotEnvPath: dir + "/missing.env",
			Env: map[string]string{
				EnvAPIKey:    testCredential,
				EnvAthleteID: "i123",
			},
		})
		if err == nil {
			t.Fatal("Load() error = nil, want error")
		}
		msg := err.Error()
		for _, want := range []string{"config file", "not found", "--config", EnvConfigPath} {
			if !strings.Contains(msg, want) {
				t.Fatalf("error %q does not contain %q", msg, want)
			}
		}
		if strings.Contains(msg, testCredential) {
			t.Fatalf("error leaked API key: %q", msg)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		path := dir + "/invalid.json"
		writeFile(t, path, `{"api_key":"`+testCredential+`","athlete_id":"i123","extra":true}`)

		_, err := Load(context.Background(), Options{Path: path, DotEnvPath: dir + "/missing.env", Env: map[string]string{}})
		if err == nil {
			t.Fatal("Load() error = nil, want error")
		}
		msg := err.Error()
		for _, want := range []string{"invalid config JSON", "expected fields", "api_key", "athlete_id"} {
			if !strings.Contains(msg, want) {
				t.Fatalf("error %q does not contain %q", msg, want)
			}
		}
		if strings.Contains(msg, testCredential) {
			t.Fatalf("error leaked API key: %q", msg)
		}
	})
}

func TestLoadValidationErrorsAreActionableAndRedacted(t *testing.T) {
	t.Parallel()

	testCredential := strings.Repeat("x", 12)
	withCredential := func(values map[string]string) map[string]string {
		values[EnvAPIKey] = testCredential
		return values
	}

	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{name: "missing API key", env: map[string]string{EnvAthleteID: "i123"}, wantErr: "missing intervals.icu API key"},
		{name: "missing athlete ID", env: withCredential(map[string]string{}), wantErr: "missing athlete ID"},
		{name: "invalid athlete ID", env: withCredential(map[string]string{EnvAthleteID: "abc"}), wantErr: "invalid athlete ID"},
		{name: "invalid timezone", env: withCredential(map[string]string{EnvAthleteID: "i123", EnvTimezone: "Mars/Base"}), wantErr: "invalid timezone"},
		{name: "invalid timeout", env: withCredential(map[string]string{EnvAthleteID: "i123", EnvHTTPTimeout: "0s"}), wantErr: "invalid HTTP timeout"},
		{name: "invalid base URL", env: withCredential(map[string]string{EnvAthleteID: "i123", EnvAPIBaseURL: "ftp://example.test"}), wantErr: "invalid API base URL"},
		{name: "invalid transport", env: withCredential(map[string]string{EnvAthleteID: "i123", EnvTransport: "websocket"}), wantErr: "invalid MCP transport"},
		{name: "invalid bind", env: withCredential(map[string]string{EnvAthleteID: "i123", EnvHTTPBind: ":8765"}), wantErr: "invalid HTTP bind address"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Load(context.Background(), Options{DotEnvPath: t.TempDir() + "/missing.env", Env: tc.env})
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.wantErr) {
				t.Fatalf("error = %q, want to contain %q", msg, tc.wantErr)
			}
			if strings.Contains(msg, testCredential) {
				t.Fatalf("error leaked API key: %q", msg)
			}
		})
	}
}
