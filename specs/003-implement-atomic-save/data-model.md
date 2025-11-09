# Data Model: Atomic Save Pattern

**Feature**: Atomic Save Pattern for Vault Operations
**Phase**: Phase 1 - Design
**Date**: 2025-11-08

## Overview

This document defines the data structures, state machines, and workflows for the atomic save pattern. All definitions are technology-agnostic representations of the logical model - implementation details belong in code.

---

## 1. Vault File State Machine

### States

| State | Description | Observable Indicators |
|-------|-------------|----------------------|
| `Consistent` | Vault file exists, is readable, and can be decrypted | `vault.enc` exists, `os.Stat()` succeeds, unlock succeeds |
| `Saving` | Save operation in progress | Temporary file exists (`vault.enc.tmp.*`), audit log shows "atomic_save_started" |
| `Corrupt` | Vault file exists but cannot be decrypted or is malformed | `vault.enc` exists, unlock fails with decryption error |
| `Missing` | Vault file does not exist | `vault.enc` absent, `os.Stat()` returns `os.ErrNotExist` |

### State Transitions

```
┌───────────┐
│  Missing  │
└─────┬─────┘
      │ vault init
      ▼
┌─────────────┐     SaveVault()      ┌────────┐
│ Consistent  ├─────────────────────►│ Saving │
└─────────────┘                      └───┬─┬──┘
      ▲                                  │ │
      │                                  │ │ verification success
      │         atomic rename succeeds   │ │ + rename succeeds
      └──────────────────────────────────┘ │
      │                                    │
      │         verification fails         │
      │         or rename fails            │
      └────────────────────────────────────┘
                  (rollback)

┌─────────────┐     manual filesystem     ┌─────────┐
│ Consistent  │     corruption (external)  │ Corrupt │
└─────────────┘◄───────────────────────────┴─────────┘
      │                                         │
      │                                         │
      └─────────────restore from backup────────┘
```

### Invariants

1. **Atomicity Guarantee**: State NEVER transitions from `Consistent` to `Corrupt` via `SaveVault()` operation
2. **Single Writer**: Only one save operation in `Saving` state at a time (single-process assumption)
3. **Backup Consistency**: When in `Consistent` state, `vault.enc.backup` (if exists) represents N-1 generation
4. **No Orphans**: When in `Consistent` state, no temporary files (`vault.enc.tmp.*`) exist

---

## 2. Save Operation Workflow

### Inputs

| Parameter | Type | Constraints | Source |
|-----------|------|-------------|--------|
| `vaultData` | `[]byte` | Encrypted vault JSON, non-empty | Vault service after credential modification |
| `masterPassword` | `string` | Valid master password | User input or OS keychain |
| `vaultPath` | `string` | Absolute path to vault file | Configuration (`~/.config/pass-cli/vault.enc`) |

### Outputs

| Output | Type | Values | Description |
|--------|------|--------|-------------|
| `error` | `error` | `nil` or specific error type | Success (nil) or failure with user-facing message |

### Workflow Steps

```
Step 1: Generate temp filename
├─► Input: vaultPath, current timestamp, crypto random bytes
├─► Output: tempPath = "vault.enc.tmp.YYYYMMDD-HHMMSS.XXXXXX"
└─► Error: None (crypto.rand.Read never fails)

Step 2: Write encrypted data to temp file
├─► Action: os.OpenFile(tempPath, O_WRONLY|O_CREATE|O_TRUNC, 0600)
├─► Action: Write(vaultData)
├─► Action: Fsync() [force disk commit]
├─► Output: Temp file on disk with vault permissions
└─► Error: ErrDiskSpaceExhausted, ErrPermissionDenied

Step 3: Verify temp file is decryptable
├─► Action: Read temp file into memory
├─► Action: Decrypt using masterPassword
├─► Action: Validate JSON structure
├─► Action: Clear decrypted memory (defer crypto.ClearBytes)
├─► Output: Verification passed
└─► Error: ErrVerificationFailed (cannot decrypt, invalid JSON, wrong password)

Step 4: Atomic rename (vault → backup)
├─► Action: os.Rename(vaultPath, vaultPath+".backup")
├─► Output: Old vault becomes backup, overwrites previous backup
└─► Error: ErrPermissionDenied, ErrFilesystemNotAtomic

Step 5: Atomic rename (temp → vault)
├─► Action: os.Rename(tempPath, vaultPath)
├─► Output: Verified temp becomes new vault
└─► Error: ErrPermissionDenied (extremely rare, Step 4 succeeded)

Step 6: Cleanup orphaned temp files (best effort)
├─► Action: Find all vault.enc.tmp.* NOT matching current tempPath
├─► Action: os.Remove each orphaned file
├─► Output: Orphaned files removed
└─► Error: Log warning, continue (cleanup is non-critical)

Step 7: Log success
├─► Action: Audit log "atomic_rename_completed"
└─► Output: State transition recorded
```

