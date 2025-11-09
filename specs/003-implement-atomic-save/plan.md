# Implementation Plan: Atomic Save Pattern for Vault Operations

**Branch**: `003-implement-atomic-save` | **Date**: 2025-11-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-implement-atomic-save/spec.md`

## Summary

Replace current backup-before-write vault save pattern with atomic save using temporary files and rename operations. This ensures crash-safe writes, prevents vault corruption during power loss or application crashes, and verifies data integrity before committing changes. The new pattern writes to temporary files, verifies decryptability, uses atomic rename operations to swap files, and maintains single N-1 backup generation for manual recovery.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Go standard library (os, io, crypto/aes, crypto/rand, filepath), existing internal packages (crypto, storage, vault)
**Storage**: File-based encrypted vault storage (`vault.enc`), backup files (`vault.enc.backup`), temporary files (`vault.enc.tmp.TIMESTAMP.RANDOM`)
**Testing**: Go testing framework (`go test`), table-driven tests, integration tests with real file operations
**Target Platform**: Cross-platform (Windows, macOS, Linux) - single binary per platform
**Project Type**: Single CLI application with library-first architecture
**Performance Goals**: <5 second save operations (including write + verification), <1 second rollback on failure
**Constraints**: Must work on filesystems supporting atomic rename, requires 2x vault size disk space during save, owner-only file permissions (0600 Unix equivalent)
**Scale/Scope**: Modifies 2 core packages (internal/storage, internal/vault), affects all vault modification operations (add/update/delete/change-password), impacts ~500 lines across storage layer

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Security-First Development ✓

- **Encryption Standards**: No changes to AES-256-GCM or PBKDF2-SHA256 (600k iterations) - only modifying save mechanism
- **No Secret Logging**: FR-015 requires logging state transitions only, not vault contents - compliant
- **Secure Memory Handling**: Verification step will decrypt in-memory but must clear after verification - requires attention
- **Threat Modeling**: Addresses threat of vault corruption from crashes/power loss - reduces attack surface for data loss
- **Temporary File Security**: FR-016 ensures temp files inherit vault permissions (0600) - compliant

**Status**: PASS - Feature improves security posture by preventing corruption vectors

### Principle II: Library-First Architecture ✓

- **Changes isolated to internal/storage package**: Save logic remains in library layer
- **No CLI dependencies**: Storage layer operations remain pure library functions
- **Clear interfaces**: Existing SaveVault API remains, implementation changes internal
- **Independently testable**: Unit tests for atomic save logic separate from CLI integration tests

**Status**: PASS - Maintains clean library separation

### Principle III: CLI Interface Standards ✓

- **No CLI changes required**: Feature is transparent to CLI commands
- **Error output to stderr**: FR-011 specifies clear error messages with actionable guidance
- **Exit codes preserved**: Existing error handling and exit codes remain unchanged

**Status**: PASS - No CLI interface impact

### Principle IV: Test-Driven Development (NON-NEGOTIABLE) ⚠️

- **Tests required**: Must write tests BEFORE implementation per TDD mandate
- **Coverage types needed**:
  - Unit tests: Atomic save operations, verification logic, cleanup operations
  - Integration tests: Crash simulation (kill process mid-save), power-loss scenarios, orphaned file cleanup
  - Security tests: Temp file permissions, memory clearing after verification, backup integrity
- **Coverage target**: 80% minimum for modified storage package code

**Status**: REQUIRES VIGILANCE - Must follow Red-Green-Refactor strictly, tests must fail before implementation

### Principle V: Cross-Platform Compatibility ✓

- **os.Rename atomic on all platforms**: Windows, macOS, Linux support atomic file rename
- **Platform-specific considerations**: File permissions (0600 vs ACLs on Windows) - already handled in existing code
- **Path handling**: Using filepath.Join for all path operations
- **CI matrix**: Existing CI already tests Windows/macOS/Linux

**Status**: PASS - No new platform-specific code required

### Principle VI: Observability & Auditability ✓

- **FR-015**: Requires logging all state transitions (temp file creation, verification pass/fail, rename, cleanup)
- **No credential logging**: Verification logs success/failure only, never decrypted content
- **Audit trail**: Atomic save events logged to existing audit.log infrastructure
- **Existing vault operations already logged**: Save success/failure will be enhanced with atomic save details

**Status**: PASS - Enhances observability without compromising security

### Principle VII: Simplicity & YAGNI ✓

- **Concrete user need**: Prevents data loss from crashes - critical for password manager
- **Standard library only**: Uses os, io, crypto packages - no new dependencies
- **No speculative features**: Explicitly excludes multi-process locking, network filesystems, multiple backup generations (see spec Out of Scope)
- **Direct solution**: Atomic rename is industry-standard pattern, not "clever" optimization

**Status**: PASS - Addresses real user need with simple, proven pattern

### Overall Gate Decision: **PASS WITH CONDITIONS**

**Conditions**:
1. Must write tests FIRST (Principle IV violation will reject implementation)
2. Must verify memory clearing after verification step (Principle I requirement)
3. Must log state transitions without credential leakage (Principle VI requirement)

**Re-check required after Phase 1 design**: Verify implementation plan maintains test-first approach and security guarantees

## Project Structure

### Documentation (this feature)

```
specs/003-implement-atomic-save/
├── plan.md              # This file
├── research.md          # Phase 0: Design pattern research, filesystem atomicity guarantees
├── data-model.md        # Phase 1: File state machine, atomic save workflow
├── quickstart.md        # Phase 1: Developer guide for atomic save pattern
├── contracts/           # Phase 1: Storage layer interface contracts
│   └── storage-save.md  # SaveVault method contract with atomic semantics
├── checklists/
│   └── requirements.md  # Spec quality checklist (already complete)
└── tasks.md             # Phase 2: NOT created by this command
```

### Source Code (repository root)

```
internal/
├── storage/
│   ├── storage.go           # MODIFIED: Replace SaveVault with atomic save pattern
│   ├── storage_test.go      # MODIFIED: Add atomic save unit tests
│   └── atomic_save.go       # NEW: Atomic save implementation (temp file + verify + rename)
├── vault/
│   └── vault.go             # MODIFIED: Update backup cleanup logic (remove old backup after unlock)
└── crypto/
    └── crypto.go            # REVIEWED: Ensure ClearBytes used for verification memory

