# Storage Layer Save Contract

**Feature**: Atomic Save Pattern for Vault Operations
**Contract Type**: API Method Contract
**Phase**: Phase 1 - Design
**Date**: 2025-11-08

---

## Method Signature

```go
func (s *StorageService) SaveVault(data []byte, password string) error
```

### Purpose

Atomically save encrypted vault data to disk with crash-safe guarantees. This method replaces the existing backup-before-write pattern with a temporary-file-and-atomic-rename pattern to prevent vault corruption during power loss or application crashes.

---

## Preconditions

The following conditions MUST be true before calling `SaveVault()`:

| Precondition | Validation | Violation Behavior |
|--------------|------------|-------------------|
| Vault file exists (initialized) | `s.VaultExists()` returns `true` | Return error: "vault not initialized" |
| `data` contains valid encrypted vault structure | Non-empty `[]byte`, proper EncryptedVault JSON | Detected during verification step |
| `password` is correct master password | Verification decrypt succeeds | Return `ErrVerificationFailed` |
| Sufficient disk space (2x vault size minimum) | File write succeeds | Return `ErrDiskSpaceExhausted` |
| Write permissions on vault directory | `os.OpenFile()` succeeds | Return `ErrPermissionDenied` |

**Caller Responsibility**: Vault service must ensure vault is unlocked and data is properly encrypted before calling `SaveVault()`.

---

## Postconditions (Success Path)

When `SaveVault()` returns `nil`, the following conditions are GUARANTEED:

| Postcondition | Observable Evidence | Audit Log Entry |
|---------------|-------------------|-----------------|
| Primary vault file contains new data | `vault.enc` modified timestamp updated, content matches `data` parameter | `atomic_rename_completed` |
| Backup file contains N-1 generation | `vault.enc.backup` exists, content is previous vault state | `atomic_rename_started` |
| No temporary files remain | No files matching `vault.enc.tmp.*` in vault directory | `cleanup_orphaned_files` |
| Vault file permissions preserved | `vault.enc` has same permissions as before save (0600 equivalent) | N/A |
| State transitions logged | audit.log contains event sequence | All events logged |
| Operation completes in <5 seconds | Save duration from start to finish | `atomic_rename_completed` includes `duration_ms` |

**Atomicity Guarantee**: The vault file is NEVER in a partially written or corrupted state. Either the save fully succeeds (new data committed) or fully fails (old data preserved).

---

## Postconditions (Failure Path)

When `SaveVault()` returns non-nil error, the following conditions are GUARANTEED:

| Postcondition | Observable Evidence | Audit Log Entry |
|---------------|---------------------|-----------------|
| Primary vault file unchanged | `vault.enc` modified timestamp unchanged, content matches pre-save state | `rollback_completed` |
| Backup file unchanged | `vault.enc.backup` unmodified (if existed before save) | N/A |
| Temporary file removed (best-effort) | No `vault.enc.tmp.TIMESTAMP.RANDOM` for current save operation | `rollback_completed` |
| Error message includes failure reason, vault status, and actionable guidance | Error string matches FR-011 format | Event with `outcome: failed` |
| Failure logged to audit.log | audit.log contains failure event with reason | Event matching error type |
| Operation fails fast (<1 second rollback) | Rollback duration logged | `rollback_completed` includes `duration_ms` |

**Rollback Guarantee**: If any step fails, the vault returns to its pre-save state. No partial updates are ever committed.

---

## Error Types

### ErrVerificationFailed

**When**: Temporary file fails decryption verification (Step 3)

**Reasons**:
- Encrypted data is corrupted (AEAD authentication tag mismatch)
- Incorrect master password (PBKDF2 key derivation mismatch)
- Invalid JSON structure (parse error)
- Filesystem read error on temporary file

**Message Format**:
```
"save failed during verification. Your vault was not modified. {specific reason}. {actionable guidance}"
```

**Example**:
```
"save failed during verification. Your vault was not modified. The encrypted data could not be decrypted. Check your master password and try again."
```

**Recovery**: User retries save operation after fixing issue (e.g., correct password, free up disk space)

---

### ErrDiskSpaceExhausted

**When**: Insufficient disk space to write temporary file (Step 2)

**Reasons**:
- Vault size + temp file size exceeds available disk space
- Filesystem quota exceeded
- Disk partition full

**Message Format**:
```
"save failed: insufficient disk space. Your vault was not modified. Free up at least {requiredMB} MB and try again."
```

**Example**:
```
"save failed: insufficient disk space. Your vault was not modified. Free up at least 50 MB and try again."
```

**Recovery**: User frees disk space, retries save operation

---

### ErrPermissionDenied

**When**: Cannot write to vault directory or temp file (Step 2, 4, or 5)

