# Tasks: Comprehensive Documentation Restructuring

**Input**: Design documents from `/specs/002-comprehensive-documentation-restructuring/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: Validation tasks replace traditional tests for this documentation project

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

## Path Conventions
- Documentation files in `docs/`
- Hugo configuration in `docsite/`
- All file operations use `git mv` to preserve history

### File Split Methodology
All file splits preserve git history using this workflow:
1. `git mv source.md target.md`
2. Edit `target.md` to keep only migrated content (remove unwanted sections)
3. If `source.md` still needed: Edit original path to remove migrated content
4. If `source.md` obsolete: Leave deleted (git mv already handled)

**Never** create new files from scratch - always start with `git mv` to preserve history.

### Consolidation Methodology
When merging content from multiple sources into single canonical doc:
1. Choose the largest/most complete source as base (specified per task)
2. `git mv base-source.md canonical-target.md`
3. Edit `canonical-target.md` to merge content from other sources
4. Remove consolidated content from other source files

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create new section structure and prepare workspace

- [x] T001 Create new section directories: `docs/02-guides/`, `docs/04-troubleshooting/`, `docs/05-operations/`
- [x] T002 [P] Backup current docs structure for reference (snapshot current state)
- [x] T003 [P] Document current file-to-line mappings for audit trail

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Section restructuring that MUST be complete before ANY user story documentation can be reorganized

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Renumber section 02-usage → 02-usage-temp (prepare for reorganization)
- [x] T005 Renumber section 03-guides → 03-guides-temp (prepare for reorganization)
- [x] T006 Renumber section 04-reference → 03-reference (final position)
- [x] T007 Renumber section 05-development → 06-development (final position)
- [x] T008 Move backup-restore.md from 03-guides-temp to 02-guides/ using `git mv docs/03-guides-temp/backup-restore.md docs/02-guides/backup-restore.md`
- [x] T009 Update backup-restore.md front matter (section weight, title verification)

**Checkpoint**: Section structure ready - user story documentation work can now begin

---

## Phase 3: User Story 1 - New User Quick Start (Priority: P1) 🎯 MVP

**Goal**: Enable new users to install and add first credential within 5 minutes via quick-install and quick-start guides

**Independent Test**: Follow quick-install.md + quick-start.md from fresh system → successfully init vault + add + get credential

### Implementation for User Story 1

- [x] T010 [P] [US1] Split installation.md → quick-install.md: Extract package manager instructions (lines 1-150) to `docs/01-getting-started/quick-install.md`
- [x] T011 [P] [US1] Split installation.md → manual-install.md: Extract binary/source installation (lines 151-550, 600-708) to `docs/01-getting-started/manual-install.md`
- [x] T012 [P] [US1] Split installation.md → uninstall.md: Extract uninstall instructions (lines 550-600) to `docs/01-getting-started/uninstall.md`
- [x] T013 [P] [US1] Split first-steps.md → quick-start.md: Extract quick-start guide (lines 1-200) to `docs/01-getting-started/quick-start.md`
- [x] T014 [US1] Add front matter to quick-install.md (title: "Quick Install", weight: 1, bookToc: true)
- [x] T015 [US1] Add front matter to manual-install.md (title: "Manual Installation", weight: 2, bookToc: true)
- [x] T016 [US1] Add front matter to quick-start.md (title: "Quick Start Guide", weight: 3, bookToc: true)
- [x] T017 [US1] Add front matter to uninstall.md (title: "Uninstall", weight: 4, bookToc: true)
- [x] T018 [US1] Remove original installation.md and first-steps.md sections related to US1 content
- [x] T019 [US1] Update 01-getting-started/_index.md with new file descriptions and links
- [x] T020 [US1] Convert internal links in getting-started docs to Hugo relref format

**Checkpoint**: New users can now complete installation and quick-start independently (MVP complete!)

---

## Phase 4: User Story 2 - Daily Command Reference Lookup (Priority: P1)

**Goal**: Enable daily users to quickly find CLI command syntax without scrolling through unrelated content

**Independent Test**: Time how long it takes to find `pass-cli get` syntax in command-reference.md vs old cli-reference.md (should be <30 seconds in new structure)

### Implementation for User Story 2

- [x] T021 [P] [US2] Split cli-reference.md → command-reference.md: Extract command syntax only (lines 1-800, 1900-2040) to `docs/03-reference/command-reference.md`
- [x] T022 [P] [US2] Split cli-reference.md → tui-guide.md: Extract TUI documentation (lines 1206-1580) to `docs/02-guides/tui-guide.md`
- [x] T023 [P] [US2] Split cli-reference.md → scripting-guide.md: Extract automation/quiet/JSON examples (lines 1722-1800 + scattered) to `docs/02-guides/scripting-guide.md`
- [x] T024 [P] [US2] Consolidate configuration.md: Use `git mv docs/02-usage-temp/cli-reference.md docs/03-reference/configuration.md` as base, edit to keep config section (lines 1089-1205), then merge config content from first-steps.md (lines 300-400)
- [x] T025 [P] [US2] Split cli-reference.md → usage-tracking.md: Extract usage tracking section (lines 1581-1721) to `docs/02-guides/usage-tracking.md`
- [x] T026 [US2] Add front matter to command-reference.md (title: "Command Reference", weight: 1)
- [x] T027 [US2] Add front matter to tui-guide.md (title: "TUI Guide", weight: 5)
- [x] T028 [US2] Add front matter to scripting-guide.md (title: "Scripting Guide", weight: 6)
- [x] T029 [US2] Add front matter to configuration.md (title: "Configuration", weight: 2)
- [ ] T030 [US2] Add front matter to usage-tracking.md (title: "Usage Tracking", weight: 3)
- [ ] T031 [US2] Create/update 02-guides/_index.md with guide descriptions
- [ ] T032 [US2] Create/update 03-reference/_index.md with reference descriptions
- [ ] T033 [US2] Remove original cli-reference.md after verifying all content migrated
- [ ] T034 [US2] Convert internal links in guides and reference docs to Hugo relref format

**Checkpoint**: Daily users can now quickly lookup command syntax independently

---

## Phase 5: User Story 3 - Troubleshooting Error Resolution (Priority: P2)

**Goal**: Enable users to find error solutions without reading 1,404 lines of mixed troubleshooting content

**Independent Test**: Given vault corruption error → navigate to troubleshooting-vault.md → find recovery steps in <2 minutes

### Implementation for User Story 3

- [ ] T035 [P] [US3] Create 04-troubleshooting/ section directory
- [ ] T036 [P] [US3] Split troubleshooting.md → installation.md: Extract install/init issues (lines 1-300) + installation.md troubleshooting to `docs/04-troubleshooting/installation.md`
- [ ] T037 [P] [US3] Split troubleshooting.md → vault.md: Extract vault access/corruption/recovery (lines 743-1240) to `docs/04-troubleshooting/vault.md`
- [ ] T038 [P] [US3] Split troubleshooting.md → keychain.md: Extract keychain platform issues (lines 485-742) to `docs/04-troubleshooting/keychain.md`
- [ ] T039 [P] [US3] Split troubleshooting.md → tui.md: Extract TUI rendering/interaction issues (TUI sections) to `docs/04-troubleshooting/tui.md`
- [ ] T040 [P] [US3] Consolidate FAQ: Use `git mv docs/04-reference/troubleshooting.md docs/04-troubleshooting/faq.md` as base, edit to keep only FAQ section (lines 1242-1368), then merge FAQ content from first-steps.md (lines 443-492, 550+), cli-reference.md (lines 1802-2032), and migration.md (lines 388-434)
- [ ] T041 [US3] Add front matter to all troubleshooting docs (titles, weights 1-5)
- [ ] T042 [US3] Create 04-troubleshooting/_index.md with category descriptions
- [ ] T043 [US3] Remove FAQ sections from original source files (first-steps.md, cli-reference.md remnants, migration.md)
- [ ] T044 [US3] Remove original troubleshooting.md after verifying all content migrated
- [ ] T045 [US3] Convert internal links in troubleshooting docs to Hugo relref format

**Checkpoint**: Users experiencing problems can now find category-specific solutions independently

---

## Phase 6: User Story 4 - Advanced Feature Discovery (Priority: P2)

**Goal**: Enable power users to explore advanced features without information scattered across multiple documents

**Independent Test**: Power user wants keychain integration → finds keychain-setup.md (consolidated guide) → completes setup using single doc

### Implementation for User Story 4

- [ ] T046 [P] [US4] Split first-steps.md → basic-workflows.md: Extract list/update/delete/generate workflows (lines 201-450) to `docs/02-guides/basic-workflows.md`
- [ ] T047 [P] [US4] Consolidate keychain-setup.md: Use `git mv docs/01-getting-started/first-steps.md docs/02-guides/keychain-setup.md` as base, edit to keep only keychain section (lines 200-250), then merge keychain content from cli-reference.md (lines 800-900) and troubleshooting.md (lines 485-550)
- [ ] T048 [US4] Add front matter to basic-workflows.md (title: "Basic Workflows", weight: 1)
- [ ] T049 [US4] Add front matter to keychain-setup.md (title: "Keychain Setup", weight: 2)
- [ ] T050 [US4] Remove keychain content from original source files (first-steps.md remnants, portions from cli-reference/troubleshooting splits)
- [ ] T051 [US4] Remove first-steps.md after verifying all content migrated to quick-start, basic-workflows, keychain-setup
- [ ] T052 [US4] Update 02-guides/_index.md with advanced feature descriptions
- [ ] T053 [US4] Convert internal links in advanced guides to Hugo relref format

**Checkpoint**: Power users can now discover and use advanced features from dedicated guides independently

---

## Phase 7: User Story 5 - Security Audit and Operations (Priority: P3)

**Goal**: Enable security engineers to review architecture and ops procedures separately

**Independent Test**: Security auditor reviews security-architecture.md for crypto details → ops engineer reviews security-operations.md for incident response → both complete tasks without overlap

### Implementation for User Story 5

- [ ] T054 [P] [US5] Create 05-operations/ section directory
- [ ] T055 [P] [US5] Split security.md → security-architecture.md: Extract crypto/threat model (lines 1-500) to `docs/03-reference/security-architecture.md`
- [ ] T056 [P] [US5] Split security.md → security-operations.md: Extract best practices/incident response (lines 501-750) to `docs/05-operations/security-operations.md`
- [ ] T057 [P] [US5] Move doctor-command.md: Relocate from 06-development/ to 05-operations/health-checks.md using `git mv docs/06-development/doctor-command.md docs/05-operations/health-checks.md`
- [ ] T058 [US5] Add front matter to security-architecture.md (title: "Security Architecture", weight: 3)
- [ ] T059 [US5] Add front matter to security-operations.md (title: "Security Operations", weight: 2)
- [ ] T060 [US5] Add front matter to health-checks.md (title: "Health Checks", weight: 1)
- [ ] T061 [US5] Create 05-operations/_index.md with operations section description
- [ ] T062 [US5] Update 03-reference/_index.md to include security-architecture.md
- [ ] T063 [US5] Update 06-development/_index.md (remove doctor-command reference)
- [ ] T064 [US5] Remove original security.md after verifying all content migrated
- [ ] T065 [US5] Convert internal links in security and operations docs to Hugo relref format

**Checkpoint**: Security and operations teams can now review documentation independently

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Finalize documentation structure, update all cross-references, validate

- [ ] T066 [P] Update docs/README.md quick links to point to new file paths (all 6 new sections)
- [ ] T067 [P] Copy docs/README.md to docs/_index.md (Hugo homepage)
- [ ] T068 [P] Update root README.md documentation links to new structure
- [ ] T069 [P] Update all section _index.md files with accurate weights (01-06)
- [ ] T070 [P] Verify all moved files have correct front matter (title, weight, bookToc)
- [ ] T071 Move known-limitations.md from old 04-reference to 03-reference using `git mv`
- [ ] T072 Move migration.md from old 04-reference to 03-reference using `git mv`
- [ ] T073 Update known-limitations.md and migration.md front matter (weights 4-5)
- [ ] T074 Search all docs for remaining markdown links `[text](*.md)` and convert to relref format
- [ ] T075 Verify no duplicate content remains (grep for sample paragraphs from keychain, FAQ, config)
- [ ] T076 Clean up temporary section directories (02-usage-temp, 03-guides-temp)

---

## Phase 9: Validation & Testing

**Purpose**: Verify all success criteria from spec.md are met

- [ ] T077 **Validation**: Run line count check - `wc -l docs/**/*.md | awk '{sum+=$1; count++} END {print sum/count}'` → must be ≤300 average
- [ ] T078 **Validation**: Find longest doc - `find docs -name '*.md' -exec wc -l {} + | sort -rn | head -1` → must be <700 lines
- [ ] T079 **Validation**: Count total files - `find docs -name '*.md' -not -name '_index.md' -not -name 'README.md' | wc -l` → must equal 29
- [ ] T080 **Validation**: Hugo build test - `cd docsite && hugo --buildDrafts` → must complete without errors
- [ ] T081 **Validation**: Link validation - Hugo build output must show zero broken relref links
- [ ] T082 **Validation**: Git history check - `git log --follow docs/02-guides/keychain-setup.md` → must show history from original sources
- [ ] T083 **Validation**: Git history check - `git log --follow docs/04-troubleshooting/faq.md` → must show consolidated history
- [ ] T084 **Validation**: Render test - Visit http://localhost:1313/pass-cli/ → all 29 docs must load without 404s
- [ ] T085 **Validation**: Navigation test - Homepage → any common task (install, add credential, troubleshoot) in ≤3 clicks
- [ ] T086 **Final Review**: Compare word counts - `wc -w docs/**/*.md` → should show ~27,000 words (down from ~33,000)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - User Story 1 (P1) - New User Quick Start
  - User Story 2 (P1) - Daily Command Reference
  - User Story 3 (P2) - Troubleshooting
  - User Story 4 (P2) - Advanced Features
  - User Story 5 (P3) - Security/Operations
- **Polish (Phase 8)**: Depends on all user stories being complete
- **Validation (Phase 9)**: Depends on Polish completion

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 3 (P2)**: Can start after Foundational - Minimal dependency on US2 (references command-reference.md)
- **User Story 4 (P2)**: Can start after Foundational - Uses content split in US2 (guides section)
- **User Story 5 (P3)**: Can start after Foundational - No dependencies on other stories

### Within Each User Story

- File splits marked [P] can run in parallel (different source files)
- Front matter updates happen sequentially after splits
- _index.md updates happen after front matter
- Link conversion happens after all files created
- Original file removal happens last (after content verified migrated)

### Parallel Opportunities

- **Setup (Phase 1)**: All 3 tasks can run in parallel
- **Foundational (Phase 2)**: Sequential (section renumbering must happen in order)
- **User Story 1 (P1)**: T010-T013 can run in parallel (4 different splits)
- **User Story 2 (P1)**: T021-T025 can run in parallel (5 different splits)
- **User Story 3 (P2)**: T036-T040 can run in parallel (5 different splits/consolidations)
- **User Story 4 (P2)**: T046-T047 can run in parallel (2 different operations)
- **User Story 5 (P3)**: T054-T057 can run in parallel (4 different operations)
- **Polish (Phase 8)**: T066-T070 can run in parallel (different files)
- **Validation (Phase 9)**: All validation tasks can run in parallel

---

## Parallel Example: User Story 2

```bash
# Launch all file splits for User Story 2 together:
# - Split cli-reference → command-reference.md
# - Split cli-reference → tui-guide.md
# - Split cli-reference → scripting-guide.md
# - Consolidate config → configuration.md
# - Split cli-reference → usage-tracking.md

