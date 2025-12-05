# Feature Specification: Recovery Key Integration

**Feature Branch**: `013-recovery-key-integration`
**Created**: 2025-12-05
**Status**: Draft
**Input**: User description: "Fix recovery key integration - the vault recovery key generated during BIP39 setup is not integrated with vault encryption, so recovery phrase cannot actually unlock the vault to enable password change. Need key-wrapping scheme where vault DEK is recoverable via either master password OR recovery phrase."

## Problem Statement

The current BIP39 recovery system generates a `VaultRecoveryKey` during vault initialization but **never integrates it with vault encryption**. The vault is encrypted solely with a key derived from the master password, making the recovery phrase useless for its intended purpose: allowing users to regain access and change their password when they've forgotten it.

**Current broken flow:**
1. User initializes vault → recovery phrase generated, `VaultRecoveryKey` created
2. Vault encrypted with password-derived key only
3. User forgets password, enters recovery phrase
4. `VaultRecoveryKey` recovered successfully
5. ❌ **FAILS**: Vault cannot be decrypted because it was encrypted with a different key

**Required fix:** Implement key wrapping so the vault's encryption key (DEK) can be recovered via either the master password OR the recovery phrase.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Recover Access with Recovery Phrase (Priority: P1)

A user who has forgotten their master password uses their 24-word recovery phrase to regain access to their vault and set a new password.

**Why this priority**: This is the core functionality that is currently broken. Without this, the recovery phrase feature provides false security - users believe they have a backup but it doesn't actually work.

**Independent Test**: Can be fully tested by creating a vault with recovery enabled, then using `pass-cli change-password --recover` to change the password using only the recovery phrase.

**Acceptance Scenarios**:

1. **Given** a vault with recovery enabled and a written-down 24-word phrase, **When** user runs `pass-cli change-password --recover` and enters the 6 challenge words correctly, **Then** user is prompted for a new password and vault is re-encrypted with the new password.

2. **Given** a vault with recovery enabled, **When** user enters incorrect recovery words, **Then** recovery fails with a clear error message and vault remains unchanged.

3. **Given** a vault with recovery and passphrase (25th word) enabled, **When** user enters correct 6 words but wrong passphrase, **Then** recovery fails with appropriate error message.

---

### User Story 2 - Initialize Vault with Working Recovery (Priority: P1)

A user initializes a new vault and the recovery phrase is properly integrated so it will actually work when needed.

**Why this priority**: Equal to P1 because new vaults must be created correctly for the feature to have any value.

**Independent Test**: Can be tested by initializing a new vault, then immediately testing recovery works before adding any credentials.

**Acceptance Scenarios**:

1. **Given** user runs `pass-cli init`, **When** user completes setup including recovery phrase backup, **Then** vault is encrypted using a key-wrapping scheme that supports both password and recovery access.

2. **Given** a freshly initialized vault, **When** user immediately runs `pass-cli change-password --recover`, **Then** recovery succeeds and user can set a new password.

3. **Given** user runs `pass-cli init --no-recovery`, **When** setup completes, **Then** vault uses traditional password-only encryption (no key wrapping overhead).

---

### User Story 3 - Migrate Existing Vault to Key-Wrapped Format (Priority: P2)

A user with an existing vault (created before this fix) upgrades their vault to the new key-wrapped format so recovery will work.

**Why this priority**: Important for existing users but not blocking - new users get working recovery immediately.

**Independent Test**: Can be tested by loading a vault created with the old format, running migration, then verifying recovery works.

**Acceptance Scenarios**:

1. **Given** a vault in the old format (password-only encryption), **When** user unlocks vault with their password, **Then** system detects old format and offers migration to key-wrapped format.

2. **Given** user accepts migration, **When** migration completes, **Then** vault is re-encrypted with key wrapping and recovery metadata is regenerated.

3. **Given** user declines migration, **When** they continue using vault, **Then** vault continues working with password-only access (recovery will not work).

4. **Given** a vault already in key-wrapped format, **When** user unlocks vault, **Then** no migration is offered.

---

### User Story 4 - Normal Password Unlock Unchanged (Priority: P1)

Users who remember their password experience no change in their daily workflow - the key-wrapping scheme is transparent.

**Why this priority**: Critical that the fix doesn't break the primary use case.

**Independent Test**: Can be tested by unlocking a key-wrapped vault with the correct password.

**Acceptance Scenarios**:

1. **Given** a vault with key wrapping enabled, **When** user unlocks with correct master password, **Then** vault unlocks successfully with no additional prompts or delays.

2. **Given** a vault with key wrapping enabled, **When** user enters wrong password, **Then** vault remains locked with appropriate error message.

---

### Edge Cases

- What happens when recovery metadata is corrupted but vault data is intact?
  - System should detect corruption and warn user; password-based access should still work if possible.

- What happens if user tries recovery on a vault without recovery enabled?
  - Clear error message: "Recovery not enabled for this vault."

