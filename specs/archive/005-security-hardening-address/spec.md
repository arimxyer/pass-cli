# Feature Specification: Security Hardening

**Feature Branch**: `005-security-hardening-address`
**Created**: 2025-01-11
**Status**: Draft
**Input**: User description: "Security hardening: address memory security, crypto parameters, password policy, and audit logging vulnerabilities"

## Clarifications

### Session 2025-10-11

- Q: FR-017 real-time strength indicator - Where should it be implemented (CLI, TUI, both, or neither)? → A: Both CLI and TUI with mode-appropriate UX patterns
- Q: Does current `gopass` library support byte-array (`[]byte`) password input, or is migration to `golang.org/x/term` required? → A: Research needed - requires codebase inspection during planning phase
- Q: When a password change migration to 600k iterations fails mid-operation (e.g., power loss during vault re-encryption), what should the recovery behavior be? → A: Option B - Implement atomic write with backup; auto-rollback on next unlock if migration incomplete; include pre-flight checks (disk space, write permissions) and retain backup until next successful unlock
- Q: FR-027 requires HMAC signatures to protect audit log integrity, but the key derivation method is unspecified. Where should the HMAC key come from? → A: Option D - Use OS keychain (Windows Credential Manager, macOS Keychain, Linux Secret Service) with per-vault HMAC keys identified by vault UUID; allows log verification without master password; gracefully disable audit logging if keychain unavailable; use github.com/99designs/keyring library
- Q: When a user fails password complexity validation during vault initialization, what should the retry behavior be? → A: Option B - Apply rate limiting after 3 failed attempts (5-second cooldown between attempts)
- Q: FR-029 specifies 10MB default threshold for audit log rotation, but retention duration is undefined. How long should old audit logs be kept? → A: Option A - Keep last 7 days of logs (delete older rotated logs automatically)
- Q: FR-004 requires memory clearing with `crypto/subtle`, but doesn't specify how to validate this behavior in tests. What testing approach should be used? → A: Option B - Targeted unit tests that verify specific byte slices are zeroed after use (not heap scanning). Test retains reference to password slice passed to function, then verifies slice is zeroed after function completes. Can force GC with runtime.GC() for reliability. Supplemented by manual profiling during development and code review for defer patterns.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Protected Master Password in Memory (Priority: P1) 🎯 MVP

Users expect their master password to be securely handled in memory and not left exposed to memory inspection attacks. When the application handles the master password for vault operations, it should be immediately cleared from memory after use, preventing extraction via memory dumps or debuggers.

**Why this priority**: This is a CRITICAL security vulnerability (High severity). An attacker with local access who can inspect process memory can extract the plaintext master password, leading to complete vault compromise. This undermines the entire security model of the password manager.

**Independent Test**: Can be fully tested using targeted unit tests that retain a reference to the password byte slice passed to the function under test, then verify the slice is zeroed after execution (not heap scanning). Tests can force GC with `runtime.GC()` to increase reliability. Supplemented by manual memory profiling (debuggers like delve, pprof) during development and mandatory code review to verify defer patterns.

**Acceptance Scenarios**:

1. **Given** a user initializes a new vault with a master password, **When** the password is used for key derivation, **Then** the password data is immediately cleared from memory after the encryption key is generated
2. **Given** a user unlocks an existing vault, **When** the master password is entered and used for decryption, **Then** the password bytes are zeroed in memory within milliseconds of successful vault unlock
3. **Given** a vault operation completes (lock, change password, save), **When** memory is inspected after the operation, **Then** no plaintext master password remains in heap or stack memory
4. **Given** the application crashes or is terminated, **When** a core dump is analyzed, **Then** master password data should not be recoverable in plaintext form

---

### User Story 2 - Stronger Brute-Force Protection (Priority: P1) 🎯 MVP

Users storing sensitive credentials need protection against offline brute-force attacks if their vault file is stolen. When an attacker obtains the encrypted vault file, the key derivation function should make password cracking computationally expensive enough to render attacks impractical with current hardware.

**Why this priority**: This is a HIGH priority security issue (Medium severity). The current 100,000 PBKDF2 iterations are 6x weaker than OWASP 2023 recommendations (600,000+). This significantly reduces the time required to crack passwords, especially with modern GPU-based attacks. For a password manager, this is critical infrastructure.

**Independent Test**: Can be fully tested by benchmarking key derivation time on reference hardware (should take 500-1000ms), verifying the iteration count is stored correctly in vault metadata, and confirming that existing vaults can be migrated to the new iteration count during password changes.

