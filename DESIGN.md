# cliq

## Purpose

CLI toolkit library - terminal utilities, interactive menus, man-page guides, shell-like utilities, and Cobra integration helpers

**Tags:** cli, terminal, menu, guide, sh, http, websocket, json, yaml, tail, fsnotify, prompts, toolkit, go, tui, interactive, state-machine, pipeline

## Overview

cliq is a CLI toolkit library — reusable building blocks for
interactive command-line apps and for writing process-level Go tests
that drive external commands. It's not a single tool; it's five
loosely-coupled packages plus a tiny `cmd/cliq` demo binary.

The five packages are intentionally narrow:

- **term/** — terminal primitives (ANSI styling, prompts, boxes,
  single/multi-select menus, autocomplete) over `golang.org/x/term`,
  with no other deps
- **menu/** — YAML-driven interactive menus built on `term/` and
  `promptui`, with categories, items, and templated commands
- **guide/** — man-page style documentation rendered from YAML
- **sh/** — shell-like helpers for tests and scripts: command run /
  chain / pipe, an HTTP client with JSON+form helpers, generic
  JSON/YAML `Data`, and a multi-file `tail` over `fsnotify`
- **ws/** — minimal WebSocket `Conn` + state machine
- **cobrautil/** — adapters that turn `menu` and `guide` into
  Cobra subcommands

Consumers pick what they need. The dependency flow is shallow:
`term ← menu ← cobrautil`, `guide ← cobrautil`, with `sh/` and `ws/`
standalone. The library is designed to be embedded in other CLIs
across the workspace, not to be a framework.

## Key Entities

- **Menu** (`menu/menu.go`) — Interactive menu instance with config and options
- **Config** (`menu/types.go`) — YAML-loadable menu configuration with categories and items
- **Options** (`menu/types.go`) — Runtime options including ActionHandler for custom actions
- **SubstituteValues** (`menu/exec.go`) — Replace {{.key}} placeholders with values from map
- **Guide** (`guide/guide.go`) — Man-page style documentation from embedded YAML files
- **Content** (`guide/types.go`) — Guide topic with sections, categories, flags, and environment vars
- **Dim** (`term/style.go`) — Return dimmed ANSI-styled text
- **MultiSelect** (`term/multiselect.go`) — Interactive multi-select with checkboxes, returns selected IDs
- **MultiSelectDeselected** (`term/multiselect.go`) — Like MultiSelect but returns deselected IDs (for skip lists)
- **SelectItem** (`term/multiselect.go`) — Item for multi-select with ID, label, description, required flag
- **Select** (`term/select.go`) — Interactive single-select menu with arrow/j-k navigation
- **SelectResult** (`term/select.go`) — Result from Select with Index, Value, and Cancelled state
- **Autocomplete** (`term/autocomplete.go`) — Input with tab-completion suggestions
- **Completer** (`term/autocomplete.go`) — Function type for providing autocomplete suggestions
- **FuzzyCompleter** (`term/completers.go`) — Completer using fuzzy matching with scoring
- **PathCompleter** (`term/completers.go`) — Completer for filesystem paths with ~ expansion
- **Copyable** (`term/box.go`) — Display content in a box with copy hint
- **MenuCommand** (`cobrautil/menu.go`) — Create a Cobra command that runs an interactive menu
- **GuideCommand** (`cobrautil/guide.go`) — Create a Cobra command for displaying guides
- **Run** (`sh/cmd.go`) — Execute a command and return Result with stdout/stderr/exitcode
- **Pipe** (`sh/cmd.go`) — Run command with input from previous Result's stdout
- **Command** (`sh/cmd.go`) — Create command builder with Dir, Env, Stdin, Timeout options
- **HTTPClient** (`sh/http.go`) — HTTP client with BaseURL, Auth, and JSON/form helpers
- **Conn** (`ws/conn.go`) — WebSocket connection with Send, Recv, RecvJSON, ExpectType
- **Dialer** (`ws/conn.go`) — WebSocket dialer with Auth, Header, Timeout options
- **Message** (`ws/conn.go`) — Received WebSocket message with Text, JSON, MessageType helpers
- **StateMachine** (`ws/state.go`) — Generic state machine with defined transitions and callbacks
- **StateHistory** (`ws/state.go`) — Tracks state transitions for debugging/testing
- **Stream** (`sh/pipe.go`) — Data stream that can be piped through commands with transform
- **Data** (`sh/data.go`) — Generic map with dot-path access (Get, Set, GetString, etc.)
- **Tail** (`sh/tail.go`) — Follow multiple files with colored output using fsnotify

## Gotchas

- **custom-action-handler** _(note)_ — Custom actions require ActionHandler in Options or they fail silently
- **template-syntax** _(note)_ — Use {{.var}} not ${var} for command template variables
- **input-id-match** _(note)_ — Input IDs must exactly match template variable names
- **term-only-xterm** _(note)_ — term/ package only uses golang.org/x/term, no other external deps
- **guide-index-required** _(note)_ — Guide package requires index.yaml in the embedded filesystem
- **guide-prefix-match** _(note)_ — Options.Prefix must match the embed directory path
- **sh-result-err-vs-exitcode** _(note)_ — Result.Err is for exec errors, ExitCode for command exit status - use OK()
- **sh-data-path-syntax** _(note)_ — Data.Get uses dot-separated paths with array indexing (e.g., users.0.name)
- **ws-close-required** _(note)_ — ws.Conn starts a goroutine on Dial - always call Close() to stop it
- **select-cancelled-not-error** _(note)_ — Select returns Cancelled=true on Esc/Ctrl+C, not an error
- **autocomplete-nil-completer** _(note)_ — Pass nil for completer if you just need an input field without suggestions

_Run `poi gotchas cliq` for full bodies and triggers._

## Dependencies

**Uses:**

- github.com/spf13/cobra
- github.com/gorilla/websocket
- github.com/fsnotify/fsnotify
- golang.org/x/term
- gopkg.in/yaml.v3

**Used by:**

- asterapi

---

_Generated by `poi render cliq` from `.poi.yaml` (schema v2). Edit `.poi.yaml`, not this file._
