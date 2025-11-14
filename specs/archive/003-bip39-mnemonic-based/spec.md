# Feature Specification: BIP39 Mnemonic Recovery

**Feature Branch**: `003-bip39-mnemonic-based`
**Created**: 2025-01-13
**Status**: Draft
**Input**: User description: "BIP39 mnemonic-based vault recovery system with 6-word challenge to enable password reset without knowing current master password"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Vault Recovery After Forgotten Password (Priority: P1)

A user forgets their master password and keychain integration is not enabled. Without recovery, their vault would be permanently inaccessible. Using the 24-word recovery phrase they wrote down during initialization, they can reset their master password by providing 6 randomly-selected words from that phrase.

**Why this priority**: Core feature solving the primary problem (permanent lockout). Delivers immediate value by enabling vault recovery in disaster scenarios.

**Dependency**: Requires User Story 2 (recovery setup during init) to be complete before this story can be implemented. Both are P1, but US2 is the foundation.

**Independent Test**: Can be fully tested by initializing a vault with recovery enabled, intentionally "forgetting" the password, and successfully recovering access using the mnemonic phrase. Delivers the value of vault recovery without requiring any other features.

**Acceptance Scenarios**:

1. **Given** user has forgotten their master password and has their 24-word recovery phrase, **When** they run the recovery command and correctly provide the 6 requested words, **Then** the vault unlocks and they can set a new master password
2. **Given** user attempts recovery with incorrect words, **When** they submit the words, **Then** the system displays an error and allows them to retry
3. **Given** user's vault was created with a recovery passphrase, **When** they provide correct words but wrong passphrase, **Then** recovery fails with a clear error message
4. **Given** user provides words that aren't in the standard word list, **When** they enter an invalid word, **Then** the system provides immediate feedback and prompts for retry

---

### User Story 2 - Initial Vault Setup with Recovery (Priority: P1)

When creating a new vault, the user is presented with a 24-word recovery phrase and instructed to write it down securely. The system verifies they've recorded it correctly by asking for 3 random words before completing initialization. This establishes the recovery mechanism for future use.

**Why this priority**: Required foundation for P1 recovery scenario. Without proper setup, recovery cannot work. Must be delivered together with P1.

**Independent Test**: Can be tested by running vault initialization and verifying that a valid 24-word phrase is generated, displayed to the user, verification is performed, and the vault is successfully created with recovery metadata stored.

**Acceptance Scenarios**:

1. **Given** user runs vault initialization, **When** they complete password setup, **Then** the system generates and displays a 24-word recovery phrase with security warnings
2. **Given** user is shown the recovery phrase, **When** they choose to verify their backup, **Then** the system asks for 3 random words and confirms when correct
3. **Given** user fails verification, **When** they enter incorrect words, **Then** the system prompts them to retry with clear error messages
4. **Given** user chooses advanced options, **When** they opt to add a passphrase, **Then** the system prompts for passphrase entry with confirmation and displays a warning about storing this additional secret
5. **Given** user wants to skip recovery entirely, **When** they use the opt-out flag during initialization, **Then** the vault is created without recovery capability and a warning is displayed

---

### User Story 3 - Recovery with Enhanced Security (Priority: P2)

A security-conscious user wants additional protection beyond the 24-word phrase. During vault initialization, they choose to add a passphrase (25th word) which provides an extra layer of security. During recovery, the system automatically detects this requirement and prompts for the passphrase only if it was set.

**Why this priority**: Valuable security enhancement but not required for basic recovery functionality. Can be added after P1 is working.

**Independent Test**: Can be tested by initializing a vault with passphrase protection, performing recovery, and verifying that the passphrase prompt appears and is required for successful unlock.

**Acceptance Scenarios**:

1. **Given** user initializes a vault and chooses passphrase protection, **When** they enter and confirm a passphrase, **Then** the system sets the passphrase requirement flag and displays appropriate warnings
2. **Given** user's vault has passphrase protection enabled, **When** they perform recovery and provide correct words, **Then** the system prompts for the passphrase before unlocking
3. **Given** user's vault does not have passphrase protection, **When** they perform recovery and provide correct words, **Then** the vault unlocks immediately without passphrase prompt

---

### User Story 4 - Skipping Recovery Setup (Priority: P3)

A user who has alternative backup strategies (like always-enabled keychain) wants to skip recovery setup during initialization to save time. They use a command flag to bypass recovery phrase generation and proceed directly to vault creation.

**Why this priority**: Nice-to-have for advanced users with specific workflows. Not critical for primary use cases.

**Independent Test**: Can be tested by initializing a vault with the no-recovery flag and verifying that no recovery phrase is generated, no verification is performed, and the vault is created successfully without recovery metadata.