**Acceptance Scenarios**:

1. **Given** a user creates a new vault, **When** the master password is set, **Then** key derivation uses a minimum of 600,000 PBKDF2-SHA256 iterations
2. **Given** a user with an existing vault (100k iterations), **When** they change their master password, **Then** the vault is automatically re-encrypted with 600,000 iterations
3. **Given** a vault is being initialized, **When** key derivation completes, **Then** the operation takes between 500-1000ms on modern consumer hardware (calibrated delay)
4. **Given** an attacker attempts to brute-force a stolen vault file, **When** testing 1 million password candidates per second on a high-end GPU, **Then** the attack should take months or years to exhaust common password spaces

---

### User Story 3 - Enforced Strong Password Requirements (Priority: P2)

Users may not understand password security best practices and could choose weak master passwords (like "password123"). When initializing a vault or changing the master password, the application should enforce minimum complexity requirements to prevent easily guessable passwords.

**Why this priority**: While cryptography protects the vault, a weak master password undermines all other security measures. This is rated P2 because it prevents user-caused vulnerabilities but doesn't fix existing code flaws. It should be implemented after P1 memory/crypto fixes but before nice-to-have features.

**Independent Test**: Can be fully tested by attempting to create vaults with various password combinations (weak, medium, strong) and verifying that only passwords meeting complexity requirements (12+ chars, mixed case, digits, symbols) are accepted, with clear error messages explaining what's missing.

**Acceptance Scenarios**:

1. **Given** a user attempts to initialize a vault with "password", **When** the password is submitted, **Then** the system rejects it with a message: "Password must be at least 12 characters and contain uppercase, lowercase, digit, and symbol"
2. **Given** a user attempts to set password "MyPassword123!", **When** the password is validated, **Then** the system accepts it as it meets all complexity requirements
3. **Given** a user enters a password during vault initialization, **When** the password field is active, **Then** a real-time strength indicator shows current password strength (weak/medium/strong) and missing requirements
4. **Given** a user with an existing vault (created with old 8-char policy), **When** they change their password, **Then** the new password must meet the updated complexity requirements

---

### User Story 4 - Audit Trail for Security Monitoring (Priority: P3)

Security-conscious users and organizations need visibility into when and how credentials are accessed. When credentials are retrieved, modified, or deleted, and when vault operations occur (unlock, lock, password changes), these events should be logged with timestamps to enable forensic analysis if unauthorized access is suspected.

**Why this priority**: This is an INFO-level security improvement that adds defense-in-depth. It doesn't fix an active vulnerability but provides detection capabilities. Rated P3 because it's valuable but not critical for basic security - users can function without it, and it should be implemented after addressing the critical memory/crypto/password issues.

**Independent Test**: Can be fully tested by performing various vault operations (unlock, get credential, add credential, delete) and verifying that each operation appears in the audit log with correct timestamps, operation type, and relevant context (credential name but NOT password values), and that the log cannot be tampered with after creation.

**Acceptance Scenarios**:

1. **Given** audit logging is enabled, **When** a user unlocks the vault, **Then** an entry is logged with: timestamp, event type "vault_unlock", success/failure status
2. **Given** a user retrieves a credential, **When** the "get" command is executed, **Then** an entry is logged with: timestamp, event type "credential_access", credential service name (NOT the password)
3. **Given** a credential is modified or deleted, **When** the operation completes, **Then** an entry is logged with: timestamp, event type, credential name, operation success/failure
4. **Given** an audit log exists, **When** an attacker attempts to modify or delete log entries, **Then** the log integrity check (HMAC signature) detects the tampering and alerts the user
5. **Given** audit logging is enabled, **When** log file size exceeds a threshold, **Then** logs are rotated automatically with retention of the most recent entries

---

### Edge Cases