# All operate on different target files, no conflicts
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only - Both P1)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (New User Quick Start)
4. Complete Phase 4: User Story 2 (Daily Command Reference)
5. **STOP and VALIDATE**: Test both P1 stories independently
6. Deploy/demo if ready

This gives new users quick-start capability AND daily users fast command lookup - the two most critical user journeys.

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Covers new user journey
3. Add User Story 2 → Test independently → Covers daily user journey (MVP!)
4. Add User Story 3 → Test independently → Covers troubleshooting journey
5. Add User Story 4 → Test independently → Covers power user journey
6. Add User Story 5 → Test independently → Covers security/ops journey
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple people:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Person A: User Story 1 (Getting Started docs)
   - Person B: User Story 2 (Command Reference + Guides splits)
   - Person C: User Story 3 (Troubleshooting splits)
3. Stories complete and integrate independently

---

## Notes

- **git mv REQUIRED**: Always use `git mv` for file operations (preserves history)
- **Hugo relref REQUIRED**: Convert all `[text](path.md)` to `{{< relref "path/file" >}}`
- **[P] tasks**: Different files, no dependencies, can run simultaneously
- **[Story] labels**: Map tasks to user stories for traceability
- **Target line counts**: 300 average, 700 max (enforced in validation)
- **Independent testing**: Each user story should be verifiable without others
- **Commit frequently**: After each task or logical group (per CLAUDE.md)
- **Checkpoints**: Stop after each user story to validate independently
- **Validation last**: All T077-T086 validation tasks run at the end

---

## Summary

**Total Tasks**: 86 tasks across 9 phases
**User Stories**: 5 stories (2 x P1, 2 x P2, 1 x P3)
**Parallel Opportunities**: 35+ tasks marked [P] for concurrent execution
**MVP Scope**: User Stories 1 & 2 (Phases 1-4) = New users + Daily users covered
**Critical Path**: Setup (1-3) → Foundational (4-9) → User Story 2 (21-34) is longest

**Task Breakdown by User Story**:
- User Story 1 (P1 - New Users): 11 tasks (T010-T020)
- User Story 2 (P1 - Daily Users): 14 tasks (T021-T034)
- User Story 3 (P2 - Troubleshooting): 11 tasks (T035-T045)
- User Story 4 (P2 - Power Users): 8 tasks (T046-T053)
- User Story 5 (P3 - Security/Ops): 12 tasks (T054-T065)
- Polish & Validation: 21 tasks (T066-T086)
