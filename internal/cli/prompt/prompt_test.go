package prompt

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestTerminalConfirm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		prompt     string
		defaultYes bool
		want       bool
		wantOut    string
	}{
		{name: "default yes", input: "\n", prompt: "Use detected?", defaultYes: true, want: true, wantOut: "Use detected? "},
		{name: "default no", input: "\n", prompt: "Overwrite?", defaultYes: false, want: false, wantOut: "Overwrite? "},
		{name: "yes", input: " yes \n", prompt: "Overwrite?", want: true, wantOut: "Overwrite? "},
		{name: "no", input: "N\n", prompt: "Overwrite?", want: false, wantOut: "Overwrite? "},
		{name: "retry invalid", input: "maybe\ny\n", prompt: "Overwrite?", want: true, wantOut: "Overwrite? Please answer y or n.\nOverwrite? "},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			got, err := NewTerminal(strings.NewReader(tc.input), &out).Confirm(context.Background(), tc.prompt, tc.defaultYes)
			if err != nil {
				t.Fatalf("Confirm() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("Confirm() = %v, want %v", got, tc.want)
			}
			if out.String() != tc.wantOut {
				t.Fatalf("output = %q, want %q", out.String(), tc.wantOut)
			}
		})
	}
}

func TestTerminalReadLineTrimsAnswer(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	got, err := NewTerminal(strings.NewReader("  Europe/Madrid  \n"), &out).ReadLine(context.Background(), "Timezone:")
	if err != nil {
		t.Fatalf("ReadLine() error = %v", err)
	}
	if got != "Europe/Madrid" {
		t.Fatalf("ReadLine() = %q, want Europe/Madrid", got)
	}
	if out.String() != "Timezone:\n> " {
		t.Fatalf("output = %q, want prompt", out.String())
	}
}

func TestTerminalReadSecretRequiresInteractiveTerminal(t *testing.T) {
	t.Parallel()

	_, err := NewTerminal(strings.NewReader("secret\n"), &bytes.Buffer{}).ReadSecret(context.Background(), "Secret:")
	if err == nil || !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("ReadSecret() error = %v, want interactive terminal error", err)
	}
}

func TestTerminalHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewTerminal(strings.NewReader("y\n"), &bytes.Buffer{}).Confirm(ctx, "Continue?", true)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Confirm() error = %v, want context.Canceled", err)
	}
}
