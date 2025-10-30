# Security Hardening Research

**Feature**: Security Hardening
**Branch**: `005-security-hardening-address`
**Date**: 2025-10-11
**Phase**: 0 (Research & Discovery)

## Overview

This document consolidates research findings for the security hardening implementation. All technical unknowns from `plan.md` have been investigated and decisions documented.

---

## Decision 1: Terminal Input Library

**Question**: Does current `gopass` library support `[]byte` password input, or is migration to `golang.org/x/term` required for FR-001/FR-005 compliance?

**Chosen**: Keep existing `gopass` library (no migration needed)

**Rationale**:
- `cmd/helpers.go:25` shows `gopass.GetPasswdMasked()` already returns `[]byte`
- Current implementation at line 30: `return string(passwordBytes), nil` converts to string unnecessarily
- Simple fix: Remove `string()` conversion in `readPassword()` function
- No breaking external API change required
- `gopass` is actively maintained and cross-platform compatible

**Alternatives Considered**:
- `golang.org/x/term.ReadPassword()`: Also returns `[]byte`, but requires dependency change
  - Rejected: No benefit over current library, introduces migration risk
- Custom implementation: Could use `term.IsTerminal()` + syscalls
  - Rejected: Reinventing the wheel, cross-platform complexity

**Implementation Impact**: Minimal - single line change in `cmd/helpers.go`

---

## Decision 2: Current Crypto Package API

**Current Signature**: `DeriveKey(password string, salt []byte) ([]byte, error)` (crypto.go:44)

**Refactoring Plan**:

### Phase 1: Add Iteration Parameter
```go
// New signature (backward compatible via function overload pattern)
func (c *CryptoService) DeriveKeyWithIterations(password []byte, salt []byte, iterations int) ([]byte, error) {
    defer ClearBytes(password) // Caller-managed clearing
    key := pbkdf2.Key(password, salt, iterations, KeyLength, sha256.New)
    return key, nil
}

// Deprecated: Use DeriveKeyWithIterations
func (c *CryptoService) DeriveKey(password string, salt []byte) ([]byte, error) {
    return c.DeriveKeyWithIterations([]byte(password), salt, Iterations)
}
```

### Phase 2: Update Callers
- `storage.go:106, 239`: Pass iteration count from metadata
- `vault.go`: Convert master password to `[]byte` at entry points

**Memory Clearing Discovery**:
- `crypto.go:150-159`: `clearBytes()` function already exists!
- Uses `crypto/subtle.ConstantTimeCompare` to prevent compiler optimization (line 158)
- Currently only used internally (line 50) - expose as public `ClearBytes(data []byte)`

**Migration Strategy**:
1. Extract `clearBytes` → public `ClearBytes` helper
2. Add `DeriveKeyWithIterations` with `[]byte` password
3. Update all callers to use new signature
4. Remove deprecated `DeriveKey` after migration complete

---

## Decision 3: Current Vault Package API

**Current Signatures**:
- `Initialize(masterPassword string, useKeychain bool)` (vault.go:96)
- `Unlock(masterPassword string)` (vault.go:141)
- `ChangePassword(newPassword string)` (vault.go:514)

**Critical Vulnerability Found**: Line 65 stores `masterPassword string` in memory!
- String in Go is immutable → cannot be zeroed
- Remains in memory until GC collects AND overwrites
- Lines 185-191: Attempted string clearing ineffective (Go compiler optimizes away)

**Refactoring Plan**:

### Change `masterPassword` field type
```go
type VaultService struct {
    // OLD: masterPassword string
    masterPassword []byte // NEW: Can be securely zeroed
    // ...
}
```

### Update method signatures
```go
func (v *VaultService) Initialize(masterPassword []byte, useKeychain bool) error
func (v *VaultService) Unlock(masterPassword []byte) error
func (v *VaultService) ChangePassword(newPassword []byte) error
```

### Fix Lock() method (lines 180-194)
```go
func (v *VaultService) Lock() {
    v.unlocked = false

    // Secure clearing for []byte
    if v.masterPassword != nil {
        crypto.ClearBytes(v.masterPassword)
        v.masterPassword = nil
    }

    v.vaultData = nil
}
```

**Backward Compatibility**: Internal API only - no public breaking changes

---

## Decision 4: TUI Framework Identification

**Framework**: `github.com/rivo/tview` (confirmed at cmd/tui/components/forms.go:10)

**Real-Time Strength Indicator Implementation** (FR-017):

### CLI Mode (cmd/add.go, cmd/init.go)
```go
// Text-based indicator after each keystroke
fmt.Fprintf(os.Stderr, "\rPassword strength: %s ", strengthIndicator(password))
// strengthIndicator() returns: "⚠ Weak", "⚠ Medium", "✓ Strong"
```

