package sh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Stream represents a data stream that can be piped through commands.
type Stream struct {
	input    io.Reader
	err      error
	timeout  time.Duration
	dir      string
	env      []string
	maxBytes int64 // 0 = unlimited
}

// Exec starts a pipeline with a command.
func Exec(name string, args ...string) *Stream {
	s := &Stream{}
	return s.run(name, args...)
}

// From starts a pipeline with string input.
func From(input string) *Stream {
	return &Stream{input: strings.NewReader(input)}
}

// FromBytes starts a pipeline with byte input.
func FromBytes(input []byte) *Stream {
	return &Stream{input: bytes.NewReader(input)}
}

// FromReader starts a pipeline with a reader.
func FromReader(r io.Reader) *Stream {
	return &Stream{input: r}
}

// FromFile starts a pipeline with file contents.
func FromFile(path string) *Stream {
	f, err := os.Open(path)
	if err != nil {
		return &Stream{err: err}
	}
	return &Stream{input: f}
}

// Timeout sets timeout for subsequent commands in the pipeline.
func (s *Stream) Timeout(d time.Duration) *Stream {
	s.timeout = d
	return s
}

// Dir sets working directory for subsequent commands.
func (s *Stream) Dir(dir string) *Stream {
	s.dir = dir
	return s
}

// Env adds environment variables for subsequent commands.
func (s *Stream) Env(env ...string) *Stream {
	s.env = append(s.env, env...)
	return s
}

// Limit sets maximum bytes to buffer (0 = unlimited).
// If exceeded, output is truncated and Err() returns ErrTruncated.
func (s *Stream) Limit(maxBytes int64) *Stream {
	s.maxBytes = maxBytes
	return s
}

// ErrTruncated indicates output was truncated due to Limit().
var ErrTruncated = fmt.Errorf("output truncated: exceeded byte limit")

// Pipe pipes the stream through a command (buffered).
func (s *Stream) Pipe(name string, args ...string) *Stream {
	if s.err != nil {
		return s
	}
	return s.run(name, args...)
}

// run executes a command with current stream as stdin (buffered).
func (s *Stream) run(name string, args ...string) *Stream {
	ctx := context.Background()
	if s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	if s.input != nil {
		cmd.Stdin = s.input
	}
	if s.dir != "" {
		cmd.Dir = s.dir
	}
	if len(s.env) > 0 {
		cmd.Env = append(os.Environ(), s.env...)
	}

	var stdout bytes.Buffer
	var writer io.Writer = &stdout
	var truncated bool

	// Apply limit if set
	if s.maxBytes > 0 {
		writer = &limitedWriter{w: &stdout, max: s.maxBytes, truncated: &truncated}
	}
	cmd.Stdout = writer

	err := cmd.Run()
	if truncated {
		err = ErrTruncated
	}

	if err != nil {
		return &Stream{
			err:      err,
			input:    &stdout,
			maxBytes: s.maxBytes,
		}
	}

	return &Stream{
		input:    &stdout,
		timeout:  s.timeout,
		dir:      s.dir,
		env:      s.env,
		maxBytes: s.maxBytes,
	}
}

// limitedWriter stops writing after max bytes.
type limitedWriter struct {
	w         io.Writer
	max       int64
	written   int64
	truncated *bool
}

func (lw *limitedWriter) Write(p []byte) (int, error) {
	remaining := lw.max - lw.written
	if remaining <= 0 {
		*lw.truncated = true
		return len(p), nil // pretend we wrote it
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
		*lw.truncated = true
	}
	n, err := lw.w.Write(p)
	lw.written += int64(n)
	return n, err
}

// --- Streaming Pipe (no buffering) ---

// StreamPipe pipes through a command using OS pipes (true streaming).
// Data flows directly between processes without buffering.
// Note: errors from the command are not available until Drain() is called.
func (s *Stream) StreamPipe(name string, args ...string) *Stream {
	if s.err != nil {
		return s
	}

	ctx := context.Background()
	if s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		_ = cancel // will be called when command completes
	}

	cmd := exec.CommandContext(ctx, name, args...)
	if s.input != nil {
		cmd.Stdin = s.input
	}
	if s.dir != "" {
		cmd.Dir = s.dir
	}
	if len(s.env) > 0 {
		cmd.Env = append(os.Environ(), s.env...)
	}

	pr, pw := io.Pipe()
	cmd.Stdout = pw

	var cmdErr error
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer pw.Close()
		cmdErr = cmd.Run()
	}()

	// Return a stream that reads from the pipe
	// The error will be captured when reading completes
	return &Stream{
		input: &streamReader{
			r:      pr,
			wg:     &wg,
			cmdErr: &cmdErr,
		},
		timeout:  s.timeout,
		dir:      s.dir,
		env:      s.env,
		maxBytes: s.maxBytes,
	}
}

