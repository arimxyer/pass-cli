# Atomic Save Pattern Quickstart Guide

**Feature**: Atomic Save Pattern for Vault Operations
**Audience**: Developers implementing this feature
**Phase**: Phase 1 - Design
**Date**: 2025-11-08

---

## Overview

This guide provides a step-by-step walkthrough for implementing the atomic save pattern. This feature replaces the backup-before-write pattern with atomic file operations to prevent vault corruption during crashes or power loss.

### What You're Building

**Old Pattern** (backup-then-overwrite):
```
1. Copy vault.enc → vault.enc.backup
2. Overwrite vault.enc with new data
3. If crash during step 2 → corrupted vault, restore from backup manually
```

**New Pattern** (atomic save):
```
1. Write new data → vault.enc.tmp.TIMESTAMP.RANDOM
2. Verify temp file is decryptable (in-memory)
3. os.Rename(vault.enc, vault.enc.backup) [atomic operation 1]
4. os.Rename(temp, vault.enc) [atomic operation 2]
5. If crash at any point → vault remains in consistent state automatically
```

**Key Improvement**: Vault file is NEVER partially written or corrupted. Either save fully succeeds or fully fails with vault unchanged.

---

## Prerequisites

Before implementing, read these documents:

- ✅ [spec.md](./spec.md) - Feature requirements and user stories
- ✅ [research.md](./research.md) - Design decisions and rationale
- ✅ [data-model.md](./data-model.md) - State machine and workflow
- ✅ [contracts/storage-save.md](./contracts/storage-save.md) - API contract and error types

**Estimated Time**: 2-3 days (following TDD strictly)

---

## Implementation Checklist

### Phase 1: Setup & Research (Day 1 Morning)

- [ ] Read all prerequisite documents above
- [ ] Review existing `internal/storage/storage.go` implementation
- [ ] Identify existing `SaveVault()` method and backup creation logic
- [ ] Review `internal/crypto/crypto.go` for `ClearBytes()` usage patterns
- [ ] Check `internal/vault/vault.go` for backup cleanup in `Unlock()`

### Phase 2: Test-Driven Development (Day 1 Afternoon - Day 2)

**CRITICAL**: Write tests BEFORE implementation per Constitution Principle IV

#### 2.1 Unit Tests (internal/storage/storage_test.go)

- [ ] **Test 1**: `TestAtomicSave_HappyPath`
  - Setup: Create test vault, prepare valid encrypted data
  - Execute: Call SaveVault()
  - Assert: vault.enc contains new data, vault.enc.backup contains old data, no temp files

- [ ] **Test 2**: `TestAtomicSave_VerificationFailure`
  - Setup: Create test vault, prepare INVALID encrypted data
  - Execute: Call SaveVault()
  - Assert: Returns ErrVerificationFailed, vault.enc unchanged, temp removed

- [ ] **Test 3**: `TestAtomicSave_DiskSpaceExhausted`
  - Setup: Mock filesystem with insufficient space
  - Execute: Call SaveVault()
  - Assert: Returns ErrDiskSpaceExhausted, vault.enc unchanged

- [ ] **Test 4**: `TestAtomicSave_PermissionsInherited`
  - Setup: Create vault with 0600 permissions
  - Execute: Call SaveVault()
  - Assert: Temp file created with 0600, vault.enc retains 0600 after rename

- [ ] **Test 5**: `TestAtomicSave_OrphanedFileCleanup`
  - Setup: Create fake orphaned temp files (vault.enc.tmp.old*)
  - Execute: Call SaveVault()
  - Assert: Orphaned files removed, new save completes successfully

- [ ] **Test 6**: `TestAtomicSave_Performance`
  - Benchmark: SaveVault() with 50KB vault
  - Assert: Completes in <5 seconds (SC-009)

- [ ] **Test 7**: `TestAtomicSave_RollbackPerformance`
  - Benchmark: SaveVault() with verification failure
  - Assert: Rollback completes in <1 second (SC-008)

#### 2.2 Integration Tests (test/atomic_save_test.go)

- [ ] **Test 8**: `TestAtomicSave_CrashSimulation`
  - Setup: Start save operation in subprocess
  - Execute: Kill process mid-save (`kill -9`)
  - Assert: Vault still readable after restart, no corruption

- [ ] **Test 9**: `TestAtomicSave_PowerLossSimulation`
  - Setup: Start save operation
  - Execute: Interrupt at each step (temp write, verification, rename)
  - Assert: Vault recoverable to consistent state in all cases

- [ ] **Test 10**: `TestAtomicSave_AuditLogging`
  - Setup: Enable audit logging
  - Execute: Successful save operation
  - Assert: Audit log contains all 7 events in correct order

