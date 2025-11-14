# Research: BIP39 Mnemonic Recovery

**Phase**: 0 (Pre-Implementation Research)
**Date**: 2025-01-13
**Status**: Complete

## Overview

Research conducted to resolve technical unknowns and establish best practices for implementing BIP39-based vault recovery in pass-cli. Focus areas: BIP39 library usage, cryptographic parameters, security patterns, and UX considerations.

---

## 1. BIP39 Library Selection & Usage

### Decision: Use `tyler-smith/go-bip39`

**Rationale**:
- Most widely adopted Go BIP39 library (used in major crypto wallets)
- Pure Go implementation (no CGO dependencies)
- Standard-compliant with BIP39 specification
- Minimal dependency footprint (reduces supply chain risk per Constitution Principle VII)
- Active maintenance and proven in production

**Alternatives Considered**:
- `pactus-project/pactus/util/bip39`: Forked version, less adoption
- `skycoin/skycoin/src/cipher/bip39`: Coupled to Skycoin ecosystem
- Rolling our own: Rejected per Principle VII (complexity, security risk)

### Usage Patterns

**Mnemonic Generation** (256-bit entropy for 24 words):
```go
import "github.com/tyler-smith/go-bip39"

// Generate 256-bit entropy (24 words)
entropy, err := bip39.NewEntropy(256)
if err != nil {
    return fmt.Errorf("entropy generation failed: %w", err)
}

// Convert to mnemonic
mnemonic, err := bip39.NewMnemonic(entropy)
if err != nil {
    return fmt.Errorf("mnemonic generation failed: %w", err)
}
// mnemonic is space-separated string: "word1 word2 ... word24"
```

**Seed Derivation** (with optional passphrase):
```go
// Derive 64-byte seed from mnemonic + passphrase
seed := bip39.NewSeed(mnemonic, passphrase)
// If no passphrase: seed := bip39.NewSeed(mnemonic, "")
```

**Validation**:
```go
// Validate mnemonic (checks wordlist + checksum)
valid := bip39.IsMnemonicValid(mnemonic)
```

**Word Splitting** (for puzzle piece approach):
```go
words := strings.Split(mnemonic, " ")
// words[0] = first word, words[23] = 24th word
```

**Security Note**: Never log `mnemonic` or `seed` variables. Use `defer crypto.ClearBytes()` immediately after allocation.

---

## 2. Argon2 Parameter Tuning

### Decision: Two KDF Configurations

**Challenge KDF** (6-word mnemonic → encryption key):
```go
import "golang.org/x/crypto/argon2"

const (
    ChallengeTime      = 1
    ChallengeMemory    = 65536  // 64 MB
    ChallengeThreads   = 4
    ChallengeKeyLen    = 32     // AES-256
)

// Derive key from 6-word challenge
seed := bip39.NewSeed(challengeMnemonic, passphrase) // challengeMnemonic = 6 words joined
key := argon2.IDKey(seed, saltChallenge, ChallengeTime, ChallengeMemory, ChallengeThreads, ChallengeKeyLen)
```

**Recovery KDF** (full 24-word mnemonic → vault recovery key):
```go
const (
    RecoveryTime      = 1
    RecoveryMemory    = 65536  // 64 MB
    RecoveryThreads   = 4
    RecoveryKeyLen    = 32
)

// Derive key from full 24-word mnemonic
seed := bip39.NewSeed(fullMnemonic, passphrase)
key := argon2.IDKey(seed, saltRecovery, RecoveryTime, RecoveryMemory, RecoveryThreads, RecoveryKeyLen)
```

**Rationale**:
- **Memory Cost (64 MB)**: Balances security vs. user experience. RFC 9106 draft recommends 64 MB as sensible default for password hashing.
- **Time Cost (1 iteration)**: Sufficient with high memory cost. Increasing time is less effective than increasing memory for GPU resistance.
- **Threads (4)**: Matches existing vault KDF parameters. Allows parallelism on modern CPUs.
- **Single iteration acceptable**: With 2^66 entropy from 6-word challenge, brute force is computationally infeasible even with time=1.

**Performance Estimate**:
- 64 MB, time=1, parallelism=4: ~50-100ms per derivation on consumer hardware
- Recovery flow: 2 derivations (challenge + recovery) = 100-200ms total (well under 30s goal)

**Alternatives Considered**:
- Higher time cost (3-5 iterations): Rejected. Adds 200-400ms delay without meaningful security gain given 2^66 entropy.
- Lower memory (32 MB): Rejected. Standard BIP39 seed has 512 bits of entropy; higher memory justified.
- Dynamic tuning based on hardware: Deferred to future enhancement (Principle VII: YAGNI).

