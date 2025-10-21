# Feature Specification: Keychain Lifecycle Management

**Feature Branch**: `011-keychain-lifecycle-management`
**Created**: 2025-10-20
**Status**: Draft
**Input**: User description: "Keychain Lifecycle Management - Complete the keychain lifecycle for pass-cli by adding three critical commands that are currently missing."

## Problem Statement

Pass-cli currently supports creating vaults with keychain integration (`--use-keychain` flag during init) and updating keychain passwords during password changes, but the keychain lifecycle is incomplete. Users face three critical gaps:

1. **Cannot enable keychain for existing vaults** - Users who created vaults without keychain must recreate the entire vault to enable keychain integration
2. **Cannot inspect keychain status** - When prompted for passwords unexpectedly, users have no visibility into whether keychain is enabled, available, or functioning correctly
3. **Orphaned keychain entries** - When vaults are manually deleted or moved, keychain entries remain in the system keychain (Windows Credential Manager, macOS Keychain, Linux Secret Service) indefinitely with no cleanup mechanism

**Root cause**: The keychain lifecycle was partially implemented - create and update exist (cmd/init.go:44, internal/vault/vault.go:819-823), but enable post-init, status inspection, and cleanup on removal were never added. The internal/keychain/keychain.go:94-105 has Delete/Clear methods but they're never called.

**Evidence from investigation** (n-n-o-1020.md):
- Q: Can keychain only be enabled during vault init? A: Yes, currently init-only at cmd/init.go:44
- Q: If vault is removed, are keychain entries cleaned up? A: No cleanup occurs, keychain entries orphaned
- Q: Command to inspect keychain status? A: Does not exist

## Clarifications

### Session 2025-10-20

- Q: Should keychain lifecycle operations (enable, status checks, removal) be logged for security auditing purposes? → A: Yes - Log all keychain operations (enable, status, remove) with timestamps and vault paths to existing audit log system

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Enable Keychain for Existing Vaults (Priority: P1) 🎯 MVP

Users who created vaults without system keychain integration want to enable it later without recreating their vault and losing all stored credentials. When a user runs the enable command and provides their vault password, the system should store it in the system keychain so future commands don't require password entry.

**Why this priority**: This is the highest-value improvement. Users currently must destroy and recreate vaults to enable keychain, losing all credentials and usage history in the process. This is the most requested missing functionality and directly improves daily UX for all users who initially opted out of keychain.

**Independent Test**: Can be fully tested by: (1) creating a vault without `--use-keychain`, (2) running the enable command and entering password, (3) verifying subsequent commands (get, add, list) don't prompt for password. Delivers immediate value - users can opt-in to convenience without data loss.

**Acceptance Scenarios**:

1. **Given** a user has an existing vault created without system keychain, **When** they run the enable command and provide their vault password, **Then** the password is stored in system keychain and subsequent vault operations don't prompt for password
2. **Given** system keychain is already enabled for a vault, **When** user runs enable command, **Then** system displays message "Keychain already enabled for this vault" and exits gracefully
3. **Given** system keychain is unavailable (e.g., headless SSH session), **When** user runs enable command, **Then** system displays clear error message explaining system keychain is not available on this system
4. **Given** user provides incorrect vault password during enable, **When** password verification fails, **Then** system displays error and allows retry without storing incorrect password
5. **Given** a user runs any keychain lifecycle command (enable, status, remove), **When** the operation completes (success or failure), **Then** the operation is logged to the audit log with timestamp, operation type, vault path, and outcome

---

### User Story 2 - Inspect Keychain Status (Priority: P2)

Users need visibility into system keychain state to troubleshoot issues when they're unexpectedly prompted for passwords or want to verify system keychain is working correctly. When a user runs the status command, they should see whether system keychain is available, whether their vault has a stored password, and which keychain backend is in use.

**Why this priority**: This is critical for troubleshooting but doesn't block daily operations. Users who encounter keychain issues (password prompts when unexpected) need diagnostic information to understand what's wrong. Rated P2 because it enables self-service troubleshooting but isn't needed for basic vault operations.

**Independent Test**: Can be fully tested by: (1) checking status with keychain disabled, (2) enabling keychain and checking status again, (3) testing on systems with unavailable keychain. Delivers value by providing clear diagnostic information for common keychain issues.

