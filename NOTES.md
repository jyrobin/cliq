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
command: "poi bootstrap --module ${module}"

# RIGHT
command: "poi bootstrap --module {{.module}}"
```

### Input IDs Must Match Template Variables
```yaml
# WRONG - ID mismatch
inputs:
  - id: path        # ID is "path"
command: "poi scan {{.module}}"  # Template uses "module"

# RIGHT - IDs match
inputs:
  - id: module
command: "poi scan {{.module}}"
```

### term Package Has Zero Dependencies
The term/ package intentionally has no external dependencies. Don't add imports:
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
| menu/types.go | Config structures for YAML menus |
| menu/menu.go | Main menu rendering loop |
| menu/exec.go | Command execution and input collection |
| cobrautil/menu.go | Create Cobra commands from menu config |
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

### promptui Over Other Libraries
Chose promptui because:
- Active maintenance
- Simple API for select/prompt
- Works well in terminals
- Used by other popular CLIs