test/
├── integration_test.go      # MODIFIED: Add crash simulation tests
└── atomic_save_test.go      # NEW: Integration tests for atomic save (power loss, crash scenarios)

cmd/
├── add.go                   # NO CHANGES: Uses storage.SaveVault (transparent)
├── update.go                # NO CHANGES: Uses storage.SaveVault (transparent)
├── delete.go                # NO CHANGES: Uses storage.SaveVault (transparent)
└── change_password.go       # NO CHANGES: Uses storage.SaveVault (transparent)
```

**Structure Decision**: Single project structure (Option 1). Changes concentrated in `internal/storage` package with ripple verification in `internal/vault` for backup cleanup. CLI layer (`cmd/`) requires no changes due to library-first architecture - atomic save is transparent to command implementations.

## Complexity Tracking

*No violations to justify - all Constitution gates passed*

---

## Phase 0: Outline & Research

### Research Tasks

1. **Filesystem Atomicity Guarantees**
   - Research: How atomic is `os.Rename()` on Windows/macOS/Linux?
   - Research: Edge cases where rename might not be atomic (network filesystems, containerized environments)
   - Research: How to detect filesystem doesn't support atomic rename?
   - Output: Document guarantees and failure modes in research.md

2. **Atomic Save Pattern Best Practices**
   - Research: Industry-standard atomic file save patterns (databases, config files, password managers)
   - Research: Bitwarden, 1Password, KeePassXC implementations (if open source available)
   - Research: Go stdlib patterns for atomic file writes
   - Output: Reference implementations and pattern variations in research.md

3. **Verification Strategy**
   - Research: In-memory decryption overhead for verification (performance impact)
   - Research: Memory clearing best practices in Go (crypto.ClearBytes effectiveness)
   - Research: Verification failure scenarios (corrupted encryption, wrong password, malformed JSON)
   - Output: Verification approach decision with performance tradeoffs in research.md

4. **Temporary File Cleanup Strategies**
   - Research: When to clean orphaned temp files (startup vs before-save vs background)
   - Research: Safe temp file deletion patterns (avoid race conditions)
   - Research: Handling temp files from crashed processes with same PID reused
   - Output: Cleanup timing decision with rationale in research.md

5. **Random Suffix Generation**
   - Research: crypto/rand vs math/rand for file suffix uniqueness
   - Research: Suffix length vs collision probability
   - Research: Timestamp format (RFC3339 vs custom) for debuggability
   - Output: Naming convention decision in research.md

### Research Deliverable: research.md

**Format**:
```markdown
# Atomic Save Pattern Research

## Decision: Filesystem Atomicity
- **Chosen**: Use os.Rename() with pre-flight checks
- **Rationale**: [findings from research]
- **Alternatives**: [temp + copy + delete vs atomic rename]

## Decision: Verification Approach
- **Chosen**: [in-memory decrypt entire vault vs partial verification]
- **Rationale**: [performance vs completeness tradeoff]
- **Alternatives**: [checksum-only vs full decrypt]

