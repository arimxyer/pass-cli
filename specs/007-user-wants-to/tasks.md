# Tasks: User Configuration File

**Input**: Design documents from `/specs/007-user-wants-to/`
**Prerequisites**: plan.md (complete), spec.md (complete), research.md (complete), data-model.md (complete), quickstart.md (complete)

**Tests**: Tests are REQUIRED per Constitution Principle IV (TDD - NON-NEGOTIABLE). All test tasks must be written FIRST and must FAIL before implementation begins (red-green-refactor cycle).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
- Single project structure: `internal/`, `cmd/`, `tests/` at repository root
- Go module with CLI + TUI frontends

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure for config feature

- [X] T001 Create `internal/config/` package directory structure
- [X] T002 [P] Create `tests/config/` integration test directory structure
- [X] T003 [P] Create placeholder test fixtures directory at `tests/config/fixtures/`

**Checkpoint**: Directory structure ready for implementation

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core config infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Create `Config` struct in `internal/config/config.go` with fields for Terminal, Keybindings, LoadErrors
- [X] T005 Create `TerminalConfig` struct in `internal/config/config.go` with fields for WarningEnabled, MinWidth, MinHeight
- [X] T006 Create `Keybinding` struct in `internal/config/keybinding.go` with fields for Action, KeyString, Key, Rune, Modifiers
- [X] T007 Create `ValidationResult`, `ValidationError`, `ValidationWarning` structs in `internal/config/config.go`
- [X] T008 Implement `GetDefaults()` function in `internal/config/config.go` returning default Config with hardcoded terminal and keybinding values
- [X] T009 Implement `GetConfigPath()` function in `internal/config/config.go` using `os.UserConfigDir()` with cross-platform support
- [X] T010 Implement `Load()` function skeleton in `internal/config/config.go` (returns defaults for now)
- [X] T011 Implement `Validate()` method skeleton on Config struct (returns valid for now)
- [X] T012 [P] Create unit test file `internal/config/config_test.go` for config package
- [X] T013 [P] Create unit test file `internal/config/keybinding_test.go` for keybinding parsing
- [X] T014 [P] Add file size limit check (100 KB) to `config.Load()` before parsing in `internal/config/config.go` (safety check to prevent DoS)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Terminal Size Threshold Customization (Priority: P1) 🎯 MVP

**Goal**: Enable users to customize terminal size warning thresholds or disable warnings entirely through config file

**Independent Test**: Create a config file with custom `terminal.min_width` and `terminal.min_height`, resize terminal to trigger warning at configured thresholds. Test with warning disabled to verify no warning appears.

### Tests for User Story 1 (REQUIRED - TDD)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T015 [P] [US1] Unit test for TerminalConfig validation (positive/negative values, range checks) in `internal/config/config_test.go`
- [X] T016 [P] [US1] Unit test for terminal config merging with defaults in `internal/config/config_test.go`
- [X] T017 [P] [US1] Integration test for loading terminal config from YAML in `tests/config/cli_test.go`

### Implementation for User Story 1

- [X] T018 [P] [US1] Implement Viper-based YAML loading for terminal section in `internal/config/config.go` Load() function
- [X] T019 [P] [US1] Implement terminal config validation in `internal/config/config.go` Validate() method (check min_width > 0, min_height > 0, range 1-10000/1-1000)
- [X] T020 [US1] Implement terminal config merging with defaults in `internal/config/config.go` Load() function (depends on T018, T019)
- [X] T021 [US1] Add terminal config unusually large size warnings (warn if >500 width or >200 height) in `internal/config/config.go` Validate() method
- [X] T022 [US1] Modify TUI startup in `cmd/tui/main.go` to call `config.Load()` and use terminal config for size warning thresholds
- [X] T023 [US1] Update size warning modal display logic in TUI to respect `WarningEnabled` flag from config
- [X] T024 [US1] Add config validation error modal display in TUI startup (call PageManager method when Load() returns errors - discover actual TUI file with PageManager, likely cmd/tui/layout/pages.go or similar)

**Checkpoint**: At this point, User Story 1 should be fully functional - users can customize terminal thresholds or disable warnings via config file

---

## Phase 4: User Story 2 - Keyboard Shortcut Remapping (Priority: P2)

**Goal**: Allow users to remap keyboard shortcuts and see custom bindings reflected in all UI hints (status bar, help modal, form hints)

**Independent Test**: Create config file with custom keybindings (e.g., 'n' for add-credential instead of 'a'), verify the new key works in TUI and all UI hints reflect the change. Test with conflicting bindings to verify validation errors appear.