---

## 3. AES-GCM Encryption for Stored Words

### Decision: AES-256-GCM with Unique Nonces

**Pattern** (encrypt 18 stored words):
```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
)

func encryptStoredWords(words []string, key []byte) (ciphertext, nonce []byte, err error) {
    // Serialize words (JSON for simplicity, can optimize later)
    plaintext, err := json.Marshal(words)
    if err != nil {
        return nil, nil, err
    }

    // Create AES-256-GCM cipher
    block, err := aes.NewCipher(key) // key is 32 bytes
    if err != nil {
        return nil, nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, nil, err
    }

    // Generate unique nonce (12 bytes for GCM)
    nonce = make([]byte, gcm.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, nil, err
    }

    // Encrypt
    ciphertext = gcm.Seal(nil, nonce, plaintext, nil)

    return ciphertext, nonce, nil
}
```

**Decrypt**:
```go
func decryptStoredWords(ciphertext, nonce, key []byte) ([]string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, fmt.Errorf("decryption failed: %w", err)
    }

    var words []string
    if err := json.Unmarshal(plaintext, &words); err != nil {
        return nil, err
    }

    return words, nil
}
```

**Rationale**:
- **GCM Mode**: Authenticated encryption (detects tampering). Required per Constitution Principle I.
- **Unique Nonces**: Generated via `crypto/rand` (CSPRNG). Prevents nonce reuse attacks.
- **No Associated Data**: Stored words don't require additional authenticated data (AD parameter nil).
- **JSON Serialization**: Simple, human-readable in metadata. Alternative: binary encoding (gob) for space efficiency (deferred per YAGNI).

**Security Properties**:
- Confidentiality: AES-256 encryption
- Integrity: GCM authentication tag (detects corruption/tampering)
- Nonce uniqueness: Guaranteed by CSPRNG

---

## 4. Challenge Position Selection

### Decision: Crypto-Secure Random Selection

**Pattern**:
```go
import (
    "crypto/rand"
    "math/big"
)

func selectChallengePositions(totalWords, count int) ([]int, error) {
    if count > totalWords {
        return nil, errors.New("count exceeds total words")
    }

    positions := make([]int, 0, count)
    used := make(map[int]bool)

    for len(positions) < count {
        // Generate random position [0, totalWords)
        n, err := rand.Int(rand.Reader, big.NewInt(int64(totalWords)))
        if err != nil {
            return nil, err
        }

        pos := int(n.Int64())
        if !used[pos] {
            positions = append(positions, pos)
            used[pos] = true
        }
    }

    sort.Ints(positions) // Sort for consistent metadata storage
    return positions, nil
}

// Example: selectChallengePositions(24, 6) → [3, 7, 11, 15, 18, 22]
```

**Rationale**:
- **crypto/rand**: CSPRNG required per FR-029
- **No Duplicates**: Map-based deduplication
- **Sorted Output**: Easier metadata inspection, no security impact
- **Fixed at Init**: Positions never change for vault lifetime (simplicity)

**Randomization for Recovery** (shuffle order of prompts):
```go
import "math/rand"

func shufflePositions(positions []int) []int {
    shuffled := make([]int, len(positions))
    copy(shuffled, positions)

    rand.Shuffle(len(shuffled), func(i, j int) {
        shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
    })

    return shuffled
}

// Used during recovery to randomize prompt order (not crypto-critical)
```

**Note**: Recovery order randomization uses `math/rand` (not `crypto/rand`) because it's non-security-critical (positions are fixed; only prompt order varies for UX).

---

## 5. Memory Clearing Best Practices

### Decision: Use Existing `crypto.ClearBytes()` Pattern

**Pattern** (from existing `internal/crypto` package):
```go
// In internal/recovery functions
func GenerateMnemonic() (string, error) {
    entropy, err := bip39.NewEntropy(256)
    if err != nil {
        return "", err
    }
    defer crypto.ClearBytes(entropy) // Clear entropy from memory

    mnemonic, err := bip39.NewMnemonic(entropy)
    if err != nil {
        return "", err
    }

    // mnemonic is string (returned to caller for display)
    // Caller responsible for clearing after use
    return mnemonic, nil
}

func PerformRecovery(challengeWords []string, passphrase []byte) error {
    defer crypto.ClearBytes(passphrase) // Clear passphrase

    // Reconstruct mnemonic
    fullMnemonic := reconstructMnemonic(challengeWords, storedWords)
    defer func() {
        // Clear string by converting to []byte and zeroing
        mnemonicBytes := []byte(fullMnemonic)
        crypto.ClearBytes(mnemonicBytes)
    }()

    seed := bip39.NewSeed(fullMnemonic, string(passphrase))
    defer crypto.ClearBytes(seed) // Clear seed

    key := argon2.IDKey(seed, salt, time, memory, threads, keyLen)
    defer crypto.ClearBytes(key) // Clear derived key

    // ... use key to unlock vault ...

    return nil
}
```