[... continue for each research task]
```

---

## Phase 1: Design & Contracts

**Prerequisites**: research.md complete, all Technical Context NEEDS CLARIFICATION resolved

### 1. Data Model (data-model.md)

**Entities**:

1. **Vault File State Machine**
   - States: `Consistent`, `Saving`, `Corrupt`, `Missing`
   - Transitions:
     - `Consistent` → `Saving` (SaveVault called)
     - `Saving` → `Consistent` (atomic rename succeeds)
     - `Saving` → `Consistent` (rollback on verification failure)
     - `Consistent` → `Corrupt` (manual filesystem damage)
   - Invariant: Never transition to `Corrupt` via save operation

2. **Save Operation Workflow**
   - Inputs: vault data ([]byte), master password (string), vault path (string)
   - Outputs: success (nil) or error with user-facing message
   - Steps:
     1. Create temp file with timestamp+random suffix
     2. Write encrypted data to temp file
     3. Fsync temp file
     4. Verify temp file (decrypt in-memory)
     5. os.Rename(vaultPath, backupPath) [atomic operation 1]
     6. os.Rename(tempPath, vaultPath) [atomic operation 2]
     7. Log success, clean temp file
   - Failure Handling: At any step, remove temp file and return error (vault unchanged)

3. **File Naming Convention**
   - Primary vault: `vault.enc`
   - Backup: `vault.enc.backup` (N-1 generation)
   - Temporary: `vault.enc.tmp.20251108-143022.a3f8c2` (timestamp + 6-char random hex)
   - Orphaned temps: Any `vault.enc.tmp.*` older than current save operation

4. **Verification Data Structure**
   - Input: Temporary file path, master password
   - Process: Read file → Decrypt in-memory → Validate JSON structure → Clear decrypted memory
   - Output: Verification result (pass/fail) + specific error if failed
   - Memory constraint: Decrypted vault must fit in memory (assume <100MB per spec assumption)

### 2. API Contracts (contracts/storage-save.md)

```markdown
# Storage Layer Save Contract

## SaveVault(data []byte, password string) error

### Preconditions
- Vault file exists (initialized)
- `data` contains valid encrypted vault structure
- `password` is correct master password
- Sufficient disk space (2x vault size minimum)
- Write permissions on vault directory

### Postconditions (Success)
- Primary vault file (`vault.enc`) contains new data
- Backup file (`vault.enc.backup`) contains N-1 generation (previous vault state)
- No temporary files remain in vault directory
- Vault file permissions match original (0600 equivalent)
- State transitions logged to audit.log
- Operation completes in <5 seconds

### Postconditions (Failure)
- Primary vault file unchanged (atomic guarantee)
- Backup file unchanged
- Temporary file removed
- Error message includes: failure reason, vault status ("not modified"), actionable guidance
- Failure logged to audit.log
- Operation fails fast (<1 second rollback)

### Error Types
- `ErrVerificationFailed`: Temp file failed decryption verification
- `ErrDiskSpaceExhausted`: Insufficient disk space for temp file
- `ErrPermissionDenied`: Cannot write to vault directory
- `ErrFilesystemNotAtomic`: Rename operation not supported on filesystem

### Observable Side Effects
- Audit log entry: "Atomic save started"
- Audit log entry: "Temp file created: vault.enc.tmp.TIMESTAMP.RANDOM"
- Audit log entry: "Verification: passed" or "Verification: failed - [reason]"
- Audit log entry: "Atomic rename completed" or "Rollback completed"
- Audit log entry: "Cleanup: temp file removed"

### Thread Safety
- NOT thread-safe: Assumes single-process vault access (per constitution assumption)
- Concurrent calls will create conflicting temp files (random suffix prevents collision but last writer wins)

### Platform Behavior
- Windows: Uses MoveFileEx with MOVEFILE_REPLACE_EXISTING
- macOS/Linux: Uses rename(2) syscall (atomic per POSIX)
- All platforms: Verifies rename atomicity in pre-flight check

### Performance Guarantees
- Normal case: <5 seconds (write + verify + rename)
- Rollback case: <1 second (remove temp file)
- Disk I/O: 2x vault size (write temp + rename to primary)
- Memory: 1x vault size (in-memory verification)
```

### 3. Quickstart Guide (quickstart.md)

```markdown
# Atomic Save Pattern Developer Guide

## Overview

This feature replaces the backup-before-write pattern with atomic file operations to prevent vault corruption during crashes or power loss.

## Key Concepts

**Old Pattern** (backup-then-overwrite):
1. Copy `vault.enc` → `vault.enc.backup`
2. Overwrite `vault.enc` with new data
3. If crash during step 2 → corrupted vault, restore from backup

