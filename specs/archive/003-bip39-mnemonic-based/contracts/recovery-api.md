# API Contract: internal/recovery Package

**Package**: `github.com/yourusername/pass-cli/internal/recovery`
**Version**: 1.0.0
**Status**: Draft
**Date**: 2025-01-13

## Overview

Public API contract for the `internal/recovery` package. Defines exported functions, types, and behaviors for BIP39 mnemonic-based vault recovery.

**Design Principles**:
- Library-first architecture (Constitution Principle II)
- No CLI dependencies (accepts/returns Go types only)
- Security-first (memory clearing, no logging)
- Testable (deterministic given inputs)

---

## Public Types

### SetupConfig

Configuration for recovery setup during vault initialization.

```go
type SetupConfig struct {
    // Optional passphrase (25th word). Empty = no passphrase.
    Passphrase []byte

    // Custom Argon2 parameters (optional, uses defaults if nil)
    KDFParams *KDFParams
}
```

**Fields**:
- `Passphrase`: Optional BIP39 passphrase. Caller must clear after use.
- `KDFParams`: Custom KDF settings (nil = use defaults from research.md)

---

### SetupResult

Result of recovery setup operation.

```go
type SetupResult struct {
    // 24-word mnemonic to display to user
    Mnemonic string

    // Recovery metadata to store in vault
    Metadata *RecoveryMetadata

    // Vault recovery key (32 bytes) to integrate with vault unlocking
    VaultRecoveryKey []byte
}
```

**Fields**:
- `Mnemonic`: Space-separated 24-word phrase. **Caller must clear** (convert to []byte and call `crypto.ClearBytes`).
- `Metadata`: Recovery metadata structure. Store in `VaultMetadata.Recovery`.
- `VaultRecoveryKey`: 32-byte key that unlocks vault. **Caller must clear** with `crypto.ClearBytes`.

**Lifecycle**: All sensitive fields must be cleared by caller after use.

---

### RecoveryConfig

Configuration for recovery execution.

```go
type RecoveryConfig struct {
    // Words entered by user (from challenge positions)
    ChallengeWords []string

    // Passphrase (if required). Empty = no passphrase.
    Passphrase []byte

    // Recovery metadata from vault
    Metadata *RecoveryMetadata
}
```

**Fields**:
- `ChallengeWords`: 6 words provided by user (in any order). Must match positions in `Metadata.ChallengePositions`.
- `Passphrase`: Optional passphrase. Caller must clear.
- `Metadata`: Recovery metadata loaded from vault.

**Validation** (performed by `PerformRecovery`):
- `len(ChallengeWords) == 6`
- All words in BIP39 wordlist
- `Metadata != nil && Metadata.Enabled == true`

---

### VerifyConfig

Configuration for backup verification during init.

```go
type VerifyConfig struct {
    // Full 24-word mnemonic (generated during setup)
    Mnemonic string

    // Positions to prompt for (randomly selected)
    VerifyPositions []int

    // Words entered by user
    UserWords []string
}
```

**Fields**:
- `Mnemonic`: Full mnemonic from `SetupResult`
- `VerifyPositions`: 3 random positions [0-23] to verify
- `UserWords`: Words entered by user (must match positions)

---

## Public Functions

### 1. SetupRecovery

Generates BIP39 mnemonic and prepares recovery metadata.

**Signature**:
```go
func SetupRecovery(config *SetupConfig) (*SetupResult, error)
```

**Parameters**:
- `config`: Setup configuration (passphrase, KDF params)

**Returns**:
- `*SetupResult`: Mnemonic, metadata, and vault recovery key
- `error`: Error if entropy generation, mnemonic generation, or encryption fails

**Behavior**:
1. Generate 256-bit entropy (`crypto/rand`)
2. Convert to 24-word BIP39 mnemonic
3. Randomly select 6 challenge positions [0-23]
4. Split mnemonic into challenge words (6) + stored words (18)
5. Derive challenge key from 6-word mnemonic + passphrase (Argon2id)
6. Encrypt stored words (18) with challenge key (AES-256-GCM)
7. Generate vault recovery key (32 random bytes)
8. Derive recovery key from full 24-word mnemonic + passphrase (Argon2id)
9. Encrypt vault recovery key with recovery key (AES-256-GCM)
10. Build `RecoveryMetadata` with all encrypted data
11. Return `SetupResult`

