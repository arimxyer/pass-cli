# Implementation Plan: Remove --vault Flag and Simplify Vault Path Configuration

**Branch**: `001-remove-vault-flag` | **Date**: 2025-10-30 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-remove-vault-flag/spec.md`

## Summary

Remove the `--vault` command-line flag from all commands and simplify vault path configuration to use a two-tier approach: config-based customization (via `vault_path` in config.yml) or default location (`$HOME/.pass-cli/vault.enc`). This refactor eliminates complexity from the CLI interface while maintaining flexibility for advanced users through configuration files. The change affects 15+ commands, test infrastructure, and documentation.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**: Cobra (CLI framework), Viper (configuration), spf13/pflag (flag parsing)
**Storage**: File-based (encrypted vault files, YAML config files)
**Testing**: Go standard testing (`go test`), integration tests with real file I/O
**Target Platform**: Cross-platform (Windows, macOS, Linux - amd64, arm64)
**Project Type**: Single project (CLI application with internal libraries)
**Performance Goals**: Instant vault path resolution (<1ms), no performance degradation
**Constraints**: Backward compatible for users without `--vault` flag usage, clear migration path
**Scale/Scope**: 15+ commands to modify, ~23 command files, 12+ test files, 7 documentation files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Security-First Development ✅ PASS
- **No security impact**: This is a refactoring of path resolution, not credential handling
- **No new attack surface**: Removes a flag, doesn't add new functionality
- **Config validation**: FR-012 requires validation of vault_path during config loading
- **Clear error messages**: FR-011, FR-016 ensure users understand vault path issues

### Principle II: Library-First Architecture ✅ PASS
- **Existing architecture**: Config package (`internal/config`) already exists
- **Clean separation**: Path resolution stays in `cmd/root.go` (GetVaultPath function)
- **Library reuse**: Config package handles YAML parsing, validation independent of CLI
- **No new libraries**: Uses existing vault and config packages

### Principle III: CLI Interface Standards ✅ PASS
- **Simplified interface**: Removes flag complexity, improves consistency
- **Script-friendly**: Config-based approach better for automation than repeated flags
- **Consistent exit codes**: Error handling unchanged (0=success, 1=user error)
- **No new prompts**: Path resolution happens silently based on config

### Principle IV: Test-Driven Development ✅ PASS
- **Test-first approach**: Update tests before removing flag implementation
- **Coverage maintained**: All existing test scenarios preserved, adapted to config-based approach
- **Integration tests**: D-003 explicitly requires test suite updates
- **Security tests**: No security-critical code changes, but path validation tests required

### Principle V: Cross-Platform Compatibility ✅ PASS
- **Path expansion**: FR-007 requires `~` and `$HOME`/`%USERPROFILE%` expansion
- **Platform-agnostic**: Uses `filepath.Join` for all path operations
- **Existing patterns**: Follows established home directory resolution in codebase
- **CI coverage**: Tests run on all platforms (Windows, macOS, Linux)

### Principle VI: Observability & Auditability ✅ PASS
- **Doctor command**: FR-014 requires `pass-cli doctor` to report resolved vault path
- **Clear error messages**: FR-016 ensures vault path is included in error messages
- **No logging changes**: Audit trail unchanged, path resolution doesn't require logging
- **Debugging support**: Verbose mode can show which vault path is being used

### Principle VII: Simplicity & YAGNI ✅ PASS
- **Removes complexity**: Eliminates flag priority logic, simplifies `GetVaultPath()`
- **No speculative features**: Removes unused multi-vault capability
- **Standard library**: Config already uses Viper (existing dependency)
- **Direct solution**: Two-tier resolution (config → default) is straightforward

**GATE RESULT**: ✅ ALL PRINCIPLES PASS - Proceed to Phase 0

## Project Structure

### Documentation (this feature)

```
specs/001-remove-vault-flag/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

**Existing Structure** (modification only, no new packages):

```
cmd/
├── root.go              # MODIFY: Remove --vault flag, simplify GetVaultPath()
├── helpers.go           # REVIEW: getAuditLogPath(), getVaultID() may need updates
├── init.go              # MODIFY: Remove --vault examples from help text
├── add.go               # MODIFY: Remove vault path usage examples
├── get.go               # MODIFY: Remove vault path usage examples
├── delete.go            # MODIFY: Remove vault path usage examples
├── update.go            # MODIFY: Remove vault path usage examples
├── list.go              # MODIFY: No changes (already uses GetVaultPath())
├── tui.go               # MODIFY: No changes (already uses GetVaultPath())
├── keychain_enable.go   # MODIFY: Remove --vault examples from help text
├── keychain_status.go   # MODIFY: Remove --vault examples from help text
├── doctor.go            # MODIFY: Update to report vault path source (config vs default)
├── usage.go             # REVIEW: No changes expected
├── change_password.go   # REVIEW: No changes expected
└── verify_audit.go      # REVIEW: No changes expected

internal/
├── config/
│   ├── config.go        # MODIFY: Add VaultPath field to Config struct
│   ├── config_test.go   # MODIFY: Add tests for vault_path validation
│   └── validation.go    # MODIFY: Add vault_path validation logic
├── vault/
│   ├── vault.go         # REVIEW: No changes (already accepts vaultPath parameter)
│   ├── metadata.go      # REVIEW: No changes expected
│   └── firstrun.go      # REVIEW: May need updates to detection logic
├── storage/             # REVIEW: No changes expected
├── health/              # REVIEW: Doctor command health checks may reference vault path
└── security/            # REVIEW: No changes expected

test/
├── integration_test.go          # MODIFY: Remove --vault flag usage, use config or default
├── list_test.go                 # MODIFY: Remove --vault flag usage
├── usage_test.go                # MODIFY: Remove --vault flag usage
├── doctor_test.go               # MODIFY: Test vault path reporting
├── firstrun_test.go             # MODIFY: Update vault path detection tests
├── vault_remove_test.go         # MODIFY: Remove --vault flag usage
├── vault_metadata_test.go       # MODIFY: Remove --vault flag usage
├── keychain_integration_test.go # MODIFY: Remove --vault flag usage
├── keychain_enable_test.go      # MODIFY: Remove --vault flag usage
└── unit/
    └── keychain_lifecycle_test.go # MODIFY: Remove --vault flag usage

docs/
├── USAGE.md             # MODIFY: Remove --vault flag docs, add vault_path config docs
├── GETTING_STARTED.md   # MODIFY: Remove custom vault location section using --vault
├── MIGRATION.md         # MODIFY: Remove/update Option C (hybrid approach)
├── TROUBLESHOOTING.md   # MODIFY: Update vault path error solutions
├── DOCTOR_COMMAND.md    # MODIFY: Remove custom vault path section
├── SECURITY.md          # MODIFY: Update testing recommendations
└── README.md            # REVIEW: Verify no --vault references
```

**Structure Decision**: This is a refactoring task affecting existing codebase structure. No new packages or major architectural changes. All modifications are within established `cmd/`, `internal/config/`, `test/`, and `docs/` directories following existing patterns.

## Complexity Tracking

**No violations to track**. This feature actively reduces complexity by:
- Removing flag parsing logic
- Simplifying vault path resolution from 3-tier (flag > config > default) to 2-tier (config > default)
- Eliminating the need for users to remember and specify flags
- Reducing test complexity (no need to test flag combinations)

This aligns perfectly with Principle VII (Simplicity & YAGNI).