- What happens when a user migrates from an old vault (100k iterations) to the new standard (600k iterations)? The migration should be automatic and seamless during the next password change operation, with no data loss. Migration uses atomic write pattern: write to temporary file, then rename. If migration fails mid-operation (power loss, crash, disk full), auto-rollback restores original vault on next unlock. Pre-flight checks verify disk space and write permissions before attempting migration. Backup file (vault.bak) retained until next successful unlock for additional safety.
- How does the system handle memory clearing if the Go garbage collector runs before manual clearing? The implementation should use byte slices exclusively for passwords and clear them immediately after use, minimizing GC exposure window.
- What if a user's system doesn't have write permissions for the audit log directory? The application should degrade gracefully, warn the user, and continue functioning without audit logging (or use an alternate location).
- How does password complexity enforcement handle non-English characters or special Unicode symbols? The validation should count them appropriately (symbols count as symbols, accented letters count as letters) and document which character classes are required.
- How does audit log rotation handle very high-frequency operations? The rotation should be based on file size limits (e.g., 10MB) with configurable retention, not just operation count. Rotated logs older than 7 days are automatically deleted to prevent unbounded disk usage.
- What happens if OS keychain is unavailable for audit HMAC keys (e.g., headless SSH session)? The system should gracefully disable audit logging with a warning message to the user. Since audit logging is P3 (optional, defense-in-depth), this degradation is acceptable.
- How are audit HMAC keys managed across multiple vaults? Each vault has a unique HMAC key stored in OS keychain with identifier `pass-cli-audit-<vault-uuid>`. Vault UUID is stored in vault metadata. This ensures each vault + audit log pair is portable and self-contained.
- What happens when a user repeatedly fails password complexity validation during vault initialization? The first 3 attempts allow immediate retries. After 3 failures, a 5-second cooldown is enforced between subsequent attempts. This prevents automated abuse while allowing legitimate users to retry as they learn the requirements.

## Requirements *(mandatory)*

### Functional Requirements

#### Memory Security (P1)

- **FR-001**: System MUST handle master password as byte array (`[]byte`) from the moment of input, never converting to immutable Go string types
- **FR-002**: System MUST clear master password byte arrays from memory immediately after key derivation using secure zeroing (overwriting with zeros)
- **FR-003**: System MUST use deferred cleanup handlers to ensure password clearing occurs even if errors or panics happen during vault operations
- **FR-004**: System MUST use `crypto/subtle.ConstantTimeCompare` or equivalent to prevent compiler optimizations from removing memory clearing operations
- **FR-005**: Password input functions MUST return `[]byte` instead of `string` to enable secure memory clearing

#### Cryptographic Hardening (P1)

- **FR-006**: System MUST use a minimum of 600,000 PBKDF2-SHA256 iterations for key derivation (OWASP 2023 recommendation)
- **FR-007**: System MUST store the iteration count in vault metadata to support future iteration increases
- **FR-008**: System MUST provide automatic vault migration to higher iteration counts when users change their master password
- **FR-009**: Key derivation MUST take 500-1000ms on modern consumer hardware to balance security and usability
- **FR-010**: System MUST support configurable iteration counts for future-proofing (with 600k as minimum)
- **FR-011**: System MUST implement atomic vault migration using write-to-temp-then-rename pattern to prevent corruption during password changes
- **FR-012**: System MUST perform pre-flight checks (disk space, write permissions) before attempting vault migration
- **FR-013**: System MUST auto-rollback to original vault on next unlock if migration fails mid-operation
- **FR-014**: System MUST retain backup file (vault.bak) until next successful unlock after migration for additional safety
- **FR-015**: System MUST notify user when migration rollback occurs, explaining the failure and suggesting retry

#### Password Policy (P2)

- **FR-016**: System MUST enforce minimum master password length of 12 characters (increased from 8)
- **FR-017**: System MUST require master passwords to contain at least one uppercase letter (A-Z)
- **FR-018**: System MUST require master passwords to contain at least one lowercase letter (a-z)
- **FR-019**: System MUST require master passwords to contain at least one digit (0-9)
- **FR-020**: System MUST require master passwords to contain at least one symbol (punctuation or special character)
- **FR-021**: System MUST provide clear, actionable error messages when password requirements are not met, listing which requirements are missing
- **FR-022**: System SHOULD provide real-time password strength feedback during input in both CLI and TUI modes (CLI: text-based indicator updated per keystroke; TUI: visual strength meter component)
- **FR-023**: Existing vaults with 8-character passwords MUST continue to function but MUST require new complexity standards when password is changed
- **FR-024**: System MUST allow unlimited immediate retries for the first 3 password complexity validation failures, then enforce a 5-second cooldown between subsequent attempts to prevent automated abuse

#### Audit Logging (P3)