### TUI Mode (cmd/tui/components/forms.go)
```go
// Add visual meter component below password field
strengthMeter := tview.NewProgressBar()
strengthMeter.SetMax(3) // weak=1, medium=2, strong=3
strengthMeter.SetFilledColor(tcell.ColorGreen)

// Update on password change
passwordField.SetChangedFunc(func(text string) {
    strength := calculateStrength(text)
    strengthMeter.SetProgress(strength)
})
```

**Alternatives Considered**:
- `bubbletea`: Not used in this project
- `termui`: Rejected - tview already integrated

---

## Decision 5: Current PBKDF2 Implementation

**Current State**:
- Iteration count: Hardcoded constant `Iterations = 100000` (crypto.go:19)
- Metadata: `VaultMetadata` struct (storage.go:30-35) has NO iteration count field
- Version: All vaults currently at version 1

**Metadata Versioning Plan**:

### Add Iterations field to VaultMetadata
```go
type VaultMetadata struct {
    Version    int       `json:"version"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
    Salt       []byte    `json:"salt"`
    Iterations int       `json:"iterations"` // NEW FIELD
}
```

### Backward Compatibility Strategy
```go
// When loading old vaults (no iterations field)
if metadata.Iterations == 0 {
    metadata.Iterations = 100000 // Legacy default
}

// When saving vaults
if metadata.Iterations < 600000 {
    // Keep existing iteration count (don't force upgrade on every save)
    // Only upgrade when user explicitly changes password
}
```

### Migration Trigger (FR-008)
```go
func (v *VaultService) ChangePassword(newPassword []byte) error {
    // ... existing validation ...

    // Opportunity to upgrade iterations
    newIterations := 600000 // OWASP 2023 recommendation

    // Update metadata and re-encrypt with stronger KDF
    metadata.Iterations = newIterations
    // ... save vault ...
}
```

**Configurable Iterations** (FR-010):
```go
const (
    MinIterations = 600000 // OWASP 2023 minimum
    DefaultIterations = 600000
)

// Environment variable override for power users
func getIterations() int {
    if env := os.Getenv("PASS_CLI_ITERATIONS"); env != "" {
        if val, err := strconv.Atoi(env); err == nil && val >= MinIterations {
            return val
        }
    }
    return DefaultIterations
}
```

---

## Decision 6: Memory Clearing Best Practices in Go

**Canonical Pattern** (from crypto.go:150-159):

```go
func ClearBytes(data []byte) {
    for i := range data {
        data[i] = 0
    }

    // Prevent compiler from optimizing away the zeroing
    // Uses constant-time comparison as a side effect
    dummy := make([]byte, len(data))
    subtle.ConstantTimeCompare(data, dummy)
}
```

**Why This Works**:
- `subtle.ConstantTimeCompare` is a compiler barrier (documented in Go stdlib)
- Function has observable side effects → compiler cannot eliminate zeroing
- Cross-platform compatible (unlike `mlock`/`VirtualLock`)

**Deferred Cleanup Pattern** (FR-003):
```go
func someFunction() error {
    password := getPasswordInput() // Returns []byte
    defer ClearBytes(password)

    // Use password...
    // Even if panic occurs, defer ensures clearing

    return nil
}
```

**Limitations Documented** (Spec Assumption 4):
- Go GC may have copied data before we clear it
- Freed memory not overwritten until reallocation
- "Best effort" security within Go's memory model
- Cannot achieve military-grade security without unsafe/C bindings

**Alternatives Considered**:
- `runtime.KeepAlive()`: Not sufficient - doesn't prevent optimization
- `mlock()` syscall: Platform-specific, requires root on some systems
- `memguard` library: External dependency, supply chain risk (violates Constitution VII)

---

## Decision 7: HMAC Audit Log Format

**HMAC Signature Scheme** (FR-022, FR-034, FR-035):

### Key Management
```go
// Generate unique HMAC key per vault, store in OS keychain
// Allows log verification without master password (FR-035)
func generateAuditKey(vaultUUID string) ([]byte, error) {
    // Generate cryptographically random 32-byte key
    key := make([]byte, 32)
    if _, err := rand.Read(key); err != nil {
        return nil, err
    }

    // Store in OS keychain using 99designs/keyring
    kr, err := keyring.Open(keyring.Config{
        ServiceName: "pass-cli-audit",
    })
    if err != nil {
        return nil, err
    }

    keyID := fmt.Sprintf("pass-cli-audit-%s", vaultUUID)
    err = kr.Set(keyring.Item{
        Key:  keyID,
        Data: key,
    })

    return key, err
}

