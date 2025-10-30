# Security Hardening Data Model

**Feature**: Security Hardening
**Branch**: `005-security-hardening-address`
**Date**: 2025-10-11
**Phase**: 1 (Design & Contracts)

## Overview

This document defines the data structures, validation rules, and state transitions for security hardening entities. All entities support the functional requirements (FR-001 through FR-026) from `spec.md`.

---

## Entity 1: VaultMetadata (Extended)

**Purpose**: Stores vault configuration including cryptographic parameters for backward compatibility and migration.

**Package**: `internal/storage`

**Current Structure** (before this feature):
```go
type VaultMetadata struct {
    Version   int       `json:"version"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Salt      []byte    `json:"salt"`
}
```

**New Structure** (after FR-007):
```go
type VaultMetadata struct {
    Version    int       `json:"version"`     // Vault format version
    CreatedAt  time.Time `json:"created_at"`  // Vault initialization timestamp
    UpdatedAt  time.Time `json:"updated_at"`  // Last modification timestamp
    Salt       []byte    `json:"salt"`        // PBKDF2 salt (32 bytes)
    Iterations int       `json:"iterations"`  // NEW: PBKDF2 iteration count (FR-007)
}
```

### Fields

| Field | Type | Required | Validation | Purpose |
|-------|------|----------|------------|---------|
| `Version` | `int` | Yes | `>= 1` | Vault format version (currently 1) |
| `CreatedAt` | `time.Time` | Yes | Not zero | Vault creation timestamp |
| `UpdatedAt` | `time.Time` | Yes | `>= CreatedAt` | Last save timestamp |
| `Salt` | `[]byte` | Yes | `len == 32` | Cryptographic salt for PBKDF2 |
| `Iterations` | `int` | Yes | `>= 100000` | PBKDF2 iteration count (100k=legacy, 600k=new) |

### Validation Rules

```go
func (m *VaultMetadata) Validate() error {
    if m.Version < 1 {
        return errors.New("invalid version")
    }
    if m.CreatedAt.IsZero() {
        return errors.New("CreatedAt cannot be zero")
    }
    if m.UpdatedAt.Before(m.CreatedAt) {
        return errors.New("UpdatedAt cannot be before CreatedAt")
    }
    if len(m.Salt) != 32 {
        return errors.New("salt must be 32 bytes")
    }
    if m.Iterations < 100000 {
        return errors.New("iterations must be >= 100,000")
    }
    return nil
}
```

### State Transitions (FR-008 Migration)

```
┌─────────────────┐
│  Legacy Vault   │
│ (100k iters)    │
│ Version: 1      │
└────────┬────────┘
         │
         │ User changes password (FR-008)
         │
         ▼
