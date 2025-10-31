# Implementation Plan: Documentation Accuracy Verification and Remediation

**Branch**: `010-documentation-accuracy-verification` | **Date**: 2025-10-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/010-documentation-accuracy-verification/spec.md`

## Summary

Systematic audit and remediation of all Pass-CLI documentation to eliminate factual inaccuracies. Root cause: Initial documentation (commit 1a60ce0) was written aspirationally from design specs rather than verifying actual implementation. Known issues: non-existent `--generate` flags documented across multiple files, missing `--category` flag documentation. This feature will verify 100% accuracy across 10 categories: CLI interface, code examples, file paths, configuration, features, architecture, metadata, output samples, cross-references, and behavioral claims. Deliverable is an audit report documenting all discrepancies and remediated documentation files matching actual codebase.

## Technical Context

**Language/Version**: Go 1.21+ (existing codebase language)
**Primary Dependencies**:
- Cobra (CLI framework - already in use)
- Internal packages: cmd/, internal/config, internal/vault, internal/security, internal/keychain
- Markdown parsing (standard library or lightweight parser for extracting code blocks)

**Storage**: N/A (documentation verification only, no data storage)
**Testing**: Manual verification workflow with documented test procedures in audit report
**Target Platform**: Cross-platform (Windows, macOS, Linux) - verification must cover platform-specific path documentation
**Project Type**: Documentation audit (single-phase verification and remediation workflow)
**Performance Goals**: N/A (manual audit process, not performance-critical)
**Constraints**:
- Must not modify source code to match documentation (FR-011: update docs to match implementation)
- Must preserve git history for all deleted/modified documentation
- Manual testing required for feature verification (FR-006)

**Scale/Scope**:
- 7 primary documentation files (README.md, USAGE.md, MIGRATION.md, SECURITY.md, TROUBLESHOOTING.md, KNOWN_LIMITATIONS.md, CONTRIBUTING.md)
- All docs/ subdirectory files
- 10 verification categories
- ~14 CLI commands to verify
- Estimated 50-100 discrepancies based on initial findings

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Principle I (Security-First Development)**: ✅ PASS
- No security implications - documentation-only changes
- No credential handling, encryption, or sensitive operations
- Verification workflow does not access user vaults or credentials

**Principle II (Library-First Architecture)**: ✅ N/A
- No new libraries required
- Uses existing cmd/ and internal/ packages for verification reference only

**Principle III (CLI Interface Standards)**: ✅ PASS
- Verification ensures documented CLI behavior matches actual implementation
- Directly supports this principle by eliminating documentation drift

**Principle IV (Test-Driven Development)**: ✅ PASS (Modified)
- Traditional TDD not applicable (documentation audit, not code)
- However, verification workflow follows test-first mindset: define expected behavior (from docs) → verify against implementation → remediate failures
- Audit report serves as acceptance test documentation

**Principle V (Cross-Platform Compatibility)**: ✅ PASS
- Verification must cover platform-specific path documentation (Windows %APPDATA%, macOS ~/Library, Linux ~/.config)
- Edge case documented: platform variants must be verified (spec line 98)

**Principle VI (Observability & Auditability)**: ✅ PASS
- Audit report (FR-015) provides complete traceability
- Git commits document every remediation with rationale (FR-012)
- Directly supports observability principle by improving documentation accuracy

**Principle VII (Simplicity & YAGNI)**: ✅ PASS
- No new code, no new dependencies
- Manual verification workflow (simple, no automation complexity)
- Addresses concrete user need (broken examples discovered in usage verification)

**Gate Result**: ✅ **ALL CHECKS PASS** - Proceed to Phase 0

## Project Structure

### Documentation (this feature)

```
specs/010-documentation-accuracy-verification/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (verification methodology research)
├── audit-report.md      # Phase 1 output (discrepancy findings)
├── verification-procedures.md  # Phase 1 output (detailed test procedures)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```
# Existing structure (no new code for this feature)
cmd/                     # Reference: CLI command implementations to verify against
├── add.go
├── config.go
├── generate.go
├── get.go
├── init.go
├── list.go
├── update.go
├── verify_audit.go
└── ... (other commands)

internal/                # Reference: Internal packages to verify architecture claims
├── config/             # Config validation rules
├── vault/              # Vault operations
├── security/           # Password policies
└── keychain/           # Keychain integration

docs/                    # Target: Files to verify and remediate
├── USAGE.md
├── SECURITY.md
├── MIGRATION.md
├── TROUBLESHOOTING.md
├── KNOWN_LIMITATIONS.md
├── INSTALLATION.md
└── DOCUMENTATION_LIFECYCLE.md