- [ ] **Test 11**: `TestAtomicSave_SecurityNoCredentialLogging`
  - Setup: Enable verbose logging
  - Execute: Save operation with real credentials
  - Assert: Audit log NEVER contains decrypted vault content or passwords

**TDD Checkpoint**: ALL TESTS MUST FAIL before writing implementation code

---

### Phase 3: Implementation (Day 2)

Now that tests are written and failing, implement the feature.

#### 3.1 Create atomic_save.go

Create `internal/storage/atomic_save.go` with helper functions:

```go
package storage

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// generateTempFileName creates unique temp file name with timestamp + random suffix
func (s *StorageService) generateTempFileName() string {
	timestamp := time.Now().Format("20060102-150405")
	suffix := randomHexSuffix(6)
	return fmt.Sprintf("%s.tmp.%s.%s", s.vaultPath, timestamp, suffix)
}

// randomHexSuffix generates N-character hex suffix from crypto/rand
func randomHexSuffix(length int) string {
	bytes := make([]byte, length/2) // 2 hex chars per byte
	rand.Read(bytes)                 // crypto/rand, not math/rand
	return fmt.Sprintf("%x", bytes)
}

// writeToTempFile writes encrypted data to temporary file with vault permissions
func (s *StorageService) writeToTempFile(tempPath string, data []byte) error {
	// Create temp file with vault permissions (0600)
	file, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, VaultPermissions)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer file.Close()

	// Write encrypted vault data
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	// Force flush to disk before verification
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	return nil
}

// verifyTempFile decrypts and validates temporary file before commit
func (s *StorageService) verifyTempFile(tempPath string, password string) error {
	// Read temp file
	data, err := os.ReadFile(tempPath)
	if err != nil {
		return fmt.Errorf("verification failed: cannot read temporary file: %w", err)
	}

	// Decrypt using existing LoadVault logic (reuse password derivation)
	decryptedData, err := s.decryptVaultData(data, password)
	if err != nil {
		return fmt.Errorf("verification failed: encrypted data could not be decrypted: %w", err)
	}
	defer crypto.ClearBytes(decryptedData) // CRITICAL: Clear decrypted memory

	// Validate JSON structure
	if !json.Valid(decryptedData) {
		return fmt.Errorf("verification failed: vault structure is invalid")
	}

	return nil
}

// atomicRename performs atomic file rename (handles platform differences via os.Rename)
func (s *StorageService) atomicRename(oldPath, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("atomic rename failed: %w", err)
	}
	return nil
}

// cleanupTempFile removes temporary file (best-effort, logs warning if fails)
func (s *StorageService) cleanupTempFile(tempPath string) error {
	if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
		// Log warning but don't fail - cleanup is non-critical
		fmt.Fprintf(os.Stderr, "Warning: failed to remove temporary file %s: %v\n", tempPath, err)
		return err
	}
	return nil
}

// cleanupOrphanedTempFiles removes old temp files from crashed previous saves
func (s *StorageService) cleanupOrphanedTempFiles(currentTempPath string) {
	vaultDir := filepath.Dir(s.vaultPath)
	pattern := filepath.Join(vaultDir, "vault.enc.tmp.*")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return // Best-effort cleanup, ignore errors
	}

	for _, orphan := range matches {
		if orphan != currentTempPath { // Don't delete current temp file
			if err := os.Remove(orphan); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove orphaned temp file %s: %v\n", orphan, err)
			}
		}
	}
}
```

#### 3.2 Modify storage.go (SaveVault method)

Replace existing `SaveVault()` implementation in `internal/storage/storage.go`:

```go
// SaveVault atomically saves vault data using temporary file and atomic rename
func (s *StorageService) SaveVault(data []byte, password string) error {
	// FR-015: Log atomic save started
	log.Info("Atomic save started", "vault", s.vaultPath)

	// Step 1: Generate unique temp filename
	tempPath := s.generateTempFileName()

	// Step 2: Write encrypted data to temp file
	if err := s.writeToTempFile(tempPath, data); err != nil {
		return s.handleSaveError(tempPath, fmt.Errorf("save failed: %w. Your vault was not modified.", err))
	}
	log.Debug("Temp file created", "path", tempPath)

	// Step 3: Verify temp file is decryptable
	log.Info("Verification started", "temp", tempPath)
	if err := s.verifyTempFile(tempPath, password); err != nil {
		s.cleanupTempFile(tempPath)
		return s.handleSaveError(tempPath, fmt.Errorf("save failed during verification. Your vault was not modified. %w", err))
	}
	log.Info("Verification passed", "vault", s.vaultPath)

	// Step 4: Atomic rename (vault → backup)
	log.Info("Atomic rename started", "vault", s.vaultPath)
	backupPath := s.vaultPath + BackupSuffix
	if err := s.atomicRename(s.vaultPath, backupPath); err != nil {
		s.cleanupTempFile(tempPath)
		return s.handleSaveError(tempPath, fmt.Errorf("save failed: %w. Your vault was not modified.", err))
	}

	// Step 5: Atomic rename (temp → vault)
	if err := s.atomicRename(tempPath, s.vaultPath); err != nil {
		// CRITICAL ERROR: Backup exists but vault missing
		// Try to restore backup as last resort
		s.atomicRename(backupPath, s.vaultPath) // Best-effort restore
		return fmt.Errorf("CRITICAL: save failed during final rename. Attempted automatic restore from backup. Error: %w", err)
	}

	log.Info("Atomic rename completed", "vault", s.vaultPath)

	// Step 6: Cleanup orphaned temp files (best-effort)
	s.cleanupOrphanedTempFiles(tempPath)

	// FR-015: Log success
	log.Info("Atomic save completed successfully", "vault", s.vaultPath)

	return nil
}

// handleSaveError logs failure and returns user-facing error message
func (s *StorageService) handleSaveError(tempPath string, err error) error {
	log.Error("Atomic save failed", "error", err)
	// Temp file cleanup already attempted by caller
	return err
}
```

#### 3.3 Update vault.go (No changes needed)

**Verification step**: Confirm `internal/vault/vault.go:Unlock()` still removes backup file after successful unlock (lines 466-474). No changes needed - existing logic is correct.

---

### Phase 4: Testing & Validation (Day 3)

#### 4.1 Run Unit Tests

```bash
# Run all storage tests
go test ./internal/storage -v

# Run specific atomic save tests
go test ./internal/storage -v -run TestAtomicSave

# Check coverage (target: 80%)
go test -coverprofile=coverage.out ./internal/storage
go tool cover -html=coverage.out -o coverage.html
```

**Expected**: All tests pass, coverage >80%

#### 4.2 Run Integration Tests

```bash
# Run integration tests (requires -tags=integration)
go test -v -tags=integration ./test -run TestAtomicSave

# Crash simulation test
go test -v -tags=integration ./test -run TestAtomicSave_CrashSimulation
```

**Expected**: All integration tests pass, vault remains readable after crashes

#### 4.3 Manual Testing

```bash
# Build CLI
go build -o pass-cli .

# Initialize test vault
./pass-cli init

# Add credential (triggers SaveVault)
./pass-cli add testsite

# Verify vault integrity
./pass-cli get testsite

# Check file structure
ls -la ~/.config/pass-cli/
# Should see: vault.enc, vault.enc.backup (after add, before next unlock)

# Check audit log
cat ~/.config/pass-cli/audit.log | grep atomic_save

# Test crash recovery (advanced)
# 1. Start long save operation (large vault)
# 2. Kill process mid-save: kill -9 <pid>
# 3. Restart and verify vault still unlocks
./pass-cli get testsite  # Should succeed
```

#### 4.4 Performance Benchmarking

```bash
# Run performance benchmarks
go test -bench=BenchmarkSaveVault ./internal/storage

# Expected output:
# BenchmarkSaveVault-8    100    4500000 ns/op    # ~4.5s (within <5s target)
```

---

## Debugging Tips

### Check Audit Logs

```bash
# View recent atomic save events
tail -f ~/.config/pass-cli/audit.log | grep atomic_save

# Expected sequence (success):
# atomic_save_started
# temp_file_created
# verification_passed
# atomic_rename_completed
```

### Inspect Vault Directory

```bash
# List all vault files
ls -la ~/.config/pass-cli/vault.enc*

# Expected during save:
# vault.enc              # Active vault
# vault.enc.backup       # N-1 generation (after save, before unlock)
# vault.enc.tmp.xxx      # Temporary file (only during save operation)

# Expected after unlock:
# vault.enc              # Active vault only (backup removed)
```

### Verify File Permissions

```bash
# Check vault permissions (should be 0600)
ls -l ~/.config/pass-cli/vault.enc*

# Expected output:
# -rw------- 1 user user ... vault.enc         # 0600 (owner read/write only)
# -rw------- 1 user user ... vault.enc.backup  # 0600
```

### Test Crash Recovery

```bash
# Start save operation in background
./pass-cli add testsite2 &
PID=$!

# Kill process mid-save
kill -9 $PID

# Verify vault still readable
./pass-cli get testsite  # Should succeed (vault not corrupted)

# Check for orphaned temp files
ls ~/.config/pass-cli/vault.enc.tmp.*  # May exist if crash during save

# Next save operation will clean up orphaned files
./pass-cli add testsite3  # Orphaned files removed automatically
```

---

## Common Pitfalls

### ❌ Pitfall 1: Forgetting Memory Clearing

