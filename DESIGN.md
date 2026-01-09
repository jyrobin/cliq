# cliq

## Purpose

CLI toolkit library providing reusable components for building interactive command-line applications and writing process-level tests. Five packages: terminal utilities (zero deps), interactive menus (promptui), man-page style guides (yaml), shell-like utilities for commands/HTTP/WebSocket, and Cobra integration helpers.

Designed for reuse across CLI tools and for writing Go tests that invoke external commands.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         cliq                                 │
├─────────────────────────────────────────────────────────────┤
│  term/              │ Terminal utilities (→ x/term)         │
│  ├── style.go       │ ANSI colors: Dim, Bold, Success, etc  │
│  ├── prompt.go      │ Input: WaitForEnter, Confirm, ReadLine│
│  ├── box.go         │ Box drawing, Copyable commands        │
│  ├── multiselect.go │ Interactive checkbox list             │
│  ├── select.go      │ Interactive single-select menu        │
│  ├── autocomplete.go│ Input with tab-completion             │
│  └── completers.go  │ Pre-built completers (prefix, fuzzy)  │
├─────────────────────────────────────────────────────────────┤
│  menu/              │ Interactive menus (→ term/)           │
│  ├── types.go       │ Config, Category, Item, Options       │
│  ├── menu.go        │ Menu.Run(), LoadConfig()              │
│  └── exec.go        │ Command execution, input collection   │
├─────────────────────────────────────────────────────────────┤
│  guide/             │ Man-page style guides (→ yaml)        │
│  ├── types.go       │ Index, Content, Section, Command      │
│  └── guide.go       │ Guide.ShowIndex(), ShowTopic()        │
├─────────────────────────────────────────────────────────────┤
│  sh/                │ Shell-like utilities (→ yaml)         │
│  ├── cmd.go         │ Run, Chain, Command builder           │
│  ├── pipe.go        │ Stream pipeline: Exec, From, Pipe     │
│  ├── http.go        │ HTTP client with JSON/form helpers    │
│  ├── data.go        │ Generic JSON/YAML parsing, Data type  │
│  └── tail.go        │ Multi-file tail with fsnotify         │
├─────────────────────────────────────────────────────────────┤
│  ws/                │ WebSocket utilities (→ websocket)     │
│  ├── conn.go        │ Conn with Send, Recv, message parsing │
│  └── state.go       │ Connection state management           │
├─────────────────────────────────────────────────────────────┤
│  cobrautil/         │ Cobra integration (→ cobra)           │
│  ├── menu.go        │ MenuCommand(), MenuCommandWithLoader()│
│  ├── guide.go       │ GuideCommand(), GuideHandler()        │
│  └── common.go      │ VersionCommand(), flag helpers        │
└─────────────────────────────────────────────────────────────┘

Dependency flow:
  term/ (→ x/term) ← menu/ ← cobrautil/
                     guide/ ←─┘
  sh/ (standalone, uses yaml + fsnotify)
  ws/ (standalone, uses websocket)
```

## Directory Structure

```
cliq/
├── go.mod           # github.com/jyrobin/cliq
├── cmd/cliq/        # CLI binary (go install github.com/jyrobin/cliq/cmd/cliq@latest)
│   ├── main.go
│   ├── root.go
│   └── tail.go      # cliq tail <pattern>
├── term/            # Terminal utilities (→ golang.org/x/term)
│   ├── style.go     # ANSI styling functions
│   ├── prompt.go    # User input utilities
│   ├── box.go       # Box drawing and copyable content
│   ├── multiselect.go # Interactive multi-select with checkboxes
│   ├── select.go    # Interactive single-select menu
│   ├── autocomplete.go # Input with tab-completion suggestions
│   └── completers.go # Pre-built completers (prefix, fuzzy, path)
├── menu/            # Interactive menu system
│   ├── types.go     # YAML-loadable configuration types
│   ├── menu.go      # Menu rendering with term/select
│   └── exec.go      # Command execution with term/autocomplete
├── guide/           # Man-page style documentation
│   ├── types.go     # Index, Content, Section, Command types
│   └── guide.go     # Rendering and topic lookup
├── sh/              # Shell-like utilities for tests/scripts
│   ├── cmd.go       # Command execution, piping, chaining
│   ├── pipe.go      # Stream pipeline: Exec, From, Pipe, Transform
│   ├── http.go      # HTTP client helpers
│   ├── data.go      # Generic JSON/YAML Data type
│   └── tail.go      # Multi-file tail with fsnotify
├── ws/              # WebSocket utilities (→ gorilla/websocket)
│   ├── conn.go      # Conn with Send, Recv, JSON parsing
│   └── state.go     # Generic state machine for connections
└── cobrautil/       # Cobra framework integration
    ├── menu.go      # Create menu commands
    ├── guide.go     # Create guide commands
    └── common.go    # Common CLI helpers
