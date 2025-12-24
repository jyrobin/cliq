# cliq

## Purpose

CLI toolkit library providing reusable components for building interactive command-line applications. Four packages with layered dependencies: terminal utilities (zero deps), interactive menus (promptui), man-page style guides (yaml), and Cobra integration helpers.

Designed for reuse across CLI tools like poi, powertalk-cli, and future CLIs.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         cliq                                 │
├─────────────────────────────────────────────────────────────┤
│  term/              │ Terminal utilities (zero deps)        │
│  ├── style.go       │ ANSI colors: Dim, Bold, Success, etc  │
│  ├── prompt.go      │ Input: WaitForEnter, Confirm, ReadLine│
│  └── box.go         │ Box drawing, Copyable commands        │
├─────────────────────────────────────────────────────────────┤
│  menu/              │ Interactive menus (→ promptui)        │
│  ├── types.go       │ Config, Category, Item, Options       │
│  ├── menu.go        │ Menu.Run(), LoadConfig()              │
│  └── exec.go        │ Command execution, input collection   │
├─────────────────────────────────────────────────────────────┤
│  guide/             │ Man-page style guides (→ yaml)        │
│  ├── types.go       │ Index, Content, Section, Command      │
│  └── guide.go       │ Guide.ShowIndex(), ShowTopic()        │
├─────────────────────────────────────────────────────────────┤
│  cobrautil/         │ Cobra integration (→ cobra)           │
│  ├── menu.go        │ MenuCommand(), MenuCommandWithLoader()│
│  ├── guide.go       │ GuideCommand(), GuideHandler()        │
│  └── common.go      │ VersionCommand(), flag helpers        │
└─────────────────────────────────────────────────────────────┘

Dependency flow:
  term/ ← menu/ ← cobrautil/
           guide/ ←─┘
    │        │         │
    └────────┴─────────┴──→ External: promptui, cobra, yaml
```

## Directory Structure

```
cliq/
├── go.mod           # github.com/jyrobin/cliq
├── term/            # Zero external dependencies
│   ├── style.go     # ANSI styling functions
│   ├── prompt.go    # User input utilities
│   └── box.go       # Box drawing and copyable content
├── menu/            # Interactive menu system
│   ├── types.go     # YAML-loadable configuration types
│   ├── menu.go      # Menu rendering with promptui
│   └── exec.go      # Command execution engine
├── guide/           # Man-page style documentation
│   ├── types.go     # Index, Content, Section, Command types
│   └── guide.go     # Rendering and topic lookup
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
    Name    string  `yaml:"name"`
    Short   string  `yaml:"short"`
    Action  string  `yaml:"action"`  // run, prompt, custom
    Command string  `yaml:"command"` // with {{.var}} placeholders
    Inputs  []Input `yaml:"inputs"`
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
```

## Dependencies

### Uses
- `github.com/manifoldco/promptui` - interactive select/prompt (menu/)
- `github.com/spf13/cobra` - CLI framework integration (cobrautil/)
- `gopkg.in/yaml.v3` - YAML parsing for menu config (menu/)

### Used by
- poi - documentation CLI (interactive menu)
- powertalk-cli - diagnostic CLI (planned)

## Boundaries

### Belongs here
- Terminal styling and formatting
- Interactive prompts and menus
- Command execution helpers
- Cobra integration utilities
- YAML-driven menu configuration

### Does NOT belong here
- Application-specific business logic
- Configuration management beyond menus
- Network/HTTP utilities
- File system operations
