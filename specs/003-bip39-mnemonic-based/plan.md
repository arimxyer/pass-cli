# Implementation Plan: BIP39 Mnemonic Recovery

**Branch**: `003-bip39-mnemonic-based` | **Date**: 2025-01-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-bip39-mnemonic-based/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implements BIP39-based vault recovery system enabling users to reset their master password without knowing the current password. Users receive a standard 24-word mnemonic phrase during vault initialization and can later recover access by providing 6 randomly-selected words from that phrase (in randomized order). The system uses a "puzzle piece" approach: 6 challenge words unlock encrypted storage of the other 18 words, allowing full 24-word phrase reconstruction for vault unlocking. Optional BIP39 passphrase (25th word) provides additional security with automatic detection.

**Technical Approach**: Extend existing vault infrastructure with new `internal/recovery` package implementing BIP39 mnemonic generation, word splitting/encryption, and challenge-based recovery. Integrate recovery setup into `cmd/init.go` and recovery execution into `cmd/change_password.go` with `--recover` flag. Use `tyler-smith/go-bip39` library for standard-compliant mnemonic operations.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
- `github.com/tyler-smith/go-bip39` (new - BIP39 mnemonic generation/validation)
- `golang.org/x/crypto/argon2` (existing - key derivation)
- `crypto/aes`, `crypto/cipher` (existing - AES-GCM encryption)
- `crypto/rand` (existing - secure random generation)
- `github.com/spf13/cobra` (existing - CLI framework)

**Storage**: File-based encrypted vault storage (existing `vault.enc`), extends existing metadata JSON structure
**Testing**: `go test`, existing test infrastructure (`test/integration_test.go`, `test/unit/`)
**Target Platform**: Cross-platform (Windows, macOS, Linux) - no platform-specific code required
**Project Type**: Single (CLI application with library-first architecture)
**Performance Goals**:
- Recovery phrase generation: <2 seconds
- Recovery interaction: <30 seconds total
- Memory overhead: <500 bytes vault metadata increase

**Constraints**:
- Offline-first (no network operations)
- Memory clearing required (all sensitive data zeroed after use)
- No credential logging (recovery words, passphrases never logged)
- BIP39 standard compliance (interoperable with external tools)
- Atomic operations (no partial state corruption)

**Scale/Scope**:
- Single user per vault
- 24-word phrase (256-bit entropy)
- 6-word challenge (2^66 = 73.8 quintillion combinations)
- 5 user stories, 33 functional requirements

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Principle I: Security-First Development (NON-NEGOTIABLE)

**Compliance**: PASS

- **Encryption Standards**: Uses existing AES-256-GCM infrastructure (FR-029). New Argon2id KDF for mnemonic-derived keys follows existing patterns.
- **No Secret Logging**: FR-026 mandates no logging of recovery words, passphrases, or master passwords. Audit logs contain only operation types and timestamps.
- **Secure Memory Handling**: FR-027 requires `crypto.ClearBytes()` for all sensitive data (mnemonic, passphrase, seeds, keys). Follows existing patterns in `internal/crypto`.
- **Threat Modeling**: Spec includes security attack analysis (6-word entropy analysis, brute force resistance, passphrase protection).

**Evidence**: FR-025-030 (Security requirements), SC-008 (no leakage to logs/memory), Edge cases (corruption, invalid words).

### ✅ Principle II: Library-First Architecture

**Compliance**: PASS

- **New Library**: `internal/recovery` package handles all recovery logic independent of CLI.
- **Single Purpose**: Package focused solely on BIP39 mnemonic generation, word splitting, encryption, and challenge validation.
- **No CLI Dependencies**: Core recovery functions accept/return Go types ([]byte, error), no Cobra flags or output formatting.
- **Public APIs**: Exported functions: `GenerateMnemonic()`, `SetupRecovery()`, `PerformRecovery()`, `VerifyBackup()`.

**Evidence**: Technical approach specifies `internal/recovery` as standalone package. CLI commands (`cmd/init.go`, `cmd/change_password.go`) are thin wrappers.

### ✅ Principle III: CLI Interface Standards

**Compliance**: PASS

- **Input**: Command-line arguments (`--recover`, `--no-recovery`), stdin for word entry, optional passphrase via secure prompt.
- **Output**: Human-readable progress indicators ("✓ (3/6)"), warnings to stdout, errors to stderr.
- **Exit Codes**: Success (0), user error (1 - wrong words), system error (2 - metadata corruption), security error (3 - entropy failure).
- **Script-Friendly**: Recovery process requires interactive word entry (intentional security measure), but verification can be skipped with flag.

**Evidence**: User Story 1-5 acceptance scenarios define CLI interaction patterns. FR-015-017 specify user feedback requirements.