// streamReader wraps pipe reader and captures command error.
type streamReader struct {
	r      *io.PipeReader
	wg     *sync.WaitGroup
	cmdErr *error
	done   bool
}

func (sr *streamReader) Read(p []byte) (int, error) {
	n, err := sr.r.Read(p)
	if err == io.EOF && !sr.done {
		sr.done = true
		sr.wg.Wait() // ensure command finished
		if *sr.cmdErr != nil {
			return n, *sr.cmdErr
		}
	}
	return n, err
}

// --- Transform operations ---

// Transform applies a Go function to transform the stream.
func (s *Stream) Transform(fn func(string) string) *Stream {
	if s.err != nil {
		return s
	}
	data, err := s.bytes()
	if err != nil {
		return &Stream{err: err}
	}
	result := fn(string(data))
	return &Stream{
		input:    strings.NewReader(result),
		timeout:  s.timeout,
		dir:      s.dir,
		env:      s.env,
		maxBytes: s.maxBytes,
	}
}

// Filter applies a line filter function.
func (s *Stream) Filter(fn func(string) bool) *Stream {
	return s.Transform(func(input string) string {
		var lines []string
		for _, line := range strings.Split(input, "\n") {
			if fn(line) {
				lines = append(lines, line)
			}
		}
		return strings.Join(lines, "\n")
	})
}

// Grep filters lines matching a substring.
func (s *Stream) Grep(pattern string) *Stream {
	return s.Filter(func(line string) bool {
		return strings.Contains(line, pattern)
	})
}

// bytes reads the stream content (with limit if set).
func (s *Stream) bytes() ([]byte, error) {
	if s.input == nil {
		return nil, nil
	}
	if s.maxBytes > 0 {
		return io.ReadAll(io.LimitReader(s.input, s.maxBytes))
	}
	return io.ReadAll(s.input)
}

// --- Terminal operations ---

// Drain returns the full Result with stdout, stderr, and exit code.
func (s *Stream) Drain() *Result {
	data, _ := s.bytes()
	if s.err != nil {
		exitCode := -1
		if exitErr, ok := s.err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return &Result{
			Stdout:   string(data),
			ExitCode: exitCode,
			Err:      s.err,
		}
	}
	return &Result{
		Stdout:   string(data),
		ExitCode: 0,
	}
}

// String returns the output as a trimmed string.
func (s *Stream) String() (string, error) {
	data, err := s.bytes()
	if err != nil {
		return "", err
	}
	if s.err != nil {
		return strings.TrimRight(string(data), "\n\r\t "), s.err
	}
	return strings.TrimRight(string(data), "\n\r\t "), nil
}

// Bytes returns the raw output bytes.
func (s *Stream) Bytes() ([]byte, error) {
	data, err := s.bytes()
	if err != nil {
		return nil, err
	}
	return data, s.err
}

// Lines returns output split into lines.
func (s *Stream) Lines() ([]string, error) {
	str, err := s.String()
	if err != nil {
		return nil, err
	}
	if str == "" {
		return nil, nil
	}
	return strings.Split(str, "\n"), nil
}

// JSON parses output as JSON into Data.
func (s *Stream) JSON() (Data, error) {
	data, err := s.bytes()
	if err != nil {
		return nil, err
	}
	if s.err != nil {
		return nil, s.err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return Data(m), nil
}

// JSONArray parses output as JSON array.
func (s *Stream) JSONArray() (Array, error) {
	data, err := s.bytes()
	if err != nil {
		return nil, err
	}
	if s.err != nil {
		return nil, s.err
	}
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil, err
	}
	return Array(arr), nil
}

// WriteTo writes the output to a writer.
func (s *Stream) WriteTo(w io.Writer) (int64, error) {
	if s.input == nil {
		return 0, s.err
	}
	// Stream directly without buffering
	n, err := io.Copy(w, s.input)
	if err != nil {
		return n, err
	}
	return n, s.err
}

// WriteFile writes the output to a file.
func (s *Stream) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = s.WriteTo(f)
	return err
}

// AppendFile appends the output to a file.
func (s *Stream) AppendFile(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = s.WriteTo(f)
	return err
}

// OK returns true if there was no error in the pipeline.
func (s *Stream) OK() bool {
	return s.err == nil
}

// Err returns any error from the pipeline.
func (s *Stream) Err() error {
	return s.err
}
