# Contract: Key Wrapping API

**Package**: `internal/crypto`
**File**: `keywrap.go`
**Date**: 2025-12-05

## Overview

This contract defines the key wrapping interface for the recovery key integration feature. Key wrapping encrypts a Data Encryption Key (DEK) with a Key Encryption Key (KEK) using AES-256-GCM.

## Types

### WrappedKey

```go
// WrappedKey represents an AES-256-GCM encrypted key
type WrappedKey struct {
    Ciphertext []byte // 48 bytes: 32-byte key + 16-byte GCM auth tag
    Nonce      []byte // 12 bytes: GCM nonce (must be unique per wrap)
}
```

### KeyWrapResult

```go
// KeyWrapResult contains a DEK wrapped by multiple KEKs
type KeyWrapResult struct {
    DEK              []byte     // 32-byte plaintext DEK (MUST be cleared after use)
    PasswordWrapped  WrappedKey // DEK wrapped with password-derived KEK
    RecoveryWrapped  WrappedKey // DEK wrapped with recovery-derived KEK
}
```

## Functions

### GenerateDEK

```go
// GenerateDEK generates a cryptographically secure 256-bit Data Encryption Key.
//
// Returns:
//   - dek: 32-byte random key from crypto/rand
//   - err: error if random generation fails
//
// Security:
//   - Caller MUST clear returned DEK with crypto.ClearBytes() after use
//   - DEK MUST NOT be logged or written to disk in plaintext
func GenerateDEK() (dek []byte, err error)
```

**Preconditions**: None
**Postconditions**: Returns 32-byte cryptographically random key
**Errors**: `ErrRandomGenerationFailed` if crypto/rand fails

---

### WrapKey

```go
// WrapKey encrypts a DEK with a Key Encryption Key using AES-256-GCM.
//
// Parameters:
//   - dek: 32-byte Data Encryption Key to wrap
//   - kek: 32-byte Key Encryption Key (from password or recovery derivation)
//
// Returns:
//   - wrapped: WrappedKey containing ciphertext and nonce
//   - err: error if encryption fails
//
// Security:
//   - Each call generates a unique 12-byte nonce via crypto/rand
//   - Auth tag provides integrity verification on unwrap
func WrapKey(dek, kek []byte) (wrapped WrappedKey, err error)
```

**Preconditions**:
- `dek` MUST be exactly 32 bytes
- `kek` MUST be exactly 32 bytes

**Postconditions**:
- `wrapped.Ciphertext` is 48 bytes (32 + 16 tag)
- `wrapped.Nonce` is 12 bytes, unique

**Errors**:
- `ErrInvalidKeyLength` if dek or kek is wrong size
- `ErrEncryptionFailed` if AES-GCM fails

---

### UnwrapKey

```go
// UnwrapKey decrypts a wrapped DEK using a Key Encryption Key.
//
// Parameters:
//   - wrapped: WrappedKey from WrapKey()
//   - kek: 32-byte Key Encryption Key (must match the one used to wrap)
//
// Returns:
//   - dek: 32-byte plaintext DEK
//   - err: error if decryption or authentication fails
//
// Security:
//   - Caller MUST clear returned DEK with crypto.ClearBytes() after use
//   - Returns error (not corrupted data) if auth tag verification fails
func UnwrapKey(wrapped WrappedKey, kek []byte) (dek []byte, err error)
```

**Preconditions**:
- `wrapped.Ciphertext` MUST be 48 bytes
- `wrapped.Nonce` MUST be 12 bytes
- `kek` MUST be exactly 32 bytes

**Postconditions**:
- Returns 32-byte DEK if kek matches original wrapper

**Errors**:
- `ErrInvalidKeyLength` if kek is wrong size
- `ErrInvalidCiphertext` if wrapped data is wrong size
- `ErrDecryptionFailed` if auth tag verification fails (wrong KEK or corrupted)

---

### GenerateAndWrapDEK

