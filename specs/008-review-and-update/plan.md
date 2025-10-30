# Implementation Plan: Documentation Review and Production Release Preparation

**Branch**: `008-review-and-update` | **Date**: 2025-01-14 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/008-review-and-update/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Review and update all user-facing documentation (README, SECURITY, USAGE, INSTALLATION, TROUBLESHOOTING, MIGRATION, KNOWN_LIMITATIONS) to ensure 100% accuracy for production release. Verify all CLI commands execute successfully, security specifications are complete, TUI keyboard shortcuts match implementation, version references are current (600k PBKDF2 iterations, January 2025 features), and troubleshooting solutions are comprehensive. Target: new users onboard in 10 minutes, security professionals identify all crypto parameters without reading source code, zero outdated references.

**Technical Approach**: Documentation audit and correction workflow using validation scripts, command execution testing, cross-reference verification, and link checking. No new content creation—scope limited to accuracy verification and updates of existing 7 documentation files.

## Technical Context

**Language/Version**: Markdown (documentation), Go 1.25 (for command validation testing)
**Primary Dependencies**: Pass-CLI binary (current release), package managers (Homebrew, Scoop for installation verification)
**Storage**: Documentation files in `/docs` and `/README.md`, spec documentation in `/specs/008-review-and-update`
**Testing**: Manual command execution, documentation link validation, version reference grep, cross-reference consistency checks
**Target Platform**: Documentation consumers (web browsers for GitHub, terminal viewers, IDE markdown renderers)
**Project Type**: Documentation review (not source code implementation)
**Performance Goals**: New user completes installation + first credential in 10 minutes, troubleshooting solutions findable in 5 minutes
**Constraints**: Documentation must match current release binary behavior, zero tolerance for outdated iteration counts or unimplemented feature references
**Scale/Scope**: 7 documentation files, 15+ TUI keyboard shortcuts, 20+ CLI commands with examples, 5+ installation methods

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Applicable Principles for Documentation Review Feature:**

### I. Security-First Development ✅ PASS
- **No Secret Logging**: Documentation examples demonstrate `--quiet` mode and warn against logging credentials in scripts (FR-011)
- **Security Documentation**: SECURITY.md must document complete crypto implementation (AES-256-GCM, PBKDF2 600k) and threat model (FR-003, FR-004, FR-005)
- **TUI Security Warnings**: Documentation must include warnings for shoulder surfing, screen recording, shared terminals (FR-014)
- **Compliance**: No security violations—documentation review enforces security best practices

### III. CLI Interface Standards ✅ PASS
- **Script-Friendly Modes**: Documentation must accurately describe `--quiet`, `--field`, `--masked`, `--no-clipboard` flags (FR-006, FR-011)
- **TUI vs CLI Clarity**: USAGE.md must clearly distinguish when TUI launches (`pass-cli` no args) vs CLI mode (`pass-cli <command>`) (FR-006)
- **Compliance**: Documentation ensures users understand CLI I/O protocols for automation

### IV. Test-Driven Development ✅ PASS (Modified for Documentation)
- **Verification Tests**: All documented CLI commands must be executed to verify accuracy (SC-003: 100% command execution success)
- **Acceptance Criteria**: Each user story has 3 testable acceptance scenarios (new user onboarding, security evaluation, automation, version accuracy, troubleshooting)
- **Compliance**: Documentation review includes validation testing (command execution, link checking, cross-reference verification)

### V. Cross-Platform Compatibility ✅ PASS
- **Installation Instructions**: INSTALLATION.md must provide working commands for Windows (Scoop), macOS (Homebrew), and Linux (Homebrew/manual) (FR-008)
- **Platform-Specific Troubleshooting**: TROUBLESHOOTING.md must cover Windows Credential Manager, macOS Keychain, Linux Secret Service issues (FR-009)
- **Compliance**: Documentation ensures all platforms covered with accurate platform-specific guidance

### VI. Observability & Auditability ✅ PASS
- **Audit Logging Documentation**: SECURITY.md must document HMAC-SHA256 audit logging, log rotation, verification commands (FR-005)
- **Usage Tracking**: Documentation must explain credential usage tracking by working directory (existing in USAGE.md)
- **Compliance**: Documentation ensures users understand observability features

### VII. Simplicity & YAGNI ✅ PASS
- **No New Files**: Out-of-scope explicitly excludes creating new documentation beyond existing 7 files
- **No Scope Creep**: Out-of-scope excludes translations, video tutorials, API docs, restructuring
- **Compliance**: Documentation review follows YAGNI—only update existing docs, no speculative additions

**Constitution Check Result**: ✅ **ALL GATES PASS**