┌─────────────────┐
│  Modern Vault   │
│ (600k iters)    │
│ Version: 1      │
│ UpdatedAt: now  │
└─────────────────┘
```

**Backward Compatibility** (Spec Assumption 2):
```go
// When loading vaults without Iterations field (old format)
if metadata.Iterations == 0 {
    metadata.Iterations = 100000 // Legacy default
}
```

---

## Entity 2: PasswordPolicy

**Purpose**: Defines and validates master password complexity requirements (FR-011 through FR-016).

**Package**: `internal/security` (new package)

**Structure**:
```go
type PasswordPolicy struct {
    MinLength        int  // Minimum character count (FR-011: 12)
    RequireUppercase bool // At least one A-Z (FR-012)
    RequireLowercase bool // At least one a-z (FR-013)
    RequireDigit     bool // At least one 0-9 (FR-014)
    RequireSymbol    bool // At least one punctuation/symbol (FR-015)
}
```

### Default Policy (FR-011 through FR-015)

```go
var DefaultPasswordPolicy = PasswordPolicy{
    MinLength:        12,   // Increased from 8 (FR-011)
    RequireUppercase: true, // FR-012
    RequireLowercase: true, // FR-013
    RequireDigit:     true, // FR-014
    RequireSymbol:    true, // FR-015
}
```

### Methods

#### `Validate(password []byte) error` (FR-011-016)

Validates password against policy and returns descriptive error (FR-016).

```go
func (p *PasswordPolicy) Validate(password []byte) error {
    if len(password) < p.MinLength {
        return fmt.Errorf("password must be at least %d characters", p.MinLength)
    }

    var hasUpper, hasLower, hasDigit, hasSymbol bool
    for _, ch := range password {
        switch {
        case unicode.IsUpper(rune(ch)):
            hasUpper = true
        case unicode.IsLower(rune(ch)):
            hasLower = true
        case unicode.IsDigit(rune(ch)):
            hasDigit = true
        case unicode.IsPunct(rune(ch)) || unicode.IsSymbol(rune(ch)):
            hasSymbol = true
        }
    }

    missing := []string{}
    if p.RequireUppercase && !hasUpper {
        missing = append(missing, "uppercase letter")
    }
    if p.RequireLowercase && !hasLower {
        missing = append(missing, "lowercase letter")
    }
    if p.RequireDigit && !hasDigit {
        missing = append(missing, "digit")
    }
    if p.RequireSymbol && !hasSymbol {
        missing = append(missing, "symbol")
    }

    if len(missing) > 0 {
        return fmt.Errorf("password must contain: %s", strings.Join(missing, ", "))
    }

    return nil
}
```

#### `Strength(password []byte) string` (FR-017)

Calculates password strength for real-time indicator.

```go
func (p *PasswordPolicy) Strength(password []byte) string {
    if len(password) == 0 {
        return "empty"
    }

    score := 0

    // Length scoring
    if len(password) >= p.MinLength {
        score++
    }
    if len(password) >= 16 {
        score++
    }

    // Complexity scoring
    var hasUpper, hasLower, hasDigit, hasSymbol bool
    for _, ch := range password {
        switch {
        case unicode.IsUpper(rune(ch)):
            hasUpper = true
        case unicode.IsLower(rune(ch)):
            hasLower = true
        case unicode.IsDigit(rune(ch)):
            hasDigit = true
        case unicode.IsPunct(rune(ch)) || unicode.IsSymbol(rune(ch)):
            hasSymbol = true
        }
    }

    complexityCount := 0
    if hasUpper { complexityCount++ }
    if hasLower { complexityCount++ }
    if hasDigit { complexityCount++ }
    if hasSymbol { complexityCount++ }

    if complexityCount >= 4 {
        score += 2
    } else if complexityCount >= 3 {
        score++
    }

    // Strength mapping
    switch {
    case score >= 4:
        return "strong"
    case score >= 2:
        return "medium"
    default:
        return "weak"
    }
}
```

### Unicode Support (Spec Assumption 5)

- Unicode letters (accented characters) count as letters
- Unicode symbols count as symbols
- Supports international users

---

## Entity 3: AuditLogEntry

**Purpose**: Represents a single security event with tamper-evident HMAC signature (FR-019 through FR-026).

**Package**: `internal/security` (new package)

**Structure**:
```go
type AuditLogEntry struct {
    Timestamp      time.Time `json:"timestamp"`       // Event time (FR-019, FR-020)
    EventType      string    `json:"event_type"`      // Type of operation (see enum below)
    Outcome        string    `json:"outcome"`         // "success" or "failure"
    CredentialName string    `json:"credential_name"` // Service name (NOT password, FR-021)
    HMACSignature  []byte    `json:"hmac_signature"`  // Tamper detection (FR-022)
}
```

### Event Type Enum

```go
const (
    EventVaultUnlock        = "vault_unlock"        // FR-019
    EventVaultLock          = "vault_lock"          // FR-019
    EventVaultPasswordChange = "vault_password_change" // FR-019
    EventCredentialAccess   = "credential_access"   // FR-020 (get)
    EventCredentialAdd      = "credential_add"      // FR-020
    EventCredentialUpdate   = "credential_update"   // FR-020
    EventCredentialDelete   = "credential_delete"   // FR-020
)
```

### Outcome Enum

```go
const (
    OutcomeSuccess = "success"
    OutcomeFailure = "failure"
)
```

### HMAC Signature Calculation (FR-022)

```go
func (e *AuditLogEntry) Sign(key []byte) error {
    // Canonical serialization (order matters!)
    data := fmt.Sprintf("%s|%s|%s|%s",
        e.Timestamp.Format(time.RFC3339Nano),
        e.EventType,
        e.Outcome,
        e.CredentialName,
    )

    mac := hmac.New(sha256.New, key)
    mac.Write([]byte(data))
    e.HMACSignature = mac.Sum(nil)

    return nil
}

func (e *AuditLogEntry) Verify(key []byte) error {
    // Recalculate HMAC
    data := fmt.Sprintf("%s|%s|%s|%s",
        e.Timestamp.Format(time.RFC3339Nano),
        e.EventType,
        e.Outcome,
        e.CredentialName,
    )

    mac := hmac.New(sha256.New, key)
    mac.Write([]byte(data))
    expected := mac.Sum(nil)

    // Constant-time comparison to prevent timing attacks
    if !hmac.Equal(e.HMACSignature, expected) {
        return fmt.Errorf("HMAC verification failed at %s", e.Timestamp)
    }

    return nil
}
```

### Log Rotation (FR-024)

```go
type AuditLogger struct {
    filePath       string
    maxSizeBytes   int64 // Default: 10MB (FR-024)
    currentSize    int64
    auditKey       []byte // Derived from vault master key
}

