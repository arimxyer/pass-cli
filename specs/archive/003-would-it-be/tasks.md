# Tasks: Password Visibility Toggle

**Input**: Design documents from `/specs/003-would-it-be/`
**Prerequisites**: plan.md, spec.md, research.md, quickstart.md

**Tests**: INCLUDED - Constitution Principle IV (Test-Driven Development) is NON-NEGOTIABLE for this project

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- Single project structure (repository root)
- Forms: `cmd/tui/components/forms.go`
- Tests: `tests/unit/` and `tests/integration/`

---

## Phase 1: Foundational (No Setup Required)

**Purpose**: This TUI-only feature requires no project setup - existing codebase is ready

**⚠️ NOTE**: Since this is an enhancement to existing forms, there is no foundational blocking work. All user stories can proceed independently after understanding the codebase structure.

- [X] T001 Review existing AddForm and EditForm implementation in cmd/tui/components/forms.go (lines 27-613)
- [X] T002 Review existing Ctrl+S keyboard shortcut pattern in setupKeyboardShortcuts() methods
- [X] T003 Review existing password field creation at line 91 (AddForm) and lines 329-339 (EditForm)

**Checkpoint**: Codebase structure understood - user story implementation can now begin

---

## Phase 2: User Story 1 - Toggle Password Visibility When Adding Entries (Priority: P1) 🎯 MVP

**Goal**: Enable users to toggle password visibility in the add form via Ctrl+H shortcut to verify typos before saving

**Independent Test**: Open add form, type "test123", press Ctrl+H, verify password shows as plaintext and label shows "[VISIBLE]", press Ctrl+H again, verify password shows as "*******" and label shows "Password"

###  Tests for User Story 1 (TDD - Write First, Ensure FAIL)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T004 [P] [US1] Unit test: TestAddFormPasswordVisibilityToggle in tests/unit/tui_forms_test.go
  - Test initial state (masked with '*')
  - Test toggle to visible (mask character = 0)
  - Test toggle back to masked (mask character = '*')
  - Test label changes ("Password" → "Password [VISIBLE]" → "Password")
- [X] T005 [P] [US1] Unit test: TestAddFormCtrlHShortcut in tests/unit/tui_forms_test.go
  - Test Ctrl+H key event triggers toggle
  - Test event consumed (returns nil)
  - Test other keys not affected
- [X] T006 [P] [US1] Integration test: TestAddFormCursorPreservation in tests/integration/tui_password_toggle_test.go
  - Type "test", move cursor to position 2, press Ctrl+H
  - Verify cursor still at position 2 after toggle
  - Type more characters, verify correct insertion point

### Implementation for User Story 1

- [X] T007 [US1] Add `passwordVisible bool` field to AddForm struct in cmd/tui/components/forms.go (after line 33)
- [X] T008 [US1] Implement `togglePasswordVisibility()` method for AddForm in cmd/tui/components/forms.go (after applyStyles(), ~line 284)
  - Toggle passwordVisible flag
  - Get password field via GetFormItem(2)
  - SetMaskCharacter(0) when visible, SetMaskCharacter('*') when masked
  - Update label to "Password [VISIBLE]" or "Password"
- [X] T009 [US1] Add Ctrl+H case to AddForm.setupKeyboardShortcuts() in cmd/tui/components/forms.go (line ~235)
  - Add `case tcell.KeyCtrlH: af.togglePasswordVisibility(); return nil`
- [X] T010 [US1] Update AddForm.addKeyboardHints() text in cmd/tui/components/forms.go (line ~217)
  - Change to: "Tab: Next field • Shift+Tab: Previous • Ctrl+S: Add • Ctrl+H: Toggle password • Esc: Cancel"
- [X] T011 [US1] Run tests: `go test ./tests/unit/tui_forms_test.go -run TestAddForm -v`
- [X] T012 [US1] Run integration tests: `go test ./tests/integration/tui_password_toggle_test.go -run TestAddForm -v`
- [X] T013 [US1] Manual testing: Build TUI, open add form, verify Ctrl+H toggles visibility correctly
  - Verify label text changes match existing pattern (compare with detail panel 'p' toggle at cmd/tui/events/handlers.go:91)

**Checkpoint**: User Story 1 complete - add form has working password visibility toggle

---

## Phase 3: User Story 2 - Toggle Password Visibility When Editing Entries (Priority: P2)

