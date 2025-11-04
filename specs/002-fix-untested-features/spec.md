# Feature Specification: Fix Untested Features and Complete Test Coverage

**Feature Branch**: `002-fix-untested-features`
**Created**: 2025-11-04
**Status**: Draft
**Input**: User description: "Fix keychain enable/status/vault remove commands that have broken functionality hidden by 25 skipped tests with TODO markers. Ensure all advertised functionality has complete, passing test coverage."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Keychain Integration Must Actually Work (Priority: P1)

As a user who has enabled keychain integration for my vault, when I open the TUI or run CLI commands, the system should automatically retrieve my master password from the keychain without prompting me, so I can access my credentials seamlessly.

**Why this priority**: This is the most critical issue - users are explicitly enabling keychain integration and seeing success messages, but the feature is completely non-functional. This is a severe quality issue where advertised functionality doesn't work.

**Independent Test**: Can be fully tested by: (1) creating a vault, (2) running `pass-cli keychain enable`, (3) running `pass-cli tui` and verifying NO password prompt appears and delivers seamless vault access via keychain.

**Acceptance Scenarios**:

1. **Given** an existing vault initialized without keychain, **When** I run `pass-cli keychain enable` and provide my master password, **Then** the vault metadata is updated to indicate keychain usage and my password is stored in the system keychain
2. **Given** a vault with keychain enabled, **When** I run `pass-cli tui`, **Then** the TUI opens without prompting for password and successfully unlocks the vault using the keychain
3. **Given** a vault with keychain enabled, **When** I run any CLI command (get/list/add), **Then** the command executes without password prompt using keychain authentication
4. **Given** a vault initialized with `--use-keychain` flag, **When** I check the vault metadata, **Then** keychain usage is properly recorded

---

### User Story 2 - Vault Removal Must Be Complete (Priority: P2)

As a user who wants to remove a vault completely, when I run `pass-cli vault remove [vault-path]`, the system should delete the vault file, metadata, and keychain entry (if present), so I can cleanly remove all traces of the vault from my system.

**Why this priority**: The command skeleton exists but is not implemented. Users see "vault remove" in help output but it doesn't work. This is moderate priority because users can manually delete files as a workaround.

**Independent Test**: Can be fully tested by: (1) creating a vault with keychain enabled, (2) running `pass-cli vault remove [vault-path] --yes`, (3) verifying vault file, metadata, and keychain entry are all deleted.

**Acceptance Scenarios**:

1. **Given** a vault with keychain enabled, **When** I run `vault remove [path] --yes`, **Then** the vault file, metadata file, and keychain entry are all deleted
2. **Given** a vault without keychain, **When** I run `vault remove [path] --yes`, **Then** the vault file and metadata file are deleted
3. **Given** a vault with metadata (audit enabled), **When** I run `vault remove [path]`, **Then** audit log entries are written (attempt + success) before metadata deletion
4. **Given** a vault file that doesn't exist but keychain entry exists, **When** I run `vault remove [path] --yes`, **Then** the orphaned keychain entry is still cleaned up with a warning about missing file
5. **Given** any vault, **When** I run `vault remove [path]` without --yes flag, **Then** I am prompted to confirm removal before proceeding
6. **Given** vault remove command run 20 times, **When** measuring success rate, **Then** at least 95% succeed (vault + metadata + keychain all deleted)

---

### User Story 3 - Keychain Status Must Report Accurately (Priority: P3)

As a user checking my keychain integration status, when I run `pass-cli keychain status`, the system should accurately report whether keychain is available, whether my vault uses keychain, and provide actionable suggestions, so I understand my security configuration.

**Why this priority**: This is informational functionality - lower priority than fixing core features. Users can still use the system without accurate status reporting.

**Independent Test**: Can be fully tested by: (1) checking status before enable (shows "not enabled"), (2) running enable, (3) checking status after enable (shows "enabled" with backend name).

**Acceptance Scenarios**:

1. **Given** a vault without keychain enabled, **When** I run `keychain status`, **Then** output shows "Available: Yes", "Password Stored: No", and suggests running `pass-cli keychain enable`
2. **Given** a vault with keychain enabled, **When** I run `keychain status`, **Then** output shows "Available: Yes", "Password Stored: Yes", and displays backend name (wincred/macOS Keychain/Linux Secret Service)
3. **Given** a vault with audit enabled, **When** I run `keychain status`, **Then** an audit entry is written with event_type "keychain_status"
4. **Given** a system where keychain is unavailable, **When** I run `keychain status`, **Then** exit code is 0 and output explains keychain is not available on this system

---

### User Story 4 - All Tests Must Pass Without Skipping (Priority: P1)

As a developer/maintainer, when I run the full test suite, all tests should execute and pass without any TODO markers or feature skips, so I can trust that all advertised functionality is verified and working before releases.

**Why this priority**: This is P1 because it's a quality control issue - broken features were hidden by skipped tests. This could have been caught before reaching users if tests were complete.

**Independent Test**: Can be fully tested by: running `go test -v -tags=integration ./test` and verifying 0 tests contain `t.Skip("TODO:` in output.

**Acceptance Scenarios**:

1. **Given** the codebase after this spec implementation, **When** I run `go test -v -tags=integration ./test`, **Then** 0 tests are skipped with TODO markers
2. **Given** the full test suite, **When** I run tests, **Then** `TestIntegration_KeychainEnable` subtests execute and pass (currently 3 skipped)
3. **Given** the full test suite, **When** I run tests, **Then** `TestIntegration_KeychainStatus` subtests execute and pass (currently 3 skipped)
4. **Given** the full test suite, **When** I run tests, **Then** `TestIntegration_VaultRemove` subtests execute and pass (currently 5 skipped)
5. **Given** any new feature development, **When** merging to main, **Then** CI blocks merge if any test contains `t.Skip("TODO:`

---

### Edge Cases

- What happens when keychain enable is run on a vault already using keychain? (Should be idempotent or provide clear message)
- What happens when vault remove is run on a vault that's currently open/locked? (Should handle gracefully)
- What happens when keychain password exists but vault file is deleted? (vault remove should clean up orphaned keychain)
- What happens when vault file exists but keychain password doesn't? (Should fall back to password prompt)
- What happens when vault metadata indicates keychain but keychain password is missing? (Should detect mismatch and provide clear error)
- What happens when multiple vaults use the same keychain entry? (Current design uses single service/account - is this intended?)
- What happens when vault remove is cancelled at confirmation prompt? (Should exit cleanly with no changes)
- What happens when system keychain is locked/unavailable during enable? (Should provide platform-specific guidance)

## Requirements *(mandatory)*

### Functional Requirements

**Keychain Enable (cmd/keychain_enable.go)**:
- **FR-001**: System MUST update vault metadata file (.meta.json) when `keychain enable` runs successfully
- **FR-002**: System MUST set metadata field indicating keychain is enabled for this vault
- **FR-003**: `keychain enable` MUST verify password is correct before storing in keychain
- **FR-004**: `keychain enable` MUST handle vaults already using keychain (idempotent or clear error)
- **FR-005**: `keychain enable --force` MUST re-enable keychain even if already enabled (overwrites password)
- **FR-006**: `keychain enable` MUST write audit log entries when vault has audit enabled

**Keychain Status (cmd/keychain_status.go)**:
- **FR-007**: `keychain status` MUST check both keychain service AND vault metadata for consistency
- **FR-008**: `keychain status` MUST exit with code 0 even when keychain unavailable (informational command)
- **FR-009**: `keychain status` MUST display platform-specific backend name (wincred/macOS Keychain/Linux Secret Service)
- **FR-010**: `keychain status` MUST provide actionable suggestions when keychain not enabled
- **FR-011**: `keychain status` MUST write audit log entry when vault has audit enabled

**Vault Remove (cmd/vault_remove.go)**:
- **FR-012**: `vault remove` MUST delete vault file, metadata file, and keychain entry in single operation
- **FR-013**: `vault remove` MUST clean up orphaned keychain entries even when vault file missing
- **FR-014**: `vault remove` MUST prompt for confirmation unless --yes flag provided
- **FR-015**: `vault remove` MUST write audit entries (attempt + success) when vault has audit enabled
- **FR-016**: `vault remove` MUST write audit entries BEFORE deleting metadata file
- **FR-017**: `vault remove` MUST achieve 95% success rate across 20 consecutive runs (SC-003)

**Vault Service (internal/vault/vault.go)**:
- **FR-018**: `UnlockWithKeychain()` MUST check vault metadata to determine if keychain is enabled
- **FR-019**: `UnlockWithKeychain()` MUST return specific error when vault doesn't use keychain
- **FR-020**: `UnlockWithKeychain()` MUST return specific error when keychain password missing
- **FR-021**: Vault initialization with `--use-keychain` MUST set keychain flag in metadata

**TUI (cmd/tui/main.go)**:
- **FR-022**: TUI MUST attempt keychain unlock if vault metadata indicates keychain usage
- **FR-023**: TUI MUST fall back to password prompt only when keychain unlock fails
- **FR-024**: TUI MUST display clear error message when keychain unlock fails (not generic "unavailable")

### Key Entities

- **Vault Metadata (.meta.json)**: JSON file stored alongside vault.enc containing vault configuration. Key attributes: audit_enabled (bool), keychain_enabled (bool), created_at (timestamp), last_modified (timestamp)
- **Keychain Entry**: System keychain storage entry. Attributes: service="pass-cli", account="master-password", value=encrypted master password
- **Audit Log Entry**: JSON line in audit.log file. Attributes: timestamp, event_type, outcome, details (for keychain/vault operations)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can enable keychain on existing vault and immediately use TUI/CLI without password prompts (end-to-end workflow)
- **SC-002**: All 25 TODO-skipped tests are unskipped and passing after implementation
- **SC-003**: Vault remove command achieves 95% success rate across 20 consecutive runs (per test requirement)
- **SC-004**: Keychain status accurately reports state (0 false positives/negatives in test scenarios)
- **SC-005**: CI pipeline blocks merges containing `t.Skip("TODO:` patterns in test files
- **SC-006**: Integration test suite completes with 100% pass rate (no skips for unimplemented features)
- **SC-007**: Documentation updated to reflect all working keychain/vault commands with examples