### Tests for User Story 2 (REQUIRED - TDD)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T025 [P] [US2] Unit test for keybinding string parsing (simple keys, modifiers, special keys) in `internal/config/keybinding_test.go`
- [X] T026 [P] [US2] Unit test for keybinding conflict detection in `internal/config/keybinding_test.go`
- [X] T027 [P] [US2] Unit test for unknown action validation in `internal/config/keybinding_test.go`
- [X] T028 [P] [US2] Integration test for custom keybindings in TUI event handling in `tests/config/validation_test.go`

### Implementation for User Story 2

- [X] T029 [P] [US2] Implement keybinding string parser `ParseKeybinding()` in `internal/config/keybinding.go` (parse "key", "ctrl+key", "alt+key", "shift+key" formats to tcell types)
- [X] T030 [P] [US2] Implement valid action name list and unknown action validation in `internal/config/keybinding.go`
- [X] T031 [US2] Implement keybinding conflict detection in `internal/config/keybinding.go` Validate() (check no duplicate key assignments)
- [X] T032 [US2] Add keybinding loading to `config.Load()` in `internal/config/config.go` (parse YAML keybindings section, validate, merge with defaults)
- [X] T033 [US2] Create keybinding registry/lookup structure in `internal/config/keybinding.go` for runtime event matching
- [X] T034 [US2] Modify event handlers to use config keybindings instead of hardcoded keys (discover actual TUI event handler files, likely cmd/tui/events/ or cmd/tui/handlers/)
- [X] T035 [US2] Update status bar rendering to display custom keybinding hints (discover actual statusbar file, likely cmd/tui/components/ or cmd/tui/views/)
- [X] T036 [US2] Update help modal to display custom keybinding hints (discover actual help modal file in cmd/tui/)
- [X] T037 [US2] Update form hints to display custom keybinding hints (discover actual form files, likely cmd/tui/forms/ or cmd/tui/components/)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - terminal config works, custom keybindings work, all UI reflects custom keys

---

## Phase 5: User Story 3 - Configuration Management Commands (Priority: P3)

**Goal**: Provide CLI commands (`init`, `edit`, `validate`, `reset`) for easy config file management without manual file editing

**Independent Test**: Run `pass-cli config init` to create config, `pass-cli config validate` to check it, `pass-cli config edit` to modify it, and `pass-cli config reset` to restore defaults. Verify all commands work correctly and provide appropriate output.

### Tests for User Story 3 (REQUIRED - TDD)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T038 [P] [US3] Integration test for `config init` command in `tests/config/cli_test.go`
- [ ] T039 [P] [US3] Integration test for `config edit` command in `tests/config/cli_test.go`
- [ ] T040 [P] [US3] Integration test for `config validate` command in `tests/config/cli_test.go`
- [ ] T041 [P] [US3] Integration test for `config reset` command in `tests/config/cli_test.go`

### Implementation for User Story 3

- [X] T042 [P] [US3] Create `cmd/config.go` with cobra command structure for `config` subcommand
- [X] T043 [P] [US3] Implement `GetEditor()` function in `internal/config/config.go` (check EDITOR env var, fallback to OS defaults: notepad on Windows, nano>vim>vi on Linux/macOS; if all fail, return clear error: "No editor found. Please set EDITOR environment variable")
- [X] T044 [P] [US3] Implement `OpenEditor()` function in `internal/config/config.go` using exec.Command
- [X] T045 [US3] Implement `config init` command in `cmd/config.go` (create config file at GetConfigPath() with commented examples, exit codes 0=success/2=file error)
- [X] T046 [US3] Implement `config edit` command in `cmd/config.go` (open config file with OpenEditor(), exit codes 0=success/2=file error)
- [X] T047 [US3] Implement `config validate` command in `cmd/config.go` (call config.Load(), display validation results with line numbers, exit codes 0=valid/1=errors)
- [X] T048 [US3] Implement `config reset` command in `cmd/config.go` (backup to .backup, write defaults, exit codes 0=success/2=file error)
- [X] T049 [US3] Add structured error messages with line numbers to validation output in `internal/config/config.go`
- [X] T050 [US3] Create default config template with comments for `config init` in `internal/config/config.go`

**Checkpoint**: All user stories should now be independently functional - terminal config, keybindings, and CLI management commands all work

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T051 [P] Add audit logging for config load attempts (file path, success/failure) in `internal/config/config.go`
- [X] T052 [P] Add audit logging for validation errors in `internal/config/config.go`
- [X] T053 [P] Create test fixtures in `tests/config/fixtures/` (valid_minimal.yml, valid_full.yml, invalid_conflict.yml, invalid_terminal_size.yml, invalid_unknown_action.yml)
- [X] T054 Add unknown field warning detection (satisfies FR-020) in `internal/config/config.go` Validate()
- [X] T055 [P] Update constitution check in `specs/007-user-wants-to/plan.md` post-implementation
- [X] T056 [P] Create user-facing documentation from quickstart.md
- [X] T057 Code cleanup and refactoring for config package
- [X] T058 Run through all quickstart.md scenarios for manual validation
- [X] T059 [P] Performance testing: verify config load <10ms, validation <5ms

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Integrates with US1 (uses same config.Load()) but is independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Uses config.Load() from US1 but is independently testable