**Error Conditions**:
- Entropy generation failure → `ErrEntropyGeneration`
- Mnemonic generation failure → `ErrMnemonicGeneration`
- Encryption failure → `ErrEncryptionFailed`

**Security**:
- Clears entropy, seeds, and keys internally
- **Caller must clear** `SetupResult.Mnemonic` and `SetupResult.VaultRecoveryKey`

**Example**:
```go
config := &recovery.SetupConfig{
    Passphrase: []byte("my-secret-phrase"),
}
defer crypto.ClearBytes(config.Passphrase)

result, err := recovery.SetupRecovery(config)
if err != nil {
    return err
}
defer crypto.ClearBytes([]byte(result.Mnemonic))
defer crypto.ClearBytes(result.VaultRecoveryKey)

// Display result.Mnemonic to user
fmt.Println("Write down these 24 words:")
fmt.Println(result.Mnemonic)

// Store result.Metadata in vault
vaultMetadata.Recovery = result.Metadata

// Use result.VaultRecoveryKey to encrypt vault
vault.EncryptWithKey(result.VaultRecoveryKey, credentials)
```

---

### 2. PerformRecovery

Recovers vault access using challenge words.

**Signature**:
```go
func PerformRecovery(config *RecoveryConfig) ([]byte, error)
```

**Parameters**:
- `config`: Recovery configuration (challenge words, passphrase, metadata)

**Returns**:
- `[]byte`: Vault recovery key (32 bytes). **Caller must clear**.
- `error`: Error if validation, decryption, or reconstruction fails

**Behavior**:
1. Validate inputs (6 words, all in BIP39 wordlist, metadata enabled)
2. Map challenge words to correct positions (using `Metadata.ChallengePositions`)
3. Derive challenge key from 6-word mnemonic + passphrase (Argon2id)
4. Decrypt stored words (18) with challenge key (AES-256-GCM)
5. Combine challenge words (6) + stored words (18) → full 24-word mnemonic
6. Validate full mnemonic checksum (`bip39.IsMnemonicValid`)
7. Derive recovery key from full 24-word mnemonic + passphrase (Argon2id)
8. Decrypt vault recovery key with recovery key (AES-256-GCM)
9. Return vault recovery key

**Error Conditions**:
- Invalid word (not in BIP39 wordlist) → `ErrInvalidWord`
- Wrong challenge words (decryption fails) → `ErrDecryptionFailed`
- Invalid checksum (reconstructed mnemonic) → `ErrInvalidMnemonic`
- Wrong passphrase (decryption fails) → `ErrDecryptionFailed`
- Recovery disabled (`Metadata.Enabled == false`) → `ErrRecoveryDisabled`

**Security**:
- Clears seeds, keys, and reconstructed mnemonic internally
- **Caller must clear** returned vault recovery key

**Example**:
```go
config := &recovery.RecoveryConfig{
    ChallengeWords: []string{"device", "identify", "spin", "diary", "about", "hybrid"},
    Passphrase:     []byte("my-secret-phrase"),
    Metadata:       vaultMetadata.Recovery,
}
defer crypto.ClearBytes(config.Passphrase)

vaultKey, err := recovery.PerformRecovery(config)
if err != nil {
    if errors.Is(err, recovery.ErrInvalidWord) {
        fmt.Println("Invalid word entered")
    } else if errors.Is(err, recovery.ErrDecryptionFailed) {
        fmt.Println("Wrong words or passphrase")
    }
    return err
}
defer crypto.ClearBytes(vaultKey)

// Unlock vault with vaultKey
vault.UnlockWithKey(vaultKey)
```

---

### 3. VerifyBackup

Verifies user wrote down mnemonic correctly.

**Signature**:
```go
func VerifyBackup(config *VerifyConfig) error
```

**Parameters**:
- `config`: Verification configuration (mnemonic, positions, user words)

**Returns**:
- `error`: `nil` if verification passes, `ErrVerificationFailed` if words mismatch

**Behavior**:
1. Extract words at `VerifyPositions` from `Mnemonic`
2. Compare with `UserWords`
3. Return error if any mismatch

**Error Conditions**:
- Word mismatch → `ErrVerificationFailed`
- Invalid positions (out of range [0-23]) → `ErrInvalidPositions`

**Security**:
- No sensitive data logged
- Does not modify mnemonic

