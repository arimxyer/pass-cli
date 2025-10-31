# Tasks: Documentation Review and Production Release Preparation

**Input**: Design documents from `/specs/008-review-and-update/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/documentation-validation-schema.md, quickstart.md

**Tests**: No automated test generation for documentation review. Validation is manual + scripted checks.

**Organization**: Tasks are grouped by user story (5 user stories, P1-P3 priorities) to enable independent validation and review of each documentation concern.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

## Path Conventions
- **Documentation files**: `R:\Test-Projects\pass-cli\docs\`, `R:\Test-Projects\pass-cli\README.md`
- **Validation scripts**: `R:\Test-Projects\pass-cli\specs\008-review-and-update\validation\`
- **Spec artifacts**: `R:\Test-Projects\pass-cli\specs\008-review-and-update\`

---

## Phase 1: Setup (Validation Infrastructure)

**Purpose**: Create validation scripts and tracking infrastructure for documentation review

- [X] **T001** [P] Create validation scripts directory at `specs/008-review-and-update/validation/`
- [X] **T002** [P] Create validation tracking spreadsheet or JSON file at `specs/008-review-and-update/validation/tracking.json` (based on data-model.md structure)
- [ ] **T003** [P] Install validation tools: `npm install -g markdown-link-check` (optional but recommended)

**Checkpoint**: Validation infrastructure ready

---

## Phase 2: Foundational (Automated Validation Scripts)

**Purpose**: Core validation scripts that ALL user stories depend on for accuracy verification

**⚠️ CRITICAL**: These scripts must be complete and working before ANY documentation file review can begin

- [X] **T004** Create `specs/008-review-and-update/validation/version-audit.sh` script
  - Grep all docs for version references
  - Flag inconsistencies (multiple versions)
  - Flag outdated iteration counts (100,000 vs 600,000)
  - Output: List of files with version issues
  - Exit code 0 if clean, 1 if issues found

- [X] **T005** [P] Create `specs/008-review-and-update/validation/command-tests.sh` script
  - Extract all `pass-cli <command>` examples from docs
  - Execute `<command> --help` to verify command exists
  - Log failures with file and line number
  - Output: List of invalid commands
  - Exit code 0 if all valid, 1 if failures

- [X] **T006** [P] Create `specs/008-review-and-update/validation/link-check.sh` script
  - Use markdown-link-check or curl to validate HTTP/HTTPS links
  - Validate internal file links resolve
  - Handle 404 errors gracefully: suggest archive.org fallback URLs (edge case E1 fix)
  - Output: List of broken links with URLs and suggested fallbacks
  - Exit code 0 if all links valid, 1 if broken links

- [X] **T007** [P] Create `specs/008-review-and-update/validation/cross-reference-check.sh` script
  - Parse cross-references between files (e.g., "See MIGRATION.md")
  - Verify referenced content exists in target file
  - Output: List of broken cross-references
  - Exit code 0 if all valid, 1 if broken refs

- [X] **T008** Create `specs/008-review-and-update/validation/run-all-validation.sh` wrapper script
  - Execute all 4 validation scripts in parallel
  - Collect exit codes
  - Generate summary report
  - Exit code 0 only if all scripts pass

**Checkpoint**: Foundation ready - documentation file reviews can now begin in parallel

---

## Phase 3: User Story 1 - New User Onboarding Success (Priority: P1) 🎯 MVP

**Goal**: Ensure README.md provides accurate Quick Start that works within 10 minutes for new users

**Independent Test**: New user walkthrough with timer - install, init, add credential, retrieve credential ≤10 minutes

**FR Coverage**: FR-001, FR-002, FR-007, FR-010, FR-015

### Implementation for User Story 1

- [X] **T009** [US1] Review `R:\Test-Projects\pass-cli\README.md` - Version accuracy (FR-001)
  - **Note**: T004 version-audit.sh handles automated version validation across all files
  - Manual review: Verify version v0.0.1 appears in appropriate contexts (installation, Quick Start)
  - Verify no outdated version references in prose or examples
  - Update version numbers if needed

- [X] **T010** [US1] Review `R:\Test-Projects\pass-cli\README.md` - Quick Start completeness (FR-002)
  - Verify installation commands for Homebrew, Scoop, manual present
  - Verify First Steps demonstrates: init → add → get → copy workflow
  - Execute all Quick Start commands with timer
  - Target: ≤10 minutes for new user (SC-001)
  - Fix any command errors or missing steps

- [X] **T011** [US1] Review `R:\Test-Projects\pass-cli\README.md` - TUI keyboard shortcuts (FR-007)
  - Current: Table shows only 6 shortcuts (issue from research.md)
  - Update table to include all 20+ keyboard shortcuts (standardized terminology: "keyboard shortcuts" not "keybindings" or "key hints")
  - Verify shortcuts match spec 007 implementation
  - Add reference link to full docs: `docs/USAGE.md#tui-keyboard-shortcuts`

