# Implementation Plan: Doctor Command and First-Run Guided Initialization

**Branch**: `011-doctor-command-for` | **Date**: 2025-10-21 | **Status**: Planning Complete | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/011-doctor-command-for/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement `doctor` command for comprehensive vault health verification (binary version, vault accessibility, config validation, keychain status, orphaned entry detection, backup file status) and first-run detection that triggers guided initialization flow when no vault exists at default location. This replaces cryptic error messages with friendly prompts for new users while giving existing users proactive troubleshooting tools.

## Technical Context

**Language/Version**: Go 1.25.1 (existing codebase)
**Primary Dependencies**:
- Cobra (CLI framework, existing)
- go-keyring (OS keychain integration, existing)
- Viper (configuration management, existing)
- Standard library: `os`, `path/filepath`, `encoding/json`, `net/http`

**Storage**: File-based (encrypted vault files, config YAML, audit logs)
**Testing**: `go test` with `-tags=integration` for end-to-end tests
**Target Platform**: Cross-platform (Windows, macOS, Linux) - single binary per platform
**Project Type**: Single CLI application
**Performance Goals**:
- SC-001: Doctor command completes in <5 seconds
- SC-003: First-run guided initialization completes in <2 minutes
- Version check: <1 second (with network), instant fallback (offline)

**Build Configuration**:
- Version injection: `-ldflags "-X main.version=$(VERSION)"` or `-ldflags "-X cmd.version=$(VERSION)"`
- Version variable: Defined in `cmd/version.go` or `main.go` (verify existing location)
- Build tools: Standard `go build` or GoReleaser (existing setup)

**Constraints**:
- Offline-first: Doctor must work without network (skip version check gracefully)
- No root/admin privileges required
- Security-first: No credential logging even in doctor verbose mode
- Zero destructive actions: Doctor reports only, no automatic repairs

**Scale/Scope**:
- Single-vault model (default location only)
- 10 health checks (version, vault, config, keychain, backup)
- 2 interactive flows (doctor report, first-run guided init)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Security-First Development ✅
- **Pass**: Doctor command does not touch credentials (reads metadata only)
- **Pass**: First-run guided init delegates to existing `internal/vault.InitializeVault()` (no new crypto code)
- **Pass**: No credential logging (doctor reports vault presence, not contents)
- **Pass**: Master password clearing follows existing `defer crypto.ClearBytes(password)` pattern
- **Verification**: Security tests will verify no secrets in doctor output

### II. Library-First Architecture ✅
- **Pass**: Doctor health checks will be in `internal/health/` package (library-first)
- **Pass**: First-run detection logic in `internal/vault/firstrun.go` (reusable)
- **Pass**: CLI commands (`cmd/doctor.go`, `cmd/root.go` first-run wrapper) are thin wrappers
- **Verification**: Health check functions independently testable without Cobra

### III. CLI Interface Standards ✅
- **Pass**: Doctor outputs structured JSON with `--json` flag + human-readable default
- **Pass**: First-run prompts detect non-TTY stdin and fail fast (no pipe interaction)
- **Pass**: Exit codes: 0=healthy, 1=warnings, 2=errors, 3=security issues
- **Pass**: Script-friendly: `--quiet` mode for doctor (exit code only)
- **Verification**: Integration tests for TTY detection and output formats

### IV. Test-Driven Development ✅
- **Commitment**: All health check functions will have unit tests before implementation
- **Commitment**: First-run flow will have integration tests with mocked vault filesystem
- **Commitment**: Contract tests for doctor output format (JSON schema validation)
- **Coverage Target**: 80% minimum for `internal/health/` and `internal/vault/firstrun.go`
- **Security Tests**: Verify no credential leakage in doctor verbose mode

### V. Cross-Platform Compatibility ✅
- **Pass**: Uses existing `filepath.Join` and OS-agnostic path logic
- **Pass**: Version check via GitHub API works on all platforms (or offline fallback)
- **Pass**: Keychain detection uses existing `internal/keychain` cross-platform abstraction
- **Pass**: Home directory via existing `os.UserHomeDir()` (respects %USERPROFILE% / $HOME)
- **Verification**: CI tests on Windows, macOS, Linux (existing matrix)

### VI. Observability & Auditability ✅
- **Pass**: Doctor verbose mode (`--verbose`) shows detailed check execution (no secrets)
- **Pass**: Audit log captures first-run initialization event (operation type + timestamp only)
- **Pass**: Doctor reports logged to stdout (preserves for scripting/logging)
- **Enhancement**: Doctor could optionally write report to `~/.pass-cli/doctor.log` (non-blocking)
- **Verification**: Audit log tests verify first-run event format

### VII. Simplicity & YAGNI ✅
- **Pass**: No new dependencies (uses stdlib + existing Cobra/go-keyring)
- **Pass**: Flat structure: `internal/health/` for checks, `cmd/doctor.go` for CLI
- **Pass**: Direct implementation: health checks are simple file/API queries
- **Pass**: Avoids abstract health check "framework" (just functions returning status structs)
- **Justification**: No complexity violations to track

**Result**: ✅ All 7 principles satisfied. Proceed to Phase 0 research.

## Project Structure

### Documentation (this feature)

```
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
pass-cli/
├── cmd/                                # CLI commands (Cobra)
│   ├── doctor.go                       # NEW: Doctor command implementation
│   ├── root.go                         # MODIFY: Add first-run detection wrapper
│   ├── init.go                         # REFERENCE: Existing vault initialization
│   └── [other existing commands]
│
├── internal/                           # Internal library packages
│   ├── health/                         # NEW: Health check library
│   │   ├── checker.go                  # Health check coordinator
│   │   ├── version.go                  # Binary version check (GitHub API)
│   │   ├── vault.go                    # Vault accessibility check
│   │   ├── config.go                   # Config file validation
│   │   ├── keychain.go                 # Keychain status + orphan detection
│   │   └── backup.go                   # Backup file status check
│   │
│   ├── vault/                          # Existing vault operations
│   │   ├── firstrun.go                 # NEW: First-run detection logic
│   │   ├── vault.go                    # REFERENCE: InitializeVault()
│   │   └── [existing vault files]
│   │
│   ├── keychain/                       # Existing keychain integration
│   │   └── [existing cross-platform code]
│   │
│   ├── config/                         # Existing config handling
│   │   └── [existing config validation]
│   │
│   └── [other existing packages]
│
├── test/                               # Integration and unit tests
│   ├── doctor_test.go                  # NEW: Doctor command integration tests
│   ├── firstrun_test.go                # NEW: First-run flow integration tests
│   └── [existing test files]
│
└── main.go                             # Application entry point
```

**Structure Decision**: Single project (CLI application). This feature adds:
- New `internal/health/` package for health check library (Constitution Principle II: Library-First)
- New `cmd/doctor.go` command (thin wrapper over `internal/health/`)
- New `internal/vault/firstrun.go` for first-run detection logic
- Modifications to `cmd/root.go` to wrap command execution with first-run check
- Integration tests in `test/` directory following existing patterns

**Key Files**:
- `internal/health/checker.go`: Orchestrates all health checks, returns structured report
- `cmd/doctor.go`: CLI command that calls `health.RunChecks()` and formats output
- `internal/vault/firstrun.go`: Detects missing vault, prompts user, delegates to existing `InitializeVault()`
- `cmd/root.go`: Intercepts command execution, calls `firstrun.Detect()` before vault-requiring commands

## Complexity Tracking

*No violations detected. All Constitution principles satisfied.*
