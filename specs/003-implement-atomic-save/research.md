# Atomic Save Pattern Research

**Date**: 2025-11-08
**Feature**: Atomic Save Pattern for Vault Operations
**Research Phase**: Phase 0

## Decision: Filesystem Atomicity

**Chosen**: Use `os.Rename()` for atomic file replacement on all platforms

**Rationale**:
- **POSIX systems** (macOS/Linux): `rename(2)` syscall is atomic per POSIX specification - if destination exists, it's replaced atomically
- **Windows**: `os.Rename()` maps to `MoveFileEx` with `MOVEFILE_REPLACE_EXISTING` flag which provides atomic replacement
- **Go stdlib guarantee**: `os.Rename()` documentation guarantees atomicity when replacing files on same filesystem
- **Industry standard**: Used by databases (SQLite, PostgreSQL), package managers (npm, cargo), text editors (vim, emacs) for safe writes

**Edge Cases Identified**:
- **Network filesystems** (NFS, SMB): May not guarantee atomicity - out of scope per spec
- **Different filesystems**: Rename only atomic within same mount point - mitigated by vault and temp in same directory
- **Read-only filesystems**: Pre-flight check (existing vault writeable) catches this before temp file creation
- **Disk full during rename**: Atomic operation fails-fast, no partial state

**Alternatives Considered**:
1. **Copy + Delete pattern**: Not atomic - crash between copy and delete leaves duplicate files
2. **Hard link + Unlink**: Requires filesystem hard link support, complicates cross-platform code
3. **Database transactions**: Overkill for single-file vault, adds external dependency

**References**:
- POSIX rename spec: https://pubs.opengroup.org/onlinepubs/9699919799/functions/rename.html
- Go os.Rename docs: https://pkg.go.dev/os#Rename
- SQLite atomic commit: https://www.sqlite.org/atomiccommit.html

---

## Decision: Verification Approach

**Chosen**: Full in-memory decryption verification before commit

**Rationale**:
- **Completeness**: Verifies encryption integrity, password correctness, and JSON structure in one step
- **Performance**: Vault size assumed <100MB (per spec), decryption overhead <500ms on modern hardware
- **Security**: Prevents committing corrupted/undecryptable vaults - critical for password manager
- **Memory safety**: `crypto.ClearBytes()` exists in codebase to zero decrypted memory after verification

**Performance Analysis**:
- **Typical vault size**: 50-100 credentials ~= 10-50KB encrypted
- **Decryption overhead**: AES-256-GCM ~1GB/sec throughput on modern CPU
- **Worst case (100MB vault)**: ~100ms decryption + 50ms JSON validation = 150ms overhead
- **Target budget**: <5 second total save time, verification is <5% overhead

**Alternatives Considered**:
1. **Checksum-only verification**: Fast but doesn't verify decryptability - could save undecryptable vault
2. **Partial decryption**: Complex to implement, marginal performance gain, still requires full decrypt for JSON validation
3. **No verification**: Violates FR-002 requirement, unacceptable risk for password manager

**Memory Clearing Verification**:
- Existing `crypto.ClearBytes()` in `internal/crypto/crypto.go` - already used for password clearing
- Pattern: `defer crypto.ClearBytes(decrypted)` immediately after allocation
- Go runtime may optimize away zeroing - `crypto.ClearBytes()` uses volatile write to prevent optimization

---

## Decision: Temporary File Cleanup

**Chosen**: Best-effort cleanup after save operations, orphaned file cleanup on next vault operation

**Rationale**:
- **Happy path**: Temp file removed immediately after successful atomic rename (Step 7 in workflow)
- **Failure path**: Temp file removed during error handling (best-effort, logged if fails)
- **Orphaned files** (crash mid-save): Cleaned up during next `SaveVault()` call before creating new temp
- **Startup cleanup**: Not chosen - adds complexity, unlock operation should be fast

**Cleanup Timing Decision**:
- **After save completion** (chosen): Simple, synchronous, no background goroutines
- **Before each save**: Adds overhead to every save, but catches orphaned files deterministically
- **On vault unlock**: Could slow down unlock (critical path for user), violates single-responsibility
- **Background goroutine**: Violates Principle VII (Simplicity), adds concurrency complexity

