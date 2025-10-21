# Research: Keychain Lifecycle Implementation

**Feature**: 011-keychain-lifecycle-management
**Date**: 2025-10-20
**Purpose**: Document existing patterns for password memory handling and audit logging integration

## Overview

This research resolves all "NEEDS CLARIFICATION" items from plan.md Technical Context and Constitution Check, specifically:
- How pass-cli handles password memory zeroing (Constitution Principle I requirement)
- How to integrate with existing audit log system (FR-015 requirement)

## Decision 1: Password Memory Zeroing

**Decision**: Use existing `crypto.ClearBytes()` function with `defer` pattern immediately after password allocation.

**Rationale**: Pass-cli already implements secure password memory handling throughout the codebase:
- All password parameters use `[]byte` type (never `string`)
- `crypto.ClearBytes()` at `internal/crypto/crypto.go:152-165` uses constant-time operations to prevent compiler optimization
- Consistent `defer crypto.ClearBytes(password)` pattern used in cmd/change_password.go, internal/vault/vault.go

**Alternatives Considered**:
1. **Manual zeroing**: Less reliable, easy to forget on error paths
2. **GC finalize**: Not immediate, vulnerable to memory dumps before GC runs
3. **String conversion then zeroing**: Strings are immutable in Go, copies remain in memory

**Code References**:
- `cmd/change_password.go:54` - Current password cleared with defer
- `cmd/change_password.go:95` - Confirm password cleared with defer
- `internal/vault/vault.go:173` - Master password cleared in Initialize()
- `internal/vault/vault.go:260` - Master password cleared in Unlock()
- `internal/crypto/crypto.go:152-165` - ClearBytes implementation using subtle.ConstantTimeCompare

**Implementation Pattern**:
```go
password, err := readPassword()
if err != nil {
    return fmt.Errorf("failed to read password: %w", err)
}
defer crypto.ClearBytes(password)  // Clears on all code paths (success + error)
```

## Decision 2: Audit Log Integration

**Decision**: Reuse existing `security.AuditLogger` with new event types for keychain lifecycle operations.