### ✅ Principle IV: Test-Driven Development (NON-NEGOTIABLE)

**Compliance**: PASS (pending implementation)

- **Test Types Required**:
  - **Unit Tests**: `internal/recovery/recovery_test.go` - BIP39 generation, word splitting, encryption, KDF operations
  - **Integration Tests**: `test/recovery_test.go` - End-to-end init + recovery flows
  - **Security Tests**: Memory clearing verification, no credential leakage, BIP39 checksum validation
  - **Contract Tests**: Public API stability for `internal/recovery` package

- **Coverage Goal**: Minimum 80% for `internal/recovery` package.

**Evidence**: Spec Testing Requirements section, FR-030 (checksum validation testable), SC-003-004 (100% success/failure rates testable).

### ✅ Principle V: Cross-Platform Compatibility

**Compliance**: PASS

- **Single Binary**: Pure Go implementation, no platform-specific code.
- **Platform-Specific Code**: None required (BIP39, crypto operations are platform-agnostic).
- **Path Handling**: Uses existing vault metadata storage (already cross-platform).
- **Testing Matrix**: Existing CI runs on Windows, macOS, Linux (no changes needed).

**Evidence**: Technical Context specifies cross-platform target. No OS-specific keychain operations in recovery flow.

### ✅ Principle VI: Observability & Auditability

**Compliance**: PASS

- **Audit Trail**: FR-028 requires logging recovery attempts (success/failure) with timestamps to `~/.pass-cli/audit.log`.
- **No Credential Logging**: FR-026 explicitly forbids logging recovery words or passphrases.
- **Operation Types**: Audit logs contain: `recovery_enabled` (init), `recovery_success` (successful unlock), `recovery_failed` (wrong words/passphrase).
- **Verbose Mode**: Existing `--verbose` flag usable for debugging (spec ensures no secrets in output).

**Evidence**: FR-028 (audit logging), FR-026 (no credential logging), Security section in spec.

### ✅ Principle VII: Simplicity & YAGNI

**Compliance**: PASS

- **No Speculative Features**: Scope explicitly excludes migration tool, phrase rotation, QR codes (deferred to future).
- **Standard Library**: Uses `tyler-smith/go-bip39` (battle-tested, minimal library) instead of rolling custom BIP39.
- **Direct Solutions**: "Puzzle piece" approach (6-word unlock of 18 stored words) is straightforward encryption, not complex secret sharing.
- **Flat Architecture**: Single new package (`internal/recovery`), integrates into existing commands.

**Evidence**: Spec "Out of Scope" section lists 8 deferred features. Technical approach uses proven `go-bip39` library. No new external services.

### Gate Result: ✅ PASS

All constitutional principles satisfied. Proceed to Phase 0 research.

## Project Structure

### Documentation (this feature)

```
specs/003-bip39-mnemonic-based/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── recovery-api.md  # internal/recovery package contract
├── checklists/
│   └── requirements.md  # Spec validation checklist (completed)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
pass-cli/
├── cmd/                      # CLI commands (Cobra-based)
│   ├── init.go               # [MODIFIED] Add recovery setup flow
│   ├── change_password.go    # [MODIFIED] Add --recover flag and recovery flow
│   └── helpers.go            # [MODIFIED] Add recovery-specific prompts
│
├── internal/                 # Internal library packages
│   ├── recovery/             # [NEW] BIP39 recovery package
│   │   ├── recovery.go       # Core recovery logic
│   │   ├── recovery_test.go  # Unit tests
│   │   ├── mnemonic.go       # BIP39 mnemonic operations
│   │   ├── challenge.go      # Challenge word selection and validation
│   │   └── crypto.go         # KDF and encryption for stored words
│   │
│   ├── vault/                # [MODIFIED] Extend metadata structure
│   │   ├── metadata.go       # Add Recovery struct to VaultMetadata
│   │   └── vault.go          # Recovery key integration
│   │
│   └── crypto/               # [EXISTING] No changes (uses existing ClearBytes)
│
├── test/                     # Integration and unit tests
│   ├── recovery_test.go      # [NEW] Integration tests for recovery flows
│   ├── unit/
│   │   └── recovery_test.go  # [NEW] Unit tests for internal/recovery
│   └── integration_test.go   # [EXISTING] May add recovery scenarios
│
└── go.mod                    # [MODIFIED] Add github.com/tyler-smith/go-bip39
```

**Structure Decision**: Single project structure (default). Recovery functionality is a core vault feature, not a separate service. Library-first architecture: `internal/recovery` provides reusable functions, CLI commands orchestrate user interaction. No new platform-specific code required (pure Go cryptographic operations).

## Complexity Tracking

*No violations detected. This section left empty per template guidance.*

---

**Phase 0 and Phase 1 outputs will follow in subsequent sections as the planning workflow continues.**
