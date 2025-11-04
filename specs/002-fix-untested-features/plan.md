# Implementation Plan: Fix Untested Features and Complete Test Coverage

**Branch**: `002-fix-untested-features` | **Date**: 2025-11-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-fix-untested-features/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Fix broken keychain enable/status/vault remove commands that are currently non-functional despite appearing in CLI help and having test skeletons. The root cause is 25 tests marked with `t.Skip("TODO: Implement")` hiding incomplete implementations from quality control. This spec implements:

1. **Vault metadata system** (.meta.json files) to track keychain/audit enablement with graceful degradation for legacy vaults
2. **Keychain enable** command implementation with idempotent behavior and --force flag
3. **Keychain status** command improvements for accurate reporting with metadata consistency checks
4. **Vault remove** command implementation with 100% reliability, confirmation prompts, and complete cleanup (vault file + metadata + keychain entry)
5. **Test coverage completion** - unskip all 25 TODO-marked tests and ensure 100% pass rate

Technical approach: Extend existing vault service with metadata operations, implement missing command logic in cmd/ layer following library-first architecture, update TUI to check metadata before attempting keychain unlock.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
- `github.com/spf13/cobra` (CLI framework)
- `github.com/spf13/viper` (configuration)
- `github.com/zalando/go-keyring` (OS keychain integration)
- `github.com/rivo/tview` (TUI framework)
- Go standard library (`encoding/json`, `os`, `path/filepath`, `crypto/aes`)

**Storage**: File-based
- Encrypted vault files (vault.enc) with AES-256-GCM
- JSON metadata files (vault.enc.meta.json) - **NEW in this spec**
- JSON audit logs (audit.log)
- OS system keychain (Windows Credential Manager, macOS Keychain, Linux Secret Service)

**Testing**: Go testing framework with integration test tags
- Unit tests: `go test ./...`
- Integration tests: `go test -tags=integration ./test`
- Current issue: 25 tests with `t.Skip("TODO: Implement")` masking broken functionality

**Target Platform**: Cross-platform CLI - Windows, macOS (Intel + ARM), Linux (amd64 + arm64)

**Project Type**: Single Go project with library-first architecture (internal/ packages + cmd/ layer)

**Performance Goals**:
- CLI commands: <100ms for cached/keychain operations
- Vault operations: Minimal - dominated by PBKDF2 key derivation (intentionally slow for security)
- 100% success rate for destructive operations (vault remove)

**Constraints**:
- Offline-first (no network operations)
- Security-first: No credential logging, memory clearing via `crypto.ClearBytes()`
- Backward compatibility: Legacy vaults without .meta.json must continue working
- Single-process assumption: No concurrent access detection (documented limitation)

**Scale/Scope**:
- Single vault per user (single-vault architecture)
- 3 commands to fix: `keychain enable`, `keychain status`, `vault remove`
- 25 skipped tests to unskip and complete
- ~10 files to modify (cmd/, internal/vault/, test/)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Security-First Development (NON-NEGOTIABLE)
**Status**: ✅ PASS

- Vault metadata (.meta.json) contains only boolean flags and timestamps - no secrets
- Keychain integration uses existing secure `internal/keychain` abstraction
- All commands follow existing password clearing patterns (`defer crypto.ClearBytes()`)
- Tests will verify FR-008/FR-013/FR-017 (audit logging without credential exposure)
- No new cryptographic operations introduced (reuses existing AES-256-GCM vault operations)

### II. Library-First Architecture
**Status**: ✅ PASS

- Changes extend existing `internal/vault` library with metadata operations
- CMD layer (`cmd/keychain_enable.go`, `cmd/vault_remove.go`) delegates to library functions
- No business logic in CLI commands - follows existing pattern seen in cmd/add.go, cmd/get.go
- Metadata operations isolated in `internal/vault/metadata.go` (new file)

### III. CLI Interface Standards
**Status**: ✅ PASS