- What happens during migration if power loss occurs mid-process?
  - Atomic operations ensure either old or new format is valid; never a corrupted intermediate state.

- What happens if the vault format version is unrecognized?
  - Clear error message with guidance to update pass-cli.

- What happens when DEK is wrapped but one wrapper (password or recovery) fails verification?
  - If password wrapper valid, unlock proceeds; recovery wrapper can be regenerated later.

## Requirements *(mandatory)*

### Functional Requirements

**Key Wrapping Architecture:**

- **FR-001**: System MUST generate a random Data Encryption Key (DEK) to encrypt vault contents.
- **FR-002**: System MUST wrap (encrypt) the DEK using the password-derived key, storing the wrapped key in vault metadata.
- **FR-003**: System MUST wrap (encrypt) the DEK using the recovery-derived key, storing the wrapped key in recovery metadata.
- **FR-004**: System MUST use the same DEK for both password and recovery access paths.

**Vault Initialization:**

- **FR-005**: During vault initialization with recovery enabled, system MUST create both password-wrapped and recovery-wrapped copies of the DEK.
- **FR-006**: System MUST store vault format version to distinguish key-wrapped vaults from legacy password-only vaults.
- **FR-007**: System MUST continue supporting `--no-recovery` flag for password-only vault creation (no key wrapping overhead).

**Vault Unlock (Password Path):**

- **FR-008**: System MUST derive key from password and unwrap the DEK to decrypt vault.
- **FR-009**: Unlock performance MUST remain comparable to current implementation (within 10% of current timing).

**Vault Unlock (Recovery Path):**

- **FR-010**: System MUST derive key from recovery phrase (via challenge words + stored words) and unwrap the DEK.
- **FR-011**: After successful recovery unlock, system MUST prompt user to set a new password.
- **FR-012**: System MUST re-wrap the DEK with the new password-derived key.
- **FR-013**: System MUST preserve the recovery-wrapped DEK unchanged (recovery phrase remains valid).

**Password Change:**

- **FR-014**: When changing password (normal flow), system MUST re-wrap DEK with new password-derived key.
- **FR-015**: System MUST NOT require re-wrapping recovery key during normal password change.

**Migration:**

- **FR-016**: System MUST detect legacy vault format on unlock.
- **FR-017**: System MUST offer optional migration to key-wrapped format.
- **FR-018**: Migration MUST be atomic - either complete successfully or leave vault unchanged.
- **FR-019**: During migration, if recovery was previously enabled, system MUST regenerate recovery phrase for user to write down.

**Security:**

- **FR-020**: DEK MUST be 256-bit random key generated via cryptographically secure RNG.
- **FR-021**: Wrapped keys MUST use authenticated encryption.
- **FR-022**: DEK MUST be cleared from memory immediately after use.
- **FR-023**: System MUST NOT store unwrapped DEK anywhere except in memory during active session.

**Error Handling:**

- **FR-024**: Invalid password MUST produce clear error without revealing whether password or recovery path was attempted.
- **FR-025**: Corrupted recovery metadata MUST NOT prevent password-based access.
- **FR-026**: All errors MUST be logged to audit log (without sensitive data).

### Key Entities

- **Data Encryption Key (DEK)**: 256-bit symmetric key that encrypts vault contents. Never stored directly - only in wrapped form.

- **Password-Wrapped DEK**: DEK encrypted using key derived from master password. Stored in vault file metadata.

- **Recovery-Wrapped DEK**: DEK encrypted using key derived from recovery phrase. Stored in recovery metadata.

- **Vault Format Version**: Integer indicating encryption scheme. Version 1 = legacy password-only, Version 2 = key-wrapped.

- **Key Wrapper**: Authenticated encryption of DEK using a derived key. Contains: encrypted DEK, nonce, authentication tag.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully recover vault access using their 24-word recovery phrase and set a new password.

- **SC-002**: Recovery process completes in under 5 seconds on standard hardware (excluding user input time).

- **SC-003**: Normal password unlock time increases by no more than 10% compared to current implementation.

- **SC-004**: 100% of newly initialized vaults with recovery enabled can be recovered using the recovery phrase.

- **SC-005**: Existing vaults can be migrated to key-wrapped format without data loss.

- **SC-006**: Password changes do not invalidate the recovery phrase (recovery continues to work after password change).

- **SC-007**: All vault operations continue to pass existing security requirements (memory clearing, audit logging, file permissions).

## Assumptions

- The existing BIP39 mnemonic generation and challenge-based recovery word selection are correct and remain unchanged.
- The existing Argon2id KDF for recovery phrase key derivation is adequate.
- Users are willing to re-write their recovery phrase during migration (necessary because the wrapped key changes).
- The current 6-word challenge approach provides sufficient security for recovery key derivation.

## Out of Scope

- Changing the BIP39 mnemonic format or word count (remains 24 words).
- Changing the challenge word count (remains 6 of 24).
- Adding multiple recovery methods (e.g., security questions, trusted contacts).
- Hardware security key support for recovery.
- Cloud backup of recovery keys.
