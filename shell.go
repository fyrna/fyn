package fytask

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

func ShellEnv(
	ctx context.Context,
	env []string,
	cmd string,
	args ...string,
) error {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(), env...)

	return c.Run()
}

func Silent(ctx context.Context, command string) error {
	parts, err := splitCommand(command)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return nil
	}

	c := exec.CommandContext(ctx, parts[0], parts[1:]...)
	c.Stdout = nil
	c.Stderr = nil
	return c.Run()
}

// "abstraction" XD
func Sh(ctx context.Context, command string) error {
	parts, err := splitCommand(command)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return nil
	}
	return Shell(ctx, parts[0], parts[1:]...)
}

func ShOut(ctx context.Context, command string) (string, error) {
	parts, err := splitCommand(command)
	if err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return "", nil
	}

	out, err := ShellOut(ctx, parts[0], parts[1:]...)
	if err != nil {
		return "", err
	}
	return out, nil
}

func splitCommand(input string) ([]string, error) {
	var args []string
	var current []rune
	inSingle := false
	inDouble := false
	escape := false

	for _, c := range input {
		switch {
		case escape:
			current = append(current, c)
			escape = false
		case c == '\\':
			escape = true
		case c == '"' && !inSingle:
			inDouble = !inDouble
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == ' ' && !inSingle && !inDouble:
			if len(current) > 0 {
				args = append(args, string(current))
				current = nil
			}
		default:
			current = append(current, c)
		}
	}

	if len(current) > 0 {
		args = append(args, string(current))
	}

	if inSingle || inDouble {
		return nil, fmt.Errorf("unmatched quotes")
	}
	if escape {
		return nil, fmt.Errorf("unfinished escape sequence")
	}

	return args, nil
}
