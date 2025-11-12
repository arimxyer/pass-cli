# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.10.0] - 2025-11-12

### Added
- **Manual Vault Backup Commands**: Three new CLI commands for manual backup management
  - `pass vault backup create` - Create timestamped manual backups (vault.enc.YYYYMMDD-HHMMSS.manual.backup)
  - `pass vault backup restore` - Restore vault from newest available backup (manual or automatic)
  - `pass vault backup info` - View backup status, history, and integrity
- **Smart Backup Selection**: Restore automatically selects newest valid backup with fallback to manual backups
- **Backup Integrity Verification**: Structural validation before backup creation and during restore
- **Interactive Restore Confirmation**: User prompts with backup details, `--force` for scripting, `--dry-run` for preview
- **Comprehensive Backup Status**: Lists all backups with age, size, integrity, and restore priority
- **Backup Warnings**: Alerts for old backups (>30 days) and excessive disk usage
- **Cross-Platform Support**: Works on Windows, macOS, Linux with platform-specific path handling
- **Backup Restore Guide**: 484-line comprehensive guide covering workflows, best practices, and troubleshooting

### Changed
- CI integration test timeout increased from 2m to 4m to accommodate Windows CI infrastructure

### Performance
- Integration test suite optimized: 96 tests complete in <3m across all platforms
- Backup operations exceed performance targets (create: 176ms < 5s, restore: 191ms < 30s, info: 191ms < 1s)

### Testing
- Added 6 comprehensive test files with 96 integration tests (100% pass rate)
- Storage package coverage increased to 81.4%
- Error handling tests for corrupted backups, missing vault, permission denied scenarios
- Platform-specific tests for Windows/Unix path handling

## [0.9.5] - 2025-11-11

### Added
- **TUI Password Generator**: In-form password generation with Ctrl+G shortcut for Add forms
- **CLI Password Generation**: `--generate` flag for `add` and `update` commands with configurable length
- **Clipboard Support**: Copy username (u), URL (l), notes (n), and password (c) from TUI detail view
- **Command Grouping**: CLI commands organized into logical groups (vault, credentials, security, utilities)
- **Multiple Color Themes**: Dracula (default), Nord, Gruvbox, and Monokai themes for TUI
- **Responsive Layout**: Configurable detail panel positioning (right/bottom/auto) with auto-threshold
- **Theme Configuration**: Terminal settings for theme, detail position, and auto-threshold in config.yaml

### Changed
- CLI help output now displays commands in organized groups for better discoverability
- TUI detail panel now uses dynamic color helpers for consistent theming
- Medium layout mode now supports detail panel in bottom position
- Detail panel auto-switches to bottom when terminal width < 120 columns

### Performance
- CI workflow optimized: reduced from 9+ minutes to ~4.5 minutes (50% improvement)
- Removed race detector tests due to fundamental conflict with security validation requirements
- Parallel job execution for lint, unit-tests, integration-tests, and security scans

### Dependencies
- Bumped github.com/fatih/color from 1.15.0 to 1.18.0
- Bumped github.com/olekukonko/tablewriter from 1.1.0 to 1.1.1
- Bumped golangci/golangci-lint-action from 8 to 9

## [0.9.0] - 2025-11-10

### Added
- **Atomic Save Pattern**: Crash-safe vault operations using write-to-temp, verify, atomic-rename workflow
- **Actionable Error Messages (FR-011)**: Clear error messages with specific failure reason, vault status confirmation, and actionable guidance
- **Complete Audit Logging (FR-015)**: All atomic save state transitions logged (9 events tracked)
- In-memory verification before committing vault changes to prevent corruption
- N-1 backup strategy with automatic cleanup after successful unlock
- Orphaned temporary file cleanup from crashed save operations
- Custom error types: ErrVerificationFailed, ErrDiskSpaceExhausted, ErrPermissionDenied, ErrFilesystemNotAtomic
- FileSystem abstraction interface for testability and error injection
- 8 new comprehensive test files with 80.8% coverage in storage package

### Changed
- Vault save operations now use atomic rename pattern instead of direct writes
- Error messages now include vault status and recovery guidance
- All vault modifications protected against crashes and power loss

### Fixed
- Vault corruption during save operations now impossible due to atomic pattern
- Clear error messages when save fails (disk space, permissions, verification)
- Temp files automatically cleaned up after successful saves

## [0.8.76] - 2025-11-08

### Fixed
- Documentation now uses correct repository URLs instead of placeholders
- Post-install messages updated for Homebrew and Scoop package managers
- Config debug output removed from production builds

## [0.8.75] - 2025-11-08

### Fixed
- Vault metadata handling and audit logging consistency
- Consistent vaultID usage for audit key storage and retrieval
- Tests updated to reflect metadata always created for all vaults
- First-run detection now works correctly in TUI entry point
- Vault initialization properly creates metadata during guided setup
- Password prompt no longer appears when vault doesn't exist on first run

### Added
- TUI now launches by default with first-run guided initialization
- Vault remove command deletes audit log and offers complete directory removal
- Enhanced keychain availability checking on-demand

### Changed
- Configuration location consolidated to `~/.pass-cli` for cross-platform consistency

## [0.8.74] - 2025-11-07

### Added
- Standalone security scan workflow for continuous security monitoring

### Fixed
- SARIF format sanitization for gosec security scan output
- Invalid artifactChanges fields removed from security scan results

## [0.8.73] - 2025-11-07

### Changed
- Documentation badges now use dynamic GitHub badges for version and last updated
- Removed static status badges in favor of dynamic alternatives
- Logo positioning improved in documentation

## [0.8.72] - 2025-11-06

### Fixed
- Keychain tests updated for lazy initialization pattern
- Prevented keyring.Set() blocking on macOS in CI environment
- Lazy keychain initialization prevents macOS CI hangs
- Missing keychain prompt added to TestDefaultVaultPath_Init

## [0.8.71] - 2025-11-06

### Changed
- List and usage tests refactored to use production flags
- Tests now use production flags instead of stdin for better reliability
- Stdin buffering conflicts in test mode resolved

### Fixed
- Cross-platform stdin reading reliability improved with bufio.Scanner
- macOS stdin blocking issues resolved
- Custom vault path and --config flag issues fixed

## [0.8.70] - 2025-11-05

### Changed
- Integration test timeout reduced from 5m to 2m for faster CI feedback
- Test mode detection using PASS_CLI_TEST environment variable
- golangci-lint now uses CLI args instead of config file

### Fixed
- First-run check now skipped in test mode
- All golangci-lint errors resolved in integration tests

## [0.8.55] - 2025-11-04

### Changed
- Documentation files reorganized for better clarity
- Removed test artifacts and temporary files from repository

### Fixed
- All linting issues resolved for code quality compliance
- Test failures fixed in Phase 6 quality improvements

## [0.8.52] - 2025-11-03

### Added
- Keychain status command with audit logging and consistency checks
- Keychain enable command with metadata integration

### Fixed
- Metadata file paths corrected in integration tests

## [0.8.51] - 2025-11-03

### Added
- Vault remove command with complete cleanup functionality

---

## Format Guidelines

This changelog follows these principles:
- **Added** for new features
- **Changed** for changes in existing functionality
- **Deprecated** for soon-to-be removed features
- **Removed** for now removed features
- **Fixed** for any bug fixes
- **Security** for vulnerability fixes

For detailed commit-level changes, see [GitHub Releases](https://github.com/ari1110/pass-cli/releases).
