# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
