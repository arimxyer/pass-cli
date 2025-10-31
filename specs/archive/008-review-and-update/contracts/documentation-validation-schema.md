# Documentation Validation Contract

**Feature**: Documentation Review and Production Release Preparation
**Branch**: `008-review-and-update`
**Date**: 2025-01-14
**Version**: 1.0.0

## Overview

This contract defines the validation requirements, acceptance criteria, and testing methods for each documentation file in the Pass-CLI repository. All documentation must pass these validation checks before production release.

## Validation Schema

### README.md

**Location**: `R:\Test-Projects\pass-cli\README.md`

**Functional Requirements**: FR-001, FR-002, FR-015

**Validation Criteria**:

1. **Version Accuracy (FR-001)**
   - Current release version number appears in file
   - No references to outdated versions
   - Links to GitHub releases page are current

2. **Quick Start Completeness (FR-002)**
   - Installation section provides commands for Homebrew, Scoop, manual
   - First Steps section demonstrates: init → add → get → copy workflow
   - All Quick Start commands execute successfully without errors
   - Workflow completable in ≤10 minutes by new user (SC-001)

3. **Feature Roadmap Accuracy (FR-015)**
   - Completed features marked with `[x]` checkbox
   - In-progress features marked with `[ ]` checkbox
   - No implemented features listed as "planned" or "future"
   - Roadmap reflects specs 001-007 implementation status

4. **TUI Keyboard Shortcuts (FR-007)**
   - Key TUI Shortcuts table includes 20+ shortcuts (not just 6)
   - Shortcuts match actual implementation from spec 007
   - References full keyboard shortcuts in USAGE.md

5. **Configuration Documentation (FR-010)**
   - Documents config.yml feature from spec 007
   - Explains keybinding customization capability
   - Provides example configuration path and format

**Test Methods**:
- Manual walkthrough with timer (SC-001: 10-minute target)
- Command execution (all Quick Start commands)
- Version grep: `grep -E "(v?[0-9]+\.[0-9]+\.[0-9]+|iteration)" README.md`
- Keyboard shortcut count: verify table has 20+ rows
- Cross-reference with spec 007 tasks.md for config features

**Acceptance Threshold**: All validation criteria PASS, zero issues

---

### SECURITY.md

**Location**: `R:\Test-Projects\pass-cli\docs\SECURITY.md`

**Functional Requirements**: FR-001, FR-003, FR-004, FR-005, FR-013, FR-014

**Validation Criteria**:

1. **Complete Cryptographic Specifications (FR-003)**
   - Algorithm: AES-256-GCM explicitly documented
   - Key Derivation: PBKDF2-SHA256 with 600,000 iterations (not 100,000)
   - Salt: 32-byte random salt per vault
   - Nonce: 12-byte unique nonce per encryption
   - Random sources: `crypto/rand` (Windows `CryptGenRandom`, Unix `/dev/urandom`)
   - NIST compliance references present (SP 800-38D for GCM, SP 800-132 for PBKDF2)

2. **Password Policy Documentation (FR-004)**
   - Minimum 12 characters enforced
   - Complexity requirements: uppercase, lowercase, digit, symbol
   - Enforcement applies to vault AND credential passwords
   - January 2025 introduction date mentioned
   - TUI strength indicator referenced

3. **Audit Logging Documentation (FR-005)**
   - HMAC-SHA256 tamper-evident signatures explained
   - HMAC key storage in OS keychain documented
   - Log rotation policy: 10MB with 7-day retention
   - Verification command documented: `pass-cli verify-audit`
   - Privacy guarantee: service names logged, passwords NEVER logged
   - Opt-in nature clarified: `--enable-audit` flag

4. **Migration Path (FR-013)**
   - Documents upgrade from 100k to 600k PBKDF2 iterations
   - References MIGRATION.md for detailed instructions
   - Explains backward compatibility (old vaults continue to work)
   - Performance impact documented (~50-100ms on modern CPUs)

