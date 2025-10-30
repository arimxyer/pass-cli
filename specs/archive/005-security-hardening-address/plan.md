# Implementation Plan: Security Hardening

**Branch**: `005-security-hardening-address` | **Date**: 2025-10-11 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/005-security-hardening-address/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Harden Pass-CLI security across four critical areas: (1) memory security - eliminate plaintext master password exposure via byte array handling and immediate zeroing; (2) cryptographic strength - increase PBKDF2 iterations from 100k to 600k (OWASP 2023); (3) password policy - enforce 12+ character complexity requirements; (4) audit logging - tamper-evident forensic trail for vault operations. Maintains backward compatibility via automatic migration during password changes.

## Technical Context

**Language/Version**: Go 1.21+ (requires crypto/subtle from stdlib)
**Primary Dependencies**: Go stdlib (`crypto/rand`, `crypto/subtle`, `crypto/sha256`), gopass terminal library (confirmed returns `[]byte`, per research.md Decision 1)
**Storage**: Encrypted vault files (local filesystem, AES-256-GCM), audit logs (`~/.pass-cli/audit.log` when enabled)
**Testing**: Go testing framework (`go test`), security-specific tests (memory inspection, crypto timing benchmarks, cross-platform validation)
**Target Platform**: Cross-platform CLI (Windows, macOS, Linux) with TUI mode support
**Project Type**: Single project (CLI + TUI interfaces sharing core libraries)
**Performance Goals**: Key derivation 500-1000ms (calibrated for security), memory clearing <1ms
**Constraints**: Zero credential logging, backward compatibility with 100k iteration vaults, memory security best-effort within Go GC limitations
**Scale/Scope**: Individual user tool, local-only operations, ~26 functional requirements across 4 categories

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Check (Pre-Research)

| Principle | Status | Notes |
|-----------|--------|-------|
| **I. Security-First Development** | ✅ PASS | Feature directly addresses critical security vulnerabilities (memory exposure, weak KDF, weak passwords, audit gaps) |
| **II. Library-First Architecture** | ⚠️ VERIFY | Must ensure crypto/vault packages remain library-first (no CLI coupling). Verify during research. |
| **III. CLI Interface Standards** | ✅ PASS | Audit logging to stderr/file, password validation errors to stderr, maintains CLI contract |
| **IV. Test-Driven Development** | ✅ PASS | Spec includes detailed Independent Test sections for each user story. Tests required for memory clearing, crypto timing, password validation, audit integrity. |
| **V. Cross-Platform Compatibility** | ✅ PASS | Spec Assumption 6 requires Windows/macOS/Linux testing. Memory clearing uses `crypto/subtle` for cross-platform consistency. |
| **VI. Observability & Auditability** | ✅ PASS | FR-019-026 add comprehensive audit logging (vault unlock, credential access, tamper detection). FR-021 prohibits credential logging. |
| **VII. Simplicity & YAGNI** | ✅ PASS | Removed unnecessary rate limiting (User Story 5) and optional security improvements (FR-027-029) during clarification to focus on real threats. |

**Security Requirements Check**:
- ❌ **FORBIDDEN**: Logging credentials → FR-021 explicitly prohibits this
- ✅ **REQUIRED**: Memory zeroing → FR-002 mandates secure zeroing after use
- ✅ **REQUIRED**: Clipboard auto-clear → Out of scope (no clipboard changes in this feature)
- ✅ **REQUIRED**: Approved crypto → FR-006 uses PBKDF2-SHA256 (approved), FR-001-005 add memory security
- ✅ **REQUIRED**: Tests for security guarantees → Spec requires memory inspection tests, crypto timing tests, audit integrity tests

**GATE RESULT**: ✅ **PASS** - No violations. Proceed to Phase 0 research.

## Project Structure

### Documentation (this feature)

```
specs/005-security-hardening-address/
├── spec.md              # Feature specification (input)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command) - N/A for this feature (no API changes)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
# Single project structure (Pass-CLI)
internal/
├── crypto/              # Cryptographic operations (PBKDF2, AES-GCM)
│   └── crypto.go        # NEEDS REFACTORING: Accept []byte passwords (FR-001)
├── vault/               # Vault storage and management
│   └── vault.go         # NEEDS REFACTORING: Byte arrays, password policy (FR-011-018)
├── security/            # NEW: Security utilities (memory clearing, password validation)
│   ├── memory.go        # NEW: Secure zeroing helpers (FR-002-004)
│   ├── password.go      # NEW: Password policy validation (FR-011-016)
│   └── audit.go         # NEW: Audit logging (FR-019-026)
├── input/               # NEW or EXISTING: Terminal input (NEEDS RESEARCH: location)
│   └── password.go      # NEEDS INVESTIGATION: Byte-based password input (FR-005)

cmd/
├── cli/                 # CLI commands
│   └── main.go          # Password strength indicator (CLI mode, FR-017)
└── tui/                 # TUI interface
    └── main.go          # Password strength indicator (TUI mode, FR-017)

tests/
├── unit/                # Unit tests for crypto, vault, security packages
├── integration/         # End-to-end tests (vault operations with new security)
└── security/            # NEW: Security-specific tests (memory inspection, crypto timing)
```