- **FR-025**: System MUST log all vault unlock attempts (successful and failed) with timestamp and outcome
- **FR-026**: System MUST log all credential access operations with timestamp, operation type (get/add/update/delete), and credential service name
- **FR-027**: System MUST NOT log credential passwords or sensitive field values in audit logs
- **FR-028**: System MUST protect audit log integrity using HMAC signatures to detect tampering
- **FR-029**: System MUST support configurable audit log location (default: `~/.pass-cli/audit.log`)
- **FR-030**: System MUST rotate audit logs when size exceeds configurable threshold (default: 10MB)
- **FR-031**: System MUST automatically delete rotated audit logs older than 7 days to prevent unbounded disk usage
- **FR-032**: Audit logging MUST be optional and disabled by default to avoid unexpected disk usage
- **FR-033**: System MUST degrade gracefully if audit logging fails (warn but continue operation)
- **FR-034**: System MUST generate unique HMAC key per vault and store in OS keychain (Windows Credential Manager, macOS Keychain, Linux Secret Service) using identifier `pass-cli-audit-<vault-uuid>`
- **FR-035**: System MUST support audit log verification without requiring master password or vault unlock
- **FR-036**: System MUST disable audit logging gracefully with warning if OS keychain is unavailable (e.g., headless environment)
- **FR-037**: System MUST store vault UUID in vault metadata to enable keychain key retrieval

### Key Entities

- **AuditLogEntry**: Represents a single security event with timestamp, event type (unlock/access/modify), outcome (success/failure), credential service name (not password), and HMAC signature for integrity. HMAC key is per-vault, stored in OS keychain.
- **PasswordPolicy**: Configuration defining minimum length (12), required character classes (uppercase, lowercase, digit, symbol), and validation rules
- **VaultMetadata**: Extended to include PBKDF2 iteration count and unique vault UUID (for keychain HMAC key identification), enabling version detection, migration, and audit log integrity verification

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Memory inspection tools (debuggers, memory profilers) MUST NOT find plaintext master password in process memory 1 second after vault operations complete
- **SC-002**: Key derivation for vault operations MUST take between 500-1000ms on reference hardware (modern consumer CPU) to ensure adequate brute-force protection
- **SC-003**: New vaults MUST reject passwords shorter than 12 characters or lacking required complexity (uppercase, lowercase, digit, symbol)
- **SC-004**: Brute-force attack against a stolen vault file (at 1 million attempts/second on high-end GPU) MUST take a minimum of 10 minutes per password candidate to verify (600k iterations * cost)
- **SC-005**: Audit logging (when enabled) MUST record 100% of vault operations (unlock, access, modify) with tamper-evident signatures
- **SC-006**: Existing vaults created with old security parameters MUST automatically migrate to new standards (600k iterations, enforced complexity) during password change operations without data loss
- **SC-007**: All security hardening changes MUST maintain backward compatibility with existing vault files until password change migration occurs

## Assumptions *(documented)*

1. **Hardware Baseline**: Success criteria timing (500-1000ms key derivation) assumes modern consumer CPUs (Intel Core i5/AMD Ryzen 5 or better from 2018+). Older hardware may experience longer delays but this is acceptable for security.

2. **Backward Compatibility Strategy**: Existing vaults will NOT be forcibly migrated on application upgrade. Migration to 600k iterations happens automatically when users voluntarily change their master password. This allows users to opt-in to the security improvements while maintaining access to existing vaults.

3. **Audit Logging Default**: Audit logging is disabled by default to avoid unexpected disk usage and maintain backward compatibility. Security-conscious users can enable it via configuration flag (e.g., `--enable-audit` or config file setting).

4. **Memory Security Limitations**: While byte array clearing significantly improves security, Go's garbage collector may still have copies of data in freed memory regions until overwritten. The implementation minimizes but cannot completely eliminate this risk without low-level memory management unavailable in Go.

5. **Unicode Password Support**: Password complexity validation counts Unicode letters (accented characters) as letters and Unicode symbols as symbols. This supports international users while maintaining security requirements.

6. **Testing Environment**: Security testing (memory inspection, brute-force timing) will be performed on Windows, macOS, and Linux to ensure cross-platform consistency, as memory management and crypto libraries may behave differently across platforms.

7. **Migration Path**: Users who forget to change their password will continue using 100k iterations indefinitely. This is acceptable as forcing immediate migration would risk vault access issues. Documentation will encourage password changes to benefit from improvements.

## Out of Scope *(explicitly excluded)*

1. **Multi-Factor Authentication (MFA)**: Adding second-factor authentication (TOTP, hardware keys) is out of scope. This would require significant UX changes and external dependencies. Consider for future release.