func (l *AuditLogger) ShouldRotate() bool {
    return l.currentSize >= l.maxSizeBytes
}

func (l *AuditLogger) Rotate() error {
    // Rename current log to .old
    // Create new empty log
    // Reset size counter
}
```

### Privacy Constraint (FR-021)

**FORBIDDEN**: Logging credential passwords or sensitive values

```go
// ❌ NEVER DO THIS
entry := AuditLogEntry{
    EventType: EventCredentialAccess,
    CredentialName: cred.Service,
    // PASSWORD: cred.Password, // FORBIDDEN by FR-021!
}

// ✅ CORRECT
entry := AuditLogEntry{
    EventType: EventCredentialAccess,
    CredentialName: cred.Service, // Service name only
}
```

---

## Entity 4: SecureBytes (Helper Type)

**Purpose**: Type-safe wrapper for sensitive byte arrays with automatic deferred cleanup (FR-003).

**Package**: `internal/security` (new package)

**Structure**:
```go
type SecureBytes struct {
    data []byte
}

func NewSecureBytes(data []byte) *SecureBytes {
    return &SecureBytes{data: data}
}

// Clear zeroes the underlying byte array
func (sb *SecureBytes) Clear() {
    if sb.data != nil {
        ClearBytes(sb.data)
        sb.data = nil
    }
}

// Get returns the underlying bytes (use with caution)
func (sb *SecureBytes) Get() []byte {
    return sb.data
}

// Len returns the length without exposing data
func (sb *SecureBytes) Len() int {
    return len(sb.data)
}
```

### Usage Pattern (FR-003 Deferred Cleanup)

```go
func processPassword(input []byte) error {
    password := NewSecureBytes(input)
    defer password.Clear() // Ensures clearing even on panic

    // Use password.Get() for operations
    key, err := crypto.DeriveKey(password.Get(), salt, iterations)
    if err != nil {
        return err // defer still executes
    }

    return nil
}
```

---

## Internal API Changes

These signatures change for FR-001/FR-005 (byte array handling):

### crypto package

```go
// OLD
func (c *CryptoService) DeriveKey(password string, salt []byte) ([]byte, error)

// NEW
func (c *CryptoService) DeriveKey(password []byte, salt []byte, iterations int) ([]byte, error)
func (c *CryptoService) ClearBytes(data []byte) // Exposed from private clearBytes()
```

### vault package

```go
// OLD
func (v *VaultService) Initialize(masterPassword string, useKeychain bool) error
func (v *VaultService) Unlock(masterPassword string) error
func (v *VaultService) ChangePassword(newPassword string) error

// NEW
func (v *VaultService) Initialize(masterPassword []byte, useKeychain bool) error
func (v *VaultService) Unlock(masterPassword []byte) error
func (v *VaultService) ChangePassword(newPassword []byte) error

// Internal field change
type VaultService struct {
    // OLD: masterPassword string
    masterPassword []byte // NEW: Secure clearable type
}
```

### cmd package

```go
// OLD (cmd/helpers.go)
func readPassword() (string, error)

// NEW
func readPassword() ([]byte, error) // Removes string() conversion from gopass
```

---

## Relationships

```
┌─────────────────┐
│  VaultMetadata  │──┐
│ (storage)       │  │ Stored in vault file JSON
└─────────────────┘  │
                     │
┌─────────────────┐  │
│ VaultService    │◄─┘
│ (vault)         │
│                 │
│ masterPassword  │──► Used with PasswordPolicy for validation
│   ([]byte)      │
└────────┬────────┘
         │
         │ Logs operations
         ▼
┌─────────────────┐
│ AuditLogEntry   │
│ (security)      │
└─────────────────┘

┌─────────────────┐
│ PasswordPolicy  │──► Validates passwords
│ (security)      │    before vault operations
└─────────────────┘

┌─────────────────┐
│  SecureBytes    │──► Wraps []byte with
│ (security)      │    deferred cleanup
└─────────────────┘
```

---

## Testing Considerations

### VaultMetadata
- Validate backward compatibility loading (Iterations=0 → 100000)
- Test migration on password change (100k → 600k)
- Verify metadata validation catches invalid values

### PasswordPolicy
- Unit tests for all complexity combinations (FR-011 through FR-015)
- Test Unicode character handling (Assumption 5)
- Verify error messages list missing requirements (FR-016)
- Test strength calculation for weak/medium/strong boundaries

### AuditLogEntry
- Verify HMAC detects tampering (modify any field, verify fails)
- Test log rotation at size threshold (FR-024)
- Ensure NO credential passwords logged (FR-021 test)

### SecureBytes
- Verify memory is zeroed after Clear()
- Test deferred cleanup during panics (FR-003)

---

**Data Model Complete** ✅
**Next**: Generate quickstart.md