**Variables Requiring Clearing**:
1. **Entropy** ([]byte): Raw random data for mnemonic generation
2. **Mnemonic** (string → []byte): Full 24-word phrase
3. **Passphrase** ([]byte): User's 25th word
4. **BIP39 Seed** ([]byte): 64-byte seed from mnemonic+passphrase
5. **Derived Keys** ([]byte): Argon2 output (challenge key, recovery key)
6. **Decrypted Stored Words** ([]string → []byte): 18 words after decryption
7. **Vault Recovery Key** ([]byte): Final key that unlocks vault

**What NOT to Clear**:
- **Challenge Positions** ([]int): Non-sensitive metadata
- **Encrypted Stored Words** ([]byte): Already encrypted, safe to persist
- **Nonces** ([]byte): Non-secret, required for decryption
- **Salts** ([]byte): Non-secret, required for KDF

---

## 6. BIP39 Checksum Validation

### Decision: Rely on `go-bip39` Built-In Validation

**Pattern**:
```go
// During recovery, after reconstructing full mnemonic
fullMnemonic := strings.Join(allWords, " ")

if !bip39.IsMnemonicValid(fullMnemonic) {
    return errors.New("invalid recovery phrase (checksum mismatch)")
}

// Proceed with seed derivation
seed := bip39.NewSeed(fullMnemonic, passphrase)
```

**Rationale**:
- BIP39 spec embeds 8-bit checksum in 24th word
- `IsMnemonicValid()` checks:
  1. All words in standard wordlist
  2. Checksum matches entropy
- Detects:
  - Typos in individual words
  - Incorrect word order
  - Corrupted stored words (if decryption succeeded but data corrupted)

**Error Handling**:
- Invalid checksum → User entered wrong words or data corruption
- Clear error message per FR-031: "Invalid recovery phrase. Verify words and try again."

---

## 7. Recovery UX Patterns (Industry Research)

### Observed Patterns in Crypto Wallets

**Ledger Live**:
- 24-word phrase displayed in 4x6 grid
- Verification: Asks for 3 random words before continuing
- Recovery: Sequential entry (paste-friendly for power users)

**MetaMask**:
- 12-word phrase (128-bit entropy)
- Verification: Click words in order from shuffled list
- Recovery: Sequential paste into single text area

**Trezor**:
- 24-word phrase with position numbers
- Verification: Asks for specific words by position
- Recovery: Hardware device prompts word-by-word (no paste)