**New Pattern** (atomic save):
1. Write new data → `vault.enc.tmp.TIMESTAMP.RANDOM`
2. Verify temp file is decryptable
3. `os.Rename(vault.enc, vault.enc.backup)` [atomic]
4. `os.Rename(temp, vault.enc)` [atomic]
5. If crash at any point → vault remains in consistent state

## Implementation Checklist

### Before Writing Code
- [ ] Read research.md for design decisions
- [ ] Review data-model.md for state machine
- [ ] Read contracts/storage-save.md for API contract

### TDD Workflow (MANDATORY)
1. **Red**: Write failing test for atomic save happy path
2. **Green**: Implement minimum code to pass test
3. **Refactor**: Clean up implementation
4. **Repeat**: For each edge case (verification failure, disk full, permissions, etc.)

### Test Cases Required
- [ ] Happy path: Save succeeds, vault updated, backup created
- [ ] Verification failure: Save aborted, vault unchanged
- [ ] Disk space exhausted: Save aborted, temp file cleaned up
- [ ] Crash simulation: Kill process mid-save, vault remains readable
- [ ] Power loss simulation: Interrupt at each step, verify recovery
- [ ] Orphaned temp files: Verify cleanup on next operation
- [ ] Permissions: Temp file inherits vault permissions
- [ ] Performance: Verify <5 second save time
- [ ] Memory: Verify decrypted vault cleared after verification

### Implementation Steps
1. Create `internal/storage/atomic_save.go` with helper functions:
   - `generateTempFileName() string`
   - `writeToTempFile(path string, data []byte) error`
   - `verifyTempFile(path string, password string) error`
   - `atomicRename(oldPath, newPath string) error`
   - `cleanupTempFile(path string) error`

2. Modify `internal/storage/storage.go`:
   - Replace `SaveVault()` implementation with atomic save workflow
   - Keep backup removal in `internal/vault/vault.go:Unlock()` (no change)

3. Update `internal/storage/storage_test.go`:
   - Add table-driven tests for atomic save scenarios
   - Add benchmark for save performance (<5s target)

4. Add `test/atomic_save_test.go`:
   - Integration tests with real file operations
   - Crash simulation tests (forceful process termination)

### Logging Integration
```go
// Example logging pattern
log.Info("Atomic save started", "vault", s.vaultPath)
log.Debug("Temp file created", "path", tempPath)
log.Info("Verification passed", "vault", s.vaultPath)
log.Info("Atomic rename completed", "vault", s.vaultPath)
log.Debug("Cleanup completed", "temp", tempPath)
```

### Error Handling Pattern
```go
if err := verifyTempFile(tempPath, password); err != nil {
    cleanupTempFile(tempPath) // Best effort cleanup
    return fmt.Errorf("save failed during verification. Your vault was not modified. %w", err)
}
```

## Testing Locally

```bash
# Run unit tests
go test ./internal/storage -v -run TestAtomicSave

# Run integration tests
go test -v -tags=integration ./test -run TestAtomicSave

# Benchmark save performance
go test -bench=BenchmarkSaveVault ./internal/storage

# Check coverage
go test -coverprofile=coverage.out ./internal/storage
go tool cover -html=coverage.out
```

## Debugging Tips

- Check audit.log for state transition logs
- Inspect vault directory for orphaned temp files: `ls -la ~/.config/pass-cli/vault.enc.tmp.*`
- Verify file permissions: `ls -l ~/.config/pass-cli/vault.enc*` (should be 0600)
- Test crash recovery: `kill -9 <pid>` during save, then verify vault still unlocks

## Common Pitfalls

❌ **Don't**: Call SaveVault concurrently from multiple goroutines
✅ **Do**: Rely on single-process assumption per constitution

❌ **Don't**: Log decrypted vault content during verification
✅ **Do**: Log only verification pass/fail status

❌ **Don't**: Skip memory clearing after verification
✅ **Do**: Always defer crypto.ClearBytes on decrypted data

❌ **Don't**: Assume rename is instant (race conditions possible)
✅ **Do**: Use fsync before rename, handle rename errors gracefully
```

### 4. Agent Context Update

Run agent context update script:
```bash
.specify/scripts/powershell/update-agent-context.ps1 -AgentType claude
```

This will update `.claude/context.md` with atomic save pattern context (if not already present).

---

## Phase 1 Complete - Stop Here

**Next command**: `/speckit.tasks` (generates task breakdown from this plan)

**Deliverables Created**:
- ✅ plan.md (this file)
- ⏳ research.md (Phase 0 - to be generated)
- ⏳ data-model.md (Phase 1 - to be generated)
- ⏳ contracts/storage-save.md (Phase 1 - to be generated)
- ⏳ quickstart.md (Phase 1 - to be generated)

**Ready for**: Task generation and implementation phase