**Goal**: Enable users to toggle password visibility in the edit form via Ctrl+H shortcut to verify changes before saving

**Independent Test**: Select existing credential, press 'e', focus password field, press Ctrl+H, verify password becomes visible with "[VISIBLE]" label, press Ctrl+H, verify password returns to masked

### Tests for User Story 2 (TDD - Write First, Ensure FAIL)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T014 [P] [US2] Unit test: TestEditFormPasswordVisibilityToggle in tests/unit/tui_forms_test.go
  - Test initial state (masked with '*')
  - Test toggle to visible (mask character = 0)
  - Test toggle back to masked (mask character = '*')
  - Test label changes ("Password" → "Password [VISIBLE]" → "Password")
- [X] T015 [P] [US2] Unit test: TestEditFormCtrlHShortcut in tests/unit/tui_forms_test.go
  - Test Ctrl+H key event triggers toggle
  - Test event consumed (returns nil)
- [X] T016 [P] [US2] Integration test: TestEditFormCursorPreservation in tests/integration/tui_password_toggle_test.go
  - Load credential, edit password to "newpass", move cursor, press Ctrl+H
  - Verify cursor position preserved after toggle

### Implementation for User Story 2

- [X] T017 [US2] Add `passwordVisible bool` field to EditForm struct in cmd/tui/components/forms.go (after line 48)
- [X] T018 [US2] Implement `togglePasswordVisibility()` method for EditForm in cmd/tui/components/forms.go (after applyStyles(), ~line 602)
  - Identical logic to AddForm toggle method
  - Toggle passwordVisible flag
  - Get password field via GetFormItem(2)
  - SetMaskCharacter(0) when visible, SetMaskCharacter('*') when masked
  - Update label to "Password [VISIBLE]" or "Password"
- [X] T019 [US2] Add Ctrl+H case to EditForm.setupKeyboardShortcuts() in cmd/tui/components/forms.go (line ~551)
  - Add `case tcell.KeyCtrlH: ef.togglePasswordVisibility(); return nil`
- [X] T020 [US2] Update EditForm.addKeyboardHints() text in cmd/tui/components/forms.go (line ~535)
  - Change to: "Tab: Next field • Shift+Tab: Previous • Ctrl+S: Save • Ctrl+H: Toggle password • Esc: Cancel"
- [X] T021 [US2] Run tests: `go test ./tests/unit/tui_forms_test.go -run TestEditForm -v`
- [X] T022 [US2] Run integration tests: `go test ./tests/integration/tui_password_toggle_test.go -run TestEditForm -v`
- [X] T023 [US2] Manual testing: Build TUI, edit credential, verify Ctrl+H toggles visibility correctly
  - Verify label text changes match existing pattern (compare with detail panel 'p' toggle)

**Checkpoint**: User Story 2 complete - edit form has working password visibility toggle

---

## Phase 4: User Story 3 - Persistent Visibility State Awareness (Priority: P3)

**Goal**: Provide clear visual feedback about password visibility state and ensure security by resetting to masked when navigating away

**Independent Test**: Open add form, toggle visibility, verify label shows "[VISIBLE]", press Esc, reopen form, verify password field defaults to masked

### Tests for User Story 3 (TDD - Write First, Ensure FAIL)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T024 [P] [US3] Unit test: TestPasswordDefaultsMasked in tests/unit/tui_forms_test.go (validates FR-009)
  - Test AddForm initializes with passwordVisible = false
  - Test EditForm initializes with passwordVisible = false
  - Test password field mask character = '*' on form creation
  - Validates FR-009: "System MUST default password fields to hidden state when forms are first opened"
- [X] T025 [P] [US3] Integration test: TestVisibilityResetOnFormClose in tests/integration/tui_password_toggle_test.go (validates FR-010)
  - **Add form cancel path**: Open add form, toggle visible, press Esc → reopen, verify passwordVisible = false
  - **Add form submit path**: Open add form, toggle visible, press Ctrl+S → reopen, verify passwordVisible = false
  - **Edit form cancel path**: Open edit form, toggle visible, press Esc → reopen, verify passwordVisible = false
  - **Edit form save path**: Open edit form, toggle visible, press Ctrl+S → reopen, verify passwordVisible = false
  - **Form switch path**: Open add form, toggle visible → switch to edit form, verify passwordVisible = false (edit form fresh state)
  - **Main menu navigation**: Open add form, toggle visible → press Esc to main → press 'a' again, verify passwordVisible = false
  - Validates FR-010: "System MUST reset password visibility to hidden when navigating away from forms" (all 6 navigation paths tested)
