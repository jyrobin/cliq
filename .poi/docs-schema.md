# POI Documentation Schema

This file defines the structure constraints for module documentation.
LLM assistants should follow these guidelines when creating or updating docs.

## DESIGN.md

Architecture and structure documentation.

### Required Sections

- **Purpose**: 1-2 paragraphs explaining what this module does and why it exists
- **Architecture Overview**: ASCII diagram or description of key components
- **Directory Structure**: Tree showing important files/folders with descriptions
- **Key Types/Interfaces**: Main types with brief Go code examples
- **Dependencies**: What this module uses and what uses it
- **Configuration**: Environment variables, flags, config files
- **Boundaries**: What belongs here vs. what doesn't

### Style

- Focus on "what" and "why", not "how" (code shows how)
- Keep current - update when architecture changes
- Include code examples for non-obvious APIs
- ASCII diagrams preferred over external images

## NOTES.md

Operational knowledge and gotchas.

### Required Sections

- **Gotchas**: Common pitfalls with before/after code examples
- **Debugging**: How to diagnose common issues
- **Historical Decisions**: Why things are the way they are
- **Testing**: How to test this module
- **Key Files**: Table of important files and their purposes

### Style

- Each gotcha should have: title, wrong way, right way, explanation
- Be specific - reference actual file paths and line numbers
- Update after incidents or surprising behaviors
- Include commands that help with debugging

## .summary.yaml

Structured extraction from DESIGN.md and NOTES.md.

### Schema

```yaml
module:
  name: <module-name>           # Required
  path: <relative-path>         # Required
  description: <1-2 sentences>  # Required
  status: active|deprecated|experimental

tags: [<tag1>, <tag2>, ...]     # Searchable keywords

dependencies:
  uses: [<dep1>, <dep2>]        # What this module imports/uses
  used_by: [<dep1>, <dep2>]     # What imports/uses this module

entities:                        # Key types, functions, components
  - name: <EntityName>
    file: <path/to/file.go>
    tags: [<tag1>, <tag2>]
    description: <1 sentence>

gotchas:                         # Extracted from NOTES.md
  - id: <kebab-case-id>
    tags: [<tag1>, <tag2>]
    summary: <1 sentence>
    section: <section heading in NOTES.md>
```

### Guidelines

- Extract, don't invent - all content should trace to DESIGN.md or NOTES.md
- Keep descriptions concise (1-2 sentences max)
- Tags should be lowercase, hyphenated
- Entity files should be relative paths from module root
- Gotcha IDs should be unique within the module
