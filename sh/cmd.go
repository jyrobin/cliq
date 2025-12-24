// Package sh provides shell-like utilities for running commands,
// piping data, and working with HTTP/WebSocket in Go code.
//
// Designed for use in tests and scripts where you'd otherwise
// reach for shell scripts or subprocess calls.
package sh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Result holds the output of a command execution.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// OK returns true if the command succeeded (exit code 0, no error).
func (r *Result) OK() bool {
	return r.ExitCode == 0 && r.Err == nil
}

// Failed returns true if the command failed.
func (r *Result) Failed() bool {
	return !r.OK()
}

// String returns stdout, trimmed of trailing whitespace.
func (r *Result) String() string {
	return strings.TrimRight(r.Stdout, "\n\r\t ")
}

// Lines returns stdout split into lines.
func (r *Result) Lines() []string {
	s := strings.TrimRight(r.Stdout, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// Combined returns stdout and stderr combined.
func (r *Result) Combined() string {
	return r.Stdout + r.Stderr
}

// Run executes a command and returns the result.
func Run(name string, args ...string) *Result {
	return RunContext(context.Background(), name, args...)
}

// RunContext executes a command with context.
func RunContext(ctx context.Context, name string, args ...string) *Result {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return &Result{Err: err, ExitCode: -1}
		}
	}

	return &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

// RunWithInput executes a command with stdin input.
func RunWithInput(input string, name string, args ...string) *Result {
	return RunWithInputContext(context.Background(), input, name, args...)
}

// RunWithInputContext executes a command with stdin input and context.
func RunWithInputContext(ctx context.Context, input string, name string, args ...string) *Result {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return &Result{Err: err, ExitCode: -1}
		}
	}

	return &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

// RunTimeout executes a command with a timeout.
func RunTimeout(timeout time.Duration, name string, args ...string) *Result {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunContext(ctx, name, args...)
}

// Pipe runs a command with input from a previous result's stdout.
func Pipe(input *Result, name string, args ...string) *Result {
	if input.Err != nil {
		return input // propagate error
	}
	return RunWithInput(input.Stdout, name, args...)
}

// PipeString runs a command with string input.
func PipeString(input string, name string, args ...string) *Result {
	return RunWithInput(input, name, args...)
}

// Chain runs multiple commands in sequence, piping stdout to stdin.
// Stops on first failure.
func Chain(cmds ...[]string) *Result {
	if len(cmds) == 0 {
		return &Result{}
	}

	result := Run(cmds[0][0], cmds[0][1:]...)
	for i := 1; i < len(cmds) && result.OK(); i++ {
		result = Pipe(result, cmds[i][0], cmds[i][1:]...)
	}
	return result
}

// Cmd is a builder for constructing commands with options.
type Cmd struct {
	name    string
	args    []string
	dir     string
	env     []string
	stdin   io.Reader
	timeout time.Duration
}

// Command creates a new command builder.
func Command(name string, args ...string) *Cmd {
	return &Cmd{name: name, args: args}
}

// Dir sets the working directory.
func (c *Cmd) Dir(dir string) *Cmd {
	c.dir = dir
	return c
}

// Env adds environment variables (KEY=VALUE format).
func (c *Cmd) Env(env ...string) *Cmd {
	c.env = append(c.env, env...)
	return c
}

// Stdin sets stdin from a string.
func (c *Cmd) Stdin(input string) *Cmd {
	c.stdin = strings.NewReader(input)
	return c
}

// StdinReader sets stdin from a reader.
func (c *Cmd) StdinReader(r io.Reader) *Cmd {
	c.stdin = r
	return c
}

// Timeout sets execution timeout.
func (c *Cmd) Timeout(d time.Duration) *Cmd {
	c.timeout = d
	return c
}

// Run executes the command.
func (c *Cmd) Run() *Result {
	ctx := context.Background()
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, c.name, c.args...)
	if c.dir != "" {
		cmd.Dir = c.dir
	}
	if len(c.env) > 0 {
		cmd.Env = append(os.Environ(), c.env...)
	}
	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return &Result{Err: err, ExitCode: -1}
		}
	}

	return &Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

// Output runs and returns stdout (convenience for simple cases).
func (c *Cmd) Output() (string, error) {
	r := c.Run()
	if r.Err != nil {
		return "", r.Err
	}
	if r.ExitCode != 0 {
		return "", fmt.Errorf("exit code %d: %s", r.ExitCode, r.Stderr)
	}
	return r.String(), nil
}