- [X] T026 [P] [US3] Integration test: TestVisualIndicatorChanges in tests/integration/tui_password_toggle_test.go
  - Verify label text changes reflect visibility state accurately
  - Test both forms show correct indicator on toggle

### Implementation for User Story 3

**NOTE**: Most functionality already implemented in US1 and US2. This phase focuses on verification and edge cases.

- [X] T027 [US3] Verify AddForm.buildFormFields() initializes password field with '*' mask (already done at line 91)
- [X] T028 [US3] Verify EditForm.buildFormFieldsWithValues() initializes with '*' mask (already done at line 332)
- [X] T029 [US3] Add edge case test: TestEmptyPasswordFieldToggle in tests/unit/tui_forms_test.go
  - Verify toggle works on empty password field (no crash, label still updates)
- [X] T030 [US3] Add unicode/emoji handling limitation to quickstart.md Common Issues section
  - **Target**: Add new issue to quickstart.md line ~295 (Common Issues section)
  - **Content**: "Issue: Unicode/emoji passwords display inconsistently across terminals. Solution: Terminal rendering of wide characters (CJK, emoji) varies - tview masks each rune as single '*', but visible display depends on terminal's Unicode support. This is expected behavior and outside our control."
  - **Note**: Technical details already documented in research.md:18-19. This task adds user-facing troubleshooting guidance.
  - Mark as complete after verifying quickstart.md contains this Common Issue entry
- [ ] T030a [US3] **[DEFERRED]** Add edge case test: TestCopyPasteWithVisiblePassword in tests/integration/tui_password_toggle_test.go
  - Type password "SecurePass123", toggle visible
  - Select all text (Ctrl+A if supported), copy (Ctrl+C)
  - Clear field, paste (Ctrl+V)
  - Verify pasted content matches original
  - Toggle to masked, verify mask applied correctly to pasted content
  - **DEFERRED REASON**: tview InputField does not provide Ctrl+C/Ctrl+V keyboard shortcuts (requires custom clipboard implementation beyond MVP scope per Principle VII - Simplicity)
  - **UN-DEFERRAL CRITERIA**: Re-enable when (1) tview adds native clipboard support in future release, OR (2) custom clipboard implementation approved by project maintainer with complexity justification in plan.md
- [X] T031 [US3] Add security test: TestNoPasswordLogging in tests/integration/tui_password_toggle_test.go
  - Run TUI with `--verbose` flag: `./pass-cli --verbose tui`
  - Toggle visibility multiple times
  - Verify no password content appears in stdout/stderr (only state changes logged)
- [X] T032 [US3] Run all tests: `go test ./tests/... -v`
- [X] T033 [US3] Manual testing: Verify all edge cases and security requirements from spec.md
  - Test unicode/emoji passwords (e.g., "测试🔐emoji") - toggle visible, verify characters display (terminal-dependent)
  - Verify no password logging when toggling
  - Verify form reset behavior on cancel/save/navigation

**Checkpoint**: All user stories complete - full password visibility toggle feature implemented

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and documentation

- [X] T034 [P] Run full test suite with coverage: `go test ./... -cover`
- [X] T035 [P] Verify >80% code coverage for modified files (constitution requirement)
- [X] T036 [P] Run golangci-lint on forms.go: `golangci-lint run cmd/tui/components/forms.go`
- [X] T037 [P] Run security scan: `gosec ./cmd/tui/components/...`
- [X] T038 Build final binary: `go build -o pass-cli.exe`
- [X] T039 Manual end-to-end test following quickstart.md validation checklist
- [X] T040 Update specs/003-would-it-be/tasks.md to mark all tasks complete
- [X] T041 [P] Code review: Verify constitution compliance (no secret logging, security-first, TDD followed)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: Code review only - no implementation blocking work
- **User Story 1 (Phase 2)**: Can start immediately - implements add form toggle (MVP)
- **User Story 2 (Phase 3)**: Can start immediately - implements edit form toggle (independent of US1)
- **User Story 3 (Phase 4)**: Depends on US1 and US2 - validates security and state management
- **Polish (Phase 5)**: Depends on US1, US2, US3 complete

### User Story Dependencies

- **User Story 1 (P1)**: Independent - no dependencies
- **User Story 2 (P2)**: Independent - no dependencies (parallel with US1)
- **User Story 3 (P3)**: Depends on US1 and US2 - verifies their behavior