### Failure Handling

| Failure Point | Recovery Action | Final State | User Impact |
|---------------|-----------------|-------------|-------------|
| Step 2 (write temp) | Remove temp file, return error | `Consistent` (vault unchanged) | Save rejected, vault safe |
| Step 3 (verification) | Remove temp file, return error | `Consistent` (vault unchanged) | Save rejected, vault safe |
| Step 4 (vault→backup) | Remove temp file, return error | `Consistent` (vault unchanged) | Save rejected, vault safe |
| Step 5 (temp→vault) | **Critical failure** - log error, manual recovery | `Corrupt` (backup exists) | Manual restore required |
| Step 6 (cleanup) | Log warning, continue | `Consistent` (orphaned files remain) | Non-critical, cleaned next save |

**Step 5 Critical Failure Handling**:
- **Extremely rare**: Step 4 succeeded (vault→backup rename worked), but Step 5 (temp→vault) failed
- **State**: `vault.enc.backup` exists (N-1 generation), `vault.enc` missing or corrupt, `temp` file may exist
- **Recovery**: Manual intervention - rename `vault.enc.backup` to `vault.enc`
- **Mitigation**: Pre-flight check verifies filesystem supports rename, Step 4 success implies Step 5 should succeed

---

## 3. File Naming Convention

### Primary Files

| File | Pattern | Purpose | Lifecycle |
|------|---------|---------|-----------|
| Vault | `vault.enc` | Active encrypted vault | Persistent, modified by atomic rename |
| Backup | `vault.enc.backup` | N-1 generation vault | Persistent, overwritten each save, deleted on unlock |
| Temporary | `vault.enc.tmp.YYYYMMDD-HHMMSS.XXXXXX` | Staging area for verification | Transient, created during save, removed after commit |

### Naming Rules

**Timestamp Format**: `YYYYMMDD-HHMMSS`
- Example: `20251108-143022` (November 8, 2025, 14:30:22)
- Sortable, filesystem-safe (no colons), log-correlatable

**Random Suffix**: 6 hexadecimal characters
- Example: `a3f8c2`
- Generated from `crypto/rand` (3 bytes → 6 hex chars)
- Collision probability: 1 in 16.7 million (acceptable for single-process usage)

**Complete Example**:
```
vault.enc                           ← Active vault
vault.enc.backup                    ← Previous generation
vault.enc.tmp.20251108-143022.a3f8c2 ← Temporary file during save
```

### Orphaned File Detection

**Pattern**: Any file matching glob `vault.enc.tmp.*` NOT created by current save operation

**Cleanup Strategy**:
1. Before each save, scan vault directory for `vault.enc.tmp.*`
2. Remove all temp files (they are from crashed previous saves)
3. Log warning for each orphaned file removed
4. Continue with save operation even if cleanup fails

---

## 4. Verification Data Flow

### Verification Process

```
Input: tempFilePath, masterPassword
  │
  ▼
┌──────────────────────────────┐
│ Read temp file → []byte      │
│ (file size must fit in RAM)  │
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Decrypt using AES-256-GCM     │
│ (PBKDF2 key derivation)       │
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Parse as JSON                 │
│ (validate structure)          │
└──────────────┬───────────────┘
               ▼
┌──────────────────────────────┐
│ Clear decrypted memory        │
│ (crypto.ClearBytes)           │
└──────────────┬───────────────┘
               ▼
Output: nil (success) or error (failure)
```

### Memory Constraints

| Constraint | Value | Rationale |
|------------|-------|-----------|
| Max vault size (in-memory) | 100 MB | Per spec assumption, typical vault <50 KB |
| Decryption buffer | 1x vault size | Single allocation, cleared after verification |
| JSON parsing overhead | ~2x decrypted size | Go JSON parser allocates temporary structures |

