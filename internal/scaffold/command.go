package scaffold

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Runner interface {
	Run(ctx context.Context, directory, command string, args ...string) error
}

type commandRunner struct {
	stdout io.Writer
	stderr io.Writer
}

func (r commandRunner) Run(ctx context.Context, directory, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = directory
	cmd.Stdin = os.Stdin
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", command, args, err)
	}
	return nil
}