- [X] **T012** [US1] Review `R:\Test-Projects\pass-cli\README.md` - Configuration documentation (FR-010)
  - Add section documenting config.yml feature from spec 007
  - Explain keybinding customization capability
  - Provide example config path: `~/.pass-cli/config.yaml`
  - Reference full config docs: `docs/USAGE.md#configuration`

- [X] **T013** [US1] Review `R:\Test-Projects\pass-cli\README.md` - Feature roadmap accuracy (FR-015)
  - Review specs 001-007 implementation status from research.md
  - Mark completed features with `[x]` checkbox
  - Ensure no implemented features listed as "planned"
  - Remove references to features not implemented

- [ ] **T014** [US1] Validate README.md using automated scripts
  - Run `specs/008-review-and-update/validation/command-tests.sh` on README.md
  - Run `specs/008-review-and-update/validation/link-check.sh` on README.md
  - Fix any failures identified
  - Mark validation_status = PASS in tracking.json

**Checkpoint**: README.md validated - User Story 1 complete and independently testable

---

## Phase 4: User Story 2 - Production Deployment Confidence (Priority: P1)

**Goal**: Ensure SECURITY.md provides complete crypto specs, audit logging docs, and threat model for security teams

**Independent Test**: Security professional reviews SECURITY.md and confirms all crypto parameters identifiable without reading source code (SC-002)

**FR Coverage**: FR-001, FR-003, FR-004, FR-005, FR-013, FR-014

### Implementation for User Story 2

- [X] **T015** [US2] Review `R:\Test-Projects\pass-cli\docs\SECURITY.md` - Cryptographic specifications (FR-003)
  - Verify AES-256-GCM explicitly documented
  - Verify PBKDF2-SHA256 with **600,000 iterations minimum** (not 100,000) per constitution Principle I
  - Verify 32-byte salt, 12-byte nonce documented
  - Verify random sources: `crypto/rand` documented
  - Add NIST compliance references: SP 800-38D (GCM), SP 800-132 (PBKDF2)
  - Ensure all parameters visible without reading source code

- [X] **T016** [US2] Review `R:\Test-Projects\pass-cli\docs\SECURITY.md` - Password policy documentation (FR-004)
  - Verify 12+ character minimum documented
  - Verify complexity requirements: uppercase, lowercase, digit, symbol
  - Verify applies to vault AND credential passwords
  - Add January 2025 introduction date
  - Reference TUI strength indicator feature

- [X] **T017** [US2] Review `R:\Test-Projects\pass-cli\docs\SECURITY.md` - Audit logging documentation (FR-005)
  - Verify HMAC-SHA256 tamper-evident signatures explained
  - Verify HMAC key storage in OS keychain documented
  - Verify log rotation policy: 10MB, 7-day retention
  - Verify verification command: `pass-cli verify-audit`
  - Verify privacy guarantee: service names logged, passwords NEVER logged
  - Verify opt-in nature: `--enable-audit` flag