```

## Key Types/Interfaces

### menu.Config (YAML-driven menu definition)
```go
type Config struct {
    Title      string     `yaml:"title"`
    Categories []Category `yaml:"categories"`
}

type Item struct {
    Name     string  `yaml:"name"`
    Short    string  `yaml:"short"`
    Action   string  `yaml:"action"`   // run, prompt, guide, workflow, custom
    Command  string  `yaml:"command"`  // with {{.var}} placeholders
    Inputs   []Input `yaml:"inputs"`
    Topic    string  `yaml:"topic"`    // for guide action
    Workflow string  `yaml:"workflow"` // for workflow action
}
```

### menu.Options (runtime behavior)
```go
type Options struct {
    Size          int
    ActionHandler func(item *Item) error  // For custom actions
    BeforeRun     func(command string) bool
    AfterRun      func(command string, err error)
}
```

### guide.Guide (man-page style documentation)
```go
type Guide struct { /* manages fs.FS with YAML files */ }

func New(fsys fs.FS, opts Options) *Guide
func (g *Guide) ShowIndex() error           // List available topics
func (g *Guide) ShowTopic(topic string) error  // Display a topic
func (g *Guide) LoadIndex() (*Index, error)
func (g *Guide) LoadTopic(topic string) (*Content, error)
func (g *Guide) HasTopic(topic string) bool

type Content struct {
    Title       string    `yaml:"title"`
    Description string    `yaml:"description"`
    Sections    []Section `yaml:"sections"`    // Free-form content
    Categories  []Category `yaml:"categories"` // Command groups
    Flags       []Flag    `yaml:"flags"`       // Global flags
    Environment []EnvVar  `yaml:"environment"` // Env vars
}
```

### sh (shell-like utilities)
```go
// Command execution
func Run(name string, args ...string) *Result
func Pipe(input *Result, name string, args ...string) *Result
func Chain(cmds ...[]string) *Result

type Result struct {
    Stdout, Stderr string
    ExitCode       int
}
func (r *Result) OK() bool
func (r *Result) Lines() []string

// Command builder
cmd := Command("git", "status").Dir("/repo").Timeout(5*time.Second)
result := cmd.Run()

// Stream pipeline (sh/pipe.go)
output, _ := Exec("cat", "file.txt").
    Pipe("grep", "error").
    Pipe("wc", "-l").
    String()

// Stream from various sources
From("hello world").Pipe("wc", "-w").String()
FromFile("data.json").Pipe("jq", ".users").JSON()

// Transform and filter
Exec("ls", "-la").
    Filter(func(line string) bool { return strings.Contains(line, ".go") }).
    Lines()

Exec("cat", "log.txt").Grep("ERROR").String()

// HTTP client
client := HTTP().BaseURL("http://api").Auth("token")
resp := client.PostJSON("/users", data)
m, _ := resp.JSON()  // map[string]interface{}

// Generic data (JSON/YAML)
d, _ := ParseJSON(`{"user": {"name": "alice"}}`)
name := d.GetString("user.name")  // "alice"
d.Set("user.age", 30)
```

### ws (WebSocket utilities)
```go
// Simple dial
conn, _ := ws.Dial("ws://host/path")
defer conn.Close()

