package app

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/credstore"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeSetupStore struct {
	secret           string
	getErr           error
	setErr           error
	getErrAfterSet   error
	mismatchAfterSet bool
	sets             []string
}

func (s *fakeSetupStore) Get(ctx context.Context, account string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if account != credstore.IntervalsAPIKeyAccount {
		return "", errors.New("unexpected account")
	}
	if s.getErr != nil {
		return "", s.getErr
	}
	return s.secret, nil
}

func (s *fakeSetupStore) Set(ctx context.Context, account, secret string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if account != credstore.IntervalsAPIKeyAccount {
		return errors.New("unexpected account")
	}
	if s.setErr != nil {
		return s.setErr
	}
	s.sets = append(s.sets, secret)
	if s.mismatchAfterSet {
		s.secret = strings.Repeat("x", len(secret)+1)
	} else {
		s.secret = secret
	}
	s.getErr = s.getErrAfterSet
	return nil
}

func (s *fakeSetupStore) Delete(ctx context.Context, account string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if account != credstore.IntervalsAPIKeyAccount {
		return errors.New("unexpected account")
	}
	return nil
}

type fakeSetupPrompter struct {
	confirms       []bool
	confirmPrompts []string
	lines          []string
	linePrompts    []string
	secrets        []string
	secretPrompts  []string
}

func (p *fakeSetupPrompter) Confirm(ctx context.Context, prompt string, _ bool) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	p.confirmPrompts = append(p.confirmPrompts, prompt)
	if len(p.confirms) == 0 {
		return false, errors.New("unexpected confirm prompt")
	}
	answer := p.confirms[0]
	p.confirms = p.confirms[1:]
	return answer, nil
}

func (p *fakeSetupPrompter) ReadLine(ctx context.Context, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	p.linePrompts = append(p.linePrompts, prompt)
	if len(p.lines) == 0 {
		return "", errors.New("unexpected line prompt")
	}
	line := p.lines[0]
	p.lines = p.lines[1:]
	return line, nil
}

func noOpSetupConfigWriter(context.Context, string, config.Config, config.WriteOptions) error {
	return nil
}

func (p *fakeSetupPrompter) ReadSecret(ctx context.Context, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	p.secretPrompts = append(p.secretPrompts, prompt)
	if len(p.secrets) == 0 {
		return "", errors.New("unexpected secret prompt")
	}
	secret := p.secrets[0]
	p.secrets = p.secrets[1:]
	return secret, nil
}

func TestRunSetupExistingKeyDefaultNoCancelsBeforeReadingSecret(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{false}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{secret: "stored"},
		Prompter:        prompter,
		ConfigExists: func(string) (bool, error) {
			t.Fatal("config existence must not be checked after key overwrite denial")
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Setup canceled; nothing changed.") {
		t.Fatalf("stdout %q missing cancellation", stdout.String())
	}
	if len(prompter.secretPrompts) != 0 {
		t.Fatalf("ReadSecret prompts = %v, want none", prompter.secretPrompts)
	}
	if got := prompter.confirmPrompts; len(got) != 1 || !strings.Contains(got[0], "API key is already stored") {
		t.Fatalf("confirm prompts = %v, want existing-key prompt", got)
	}
}

func TestRunSetupExistingConfigDefaultNoCancelsBeforeReadingSecret(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{false}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(path string) (bool, error) { return path == "/tmp/icuvisor.json", nil },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Setup canceled; nothing changed.") {
		t.Fatalf("stdout %q missing cancellation", stdout.String())
	}
	if len(prompter.secretPrompts) != 0 {
		t.Fatalf("ReadSecret prompts = %v, want none", prompter.secretPrompts)
	}
	if got := prompter.confirmPrompts; len(got) != 1 || !strings.Contains(got[0], "config file already exists") {
		t.Fatalf("confirm prompts = %v, want existing-config prompt", got)
	}
}