- [X] **T018** [US2] Review `R:\Test-Projects\pass-cli\docs\SECURITY.md` - Migration path (FR-013)
  - Document upgrade from 100,000 to 600,000 PBKDF2-SHA256 iterations per constitution Principle I
  - Reference MIGRATION.md for detailed instructions
  - Explain backward compatibility (old vaults work)
  - Document performance impact (~50-100ms on modern CPUs)
  - Ensure cross-reference consistency with MIGRATION.md

- [X] **T019** [US2] Review `R:\Test-Projects\pass-cli\docs\SECURITY.md` - TUI security warnings (FR-014)
  - Add shoulder surfing risk warning
  - Add screen recording exposure warning (service names visible)
  - Add shared terminal session dangers
  - Add password visibility toggle (`Ctrl+H`) security considerations
  - Add recommendations for secure usage environments

- [ ] **T020** [US2] Validate SECURITY.md using automated scripts
  - Run `specs/008-review-and-update/validation/version-audit.sh` on SECURITY.md
  - Verify zero "100,000" or "100k" references outside migration context
  - Verify ≥1 explicit "600,000 iterations" references (no shorthand "600k")
  - Run `specs/008-review-and-update/validation/link-check.sh` on SECURITY.md
  - Mark validation_status = PASS in tracking.json

**Checkpoint**: SECURITY.md validated - User Story 2 complete and independently testable

---

## Phase 5: User Story 3 - CLI vs TUI Mode Clarity (Priority: P2)

**Goal**: Ensure USAGE.md clearly distinguishes TUI vs CLI modes and documents all keyboard shortcuts for script authors

**Independent Test**: Script author reads USAGE.md and successfully writes automation that avoids accidental TUI launches

**FR Coverage**: FR-001, FR-006, FR-007, FR-011, FR-012

### Implementation for User Story 3

- [X] **T021** [US3] Review `R:\Test-Projects\pass-cli\docs\USAGE.md` - TUI vs CLI mode distinction (FR-006)
  - Verify "TUI vs CLI Mode" section exists with clear explanations
  - Ensure `pass-cli` (no args) → TUI mode clearly stated
  - Ensure `pass-cli <command>` → CLI mode clearly stated
  - Add comparison table showing mode triggers
  - Verify all script examples use explicit commands (never bare `pass-cli`)

- [X] **T022** [US3] Review `R:\Test-Projects\pass-cli\docs\USAGE.md` - TUI keyboard shortcuts (FR-007)
  - Current: May be incomplete (research.md identified 20+ shortcuts)
  - Update "TUI Keyboard Shortcuts" table to include all shortcuts (minimum 20+):
    - Navigation: Tab, Shift+Tab, ↑/↓, Enter
    - Actions: n (new), e (edit), d (delete), p (copy password), c (copy username)
    - View: i (toggle detail panel), s (toggle sidebar), / (search/filter)
    - Forms: Ctrl+S (save), Ctrl+H (toggle password visibility), Esc (cancel/clear search)
    - General: ? (help), q (quit), Ctrl+C (force quit)
  - Add context column (which view each shortcut works in: list view, detail view, form view)
  - Document custom keybinding capability via config.yml (reference spec 007)
  - Verify toggle detail shortcut is 'i' by testing TUI mode directly

- [X] **T023** [US3] Review `R:\Test-Projects\pass-cli\docs\USAGE.md` - CLI command accuracy (FR-011)
  - Run `specs/008-review-and-update/validation/command-tests.sh` on USAGE.md
  - Verify all documented commands execute successfully
  - Verify all documented flags recognized by binary
  - Remove any references to unimplemented features
  - Fix command examples to match current release

- [X] **T024** [US3] Review `R:\Test-Projects\pass-cli\docs\USAGE.md` - File path accuracy (FR-012)
  - Verify vault location: `%USERPROFILE%\.pass-cli\vault.enc` (Windows), `~/.pass-cli/vault.enc` (Unix)
  - Verify config location: `~/.pass-cli/config.yaml` (from spec 007)
  - Verify audit log location: `~/.pass-cli/audit.log`
  - Ensure all paths match actual implementation

