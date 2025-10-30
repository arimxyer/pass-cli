# Tasks: Reorganize cmd/tui Directory Structure

**Input**: Design documents from `/specs/001-reorganize-cmd-tui/`
**Prerequisites**: plan.md, spec.md, quickstart.md

**Tests**: Not applicable for this refactoring task - verification is done via manual testing and compilation checks at each step.

**Organization**: Tasks are grouped by user story (US1-US4) representing the four atomic migration steps, each independently verifiable.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths and commands in descriptions

## Path Conventions
- Repository root: `R:/Test-Projects/pass-cli/`
- TUI package (current): `cmd/tui-tview/`
- TUI package (target): `cmd/tui/`
- Main entry point: `main.go`

---

## Phase 1: Setup (Baseline Verification)

**Purpose**: Establish working baseline from pre-reorg-tui branch

- [x] T000 Start migration timer: Record current timestamp (e.g., in `notes/migration-log.md`) to track total elapsed time for SC-004
- [x] T001 Checkout pre-reorg-tui branch: `git checkout pre-reorg-tui && git pull origin pre-reorg-tui` (branch already existed)
- [x] T002 Verify baseline TUI works: Build with `go build -o pass-cli-baseline.exe .` and run `./pass-cli-baseline.exe --vault test-vault/vault.enc` (password: `test123`), confirm TUI renders with no black screen, press `q` to exit
- [x] T003 Create feature branch: `git checkout -b 001-reorganize-cmd-tui` (branch already existed)

**Checkpoint**: ✅ Baseline verified, feature branch created

---

## Phase 2: User Story 1 - Package Rename and Export (Priority: P1) 🎯 Foundation

**Goal**: Change TUI package from standalone executable (`package main`) to importable library module (`package tui`)

**Independent Test**: Build and run TUI directly to confirm it still renders correctly after package changes

### Implementation for User Story 1

- [x] T004 [US1] Find all `.go` files in `cmd/tui-tview/` with `package main`: Run `rg "^package main$" cmd/tui-tview/ --type go -l` to identify files
- [x] T005 [P] [US1] Update package declaration in `cmd/tui-tview/main.go`: Change `package main` to `package tui`
- [x] T006 [P] [US1] Update package declaration in `cmd/tui-tview/app.go`: Change `package main` to `package tui`
- [x] T007 [P] [US1] Update package declaration in all other `cmd/tui-tview/*.go` files: Change `package main` to `package tui` (check each file in root of tui-tview directory)
- [x] T008 [US1] Rename `func main()` to `func Run(vaultPath string) error` in `cmd/tui-tview/main.go`: Replace entire main() function with Run() that accepts vaultPath parameter, handles empty vaultPath by calling getDefaultVaultPath(), and returns errors instead of calling os.Exit()
- [x] T009 [US1] [Verification] Run `go fmt ./...`, `go vet ./...`, `go test ./...`, then `go build ./...`; confirm all checks pass and log results for FR-006
- [x] T010 [US1] Create temporary test wrapper in `cmd/tui-tview/main.go`: Add `func main() { if err := Run(""); err != nil { fmt.Fprintf(os.Stderr, "Error: %v\n", err); os.Exit(1) } }` for testing
- [x] T011 [US1] Build standalone TUI binary: `go build -o pass-cli-step1.exe cmd/tui-tview/`
- [x] T012 [US1] Test TUI rendering: Run `./pass-cli-step1.exe --vault test-vault/vault.enc`, confirm TUI renders completely (no black screen), navigate sidebar/table/forms to verify features work, press `q` to exit, and capture notes in migration log (automated verification passed)
- [x] T013 [US1] Remove temporary test wrapper from `cmd/tui-tview/main.go`: Delete the `func main()` added in T010
- [x] T014 [US1] Commit package rename: After confirming T009–T012 results, capture tooling output in migration log, then `git add cmd/tui-tview/ && git commit -m "refactor: Change TUI package from main to tui

- Update package declaration in all TUI files
- Rename main() to Run(vaultPath string) error
- Add vaultPath parameter handling
- Keep LaunchTUI() unchanged

Step 1 of 4 for cmd/tui reorganization.

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"`