No violations. Documentation review feature aligns with all applicable constitution principles. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```
specs/008-review-and-update/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── documentation-validation-schema.md
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
# Documentation files to review/update
docs/
├── INSTALLATION.md      # Installation instructions for all platforms
├── SECURITY.md          # Cryptographic implementation, threat model, best practices
├── TROUBLESHOOTING.md   # Platform-specific issue resolution
├── USAGE.md             # Complete command reference, TUI/CLI modes, keyboard shortcuts
├── MIGRATION.md         # PBKDF2 iteration upgrade guide (100k → 600k)
├── KNOWN_LIMITATIONS.md # Current limitations and workarounds
└── archive/             # Historical documentation (not in scope)

README.md                # Primary entry point: Quick Start, features, security summary

# Validation and testing artifacts (created during implementation)
specs/008-review-and-update/
└── validation/
    ├── command-tests.sh         # Script to execute all documented CLI commands
    ├── link-check.sh            # Validate external links (GitHub, NIST, OWASP)
    ├── version-audit.sh         # Grep for outdated version references
    └── cross-reference-check.sh # Verify internal doc links and consistency
```

**Structure Decision**: Documentation-only feature. No source code changes required. Validation scripts stored in spec directory for reproducibility and audit trail.

## Complexity Tracking

*No violations—this section not applicable. All constitution gates pass.*

---

## Phase 0: Research & Validation Strategy

### Research Tasks

1. **Current Release Version Identification**
   - **Known**: Current release version is v0.0.1 (per tasks.md line 88)
   - **Research Task**: Verify `git describe --tags --abbrev=0` matches v0.0.1 and confirm GitHub releases page accuracy
   - **Rationale**: All documentation must reference current version (FR-001)

2. **Implemented Features from Specs 001-007**
   - **Unknown**: Which features from specs 001-007 are implemented and merged to main?
   - **Research Task**: Review git log, merged spec branches, CHANGELOG.md
   - **Rationale**: Documentation must remove outdated references and add new features (FR-010, FR-007)

3. **TUI Keyboard Shortcuts Inventory**
   - **Unknown**: Complete list of TUI keyboard shortcuts in current implementation
   - **Research Task**: Test TUI mode, review spec 007 tasks.md, grep source code for keybindings
   - **Rationale**: USAGE.md must document all shortcuts accurately (FR-007, SC-006: 20+ shortcuts)

4. **Documentation Validation Best Practices**
   - **Research Task**: Identify tools for link checking (linkchecker, markdown-link-check), command execution testing, version reference auditing
   - **Rationale**: Establish validation methodology for SC-003 (100% command accuracy), SC-007 (zero outdated references)

5. **Package Manager Update Status**
   - **Unknown**: Are Homebrew tap and Scoop bucket updated to current release?
   - **Research Task**: Check `brew info pass-cli`, `scoop info pass-cli`, compare versions to git tags
   - **Rationale**: INSTALLATION.md commands must work with current release (FR-008, SC-005)

### Research Consolidation

**Output**: `research.md` containing:
- Current release version and feature inventory
- Complete TUI keyboard shortcuts list (verified against implementation)
- Validation tooling and methodology
- Documentation accuracy baseline (current state assessment)
- Package manager version status

---

## Phase 1: Design & Contracts

### Data Model

**Entity**: Documentation File Validation Record

**Attributes**:
- `file_path`: Absolute path to documentation file (e.g., `docs/SECURITY.md`)
- `validation_status`: Enum (PENDING, IN_PROGRESS, PASS, FAIL)
- `issues_found`: List of validation failures (outdated references, broken links, command errors)
- `last_updated`: Timestamp of last modification
- `validated_by`: Validation method (manual review, automated script, command execution)
- `acceptance_criteria`: List of applicable functional requirements (FR-001, FR-003, etc.)

**Relationships**:
- Documentation file → Functional requirements (many-to-many)
- Documentation file → User stories (many-to-many, traceability)
- Documentation file → CLI commands referenced (one-to-many)

**Validation Rules**:
- All 7 documentation files must have `validation_status = PASS` before completion
- `issues_found` must be empty list for PASS status
- Each file must map to at least one functional requirement

**Output**: `data-model.md`

### Contracts

**Documentation Validation Schema** (OpenAPI-style specification):