- [ ] **T025** [US3] Validate USAGE.md using automated scripts
  - Run `specs/008-review-and-update/validation/command-tests.sh` on USAGE.md
  - Target: 100% command accuracy (SC-003)
  - Verify 20+ TUI shortcuts documented (SC-006)
  - Run `specs/008-review-and-update/validation/link-check.sh` on USAGE.md
  - Mark validation_status = PASS in tracking.json

**Checkpoint**: USAGE.md validated - User Story 3 complete and independently testable

---

## Phase 6: User Story 4 - Version-Specific Accuracy (Priority: P2)

**Goal**: Ensure INSTALLATION.md and MIGRATION.md provide working installation/migration commands for current release

**Independent Test**: Fresh install test on 3 platforms (Homebrew, Scoop, manual) - all succeed

**FR Coverage**: FR-001, FR-008, FR-013

### Implementation for User Story 4

- [X] **T025a** [US4] Verify MIGRATION.md exists or create skeleton (FR-013 prerequisite)
  - Check if `R:\Test-Projects\pass-cli\docs\MIGRATION.md` exists
  - If missing, create skeleton documenting PBKDF2-SHA256 iteration count migration from 100,000 to 600,000 iterations
  - Include sections: Security Rationale, Migration Command, Backward Compatibility, Performance Impact
  - Ensure consistency with constitution Principle I (600,000 iterations minimum)
  - If exists, proceed to T030

- [X] **T026** [US4] Review `R:\Test-Projects\pass-cli\docs\INSTALLATION.md` - Homebrew installation (FR-008)
  - Verify tap command: `brew tap ari1110/homebrew-tap`
  - Verify install command: `brew install pass-cli`
  - Verify verify command: `pass-cli version`
  - Test on macOS/Linux if available, or verify Homebrew tap is current
  - Check package manager status from research.md

- [X] **T027** [US4] Review `R:\Test-Projects\pass-cli\docs\INSTALLATION.md` - Scoop installation (FR-008)
  - Verify bucket command: `scoop bucket add pass-cli https://github.com/ari1110/scoop-bucket`
  - Verify install command: `scoop install pass-cli`
  - Verify verify command: `pass-cli version`
  - Test on Windows if available, or verify Scoop bucket is current
  - Check package manager status from research.md

- [X] **T028** [US4] Review `R:\Test-Projects\pass-cli\docs\INSTALLATION.md` - Manual installation (FR-008)
  - Verify download links point to current release (v0.0.1 or later from research.md)
  - Verify checksum verification instructions accurate
  - Verify binary placement instructions for all platforms
  - Verify PATH setup instructions for Windows, macOS, Linux

- [X] **T029** [US4] Review `R:\Test-Projects\pass-cli\docs\INSTALLATION.md` - Build from source (FR-008)
  - Verify Go version requirement: 1.25 or later
  - Verify build commands: `go build -o pass-cli .` or `make build`
  - Verify test commands: `go test ./...` or `make test`
  - Test build commands if Go installed

- [X] **T030** [US4] Review `R:\Test-Projects\pass-cli\docs\MIGRATION.md` - PBKDF2 iteration upgrade (FR-013)
  - **Prerequisite**: T025a must complete (MIGRATION.md exists)
  - Document migration from 100,000 to 600,000 PBKDF2-SHA256 iterations per constitution Principle I
  - Explain security rationale (OWASP recommendations, NIST SP 800-132)
  - Provide migration command: `pass-cli migrate` or equivalent
  - Explain backward compatibility (old vaults work)
  - Document performance impact (~50-100ms increase)
  - Ensure consistency with SECURITY.md (cross-reference T020 completion recommended)

- [ ] **T031** [US4] Validate INSTALLATION.md and MIGRATION.md using automated scripts
  - **Prerequisite**: T020 completion recommended (SECURITY.md validated before cross-reference check)
  - Run `specs/008-review-and-update/validation/version-audit.sh` on both files
  - Run `specs/008-review-and-update/validation/command-tests.sh` on both files
  - Run `specs/008-review-and-update/validation/link-check.sh` on both files
  - Run `specs/008-review-and-update/validation/cross-reference-check.sh` (SECURITY ↔ MIGRATION consistency - requires T020 SECURITY.md accuracy)
  - Mark validation_status = PASS in tracking.json for both files

