# Implementation Plan: Keychain Lifecycle Management

**Branch**: `011-keychain-lifecycle-management` | **Date**: 2025-10-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/011-keychain-lifecycle-management/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Complete the keychain lifecycle for pass-cli by adding three critical commands: `pass-cli keychain enable` (enable keychain for existing vaults without recreation), `pass-cli keychain status` (inspect keychain state for troubleshooting), and `pass-cli vault remove` (clean vault removal with automatic keychain cleanup). Technical approach leverages existing internal/keychain package Delete/Clear methods (lines 94-105) and vault unlock/validation logic, adds new Cobra commands under `cmd/`, and integrates with existing audit log system for security tracking per FR-015.

## Technical Context

**Language/Version**: Go 1.21+ (existing codebase)
**Primary Dependencies**:
- github.com/spf13/cobra (CLI framework - existing)
- github.com/99designs/keyring (OS keychain integration - existing)
- Existing internal packages: internal/keychain, internal/vault, internal/storage

**Storage**: Encrypted vault files (vault.enc) with JSON structure, OS-native keychain for master passwords (per-vault entries using service name format "pass-cli:/absolute/path/to/vault.enc" per spec.md FR-003)
**Testing**: Go standard testing (`go test`), integration tests with real vault files and keychain operations
**Target Platform**: Cross-platform (Windows, macOS, Linux) with platform-specific keychain backends
**Project Type**: Single CLI application with library-first architecture (per Constitution Principle II)
**Performance Goals**:
- Enable command: <1 minute (SC-001)
- Status command: <30 seconds diagnostic output (SC-002)
- Remove command: 95% success rate for both file + keychain deletion (SC-003)

**Constraints**:
- Must respect existing service name format "pass-cli:/absolute/path/to/vault.enc" (FR-003)
- Must work across Windows/macOS/Linux with 100% platform coverage (SC-004, FR-009)
- Must integrate with existing audit log system (FR-015)
- Must not break existing keychain functionality (create during init, update during password change)
- Must handle keychain-unavailable gracefully (FR-007, SC-005)

**Scale/Scope**:
- 3 new commands (enable, status, remove)
- 15 functional requirements
- 14 acceptance scenarios across 3 user stories
- 7 edge cases to handle

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Principle I: Security-First Development (NON-NEGOTIABLE)

- **Pass**: All keychain operations use existing OS-provided secure storage (Windows Credential Manager, macOS Keychain, Linux Secret Service) per FR-009
- **Pass**: No new credential logging introduced; FR-015 explicitly logs operation types, timestamps, vault paths but NOT passwords
- **Pass**: Password validation before keychain storage (FR-002) prevents storing incorrect passwords
- **Pass**: Remove command cleanly deletes keychain entries, preventing orphaned secrets (FR-005, FR-012)
- **Action Required**: Phase 0 research must verify memory zeroing practices for password prompts in enable command

### ✅ Principle II: Library-First Architecture

- **Pass**: Commands delegate to internal/keychain package methods (Delete/Clear at lines 94-105, existing Get/Set)
- **Pass**: New functionality isolated in existing internal/vault and internal/keychain packages
- **Pass**: CLI commands (cmd/) are thin wrappers calling library functions
- **No Violations**: No new CLI-specific logic in libraries

### ✅ Principle III: CLI Interface Standards

- **Pass**: All commands output to stdout (status display) and stderr (errors) per FR-007
- **Pass**: Exit codes follow standard (0=success, non-zero=error per SC-005)
- **Pass**: Remove command supports `--yes`/`--force` flag for automation (edge case per spec.md lines 91-92)
- **Pass**: All commands respect `--vault` flag for non-default locations (FR-013)
- **Pass**: Status command includes actionable suggestions (FR-014)

### ✅ Principle IV: Test-Driven Development (NON-NEGOTIABLE)

- **Action Required**: Phase 2 tasks must include test-first development for all 3 commands
- **Required Test Types**:
  - Unit tests: keychain enable logic, status detection, remove cleanup
  - Integration tests: Real vault files + real keychain operations on all platforms
  - Contract tests: Existing keychain API stability (ensure no breaking changes)
  - Security tests: Verify FR-015 audit logging, password validation, error handling

### ✅ Principle V: Cross-Platform Compatibility

- **Pass**: Leverages existing cross-platform keychain package (github.com/99designs/keyring)
- **Pass**: FR-009 explicitly requires Windows/macOS/Linux support
- **Pass**: SC-004 requires 100% platform coverage
- **Action Required**: Phase 1 must document platform-specific error messages for FR-007 (e.g., "Access denied to Windows Credential Manager")

### ✅ Principle VI: Observability & Auditability

- **Pass**: FR-015 requires audit logging for all keychain lifecycle operations (enable, status, remove)
- **Pass**: Logs include timestamps, operation type, vault path, outcome (but NOT passwords)
- **Pass**: Status command provides visibility into keychain state (FR-004)
- **No Violations**: Follows existing audit log patterns

