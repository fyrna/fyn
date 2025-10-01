package fytask

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// ShellEnv executes a command with environment variables, streaming stdout and stderr.
func ShellEnv(
	ctx context.Context,
	env []string,
	cmd string,
	args ...string,
) error {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if env != nil {
		c.Env = append(os.Environ(), env...)
	}

	if err := c.Run(); err != nil {
		return fmt.Errorf("shell: %w", err)
	}
	return nil
}

// ShellOutEnv executes a command with environment variables and returns its output.
func ShellOutEnv(
	ctx context.Context,
	env []string,
	cmd string,
	args ...string,
) (string, error) {
	c := exec.CommandContext(ctx, cmd, args...)

	if env != nil {
		c.Env = append(os.Environ(), env...)
	}

	out, err := c.Output()
	if err != nil {
		return "", fmt.Errorf("shellout: %w", err)
	}
	return string(out), nil
}

// ShellCombinedOutEnv executes a command with environment variables and returns both stdout and stderr combined.
func ShellCombinedOutEnv(
	ctx context.Context,
	env []string,
	cmd string,
	args ...string,
) (string, error) {
	c := exec.CommandContext(ctx, cmd, args...)

	if env != nil {
		c.Env = append(os.Environ(), env...)
	}

	out, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("shellcombined: %w", err)
	}
	return string(out), nil
}

// SilentEnv executes a command with environment variables without capturing any output.
func SilentEnv(
	ctx context.Context,
	env []string,
	cmd string,
	args ...string,
) error {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = nil
	c.Stderr = nil

	if env != nil {
		c.Env = append(os.Environ(), env...)
	}

	if err := c.Run(); err != nil {
		return fmt.Errorf("silent: %w", err)
	}
	return nil
}

// Shell executes a command with stdout and stderr streaming.
func Shell(ctx context.Context, cmd string, args ...string) error {
	return ShellEnv(ctx, nil, cmd, args...)
}

// ShellOut executes a command and returns its output.
func ShellOut(ctx context.Context, cmd string, args ...string) (string, error) {
	return ShellOutEnv(ctx, nil, cmd, args...)
}

// ShellCombinedOut executes a command and returns both stdout and stderr combined.
func ShellCombinedOut(ctx context.Context, cmd string, args ...string) (string, error) {
	return ShellCombinedOutEnv(ctx, nil, cmd, args...)
}

// Silent executes a command without capturing any output.
func Silent(ctx context.Context, command string) error {
	parts, err := splitCommand(command)
	if err != nil {
		return fmt.Errorf("silent: %w", err)
	}
	if len(parts) == 0 {
		return nil
	}

	return SilentEnv(ctx, nil, parts[0], parts[1:]...)
}

// Sh is a convenience wrapper for executing shell commands from a string.
func Sh(ctx context.Context, command string) error {
	parts, err := splitCommand(command)
	if err != nil {
		return fmt.Errorf("sh: %w", err)
	}
	if len(parts) == 0 {
		return nil
	}
	return ShellEnv(ctx, nil, parts[0], parts[1:]...)
}

// ShEnv is a convenience wrapper for executing shell commands from a string with environment variables.
func ShEnv(
	ctx context.Context,
	env []string,
	command string,
) error {
	parts, err := splitCommand(command)
	if err != nil {
		return fmt.Errorf("shenv: %w", err)
	}
	if len(parts) == 0 {
		return nil
	}
	return ShellEnv(ctx, env, parts[0], parts[1:]...)
}

// ShOut executes a shell command from a string and returns its output.
func ShOut(ctx context.Context, command string) (string, error) {
	parts, err := splitCommand(command)
	if err != nil {
		return "", fmt.Errorf("shout: %w", err)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return ShellOutEnv(ctx, nil, parts[0], parts[1:]...)
}

// ShOutEnv executes a shell command from a string with environment variables and returns its output.
func ShOutEnv(
	ctx context.Context,
	env []string,
	command string,
) (string, error) {
	parts, err := splitCommand(command)
	if err != nil {
		return "", fmt.Errorf("shoutenv: %w", err)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return ShellOutEnv(ctx, env, parts[0], parts[1:]...)
}

// splitCommand splits a command string into arguments, respecting quotes and escape sequences.
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
		return nil, fmt.Errorf("unmatched quotes in command: %s", input)
	}
	if escape {
		return nil, fmt.Errorf("unfinished escape sequence in command: %s", input)
	}

	return args, nil
}