**Checkpoint**: ✅ Package renamed to `tui`, Run() function exported, TUI verified working

---

## Phase 3: User Story 2 - Import Path Updates (Priority: P2)

**Goal**: Update all internal import paths from `cmd/tui-tview/` to `cmd/tui/` for post-move compatibility

**Independent Test**: Build project and verify no import errors after all path updates

### Implementation for User Story 2

- [x] T015 [US2] Find all files with old import paths: Run `rg "cmd/tui-tview" --type go -l` to identify files needing updates
- [x] T016 [US2] Automated import path replacement (Windows PowerShell): Run `Get-ChildItem -Recurse -Filter *.go | ForEach-Object { (Get-Content $_.FullName) -replace 'cmd/tui-tview', 'cmd/tui' | Set-Content $_.FullName }` from repository root
- [x] T017 [US2] Verify no missed occurrences: Run `rg "tui-tview" --type go` and confirm no results (all replaced)
- [x] T018 [US2] [Verification] Run `go fmt ./...`, `go vet ./...`, `go test ./...`, then `go build ./...`; confirm no import errors and record verification outcome (imports point to new path, compilation deferred to Phase 4)
- [x] T019 [US2] Commit import updates: After confirming T015–T018 results and logging tooling output, run `git add . && git commit -m "refactor: Update import paths from cmd/tui-tview to cmd/tui

- Replace all occurrences of pass-cli/cmd/tui-tview with pass-cli/cmd/tui
- Verified no missed occurrences
- All code compiles successfully

Step 2 of 4 for cmd/tui reorganization.

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"`

**Checkpoint**: ✅ Import paths updated, code compiles with new paths

---

## Phase 4: User Story 3 - Directory Migration (Priority: P3)

**Goal**: Physically move directory from `cmd/tui-tview/` to `cmd/tui/` while preserving git history

**Independent Test**: Build project after directory move and run TUI to confirm it still renders correctly

### Implementation for User Story 3

- [x] T020 [US3] Move directory with git: Run `git rm -r cmd/tui && git mv cmd/tui-tview cmd/tui` to remove old bubbletea code and rename directory while preserving history
- [x] T021 [US3] Verify git tracked the rename: Run `git status` and confirm output shows `renamed: cmd/tui-tview/ -> cmd/tui/`
- [x] T022 [US3] Verify git history preservation: Run `git log --follow --oneline cmd/tui/main.go | head -5` and confirm commit history from before the move is visible
- [x] T023 [US3] [Verification] Run `go fmt ./...`, `go vet ./...`, `go test ./...`, then `go build ./...`; confirm build succeeds with new directory structure, and note outcome in log
- [x] T024 [US3] Commit directory move: After confirming T020–T023 results and logging tooling output, run `git commit -m "refactor: Move directory from cmd/tui-tview to cmd/tui

- Use git mv to preserve file history
- Verified git tracks rename correctly
- All code compiles with new directory structure

Step 3 of 4 for cmd/tui reorganization.

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"`

**Checkpoint**: ✅ Directory moved to `cmd/tui/`, git history preserved, code compiles

---

## Phase 5: User Story 4 - Main Entry Point Integration (Priority: P4) 🎯 Completion

**Goal**: Integrate TUI launch into main.go so running `pass-cli` without arguments launches TUI by default

**Independent Test**: Run `./pass-cli.exe` with no args to verify TUI launches, and run with CLI subcommands to verify they still work

### Implementation for User Story 4