**Acceptance Scenarios**:

1. **Given** user runs initialization with the no-recovery flag, **When** vault creation completes, **Then** no recovery phrase is displayed and a warning about unrecoverability is shown
2. **Given** user has a vault without recovery enabled, **When** they attempt recovery, **Then** the system displays an error indicating recovery is not available for this vault

---

### User Story 5 - Verification Skip During Setup (Priority: P3)

A user is in a hurry during vault initialization and wants to skip the verification step where they must re-enter 3 random words. They decline verification but receive a strong warning to ensure they've written down all 24 words correctly.

**Why this priority**: Minor convenience feature. Verification is a safety measure, so skipping it is less important than enabling it.

**Independent Test**: Can be tested by initializing a vault and declining verification, confirming that the vault is still created successfully with a warning displayed.

**Acceptance Scenarios**:

1. **Given** user is prompted to verify their recovery phrase backup, **When** they decline verification, **Then** the system displays a warning and continues with vault creation
2. **Given** user accepts verification, **When** they correctly enter the 3 requested words, **Then** verification succeeds and vault creation completes

---

### Edge Cases

- What happens when the user enters a word not in the standard BIP39 word list during recovery?
- How does the system handle recovery attempts on a vault created without recovery enabled?
- What occurs if vault metadata is corrupted and recovery information cannot be read?
- How does the system behave if the user provides correct words but the stored encrypted data cannot be decrypted (data corruption)?
- What happens during initialization if random number generation fails (system entropy unavailable)?
- How does the system handle a user who fails verification 3+ times during initialization?
- What occurs if a user with passphrase protection forgets the passphrase but remembers all 24 words?
- How does the system respond when different recovery positions are asked in different orders across multiple recovery attempts?

## Requirements *(mandatory)*

### Functional Requirements

**Recovery Setup (Initialization)**

- **FR-001**: System MUST generate a valid 24-word recovery phrase using industry-standard word lists during vault initialization
- **FR-002**: System MUST display the 24-word phrase to the user exactly once during initialization with clear instructions to write it down securely
- **FR-003**: System MUST include security warnings that anyone with the phrase can access the vault and that offline storage is required
- **FR-004**: System MUST randomly select 6 positions from the 24-word phrase to use as challenge positions for future recovery attempts
- **FR-005**: System MUST allow users to optionally add a passphrase for enhanced security during initialization
- **FR-006**: System MUST display a warning when passphrase is added, explaining it is an additional secret that must be stored separately
- **FR-007**: System MUST prompt users to verify their backup by entering 3 randomly selected words from the phrase
- **FR-008**: System MUST allow users to skip verification with a warning
- **FR-009**: System MUST allow users to opt out of recovery entirely using a command flag during initialization
- **FR-010**: System MUST display a warning about permanent data loss when recovery is disabled
- **FR-011**: System MUST store recovery metadata (challenge positions, encrypted word data, passphrase requirement flag) securely with the vault

**Recovery Execution**

- **FR-012**: System MUST support a recovery command that allows password reset without knowing the current master password
- **FR-013**: System MUST prompt for 6 specific words from the 24-word phrase during recovery
- **FR-014**: System MUST randomize the order in which the 6 challenge positions are requested during each recovery attempt
- **FR-015**: System MUST provide immediate feedback if a user enters a word not in the standard word list
- **FR-016**: System MUST allow retry for individual word entry if validation fails
- **FR-017**: System MUST display progress indicators showing which word number is being entered (e.g., "3/6")
- **FR-018**: System MUST detect whether passphrase protection is enabled and prompt for passphrase only if required
- **FR-019**: System MUST unlock the vault when correct words (and passphrase if required) are provided
- **FR-020**: System MUST prompt for a new master password after successful recovery
- **FR-021**: System MUST re-encrypt the vault with the new master password
- **FR-022**: System MUST display clear error messages when recovery fails due to incorrect words or passphrase
- **FR-023**: System MUST allow users to retry recovery after failure
- **FR-024**: System MUST prevent recovery attempts on vaults that were created without recovery enabled

**Security**

- **FR-025**: System MUST never store the 24-word recovery phrase in any form (only encrypted derived data)
- **FR-026**: System MUST never log recovery words, passphrases, or master passwords
- **FR-027**: System MUST clear all sensitive data from memory immediately after use
- **FR-028**: System MUST log recovery attempts (success and failure) in audit logs without sensitive data
- **FR-029**: System MUST use cryptographically secure random number generation for all random operations (word selection, position selection, encryption)
- **FR-030**: System MUST validate the mathematical checksum embedded in the 24-word phrase before accepting it for recovery