func TestRunSetupForceSkipsOnlyConfigPrompt(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{true}, secrets: []string{"api-key"}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Force:           true,
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(path string) (bool, error) { return path == "/tmp/icuvisor.json", nil },
		ConfigWriter: func(_ context.Context, _ string, _ config.Config, opts config.WriteOptions) error {
			if !opts.AllowOverwrite {
				t.Fatal("--force must allow config overwrite")
			}
			return nil
		},
		ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
			return SetupProfile{AthleteID: "12345", DisplayName: "Jane Doe", FTP: 245}, nil
		},
		TimezoneDetector: func() string { return "UTC" },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if len(prompter.confirmPrompts) != 1 || !strings.Contains(prompter.confirmPrompts[0], "Detected timezone") {
		t.Fatalf("confirm prompts = %v, want only timezone prompt", prompter.confirmPrompts)
	}
	if got := prompter.secretPrompts; len(got) != 1 || !strings.Contains(got[0], "intervals.icu API key") {
		t.Fatalf("secret prompts = %v, want API-key prompt", got)
	}
}

func TestRunSetupFetchesProfileNormalizesIDAndConfirmsTimezone(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{true}, secrets: []string{"api-key"}}
	var gotKey string
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ConfigWriter:    noOpSetupConfigWriter,
		ProfileFetcher: func(_ context.Context, apiKey string) (SetupProfile, error) {
			gotKey = apiKey
			return SetupProfile{AthleteID: "12345", DisplayName: "Jane Doe", FTP: 245}, nil
		},
		TimezoneDetector: func() string { return "Europe/Madrid" },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if gotKey != "api-key" {
		t.Fatalf("profile fetcher key = %q, want api-key", gotKey)
	}
	for _, want := range []string{"connected as \"Jane Doe\"", "athlete i12345", "FTP 245 W", "timezone Europe/Madrid"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout %q missing %q", stdout.String(), want)
		}
	}
}

func TestRunSetupAllowsTimezoneOverride(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	prompter := &fakeSetupPrompter{confirms: []bool{false}, lines: []string{"America/Sao_Paulo"}, secrets: []string{"api-key"}}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		Stdout:          &stdout,
		CredentialStore: &fakeSetupStore{getErr: credstore.ErrNotFound},
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ConfigWriter:    noOpSetupConfigWriter,
		ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
			return SetupProfile{AthleteID: "i12345", DisplayName: "Jane Doe"}, nil
		},
		TimezoneDetector: func() string { return "Europe/Madrid" },
	})
	if err != nil {
		t.Fatalf("RunSetup() error = %v", err)
	}
	if got := prompter.linePrompts; len(got) != 1 || !strings.Contains(got[0], "Timezone") {
		t.Fatalf("line prompts = %v, want timezone override prompt", got)
	}
	if !strings.Contains(stdout.String(), "timezone America/Sao_Paulo") {
		t.Fatalf("stdout %q missing timezone override", stdout.String())
	}
}

func TestRunSetupUnauthorizedKeyReturnsFixURL(t *testing.T) {
	t.Parallel()

	prompter := &fakeSetupPrompter{secrets: []string{"bad-key"}}
	store := &fakeSetupStore{getErr: credstore.ErrNotFound}
	err := RunSetup(context.Background(), SetupOptions{
		ConfigPath:      "/tmp/icuvisor.json",
		CredentialStore: store,
		Prompter:        prompter,
		ConfigExists:    func(string) (bool, error) { return false, nil },
		ProfileFetcher: func(context.Context, string) (SetupProfile, error) {
			return SetupProfile{}, errors.Join(errors.New("wrapped"), intervals.ErrUnauthorized)
		},
	})
	if err == nil {
		t.Fatal("RunSetup() error = nil, want unauthorized error")
	}
	if !strings.Contains(err.Error(), "API key not accepted") || !strings.Contains(err.Error(), "https://intervals.icu/settings") {
		t.Fatalf("error = %q, want fix URL", err.Error())
	}
	if len(store.sets) != 0 {
		t.Fatalf("Set calls = %v, want none", store.sets)
	}
	if len(prompter.linePrompts) != 0 || len(prompter.confirmPrompts) != 0 {
		t.Fatalf("prompts after unauthorized = confirms %v lines %v, want none", prompter.confirmPrompts, prompter.linePrompts)
	}
}