5. **TUI Security Warnings (FR-014)**
   - Shoulder surfing risk in TUI mode
   - Screen recording exposure (service names, usernames visible)
   - Shared terminal session dangers
   - Password visibility toggle (`Ctrl+H`) security considerations
   - Recommendations for secure usage environments

**Test Methods**:
- Specification review (security professional validates completeness per SC-002)
- Iteration count grep: `grep -c "600,000\|600k" docs/SECURITY.md` (should be >0)
- Old iteration grep: `grep -c "100,000\|100k" docs/SECURITY.md` (should be 0 unless in migration context)
- NIST reference check: verify links to SP 800-38D and SP 800-132
- Cross-reference with MIGRATION.md for consistency

**Acceptance Threshold**: Security professional can identify all crypto parameters without reading source code (SC-002)

---

### USAGE.md

**Location**: `R:\Test-Projects\pass-cli\docs\USAGE.md`

**Functional Requirements**: FR-001, FR-006, FR-007, FR-011, FR-012

**Validation Criteria**:

1. **TUI vs CLI Mode Distinction (FR-006)**
   - Clear explanation: `pass-cli` (no args) launches TUI
   - Clear explanation: `pass-cli <command>` uses CLI mode
   - "TUI vs CLI Mode" section exists with comparison table
   - Script examples use explicit commands (never bare `pass-cli`)

2. **TUI Keyboard Shortcuts (FR-007)**
   - All 20+ keyboard shortcuts documented in table
   - Shortcuts match spec 007 implementation:
     - Navigation: Tab, Shift+Tab, ↑/↓, Enter
     - Actions: n (new), e (edit), d (delete), p (password visibility), c (copy)
     - View: i (toggle detail), s (toggle sidebar), / (search)
     - Forms: Ctrl+S (save), Ctrl+H (toggle password), Esc (cancel)
     - General: ? (help), q (quit), Ctrl+C (quit)
   - Context column indicates where each shortcut works
   - Custom keybinding documentation references config.yml

3. **CLI Command Accuracy (FR-011)**
   - All documented commands execute successfully
   - All documented flags recognized by binary
   - Examples produce expected output format
   - No references to removed or unimplemented features

4. **File Path Accuracy (FR-012)**
   - Vault location: `%USERPROFILE%\.pass-cli\vault.enc` (Windows), `~/.pass-cli/vault.enc` (Unix)
   - Config location: `~/.pass-cli/config.yaml` (if documented)
   - Audit log location: `~/.pass-cli/audit.log` (if audit enabled)
   - All paths match actual implementation

**Test Methods**:
- TUI mode clarity: manual review of "TUI Mode" section
- Keyboard shortcut count: verify table has 20+ rows with context
- Command execution script:
  ```bash
  # Extract all `pass-cli <command>` examples
  # Execute each command with --help to verify it exists
  # Log failures
  ```
- Path validation: grep for file paths and cross-reference with source code
- Cross-reference with spec 007 for keybinding completeness

**Acceptance Threshold**: 100% of documented CLI commands execute successfully (SC-003), all 20+ TUI shortcuts documented (SC-006)

---

### INSTALLATION.md

**Location**: `R:\Test-Projects\pass-cli\docs\INSTALLATION.md`

**Functional Requirements**: FR-001, FR-008

**Validation Criteria**:

1. **Homebrew Installation (FR-008)**
   - Tap command: `brew tap ari1110/homebrew-tap`
   - Install command: `brew install pass-cli`
   - Verify command: `pass-cli version`
   - Commands execute successfully on macOS and Linux

2. **Scoop Installation (FR-008)**
   - Bucket command: `scoop bucket add pass-cli https://github.com/ari1110/scoop-bucket`
   - Install command: `scoop install pass-cli`
   - Verify command: `pass-cli version`
   - Commands execute successfully on Windows