**Problem**: Decrypted vault left in memory after verification

**Solution**: Always use `defer crypto.ClearBytes(decrypted)` immediately after allocation

```go
decrypted, err := decrypt(data, password)
if err != nil {
    return err
}
defer crypto.ClearBytes(decrypted) // CRITICAL: Place immediately after allocation
```

---

### ❌ Pitfall 2: Logging Decrypted Content

**Problem**: Accidentally logging vault content or passwords

**Solution**: Log only operation outcomes, never data

```go
// ❌ WRONG
log.Debug("Decrypted vault", "data", string(decrypted))

// ✅ CORRECT
log.Info("Verification passed", "vault", s.vaultPath)
```

---

### ❌ Pitfall 3: Not Handling Step 5 Critical Failure

**Problem**: Vault → backup rename succeeds, but temp → vault fails (vault missing!)

**Solution**: Best-effort restore backup, log critical error

```go
if err := s.atomicRename(tempPath, s.vaultPath); err != nil {
    // CRITICAL: Try to restore backup
    s.atomicRename(backupPath, s.vaultPath)
    return fmt.Errorf("CRITICAL: save failed during final rename. Attempted restore from backup. Error: %w", err)
}
```

---

### ❌ Pitfall 4: Using math/rand Instead of crypto/rand

**Problem**: Predictable temp file names (security risk)

**Solution**: Always use `crypto/rand` for random suffixes

```go
// ❌ WRONG
suffix := fmt.Sprintf("%x", math.Intn(1000000))

// ✅ CORRECT
bytes := make([]byte, 3)
crypto.Read(bytes)
suffix := fmt.Sprintf("%x", bytes)
```

---

### ❌ Pitfall 5: Skipping Fsync Before Rename

**Problem**: Temp file not fully written to disk before rename

**Solution**: Always call `file.Sync()` before closing temp file

```go
file, _ := os.OpenFile(tempPath, ...)
file.Write(data)
file.Sync() // ← CRITICAL: Force disk flush
file.Close()
```

---

## Success Criteria Checklist

Before marking feature complete, verify:

- [ ] **SC-001**: Zero corruption in manual testing (add/update/delete/change-password operations)
- [ ] **SC-002**: Vault unlocks 100% after crash simulation tests
- [ ] **SC-003**: `vault.enc.backup` always contains N-1 generation (verified in tests)
- [ ] **SC-004**: Verification rejects invalid data 100% (test with corrupted vaults)
- [ ] **SC-005**: Manual recovery from backup succeeds (tested with corrupted vault.enc)
- [ ] **SC-006**: No orphaned temp files after save operations
- [ ] **SC-007**: Vault readable after simulated power loss at each step
- [ ] **SC-008**: Rollback completes in <1s (benchmarked)
- [ ] **SC-009**: Save completes in <5s for typical vault (benchmarked)

---

## Next Steps

After implementation and testing complete:

1. **Commit changes** (following CLAUDE.md commit guidelines):
   ```bash
   git add internal/storage/ test/
   git commit -m "feat: implement atomic save pattern for vault operations

   - Replace backup-before-write with temp-file-and-atomic-rename
   - Add full in-memory verification before commit
   - Implement crash-safe writes using os.Rename atomicity
   - Add orphaned temp file cleanup
   - Log all state transitions to audit.log
   - Temp files inherit vault permissions (0600)
   - Achieve <5s save, <1s rollback performance targets

   Tests:
   - Unit tests for atomic save workflow
   - Integration tests for crash simulation
   - Security tests for no credential logging
   - Performance benchmarks for SC-008 and SC-009

   Generated with Claude Code

   Co-Authored-By: Claude <noreply@anthropic.com>"
   ```

2. **Run full test suite**:
   ```bash
   go test ./...                          # All tests
   go test -race ./...                    # Race detection
   golangci-lint run                      # Linting
   gosec ./...                            # Security scan
   ```

3. **Update documentation** (if needed):
   - Update GETTING_STARTED.md if vault behavior changed
   - Update TROUBLESHOOTING.md with recovery steps
   - Update SECURITY.md if atomic save improves security posture

4. **Create pull request** (if working on fork):
   - Reference spec: `specs/003-implement-atomic-save/spec.md`
   - Include test results and benchmark outputs
   - Note: Constitution compliance verified in plan.md

---

## Questions?

If stuck or need clarification:

1. **Check spec docs first**: [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md)
2. **Review contract**: [contracts/storage-save.md](./contracts/storage-save.md)
3. **Consult constitution**: `.specify/memory/constitution.md` for project principles
4. **Ask in PR comments**: Tag reviewers for implementation questions

**Estimated Total Time**: 2-3 days (strict TDD)

Good luck! 🚀