**Rationale**:
- Pass-cli has mature audit logging system at `internal/security/audit.go`
- System already logs vault operations (unlock, lock, password change, credential access)
- Supports HMAC signatures for tamper detection
- Graceful degradation (logs warnings but doesn't fail operations if logging fails)

**Implementation Pattern**:
```go
// 1. Enable audit logging (if not already enabled)
if vaultData.AuditEnabled {
    auditLogPath := vaultData.AuditLogPath
    vaultID := vaultData.VaultID
    if err := vaultService.EnableAudit(auditLogPath, vaultID); err != nil {
        fmt.Fprintf(os.Stderr, "Warning: failed to enable audit logging: %v\n", err)
    }
}

// 2. Log keychain lifecycle operations
vaultService.logAudit("keychain_enable", security.OutcomeSuccess, "")
vaultService.logAudit("keychain_status", security.OutcomeSuccess, "")
vaultService.logAudit("vault_remove", security.OutcomeSuccess, "")
```

**New Event Type Constants** (to add to `internal/security/audit.go:28-40`):
```go
EventKeychainEnable = "keychain_enable"
EventKeychainStatus = "keychain_status"
EventVaultRemove    = "vault_remove"
```

**Audit Entry Structure** (`security.AuditLogEntry`):
- `Timestamp`: Automatic (time.Now())
- `EventType`: One of the constants above
- `Outcome`: "success" or "failure"
- `CredentialName`: Empty string for vault-level operations (not credential-specific)
- `HMACSignature`: Automatic (calculated by AuditLogger)

**Code References**:
- `internal/security/audit.go:18-24` - AuditLogEntry struct
- `internal/security/audit.go:146` - Log() function
- `internal/security/audit.go:236` - NewAuditLogger() constructor
- `internal/vault/vault.go:115-140` - EnableAudit() implementation
- `internal/vault/vault.go:150-166` - logAudit() helper
- `cmd/helpers.go:35-57` - getAuditLogPath(), getVaultID() helpers

**Alternative Considered**: Create separate audit log for keychain operations
- **Rejected**: Fragmenting audit trail reduces security visibility; single consolidated log is better for forensics

## Decision 3: Keychain Service Integration

**Decision**: Use existing `internal/keychain` package with service name format "pass-cli:/absolute/path/to/vault.enc".

**Rationale**:
- Package already implements cross-platform abstractions (Windows Credential Manager, macOS Keychain, Linux Secret Service)
- Delete/Clear methods exist but unused (lines 94-105) - perfect for remove command
- Service name format documented in spec.md FR-003

**Existing API** (`internal/keychain/keychain.go`):
```go
type KeychainService struct {
    serviceName string  // Format: "pass-cli:/absolute/path/to/vault.enc"
    accountName string  // Always "master-password"
}

func New(serviceName, accountName string) (*KeychainService, error)
func (ks *KeychainService) IsAvailable() bool
func (ks *KeychainService) Store(password string) error
func (ks *KeychainService) Retrieve() (string, error)
func (ks *KeychainService) Delete() error  // Line 94 - UNUSED, perfect for vault remove
func (ks *KeychainService) Clear() error   // Line 100 - UNUSED
```

**Implementation Notes**:
1. `IsAvailable()` for status command to check keychain availability
2. `Store()` for enable command (after unlocking vault)
3. `Retrieve()` for status command to verify password is stored
4. `Delete()` for remove command to clean up keychain entry
5. Service name MUST use absolute path from GetVaultPath() to match existing entries

**Code References**:
- `internal/keychain/keychain.go:94-105` - Delete/Clear methods (currently unused)
- `internal/keychain/keychain.go:74` - IsAvailable() implementation
- `cmd/init.go:124-133` - Example of storing password in keychain
- `cmd/helpers.go:75-93` - unlockVault() helper shows keychain retrieval

## Decision 4: Password Prompt Pattern

**Decision**: Use existing `readPassword()` helper from `cmd/helpers.go:12-33`.

**Rationale**:
- Handles both terminal (asterisk masking) and non-terminal input
- Returns `[]byte` directly (secure type)
- Already used consistently across all commands (add, change_password, init, update)

**Implementation Pattern**:
```go
fmt.Print("Master password: ")
password, err := readPassword()
if err != nil {
    return fmt.Errorf("failed to read password: %w", err)
}
defer crypto.ClearBytes(password)
fmt.Println() // newline after password input
```

**Code Reference**: `cmd/helpers.go:12-33` - readPassword() implementation using gopass.GetPasswdMasked()

## Decision 5: Error Messages for Keychain Unavailability

**Decision**: Use platform-specific error messages with actionable guidance.

**Rationale**: FR-007 requires "clear error messages", SC-005 requires "clear, actionable" messages

**Platform-Specific Messages**:
```go
var unavailableMessages = map[string]string{
    "windows": "System keychain not available: Windows Credential Manager access denied.\nTroubleshooting: Check user permissions for Credential Manager access.",
    "darwin":  "System keychain not available: macOS Keychain access denied.\nTroubleshooting: Check Keychain Access.app permissions for pass-cli.",
    "linux":   "System keychain not available: Linux Secret Service not running or accessible.\nTroubleshooting: Ensure gnome-keyring or KWallet is installed and running.",
}

func getKeychainUnavailableMessage() string {
    msg, ok := unavailableMessages[runtime.GOOS]
    if !ok {
        return "System keychain not available on this platform."
    }
    return msg
}
```

**Code Location**: Add to `cmd/helpers.go` or new `cmd/keychain_helpers.go`

**Alternative Considered**: Generic error message across all platforms
- **Rejected**: Users need platform-specific troubleshooting guidance (Constitution Principle III - actionable errors)

## Implementation Checklist

Based on research findings, here's what needs to be implemented:

### cmd/keychain_enable.go
- [ ] Use `readPassword()` from helpers.go
- [ ] Apply `defer crypto.ClearBytes(password)` immediately
- [ ] Call `vaultService.Unlock(password)` to validate
- [ ] Call keychain.Store() after successful unlock
- [ ] Handle keychain unavailable with platform-specific message
- [ ] Log audit event (keychain_enable, success/failure)

### cmd/keychain_status.go
- [ ] No password prompt (read-only operation per FR-011)
- [ ] Call keychain.IsAvailable() to check system keychain
- [ ] Call keychain.Retrieve() to check if password stored (don't use retrieved password)
- [ ] Display backend name (Windows Credential Manager / macOS Keychain / Linux Secret Service)
- [ ] Include actionable suggestions (FR-014) if keychain not enabled
- [ ] Log audit event (keychain_status, success)

### cmd/vault_remove.go
- [ ] Prompt for confirmation (y/n) unless --yes/--force flag provided
- [ ] Delete vault file using os.Remove()
- [ ] Call keychain.Delete() to remove keychain entry
- [ ] Handle "vault file missing but keychain exists" edge case (FR-012)
- [ ] Check for file locks (optional: use flock-style library)
- [ ] Log audit event (vault_remove, success/failure) BEFORE deleting audit log itself

### internal/security/audit.go
- [ ] Add event type constants: EventKeychainEnable, EventKeychainStatus, EventVaultRemove
- [ ] No code changes needed (existing Log() function handles new event types)

### tests
- [ ] Unit tests for password memory zeroing (verify bytes cleared)
- [ ] Integration tests for enable command (vault without keychain → enable → subsequent commands don't prompt)
- [ ] Integration tests for status command (check output format, suggestions)
- [ ] Integration tests for remove command (verify both file and keychain deleted)
- [ ] Security tests for audit logging (verify FR-015 acceptance scenario #5)

## Risks & Mitigations

### Risk 1: Keychain unavailable on headless SSH sessions
**Mitigation**: Clear error message with graceful degradation (commands fail fast with actionable message, per FR-007)

### Risk 2: Partial failure (file deleted but keychain removal fails)
**Mitigation**: FR-012 allows remove command to succeed even if file missing - handle both failure modes gracefully

### Risk 3: Audit log deleted before logging removal operation
**Mitigation**: Log vault_remove event BEFORE deleting audit log file itself

### Risk 4: Platform differences in keychain behavior
**Mitigation**: Leverage existing 99designs/keyring package which abstracts platform differences; integration tests on all platforms (Windows/macOS/Linux) per CI matrix

## Next Steps

1. **Phase 1: Design** - Create data-model.md and contracts/commands.md
2. **Phase 2: Tasks** - Break down implementation into testable tasks (via `/speckit.tasks`)
3. **Implementation** - Follow TDD workflow (tests first, then implementation)