3. **Manual Installation (FR-008)**
   - Download links point to current release (v0.0.1 or later)
   - Checksum verification instructions accurate
   - Binary placement instructions for all platforms
   - PATH setup instructions for Windows, macOS, Linux

4. **Build from Source (FR-008)**
   - Go version requirement: 1.25 or later
   - Build commands: `go build -o pass-cli .` or `make build`
   - Test commands: `go test ./...` or `make test`

**Test Methods**:
- Fresh install test on 3 platforms:
  - macOS: Homebrew installation
  - Windows: Scoop installation
  - Linux: Manual binary installation
- Verify all commands execute without errors
- Check package manager versions: `brew info pass-cli`, `scoop info pass-cli`
- Validate release links point to existing release assets

**Acceptance Threshold**: All installation methods complete successfully (SC-005)

---

### TROUBLESHOOTING.md

**Location**: `R:\Test-Projects\pass-cli\docs\TROUBLESHOOTING.md`

**Functional Requirements**: FR-001, FR-009

**Validation Criteria**:

1. **TUI-Specific Issues (FR-009)**
   - TUI Display Garbled or Has Rendering Artifacts
   - Keyboard Shortcuts Not Working
   - TUI Launches to Black or Unresponsive Screen
   - Search Function (`/`) Not Filtering Results
   - Ctrl+H Password Toggle Not Working
   - Sidebar or Detail Panel Not Visible
   - Usage Locations Not Appearing in Detail Panel

2. **Platform-Specific Solutions (FR-009)**
   - Windows: Credential Manager issues, PowerShell execution policy
   - macOS: Keychain prompt issues, Gatekeeper warnings
   - Linux: D-Bus session issues, Secret Service setup, SELinux/AppArmor

3. **Solution Quality**
   - Each issue has actionable solution (not just "check the docs")
   - Solutions include specific commands or configuration steps
   - Platform-specific instructions clearly labeled
   - Common issues findable in ≤5 minutes via search/scan (SC-004)

**Test Methods**:
- Issue simulation: trigger common errors and verify solutions work
- Search time measurement: find solution for "TUI rendering" issue
- Platform coverage check: verify Windows, macOS, Linux each have ≥5 unique troubleshooting entries
- Command validation: execute all troubleshooting commands to verify they work

**Acceptance Threshold**: Users find solutions in ≤5 minutes (SC-004)

---

### MIGRATION.md

**Location**: `R:\Test-Projects\pass-cli\docs\MIGRATION.md`

**Functional Requirements**: FR-001, FR-013

**Validation Criteria**:

1. **PBKDF2 Iteration Upgrade (FR-013)**
   - Documents migration from 100,000 to 600,000 iterations
   - Explains why upgrade needed (security hardening, OWASP recommendations)
   - Provides migration command: `pass-cli migrate` or equivalent
   - Explains backward compatibility (old vaults work, but should upgrade)

2. **Migration Process**
   - Step-by-step instructions
   - Backup recommendations before migration
   - Expected performance impact (~50-100ms increase)
   - Verification steps after migration

3. **Consistency with SECURITY.md**
   - Iteration counts match SECURITY.md (600k is current)
   - Migration rationale aligns with security documentation

**Test Methods**:
- Command execution: verify migration command exists and works
- Cross-reference with SECURITY.md for iteration count consistency
- Grep for outdated references: `grep -c "100,000\|100k" docs/MIGRATION.md`

**Acceptance Threshold**: Migration path clearly documented, cross-references consistent

---

### KNOWN_LIMITATIONS.md

**Location**: `R:\Test-Projects\pass-cli\docs\KNOWN_LIMITATIONS.md`

**Functional Requirements**: FR-001

**Validation Criteria**:

1. **Current Limitations Accurate**
   - Listed limitations exist in current release
   - No limitations that have been resolved (e.g., no mention of "no TUI mode" if TUI now exists)

2. **Outdated Limitations Removed**
   - Review specs 001-007 to identify resolved limitations
   - Remove any limitations that were addressed by implemented features

