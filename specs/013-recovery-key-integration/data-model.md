# Data Model: Recovery Key Integration

**Feature**: 013-recovery-key-integration
**Date**: 2025-12-05

## Overview

This document defines the data structures modified or introduced for key wrapping support. The core change is introducing a Data Encryption Key (DEK) that encrypts vault data, with the DEK itself wrapped by multiple Key Encryption Keys (KEKs).

## Entity Changes

### 1. VaultMetadata (Modified)

**Location**: `internal/storage/storage.go`

**Current Structure**:
```go
type VaultMetadata struct {
    Version    int       `json:"version"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
    Salt       []byte    `json:"salt"`
    Iterations int       `json:"iterations"`
}
```

**New Structure** (Version 2):
```go
type VaultMetadata struct {
    Version         int       `json:"version"`          // 1 = legacy, 2 = key-wrapped
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    Salt            []byte    `json:"salt"`             // Salt for password KDF
    Iterations      int       `json:"iterations"`       // PBKDF2 iterations
    WrappedDEK      []byte    `json:"wrapped_dek,omitempty"`       // NEW: DEK wrapped by password KEK
    WrappedDEKNonce []byte    `json:"wrapped_dek_nonce,omitempty"` // NEW: GCM nonce for wrap
}
```

**Field Descriptions**:
| Field | Type | Description | Version |
|-------|------|-------------|---------|
| Version | int | Format version (1=legacy, 2=key-wrapped) | 1+ |
| CreatedAt | time.Time | Vault creation timestamp | 1+ |
| UpdatedAt | time.Time | Last modification timestamp | 1+ |
| Salt | []byte | 32-byte salt for PBKDF2 | 1+ |
| Iterations | int | PBKDF2 iteration count (600000) | 1+ |
| WrappedDEK | []byte | DEK encrypted with password-derived KEK | 2+ |
| WrappedDEKNonce | []byte | 12-byte GCM nonce for DEK wrapping | 2+ |

**Validation Rules**:
- Version MUST be 1 or 2
- If Version = 2, WrappedDEK MUST be present (32 bytes encrypted + 16 byte tag = 48 bytes)
- If Version = 2, WrappedDEKNonce MUST be 12 bytes
- Salt MUST be 32 bytes
- Iterations MUST be >= 600000

---

### 2. RecoveryMetadata (Modified)

**Location**: `internal/vault/metadata.go`

**Current Structure**:
```go
type RecoveryMetadata struct {
    Enabled              bool      `json:"enabled"`
    Version              string    `json:"version"`
    PassphraseRequired   bool      `json:"passphrase_required"`
    ChallengePositions   []int     `json:"challenge_positions"`
    KDFParams            KDFParams `json:"kdf_params"`
    EncryptedStoredWords []byte    `json:"encrypted_stored_words"`
    NonceStored          []byte    `json:"nonce_stored"`
    EncryptedRecoveryKey []byte    `json:"encrypted_recovery_key"` // Currently: random key
    NonceRecovery        []byte    `json:"nonce_recovery"`
}
```

**Modified Usage** (same structure, different semantics):
```go
type RecoveryMetadata struct {
    Enabled              bool      `json:"enabled"`
    Version              string    `json:"version"`               // "1" = old (broken), "2" = key-wrapped
    PassphraseRequired   bool      `json:"passphrase_required"`
    ChallengePositions   []int     `json:"challenge_positions"`
    KDFParams            KDFParams `json:"kdf_params"`
    EncryptedStoredWords []byte    `json:"encrypted_stored_words"`
    NonceStored          []byte    `json:"nonce_stored"`
    EncryptedRecoveryKey []byte    `json:"encrypted_recovery_key"` // REPURPOSED: DEK wrapped by recovery KEK
    NonceRecovery        []byte    `json:"nonce_recovery"`
}
```

**Semantic Change**:
| Version | EncryptedRecoveryKey Contains |
|---------|------------------------------|
| "1" | Random 32-byte VaultRecoveryKey (broken - not used for vault encryption) |
| "2" | DEK wrapped by recovery-derived KEK (working) |

**Validation Rules**:
- If Enabled = true and Version = "2", EncryptedRecoveryKey MUST be 48 bytes (32 + 16 tag)
- NonceRecovery MUST be 12 bytes
- ChallengePositions MUST have exactly 6 elements, each in [0, 23]

---

### 3. WrappedKey (New)

**Location**: `internal/crypto/keywrap.go`

**Structure**:
```go
// WrappedKey represents a key encrypted with AES-256-GCM
type WrappedKey struct {
    Ciphertext []byte // 48 bytes: 32-byte key + 16-byte auth tag
    Nonce      []byte // 12 bytes: GCM nonce
}
```

**Usage**: Returned by `WrapKey()`, input to `UnwrapKey()`.

---

### 4. KeyWrapResult (New)

**Location**: `internal/crypto/keywrap.go`

**Structure**:
```go
// KeyWrapResult contains both wrapped versions of a DEK
type KeyWrapResult struct {
    DEK              []byte     // 32-byte Data Encryption Key (clear immediately after use!)
    PasswordWrapped  WrappedKey // DEK wrapped with password-derived KEK
    RecoveryWrapped  WrappedKey // DEK wrapped with recovery-derived KEK
}
```

**Usage**: Returned by `GenerateAndWrapDEK()` during vault initialization.

---

## State Transitions

### Vault Version Migration

```
┌─────────────────┐     User accepts     ┌─────────────────┐
│  Version 1      │ ─────migration─────> │  Version 2      │
│  (legacy)       │                      │  (key-wrapped)  │
└─────────────────┘                      └─────────────────┘
        │                                        │
        │ Password unlock                        │ Password unlock OR
        │ (direct decrypt)                       │ Recovery unlock
        ▼                                        ▼
   ┌─────────┐                             ┌─────────┐
   │ Unlocked │                             │ Unlocked │
   └─────────┘                             └─────────┘