**Acceptance Scenarios**:

1. **Given** system keychain is enabled for current vault, **When** user runs status command, **Then** system displays: system keychain availability (yes/no), password stored status (yes/no for current vault), backend in use (Windows Credential Manager, macOS Keychain, or Linux Secret Service)
2. **Given** system keychain is unavailable, **When** user runs status command, **Then** system displays message "System keychain not available on this platform" with helpful context
3. **Given** multiple vaults exist with system keychain enabled, **When** user runs status command with `--vault` flag, **Then** system displays system keychain status for the specified vault path
4. **Given** system keychain is not enabled for current vault, **When** user runs status command, **Then** system displays "Keychain not enabled for this vault" and suggests running enable command

---

### User Story 3 - Clean Vault Removal (Priority: P3)

Users who want to remove test vaults or permanently delete vaults need a clean removal process that doesn't leave orphaned credentials in their system keychain. When a user runs the remove command with explicit confirmation, both the vault file and its system keychain entry should be deleted.

**Why this priority**: This prevents security hygiene issues but doesn't affect daily usage. Orphaned keychain entries are a cleanup problem, not an immediate usability crisis. Rated P3 because it's important for complete lifecycle management but users can function without it (manual vault deletion still works, just leaves orphaned entries).

**Independent Test**: Can be fully tested by: (1) creating vault with keychain, (2) running remove command with confirmation, (3) verifying both vault file and keychain entry are deleted. Delivers value by ensuring clean uninstall and preventing keychain pollution.

**Acceptance Scenarios**:

1. **Given** a vault exists with system keychain enabled, **When** user runs remove command and confirms the action, **Then** both vault file and system keychain entry are deleted completely
2. **Given** a vault exists without system keychain, **When** user runs remove command and confirms, **Then** vault file is deleted and system confirms no system keychain entry existed
3. **Given** user runs remove command without confirmation flag, **When** system prompts for confirmation, **Then** vault is only removed after user explicitly confirms (prevents accidental deletion)
4. **Given** vault file doesn't exist but keychain entry does, **When** user runs remove command, **Then** system removes keychain entry and displays warning that vault file was not found
5. **Given** user attempts to remove the default vault location while it's locked, **When** remove command executes, **Then** system deletes vault file and keychain entry without requiring unlock

---

### Edge Cases

- What happens when keychain is temporarily unavailable during enable command (e.g., network drive disconnection during operation)? System should display clear error and exit without corrupting vault or partial keychain entries.
- How does status command handle multiple vaults at different paths? Command should default to checking the vault at default location or path specified by `--vault` flag, displaying status for that specific vault only.
- What happens when user removes a vault that other processes are actively using (e.g., TUI open in another window)? System should check for file locks and warn user if vault appears to be in use, requiring `--force` flag to proceed.
- How does enable command handle permission denied errors when writing to system keychain? System should display clear error message with platform-specific troubleshooting (e.g., "Access denied to Windows Credential Manager - check user permissions").
- What happens when vault file exists but keychain entry is orphaned/corrupted? Status command should detect mismatch and suggest re-enabling keychain to fix.
- How does remove command handle confirm when running in CI/CD or non-interactive scripts? Command should support `--yes` or `--force` flag to bypass confirmation prompt for automation scenarios.
- What happens when keychain backend changes between operations (e.g., user switches Linux secret service implementations)? Status command should detect and display current backend; enable command should work with newly available backend.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide command to enable keychain for existing vaults without requiring vault recreation
- **FR-002**: System MUST prompt user for vault password when enabling keychain and validate password before storing in keychain
- **FR-003**: System MUST store keychain credentials using existing service name format: "pass-cli:/absolute/path/to/vault.enc"
- **FR-004**: System MUST provide command to display keychain status showing: (a) keychain availability, (b) password stored status for current vault, (c) keychain backend in use
- **FR-005**: System MUST provide command to remove vaults that deletes both vault file and associated keychain entry
- **FR-006**: System MUST require explicit user confirmation before removing vault (either via prompt or `--yes`/`--force` flag)
- **FR-007**: System MUST handle keychain-unavailable scenarios gracefully with clear error messages (no crashes or undefined behavior)
- **FR-008**: System MUST prevent duplicate keychain entries when enable command is run multiple times for same vault
- **FR-009**: System MUST work across all supported platforms: Windows (Credential Manager), macOS (Keychain), Linux (Secret Service)
- **FR-010**: Enable command MUST verify vault exists and is accessible before prompting for password
- **FR-011**: Status command MUST work without unlocking the vault (read-only operation)
- **FR-012**: Remove command MUST succeed even if vault file is missing (clean up orphaned keychain entry)
- **FR-013**: All commands MUST respect `--vault` flag for specifying non-default vault locations
- **FR-014**: Status command output MUST include actionable suggestions (e.g., "Run 'pass-cli keychain enable' to store password")
- **FR-015**: System MUST log all keychain lifecycle operations (enable, status, remove) to existing audit log system with timestamps and vault paths for security auditing