**Structure Decision**: Single project structure maintained. New `internal/security/` package consolidates memory, password, and audit utilities. Existing `internal/crypto/` and `internal/vault/` packages require refactoring to accept `[]byte` passwords. Terminal input location and library choice requires research (Phase 0).

## Complexity Tracking

*No violations identified. All changes align with constitution principles.*

---

## Phase 0: Outline & Research

### Research Tasks

1. **Terminal Input Library Decision** (from Clarification 2)
   - **Question**: Does current `gopass` library support `[]byte` password input, or is migration to `golang.org/x/term` required for FR-001/FR-005?
   - **Investigation**: Inspect codebase for current password input mechanism, check library API documentation
   - **Deliverable**: Library decision + migration plan (if needed)

2. **Current Crypto Package API**
   - **Question**: What is the current signature of key derivation functions in `internal/crypto/crypto.go`?
   - **Investigation**: Read `internal/crypto/crypto.go` to understand current string-based API
   - **Deliverable**: API refactoring plan (string → []byte conversion)

3. **Current Vault Package API**
   - **Question**: What is the current signature of vault unlock/init functions in `internal/vault/vault.go`?
   - **Investigation**: Read `internal/vault/vault.go` to understand current password handling
   - **Deliverable**: API refactoring plan + backward compatibility strategy

4. **TUI Framework Identification**
   - **Question**: What TUI framework is used for password input? (needed for FR-017 strength indicator)
   - **Investigation**: Inspect `cmd/tui/` to identify framework (likely bubbletea or tview)
   - **Deliverable**: Framework-specific implementation plan for visual strength meter

5. **Current PBKDF2 Implementation**
   - **Question**: Where is iteration count stored in vault metadata? How to support configurable iterations?
   - **Investigation**: Read vault file format, identify metadata structure
   - **Deliverable**: Metadata versioning plan for FR-007/FR-008 migration

6. **Memory Clearing Best Practices in Go**
   - **Question**: What is the canonical pattern for secure memory zeroing with `crypto/subtle`?
   - **Investigation**: Research Go stdlib examples, verify compiler optimization behavior
   - **Deliverable**: Reusable `ClearBytes([]byte)` helper design

7. **HMAC Audit Log Format**
   - **Question**: What HMAC structure supports tamper detection (FR-022)?
   - **Investigation**: Research HMAC-based log integrity patterns, key management for audit HMAC
   - **Deliverable**: Audit log entry format + HMAC signature scheme

### Research Consolidation

**Output**: `research.md` with all decisions documented using format:
```
## Decision: [Terminal Input Library]
**Chosen**: [gopass/golang.org/x/term]
**Rationale**: [byte array support, cross-platform compatibility, maintenance status]
**Alternatives Considered**: [other options and why rejected]
```

---

## Phase 1: Design & Contracts

*Prerequisites: research.md complete*

### 1. Data Model

**Generate**: `data-model.md` with entities:

- **VaultMetadata** (extended)
  - Fields: `iterationCount int`, `version string`, `createdAt time.Time`, `lastModified time.Time`
  - Validation: `iterationCount >= 600000` for new vaults, support legacy 100k
  - State transitions: Migration trigger on password change (FR-008)

- **PasswordPolicy**
  - Fields: `minLength int`, `requireUppercase bool`, `requireLowercase bool`, `requireDigit bool`, `requireSymbol bool`
  - Validation: Enforce FR-011-015 rules
  - Methods: `Validate(password []byte) error`, `Strength(password []byte) string` (weak/medium/strong)

- **AuditLogEntry**
  - Fields: `timestamp time.Time`, `eventType string`, `outcome string`, `credentialName string` (NOT password), `hmacSignature []byte`
  - Validation: `eventType` enum (vault_unlock, credential_access, credential_modify, credential_delete)
  - Integrity: HMAC-SHA256 signature over all fields (FR-022)

- **SecureBytes** (wrapper for `[]byte`)
  - Purpose: Type-safe wrapper for sensitive byte arrays with deferred cleanup
  - Methods: `Clear()`, `defer sb.Clear()` pattern
  - Rationale: Enforce FR-003 deferred cleanup discipline