- All commands follow existing Cobra patterns (flags, stdin/stdout/stderr)
- Exit codes: 0 for success/idempotent, 1 for user errors, 2 for system errors
- `--yes` flag for `vault remove` follows existing confirmation pattern (see cmd/delete.go)
- No breaking changes to existing command interfaces

### IV. Test-Driven Development (NON-NEGOTIABLE)
**Status**: ✅ PASS

- **This spec exists because TDD was NOT followed** - tests were marked TODO instead of driving implementation
- Fixes the root cause: Unskips all 25 TODO-marked tests
- Implementation approach: Unskip tests → verify they fail → implement → tests pass
- Success criteria SC-002/SC-006: 100% test pass rate, 0 TODO skips
- CI gate (SC-005): Block merges containing `t.Skip("TODO:`

### V. Cross-Platform Compatibility
**Status**: ✅ PASS

- Uses existing `github.com/zalando/go-keyring` abstraction (already cross-platform)
- Metadata files use platform-agnostic `encoding/json` and `path/filepath`
- No new platform-specific code introduced
- CI already tests Windows, macOS (Intel + ARM), Linux (amd64 + arm64)

### VI. Observability & Auditability
**Status**: ✅ PASS

- FR-008: `keychain enable` writes audit entries when vault has audit enabled
- FR-013: `keychain status` writes audit entries
- FR-017/FR-018: `vault remove` writes audit entries BEFORE deleting metadata
- Audit entries contain only operation types, timestamps - never credentials (follows constitution)

### VII. Simplicity & YAGNI
**Status**: ✅ PASS

- Implements only what's already advertised in CLI help (no new features)
- Deferred complex features (concurrent access detection) to future specs (documented in design-improvements-post-spec.md)
- Reuses existing patterns: metadata file format mirrors audit.log (JSON lines)
- Minimal file changes: ~10 files (cmd/, internal/vault/, test/)

**GATE RESULT**: ✅ ALL PRINCIPLES SATISFIED - Proceed to Phase 0

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
cmd/                              # CLI command implementations (Cobra)
├── keychain_enable.go            # ← MODIFIED: Implement metadata update + idempotent logic
├── keychain_status.go            # ← MODIFIED: Add metadata consistency checks
├── vault_remove.go               # ← MODIFIED: Implement full removal logic
└── tui/
    └── main.go                   # ← MODIFIED: Check metadata before keychain unlock

internal/                         # Library packages (business logic)
├── vault/
│   ├── vault.go                  # ← MODIFIED: Update UnlockWithKeychain() to check metadata
│   ├── metadata.go               # ← NEW: Metadata CRUD operations
│   └── metadata_test.go          # ← NEW: Unit tests for metadata operations
├── keychain/
│   └── keychain.go               # (existing, no changes)
└── security/
    └── audit.go                  # (existing, used for FR-008/FR-013/FR-017)

test/                             # Integration tests
├── keychain_enable_test.go       # ← MODIFIED: Unskip 3 TODO tests (lines 68, 105, 134)
├── keychain_status_test.go       # ← MODIFIED: Unskip 3 TODO tests (lines 64, 95, 117)
└── vault_remove_test.go          # ← MODIFIED: Unskip 5 TODO tests (lines 70, 100, 136, 178, +1 more)

specs/002-fix-untested-features/  # This spec's documentation
├── spec.md
├── plan.md                       # ← This file
├── research.md                   # ← Phase 0 output (next step)
├── data-model.md                 # ← Phase 1 output
└── checklists/
    └── requirements.md
```

**Structure Decision**: Single Go project following existing pass-cli architecture. All changes follow library-first principle:
- Business logic in `internal/vault/metadata.go` (new)
- CMD layer delegates to library functions
- Integration tests in `test/` (existing structure)
- No new packages or directories required beyond `internal/vault/metadata.go`

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

**No violations** - All constitution principles satisfied. This spec actually **reduces** complexity by:
- Fixing broken advertised features (less user confusion)
- Completing test coverage (reduces technical debt)
- Removing 25 TODO markers (cleaner codebase)
- Using existing patterns (no new abstractions)