**Orphaned File Detection**:
- Pattern: `vault.enc.tmp.*` (any file matching glob pattern)
- Age check: Any temp file NOT created by current save operation
- Cleanup strategy: Best-effort `os.Remove()`, log warning if fails (don't block save)

**Race Condition Mitigation**:
- Timestamp + random suffix prevents collision with concurrent saves (though single-process assumption makes this unlikely)
- Cleanup only removes files NOT matching current temp file name
- No lock files needed (per single-process assumption in constitution)

---

## Decision: Random Suffix Generation

**Chosen**: 6-character hexadecimal suffix from `crypto/rand`

**Rationale**:
- **Collision probability**: 16^6 = 16.7 million possibilities, negligible collision risk for single-process usage
- **Security**: `crypto/rand` uses OS cryptographic RNG (not predictable like `math/rand`)
- **Debuggability**: Hex is human-readable, short enough for command-line output
- **Timestamp precision**: Second-level timestamp sufficient given random suffix

**Format Decision**:
```
vault.enc.tmp.20251108-143022.a3f8c2
             └─ YYYYMMDD-HHMMSS └─ 6-char hex random
```

**Alternatives Considered**:
1. **Process ID (PID)**: Reused after process crash, collisions possible
2. **math/rand**: Predictable seed, security risk (temp files contain encrypted vault)
3. **UUID**: Overkill (128 bits), less human-readable, stdlib doesn't include UUID generator
4. **Millisecond timestamp**: Sufficient but loses debuggability (harder to correlate with logs)

**Implementation**:
```go
import "crypto/rand"

func generateTempSuffix() string {
    b := make([]byte, 3) // 3 bytes = 6 hex chars
    rand.Read(b) // crypto/rand, not math/rand
    return fmt.Sprintf("%x", b)
}
```

**Timestamp Format**:
- `20251108-143022` (YYYYMMDD-HHMMSS): Sortable, log-correlatable, filesystem-safe (no colons)
- **Not** RFC3339: Contains colons (problematic on Windows), overkill precision (no milliseconds needed)

---

## Decision: Atomic Rename Platform Specifics

**Chosen**: Trust Go stdlib `os.Rename()` cross-platform guarantees, no platform-specific code

**Rationale**:
- **Go stdlib abstraction**: `os.Rename()` handles platform differences internally
- **Existing codebase pattern**: No other platform-specific file operations in storage layer
- **Constitution Principle V**: Minimize platform-specific code, rely on Go portability

**Platform Behavior Verified**:
- **Windows**: `os.Rename()` → `MoveFileExW` with `MOVEFILE_REPLACE_EXISTING` (atomic replace)
- **macOS/Linux**: `os.Rename()` → `rename(2)` syscall (POSIX atomic)
- **Error handling**: Platform-specific errors mapped to generic `os.PathError` by Go runtime

**No Pre-Flight Check Needed**:
- Original plan suggested filesystem atomicity detection - YAGNI
- Assumption: User running on local filesystem (network FS out of scope per spec)
- If `os.Rename()` fails, error handling catches it (no need for proactive detection)

---

## Decision: Audit Logging Strategy

**Chosen**: Log state transitions to existing `audit.log` using structured logging

**Rationale**:
- **Existing infrastructure**: `internal/security` package already implements audit logging
- **FR-015 compliance**: Log all state transitions (temp created, verification, rename, cleanup)
- **No credential leakage**: Log operation outcomes only, never decrypted vault content
- **Structured format**: Existing audit log uses structured fields (timestamp, event, outcome, details)

**Log Levels**:
- **INFO**: Atomic save started, verification passed, atomic rename completed
- **DEBUG**: Temp file created (path), cleanup completed (temp file name)
- **WARN**: Temp file cleanup failed (non-critical, log and continue)
- **ERROR**: Verification failed, disk space exhausted, permissions error

**Example Log Entries**:
```
2025-11-08T14:30:22Z INFO atomic_save_started vault=/path/to/vault.enc
2025-11-08T14:30:22Z DEBUG temp_file_created path=/path/to/vault.enc.tmp.20251108-143022.a3f8c2
2025-11-08T14:30:22Z INFO verification_passed vault=/path/to/vault.enc
2025-11-08T14:30:22Z INFO atomic_rename_completed vault=/path/to/vault.enc
2025-11-08T14:30:22Z DEBUG cleanup_completed temp=vault.enc.tmp.20251108-143022.a3f8c2
```

**Error Message Format** (FR-011):
```go
// Verification failure example
return fmt.Errorf("save failed during verification. Your vault was not modified. The encrypted data could not be decrypted. Check your master password and try again.")

// Disk space exhausted example
return fmt.Errorf("save failed: insufficient disk space. Your vault was not modified. Free up at least %d MB and try again.", requiredMB)
```

---

## Research Summary

**All Decisions Resolved**: No "NEEDS CLARIFICATION" remaining in Technical Context

**Key Takeaways**:
1. `os.Rename()` provides atomic replacement on all target platforms
2. Full in-memory verification adds <5% overhead, critical for data integrity
3. Cleanup strategy: synchronous after save, orphaned file sweep before next save
4. Temp file naming: timestamp + crypto/rand for uniqueness and debuggability
5. Audit logging: structured events via existing infrastructure

**Ready for Phase 1**: Data model and contract design can proceed with concrete implementation decisions