### Within Each User Story

- Tests (T004-T006, T014-T016, T024-T026) MUST be written FIRST and FAIL
- Implementation tasks (T007-T013, T017-T023, T027-T033) follow tests
- Test validation after implementation confirms tests now PASS
- Manual testing verifies integration before moving to next story

### Parallel Opportunities

- **Foundational Phase**: All review tasks (T001-T003) can be done in parallel
- **User Story 1 Tests**: T004, T005, T006 can be written in parallel (different test files/functions)
- **User Story 2 Tests**: T014, T015, T016 can be written in parallel
- **User Story 3 Tests**: T024, T025, T026, T029, T030, T031 can all be written in parallel
- **US1 and US2 Implementation**: These can proceed in parallel once their tests are written (different parts of same file, non-conflicting)
- **Polish Phase**: T034, T035, T036, T037, T041 can run in parallel

---

## Parallel Example: User Story 1

```bash
# Phase 2: Launch all User Story 1 tests in parallel
Agent 1: "Write unit test TestAddFormPasswordVisibilityToggle in tests/unit/tui_forms_test.go"
Agent 2: "Write unit test TestAddFormCtrlHShortcut in tests/unit/tui_forms_test.go"
Agent 3: "Write integration test TestAddFormCursorPreservation in tests/integration/tui_password_toggle_test.go"

# Verify all tests FAIL (no implementation yet)
go test ./tests/... -v

# Sequential implementation (same file)
Task T007: Add passwordVisible field
Task T008: Implement togglePasswordVisibility method
Task T009: Add Ctrl+H shortcut
Task T010: Update keyboard hints

# Verify tests now PASS
go test ./tests/unit/tui_forms_test.go -run TestAddForm -v
```

---

## Parallel Example: User Story 1 & 2 Together

```bash
# If team has 2+ developers, US1 and US2 can proceed in parallel:

Developer A works on User Story 1 (Add Form):
- T004-T006: Write tests
- T007-T013: Implement add form toggle

Developer B works on User Story 2 (Edit Form):
- T014-T016: Write tests
- T017-T023: Implement edit form toggle

# Both can work simultaneously because they modify different methods in forms.go:
# - AddForm methods (lines 27-294)
# - EditForm methods (lines 296-613)
# Minimal merge conflicts expected
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Foundational (review existing code)
2. Complete Phase 2: User Story 1 (add form toggle only)
3. **STOP and VALIDATE**: Test add form independently with manual testing
4. Commit: `git commit -m "feat: Add password visibility toggle to add form (US1)"`
5. Demo/validate with stakeholder if needed

**At this point, users can verify passwords in add form - core value delivered!**

### Incremental Delivery

1. Complete Phase 1: Foundational → Codebase understood
2. Add User Story 1 (P1) → Test independently → Commit → **MVP Delivered!**
3. Add User Story 2 (P2) → Test independently → Commit → Edit form now supported
4. Add User Story 3 (P3) → Test independently → Commit → Security/UX polish complete
5. Complete Phase 5: Polish → Final validation → Ready for PR

Each story adds value without breaking previous stories.

### Parallel Team Strategy

With multiple developers:

1. Team reviews codebase together (Phase 1) - 15 minutes
2. Once reviews complete:
   - **Developer A**: User Story 1 (T004-T013) - add form
   - **Developer B**: User Story 2 (T014-T023) - edit form
3. Stories complete independently and can be committed separately
4. Both developers collaborate on User Story 3 (T024-T033) - edge cases and security tests
5. Team runs Polish phase together (T034-T041)

---

## Notes

- **Constitution Compliance**: TDD is NON-NEGOTIABLE (Principle IV) - tests MUST be written first and fail
- **Security**: No logging of password content or visibility state (Principle I)
- **Simplicity**: Mouse activation (FR-006) **EXPLICITLY DEFERRED** per research.md Section 4 - keyboard-only for MVP. FR-006 marked as deferred in spec.md. Future implementation will require custom Form rendering or separate clickable widget (estimated 4-6 hours additional work).
- **[P] tasks**: Different files or independent test functions - can run in parallel
- **[Story] label**: Maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (Red-Green-Refactor)
- Commit after each story phase for clean history
- Stop at any checkpoint to validate story independently
- Estimated total time: 2-3 hours for single developer (per quickstart.md)