**Reasons**:
- Vault directory is read-only
- File permissions prevent write (not owner, no write bit)
- SELinux/AppArmor policy denies write

**Message Format**:
```
"save failed: permission denied for {location}. Your vault was not modified. Check file permissions for {vaultDir} and try again."
```

**Example**:
```
"save failed: permission denied for vault directory. Your vault was not modified. Check file permissions for /home/user/.config/pass-cli and try again."
```

**Recovery**: User fixes permissions (`chmod 0700 ~/.config/pass-cli`), retries save

---

### ErrFilesystemNotAtomic

**When**: Rename operation not supported on filesystem (Step 4 or 5)

**Reasons**:
- Network filesystem (NFS, SMB) doesn't guarantee atomic rename
- Filesystem bug or limitation
- Cross-device rename attempted (vault and temp on different mount points)

**Message Format**:
```
"save failed: filesystem does not support atomic operations. Your vault was not modified. Move your vault to a local filesystem (not NFS/SMB)."
```

**Example**:
```
"save failed: filesystem does not support atomic operations. Your vault was not modified. Move your vault to a local filesystem (not NFS/SMB)."
```

**Recovery**: User moves vault to local filesystem, updates config, retries

---

## Observable Side Effects

### Audit Log Events (in order)

```
1. atomic_save_started        { "vault": "/path/to/vault.enc" }
2. temp_file_created          { "path": "/path/to/vault.enc.tmp.20251108-143022.a3f8c2" }
3. verification_started       { "temp": "/path/to/vault.enc.tmp.20251108-143022.a3f8c2" }
4. verification_passed        { "temp": "/path/to/vault.enc.tmp.20251108-143022.a3f8c2", "duration_ms": 150 }
5. atomic_rename_started      { "vault": "/path/to/vault.enc" }
6. atomic_rename_completed    { "vault": "/path/to/vault.enc", "duration_ms": 4850 }
7. cleanup_orphaned_files     { "removed": [] }  // empty if no orphaned files
```

### Failure Path Audit Log (verification fails)

```
1. atomic_save_started        { "vault": "/path/to/vault.enc" }
2. temp_file_created          { "path": "/path/to/vault.enc.tmp.20251108-143022.a3f8c2" }
3. verification_started       { "temp": "/path/to/vault.enc.tmp.20251108-143022.a3f8c2" }
4. verification_failed        { "temp": "/path/to/vault.enc.tmp.20251108-143022.a3f8c2", "reason": "corrupted encryption" }
5. rollback_completed         { "temp": "/path/to/vault.enc.tmp.20251108-143022.a3f8c2", "duration_ms": 50 }
```

---

## Thread Safety

**NOT thread-safe**: `SaveVault()` assumes single-process vault access per Constitution assumption.

**Concurrent Call Behavior**:
- Multiple goroutines calling `SaveVault()` will create conflicting temporary files
- Random suffix prevents file name collision, but **last writer wins**
- No mutex/lock protection - callers must serialize access

**Rationale**: Single-process assumption (Constitution Principle VII: Simplicity) - adding locking would violate YAGNI. If multi-process access needed in future, use file locking (flock) or advisory locks.

---

## Platform-Specific Behavior

### Windows

**Rename Implementation**: `os.Rename()` → `MoveFileExW` with `MOVEFILE_REPLACE_EXISTING` flag

**Atomicity**: Guaranteed atomic replacement if source and destination on same volume

**File Permissions**: Inherits ACLs from parent directory (equivalent to 0600 Unix - owner-only access)

**Error Codes**:
- `ERROR_ACCESS_DENIED` → `ErrPermissionDenied`
- `ERROR_DISK_FULL` → `ErrDiskSpaceExhausted`
- `ERROR_NOT_SAME_DEVICE` → `ErrFilesystemNotAtomic`

---

### macOS / Linux (POSIX)

**Rename Implementation**: `os.Rename()` → `rename(2)` syscall

**Atomicity**: POSIX guarantees atomic replacement per specification

**File Permissions**: Explicit 0600 (owner read/write only) set via `os.OpenFile()` mode parameter

**Error Codes**:
- `EACCES` / `EPERM` → `ErrPermissionDenied`
- `ENOSPC` / `EDQUOT` → `ErrDiskSpaceExhausted`
- `EXDEV` → `ErrFilesystemNotAtomic`

---

## Performance Guarantees

### Success Path Latency

| Operation | Target | Typical | Worst Case |
|-----------|--------|---------|------------|
| Temp file write | <1s | 50-100ms | 2s (large vault) |
| Verification decrypt | <500ms | 100-150ms | 1s (100MB vault) |
| Atomic rename (2x) | <100ms | 10-20ms | 200ms (slow disk) |
| Orphaned file cleanup | <100ms | 5-10ms | 500ms (many orphans) |
| **Total** | **<5s** | **<500ms** | **4s** |