### Key Entities

- **Vault**: The encrypted credential storage file, identified by absolute file path, may or may not have associated keychain entry
- **Keychain Entry**: System keychain credential with service name format "pass-cli:/absolute/path/to/vault.enc", stores vault master password
- **Keychain Backend**: Platform-specific keychain implementation (Windows Credential Manager, macOS Keychain, Linux Secret Service)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can enable keychain for existing vault in 45-60 seconds (measured on reference hardware: 2GHz CPU, SSD) without recreating vault or losing credentials
- **SC-002**: Users can diagnose keychain issues within 30 seconds using status command output
- **SC-003**: 95% of vault removal operations successfully delete both vault file and keychain entry (minimum 20 test runs per platform, no orphaned entries)
- **SC-004**: All keychain lifecycle commands work correctly on Windows, macOS, and Linux (100% platform coverage)
- **SC-005**: Users receive clear, actionable error messages when keychain is unavailable (no generic errors or crashes)
- **SC-006**: Enable command prevents accidental password overwrites (rejects operation if keychain already enabled without `--force` flag)

## Assumptions

- Existing keychain infrastructure (internal/keychain package) is functional and tested across platforms
- Service name format "pass-cli:/absolute/path/to/vault.enc" is established pattern and won't change
- Users have basic understanding of what keychain is (or documentation explains it)
- Default vault location is $HOME/.pass-cli/vault.enc as documented in pass-cli architecture
- The `--vault` flag is available globally for specifying alternate vault paths (existing functionality)
- Confirmation prompts follow standard CLI patterns (y/n prompt or --yes flag)

## Dependencies

- Depends on internal/keychain package Delete/Clear methods (lines 94-105) being functional
- Depends on existing vault unlock/validation logic for password verification during enable
- Depends on existing vault path resolution logic (GetVaultPath() function)
- Depends on existing keychain service name generation logic
- Depends on existing audit log system for recording keychain lifecycle operations (per FR-015)

## Known Limitations

### Audit Logging for Non-Unlocking Operations

**Issue**: Commands that operate without unlocking the vault (`keychain status`, `vault remove`) cannot write audit entries because audit configuration is loaded only during vault unlock.

**Impact**:
- `keychain enable` → ✅ Audit logging works (unlocks vault to validate password)
- `keychain status` → ⚠️ No audit logging (read-only, doesn't unlock vault)
- `vault remove` → ⚠️ No audit logging (works on locked vault by design)

**Why**: Audit configuration (`AuditLogPath`, `VaultID`) is stored inside the encrypted vault data, requiring decryption to access.

**Partial FR-015 Compliance**: FR-015 states "System MUST log all keychain lifecycle operations" - currently only `enable` is logged. The `status` command has low security impact (read-only query), but `vault remove` is a destructive operation that should ideally be audited.

**Future Work**: See `FOLLOW_UP.md` for proposed metadata-based solution that would enable audit logging for all operations without requiring vault unlock.

**Workaround** (for `vault remove`): Unlock the vault before removal to load audit configuration:
```bash
# Option 1: Get any credential first (loads audit config)
pass-cli get <any-credential> --vault /path/to/vault.enc
pass-cli vault remove /path/to/vault.enc --yes

# Option 2: Use keychain to auto-unlock
pass-cli keychain enable --vault /path/to/vault.enc
pass-cli vault remove /path/to/vault.enc --yes
```