```

### DEK Lifecycle States

```
┌──────────────┐
│ Not Created  │ (vault doesn't exist)
└──────┬───────┘
       │ vault init
       ▼
┌──────────────┐
│  Generated   │ (in memory, 32 bytes random)
└──────┬───────┘
       │ wrap with KEKs
       ▼
┌──────────────────────┐
│  Wrapped & Stored    │ (password + recovery wrapped versions on disk)
└──────┬───────────────┘
       │ vault unlock
       ▼
┌──────────────┐
│  Unwrapped   │ (in memory, decrypt vault)
└──────┬───────┘
       │ crypto.ClearBytes()
       ▼
┌──────────────┐
│   Cleared    │ (zeroed from memory)
└──────────────┘
```

## Backward Compatibility

### Version 1 Vaults (Legacy)
- Continue to work with password-only unlock
- Recovery feature non-functional (existing behavior)
- Migration available but not required

### Version 2 Vaults (New)
- Full recovery support
- Password unlock unchanged from user perspective
- New vaults default to version 2 when recovery enabled
- `--no-recovery` creates version 1 vault (no key wrapping overhead)

## File Format Examples

### Version 2 Vault File (`vault.enc`)
```json
{
  "metadata": {
    "version": 2,
    "created_at": "2025-12-05T10:00:00Z",
    "updated_at": "2025-12-05T10:00:00Z",
    "salt": "base64-encoded-32-bytes",
    "iterations": 600000,
    "wrapped_dek": "base64-encoded-48-bytes",
    "wrapped_dek_nonce": "base64-encoded-12-bytes"
  },
  "data": "base64-encoded-encrypted-vault-data"
}
```

### Version 2 Metadata File (`.meta.json`)
```json
{
  "version": "1.0",
  "keychain_enabled": false,
  "audit_enabled": true,
  "recovery": {
    "enabled": true,
    "version": "2",
    "passphrase_required": false,
    "challenge_positions": [2, 7, 11, 15, 19, 22],
    "kdf_params": {
      "algorithm": "argon2id",
      "time": 1,
      "memory": 65536,
      "threads": 4,
      "salt_challenge": "base64-encoded-32-bytes",
      "salt_recovery": "base64-encoded-32-bytes"
    },
    "encrypted_stored_words": "base64-encoded-encrypted-words",
    "nonce_stored": "base64-encoded-12-bytes",
    "encrypted_recovery_key": "base64-encoded-48-bytes",
    "nonce_recovery": "base64-encoded-12-bytes"
  }
}
```