2. **Argon2id Migration**: While Argon2id is more resistant to GPU attacks than PBKDF2, migrating to it requires new dependencies and breaks vault format compatibility. Increasing PBKDF2 iterations provides immediate security improvement without breaking changes. Argon2id can be considered for v2.0.

3. **Encrypted Audit Logs**: Audit logs are stored in plaintext (though credentials are not included). Encrypting logs adds complexity and key management challenges. If needed, users can enable full-disk encryption at OS level.

4. **Network-Based Rate Limiting**: The rate limiting is local (per-device) only. Synchronized rate limiting across multiple devices would require a backend service, which conflicts with pass-cli's offline-first design.

5. **Password Strength Estimation (zxcvbn)**: Real-time password strength estimation using advanced algorithms (like zxcvbn) is out of scope. The implementation uses simpler character class requirements which are sufficient and don't require external dependencies.

6. **Automated Security Scanning**: Integration with CVE databases or automated dependency scanning (govulncheck) is not included in this feature. These are development/CI concerns, not runtime features.

7. **Biometric Authentication**: Using fingerprint/face recognition for vault unlock is out of scope due to platform-specific APIs and varying security models. System keychain integration already provides some convenience.

## Dependencies

- **Go Standard Library**: `crypto/rand`, `crypto/subtle`, `crypto/sha256` for secure operations
- **Existing Crypto Package**: `internal/crypto/crypto.go` requires refactoring to accept `[]byte` passwords
- **Existing Vault Package**: `internal/vault/vault.go` requires changes to use byte arrays and enforce new password policy
- **Terminal Input Library**: Requires investigation during planning - determine if current `gopass` library supports `[]byte` return values or if migration to `golang.org/x/term` is necessary for FR-001/FR-005 compliance
- **OS Keychain Library**: `github.com/99designs/keyring` for cross-platform OS keychain integration (Windows Credential Manager, macOS Keychain, Linux Secret Service) to store per-vault HMAC keys for audit log integrity
- **Backward Compatibility**: Must maintain ability to read vaults created with 100k iterations until migration

## Risks

1. **User Experience Impact**: Increasing key derivation time from ~100ms to ~1000ms (10x) will make vault unlock noticeably slower. Mitigation: Clearly communicate this is a security improvement, make iteration count configurable for power users.

2. **Migration Friction**: Users with weak passwords (8 chars, simple) will be forced to choose new passwords when upgrading. Mitigation: Only enforce on password change, not on unlock. Provide clear migration messaging.

3. **Breaking Changes**: While vaults remain compatible, the password input API changes from string to []byte might affect TUI integration. Mitigation: Test thoroughly with both CLI and TUI modes.

4. **Memory Security Limitations**: Even with byte array clearing, Go's GC may leave copies in freed memory. Cannot achieve military-grade memory security without dropping to unsafe or C bindings. Mitigation: Document limitations, focus on "best effort" security.

5. **Cross-Platform Consistency**: Memory clearing effectiveness may vary across Windows/macOS/Linux due to different memory allocators. Mitigation: Test on all platforms, use Go's `crypto/subtle` which is designed for cross-platform constant-time ops.

6. **Audit Log Storage**: Audit logs could grow unbounded if rotation fails. Mitigation: Implement size-based rotation with configurable limits, make logging optional.

## Notes

- **Backward Compatibility**: This feature maintains vault file compatibility. Old vaults work with new code, new vaults work with new code. Old vaults get security upgrades only when password is changed.

- **Testing Strategy**: Security features require specialized testing:
  - Memory clearing: Targeted unit tests that retain reference to password byte slice passed to functions under test, then verify slice is zeroed after execution. Tests can force GC with `runtime.GC()` to increase reliability. Supplemented by manual profiling (delve, pprof) during development and mandatory code review for defer patterns.
  - Crypto timing: Benchmark on representative hardware to verify 500-1000ms target
  - Password validation: Unit tests for all complexity rules and edge cases
  - Audit logging: Verify HMAC integrity and tamper detection

- **Documentation Impact**: README, SECURITY.md, and USAGE.md will need updates explaining:
  - Why vault unlock is slower (security improvement)
  - New password requirements
  - How to enable audit logging
  - Migration path for existing vaults

- **Versioning**: This is a minor version bump (e.g., 0.1.0 → 0.2.0) as it changes security behavior but maintains backward compatibility. Mark as "Security Release" in release notes.
