# Implementation Plan: User Configuration File

**Branch**: `007-user-wants-to` | **Date**: 2025-10-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-user-wants-to/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add YAML-based user configuration system allowing customization of terminal size warning thresholds and keyboard shortcuts. Config lives at `~/.config/pass-cli/config.yml` (or OS equivalent), is optional with hardcoded defaults, validates on startup with modal warnings for errors, and includes CLI management commands (init, edit, validate, reset). All UI elements (status bar, help modal, form hints) automatically reflect custom keybindings.

## Technical Context

**Language/Version**: Go 1.25.1
**Primary Dependencies**:
- `github.com/spf13/viper v1.21.0` (YAML config loading - already in project)
- `github.com/rivo/tview v0.42.0` (TUI modal warnings - already in project)
- `github.com/gdamore/tcell/v2 v2.9.0` (keyboard event handling - already in project)
- `github.com/spf13/cobra v1.10.1` (CLI commands - already in project)

**Storage**: YAML file at `~/.config/pass-cli/config.yml` (Linux/macOS) or `%APPDATA%\pass-cli\config.yml` (Windows)
**Testing**: Go standard testing (`go test`), table-driven tests for validation logic
**Target Platform**: Cross-platform (Windows, macOS, Linux) - must respect OS-specific config locations and default editors
**Project Type**: Single project with CLI + TUI frontends
**Performance Goals**: Config load <10ms, validation <5ms (non-blocking startup)
**Constraints**:
- Config file size limit: 100 KB hard limit
- Validation must not block TUI startup (show modal, continue with defaults)
- Must maintain backward compatibility (missing config = use defaults)

**Scale/Scope**:
- ~20 configurable settings (3 terminal, ~10 keybindings, potential future expansion)
- Single global config per user (not per-vault)
- Expected typical config file size: 2-5 KB with comments

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Security-First Development ✅

**Status**: PASS - No security concerns

- Config file contains no secrets (only UI preferences)
- No credential data stored or transmitted
- File permissions will follow OS defaults (user-only access recommended but not critical)
- Validation errors do not leak sensitive information
- No encryption required for config content

### Principle II: Library-First Architecture ✅

**Status**: PASS - Follows library-first pattern

**Plan**:
- Create `internal/config/` package for config loading, validation, and defaults
- Config package has no CLI/TUI dependencies
- Public API: `Load()`, `Validate()`, `GetDefaults()`, `Merge()`
- Keybinding parsing and validation isolated in config package
- CLI commands in `cmd/config.go` consume library, do not embed logic

### Principle III: CLI Interface Standards ✅

**Status**: PASS - Follows CLI conventions

**Plan**:
- New subcommands: `pass-cli config [init|edit|validate|reset]`
- Output: Success/error messages to stdout/stderr
- Exit codes: 0=success, 1=validation error, 2=file system error
- Non-interactive: All commands work in scripts (no TTY required)
- Validation output includes structured error messages with line numbers

### Principle IV: Test-Driven Development (NON-NEGOTIABLE) ✅

**Status**: PASS - TDD approach planned

**Plan**:
- Unit tests for config validation (conflicting keys, unknown actions, size limits)
- Unit tests for YAML parsing edge cases (empty file, malformed syntax, partial config)
- Unit tests for keybinding string-to-tcell conversion
- Integration tests for CLI commands (init creates file, edit opens editor, validate reports errors)
- Integration tests for TUI modal display on validation failure
- Contract tests for config schema stability
- Target: >80% coverage for `internal/config/` package

### Principle V: Cross-Platform Compatibility ✅

**Status**: PASS - Cross-platform considerations included

**Plan**:
- Use `os.UserConfigDir()` for cross-platform config paths
- Editor selection: Check `EDITOR` env var, fallback to OS defaults via `os/exec` platform detection
- Test on Windows (notepad), macOS (nano/vim), Linux (nano/vi)
- Path handling with `filepath.Join` throughout
- CI tests on all platforms (Windows, macOS, Linux)

### Principle VI: Observability & Auditability ✅

**Status**: PASS - Audit logging considerations

**Plan**:
- Log config load attempts to audit log (file path, success/failure)
- Log validation errors (without exposing config content)
- Verbose mode shows config load details (path, loaded settings summary)
- No credential logging (config contains none)

### Principle VII: Simplicity & YAGNI ✅

**Status**: PASS - Minimal complexity

**Plan**:
- Use existing `spf13/viper` library (already in project) - no new dependencies
- Flat config structure (no nested complexity beyond 2 levels)
- No hot-reload (changes apply on next startup - simpler)
- No config versioning/migration for v1 (add if needed in future)
- Backup strategy: Simple overwrite to `.backup` (no timestamped history)

**Re-evaluation Post-Phase 1**: _(To be completed after design phase)_

## Project Structure

