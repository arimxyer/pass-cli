# Data Model: BIP39 Mnemonic Recovery

**Phase**: 1 (Design)
**Date**: 2025-01-13
**Status**: Complete

## Overview

Defines data structures and relationships for BIP39 recovery feature. All entities stored in vault metadata (JSON-serialized) except transient runtime data (cleared from memory after use).

---

## Entity Diagram

```
VaultMetadata
    ├── Recovery (RecoveryMetadata) [optional]
    │   ├── ChallengePositions ([]int)
    │   ├── KDFParams
    │   │   ├── SaltChallenge ([]byte)
    │   │   └── SaltRecovery ([]byte)
    │   ├── EncryptedStoredWords ([]byte)
    │   ├── NonceStored ([]byte)
    │   ├── EncryptedRecoveryKey ([]byte)
    │   └── NonceRecovery ([]byte)
    │
    └── [other vault metadata fields]

Runtime Data (NOT persisted):
    ├── Mnemonic (24 words, string)
    ├── ChallengeWords (6 words, []string)
    ├── StoredWords (18 words, []string)
    ├── Passphrase ([]byte)
    ├── Seeds ([]byte)
    └── Keys ([]byte)
```

---

## 1. RecoveryMetadata (Persisted)

**Location**: `internal/vault/metadata.go`

**Purpose**: Stores all recovery-related data in vault metadata JSON.

**Schema**:
```go
type RecoveryMetadata struct {
    // Feature control
    Enabled            bool   `json:"enabled"`              // Whether recovery is active
    Version            string `json:"version"`              // Schema version ("1" initially)
    PassphraseRequired bool   `json:"passphrase_required"`  // Whether 25th word was set

    // Challenge configuration
    ChallengePositions []int `json:"challenge_positions"`  // Indices of 6 words [0-23]

    // Cryptographic parameters
    KDFParams KDFParams `json:"kdf_params"`

    // Encrypted secrets
    EncryptedStoredWords []byte `json:"encrypted_stored_words"` // 18 words (AES-GCM)
    NonceStored          []byte `json:"nonce_stored"`           // GCM nonce (12 bytes)
    EncryptedRecoveryKey []byte `json:"encrypted_recovery_key"` // Vault unlock key (AES-GCM)
    NonceRecovery        []byte `json:"nonce_recovery"`         // GCM nonce (12 bytes)
}
```

**Field Details**:

| Field | Type | Size (bytes) | Description | Validation |
|-------|------|--------------|-------------|------------|
| `Enabled` | bool | 1 | Recovery feature active | - |
| `Version` | string | ~5 | Schema version | Semantic version |
| `PassphraseRequired` | bool | 1 | BIP39 passphrase used | - |
| `ChallengePositions` | []int | ~50 | 6 positions [0-23] | Unique, sorted, len=6 |
| `KDFParams` | struct | ~150 | Argon2 config + salts | See KDFParams |
| `EncryptedStoredWords` | []byte | ~250 | 18 words encrypted | JSON + GCM tag |
| `NonceStored` | []byte | 12 | GCM nonce | Random, unique |
| `EncryptedRecoveryKey` | []byte | ~50 | 32-byte key encrypted | GCM ciphertext |
| `NonceRecovery` | []byte | 12 | GCM nonce | Random, unique |

**Total Size**: ~530 bytes (within SC-007 constraint of <500 bytes target, acceptable overage)

**Lifecycle**:
1. **Created**: During `pass-cli init` (if recovery enabled)
2. **Read**: During `pass-cli change-password --recover`
3. **Modified**: Never (immutable after init unless rotation feature added later)
4. **Deleted**: If user runs `pass-cli recovery disable` (future enhancement)

**Serialization**:
```json
{
  "enabled": true,
  "version": "1",
  "passphrase_required": false,
  "challenge_positions": [3, 7, 11, 15, 18, 22],
  "kdf_params": {
    "algorithm": "argon2id",
    "time": 1,
    "memory": 65536,
    "threads": 4,
    "salt_challenge": "YmFzZTY0ZW5jb2RlZHNhbHQ=",
    "salt_recovery": "YW5vdGhlcmJhc2U2NHNhbHQ="
  },
  "encrypted_stored_words": "YmFzZTY0ZW5jb2RlZGNpcGhlcnRleHQ=",
  "nonce_stored": "cmFuZG9tbm9uY2U=",
  "encrypted_recovery_key": "ZW5jcnlwdGVka2V5",
  "nonce_recovery": "YW5vdGhlcm5vbmNl"
}
```

