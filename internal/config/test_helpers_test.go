package config

import (
	"context"
	"os"
	"testing"
)

type fakeCredentialStore struct {
	value string
	err   error
	calls int
}

func (f *fakeCredentialStore) Get(ctx context.Context, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.value, nil
}

func (f *fakeCredentialStore) Set(context.Context, string, string) error {
	return nil
}

func (f *fakeCredentialStore) Delete(context.Context, string) error {
	return nil
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}