// Retrieve audit key from keychain for verification
func getAuditKey(vaultUUID string) ([]byte, error) {
    kr, err := keyring.Open(keyring.Config{
        ServiceName: "pass-cli-audit",
    })
    if err != nil {
        return nil, err
    }

    keyID := fmt.Sprintf("pass-cli-audit-%s", vaultUUID)
    item, err := kr.Get(keyID)
    if err != nil {
        return nil, err
    }

    return item.Data, nil
}
```

### Log Entry Format
```json
{
  "timestamp": "2025-10-11T14:30:00Z",
  "event_type": "vault_unlock",
  "outcome": "success",
  "credential_name": "",
  "hmac": "a1b2c3d4..."
}
```

### HMAC Calculation
```go
func signEntry(entry AuditLogEntry, key []byte) []byte {
    // Canonical serialization (order matters for verification)
    data := fmt.Sprintf("%s|%s|%s|%s",
        entry.Timestamp.Format(time.RFC3339),
        entry.EventType,
        entry.Outcome,
        entry.CredentialName,
    )

    mac := hmac.New(sha256.New, key)
    mac.Write([]byte(data))
    return mac.Sum(nil)
}
```

### Tamper Detection
```go
func verifyLogIntegrity(entries []AuditLogEntry, key []byte) error {
    for _, entry := range entries {
        expected := signEntry(entry, key)
        if !hmac.Equal(entry.HMAC, expected) {
            return fmt.Errorf("tamper detected at %s", entry.Timestamp)
        }
    }
    return nil
}
```

**Key Rotation**: Audit key stored in OS keychain means:
- Key persists independently of master password
- Old logs remain verifiable after password change
- Key tied to vault UUID, not master password
- Users can verify logs without unlocking vault (FR-035)

**Graceful Degradation** (FR-036):
- If OS keychain unavailable (headless, SSH session), disable audit logging with warning
- Vault operations continue normally without audit capability

**Alternatives Considered**:
- Master key derivation: Cannot verify without password, rejected
- Blockchain/Merkle tree: Overkill for local-only tool
- Separate audit key file: Key management complexity, less secure than OS keychain
- Encrypted log entries: Doesn't prevent deletion, adds complexity

---

## Implementation Priorities

Based on research findings, recommended implementation order:

### Phase A: Memory Security Foundation (P1 - Critical)
1. Extract `clearBytes` → public `ClearBytes` in crypto package
2. Change `readPassword()` to return `[]byte` (cmd/helpers.go)
3. Refactor `VaultService.masterPassword` from `string` → `[]byte`
4. Update `crypto.DeriveKey` to accept `[]byte` password
5. Add deferred cleanup handlers to all password-handling functions

**Rationale**: Fixes critical vulnerability (master password exposure in memory)

### Phase B: Cryptographic Hardening (P1 - High)
6. Add `Iterations int` field to `VaultMetadata`
7. Implement backward-compatible iteration count loading
8. Update `DeriveKey` to accept iteration parameter
9. Set new vaults to 600,000 iterations
10. Add migration logic to `ChangePassword`

**Rationale**: Addresses OWASP compliance gap, enables future iteration increases

### Phase C: Password Policy (P2 - Medium)
11. Create `internal/security/password.go` with validation functions
12. Implement complexity checks (FR-011 through FR-015)
13. Add strength calculation for FR-017 indicator
14. Update vault init/change password flows to validate policy
15. Implement CLI and TUI strength indicators

**Rationale**: Prevents weak passwords, user-visible improvement

### Phase D: Audit Logging (P3 - Low)
16. Create `internal/security/audit.go` with HMAC logging
17. Add audit log configuration (default: disabled)
18. Instrument vault operations with audit calls
19. Implement log rotation (FR-024)
20. Add tamper detection verification command

**Rationale**: Defense-in-depth, opt-in feature, lowest priority

---

## Open Questions / Future Work

1. **Testing Strategy**: How to verify memory clearing in unit tests?
   - **Answer**: Use `delve` debugger in integration tests, inspect memory after vault operations
   - **Alternative**: Benchmark memory retention with `pprof` heap snapshots

2. **Cross-Platform Validation**: Does `crypto/subtle` barrier work consistently?
   - **Answer**: Yes - Go stdlib guarantees across Windows/macOS/Linux
   - **Verification**: Include in CI test matrix (already required by Constitution V)

3. **Performance Impact**: Will 600k iterations cause UX issues on older hardware?
   - **Answer**: Spec Assumption 1 accepts longer delays on older hardware
   - **Mitigation**: Configurable via environment variable for edge cases

---

## References

- **OWASP 2023 Password Storage Cheat Sheet**: 600,000 PBKDF2 iterations minimum
- **Go crypto/subtle Documentation**: Compiler barrier guarantees
- **RFC 2104**: HMAC-SHA256 specification
- **Go Memory Model**: https://go.dev/ref/mem (GC limitations documented)

---

**Research Phase Complete** ✅
**Next Step**: Phase 1 - Generate data-model.md and quickstart.md