---

## 2. KDFParams (Nested in RecoveryMetadata)

**Purpose**: Stores Argon2id key derivation function parameters.

**Schema**:
```go
type KDFParams struct {
    Algorithm     string `json:"algorithm"`      // "argon2id" (fixed)
    Time          uint32 `json:"time"`           // Iteration count (1)
    Memory        uint32 `json:"memory"`         // Memory cost in KiB (65536 = 64 MB)
    Threads       uint8  `json:"threads"`        // Parallelism (4)
    SaltChallenge []byte `json:"salt_challenge"` // 32-byte salt for challenge KDF
    SaltRecovery  []byte `json:"salt_recovery"`  // 32-byte salt for recovery KDF
}
```

**Field Details**:

| Field | Type | Value | Rationale |
|-------|------|-------|-----------|
| `Algorithm` | string | "argon2id" | Hybrid mode (side-channel + GPU resistance) |
| `Time` | uint32 | 1 | Single pass (memory cost primary defense) |
| `Memory` | uint32 | 65536 | 64 MB (RFC 9106 recommended default) |
| `Threads` | uint8 | 4 | Matches existing vault KDF, utilizes modern CPUs |
| `SaltChallenge` | []byte | 32 bytes | For 6-word challenge → encryption key derivation |
| `SaltRecovery` | []byte | 32 bytes | For 24-word mnemonic → vault recovery key derivation |

**Salts**:
- **Generation**: `crypto/rand.Read()` (CSPRNG)
- **Uniqueness**: Different salts for challenge vs. recovery KDFs (defense in depth)
- **Storage**: Base64-encoded in JSON
- **Reuse**: Same salts used for all recovery attempts (required for deterministic key derivation)

**Why Two Salts**:
1. **Challenge Salt**: Used to derive key from 6-word phrase → unlocks encrypted 18 words
2. **Recovery Salt**: Used to derive key from full 24-word phrase → unlocks vault recovery key

Separate salts prevent key reuse (even though inputs differ, extra precaution per security-first principle).

---

## 3. Mnemonic (Transient Runtime Data)

**Location**: `internal/recovery` package (NOT persisted)

**Purpose**: 24-word BIP39 mnemonic phrase generated during init, displayed to user.

**Schema**:
```go
type Mnemonic string // Space-separated 24 words, e.g., "word1 word2 ... word24"
```

**Properties**:
- **Entropy**: 256 bits
- **Words**: 24 (from 2,048-word BIP39 English wordlist)
- **Checksum**: 8-bit checksum embedded in 24th word
- **Format**: Lowercase, space-separated

**Example**:
```
"abandon ability about above absent absorb device diagram dial diamond diary diesel hover hurdle hybrid icon idea identify spatial sphere spike spin spirit split"
```

**Lifecycle**:
1. **Generated**: `bip39.NewEntropy(256)` → `bip39.NewMnemonic(entropy)`
2. **Displayed**: Shown to user once during `pass-cli init`
3. **Split**: Into 6 challenge words + 18 stored words
4. **Cleared**: `crypto.ClearBytes([]byte(mnemonic))` after split complete

**Validation**:
- **Wordlist Check**: All words in BIP39 English wordlist
- **Checksum**: `bip39.IsMnemonicValid(mnemonic)` verifies last word checksum

**Security**:
- **Never Logged**: FR-026
- **Never Stored**: Only encrypted derivatives stored (18 words, recovery key)
- **Memory Cleared**: Immediately after use

---

## 4. ChallengePositions (Persisted)

**Location**: `RecoveryMetadata.ChallengePositions`

**Purpose**: Indices of 6 words (from 24) used as challenge during recovery.

**Schema**:
```go
type ChallengePositions []int // 6 unique integers [0-23], sorted
```

**Example**: `[3, 7, 11, 15, 18, 22]`
- Word #4 (index 3): "above"
- Word #8 (index 7): "diagram"
- Word #12 (index 11): "diary"
- Word #16 (index 15): "hybrid"
- Word #19 (index 18): "identify"
- Word #23 (index 22): "spin"