### Documentation (this feature)

```
specs/007-user-wants-to/
├── spec.md              # Feature specification (complete)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (PENDING)
├── data-model.md        # Phase 1 output (PENDING)
├── quickstart.md        # Phase 1 output (PENDING)
├── contracts/           # Phase 1 output (PENDING)
│   └── config-schema.yml
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
# Existing structure (Single project with CLI + TUI)
cmd/
├── config.go            # NEW: Config management commands (init, edit, validate, reset)
├── tui/
│   ├── main.go          # MODIFIED: Load config, display validation modal if errors
│   ├── events/
│   │   └── handlers.go  # MODIFIED: Use config keybindings instead of hardcoded
│   └── components/
│       ├── statusbar.go # MODIFIED: Reflect custom keybindings in hints
│       └── forms.go     # MODIFIED: Reflect custom keybindings in form hints
└── ...

internal/
├── config/              # NEW: Config library package
│   ├── config.go        # Config struct, Load(), Validate(), Defaults()
│   ├── config_test.go   # Unit tests for config logic
│   ├── keybinding.go    # Keybinding parsing, tcell conversion, conflict detection
│   └── keybinding_test.go
└── ...

tests/
└── config/              # NEW: Integration tests
    ├── cli_test.go      # Test config commands end-to-end
    └── validation_test.go

# Example config file (for reference, not in repo)
~/.config/pass-cli/config.yml  # User's config (Linux/macOS)
%APPDATA%\pass-cli\config.yml  # User's config (Windows)
```

**Structure Decision**: Pass-CLI follows a single-project structure with `cmd/` for CLI/TUI entry points and `internal/` for library packages. This aligns with Go conventions and maintains clear separation between interface (cmd) and business logic (internal). The new config feature fits naturally into this pattern with `internal/config/` for the library and `cmd/config.go` for CLI commands.

## Complexity Tracking

*No violations of Constitution principles. This section is empty.*

---

## Phase 0: Research & Architecture Decisions

### Research Questions

1. **YAML Library Best Practices**: How to use `spf13/viper` for config loading with validation?
   - Viper supports YAML out of box, defaults, merging
   - Need to investigate: schema validation, line number error reporting, size limit checking

2. **Keybinding String Format**: What string format for keybindings in YAML?
   - Options: `"a"`, `"ctrl+a"`, `"alt+shift+f1"`
   - Research tcell key parsing, modifier representation
   - Consider case sensitivity, platform differences

3. **OS Config Directory**: How to determine cross-platform config paths?
   - Go `os.UserConfigDir()` returns appropriate path per platform
   - Need to handle: directory creation, permission errors, fallback behavior

4. **Default Editor Selection**: How to detect OS-default editor when `EDITOR` not set?
   - Research platform-specific defaults
   - Error handling when no editor available

5. **Modal Warning UX**: How to display validation errors in tview without blocking?
   - Research tview modal patterns (already have size warning modal as example)
   - Ensure error messages are readable, dismissible, don't prevent app usage

### Technology Choices

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Config Format | YAML | Human-readable, supports comments, viper native support |
| Config Library | spf13/viper | Already in project, mature, supports defaults/merging |
| Config Location | OS UserConfigDir | Cross-platform standard (XDG on Linux, AppData on Windows, Library on macOS) |
| Validation Strategy | Load-time with modal warning | Non-blocking, graceful degradation, aligns with size warning pattern |
| Keybinding Storage | Action-to-key map | Natural for users, enables conflict detection |
| Backup Strategy | Simple `.backup` overwrite | YAGNI - no need for timestamped history yet |
| Size Limit | 100 KB hard reject | Prevents DoS, allows 20-50x growth, catches user errors |

### Architecture Decisions

**Config Loading Flow**:
1. App startup → `config.Load()`
2. Check file exists → if not, return defaults (no error)
3. Check file size → if >100KB, return error + defaults
4. Parse YAML → if syntax error, return error + defaults
5. Validate schema → if validation error, return error + defaults
6. Merge with defaults → return merged config
7. TUI displays modal if errors exist, then continues

**Keybinding Resolution**:
- Config file: `keybindings.add_credential: "n"`
- Parser: String `"n"` → tcell.KeyRune('n')
- Runtime: Event matches tcell.KeyRune('n') → trigger add action
- UI update: Status bar reads config, shows "n: Add" instead of "a: Add"

**Error Handling Philosophy**:
- Config errors are non-fatal (app must always work)
- Show clear error message in TUI modal
- Log error to audit log for troubleshooting
- Continue with defaults (safe fallback)

---

## Phase 1: Data Model & Contracts

_(Phase 1 outputs will be generated in research.md, data-model.md, contracts/, and quickstart.md)_

---

## Phase 2: Task Generation

_(Phase 2 happens via `/speckit.tasks` command - not part of this plan)_
