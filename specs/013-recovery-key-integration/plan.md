# Implementation Plan: Recovery Key Integration

**Branch**: `013-recovery-key-integration` | **Date**: 2025-12-05 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/013-recovery-key-integration/spec.md`

## Summary

Fix the broken recovery phrase system by implementing key wrapping. Currently, the vault is encrypted directly with a password-derived key, making the recovery phrase useless. The fix introduces a Data Encryption Key (DEK) that encrypts the vault, with the DEK itself wrapped by both the password-derived key and the recovery-derived key, enabling either path to unlock the vault.

## Technical Context

**Language/Version**: Go 1.25.1
**Primary Dependencies**:
- `golang.org/x/crypto` (PBKDF2, Argon2id)
- `github.com/tyler-smith/go-bip39` (mnemonic generation)
- `crypto/aes`, `crypto/cipher` (AES-256-GCM)
**Storage**: File-based encrypted vault (`vault.enc`), JSON metadata (`.meta.json`)
**Testing**: Go testing framework (`go test`), testify assertions
**Target Platform**: Windows, macOS, Linux (cross-platform)
**Project Type**: Single CLI application with library packages
**Performance Goals**: Unlock time within 10% of current (~500-1000ms for KDF)
**Constraints**: Offline-only, no external services, atomic file operations
**Scale/Scope**: Single-user local vault, typical credential count <1000

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Security-First (NON-NEGOTIABLE) | ✅ PASS | DEK uses AES-256-GCM, 256-bit random key via crypto/rand. Both KDFs (PBKDF2, Argon2id) compliant with constitution. Memory clearing via `crypto.ClearBytes()`. |
| II. Library-First Architecture | ✅ PASS | Key wrapping implemented in `internal/crypto/keywrap.go`. Vault integration in `internal/vault/`. CLI remains thin wrapper. |
| III. CLI Interface Standards | ✅ PASS | Existing `change-password --recover` interface unchanged. Errors to stderr, exit codes follow standard. |
| IV. Test-Driven Development (NON-NEGOTIABLE) | ✅ PASS | Unit tests for key wrapping, integration tests for full recovery flow. TDD required per constitution. |
| V. Cross-Platform Compatibility | ✅ PASS | No platform-specific code needed. Uses Go standard library crypto. |
| VI. Observability & Auditability | ✅ PASS | Recovery events logged to audit.log. No secrets in logs (FR-026). |
| VII. Simplicity & YAGNI | ✅ PASS | Minimal changes to existing structure. Key wrapping is industry-standard pattern, not over-engineered. |

**Security Requirements Check**:
- ✅ AES-256-GCM for all encryption (DEK wrapping, vault encryption)
- ✅ DEK is 256-bit random from crypto/rand
- ✅ DEK cleared from memory after use
- ✅ No secrets logged
- ✅ Vault file permissions unchanged (0600)

## Project Structure

### Documentation (this feature)

```text
specs/013-recovery-key-integration/
├── plan.md              # This file
├── research.md          # Key wrapping patterns research
├── data-model.md        # Updated vault data structures
├── quickstart.md        # Developer setup guide
├── contracts/           # Internal API contracts
│   └── keywrap.md       # Key wrapping interface contract
└── tasks.md             # Implementation tasks (created by /speckit.tasks)
```

### Source Code (repository root)

```text
pass-cli/
├── cmd/
│   ├── init.go              # MODIFY: Integrate DEK generation and wrapping
│   └── change_password.go   # MODIFY: Use DEK unwrapping for recovery
├── internal/
│   ├── crypto/
│   │   ├── crypto.go        # EXISTING: AES-GCM, PBKDF2
│   │   └── keywrap.go       # NEW: Key wrapping operations
│   ├── storage/
│   │   └── storage.go       # MODIFY: Support wrapped DEK in vault format
│   ├── vault/
│   │   ├── vault.go         # MODIFY: UnlockWithDEK, password change with DEK
│   │   └── metadata.go      # MODIFY: Add WrappedDEK fields to metadata
│   └── recovery/
│       └── recovery.go      # MODIFY: Return DEK instead of VaultRecoveryKey
└── test/
    ├── unit/
    │   └── keywrap_test.go  # NEW: Key wrapping unit tests
    └── recovery_integration_test.go  # NEW: End-to-end recovery tests
```

**Structure Decision**: Single project structure maintained. New code added to existing packages following library-first architecture. Key wrapping is a crypto concern, vault integration is a vault concern.

## Complexity Tracking

> No constitution violations requiring justification. Key wrapping is the minimal viable solution for the dual-key requirement.

| Consideration | Decision | Rationale |
|--------------|----------|-----------|
| Key wrapping vs. dual encryption | Key wrapping | Industry standard (NIST SP 800-38F). Single DEK means single encryption/decryption path. Simpler than encrypting vault twice. |
| Vault format version | Version 2 | Clear distinction from legacy vaults. Enables graceful migration. |
| Migration approach | On-unlock prompt | Non-disruptive. User controls when to migrate. Atomic operation. |