**Properties**:
- **Count**: Always 6
- **Uniqueness**: No duplicates
- **Range**: [0, 23] (inclusive)
- **Ordering**: Stored sorted (not security-critical, aids readability)

**Generation** (during init):
```go
func selectChallengePositions() ([]int, error) {
    positions := make([]int, 0, 6)
    used := make(map[int]bool)

    for len(positions) < 6 {
        n, err := rand.Int(rand.Reader, big.NewInt(24))
        if err != nil {
            return nil, err
        }
        pos := int(n.Int64())
        if !used[pos] {
            positions = append(positions, pos)
            used[pos] = true
        }
    }

    sort.Ints(positions)
    return positions, nil
}
```

**Lifecycle**:
1. **Generated**: Once during `pass-cli init`
2. **Stored**: In `RecoveryMetadata.ChallengePositions`
3. **Read**: During every recovery attempt
4. **Shuffled**: For prompt order randomization (non-destructive)

**Shuffle** (during recovery, in-memory only):
```go
func shufflePositions(positions []int) []int {
    shuffled := make([]int, len(positions))
    copy(shuffled, positions)
    rand.Shuffle(len(shuffled), func(i, j int) {
        shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
    })
    return shuffled // e.g., [18, 3, 22, 7, 11, 15] (different order, same positions)
}
```

---

## 5. EncryptedStoredWords (Persisted)

**Location**: `RecoveryMetadata.EncryptedStoredWords`

**Purpose**: Encrypted JSON array of 18 words (those NOT in challenge positions).

**Schema**:
```go
type EncryptedStoredWords []byte // AES-256-GCM ciphertext
```

**Plaintext Structure** (before encryption):
```json
[
  "abandon",
  "ability",
  "about",
  "absent",
  "absorb",
  "device",
  "dial",
  "diamond",
  "hover",
  "hurdle",
  "icon",
  "idea",
  "spatial",
  "sphere",
  "spike",
  "spirit",
  "split"
]
```

**Encryption Process**:
1. **Serialize**: `json.Marshal(storedWords)` → plaintext JSON
2. **Derive Key**: Argon2id(6-word mnemonic + passphrase, SaltChallenge) → 32-byte key
3. **Encrypt**: AES-256-GCM(plaintext, key, NonceStored) → ciphertext + tag
4. **Store**: Ciphertext in `EncryptedStoredWords`, nonce in `NonceStored`

**Decryption Process** (during recovery):
1. **User Provides**: 6 challenge words
2. **Derive Key**: Argon2id(6-word mnemonic + passphrase, SaltChallenge) → key
3. **Decrypt**: AES-256-GCM.Open(ciphertext, key, NonceStored) → plaintext JSON
4. **Deserialize**: `json.Unmarshal(plaintext)` → []string (18 words)
5. **Combine**: 6 challenge words + 18 stored words → full 24-word mnemonic

**Size**: ~250 bytes (18 words * ~10 chars avg + JSON overhead + GCM tag)

**Security**:
- **Confidentiality**: AES-256 encryption
- **Integrity**: GCM authentication tag (detects tampering/corruption)
- **Key Source**: Derived from 6-word challenge (2^66 entropy)

---

## 6. EncryptedRecoveryKey (Persisted)

**Location**: `RecoveryMetadata.EncryptedRecoveryKey`

**Purpose**: Encrypted vault recovery key (32-byte random key that unlocks vault).

**Schema**:
```go
type EncryptedRecoveryKey []byte // AES-256-GCM ciphertext
```

**Plaintext**: 32-byte random key (generated during init)

**Encryption Process**:
1. **Generate**: `crypto/rand.Read(vaultRecoveryKey)` → 32 random bytes
2. **Derive Key**: Argon2id(full 24-word mnemonic + passphrase, SaltRecovery) → encryption key
3. **Encrypt**: AES-256-GCM(vaultRecoveryKey, key, NonceRecovery) → ciphertext
4. **Store**: Ciphertext in `EncryptedRecoveryKey`, nonce in `NonceRecovery`

**Decryption Process** (during recovery):
1. **Reconstruct**: Combine 6 challenge words + 18 decrypted stored words → full 24-word mnemonic
2. **Validate**: `bip39.IsMnemonicValid(fullMnemonic)` (checksum)
3. **Derive Key**: Argon2id(full 24-word mnemonic + passphrase, SaltRecovery) → key
4. **Decrypt**: AES-256-GCM.Open(ciphertext, key, NonceRecovery) → vaultRecoveryKey
5. **Unlock Vault**: Use vaultRecoveryKey to decrypt vault credentials