**Error Handling**

- **FR-031**: System MUST display user-friendly error messages for all failure scenarios (invalid words, wrong recovery data, missing vault, corrupted metadata)
- **FR-032**: System MUST handle entropy generation failures gracefully with clear error messages
- **FR-033**: System MUST detect and report metadata corruption before attempting recovery operations

### Key Entities

- **Recovery Phrase**: A 24-word mnemonic generated from cryptographically random data, using a standard word list. Represents the master recovery credential for the vault.

- **Recovery Passphrase**: An optional additional secret (25th word) that enhances security. When set, both the recovery phrase and passphrase are required for successful recovery.

- **Challenge Positions**: A set of 6 positions (indices) randomly selected from the 24 words during initialization. These positions remain fixed for the lifetime of the vault and determine which words are requested during recovery.

- **Encrypted Stored Words**: The 18 words from the recovery phrase that are not challenge words, encrypted and stored with the vault metadata. These are decrypted during recovery to reconstruct the full 24-word phrase.

- **Recovery Metadata**: Data structure stored with the vault containing challenge positions, encrypted words, passphrase requirement flag, encryption parameters, and cryptographic nonces.

- **Vault Recovery Key**: A random encryption key that can unlock the vault, itself encrypted using the full 24-word recovery phrase. This is the actual key that unlocks vault credentials.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can successfully recover vault access by providing 6 correct words from their 24-word phrase in under 30 seconds of interaction time
- **SC-002**: Recovery phrase generation completes in under 2 seconds during vault initialization
- **SC-003**: 100% of valid recovery phrases (correct words and passphrase) successfully unlock the vault
- **SC-004**: 100% of invalid recovery attempts (wrong words or passphrase) are rejected with clear error messages
- **SC-005**: Users who skip verification can still successfully recover their vault if they wrote down the phrase correctly
- **SC-006**: Vaults created without recovery display appropriate error messages when recovery is attempted
- **SC-007**: Recovery metadata adds less than 520 bytes to vault file size (~511 bytes actual)
- **SC-008**: All recovery operations complete without leaking sensitive data to logs or memory dumps
- **SC-009**: Random word selection during verification and recovery produces different sequences on each attempt (verifiable over 10+ attempts)
- **SC-010**: Passphrases are only prompted when required, never when vault was created without passphrase protection

### User Experience Goals

- Users understand recovery phrase importance through clear warnings during initialization
- Users can complete recovery without consulting documentation (clear prompts guide them)
- Users receive immediate feedback on word entry errors (invalid words caught before submission)
- Users experience no ambiguity about whether passphrase is required (automatic detection)

## Assumptions

- Users have a secure way to store 24 words offline (paper, secure note, safe)
- Users understand basic security concepts (not sharing phrases, offline storage)
- Vault file remains intact and uncorrupted for recovery to work
- System has access to cryptographically secure random number generation
- Users can accurately transcribe and read back 24 words from their backup
- Standard BIP39 word list remains stable and compatible with verification tools
- Recovery is a rare operation (design optimizes for security over recovery speed)
- Users who enable passphrase protection understand they're creating an additional secret to manage

## Dependencies

- Requires standard BIP39 word list (2,048 words)
- Requires cryptographic library supporting secure random generation, key derivation, and AES-GCM encryption
- Requires existing vault metadata structure to be extended with recovery fields
- Requires existing password change functionality to integrate with recovery flow
- Requires existing audit logging system to record recovery events

## Scope

### In Scope

- Recovery phrase generation during vault initialization
- Recovery phrase display and user warnings
- Optional passphrase protection (25th word)
- Backup verification during initialization (3-word challenge)
- Recovery command with 6-word challenge (randomized order)
- Automatic passphrase detection and prompting
- Password reset after successful recovery
- Opt-out capability for users who don't want recovery
- Error handling for invalid words, wrong phrases, corrupted data
- Audit logging of recovery events
- Security measures (no storage of phrase, memory clearing, encrypted metadata)

### Out of Scope

- Migration tool for adding recovery to existing vaults (deferred to future enhancement)
- Recovery phrase rotation/regeneration (deferred to future enhancement)
- Ability to disable recovery after vault creation (deferred to future enhancement)
- Multi-language word lists beyond English (deferred to future enhancement)
- QR code generation for recovery phrase backup (deferred to future enhancement)
- Integration with hardware security devices (deferred to future enhancement)
- Social recovery (splitting phrase among trusted contacts) (deferred to future enhancement)
- Automatic cloud backup of recovery phrase (explicitly excluded for security reasons)

## Open Questions

None - all design decisions finalized during specification discussion.