**Test Methods**:
- Manual review against current binary capabilities
- Cross-reference with specs 001-007 to identify resolved issues
- Verify each limitation with actual testing

**Acceptance Threshold**: All listed limitations current, no resolved limitations remain

---

## Global Validation Rules

**Applies to ALL documentation files**:

1. **Version Consistency (FR-001)**
   - All files reference the same current release version
   - No conflicting version numbers across files
   - PBKDF2 iteration count is consistently 600,000 (not 100,000)

2. **Link Validity**
   - All internal links (to other docs) resolve correctly
   - All external links (GitHub, NIST, OWASP) return HTTP 200
   - Package manager links (Homebrew tap, Scoop bucket) are current

3. **Cross-Reference Consistency**
   - When file A references file B, file B contains the referenced content
   - Example: if SECURITY.md references MIGRATION.md, MIGRATION.md must exist and contain migration instructions

4. **Formatting Standards**
   - Valid GitHub-flavored Markdown
   - Code blocks have language specifiers
   - Tables formatted correctly
   - Headings follow hierarchy (no H3 without H2)

## Validation Automation

**Automated Scripts** (to be created in `specs/008-review-and-update/validation/`):

1. **command-tests.sh**
   - Extracts all `pass-cli <command>` examples from documentation
   - Executes `<command> --help` to verify command exists
   - Logs failures with file and line number
   - Exit code 0 if all commands valid, 1 if any failures

2. **version-audit.sh**
   - Greps all documentation for version references
   - Identifies inconsistencies (multiple different versions)
   - Flags outdated iteration counts (100,000 instead of 600,000)
   - Outputs: list of files with version issues

3. **link-check.sh**
   - Uses `markdown-link-check` or equivalent
   - Validates all HTTP/HTTPS links return 200
   - Validates all internal file links resolve
   - Outputs: list of broken links with URLs

4. **cross-reference-check.sh**
   - Parses cross-references between files
   - Verifies referenced content exists in target file
   - Example: "See MIGRATION.md" → MIGRATION.md must exist
   - Outputs: list of broken cross-references

## Success Criteria Mapping

| Success Criterion | Validation Files | Contract Sections |
|-------------------|------------------|-------------------|
| SC-001: 10-minute onboarding | README.md | README.md validation criteria #2 |
| SC-002: Crypto params identifiable | SECURITY.md | SECURITY.md validation criteria #1 |
| SC-003: 100% command accuracy | USAGE.md, README.md | USAGE.md validation criteria #3 + automation script |
| SC-004: 5-minute troubleshooting | TROUBLESHOOTING.md | TROUBLESHOOTING.md validation criteria #3 |
| SC-005: Installation methods work | INSTALLATION.md | INSTALLATION.md validation criteria #1-#3 |
| SC-006: 20+ shortcuts documented | USAGE.md | USAGE.md validation criteria #2 |
| SC-007: Zero outdated references | ALL FILES | Global validation rule #1 + automation script |
| SC-008: Script examples run | USAGE.md, README.md | Command execution testing |

## Validation Workflow

1. **Initialize**: Create validation tracking for all 7 files (per data-model.md)
2. **Automate**: Run 4 validation scripts (command-tests, version-audit, link-check, cross-reference)
3. **Manual Review**: Validate criteria not covered by automation (10-minute onboarding, security completeness, troubleshooting quality)
4. **Document Issues**: Record all validation failures in data model (ValidationIssue entities)
5. **Fix**: Update documentation files to resolve issues
6. **Re-validate**: Run automation + manual review again
7. **Sign-off**: Mark each file `validation_status = PASS` when all criteria met

## Contract Versioning

- **Version 1.0.0**: Initial contract (2025-01-14)
- Future versions will document changes to validation criteria as release requirements evolve

---

**Acceptance**: All 7 documentation files must pass ALL validation criteria in this contract before production release approval.