### 2. API Contracts

**N/A for this feature** - No external API changes. All changes are internal library refactoring.

**Internal Interface Changes** (documented in `data-model.md`):
- `crypto.DeriveKey(password []byte, salt []byte, iterations int) ([]byte, error)` - was `(password string, ...)`
- `vault.Unlock(password []byte) error` - was `(password string)`
- `vault.Init(password []byte, policy PasswordPolicy) error` - was `(password string)`
- `security.ValidatePassword(password []byte, policy PasswordPolicy) error` - new function
- `security.ClearBytes(data []byte)` - new helper for FR-002

### 3. Quickstart Guide

**Generate**: `quickstart.md` with developer onboarding:

```markdown
# Security Hardening Quickstart

## For Developers

### Running Security Tests
\`\`\`bash
# Memory inspection tests (requires debugger)
go test -v ./tests/security -run TestMemoryClear

# Crypto timing benchmarks
go test -bench=. ./internal/crypto -benchtime=5s

# Cross-platform test matrix
make test-all-platforms
\`\`\`

### Adding New Password Validation Rules
See `internal/security/password.go` for PasswordPolicy extension points.

### Enabling Audit Logging (for testing)
\`\`\`bash
pass-cli init --enable-audit
export PASS_AUDIT_LOG=~/.pass-cli/test-audit.log
\`\`\`
```

### 4. Agent Context Update

**Run**: `.specify/scripts/powershell/update-agent-context.ps1 -AgentType claude`

**New Technology to Add**:
- `golang.org/x/term` (if migration required from research decision)
- `crypto/subtle` (for constant-time operations)
- HMAC-SHA256 for audit log integrity

---

## Phase 2: Task Generation

**Not executed by /speckit.plan** - User must run `/speckit.tasks` separately after reviewing this plan.

---

## Post-Phase 1: Constitution Re-Check

*Execute after Phase 1 design complete*

### Design Review Against Constitution

| Principle | Status | Design Notes |
|-----------|--------|--------------|
| **I. Security-First Development** | ✅ PASS | `SecureBytes` wrapper enforces memory clearing discipline. No credential logging in audit design. |
| **II. Library-First Architecture** | ✅ PASS | All changes in `internal/` packages (crypto, vault, security) - no CLI coupling. |
| **III. CLI Interface Standards** | ✅ PASS | Audit log to file (not stdout), validation errors to stderr, maintains existing CLI contract. |
| **IV. Test-Driven Development** | ✅ PASS | Research identified 3 test categories: unit (password validation), integration (vault operations), security (memory/timing). |
| **V. Cross-Platform Compatibility** | ✅ PASS | Terminal library decision in research phase considered cross-platform support. `crypto/subtle` is cross-platform. |
| **VI. Observability & Auditability** | ✅ PASS | AuditLogEntry design provides tamper-evident logging. FR-021 prevents credential leakage. |
| **VII. Simplicity & YAGNI** | ✅ PASS | Removed speculative features during clarification. Data model is minimal (4 entities, all necessary). |

**FINAL GATE RESULT**: ✅ **PASS** - Design maintains constitution compliance. Proceed to `/speckit.tasks`.

---

## Next Steps

1. ✅ **Phase 0 Complete**: `research.md` generated with all technical decisions documented
2. ✅ **Phase 1 Complete**: `data-model.md` and `quickstart.md` generated
3. ⏳ **Agent Context Update**: User should run `.specify/scripts/powershell/update-agent-context.ps1 -AgentType claude` to add new technology to CLAUDE.md
4. ⏳ **Phase 2**: User runs `/speckit.tasks` to generate actionable task breakdown from this plan

---

## Planning Complete ✅

**Generated Artifacts**:
- `plan.md` (this file) - Implementation plan with technical context, constitution check, and phase breakdown
- `research.md` - 7 technical decisions documented with rationale and alternatives
- `data-model.md` - 4 entity definitions (VaultMetadata, PasswordPolicy, AuditLogEntry, SecureBytes) with validation rules
- `quickstart.md` - Developer onboarding guide with testing instructions

**Key Findings**:
- **Terminal library**: Keep existing `gopass` (already returns `[]byte`)
- **Critical vulnerability**: `VaultService.masterPassword` stored as `string` (cannot be zeroed)
- **Memory clearing**: Existing `clearBytes()` function works correctly with `crypto/subtle`
- **TUI framework**: `tview` (confirmed)
- **Metadata**: Need to add `Iterations int` field for FR-007

**Next Command**: `/speckit.tasks` to generate dependency-ordered task breakdown