**Checkpoint**: INSTALLATION.md and MIGRATION.md validated - User Story 4 complete

---

## Phase 7: User Story 5 - Troubleshooting Self-Service (Priority: P3)

**Goal**: Ensure TROUBLESHOOTING.md and KNOWN_LIMITATIONS.md provide actionable solutions findable in ≤5 minutes

**Independent Test**: Simulate common errors, verify solutions found in ≤5 minutes via search/scan (SC-004)

**FR Coverage**: FR-001, FR-009

### Implementation for User Story 5

- [X] **T032** [US5] Review `R:\Test-Projects\pass-cli\docs\TROUBLESHOOTING.md` - TUI-specific issues (FR-009)
  - Verify coverage: TUI rendering artifacts, keyboard shortcuts not working, black screen, search not filtering, Ctrl+H toggle, sidebar/detail panel visibility
  - Ensure each issue has actionable solution with specific commands
  - Verify platform-specific solutions: Windows (Credential Manager), macOS (Keychain), Linux (Secret Service, D-Bus)
  - Add TUI issues from spec 007 if missing (e.g., toggle detail panel 'i' key)

- [X] **T033** [US5] Review `R:\Test-Projects\pass-cli\docs\TROUBLESHOOTING.md` - Solution quality
  - Time yourself finding solution for "TUI rendering" issue (should be ≤5 minutes)
  - Verify each solution is actionable (not just "check the docs")
  - Verify commands in solutions execute successfully
  - Organize sections for findability (clear headings, table of contents)

- [X] **T034** [US5] Review `R:\Test-Projects\pass-cli\docs\KNOWN_LIMITATIONS.md` - Current limitations accurate
  - Review specs 001-007 from research.md to identify resolved limitations
  - Remove any limitations that were addressed (e.g., no "no TUI mode" if TUI exists)
  - Verify each listed limitation exists in current release
  - Test each limitation to confirm accuracy

- [ ] **T035** [US5] Validate TROUBLESHOOTING.md and KNOWN_LIMITATIONS.md using automated scripts
  - Run `specs/008-review-and-update/validation/command-tests.sh` on TROUBLESHOOTING.md
  - Run `specs/008-review-and-update/validation/link-check.sh` on both files
  - Measure search time for common issue (should be ≤5 minutes for SC-004)
  - Mark validation_status = PASS in tracking.json for both files

**Checkpoint**: TROUBLESHOOTING.md and KNOWN_LIMITATIONS.md validated - User Story 5 complete

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, cross-reference consistency, and documentation quality improvements

- [ ] **T036** [P] Run comprehensive validation across ALL documentation files
  - Execute `specs/008-review-and-update/validation/run-all-validation.sh`
  - Target: Exit code 0 (all scripts pass)
  - Generate validation summary report

- [ ] **T037** [P] Verify cross-reference consistency across all files
  - SECURITY.md ↔ MIGRATION.md (iteration counts match)
  - README.md → USAGE.md (shortcut references match)
  - README.md → INSTALLATION.md (install commands match)
  - All files → current release version (consistent)

- [ ] **T037a** [P] Verify FR-012 file path accuracy across all documentation (coverage gap fix)
  - Grep all 7 documentation files for file path references
  - Verify vault location: `%USERPROFILE%\.pass-cli\vault.enc` (Windows), `~/.pass-cli/vault.enc` (Unix)
  - Verify config location: `~/.pass-cli/config.yaml`
  - Verify audit log location: `~/.pass-cli/audit.log`
  - Cross-reference with actual implementation (test Pass-CLI binary if available)
  - Mark all path references as validated in tracking.json