README.md                # Target: Primary documentation file
CONTRIBUTING.md          # Target: Contributor guidelines
```

**Structure Decision**: This is a documentation-only feature. No source code changes. Verification workflow references existing cmd/ and internal/ packages to validate documentation accuracy. Remediation modifies only markdown files in docs/ and root directory.

## Complexity Tracking

*No constitution violations requiring justification.*

---

## Phase 0: Research & Verification Methodology

### Research Questions

1. **CLI Verification Approach**: What is the most reliable method to extract and compare documented CLI flags against actual cobra command definitions?
   - Option A: Parse cmd/*.go files directly with AST
   - Option B: Execute `pass-cli [command] --help` and parse output
   - Option C: Manual inspection with checklist
   - **Decision needed**: Most reliable + least error-prone method

2. **Code Example Testing**: How should code examples be extracted and tested systematically?
   - Markdown code block extraction patterns
   - Test vault setup for example execution
   - Platform-specific example handling (bash vs PowerShell)

3. **Discrepancy Categorization**: What severity/priority levels should be assigned to different discrepancy types?
   - Critical: Broken commands (immediate user impact)
   - High: Incorrect flags (user frustration)
   - Medium: Missing documentation (feature exists, undocumented)
   - Low: Outdated metadata (cosmetic)

4. **Audit Report Format**: What structure enables efficient remediation tracking?
   - Per-file grouping vs. per-category grouping
   - Discrepancy ID scheme for traceability
   - Remediation status tracking

### Research Tasks

- [R1] Investigate cobra command introspection capabilities (can we programmatically list all flags from cobra.Command?)
- [R2] Research markdown parsing libraries in Go standard library or lightweight options
- [R3] Review existing specs 008/009 methodology for documentation review patterns (what worked, what didn't)
- [R4] Define verification test procedure template (Given/When/Then format for each category)
- [R5] Research best practices for documentation audit reports (industry standards, examples from other projects)

**Output**: `research.md` with decisions on verification methodology, tooling choices (if any), audit report structure, and test procedure format.

---

## Phase 1: Audit Design & Verification Procedures

### Prerequisites
- Phase 0 research complete
- Verification methodology decided
- Audit report structure defined

### Design Artifacts

1. **audit-report.md** (structured findings document):
   - Header: Audit date, scope (files covered), methodology
   - Summary statistics: Total discrepancies by type, files affected, severity breakdown
   - Discrepancy records (one per finding):
     - ID: DISC-### (sequential)
     - File: Path + line number
     - Category: CLI/Example/Path/Config/Feature/Architecture/Metadata/Output/Link
     - Severity: Critical/High/Medium/Low
     - Description: What is incorrect
     - Actual: What the code/implementation shows
     - Documented: What the docs currently claim
     - Remediation: Add/Remove/Update + specific action
     - Status: Open/Fixed/Verified
   - Appendix: Verification test results (per-category test execution log)

2. **verification-procedures.md** (repeatable test procedures):
   - Category 1: CLI Interface Verification
     - Procedure: For each command, run `--help`, compare against USAGE.md table
     - Expected: Flag name, type, short flag, description match exactly
     - Test checklist: 14 commands × ~5 flags each = 70+ verifications
   - Category 2: Code Examples Verification
     - Procedure: Extract code blocks with language tags, execute in test environment
     - Expected: Zero errors, output matches documented samples
     - Test checklist: ~30 code examples across all docs
   - (Repeat for all 10 categories)

3. **Agent Context Update**:
   - Run `.specify/scripts/powershell/update-agent-context.ps1 -AgentType claude`
   - Add: "Documentation verification workflow for Pass-CLI (spec 010)"
   - Add: "Audit report structure and discrepancy tracking methodology"
   - Preserve: Existing security, architecture, and testing context

### Deliverables

- `audit-report.md` (initial template, populated during implementation)
- `verification-procedures.md` (detailed test procedures for each category)
- Updated `.specify/memory/claude.md` with audit workflow context

**Constitution Re-Check Post-Design**: ✅ All principles still satisfied (no design changes affecting compliance)

---

## Next Steps

After Phase 1 completion, proceed to `/speckit.tasks` to generate task breakdown from this plan. Implementation will follow this sequence:

1. **Setup** (Phase 0 execution): Populate research.md with methodology decisions
2. **Audit Execution** (Phase 1 implementation): Run verification procedures, populate audit-report.md with findings
3. **Remediation** (Phase 2 implementation): Fix all discrepancies per audit report, verify fixes, close audit report
4. **Validation** (Phase 3): Verify all success criteria met (SC-001 through SC-012)

---

**Plan Status**: ✅ **COMPLETE** - Ready for `/speckit.tasks`