### Within Each User Story

- Tests MUST be written FIRST and MUST FAIL before implementation (TDD mandatory)
- Core config structs before parsing/validation logic
- Config loading before TUI integration
- CLI commands depend on config library functions
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel (Phase 1)
- All Foundational tasks marked [P] can run in parallel (Phase 2)
- Once Foundational phase completes, all user stories CAN start in parallel (if team capacity allows)
  - BUT recommended to do sequentially (P1 → P2 → P3) for single developer
- All tests for a user story marked [P] can run in parallel
- Within each story, implementation tasks marked [P] can run in parallel
- All Polish tasks marked [P] can run in parallel (Phase 6)

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (if tests requested):
Task: "Unit test for TerminalConfig validation in internal/config/config_test.go"
Task: "Unit test for terminal config merging in internal/config/config_test.go"
Task: "Integration test for loading terminal config in tests/config/cli_test.go"

# After tests written, launch parallel implementation:
Task: "Viper YAML loading for terminal section in internal/config/config.go"
Task: "Terminal config validation in internal/config/config.go"
```

---

## Parallel Example: User Story 2

```bash
# Launch all keybinding tasks together:
Task: "Keybinding string parser in internal/config/keybinding.go"
Task: "Valid action name list and validation in internal/config/keybinding.go"

# Then launch UI update tasks in parallel:
Task: "Update status bar rendering in cmd/tui/components/statusbar.go"
Task: "Update form hints in cmd/tui/components/forms.go"
```

---

## Parallel Example: User Story 3

```bash
# Launch config command implementations in parallel:
Task: "Create cmd/config.go with cobra structure"
Task: "GetEditor() function in internal/config/config.go"
Task: "OpenEditor() function in internal/config/config.go"

# Then launch individual command implementations:
Task: "config init command in cmd/config.go"
Task: "config edit command in cmd/config.go"
Task: "config validate command in cmd/config.go"
Task: "config reset command in cmd/config.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T014) - CRITICAL, blocks all stories
3. Complete Phase 3: User Story 1 (T015-T024)
4. **STOP and VALIDATE**: Test User Story 1 independently using quickstart.md scenarios 1.1-1.4
5. Commit and deploy/demo if ready

**MVP Delivers**: Terminal size threshold customization - users can set custom thresholds or disable warnings

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently (quickstart scenarios 1.1-1.4) → Commit → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently (quickstart scenarios 2.1-2.5) → Commit → Deploy/Demo
4. Add User Story 3 → Test independently (quickstart scenarios 3.1-3.5) → Commit → Deploy/Demo
5. Complete Polish phase → Final validation with all edge cases → Release
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers (not typical for this project, but possible):

1. Team completes Setup + Foundational together (T001-T014)
2. Once Foundational is done:
   - Developer A: User Story 1 (Terminal config)
   - Developer B: User Story 2 (Keybindings)
   - Developer C: User Story 3 (CLI commands)
3. Stories complete and integrate independently (all use common config.Load())
4. Integration testing to ensure all three work together

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Tests MUST be written first and MUST fail before implementation (TDD red-green-refactor)
- Commit after each task or logical group of tasks
- Stop at any checkpoint to validate story independently
- Constitution Principle IV (TDD) is satisfied: tests are REQUIRED and must be written first
- Cross-platform support is embedded in tasks (Windows/macOS/Linux)
- File size limit (100 KB) enforced in Phase 2 foundational tasks (T014)
- All UI updates (status bar, help, forms) must reflect custom keybindings

---

## Task Count Summary

- **Total Tasks**: 59
- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 11 tasks (includes file size limit check)
- **Phase 3 (User Story 1)**: 10 tasks (3 test tasks + 7 implementation tasks, includes terminal size warnings)
- **Phase 4 (User Story 2)**: 13 tasks (4 test tasks + 9 implementation tasks)
- **Phase 5 (User Story 3)**: 13 tasks (4 test tasks + 9 implementation tasks)
- **Phase 6 (Polish)**: 9 tasks (removed duplicate file size check and merged terminal warnings)

**Parallel Opportunities**: 26 tasks marked [P] can run in parallel within their phases

**MVP Scope (Recommended)**: Phase 1 + Phase 2 + Phase 3 = 24 tasks (delivers User Story 1)