- [ ] **T038** [P] Verify all 8 success criteria met
  - SC-001: README Quick Start ≤10 minutes (manual walkthrough)
  - SC-002: SECURITY.md crypto params complete (security review)
  - SC-003: 100% CLI command accuracy (command-tests.sh)
  - SC-004: TROUBLESHOOTING.md ≤5 minutes (search time test)
  - SC-005: INSTALLATION.md all methods work (platform testing)
  - SC-006: USAGE.md 20+ shortcuts documented (count verification)
  - SC-007: Zero outdated references (version-audit.sh)
  - SC-008: Script examples run successfully (command execution)

- [ ] **T038a** [P] Execute script examples for SC-008 validation (missing coverage fix)
  - Extract all script examples from documentation (Bash, PowerShell, Python)
  - Create test environment: Bash (Git Bash/WSL on Windows, native on Unix), PowerShell (pwsh), Python 3.8+
  - Execute each example without modification
  - Log failures with file, line number, and error message
  - Target: 100% script example success rate (SC-008)
  - Mark SC-008 validation complete in tracking.json

- [ ] **T039** Update validation tracking file
  - Mark all 7 documentation files as validation_status = PASS
  - Record issue counts (total, by severity, by category)
  - Record actual review times per file
  - Generate final metrics report

- [ ] **T040** Create commit with documentation updates
  - Commit message: "docs: Review and update for production release"
  - Include all updated documentation files
  - Reference spec 008-review-and-update
  - Include validation report as commit description

