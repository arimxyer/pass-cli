# Research: Recovery Key Integration

**Feature**: 013-recovery-key-integration
**Date**: 2025-12-05

## Research Questions

### Q1: Key Wrapping Standard and Implementation

**Question**: What is the best approach for wrapping a DEK with multiple key encryption keys (KEKs)?

**Decision**: Use AES-256-GCM for key wrapping (AES Key Wrap with Padding alternative considered).

**Rationale**:
- AES-GCM is already used throughout pass-cli for vault encryption
- Provides authenticated encryption (integrity + confidentiality)
- No additional dependencies needed
- NIST SP 800-38F defines AES Key Wrap, but AES-GCM is equally secure for our use case
- Simpler implementation since we already have AES-GCM infrastructure

**Alternatives Considered**:
| Alternative | Pros | Cons | Decision |
|-------------|------|------|----------|
| AES Key Wrap (RFC 3394) | NIST standard, dedicated key wrap | Additional implementation, no auth tag | Rejected |
| AES-GCM | Already implemented, auth tag | Slightly larger output (12-byte nonce + 16-byte tag) | **Selected** |
| XChaCha20-Poly1305 | Modern, larger nonce | New dependency, not in current stack | Rejected |

### Q2: DEK Generation and Lifecycle

**Question**: How should the DEK be generated, stored, and managed?

**Decision**: Generate 256-bit DEK via crypto/rand at vault init. Store only wrapped versions. Clear from memory after each operation.

**Rationale**:
- 256-bit aligns with AES-256-GCM key requirement
- crypto/rand uses OS CSPRNG (CryptGenRandom on Windows, /dev/urandom on Unix)
- Never store unwrapped DEK - always wrapped by at least one KEK
- Clear immediately after use per constitution (Principle I: Secure Memory Handling)

**Lifecycle**:
```
Init: Generate DEK → Wrap with password KEK → Wrap with recovery KEK → Store both → Clear DEK
Unlock (password): Derive KEK from password → Unwrap DEK → Decrypt vault → Clear DEK
Unlock (recovery): Derive KEK from recovery → Unwrap DEK → Prompt new password → Re-wrap with new KEK → Clear DEK
Password change: Unwrap DEK with old KEK → Wrap with new KEK → Clear DEK
```

### Q3: Vault Format Versioning

**Question**: How to distinguish key-wrapped vaults from legacy password-only vaults?

**Decision**: Increment vault metadata version from 1 to 2. Store version in `storage.VaultMetadata.Version`.

**Rationale**:
- Clear signal for code to choose decryption path
- Enables gradual migration - version 1 vaults continue working
- Simple integer comparison, no complex detection logic

**Version Semantics**:
| Version | Encryption Scheme | Recovery Support |
|---------|-------------------|------------------|
| 1 | Password-derived key directly encrypts vault | Broken (current) |
| 2 | DEK encrypts vault, wrapped by password KEK and recovery KEK | Working |

### Q4: Migration Path for Existing Vaults

**Question**: How should users with existing vaults migrate to the new format?

**Decision**: On-unlock migration prompt. User must know current password. Atomic operation with backup.

**Rationale**:
- Non-disruptive - users choose when to migrate
- Requires password (can't migrate locked vault)
- Must regenerate recovery phrase (old VaultRecoveryKey is incompatible)
- Atomic: temp file → verify → backup old → rename new

**Migration Flow**:
```
1. User unlocks vault (version 1) with password
2. System detects version 1, prompts: "Upgrade to enable recovery phrase?"
3. If yes:
   a. Generate new DEK
   b. Generate new mnemonic + recovery metadata
   c. Display mnemonic for user to write down
   d. Wrap DEK with password KEK and recovery KEK
   e. Re-encrypt vault with DEK
   f. Atomic save as version 2
4. If no: Continue with version 1 (recovery won't work)
```

### Q5: Wrapped DEK Storage Location

**Question**: Where should the wrapped DEK(s) be stored?

**Decision**: Store in vault file metadata (`storage.VaultMetadata`), not in `.meta.json`.

**Rationale**:
- Password-wrapped DEK is essential for vault access - should be in vault file
- Single atomic unit - vault + its decryption metadata together
- `.meta.json` for non-critical metadata (keychain enabled, audit enabled)
- Recovery-wrapped DEK stored in `vault.RecoveryMetadata` (in `.meta.json`)

**Storage Structure**:
```
vault.enc (JSON):
{
  "metadata": {
    "version": 2,
    "salt": "...",
    "iterations": 600000,
    "wrapped_dek": "...",      // NEW: DEK wrapped by password KEK
    "wrapped_dek_nonce": "..." // NEW: GCM nonce for password wrap
  },
  "data": "..." // Encrypted with DEK (unchanged format)
}

.meta.json:
{
  "recovery": {
    "encrypted_recovery_key": "...",  // REPURPOSE: Now stores DEK wrapped by recovery KEK
    "nonce_recovery": "...",           // Nonce for recovery wrap
    ...
  }
}
```

### Q6: Error Handling for Corrupted Wrappers

**Question**: What happens if one wrapper is corrupted but not the other?

**Decision**: Graceful degradation. If password wrapper works, unlock succeeds. Warn about recovery wrapper corruption.

**Rationale**:
- User's primary goal is accessing their credentials
- Don't block access if recovery is broken but password works
- Offer to regenerate recovery after successful password unlock
- Audit log the corruption for security visibility

**Error Matrix**:
| Password Wrapper | Recovery Wrapper | Outcome |
|------------------|------------------|---------|
| Valid | Valid | Normal operation |
| Valid | Corrupted | Unlock succeeds, warn user, offer recovery regeneration |
| Corrupted | Valid | Unlock fails (can't derive password KEK) |
| Corrupted | Corrupted | Unlock fails, suggest restore from backup |

## Implementation Recommendations

1. **New Package**: No new package needed. Add `keywrap.go` to `internal/crypto/`.

2. **Interface Design**:
   ```go
   type KeyWrapper interface {
       WrapKey(dek, kek []byte) (wrapped []byte, nonce []byte, err error)
       UnwrapKey(wrapped, nonce, kek []byte) (dek []byte, err error)
   }
   ```

3. **Memory Safety**: All DEK operations must use `defer crypto.ClearBytes(dek)`.

4. **Test Strategy**:
   - Unit: Wrap/unwrap round-trip, invalid KEK rejection, nonce uniqueness
   - Integration: Full init→recover→change-password cycle

## References

- NIST SP 800-38F: Recommendation for Block Cipher Modes of Operation: Methods for Key Wrapping
- RFC 3394: Advanced Encryption Standard (AES) Key Wrap Algorithm
- Pass-CLI Constitution v1.2.0: Security Requirements, Key Derivation Functions