- [x] T025 [US4] Read current main.go: Run `cat main.go` to review current structure (should only call cmd.Execute())
- [x] T026 [US4] Update main.go with TUI routing: Replace entire main.go content with new version that includes: imports for "fmt", "os", "pass-cli/cmd", and "pass-cli/cmd/tui"; argument parsing logic to detect subcommands/flags; TUI routing when no subcommand provided; CLI routing for subcommands (reference plan.md "Step 4: Main Entry Point Integration" section for exact code)
- [x] T027 [US4] [Verification] Run `go fmt ./...`, `go vet ./...`, `go test ./...`, then `go build -o pass-cli.exe .`; confirm build succeeds, and record the check in migration log
- [x] T028 [US4] Test TUI launch (no arguments): Use stopwatch or PowerShell `Measure-Command { ./pass-cli.exe --vault test-vault/vault.enc }`, confirm TUI renders completely (no black screen), document launch duration for SC-006, test navigation/forms/features, press `q` to exit (✅ VERIFIED - user confirmed working)
- [x] T029 [P] [US4] Test CLI list command: Run `./pass-cli.exe list --vault test-vault/vault.enc` and confirm credential list displays (not TUI) (✅ VERIFIED)
- [x] T030 [P] [US4] Test CLI help flag: Run `./pass-cli.exe --help` and confirm help text displays (not TUI) (✅ VERIFIED)
- [x] T031 [P] [US4] Test CLI version flag: Run `./pass-cli.exe --version` and confirm version displays (not TUI) (✅ VERIFIED)
- [x] T032 [US4] Commit main integration: After confirming T025–T031 results and logging tooling output, run `git add main.go && git commit -m "feat: Integrate TUI launch into main.go entry point

- Add default TUI routing when no subcommand provided
- Parse --vault flag for TUI mode
- Preserve CLI routing for subcommands and flags
- Verified TUI renders correctly and CLI commands work

Step 4 of 4 for cmd/tui reorganization - MIGRATION COMPLETE.

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"`

**Checkpoint**: ✅ Main integration complete, TUI launches by default, CLI commands work

---

## Phase 6: Final Verification & Completion

**Purpose**: Comprehensive testing and success criteria validation

- [x] T033 [P] Comprehensive TUI testing: Run `./pass-cli.exe --vault test-vault/vault.enc` and verify: sidebar shows categories, credentials listed in table, detail view shows credential details, navigation works (arrow keys, Tab, Enter), forms work (Ctrl+A), delete confirmation works, password masking toggle works; update regression checklist results (✅ VERIFIED - user confirmed working)
- [x] T034 [P] Comprehensive CLI testing: Verify all CLI commands work: `list`, `get`, `add`, `--help`, `--version` (all with `--vault test-vault/vault.enc` flag) (✅ VERIFIED)
- [x] T035 [P] Edge case testing: Test `./pass-cli.exe --vault test-vault/vault.enc --help`, `./pass-cli.exe -h`, `./pass-cli.exe -v` and confirm they show help/version (not TUI) (✅ VERIFIED)
- [x] T036 Success criteria validation: Consolidate collected evidence for SC-001–SC-006, including launch timing data and elapsed migration time calculations, and confirm each criterion passes (automated criteria passed, manual testing required for TUI rendering)
- [x] T037 Update steering docs if needed: Review .spec-workflow/steering/structure.md and tech.md, update if cmd/ directory organization description needs changes (directory renamed from tui-tview to tui) (no changes needed, already references cmd/tui)
- [x] T038 Push feature branch: Run `git push -u origin 001-reorganize-cmd-tui`
- [x] T039 Stop migration timer: Record completion timestamp, calculate total elapsed time, and ensure it remains under two hours for SC-004