- [ ] **T040a** Add version compatibility notice to documentation (edge case E2 fix)
  - Add header to README.md: "**Version**: v0.0.1 | These docs match release v0.0.1. For older versions, see git tags."
  - Add header to each docs/*.md file: "**Documentation Version**: v0.0.1 | Last Updated: [date]"
  - Addresses edge case: users on older Pass-CLI versions reading latest documentation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - User Story 1 (README): Independent - can start after Foundational
  - User Story 2 (SECURITY): Independent - can start after Foundational
  - User Story 3 (USAGE): Independent - can start after Foundational
  - User Story 4 (INSTALLATION, MIGRATION): Independent - can start after Foundational
  - User Story 5 (TROUBLESHOOTING, KNOWN_LIMITATIONS): Independent - can start after Foundational
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1 - README)**: Can start after Foundational (T004-T008) - No dependencies on other stories
- **User Story 2 (P1 - SECURITY)**: Can start after Foundational - No dependencies on other stories
- **User Story 3 (P2 - USAGE)**: Can start after Foundational - No dependencies on other stories
- **User Story 4 (P2 - INSTALLATION, MIGRATION)**: Can start after Foundational - References SECURITY (T018 ↔ T030 cross-ref)
- **User Story 5 (P3 - TROUBLESHOOTING, KNOWN_LIMITATIONS)**: Can start after Foundational - No dependencies on other stories

### Within Each User Story

- Documentation file review tasks are sequential (same file edits)
- Validation tasks run after review tasks complete
- Cross-reference verification after both files updated (T031 depends on T020 for SECURITY ↔ MIGRATION)

### Parallel Opportunities

- **Setup Phase**: All T001-T003 marked [P] can run in parallel
- **Foundational Phase**: T005, T006, T007 marked [P] can run in parallel (T004 sequential - creates base script)
- **User Stories**: Once Foundational complete (T008 done), ALL user stories (US1-US5) can be worked on in parallel by different reviewers
- **Polish Phase**: T036, T037, T038 marked [P] can run in parallel

---

## Parallel Example: After Foundational Phase

```bash
# Once T008 completes, launch all user story reviews in parallel:

# Reviewer 1:
Task: "Review README.md - Version accuracy" (T009)
Task: "Review README.md - Quick Start" (T010)
Task: "Review README.md - TUI shortcuts" (T011)
...

# Reviewer 2:
Task: "Review SECURITY.md - Crypto specs" (T015)
Task: "Review SECURITY.md - Password policy" (T016)
Task: "Review SECURITY.md - Audit logging" (T017)
...

# Reviewer 3:
Task: "Review USAGE.md - TUI vs CLI" (T021)
Task: "Review USAGE.md - Keyboard shortcuts" (T022)
Task: "Review USAGE.md - CLI commands" (T023)
...

# Reviewer 4:
Task: "Review INSTALLATION.md - Homebrew" (T026)
Task: "Review INSTALLATION.md - Scoop" (T027)
Task: "Review MIGRATION.md - Iteration upgrade" (T030)
...

# Reviewer 5:
Task: "Review TROUBLESHOOTING.md - TUI issues" (T032)
Task: "Review KNOWN_LIMITATIONS.md - Accuracy" (T034)
...
```

---

## Implementation Strategy

### MVP First (User Story 1 + 2 Only - Both P1)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T008) - CRITICAL, blocks all stories
3. Complete Phase 3: User Story 1 - README.md (T009-T014)
4. Complete Phase 4: User Story 2 - SECURITY.md (T015-T020)
5. **STOP and VALIDATE**: Test README Quick Start (≤10 min), verify SECURITY.md completeness
6. These two files are highest priority (both P1) - ready for production if validated

### Incremental Delivery

1. Complete Setup + Foundational → Validation infrastructure ready
2. Add User Story 1 (README) → Test independently → Commit
3. Add User Story 2 (SECURITY) → Test independently → Commit (MVP! - Both P1 complete)
4. Add User Story 3 (USAGE) → Test independently → Commit (P2)
5. Add User Story 4 (INSTALLATION, MIGRATION) → Test independently → Commit (P2)
6. Add User Story 5 (TROUBLESHOOTING, KNOWN_LIMITATIONS) → Test independently → Commit (P3)
7. Polish & Cross-Cutting → Final validation → Release

### Parallel Team Strategy

With multiple reviewers:

1. Team completes Setup + Foundational together (T001-T008)
2. Once T008 completes:
   - Reviewer A: User Story 1 (README) - T009-T014
   - Reviewer B: User Story 2 (SECURITY) - T015-T020
   - Reviewer C: User Story 3 (USAGE) - T021-T025
   - Reviewer D: User Story 4 (INSTALLATION, MIGRATION) - T026-T031
   - Reviewer E: User Story 5 (TROUBLESHOOTING, KNOWN_LIMITATIONS) - T032-T035
3. Stories complete independently, merge sequentially or in parallel
4. Team reconvenes for Polish phase (T036-T040)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label (US1-US5) maps task to specific user story for traceability
- Each user story validates different documentation files independently
- Foundational validation scripts (T004-T008) are CRITICAL - block all file reviews
- Commit after each user story phase completes (5 commits + 1 polish commit = 6 total)
- Stop at any checkpoint to validate story independently
- Avoid: same-file conflicts (sequential tasks within story), cross-story dependencies that break independence

---

## Task Count Summary

- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 5 tasks (CRITICAL - blocks all user stories)
- **Phase 3 (US1 - README)**: 6 tasks
- **Phase 4 (US2 - SECURITY)**: 6 tasks
- **Phase 5 (US3 - USAGE)**: 5 tasks
- **Phase 6 (US4 - INSTALLATION, MIGRATION)**: 6 tasks
- **Phase 7 (US5 - TROUBLESHOOTING, KNOWN_LIMITATIONS)**: 4 tasks
- **Phase 8 (Polish)**: 5 tasks

**Total**: 40 tasks across 8 phases (5 user story phases + 3 infrastructure phases)

**Parallel Opportunities**:
- Setup phase: 3 tasks can run in parallel
- Foundational phase: 3 tasks can run in parallel (T005-T007)
- User story phases: All 5 user stories can be worked on in parallel after Foundational completes
- Polish phase: 3 tasks can run in parallel (T036-T038)

**Independent Test Criteria**:
- US1: README Quick Start walkthrough ≤10 minutes
- US2: Security professional confirms crypto params complete
- US3: Script author successfully writes automation avoiding TUI launches
- US4: Fresh install succeeds on 3 platforms
- US5: Common issues findable in ≤5 minutes

**Suggested MVP Scope**: User Story 1 + User Story 2 (README + SECURITY - both P1, highest priority for production release)
