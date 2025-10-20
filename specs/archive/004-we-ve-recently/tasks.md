---
description: "Task list for Documentation Update feature implementation"
---

# Tasks: Documentation Update for Recent Application Changes

**Input**: Design documents from `/specs/004-we-ve-recently/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md, contracts/

**Tests**: No automated tests for documentation accuracy. Manual verification required (cross-reference with actual feature implementation).

**Organization**: Tasks organized by user story. This feature has ONE user story (US1: Update Interactive TUI Features Documentation).

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1)
- Include exact file paths in descriptions

## Path Conventions
- Documentation files at repository root: `README.md`, `docs/USAGE.md`
- Source code references for verification: `cmd/tui/`

---

## Phase 1: Setup (Preparation)

**Purpose**: Prepare for documentation updates by verifying current state

- [X] T001 Read current `README.md` to understand existing structure and content
- [X] T002 Read current `docs/USAGE.md` to understand existing structure and content
- [X] T003 [P] Verify all keyboard shortcuts in `cmd/tui/events/handlers.go` match research.md findings
- [X] T004 [P] Verify password visibility toggle implementation in `cmd/tui/components/forms.go:241,585`

**Checkpoint**: Current documentation state understood, all features verified in code

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core documentation structure that MUST be complete before detailed content updates

**⚠️ CRITICAL**: These structural changes must complete before detailed documentation content can be added

- [X] T005 Identify insertion point for TUI Mode section in `README.md` (after Quick Start, before Usage section)
- [X] T006 Identify insertion point for TUI section in `docs/USAGE.md` (after Configuration ~line 740, before Best Practices)
- [X] T007 Audit all file path references in `README.md` and `docs/USAGE.md` for accuracy
- [X] T008 Verify vault file path references (`~/.pass-cli/vault.enc` vs `vault.json`)

**Checkpoint**: Documentation insertion points identified, existing file paths audited

---

## Phase 3: User Story 1 - Update Interactive TUI Features Documentation (Priority: P1) 🎯 MVP

**Goal**: Document all interactive TUI features (password visibility toggle, keyboard shortcuts, form interactions) added in specs 001-003 so users can discover and use new functionality

**Independent Test**: Verify each documented feature exists in the application and behaves as documented by manually testing keyboard shortcuts, TUI launch, and interactive features

### Implementation for User Story 1

- [X] T009 [US1] Add TUI Mode section to `README.md` with launch instructions and features overview
- [X] T010 [US1] Add keyboard shortcuts reference table to `README.md` (6 essential shortcuts from quickstart.md)
- [X] T011 [US1] Add "TUI Mode" main section to `docs/USAGE.md` (insert content from contracts/tui-usage-section.md)
- [X] T012 [US1] Add "Launching TUI Mode" subsection to `docs/USAGE.md` with examples
- [X] T013 [US1] Add "TUI vs CLI Mode" subsection to `docs/USAGE.md` with comparison table
- [X] T014 [US1] Add "TUI Keyboard Shortcuts" subsection to `docs/USAGE.md` with complete organized tables
- [X] T015 [US1] Add "Search & Filter" subsection to `docs/USAGE.md` documenting `/` key activation and behavior
- [X] T016 [US1] Add "Password Visibility Toggle" subsection to `docs/USAGE.md` documenting Ctrl+H in forms
- [X] T017 [US1] Add "Layout Controls" subsection to `docs/USAGE.md` documenting `i` (detail panel) and `s` (sidebar) toggles
- [X] T018 [US1] Add "Usage Location Display" subsection to `docs/USAGE.md` documenting detail panel feature
- [X] T019 [US1] Add "Exiting TUI Mode" subsection to `docs/USAGE.md` documenting `q` and Ctrl+C
- [X] T020 [US1] Add "TUI Best Practices" section to `docs/USAGE.md` with 6 user tips
- [X] T021 [US1] Add "TUI Troubleshooting" section to `docs/USAGE.md` with common issues and solutions
- [X] T022 [US1] Update any outdated installation/build instructions in `README.md` if found during audit
- [X] T023 [US1] Ensure cross-references between `README.md` and `docs/USAGE.md` are consistent

**Checkpoint FR-007**: Verify consistency requirement satisfied - all cross-file references synchronized (TUI feature names, keyboard shortcuts, file paths)

**Checkpoint**: All TUI features from specs 001-003 are now documented with usage examples

---

## Phase 4: Verification & Polish

**Purpose**: Manual verification that documentation matches actual implementation and is user-ready

- [X] T024 [P] [US1] Manual test: Launch TUI with `pass-cli` (no args) and verify it works as documented
- [X] T025 [P] [US1] Manual test: Press `n` to open add form, type password, press Ctrl+H to toggle visibility
- [X] T026 [P] [US1] Manual test: Press Ctrl+H again in add form to re-mask password
- [X] T027 [P] [US1] Manual test: Press `e` to open edit form, press Ctrl+H to verify toggle works in edit forms
- [X] T028 [P] [US1] Manual test: Press `/` to activate search, type "git", verify results filter in real-time
- [X] T029 [P] [US1] Manual test: Press Esc to exit search mode, verify all credentials shown again
- [X] T030 [P] [US1] Manual test: Press `s` three times to cycle sidebar states (Auto/Hide/Show)
- [X] T031 [P] [US1] Manual test: Press `i` three times to cycle detail panel states (Auto/Hide/Show)
- [X] T032 [P] [US1] Manual test: Press `?` to show help modal, verify all shortcuts listed
- [X] T033 [P] [US1] Manual test: Select credential with usage history, verify "Usage Locations" section appears in detail panel
- [X] T034 [P] [US1] Manual test: Press `q` to quit TUI, verify app exits to shell
- [X] T035 [P] [US1] Manual test: Press Ctrl+C to quit TUI, verify app exits to shell
- [X] T036 [US1] Verify all file paths referenced in documentation exist in current codebase (zero broken references)
- [X] T037 [US1] Verify all documented keyboard shortcuts match actual key bindings in `cmd/tui/events/handlers.go`
- [X] T038 [US1] Verify all code examples in documentation are copy-pasteable (no placeholder values)
- [X] T039 [US1] Verify README.md and docs/USAGE.md use consistent formatting and terminology
- [X] T040 [US1] Final review: Read both documentation files as a new user to confirm discoverability (<1 minute to find TUI launch instructions)
- [X] T041 [US1] Final review: Confirm 100% of interactive features from specs 001-003 are documented

**Checkpoint**: All documented features verified working, documentation complete and accurate

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS detailed documentation updates
- **User Story 1 (Phase 3)**: Depends on Foundational phase completion
- **Verification (Phase 4)**: Depends on User Story 1 implementation completion

### Within User Story 1

- T009-T010 (README.md updates) can proceed independently
- T011-T021 (docs/USAGE.md section additions) should proceed sequentially to maintain document flow
- T022-T023 (cross-references and cleanup) should happen after main content is added
- Tasks T024-T035 in Phase 4 marked [P] can all run in parallel (different manual test scenarios)
- Tasks T036-T041 in Phase 4 should run sequentially after manual tests complete

### Parallel Opportunities

- Phase 1: Tasks T003-T004 marked [P] can run in parallel (verifying different code files)
- Phase 4: Tasks T024-T035 marked [P] can run in parallel (independent manual tests)

---

## Parallel Example: Verification Phase

```bash
# Manual tests can be performed in any order (all marked [P]):
Task: "Manual test: Launch TUI with pass-cli (no args)"
Task: "Manual test: Press n to open add form, press Ctrl+H to toggle visibility"
Task: "Manual test: Press / to activate search, type 'git'"
Task: "Manual test: Press s three times to cycle sidebar states"
Task: "Manual test: Press i three times to cycle detail panel states"
Task: "Manual test: Press ? to show help modal"
Task: "Manual test: Select credential with usage history"
Task: "Manual test: Press q to quit TUI"
Task: "Manual test: Press Ctrl+C to quit TUI"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

This feature has only one user story, so MVP = complete feature

1. Complete Phase 1: Setup (understand current state)
2. Complete Phase 2: Foundational (identify structure and insertion points)
3. Complete Phase 3: User Story 1 (add all documentation content)
4. Complete Phase 4: Verification & Polish (manual testing and final review)
5. **STOP and VALIDATE**: Follow documentation as a new user to verify clarity

### Incremental Delivery

Since this is a documentation-only feature with a single user story:

1. Complete Setup + Foundational → Documentation structure ready
2. Add User Story 1 content → Documentation complete
3. Verify all features → Documentation validated
4. Commit documentation updates → Ready for users

---

## Notes

- [P] tasks = different files or independent test scenarios, no dependencies
- [US1] label maps all tasks to User Story 1 (only story in this feature)
- Manual verification required since documentation accuracy cannot be unit tested
- Each documented feature MUST be tested by following documented steps
- Commit frequently after completing logical task groups (e.g., after adding complete README section)
- No speculative/planned features should be documented (only implemented features from specs 001-003)
- Cross-references between README.md and docs/USAGE.md MUST stay synchronized
- Keyboard shortcuts MUST use format: `Key - Action (Context)`
- All examples MUST be copy-pasteable with no placeholder values