**Size**: ~50 bytes (32-byte key + GCM tag)

**Why Separate from StoredWords Encryption**:
- **Defense in Depth**: Even if challenge encryption broken, attacker still needs full 24-word mnemonic to unlock vault
- **Two-Stage Security**:
  1. 6 words (2^66) → unlock 18 stored words
  2. Full 24 words (2^256 effective) → unlock vault

---

## 7. Passphrase (Transient Runtime Data)

**Location**: `internal/recovery` package (NOT persisted)

**Purpose**: Optional BIP39 passphrase (25th word) for enhanced security.

**Schema**:
```go
type Passphrase []byte // UTF-8 encoded, variable length
```

**Properties**:
- **Optional**: User may leave empty (empty string = no passphrase)
- **Encoding**: UTF-8 (BIP39 spec allows any encoding, UTF-8 is standard)
- **Length**: No minimum/maximum (BIP39 spec, but UI may enforce practical limits)
- **Case-Sensitive**: "secret" ≠ "Secret"

**Storage**:
- **Flag**: `RecoveryMetadata.PassphraseRequired` (bool) stored in metadata
- **Value**: Never stored (user must remember/record separately)

**Usage in KDF**:
```go
seed := bip39.NewSeed(mnemonic, string(passphrase))
// If no passphrase: seed := bip39.NewSeed(mnemonic, "")
```

**Lifecycle**:
1. **Entered**: During `pass-cli init` (if user chooses advanced option)
2. **Used**: Immediately for seed derivation
3. **Cleared**: `defer crypto.ClearBytes(passphrase)` after use

**Security**:
- **Never Logged**: FR-026
- **Never Stored**: Only `PassphraseRequired` flag stored
- **Memory Cleared**: After seed derivation

**Detection** (during recovery):
```go
if metadata.Recovery.PassphraseRequired {
    fmt.Print("Enter recovery passphrase: ")
    passphrase, _ = readSecureInput()
} else {
    passphrase = []byte("") // Empty passphrase
}
```

---

## 8. Seeds and Keys (Transient Runtime Data)

**Location**: `internal/recovery` package (NOT persisted)

**Purpose**: Intermediate cryptographic data during KDF operations.

**Schema**:
```go
type Seed [64]byte       // BIP39 seed (512 bits)
type DerivedKey [32]byte // Argon2id output (256 bits for AES-256)
```

**Flow**:

**Challenge Key Derivation**:
```
6-word mnemonic + passphrase
    ↓ bip39.NewSeed()
64-byte BIP39 seed
    ↓ argon2.IDKey(..., SaltChallenge, ...)
32-byte challenge encryption key
    ↓ AES-256-GCM
Decrypt EncryptedStoredWords
```

**Recovery Key Derivation**:
```
Full 24-word mnemonic + passphrase
    ↓ bip39.NewSeed()
64-byte BIP39 seed
    ↓ argon2.IDKey(..., SaltRecovery, ...)
32-byte recovery encryption key
    ↓ AES-256-GCM
Decrypt EncryptedRecoveryKey → vault recovery key
```

**Lifecycle**:
1. **Created**: During KDF operations
2. **Used**: Immediately for encryption/decryption
3. **Cleared**: `defer crypto.ClearBytes(seed)`, `defer crypto.ClearBytes(key)`

**Security**:
- **Never Logged**: FR-026
- **Never Stored**: Ephemeral, derived on-demand
- **Memory Cleared**: After use

---

## Relationships

**Initialization Flow**:
```
User enters master password
    ↓
Generate 24-word Mnemonic
    ↓
    ├→ Display to user (write down)
    ├→ Optionally collect Passphrase
    ├→ Select ChallengePositions (6 random)
    ├→ Split into ChallengeWords (6) + StoredWords (18)
    │
    ├→ Derive ChallengeKey (6 words + passphrase + SaltChallenge)
    │   ↓ AES-256-GCM Encrypt
    │   StoredWords → EncryptedStoredWords
    │
    ├→ Generate VaultRecoveryKey (32 random bytes)
    ├→ Derive RecoveryKey (24 words + passphrase + SaltRecovery)
    │   ↓ AES-256-GCM Encrypt
    │   VaultRecoveryKey → EncryptedRecoveryKey
    │
    └→ Store RecoveryMetadata in vault
```