### Verification Failure Scenarios

| Scenario | Detection | Error Message |
|----------|-----------|---------------|
| Corrupted encryption | Decrypt fails (AEAD auth tag mismatch) | "Verification failed: encrypted data is corrupted" |
| Wrong password | Decrypt fails (PBKDF2 key mismatch) | "Verification failed: incorrect master password" |
| Invalid JSON | JSON parse fails | "Verification failed: vault structure is invalid" |
| Filesystem read error | File read fails | "Verification failed: cannot read temporary file" |

---

## 5. Audit Log Data Model

### Log Entry Structure

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | ISO8601 | Event occurrence time | `2025-11-08T14:30:22Z` |
| `event` | string | Event type identifier | `atomic_save_started` |
| `outcome` | string | Success or failure | `success` or `failed` |
| `vault_path` | string | Vault file path (NOT content) | `/home/user/.config/pass-cli/vault.enc` |
| `details` | map | Event-specific metadata | `{"temp_file": "vault.enc.tmp.xxx"}` |

### Logged Events

| Event | Trigger | Outcome | Details |
|-------|---------|---------|---------|
| `atomic_save_started` | SaveVault() called | Always `success` | `{"vault": path}` |
| `temp_file_created` | Step 2 completes | `success` or `failed` | `{"path": tempPath}` or `{"error": reason}` |
| `verification_started` | Step 3 begins | Always `success` | `{"temp": tempPath}` |
| `verification_passed` | Step 3 completes | `success` | `{"temp": tempPath, "duration_ms": N}` |
| `verification_failed` | Step 3 decrypt/parse fails | `failed` | `{"temp": tempPath, "reason": error}` |
| `atomic_rename_started` | Step 4 begins | Always `success` | `{"vault": path}` |
| `atomic_rename_completed` | Step 5 completes | `success` | `{"vault": path}` |
| `rollback_completed` | Error handler removes temp | `success` | `{"temp": tempPath}` |
| `cleanup_orphaned_files` | Step 6 runs | `success` or `warning` | `{"removed": [paths]}` or `{"failed": [paths]}` |

### Security Constraints

**NEVER log**:
- Decrypted vault content
- Master password
- Credential names or values
- Encryption keys or salts
- File contents (only paths)

**DO log**:
- File paths (vault, temp, backup)
- Operation outcomes (success/fail)
- Error reasons (user-facing messages only)
- Performance metrics (duration in ms)

---

## 6. Error Types & Messages

### Error Type Hierarchy

```
SaveError (base type)
├── ErrVerificationFailed
│   ├── reason: "corrupted encryption"
│   ├── reason: "incorrect password"
│   └── reason: "invalid JSON structure"
├── ErrDiskSpaceExhausted
├── ErrPermissionDenied
│   ├── location: "vault directory"
│   ├── location: "temp file"
│   └── location: "backup file"
└── ErrFilesystemNotAtomic
    └── reason: "rename operation not supported"
```

### User-Facing Error Messages (FR-011)

| Error Type | Message Template | Actionable Guidance |
|------------|------------------|---------------------|
| `ErrVerificationFailed` | "Save failed during verification. Your vault was not modified. {reason}" | "Check your master password and try again." |
| `ErrDiskSpaceExhausted` | "Save failed: insufficient disk space. Your vault was not modified." | "Free up at least {requiredMB} MB and try again." |
| `ErrPermissionDenied` | "Save failed: permission denied for {location}. Your vault was not modified." | "Check file permissions for {vaultDir} and try again." |
| `ErrFilesystemNotAtomic` | "Save failed: filesystem does not support atomic operations. Your vault was not modified." | "Move your vault to a local filesystem (not NFS/SMB)." |

---

## Data Model Summary

**Entities**: 4 (Vault File, Backup File, Temporary File, Audit Log Entry)
**State Machine**: 4 states, 6 transitions, 4 invariants
**Workflow Steps**: 7 steps with failure handling for each
**File Patterns**: 3 (vault, backup, temp with timestamp+random)
**Audit Events**: 9 logged event types
**Error Types**: 4 primary error types with user-facing messages

**Ready for**: Contract definition and implementation