```go
// GenerateAndWrapDEK creates a new DEK and wraps it with both password and recovery KEKs.
//
// Parameters:
//   - passwordKEK: 32-byte key derived from master password via PBKDF2
//   - recoveryKEK: 32-byte key derived from recovery phrase via Argon2id
//
// Returns:
//   - result: KeyWrapResult containing DEK and both wrapped versions
//   - err: error if generation or wrapping fails
//
// Security:
//   - Caller MUST clear result.DEK with crypto.ClearBytes() after encrypting vault
//   - Both wrapped versions use independent nonces
func GenerateAndWrapDEK(passwordKEK, recoveryKEK []byte) (result KeyWrapResult, err error)
```

**Preconditions**:
- `passwordKEK` MUST be exactly 32 bytes
- `recoveryKEK` MUST be exactly 32 bytes

**Postconditions**:
- `result.DEK` is 32 bytes
- `result.PasswordWrapped` and `result.RecoveryWrapped` have different nonces

**Errors**:
- `ErrInvalidKeyLength` if KEKs are wrong size
- Propagates errors from `GenerateDEK()` and `WrapKey()`

## Error Types

```go
var (
    ErrInvalidKeyLength       = errors.New("invalid key length: expected 32 bytes")
    ErrInvalidCiphertext      = errors.New("invalid ciphertext length")
    ErrDecryptionFailed       = errors.New("key unwrap failed: invalid KEK or corrupted data")
    ErrRandomGenerationFailed = errors.New("failed to generate random bytes")
    ErrEncryptionFailed       = errors.New("key wrap encryption failed")
)
```

## Usage Examples

### Vault Initialization

```go
// During vault init with recovery enabled
passwordKEK, _ := crypto.DeriveKey(password, salt, iterations)
defer crypto.ClearBytes(passwordKEK)

recoveryKEK := recovery.DeriveRecoveryKey(mnemonic, passphrase, kdfParams)
defer crypto.ClearBytes(recoveryKEK)

result, err := crypto.GenerateAndWrapDEK(passwordKEK, recoveryKEK)
if err != nil {
    return err
}
defer crypto.ClearBytes(result.DEK)

// Encrypt vault with DEK
encryptedVault, _ := crypto.Encrypt(vaultData, result.DEK)

// Store wrapped keys in metadata
metadata.WrappedDEK = result.PasswordWrapped.Ciphertext
metadata.WrappedDEKNonce = result.PasswordWrapped.Nonce
recoveryMeta.EncryptedRecoveryKey = result.RecoveryWrapped.Ciphertext
recoveryMeta.NonceRecovery = result.RecoveryWrapped.Nonce
```

### Password Unlock

```go
// During vault unlock with password
passwordKEK, _ := crypto.DeriveKey(password, metadata.Salt, metadata.Iterations)
defer crypto.ClearBytes(passwordKEK)

wrapped := crypto.WrappedKey{
    Ciphertext: metadata.WrappedDEK,
    Nonce:      metadata.WrappedDEKNonce,
}

dek, err := crypto.UnwrapKey(wrapped, passwordKEK)
if err != nil {
    return ErrInvalidPassword
}
defer crypto.ClearBytes(dek)

// Decrypt vault with DEK
vaultData, _ := crypto.Decrypt(encryptedVault, dek)
```

### Recovery Unlock

```go
// During recovery flow
recoveryKEK := recovery.DeriveRecoveryKey(reconstructedMnemonic, passphrase, kdfParams)
defer crypto.ClearBytes(recoveryKEK)

wrapped := crypto.WrappedKey{
    Ciphertext: recoveryMeta.EncryptedRecoveryKey,
    Nonce:      recoveryMeta.NonceRecovery,
}

dek, err := crypto.UnwrapKey(wrapped, recoveryKEK)
if err != nil {
    return ErrInvalidRecoveryPhrase
}
defer crypto.ClearBytes(dek)

// Decrypt vault, prompt for new password, re-wrap DEK
// ...
```

## Security Considerations

1. **Memory Safety**: All DEK and KEK variables MUST use `defer crypto.ClearBytes()` immediately after allocation.

2. **Nonce Uniqueness**: Each `WrapKey()` call generates a fresh nonce. Never reuse nonces.

3. **Error Messages**: Do not differentiate between "wrong key" and "corrupted data" to avoid oracle attacks.

4. **Timing**: Use constant-time comparison where applicable (handled by GCM auth tag verification).