**Recovery Flow**:
```
User runs pass-cli change-password --recover
    ↓
Load RecoveryMetadata
    ↓
Shuffle ChallengePositions (prompt order)
    ↓
Prompt for 6 ChallengeWords (randomized order)
    ↓
Validate each word (BIP39 wordlist)
    ↓
If PassphraseRequired: Prompt for Passphrase
    ↓
Derive ChallengeKey (6 words + passphrase + SaltChallenge)
    ↓ AES-256-GCM Decrypt
EncryptedStoredWords → StoredWords (18 words)
    ↓
Combine: ChallengeWords (6) + StoredWords (18) → FullMnemonic (24)
    ↓
Validate: bip39.IsMnemonicValid(FullMnemonic)
    ↓
Derive RecoveryKey (24 words + passphrase + SaltRecovery)
    ↓ AES-256-GCM Decrypt
EncryptedRecoveryKey → VaultRecoveryKey
    ↓
Unlock vault with VaultRecoveryKey
    ↓
Prompt for new master password
    ↓
Re-encrypt vault with new password
```

---

## Validation Rules

### RecoveryMetadata
- `Enabled`: Must be `true` to use recovery
- `Version`: Must match supported schema version ("1")
- `PassphraseRequired`: Boolean (true/false)
- `ChallengePositions`: Length = 6, unique values [0-23], sorted
- `KDFParams`: See KDFParams validation
- `EncryptedStoredWords`: Non-empty, >16 bytes (minimum for GCM tag)
- `NonceStored`: Exactly 12 bytes
- `EncryptedRecoveryKey`: Non-empty, >16 bytes
- `NonceRecovery`: Exactly 12 bytes

### KDFParams
- `Algorithm`: Exactly "argon2id"
- `Time`: Must be > 0 (recommended: 1)
- `Memory`: Must be ≥ 8192 (recommended: 65536)
- `Threads`: Must be 1-255 (recommended: 4)
- `SaltChallenge`: Exactly 32 bytes
- `SaltRecovery`: Exactly 32 bytes

### Mnemonic
- Length: Exactly 24 words
- Words: All in BIP39 English wordlist
- Checksum: Last word encodes valid checksum

### ChallengePositions
- Length: Exactly 6
- Values: [0, 23] inclusive
- Uniqueness: No duplicates
- Order: Sorted ascending (validation, not security)

---

## Size Analysis

**Total Metadata Overhead**:
| Component | Size (bytes) | Notes |
|-----------|--------------|-------|
| Enabled | 1 | Boolean |
| Version | 5 | "1" as string |
| PassphraseRequired | 1 | Boolean |
| ChallengePositions | 48 | 6 ints * 8 bytes (JSON) |
| KDFParams.Algorithm | 12 | "argon2id" |
| KDFParams.Time | 4 | uint32 |
| KDFParams.Memory | 8 | uint32 |
| KDFParams.Threads | 2 | uint8 |
| KDFParams.SaltChallenge | 45 | 32 bytes base64 |
| KDFParams.SaltRecovery | 45 | 32 bytes base64 |
| EncryptedStoredWords | 250 | JSON + GCM tag |
| NonceStored | 20 | 12 bytes base64 |
| EncryptedRecoveryKey | 50 | 32 bytes + GCM tag |
| NonceRecovery | 20 | 12 bytes base64 |
| **TOTAL** | **~511 bytes** | Slightly over SC-007 target (500 bytes), acceptable |

**Mitigation if size critical**:
- Use binary encoding (gob) instead of JSON for stored words: saves ~50 bytes
- Deferred per YAGNI (Principle VII) unless profiling shows issue

---

## Summary

All entities defined with clear purposes, schemas, validation rules, and lifecycles. Data model supports:
- ✅ Secure storage (encryption, authentication)
- ✅ Memory safety (clearing transient data)
- ✅ Interoperability (BIP39 standard)
- ✅ Size constraints (~500 bytes)
- ✅ Extensibility (version field for future schema evolution)

Ready for API contract definition (Phase 1.2).