// With options
conn, _ := ws.NewDialer().
    Auth("token").
    Header("X-Custom", "value").
    Timeout(30 * time.Second).
    Dial("ws://host/path")

// Send/Receive
conn.Send(`{"action": "ping"}`)
conn.SendJSON(map[string]any{"action": "subscribe"})

msg := conn.Recv(5 * time.Second)
text := msg.Text()
data, _ := msg.JSONMap()
msgType := msg.MessageType()  // extracts "type" field

// Convenience methods
text, _ := conn.RecvText(5 * time.Second)
data, _ := conn.RecvJSON(5 * time.Second)

// Request/Response patterns
resp, _ := conn.ExpectJSON(`{"action": "ping"}`, 5*time.Second)
msg, _ := conn.ExpectType(request, 5*time.Second, "response")

// Collect until condition
msgs, _ := conn.RecvUntilType(30*time.Second, "complete")

// State machine for connection states
type ConnState string
const (
    StateConnecting ConnState = "connecting"
    StateConnected  ConnState = "connected"
    StateClosed     ConnState = "closed"
)

sm := ws.NewStateMachine(StateConnecting, map[ConnState][]ConnState{
    StateConnecting: {StateConnected, StateClosed},
    StateConnected:  {StateClosed},
})
sm.Transition(StateConnected)
```

### term utilities
```go
// Styling
func Dim(s string) string
func Bold(s string) string
func Success(s string) string  // "✓ " prefix, green
func Error(s string) string    // "✗ " prefix, red

// Input
func WaitForEnter()
func Confirm(prompt string, defaultYes bool) bool
func ReadLine(prompt string, defaultValue string) string

// Display
func Copyable(content string, hint string)
func Box(content string, width int, style BoxStyle) string

// Interactive multi-select (j/k or arrows, space=toggle, a=all, n=none, enter=done)
type SelectItem struct {
    ID          string
    Label       string
    Description string
    Required    bool   // Cannot be deselected
    Selected    bool   // Initial state
}

func MultiSelect(items []SelectItem, opts MultiSelectOptions) ([]string, error)
func MultiSelectDeselected(items []SelectItem, opts MultiSelectOptions) ([]string, error)

// Interactive single-select (replaces promptui.Select)
type SelectResult struct {
    Index     int
    Value     string
    Cancelled bool
}

func Select(label string, items []string, opts SelectOptions) (*SelectResult, error)
func SelectSimple(label string, items []string) (string, error)

// Autocomplete input with tab-completion
type AutocompleteResult struct {
    Value        string
    WasSuggested bool
    Cancelled    bool
}

type Completer func(input string) []string

func Autocomplete(completer Completer, opts AutocompleteOptions) (*AutocompleteResult, error)
func AutocompleteSimple(prompt string, options []string) (string, error)

// Pre-built completers
func PrefixCompleter(options []string, caseSensitive bool) Completer
func ContainsCompleter(options []string, caseSensitive bool) Completer
func FuzzyCompleter(options []string) Completer
func PathCompleter(dirsOnly bool) Completer
func CommandCompleter(commands []string, argCompleter func(cmd, args string) []string) Completer
```

## Dependencies

### Uses
- `github.com/spf13/cobra` - CLI framework integration (cobrautil/)
- `github.com/gorilla/websocket` - WebSocket client (ws/)
- `github.com/fsnotify/fsnotify` - file watching (sh/tail)
- `golang.org/x/term` - terminal raw mode (term/multiselect, term/select, term/autocomplete)
- `gopkg.in/yaml.v3` - YAML parsing (menu/, guide/, sh/)

Note: `github.com/manifoldco/promptui` was replaced by native term/ implementations.

### Used by
- CLI tools using interactive menus and guides
- Go tests for process-level testing (sh/)

## Boundaries

### Belongs here
- Terminal styling and formatting
- Interactive prompts and menus
- Man-page style documentation
- Command execution and piping
- HTTP/WebSocket client helpers
- Generic JSON/YAML data manipulation
- Cobra integration utilities

### Does NOT belong here
- Application-specific business logic
- Database or persistent storage
- Complex configuration management
- Domain-specific protocols