**Coinbase Wallet**:
- 12-word phrase
- Verification: Quiz-style (select word #X from list)
- Recovery: Sequential paste

### Design Choices for pass-cli

**Init Verification**:
- **Decision**: Ask for 3 random words (Ledger-style)
- **Rationale**: Balances security (catches transcription errors) vs. UX (3 words faster than 24)
- **Pattern**: Position numbers shown to user, e.g., "Enter word #7:", "Enter word #18:", "Enter word #22:"

**Recovery Challenge**:
- **Decision**: 6 words in randomized order (unique to pass-cli)
- **Rationale**:
  - 6 words = 2^66 entropy (secure against brute force)
  - Randomized order prevents muscle memory, forces reference to written phrase
  - Faster than 24-word entry (~10-15 seconds vs. 1-2 minutes)

**Word Entry**:
- **Decision**: Interactive one-at-a-time (not paste)
- **Rationale**:
  - Security: Prevents clipboard history leakage
  - Validation: Immediate feedback on invalid words (not in BIP39 wordlist)
  - UX: Progress indicators ("✓ (3/6)") reduce anxiety

**Progress Feedback**:
- **Decision**: Show count after each word ("✓ (1/6)", "✓ (2/6)", etc.)
- **Rationale**: User knows how many words remain, reduces uncertainty

---

## 8. Metadata Structure Design

### Decision: Extend Existing VaultMetadata JSON

**Schema** (addition to `internal/vault/metadata.go`):
```go
type VaultMetadata struct {
    // ... existing fields (Version, Created, etc.) ...

    Recovery *RecoveryMetadata `json:"recovery,omitempty"`
}

type RecoveryMetadata struct {
    Enabled             bool     `json:"enabled"`
    Version             string   `json:"version"`              // "1" for initial implementation
    PassphraseRequired  bool     `json:"passphrase_required"`
    ChallengePositions  []int    `json:"challenge_positions"`  // e.g., [3, 7, 11, 15, 18, 22]

    // KDF parameters
    KDFParams KDFParams `json:"kdf_params"`

    // Encrypted data
    EncryptedStoredWords []byte `json:"encrypted_stored_words"` // Base64-encoded
    NonceStored          []byte `json:"nonce_stored"`           // Base64-encoded
    EncryptedRecoveryKey []byte `json:"encrypted_recovery_key"` // Base64-encoded
    NonceRecovery        []byte `json:"nonce_recovery"`         // Base64-encoded
}

type KDFParams struct {
    Algorithm     string `json:"algorithm"`      // "argon2id"
    Time          uint32 `json:"time"`           // 1
    Memory        uint32 `json:"memory"`         // 65536
    Threads       uint8  `json:"threads"`        // 4
    SaltChallenge []byte `json:"salt_challenge"` // Base64-encoded
    SaltRecovery  []byte `json:"salt_recovery"`  // Base64-encoded
}
```

**Example JSON** (in `vault.enc` metadata section):
```json
{
  "version": "1.0",
  "created": "2025-01-13T10:30:00Z",
  "recovery": {
    "enabled": true,
    "version": "1",
    "passphrase_required": false,
    "challenge_positions": [3, 7, 11, 15, 18, 22],
    "kdf_params": {
      "algorithm": "argon2id",
      "time": 1,
      "memory": 65536,
      "threads": 4,
      "salt_challenge": "base64...",
      "salt_recovery": "base64..."
    },
    "encrypted_stored_words": "base64...",
    "nonce_stored": "base64...",
    "encrypted_recovery_key": "base64...",
    "nonce_recovery": "base64..."
  }
}
```

**Size Estimate**:
- Challenge positions (6 ints): ~50 bytes
- KDF params (including salts): ~150 bytes
- Encrypted stored words (18 words ~= 180 chars + overhead): ~250 bytes
- Encrypted recovery key (32 bytes + overhead): ~50 bytes
- Total: ~500 bytes (within SC-007 constraint)

**Rationale**:
- **JSON**: Human-readable, easy to inspect/debug
- **Base64 Encoding**: Binary data (nonces, ciphertexts, salts) stored as strings
- **Omitempty**: If recovery disabled, `recovery` field is null (backward compatible)
- **Version Field**: Enables future schema evolution (e.g., v2 with different KDF)

---

## 9. Error Handling Strategy

### Decision: Layered Error Types

**Custom Errors** (in `internal/recovery/errors.go`):
```go
var (
    ErrEntropyGeneration    = errors.New("failed to generate entropy")
    ErrInvalidWord          = errors.New("word not in BIP39 wordlist")
    ErrInvalidMnemonic      = errors.New("invalid recovery phrase (checksum mismatch)")
    ErrDecryptionFailed     = errors.New("decryption failed (wrong words or corrupted data)")
    ErrRecoveryDisabled     = errors.New("recovery not enabled for this vault")
    ErrMetadataCorrupted    = errors.New("vault metadata corrupted")
    ErrWrongPassphrase      = errors.New("incorrect recovery passphrase")
)
```

**User-Facing Messages** (in CLI layer):
```go
// cmd/change_password.go
err := recovery.PerformRecovery(...)
if err != nil {
    switch {
    case errors.Is(err, recovery.ErrInvalidWord):
        return fmt.Errorf("✗ Invalid word. Try again.")
    case errors.Is(err, recovery.ErrInvalidMnemonic):
        return fmt.Errorf("✗ Invalid recovery phrase. Verify words and try again.")
    case errors.Is(err, recovery.ErrDecryptionFailed):
        return fmt.Errorf("✗ Recovery failed: Incorrect recovery words")
    case errors.Is(err, recovery.ErrWrongPassphrase):
        return fmt.Errorf("✗ Recovery failed: Incorrect passphrase")
    case errors.Is(err, recovery.ErrRecoveryDisabled):
        return fmt.Errorf("✗ Recovery not enabled for this vault")
    default:
        return fmt.Errorf("✗ Recovery failed: %w", err)
    }
}
```

**Rationale**:
- **Sentinel Errors**: Testable, clear error handling
- **User-Friendly**: CLI layer translates technical errors to actionable messages
- **No Sensitive Data**: Error messages never include words, passphrases, or keys

---

## 10. Testing Strategy

### Unit Tests (`internal/recovery/recovery_test.go`)

**Coverage Areas**:
1. **Mnemonic Generation**:
   - Test 256-bit entropy → 24-word mnemonic
   - Validate checksum correctness
   - Verify all words in BIP39 wordlist

2. **Challenge Position Selection**:
   - Test randomness (run 100 times, ensure distribution)
   - Test uniqueness (no duplicate positions)
   - Test count validation (6 from 24)

3. **Word Splitting & Reconstruction**:
   - Split 24 words → 6 challenge + 18 stored
   - Reconstruct from 6 + 18 → full 24
   - Validate reconstructed mnemonic checksum

4. **Encryption/Decryption**:
   - Encrypt 18 stored words with challenge-derived key
   - Decrypt and verify match
   - Test wrong key → decryption failure

5. **KDF**:
   - Test Argon2id parameters
   - Verify key length (32 bytes)
   - Test passphrase vs. no passphrase

6. **Memory Clearing**:
   - Verify `crypto.ClearBytes()` called on sensitive variables
   - Test defer paths (error cases still clear)

### Integration Tests (`test/recovery_test.go`)

**Test Scenarios**:
1. **Init with Recovery** (User Story 2):
   - `pass-cli init` → recovery phrase generated
   - Metadata contains recovery fields
   - Verification prompt accepts correct words

2. **Recovery Success** (User Story 1):
   - Init vault with known mnemonic
   - `pass-cli change-password --recover`
   - Provide correct 6 words → vault unlocks

3. **Recovery Failure**:
   - Wrong words → error message
   - Wrong passphrase → error message
   - Invalid word (not in wordlist) → immediate feedback

4. **No Recovery**:
   - `pass-cli init --no-recovery`
   - Metadata `recovery: null`
   - `pass-cli change-password --recover` → error

5. **Passphrase Detection** (User Story 3):
   - Init with passphrase
   - Recovery prompts for passphrase
   - No passphrase → no prompt

**Security Tests**:
- Audit log contains no sensitive data
- Memory dumps don't reveal mnemonic/passphrase
- BIP39 compatibility (external tool validates mnemonic)

---

## 11. Implementation Order

### Recommended Sequence

**Phase 1: Core Library** (`internal/recovery`)
1. `mnemonic.go`: BIP39 generation, validation, seed derivation
2. `challenge.go`: Position selection, word splitting, reconstruction
3. `crypto.go`: KDF, AES-GCM encryption/decryption
4. `recovery.go`: Orchestration (SetupRecovery, PerformRecovery, VerifyBackup)
5. Unit tests for all above

**Phase 2: Metadata Integration** (`internal/vault`)
1. Extend `VaultMetadata` struct
2. Add `RecoveryMetadata` struct
3. Serialization/deserialization tests

**Phase 3: CLI Integration** (`cmd/`)
1. `cmd/init.go`: Recovery setup flow, display phrase, verification
2. `cmd/change_password.go`: Add `--recover` flag, recovery challenge flow
3. `cmd/helpers.go`: Recovery-specific prompts (word entry, passphrase)
4. Integration tests

**Phase 4: Documentation & Polishing**
1. `quickstart.md`: User guide for recovery setup/usage
2. `contracts/recovery-api.md`: Public API documentation
3. Security audit (audit log, memory clearing, no leaks)
4. Cross-platform testing (Windows, macOS, Linux)

---

## Key Takeaways

1. **BIP39 Library**: `tyler-smith/go-bip39` is battle-tested and minimal (Constitution VII)
2. **Argon2 Parameters**: 64 MB memory, 1 iteration balances security and UX
3. **6-Word Challenge**: 2^66 entropy secures against brute force, faster than 24-word entry
4. **Puzzle Piece Approach**: Encrypt 18 words with key from 6 words, reconstruct full phrase
5. **Memory Clearing**: Use existing `crypto.ClearBytes()` pattern throughout
6. **Metadata Size**: ~500 bytes fits within constraint (SC-007)
7. **Error Handling**: Layered (library errors → user-friendly CLI messages)
8. **Testing**: TDD required (unit → integration → security)
9. **Implementation Order**: Core library → metadata → CLI → docs

All technical unknowns resolved. Ready for Phase 1 (Design & Contracts).
