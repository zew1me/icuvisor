package prompt

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Prompter reads terminal confirmations, free-form answers, and masked secrets.
type Prompter interface {
	Confirm(ctx context.Context, prompt string, defaultYes bool) (bool, error)
	ReadLine(ctx context.Context, prompt string) (string, error)
	ReadSecret(ctx context.Context, prompt string) (string, error)
}

// Terminal implements Prompter using standard terminal input and output streams.
type Terminal struct {
	in     io.Reader
	out    io.Writer
	reader *bufio.Reader
}

// NewTerminal returns a terminal prompter backed by in and out.
func NewTerminal(in io.Reader, out io.Writer) *Terminal {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = io.Discard
	}
	return &Terminal{in: in, out: out, reader: bufio.NewReader(in)}
}

// Confirm asks a yes/no question until it receives a valid answer.
func (p *Terminal) Confirm(ctx context.Context, prompt string, defaultYes bool) (bool, error) {
	for {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		_, _ = fmt.Fprint(p.out, prompt+" ")
		answer, err := p.reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return false, err
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "" {
			return defaultYes, nil
		}
		switch answer {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			_, _ = fmt.Fprintln(p.out, "Please answer y or n.")
		}
	}
}

// ReadLine asks for a free-form answer and trims surrounding whitespace.
func (p *Terminal) ReadLine(ctx context.Context, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	_, _ = fmt.Fprintln(p.out, prompt)
	_, _ = fmt.Fprint(p.out, "> ")
	answer, err := p.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(answer), nil
}

// ReadSecret asks for a masked answer from an interactive terminal.
func (p *Terminal) ReadSecret(ctx context.Context, prompt string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	file, ok := p.in.(*os.File)
	if !ok {
		return "", errors.New("masked API-key input requires an interactive terminal")
	}
	_, _ = fmt.Fprintln(p.out, prompt)
	_, _ = fmt.Fprint(p.out, "> ")
	secret, err := term.ReadPassword(int(file.Fd()))
	_, _ = fmt.Fprintln(p.out)
	if err != nil {
		return "", err
	}
	return string(secret), nil
}