**Example**:
```go
positions := []int{7, 18, 22} // Random selection
config := &recovery.VerifyConfig{
    Mnemonic:        mnemonic, // From SetupResult
    VerifyPositions: positions,
    UserWords:       []string{"device", "identify", "spin"}, // User input
}

if err := recovery.VerifyBackup(config); err != nil {
    fmt.Println("Verification failed. Please check your backup.")
    return err
}

fmt.Println("Verification successful!")
```

---

### 4. SelectVerifyPositions

Randomly selects positions for backup verification.

**Signature**:
```go
func SelectVerifyPositions(count int) ([]int, error)
```

**Parameters**:
- `count`: Number of positions to select (e.g., 3)

**Returns**:
- `[]int`: Random positions [0-23], sorted
- `error`: Error if count > 24 or crypto/rand fails

**Behavior**:
1. Generate `count` unique random positions [0, 23]
2. Sort and return

**Error Conditions**:
- `count > 24` → `ErrInvalidCount`
- `crypto/rand` failure → `ErrRandomGeneration`

**Example**:
```go
positions, err := recovery.SelectVerifyPositions(3)
// positions might be [7, 18, 22]
```

---

### 5. ShuffleChallengePositions

Randomizes order of challenge positions for recovery prompts.

**Signature**:
```go
func ShuffleChallengePositions(positions []int) []int
```

**Parameters**:
- `positions`: Fixed challenge positions from metadata

**Returns**:
- `[]int`: Shuffled positions (non-destructive, creates copy)

**Behavior**:
1. Copy positions
2. Shuffle using `math/rand.Shuffle` (non-crypto random, UX only)
3. Return shuffled copy

**Security**:
- Not cryptographically secure (uses `math/rand`)
- Acceptable: Only affects prompt order, not security

**Example**:
```go
fixed := []int{3, 7, 11, 15, 18, 22}
shuffled := recovery.ShuffleChallengePositions(fixed)
// shuffled might be [18, 3, 22, 7, 11, 15]
```

---

### 6. ValidateWord

Checks if a word is in the BIP39 English wordlist.

**Signature**:
```go
func ValidateWord(word string) bool
```

**Parameters**:
- `word`: Word to validate (case-insensitive)

**Returns**:
- `bool`: `true` if word in wordlist, `false` otherwise

**Behavior**:
1. Convert word to lowercase
2. Check against BIP39 English wordlist

**Example**:
```go
if !recovery.ValidateWord("device") {
    fmt.Println("Invalid word")
}
```

---

## Error Types

### Sentinel Errors

```go
var (
    ErrEntropyGeneration    = errors.New("failed to generate entropy")
    ErrMnemonicGeneration   = errors.New("failed to generate mnemonic")
    ErrInvalidWord          = errors.New("word not in BIP39 wordlist")
    ErrInvalidMnemonic      = errors.New("invalid mnemonic (checksum mismatch)")
    ErrDecryptionFailed     = errors.New("decryption failed")
    ErrVerificationFailed   = errors.New("backup verification failed")
    ErrRecoveryDisabled     = errors.New("recovery not enabled")
    ErrInvalidPositions     = errors.New("invalid challenge positions")
    ErrInvalidCount         = errors.New("invalid position count")
    ErrRandomGeneration     = errors.New("random number generation failed")
    ErrEncryptionFailed     = errors.New("encryption failed")
)
```

**Usage**: Use `errors.Is()` for error checking.

---

## Constants

### Default KDF Parameters

```go
const (
    DefaultTime      uint32 = 1
    DefaultMemory    uint32 = 65536 // 64 MB
    DefaultThreads   uint8  = 4
    DefaultKeyLen    uint32 = 32    // AES-256
    DefaultSaltLen   int    = 32    // 256 bits
)
```

### BIP39 Constants

```go
const (
    EntropyBits     int = 256  // 24-word mnemonic
    MnemonicWords   int = 24
    WordlistSize    int = 2048
    ChallengeCount  int = 6    // Fixed: 6-word challenge
    VerifyCount     int = 3    // Default: 3-word verification
)
```

---

## Usage Example (Full Flow)

### Initialization

