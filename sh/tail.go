package sh

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// TailOptions configures tail behavior.
type TailOptions struct {
	// Colors for different files (cycles through). Empty = no colors.
	Colors []string
	// Prefix format: "name" (default), "path", "none"
	Prefix string
	// FromStart starts from beginning of file instead of end
	FromStart bool
	// OnLine callback for each line (alternative to stdout)
	OnLine func(file, line string)
}

// DefaultTailColors provides distinct colors for up to 6 files.
var DefaultTailColors = []string{
	"\033[32m", // green
	"\033[33m", // yellow
	"\033[34m", // blue
	"\033[35m", // magenta
	"\033[36m", // cyan
	"\033[91m", // bright red
}

const resetColor = "\033[0m"

// Tail follows multiple files matching a glob pattern, merging output.
// Blocks until context is cancelled.
func Tail(ctx context.Context, pattern string, opts TailOptions) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no files match pattern: %s", pattern)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Track file states
	type fileState struct {
		path   string
		name   string
		color  string
		offset int64
		mu     sync.Mutex
	}

	states := make(map[string]*fileState)
	colors := opts.Colors
	if colors == nil {
		colors = DefaultTailColors
	}

	for i, path := range files {
		absPath, _ := filepath.Abs(path)

		// Determine starting offset
		var offset int64
		if !opts.FromStart {
			if info, err := os.Stat(absPath); err == nil {
				offset = info.Size()
			}
		}

		// Determine prefix
		name := filepath.Base(path)
		if opts.Prefix == "path" {
			name = path
		} else if opts.Prefix == "none" {
			name = ""
		}

		// Assign color
		color := ""
		if len(colors) > 0 {
			color = colors[i%len(colors)]
		}

		states[absPath] = &fileState{
			path:   absPath,
			name:   name,
			color:  color,
			offset: offset,
		}

		// Watch the directory (handles file recreation/rotation)
		dir := filepath.Dir(absPath)
		watcher.Add(dir)
	}

	// Function to read new content from a file
	readNew := func(state *fileState) {
		state.mu.Lock()
		defer state.mu.Unlock()

		f, err := os.Open(state.path)
		if err != nil {
			return
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return
		}

		// Handle file truncation (log rotation)
		if info.Size() < state.offset {
			state.offset = 0
		}

		if info.Size() <= state.offset {
			return
		}

		f.Seek(state.offset, io.SeekStart)
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			line := scanner.Text()
			if opts.OnLine != nil {
				opts.OnLine(state.name, line)
			} else {
				if state.name == "" {
					fmt.Println(line)
				} else if state.color != "" {
					fmt.Printf("%s[%s]%s %s\n", state.color, state.name, resetColor, line)
				} else {
					fmt.Printf("[%s] %s\n", state.name, line)
				}
			}
		}

		state.offset, _ = f.Seek(0, io.SeekCurrent)
	}

	// Initial read (especially for FromStart)
	for _, state := range states {
		readNew(state)
	}

	// Watch for changes
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			absPath, _ := filepath.Abs(event.Name)
			if state, exists := states[absPath]; exists {
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					readNew(state)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			// Log but continue
			fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
		}
	}
}

// TailFile is a convenience for tailing a single file.
func TailFile(ctx context.Context, path string, opts TailOptions) error {
	return Tail(ctx, path, opts)
}