```yaml
# contracts/documentation-validation-schema.md

DocumentationValidationContract:
  description: Defines validation requirements for each documentation file

  README.md:
    requirements: [FR-001, FR-002, FR-015]
    validation_criteria:
      - Quick Start executes in <10 minutes (SC-001)
      - Current release version referenced (FR-001)
      - Feature roadmap accurately reflects completed features (FR-015)
    test_method: Manual walkthrough + version grep

  SECURITY.md:
    requirements: [FR-001, FR-003, FR-004, FR-005, FR-013, FR-014]
    validation_criteria:
      - Complete crypto specs (AES-256-GCM, PBKDF2-SHA256 with 600,000 iterations minimum) (FR-003)
      - Password policy documented (12+ chars, complexity) (FR-004)
      - Audit logging documented (HMAC-SHA256, verification) (FR-005)
      - Migration path documented (100k → 600k) (FR-013)
      - TUI security warnings present (FR-014)
    test_method: Specification review + grep for iteration counts

  USAGE.md:
    requirements: [FR-001, FR-006, FR-007, FR-011, FR-012]
    validation_criteria:
      - TUI vs CLI mode distinction clear (FR-006)
      - All TUI keyboard shortcuts documented (20+) (FR-007, SC-006)
      - All CLI commands execute successfully (FR-011, SC-003)
      - Correct file paths referenced (FR-012)
    test_method: Command execution + shortcut verification + path validation

  INSTALLATION.md:
    requirements: [FR-001, FR-008]
    validation_criteria:
      - Homebrew installation commands work (FR-008, SC-005)
      - Scoop installation commands work (FR-008, SC-005)
      - Manual installation steps current (FR-008)
    test_method: Fresh install test on 3 platforms

  TROUBLESHOOTING.md:
    requirements: [FR-001, FR-009]
    validation_criteria:
      - TUI-specific issues covered (rendering, shortcuts, search, toggle, layout) (FR-009)
      - Platform-specific solutions (Windows, macOS, Linux) (FR-009)
      - Solutions findable in <5 minutes (SC-004)
    test_method: Issue simulation + search time measurement

  MIGRATION.md:
    requirements: [FR-001, FR-013]
    validation_criteria:
      - PBKDF2 iteration upgrade path documented (100k → 600k) (FR-013)
      - Migration command accurate
    test_method: Command execution + cross-reference with SECURITY.md

  KNOWN_LIMITATIONS.md:
    requirements: [FR-001]
    validation_criteria:
      - Current limitations accurate
      - Outdated limitations removed
    test_method: Manual review
```

**Output**: `contracts/documentation-validation-schema.md`

### Quickstart Guide

**Purpose**: Provide implementation team with minimal steps to validate documentation accuracy.

**Steps**:

1. **Setup Validation Environment**
   ```bash
   # Ensure current release binary installed
   pass-cli version

   # Clone repository and checkout feature branch
   git checkout 008-review-and-update
   ```

2. **Run Automated Validation**
   ```bash
   # Execute command tests (SC-003: 100% command accuracy)
   cd specs/008-review-and-update/validation
   ./command-tests.sh

   # Check for outdated version references (SC-007: zero outdated refs)
   ./version-audit.sh

   # Validate external links
   ./link-check.sh

   # Verify cross-references
   ./cross-reference-check.sh
   ```

3. **Manual Review Checklist**
   - [ ] README Quick Start completes in 10 minutes (SC-001)
   - [ ] SECURITY.md contains all crypto parameters (SC-002)
   - [ ] USAGE.md documents all 20+ TUI shortcuts (SC-006)
   - [ ] TROUBLESHOOTING.md solutions findable in 5 minutes (SC-004)
   - [ ] INSTALLATION.md commands work on all platforms (SC-005)

4. **Fix Issues and Re-validate**
   - Update documentation files based on validation failures
   - Re-run automated validation
   - Repeat until all validation status = PASS

**Output**: `quickstart.md`

---

## Phase 2: Task Decomposition (Generated by `/speckit.tasks`)

**Note**: Task generation deferred to `/speckit.tasks` command. This plan provides the foundation:

**Expected Task Categories**:
1. **Validation Script Creation** (4 tasks: command tests, version audit, link check, cross-reference check)
2. **Documentation File Review** (7 tasks: one per file—README, SECURITY, USAGE, INSTALLATION, TROUBLESHOOTING, MIGRATION, KNOWN_LIMITATIONS)
3. **Manual Testing** (5 tasks: Quick Start walkthrough, security spec review, TUI shortcut verification, installation testing, troubleshooting simulation)
4. **Cross-Reference Verification** (3 tasks: internal links, external links, version consistency)
5. **Final Validation** (1 task: Execute all validation scripts and confirm PASS status)

**Estimated Total**: 20 tasks across 5 categories

---

## Success Metrics Alignment

| Success Criterion | Validation Method | Acceptance Threshold |
|-------------------|-------------------|----------------------|
| SC-001: 10-minute onboarding | Manual walkthrough with timer | New user completes install + first credential ≤10 min |
| SC-002: Crypto params identifiable | SECURITY.md specification review | All parameters (algorithm, key size, iterations, nonce) documented |
| SC-003: 100% command accuracy | Automated command execution script | All documented commands execute without errors |
| SC-004: 5-minute troubleshooting | Manual search time measurement | Common issues findable ≤5 min |
| SC-005: Installation methods work | Fresh install test on 3 platforms | Homebrew, Scoop, manual all succeed |
| SC-006: 20+ shortcuts documented | TUI shortcut verification script | All shortcuts match implementation |
| SC-007: Zero outdated references | Automated version audit script | No 100k iteration counts, no unimplemented features |
| SC-008: Script examples run | Script execution tests (Bash, PowerShell, Python) | All examples execute without modification |

---

**Next Command**: `/speckit.tasks` to generate actionable task breakdown

**Status**: Phase 0 research pending. Phase 1 design complete. Constitution gates passed.
