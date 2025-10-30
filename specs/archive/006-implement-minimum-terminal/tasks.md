# Tasks: Minimum Terminal Size Enforcement

**Input**: Design documents from `/specs/006-implement-minimum-terminal/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, quickstart.md ✅

**Tests**: ✅ INCLUDED - Constitution Principle IV (TDD) requires tests before implementation
**Organization**: Tasks grouped by user story for independent implementation and testing

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify existing project structure - no setup needed

✅ **SKIPPED** - Feature extends existing TUI infrastructure. All dependencies already in place:
- Go 1.25.1 project initialized
- tview v0.42.0 and tcell/v2 v2.9.0 already in go.mod
- Test infrastructure exists in `cmd/tui/layout/*_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add minimum size constants that all user stories depend on

**⚠️ CRITICAL**: This phase must complete before ANY user story implementation begins

- [X] T001 [Foundational] Add MinTerminalWidth=60 and MinTerminalHeight=30 constants to `cmd/tui/layout/manager.go` at package level (after import statements, before LayoutMode type definition)

**Checkpoint**: Foundation ready - constants defined, user story implementation can now begin

---

## Phase 3: User Story 1 - Terminal Size Warning Display (Priority: P1) 🎯 MVP

**Goal**: Display blocking warning overlay when terminal is resized below 60×30, showing current vs. required dimensions

**Independent Test**: Launch app in 50×20 terminal → warning appears with "Current: 50x20, Minimum required: 60x30" message

### Tests for User Story 1 (TDD - WRITE FIRST) ⚠️

**CONSTITUTION REQUIREMENT**: Write these tests FIRST, get user approval, verify they FAIL, then implement

- [X] T002 [P] [US1] Write unit test `TestShowSizeWarning` in `cmd/tui/layout/pages_test.go` - verify warning page added, state flag set, message format correct
- [X] T003 [P] [US1] Write unit test `TestSizeWarningMessage` in `cmd/tui/layout/pages_test.go` - verify message includes current dimensions (50×20) and minimum (60×30)
- [X] T004 [P] [US1] Write unit test `TestSizeWarningStateTracking` in `cmd/tui/layout/pages_test.go` - verify `IsSizeWarningActive()` returns correct boolean state
- [X] T005 [P] [US1] Write unit test `TestHandleResize_BelowMinimum` in `cmd/tui/layout/manager_test.go` - verify HandleResize calls ShowSizeWarning when width<60 OR height<30
- [X] T006 [P] [US1] Write unit test `TestHandleResize_StartupCheck` in `cmd/tui/layout/manager_test.go` - verify startup size check triggers warning if already too small

**USER APPROVAL GATE**: Present tests T002-T006 to user, verify they fail before proceeding to implementation

### Implementation for User Story 1

- [X] T007 [US1] Add `sizeWarningActive bool` field to PageManager struct in `cmd/tui/layout/pages.go` (after modalStack field)
- [X] T008 [US1] Implement `ShowSizeWarning(currentWidth, currentHeight, minWidth, minHeight int)` method in `cmd/tui/layout/pages.go` - create tview.Modal with dark red background, formatted message, add page "size-warning", set state flag, call app.Draw()
- [X] T009 [US1] Implement `IsSizeWarningActive() bool` method in `cmd/tui/layout/pages.go` - return sizeWarningActive field value
- [X] T010 [US1] Modify `HandleResize(width, height int)` in `cmd/tui/layout/manager.go` - add size check `if width < MinTerminalWidth || height < MinTerminalHeight` before existing layout mode logic, call pageManager.ShowSizeWarning() if condition true

**Checkpoint**: User Story 1 complete - warning displays on startup and during resize when terminal < 60×30. Run tests T002-T006 to verify.

---

## Phase 4: User Story 2 - Automatic Recovery (Priority: P2)

**Goal**: Automatically hide warning and restore interface when terminal is resized back to adequate dimensions (≥60×30)

**Independent Test**: Trigger warning (50×20 terminal), then resize to 80×40 → warning disappears, interface functional

### Tests for User Story 2 (TDD - WRITE FIRST) ⚠️

**CONSTITUTION REQUIREMENT**: Write these tests FIRST, get user approval, verify they FAIL, then implement

- [X] T011 [P] [US2] Write unit test `TestHideSizeWarning` in `cmd/tui/layout/pages_test.go` - verify warning page removed, state flag cleared
- [X] T012 [P] [US2] Write unit test `TestHideSizeWarning_WhenNotActive` in `cmd/tui/layout/pages_test.go` - verify safe no-op when warning not showing (idempotency test)
- [X] T013 [P] [US2] Write unit test `TestHandleResize_ExactlyAtMinimum` in `cmd/tui/layout/manager_test.go` - verify 60×30 does NOT trigger warning (inclusive boundary)
- [X] T014 [P] [US2] Write unit test `TestHandleResize_PartialFailure` in `cmd/tui/layout/manager_test.go` - verify 70×25 triggers warning (height < 30, OR logic)
- [X] T015 [P] [US2] Write integration test `TestFullResizeFlow_ShowAndHide` in `tests/integration/tui_resize_test.go` - verify end-to-end flow: start 50×20 (warning shows), resize 80×40 (warning hides), interface functional

**USER APPROVAL GATE**: Present tests T011-T015 to user, verify they fail before proceeding to implementation

### Implementation for User Story 2

- [X] T016 [US2] Implement `HideSizeWarning()` method in `cmd/tui/layout/pages.go` - check if sizeWarningActive, if true: remove "size-warning" page, clear state flag, call app.Draw()
- [X] T017 [US2] Modify `HandleResize(width, height int)` in `cmd/tui/layout/manager.go` - add else branch after size check to call pageManager.HideSizeWarning() when dimensions adequate
- [X] T018 [US2] Verify `HandleResize` does NOT return early after showing warning in `cmd/tui/layout/manager.go` - layout mode determination must continue for recovery to work

**Checkpoint**: User Stories 1 AND 2 complete - warning shows/hides automatically on resize. Run tests T011-T015 to verify recovery flow.

---

## Phase 5: User Story 3 - Visual Clarity and Feedback (Priority: P3)

**Goal**: Ensure warning is visually distinct (dark red background) with clear, actionable message

**Independent Test**: Trigger warning → verify dark red background, plain language message ("Terminal too small!"), clear instructions ("Please resize your terminal window")

### Tests for User Story 3 (TDD - WRITE FIRST) ⚠️

**CONSTITUTION REQUIREMENT**: Write these tests FIRST, get user approval, verify they FAIL, then implement

- [X] T019 [P] [US3] Write unit test `TestSizeWarningVisualStyle` in `cmd/tui/layout/pages_test.go` - verify modal has dark red background color (tcell.ColorDarkRed)
- [X] T020 [P] [US3] Write unit test `TestSizeWarningMessageClarity` in `cmd/tui/layout/pages_test.go` - verify message contains "Terminal too small!", no technical jargon
- [X] T021 [P] [US3] Write unit test `TestSizeWarningActionableInstructions` in `cmd/tui/layout/pages_test.go` - verify message contains "Please resize your terminal window"

**USER APPROVAL GATE**: Present tests T019-T021 to user, verify they fail before proceeding to implementation

### Implementation for User Story 3

- [X] T022 [US3] Update `ShowSizeWarning` message format in `cmd/tui/layout/pages.go` - ensure message reads: "Terminal too small!\n\nCurrent: {W}x{H}\nMinimum required: {minW}x{minH}\n\nPlease resize your terminal window." (plain language, clear instructions)
- [X] T023 [US3] Confirm implementation includes `SetBackgroundColor(tcell.ColorDarkRed)` call in `ShowSizeWarning` method in `cmd/tui/layout/pages.go` - ensure dark red background is set per visual style requirements (US3)

**Checkpoint**: All three user stories complete - warning is functional, auto-recovers, and visually clear. Run all tests T002-T021 to verify.

---

## Phase 6: Edge Cases & Polish

**Purpose**: Handle edge cases and cross-cutting concerns

### Edge Case Tests (TDD - WRITE FIRST) ⚠️

- [X] T024 [P] [EdgeCase] Write unit test `TestHandleResize_RapidOscillation` in `cmd/tui/layout/manager_test.go` - loop 100 times alternating 50×20 / 80×40, verify no crashes or memory leaks (per clarification: immediate show/hide acceptable)
- [X] T025 [P] [EdgeCase] Write integration test `TestModalPreservation_DuringWarning` in `tests/integration/tui_resize_test.go` - open form modal, resize below minimum, verify warning overlays on top, form state preserved
- [X] T026 [P] [EdgeCase] Write unit test `TestHandleResize_BoundaryEdgeCases` in `cmd/tui/layout/manager_test.go` - test 59×30 (warn), 60×29 (warn), 61×31 (no warn)

**USER APPROVAL GATE**: Present tests T024-T026 to user, verify they fail before proceeding

### Edge Case Implementation

- [X] T027 [EdgeCase] Verify HandleResize continues layout updates even when warning shown in `cmd/tui/layout/manager.go` - confirm no early return after ShowSizeWarning call (prevents recovery deadlock)
- [X] T028 [EdgeCase] Add explicit `app.Draw()` calls in ShowSizeWarning and HideSizeWarning in `cmd/tui/layout/pages.go` - ensures <100ms response time per success criteria

### Documentation & Polish

- [X] T029 [P] [Polish] Add code comments to ShowSizeWarning/HideSizeWarning methods in `cmd/tui/layout/pages.go` - document behavior, parameters, and state management
- [X] T030 [P] [Polish] Add code comments to MinTerminal constants in `cmd/tui/layout/manager.go` - explain why 60×30 chosen (LayoutSmall mode usability)
- [X] T031 [P] [Polish] Update CLAUDE.md context if needed - add reference to minimum size constants location (agent context script already ran, verify completeness)
- [X] T032 [Polish] Run quickstart.md manual verification tests - execute all test scenarios from `specs/006-implement-minimum-terminal/quickstart.md` (covers FR-010: warning readability at extremely small sizes)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: ✅ SKIPPED - existing infrastructure sufficient
- **Phase 2 (Foundational)**: Can start immediately - BLOCKS all user stories
  - **T001** must complete before any user story begins
- **Phase 3 (US1)**: Depends on Phase 2 completion
  - Tests T002-T006 → User approval → Implementation T007-T010
- **Phase 4 (US2)**: Depends on Phase 2 completion (independent of US1, but logically builds on it)
  - Tests T011-T015 → User approval → Implementation T016-T018
- **Phase 5 (US3)**: Depends on Phase 2 completion (refines US1 implementation)
  - Tests T019-T021 → User approval → Implementation T022-T023
- **Phase 6 (Edge Cases)**: Depends on all user stories being complete
  - Tests T024-T026 → Implementation T027-T028 → Polish T029-T032

### User Story Dependencies

- **US1 (P1)**: Requires T001 (constants) - No dependencies on other stories
- **US2 (P2)**: Requires T001 (constants) + US1 implementation (T007-T010) - Uses ShowSizeWarning from US1
- **US3 (P3)**: Requires T001 (constants) + US1 implementation (T007-T010) - Refines ShowSizeWarning from US1

**Note**: US2 and US3 both depend on US1. They could theoretically run in parallel but sequencing P1→P2→P3 is recommended for logical flow.

### TDD Workflow Per User Story

For each user story phase:
1. Write all test tasks (marked with T-numbers)
2. **STOP**: Present tests to user for approval
3. Run tests → verify they FAIL (red)
4. Implement corresponding implementation tasks
5. Run tests → verify they PASS (green)
6. Checkpoint: Verify story independently
7. Move to next priority story

### Parallel Opportunities

**Within Foundational Phase**:
- Only T001 - no parallelization possible

**Within Each User Story Test Phase**:
- All tests marked [P] for that story can run in parallel
  - US1: T002, T003, T004, T005, T006 (5 parallel test writes)
  - US2: T011, T012, T013, T014, T015 (5 parallel test writes)
  - US3: T019, T020, T021 (3 parallel test writes)
  - EdgeCase: T024, T025, T026 (3 parallel test writes)

**Within Each User Story Implementation Phase**:
- US1: T007 and T010 can run in parallel (different methods)
  - T008-T009 sequential (same file, related methods)
- US2: T016 and T017 parallel (different methods)
- US3: T022 and T023 parallel (independent verification)
- EdgeCase: T027 and T028 parallel (independent verifications)

**Across User Stories** (if team capacity allows):
- After US1 completes: US2 and US3 could theoretically proceed in parallel
- However, sequential order P1→P2→P3 recommended for logical dependencies

**Polish Phase**:
- T029, T030, T031 can all run in parallel (different files)

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all test writes for US1 together:
Task: "Write TestShowSizeWarning in cmd/tui/layout/pages_test.go"
Task: "Write TestSizeWarningMessage in cmd/tui/layout/pages_test.go"
Task: "Write TestSizeWarningStateTracking in cmd/tui/layout/pages_test.go"
Task: "Write TestHandleResize_BelowMinimum in cmd/tui/layout/manager_test.go"
Task: "Write TestHandleResize_StartupCheck in cmd/tui/layout/manager_test.go"

# Then present to user for approval before implementation
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (T001) → Constants defined
2. Complete Phase 3: User Story 1 (T002-T010)
   - Write tests T002-T006 → Get user approval → Verify FAIL
   - Implement T007-T010 → Verify tests PASS
3. **STOP and VALIDATE**: Test US1 independently per quickstart.md
4. **MVP READY**: Warning displays on small terminal, shows correct dimensions

### Incremental Delivery

1. Phase 2 (T001) → Foundation ready
2. Phase 3 (US1 - T002-T010) → Test independently → **MVP Deploy**
3. Phase 4 (US2 - T011-T018) → Test independently → Recovery works
4. Phase 5 (US3 - T019-T023) → Test independently → Visual polish complete
5. Phase 6 (EdgeCases - T024-T032) → All edge cases handled → **Feature Complete**

### TDD Compliance Strategy

**CONSTITUTION PRINCIPLE IV ENFORCEMENT**:

1. **For each user story**:
   - Agent writes all test tasks
   - Agent presents tests to user: "Here are tests T002-T006 for User Story 1. Please review and approve."
   - **USER MUST APPROVE BEFORE PROCEEDING**
   - Agent runs tests → confirms RED (failing)
   - Only then: Agent implements corresponding code
   - Agent runs tests → confirms GREEN (passing)
   - Agent marks story checkpoint complete

2. **Commit strategy per CLAUDE.md**:
   - Commit after test write phase: "test: add tests for US1 terminal size warning"
   - Commit after implementation phase: "feat: implement US1 terminal size warning display"
   - Commit after each user story phase completion

3. **No shortcuts allowed**:
   - Cannot implement without tests
   - Cannot mark tasks complete if tests failing
   - Cannot skip user approval gate

---

## Task Summary

**Total Tasks**: 32
- **Foundational**: 1 task (T001)
- **User Story 1 (MVP)**: 9 tasks (5 tests + 4 implementation)
- **User Story 2**: 8 tasks (5 tests + 3 implementation)
- **User Story 3**: 5 tasks (3 tests + 2 implementation)
- **Edge Cases & Polish**: 9 tasks (3 tests + 2 implementation + 4 polish)

**Test Tasks**: 16 (50% of total - TDD compliant)
**Implementation Tasks**: 12
**Polish Tasks**: 4

**Parallel Opportunities**:
- 16 test tasks can run in parallel (within their phases)
- 8 implementation tasks can run in parallel (with dependencies respected)
- 3 polish tasks can run in parallel

**Independent Test Criteria**:
- **US1**: Launch in 50×20 terminal → warning shows with correct message
- **US2**: Show warning, resize to 80×40 → warning hides, interface functional
- **US3**: Warning has dark red background, plain language, clear instructions

**Suggested MVP Scope**: Phase 2 + Phase 3 (US1) = T001-T010 (10 tasks)

---

## Notes

- [P] = Parallelizable (different files, no dependencies)
- [Story] = US1/US2/US3/EdgeCase/Foundational/Polish
- TDD workflow enforced per Constitution Principle IV
- Each user story independently testable per spec requirements
- Tests must FAIL before implementation (Red-Green-Refactor)
- Commit after each task or logical group per CLAUDE.md
- Stop at each checkpoint to validate story completion
- User approval gates are NON-NEGOTIABLE per constitution
