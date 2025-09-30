package fytask

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

func Shell(ctx context.Context, cmd string, args ...string) error {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func ShellOut(ctx context.Context, cmd string, args ...string) (string, error) {
	c := exec.CommandContext(ctx, cmd, args...)
	out, err := c.Output()
	return string(out), err
}

func Silent(ctx context.Context, command string) error {
	cmd, args := parseCommand(command)

	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = nil
	c.Stderr = nil
	return c.Run()
}

// "abstraction" XD
func Sh(ctx context.Context, command string) error {
	cmd, args := parseCommand(command)
	return Shell(ctx, cmd, args...)
}

func ShOut(ctx context.Context, command string) (string, error) {
	cmd, args := parseCommand(command)
	out, err := ShellOut(ctx, cmd, args...)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func parseCommand(cmd string) (string, []string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}