**Performance Target**: SC-009 requires <5 second save time (excluding user password entry)

---

### Failure Path Latency

| Operation | Target | Typical |
|-----------|--------|---------|
| Temp file removal | <100ms | 10ms |
| Error message formatting | <10ms | 1ms |
| Audit log write | <50ms | 5ms |
| **Total** | **<1s** | **<50ms** |

**Rollback Target**: SC-008 requires <1 second rollback time

---

### Resource Usage

| Resource | Usage | Peak | Cleanup |
|----------|-------|------|---------|
| Disk Space | 2x vault size | During write (temp + original) | Temp deleted after rename |
| Memory | 1x vault size | During verification | Cleared via crypto.ClearBytes |
| File Descriptors | 3 (vault, backup, temp) | During rename | Auto-closed by os.Rename |
| CPU | ~100ms decrypt | During verification | N/A |

---

## Contract Validation Tests

### Success Path Tests

```go
TestSaveVault_Success
- Given: Valid vault data, correct password
- When: SaveVault() called
- Then: Returns nil, vault.enc contains new data, vault.enc.backup contains old data

TestSaveVault_PerformanceTarget
- Given: 50KB typical vault
- When: SaveVault() called
- Then: Completes in <5 seconds (benchmark)

TestSaveVault_PermissionsPreserved
- Given: vault.enc has 0600 permissions
- When: SaveVault() called
- Then: vault.enc still has 0600 permissions
```

### Failure Path Tests

```go
TestSaveVault_VerificationFailure
- Given: Corrupted encrypted data
- When: SaveVault() called
- Then: Returns ErrVerificationFailed, vault.enc unchanged, temp removed

TestSaveVault_DiskSpaceExhausted
- Given: Insufficient disk space for temp file
- When: SaveVault() called
- Then: Returns ErrDiskSpaceExhausted, vault.enc unchanged, temp removed

TestSaveVault_PermissionDenied
- Given: Read-only vault directory
- When: SaveVault() called
- Then: Returns ErrPermissionDenied, vault.enc unchanged

TestSaveVault_RollbackPerformance
- Given: Verification will fail
- When: SaveVault() called
- Then: Rollback completes in <1 second (benchmark)
```

### Audit Logging Tests

```go
TestSaveVault_AuditLogSuccess
- Given: Valid save operation
- When: SaveVault() called
- Then: Audit log contains atomic_save_started, verification_passed, atomic_rename_completed

TestSaveVault_AuditLogFailure
- Given: Verification will fail
- When: SaveVault() called
- Then: Audit log contains verification_failed, rollback_completed
```

### Security Tests

```go
TestSaveVault_NoCredentialLogging
- Given: Valid save operation
- When: SaveVault() called
- Then: Audit log NEVER contains decrypted vault content or password

TestSaveVault_MemoryClearing
- Given: Valid save operation
- When: SaveVault() completes
- Then: Decrypted memory zeroed (verify via memory inspection or fuzzing)

TestSaveVault_TempFilePermissions
- Given: Valid save operation
- When: Temp file created
- Then: Temp file has 0600 permissions (same as vault.enc)
```

---

## Breaking Changes from Previous Contract

**Changed Behavior**:
1. **Backup timing**: Old contract created backup BEFORE save, new contract creates backup DURING save (as part of atomic rename)
2. **Verification**: Old contract did NOT verify new vault data before commit, new contract REQUIRES verification
3. **Error handling**: Old contract could leave vault partially written on crash, new contract GUARANTEES atomicity

**Preserved Behavior**:
1. **API signature**: `SaveVault(data []byte, password string) error` unchanged
2. **Return value**: Still returns `nil` on success, `error` on failure
3. **Caller expectations**: CLI commands (add/update/delete) require no changes

**Migration Notes**: Existing code calling `SaveVault()` will automatically benefit from atomic save - no changes required. Tests may need updating if they assert on backup creation timing.

---

## Contract Summary

**Method**: `SaveVault(data []byte, password string) error`

**Guarantees**:
- Atomicity: Vault never in corrupt state after save
- Verification: New data always decryptable before commit
- Logging: All state transitions recorded in audit.log
- Performance: <5s save, <1s rollback
- Security: Temp files inherit vault permissions, memory cleared after verification

**Failure Modes**: 4 error types with actionable user messages

**Observable Side Effects**: 7 audit log events (success path), 5 events (failure path)

**Thread Safety**: NOT thread-safe (single-process assumption)

**Platform Support**: Windows, macOS, Linux - all use atomic rename

**Ready for**: Implementation and test-driven development
