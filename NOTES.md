# cliq Notes

## Gotchas

### Custom Actions Require Handler
Menu items with custom actions (not "run" or "prompt") need an ActionHandler:
```go
// WRONG - custom action silently fails
opts := menu.Options{}
m := menu.New(cfg, opts)

// RIGHT - provide handler for custom actions
opts := menu.Options{
    ActionHandler: func(item *menu.Item) error {
        if item.Action == "guide" {
            return showGuide(item.Topic)
        }
        return fmt.Errorf("unknown action: %s", item.Action)
    },
}
```

### Template Variables Use Go Template Syntax
Command templates use `{{.var}}` not `${var}`:
```yaml
# WRONG
command: "mycli build --target ${target}"

# RIGHT
command: "mycli build --target {{.target}}"
```

### Input IDs Must Match Template Variables
```yaml
# WRONG - ID mismatch
inputs:
  - id: path        # ID is "path"
command: "mycli scan {{.target}}"  # Template uses "target"

# RIGHT - IDs match
inputs:
  - id: target
command: "mycli scan {{.target}}"
```

### term Package Uses Only golang.org/x/term
The term/ package only depends on `golang.org/x/term` for raw terminal mode. Don't add other external dependencies:
```go
// WRONG - adds external dependency to term/
import "github.com/fatih/color"

// RIGHT - use built-in ANSI codes
const dim = "\033[2m"
```

### Guide Requires index.yaml
The guide package expects an `index.yaml` in the embedded filesystem:
```yaml
# guide/index.yaml - required
title: My CLI Guide
description: |
  Help text shown at the top.
topics:
  - id: workflow     # Maps to guide/workflow.yaml
    title: Workflows
    short: Step-by-step guides
```

### Guide Filesystem Prefix
When creating a Guide, the Options.Prefix must match your embed path:
```go
//go:embed docs/*.yaml      // Files in docs/
var docsFS embed.FS

// WRONG - default prefix is "guide"
g := guide.New(docsFS, guide.Options{})

// RIGHT - match your embed path
g := guide.New(docsFS, guide.Options{Prefix: "docs"})
```

### sh.Result Error vs ExitCode
Check both Err and ExitCode - they indicate different failures:
```go
result := sh.Run("nonexistent-cmd")
// result.Err != nil      (command not found)
// result.ExitCode == -1

result := sh.Run("false")  // exits with code 1
// result.Err == nil       (command ran successfully)
// result.ExitCode == 1    (but returned error code)

// Use OK() to check both
if !result.OK() { ... }
```

### sh.Data Path Syntax
Data.Get() uses dot-separated paths with array indexing:
```go
d, _ := ParseJSON(`{"users": [{"name": "alice"}, {"name": "bob"}]}`)
d.Get("users.0.name")  // "alice"
d.Get("users.1.name")  // "bob"
d.Get("users.2.name")  // nil (out of bounds)
```

### WebSocket Read Loop
ws.Conn starts a goroutine on Dial(). Always call Close():
```go
conn, err := ws.Dial("ws://host/path")
if err != nil { ... }
defer conn.Close()  // Important: stops read goroutine
```

### Select and Autocomplete Return Cancelled State
Unlike promptui which returns errors for cancellation, term.Select and term.Autocomplete return a result with Cancelled=true:
```go
// WRONG - treating cancellation as error
result, err := term.Select("Choose", items, opts)
if err != nil {
    return err  // Cancellation is NOT an error
}

// RIGHT - check Cancelled field
result, err := term.Select("Choose", items, opts)
if err != nil {
    return err  // Real error (terminal issue, etc.)
}
if result.Cancelled {
    return nil  // User pressed Esc/Ctrl+C - handle gracefully
}
value := result.Value
```

### Autocomplete Completer Can Be Nil
If you don't need suggestions, pass nil for the completer:
```go
// Just an input field with no suggestions
result, err := term.Autocomplete(nil, term.AutocompleteOptions{
    Prompt: "Enter name: ",
})
```

## Debugging

### Menu Not Showing Items
- Check YAML indentation (items must be under categories)
- Verify Config loaded without error from LoadConfig()
- Check Size option (default 10, may hide items)

### Command Execution Fails
- Check command template substitution with SubstituteValues()
- Verify all Input IDs have corresponding template variables
- Test command string manually before menu execution

### ANSI Colors Not Showing
- Terminal must support ANSI escape codes
- Piping output may strip colors
- Some CI environments disable colors

## Testing

```bash
# Build and test all packages
go test ./...

# Test specific package
go test ./menu/...
go test ./term/...
```

### Testing Menus
Menus are interactive - test components separately:
- Test LoadConfig() with sample YAML
- Test SubstituteValues() with known inputs
- Test RunCommand() with safe commands

## Key Files

| File | Purpose |
|------|---------|
| term/style.go | ANSI color and styling functions |
| term/prompt.go | User input utilities (WaitForEnter, Confirm) |
| term/box.go | Box drawing for copyable content |
| term/multiselect.go | Interactive multi-select with checkboxes |
| term/select.go | Interactive single-select menu (replaces promptui) |
| term/autocomplete.go | Input with tab-completion suggestions |
| term/completers.go | Pre-built completers (prefix, fuzzy, path) |
| menu/types.go | Config structures for YAML menus |
| menu/menu.go | Main menu rendering loop (uses term/select) |
| menu/exec.go | Command execution (uses term/autocomplete) |
| guide/types.go | Guide index and content types |
| guide/guide.go | Guide rendering and loading |
| sh/cmd.go | Run, Pipe, Chain, Command builder |
| sh/pipe.go | Stream pipeline: Exec, From, Pipe, Transform |
| sh/http.go | HTTPClient with JSON/form helpers |
| sh/data.go | Generic Data type for JSON/YAML |
| ws/conn.go | WebSocket Conn with Send, Recv, JSON parsing |
| ws/state.go | Generic state machine for connection states |
| cobrautil/menu.go | Create Cobra commands from menu config |
| cobrautil/guide.go | Create Cobra commands for guides |
| cobrautil/common.go | Version command, flag helpers |

## Historical Decisions

### Three Packages Instead of One
Separated into term/, menu/, cobrautil/ to allow:
- term/ usable without promptui dependency
- menu/ usable without Cobra
- cobrautil/ optional for non-Cobra CLIs

### YAML-Driven Menus
Menu configuration in YAML allows:
- Non-programmers to modify menus
- Embedding with go:embed
- Runtime loading without recompilation

### Native term/ Implementation Over promptui
Replaced promptui with native term/select and term/autocomplete because:
- Reduces external dependencies (only golang.org/x/term needed)
- Full control over rendering and behavior
- Consistent cancellation handling (Cancelled field vs error)
- Better integration with other term/ utilities
- Fuzzy matching and path completion built-in

The menu/ package now uses term/select for menus and term/autocomplete for input collection.