```go
package main

import (
    "fmt"
    "github.com/yourusername/pass-cli/internal/recovery"
    "github.com/yourusername/pass-cli/internal/crypto"
)

func initVault() error {
    // Step 1: Setup recovery
    config := &recovery.SetupConfig{
        Passphrase: []byte("my-secret"), // Optional
    }
    defer crypto.ClearBytes(config.Passphrase)

    result, err := recovery.SetupRecovery(config)
    if err != nil {
        return err
    }
    defer crypto.ClearBytes([]byte(result.Mnemonic))
    defer crypto.ClearBytes(result.VaultRecoveryKey)

    // Step 2: Display mnemonic to user
    fmt.Println("Write down these 24 words:")
    words := strings.Split(result.Mnemonic, " ")
    for i, word := range words {
        fmt.Printf("%2d. %s\n", i+1, word)
    }

    // Step 3: Verify backup
    positions, _ := recovery.SelectVerifyPositions(3)
    userWords := promptUserForWords(positions) // Prompt user

    verifyConfig := &recovery.VerifyConfig{
        Mnemonic:        result.Mnemonic,
        VerifyPositions: positions,
        UserWords:       userWords,
    }

    if err := recovery.VerifyBackup(verifyConfig); err != nil {
        return fmt.Errorf("backup verification failed")
    }

    // Step 4: Store metadata in vault
    vaultMetadata.Recovery = result.Metadata

    // Step 5: Encrypt vault with recovery key
    vault.EncryptWithKey(result.VaultRecoveryKey, credentials)

    return nil
}
```

### Recovery

```go
func recoverVault(metadata *recovery.RecoveryMetadata) error {
    // Step 1: Shuffle challenge positions for prompts
    positions := recovery.ShuffleChallengePositions(metadata.ChallengePositions)

    // Step 2: Prompt user for words
    var challengeWords []string
    for i, pos := range positions {
        fmt.Printf("Enter word #%d: ", pos+1)
        word := readWord()

        if !recovery.ValidateWord(word) {
            fmt.Println("Invalid word. Try again.")
            // Retry logic
        }

        challengeWords = append(challengeWords, word)
        fmt.Printf("✓ (%d/6)\n", i+1)
    }

    // Step 3: Prompt for passphrase if required
    var passphrase []byte
    if metadata.PassphraseRequired {
        fmt.Print("Enter recovery passphrase: ")
        passphrase = readSecureInput()
    }
    defer crypto.ClearBytes(passphrase)

    // Step 4: Perform recovery
    config := &recovery.RecoveryConfig{
        ChallengeWords: challengeWords,
        Passphrase:     passphrase,
        Metadata:       metadata,
    }

    vaultKey, err := recovery.PerformRecovery(config)
    if err != nil {
        return fmt.Errorf("recovery failed: %w", err)
    }
    defer crypto.ClearBytes(vaultKey)

    // Step 5: Unlock vault
    vault.UnlockWithKey(vaultKey)

    return nil
}
```

---

## Testing Contract

### Unit Test Coverage

1. **SetupRecovery**:
   - Valid config → success
   - Empty passphrase → success (no passphrase)
   - Custom KDF params → success
   - Entropy failure → error

2. **PerformRecovery**:
   - Correct words + passphrase → success
   - Wrong words → `ErrDecryptionFailed`
   - Wrong passphrase → `ErrDecryptionFailed`
   - Invalid word → `ErrInvalidWord`
   - Recovery disabled → `ErrRecoveryDisabled`

3. **VerifyBackup**:
   - Correct words → success
   - Wrong words → `ErrVerificationFailed`
   - Invalid positions → `ErrInvalidPositions`

4. **ValidateWord**:
   - Valid word → `true`
   - Invalid word → `false`
   - Case-insensitive → `true`

5. **Memory Clearing**:
   - All sensitive data cleared after use
   - Defer paths execute on error

---

## Backward Compatibility

**Version 1.0.0**: Initial release (no compatibility concerns)

**Future Versions**:
- Breaking changes: Bump major version (2.0.0)
- New features: Bump minor version (1.1.0)
- Bug fixes: Bump patch version (1.0.1)

**Schema Versioning**: `RecoveryMetadata.Version` field enables schema evolution.

---

## Summary

Public API provides 6 functions covering:
- ✅ Recovery setup (generation, encryption, metadata)
- ✅ Recovery execution (decryption, vault unlocking)
- ✅ Backup verification (during init)
- ✅ Helper utilities (position selection, word validation)

All functions follow library-first architecture, require no CLI dependencies, and prioritize security (memory clearing, error handling, validation).

Ready for implementation.