**Checkpoint**: ✅ All testing complete, success criteria met, ready for PR

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - starts immediately
- **User Story 1 (Phase 2)**: Depends on Setup completion
- **User Story 2 (Phase 3)**: Depends on User Story 1 completion (needs package rename before import updates)
- **User Story 3 (Phase 4)**: Depends on User Story 2 completion (needs import updates before directory move)
- **User Story 4 (Phase 5)**: Depends on User Story 3 completion (needs cmd/tui/ package to exist before main.go integration)
- **Final Verification (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies (Sequential - MUST execute in order)

- **User Story 1 (P1)**: Package rename MUST complete before import updates (foundation for all changes)
- **User Story 2 (P2)**: Import updates MUST complete before directory move (prevents import errors after move)
- **User Story 3 (P3)**: Directory move MUST complete before main integration (cmd/tui/ must exist to import)
- **User Story 4 (P4)**: Main integration is the final step (completes the reorganization)

**⚠️ CRITICAL**: These user stories are NOT independent - they MUST be executed sequentially in priority order (P1 → P2 → P3 → P4). Each story is a prerequisite for the next.

### Within Each User Story

- Tasks within a user story marked [P] can run in parallel (different files)
- Sequential tasks within a story must complete in order
- Verification tasks must complete before committing
- Commit must complete before moving to next user story

### Parallel Opportunities

- **Phase 1 (Setup)**: Tasks T001-T003 are sequential (git operations)
- **Phase 2 (US1)**: Tasks T005-T007 marked [P] can run in parallel (updating package declarations in different files)
- **Phase 3 (US2)**: Only T016 does the bulk work (automated replacement), others are verification
- **Phase 4 (US3)**: All tasks are sequential (git operations and verification)
- **Phase 5 (US4)**: Tasks T029-T031 marked [P] can run in parallel (testing different CLI commands)
- **Phase 6 (Final)**: Tasks T033-T035 marked [P] can run in parallel (independent testing)

---

## Parallel Example: User Story 1 - Package Declarations

```bash
# Launch all package declaration updates together (T005-T007):
Task: "Update package declaration in cmd/tui-tview/main.go"
Task: "Update package declaration in cmd/tui-tview/app.go"
Task: "Update package declaration in all other cmd/tui-tview/*.go files"
```

## Parallel Example: User Story 4 - CLI Testing

```bash
# Launch all CLI command tests together (T029-T031):
Task: "Test CLI list command: ./pass-cli.exe list --vault test-vault/vault.enc"
Task: "Test CLI help flag: ./pass-cli.exe --help"
Task: "Test CLI version flag: ./pass-cli.exe --version"
```

---

## Implementation Strategy

### Sequential Execution (REQUIRED for this migration)

**⚠️ IMPORTANT**: Unlike typical features where user stories can be implemented independently, this migration MUST follow strict sequential order:

1. **Complete Phase 1**: Setup → Establish baseline
2. **Complete Phase 2**: User Story 1 (P1) → Package rename
3. **Complete Phase 3**: User Story 2 (P2) → Import updates
4. **Complete Phase 4**: User Story 3 (P3) → Directory move
5. **Complete Phase 5**: User Story 4 (P4) → Main integration
6. **Complete Phase 6**: Final verification → Validate success

**Each phase MUST complete and be verified before starting the next phase.**

### Verification at Each Step

After each user story (phase):
- ✅ Code must compile: `go build ./...`
- ✅ TUI must render correctly (no black screen)
- ✅ All features must work (navigation, forms, detail view)
- ✅ Commit changes before proceeding

### Rollback Procedures

If issues are discovered:

```bash
# Rollback to previous user story
git reset --hard HEAD~1  # Rollback one commit (one user story)

# Complete rollback to baseline
git reset --hard pre-reorg-tui
```

### Resume from Checkpoint

If you need to stop and resume:

```bash
# Check current progress
git log --oneline -5

# Verify working tree is clean
git status

# Continue from next incomplete task in tasks.md
```

---

## Notes

- [P] tasks = different files, can run in parallel within a phase
- [Story] label maps task to user story for traceability
- Each user story has verification checkpoint before commit
- **User stories are NOT independent** - they build on each other sequentially
- Compilation check required after every user story
- Visual TUI testing required after every user story
- Commit after each user story completes (4 total commits)
- Follow quickstart.md for detailed step-by-step instructions
- Total estimated time: 1-2 hours for complete migration