### ✅ Principle VII: Simplicity & YAGNI

- **Pass**: Minimal scope (exactly 3 commands to complete lifecycle, no extras)
- **Pass**: Reuses existing internal/keychain package (no new abstractions)
- **Pass**: Direct implementation using established Cobra command pattern
- **Pass**: No speculative features beyond spec requirements

### Security Review Checklist (Pre-Implementation)

- [ ] No secrets logged or printed (FR-015 specifies operation types only, not passwords)
- [ ] Memory containing passwords zeroed after enable command (action required for Phase 0 research)
- [ ] Error messages do not leak vault passwords (FR-007 requires "clear error messages")
- [ ] All operations use approved keychain backends (FR-009: Windows Credential Manager, macOS Keychain, Linux Secret Service)
- [ ] Tests verify audit logging works correctly (FR-015 acceptance scenario)

**GATE STATUS**: ✅ **PASS** - No constitution violations. Action items for Phase 0 research documented.

## Constitution Check (Post-Design Re-Evaluation)

*Performed after Phase 1 design completion*

### ✅ Principle I: Security-First Development

- **RESOLVED**: Memory zeroing practices verified in research.md (crypto.ClearBytes with defer pattern)
- **RESOLVED**: Platform-specific error messages designed in contracts/commands.md (no password leakage)
- **Pass**: All security requirements met per design artifacts

### ✅ Principle II: Library-First Architecture

- **Pass**: data-model.md confirms no new packages needed (reuses internal/keychain, internal/vault)
- **Pass**: quickstart.md shows CLI commands delegate to library methods
- **Pass**: Command implementations are thin wrappers per design

### ✅ Principle III: CLI Interface Standards

- **Pass**: contracts/commands.md defines stdout/stderr separation, exit codes, script-friendly flags
- **Pass**: All output formats documented and consistent

### ✅ Principle IV: Test-Driven Development

- **Pass**: quickstart.md outlines TDD workflow (write tests → approve → implement → refactor)
- **Pass**: Test contracts defined in contracts/commands.md (unit + integration + security tests)

### ✅ Principle V: Cross-Platform Compatibility

- **RESOLVED**: Platform-specific error messages defined in research.md Decision 5
- **Pass**: contracts/commands.md documents Windows/macOS/Linux outputs

### ✅ Principle VI: Observability & Auditability

- **Pass**: Audit logging integration documented in research.md Decision 2
- **Pass**: New event types defined in data-model.md

### ✅ Principle VII: Simplicity & YAGNI

- **Pass**: Design confirms minimal scope (3 commands, no new packages, reuses existing infrastructure)

**FINAL GATE STATUS**: ✅ **PASS** - All action items resolved. Ready for Phase 2 (task generation).

## Project Structure

### Documentation (this feature)

```
specs/011-keychain-lifecycle-management/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (memory zeroing, audit log integration patterns)
├── data-model.md        # Phase 1 output (keychain operation entities, state transitions)
├── quickstart.md        # Phase 1 output (developer guide for implementing commands)
├── contracts/           # Phase 1 output (command signatures, error responses)
│   └── commands.md      # CLI command signatures for enable, status, remove
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
# Existing structure - Commands added to existing locations
cmd/
├── keychain.go          # NEW: Parent command for keychain subcommands
├── keychain_enable.go   # NEW: Implements enable subcommand
├── keychain_status.go   # NEW: Implements status subcommand
├── vault_remove.go      # NEW: Implements vault remove command
├── root.go              # MODIFIED: Registers keychain parent command
└── helpers.go           # MODIFIED: May add shared keychain helper functions

internal/
├── keychain/
│   └── keychain.go      # EXISTING: Delete/Clear methods at lines 94-105 (already implemented)
├── vault/
│   └── vault.go         # EXISTING: Unlock/validation logic (reused for enable command)
└── storage/
    └── storage.go       # EXISTING: File operations (reused for remove command)

test/
├── integration/
│   ├── keychain_enable_test.go    # NEW: Integration tests for enable command
│   ├── keychain_status_test.go    # NEW: Integration tests for status command
│   └── vault_remove_test.go       # NEW: Integration tests for remove command
└── unit/
    └── keychain_lifecycle_test.go # NEW: Unit tests for keychain operations
```

**Structure Decision**: Single CLI project following existing pass-cli architecture. New commands added to `cmd/` directory following established Cobra pattern (see cmd/init.go:44, cmd/change_password.go as references). Functionality delegated to existing `internal/keychain` and `internal/vault` packages per Library-First Architecture (Constitution Principle II). No new packages required—feature completes existing keychain lifecycle using already-implemented infrastructure.

## Complexity Tracking

*No constitution violations requiring justification.*

All complexity is justified by functional requirements and follows existing patterns:
- 3 new commands match existing Cobra command structure (cmd/init.go, cmd/add.go, etc.)
- Audit logging integration follows existing audit.log patterns (FR-015)
- Platform-specific error messages handled by existing keychain package abstractions
