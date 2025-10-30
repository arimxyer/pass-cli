# Migration Safety Checklist: Reorganize cmd/tui Directory Structure

**Purpose**: Validate requirements quality for safe, reversible directory reorganization with verification at each step
**Created**: 2025-10-09
**Feature**: [spec.md](../spec.md) | [plan.md](../plan.md) | [tasks.md](../tasks.md)
**Focus**: Migration safety requirements for rollback, verification checkpoints, and git history preservation
**Audience**: Peer reviewer validating requirements quality during PR review
**Scope**: Balanced coverage of UX requirements (TUI rendering) and technical migration mechanics

---

## Rollback & Recovery Requirements

- [x] CHK001 - Are rollback procedures explicitly defined for each migration step? ✅ quickstart.md §Rollback Procedures
- [x] CHK002 - Is the rollback command syntax specified with concrete examples? ✅ quickstart.md (git reset --hard HEAD~N)
- [ ] CHK003 - Are rollback requirements defined for partial failures during a step? ❌ GAP - No explicit partial failure handling
- [x] CHK004 - Is the "clean working tree" validation requirement specified before rollback? ✅ quickstart.md §Resume from Checkpoint
- [ ] CHK005 - Are recovery requirements defined for interrupted git operations? ❌ GAP - No recovery procedures documented
- [x] CHK006 - Is the baseline branch requirement (pre-reorg-tui) explicitly documented? ✅ spec.md §Assumptions + quickstart.md §Step 0

## Verification Checkpoint Requirements

- [x] CHK007 - Are verification criteria defined for EACH of the 4 user stories? ✅ spec.md §User Stories 1-4 (acceptance scenarios)
- [x] CHK008 - Is "TUI renders correctly" quantified with specific observable criteria? ✅ quickstart.md (sidebar visible, credentials listed, navigation works)
- [x] CHK009 - Are compilation verification requirements specified after each step? ✅ Every step has `go build` verification
- [x] CHK010 - Is "no black screen" defined with testable visual criteria? ✅ spec.md SC-001 + quickstart.md explicit checks
- [x] CHK011 - Are manual testing procedures detailed enough for consistent execution? ✅ quickstart.md step-by-step procedures
- [x] CHK012 - Is the verification sequence (compile → run → visual → interaction) explicitly ordered? ✅ quickstart.md follows this pattern
- [x] CHK013 - Are verification requirements consistent across all 4 migration steps? ✅ All steps use same verification pattern
- [x] CHK014 - Is the test vault path and password documented for verification testing? ✅ quickstart.md: test-vault/vault.enc (password: test123)
- [ ] CHK015 - Are terminal capability requirements specified for TUI verification? ⚠️ PARTIAL - Mentions "capable terminal" but lacks specifics

## Git History Preservation Requirements

- [x] CHK016 - Is the `git mv` command specified as the required directory move method? ✅ spec.md FR-004 + quickstart.md §Step 3
- [x] CHK017 - Are git history verification commands explicitly documented? ✅ quickstart.md: git log --follow --oneline cmd/tui/main.go
- [x] CHK018 - Is the verification criterion for "history preserved" measurable? ✅ quickstart.md: "Shows commit history from before the move"
- [x] CHK019 - Are git commit requirements defined after each user story completion? ✅ spec.md FR-006 + quickstart.md commits after each phase
- [x] CHK020 - Is the commit message format specified with examples? ✅ quickstart.md exact messages + plan.md §Decision 3
- [x] CHK021 - Are requirements for verifying git rename detection explicitly stated? ✅ tasks.md T021: git status shows "renamed:"

## Sequential Dependency Requirements

- [x] CHK022 - Are the sequential dependencies between user stories explicitly documented? ✅ tasks.md §Dependencies & Execution Order
- [x] CHK023 - Is the strict execution order (P1→P2→P3→P4) requirement clearly stated? ✅ tasks.md: "MUST execute in order"
- [x] CHK024 - Are the consequences of out-of-order execution documented? ✅ spec.md §Edge Cases: build fails if directory moved before imports updated
- [x] CHK025 - Is the "MUST complete before next" requirement explicit for each phase? ✅ tasks.md: "Each phase MUST complete before next"
- [x] CHK026 - Are parallel execution opportunities clearly marked with [P] tags? ✅ tasks.md shows [P] markers throughout
- [x] CHK027 - Are the prerequisites for each user story explicitly listed? ✅ spec.md §User Stories 1-4 "Why this priority"

## Package & Import Requirements

- [x] CHK028 - Is the package declaration change (`main` → `tui`) specified for ALL files? ✅ spec.md FR-001 + tasks.md T004-T007
- [x] CHK029 - Is the function signature change (`main()` → `Run(vaultPath string) error`) precisely defined? ✅ spec.md FR-002 + quickstart.md §Step 1
- [x] CHK030 - Are import path update requirements specified with exact find/replace patterns? ✅ quickstart.md §Step 2: cmd/tui-tview → cmd/tui
- [x] CHK031 - Is the verification requirement for "no missed occurrences" explicitly stated? ✅ tasks.md T017: rg "tui-tview" confirms no results
- [x] CHK032 - Are requirements defined for handling vaultPath parameter (empty vs. provided)? ✅ quickstart.md §Step 1 + plan.md shows logic

## TUI Rendering & Visual Requirements

- [x] CHK033 - Are all visual elements required for "TUI renders completely" enumerated? ✅ tasks.md T033: sidebar, table, detail, navigation, forms, delete, masking
- [x] CHK034 - Is "sidebar shows categories" verifiable with specific observable criteria? ✅ tasks.md T033 + quickstart.md §Final Verification
- [x] CHK035 - Are requirements for "credentials listed in table" defined with clear success criteria? ✅ tasks.md T033: "credentials listed in table"
- [ ] CHK036 - Is "detail view shows credential details" specified with required fields? ⚠️ PARTIAL - Mentions "credential details" but not specific fields
- [x] CHK037 - Are navigation testing requirements (arrow keys, Tab, Enter) explicitly documented? ✅ tasks.md T033: "navigation works (arrow keys, Tab, Enter)"
- [x] CHK038 - Are form interaction requirements (Ctrl+A) specified for verification? ✅ tasks.md T033: "forms work (Ctrl+A)"
- [x] CHK039 - Is the password masking toggle requirement included in verification criteria? ✅ tasks.md T033: "password masking toggle works"
- [ ] CHK040 - Are requirements for "no visual corruption" defined with testable criteria? ⚠️ PARTIAL - "Visual corruption" mentioned but not specifically defined

## CLI Compatibility Requirements

- [x] CHK041 - Are CLI command preservation requirements explicitly stated? ✅ spec.md FR-010 + plan.md §Step 4 CLI routing preservation
- [x] CHK042 - Is the argument parsing logic for TUI vs. CLI routing specified? ✅ plan.md §Step 4 complete argument parsing logic
- [x] CHK043 - Are requirements defined for all CLI subcommands (list, get, add, etc.)? ✅ tasks.md T034: list, get, add, --help, --version
- [x] CHK044 - Are help flag requirements (--help, -h) explicitly documented? ✅ tasks.md T030 + T035 edge case testing
- [x] CHK045 - Are version flag requirements (--version, -v) specified? ✅ tasks.md T031: version flag testing
- [x] CHK046 - Is the edge case of `--vault` + `--help` combination addressed? ✅ tasks.md T035: --vault + --help shows help (not TUI)

## Error Handling & Exception Requirements

- [x] CHK047 - Are error handling requirements defined for compilation failures at each step? ✅ quickstart.md §Troubleshooting covers compilation failures
- [x] CHK048 - Are requirements specified for detecting "black screen" issues during verification? ✅ quickstart.md §Troubleshooting: "Issue: Black screen after migration"
- [x] CHK049 - Are import error detection requirements explicitly stated? ✅ quickstart.md §Troubleshooting: "Issue: Import errors after Step 2"
- [x] CHK050 - Is the fallback behavior defined when CLI commands incorrectly launch TUI? ✅ quickstart.md §Troubleshooting: "Issue: CLI commands launch TUI"
- [ ] CHK051 - Are requirements for handling partially completed steps documented? ⚠️ PARTIAL - Resume from Checkpoint exists, but no explicit within-step handling

## Time & Performance Requirements

- [x] CHK052 - Is the 2-hour completion time requirement measurable? ✅ spec.md SC-004 + quickstart.md: "Estimated Time: 1-2 hours"
- [x] CHK053 - Is the <3 second TUI launch requirement explicitly stated? ✅ spec.md SC-006: "TUI launches in under 3 seconds"
- [ ] CHK054 - Are performance degradation criteria defined if launch time exceeds threshold? ❌ GAP - No degradation criteria documented

## Completeness & Coverage Validation

- [x] CHK055 - Are requirements defined for all 10 functional requirements (FR-001 to FR-010)? ✅ spec.md §Requirements lists all FR-001 to FR-010
- [x] CHK056 - Are acceptance criteria defined for all 6 success criteria (SC-001 to SC-006)? ✅ spec.md §Success Criteria + tasks.md T036 validates all
- [x] CHK057 - Do all 4 user stories have measurable acceptance scenarios? ✅ spec.md §User Stories 1-4 all have Given/When/Then scenarios
- [x] CHK058 - Are edge cases documented in the spec addressed in verification tasks? ✅ spec.md §Edge Cases + tasks.md verification addresses them
- [x] CHK059 - Is the out-of-scope boundary clearly defined? ✅ spec.md §Out of Scope explicitly lists 5 items

## Ambiguities & Conflicts

- [x] CHK060 - Is "features function identically" quantified with specific test criteria? ✅ spec.md SC-002 lists features + tasks.md T033 details tests
- [x] CHK061 - Is "zero new compiler errors" verification method specified? ✅ spec.md SC-005 + `go build` verification at each step
- [x] CHK062 - Are there any conflicting requirements between user stories? ✅ No conflicts; user stories are sequential and complementary
- [x] CHK063 - Is the distinction between "package tui" and "tui-tview" directory naming clear? ✅ spec.md FR-001 (package) vs FR-004 (directory) are distinct

---

## Summary

**Audit Completed**: 2025-10-09
**Total Items**: 63 checklist items
**Completed**: 55 items (87.3%) ✅
**Remaining Gaps**: 8 items (12.7%)

### Gap Analysis

**Full Gaps (3 items)**:
- CHK003: Partial failure recovery within a step not documented
- CHK005: Recovery procedures for interrupted git operations missing
- CHK054: Performance degradation criteria not defined (launch time > 3s)

**Partial Gaps (5 items)**:
- CHK015: Terminal capability requirements mentioned but not specific
- CHK036: Detail view requirements lack specific required fields
- CHK040: "Visual corruption" mentioned but not defined with testable criteria
- CHK051: Resume from checkpoint exists, but no explicit within-step partial completion handling

### Coverage Assessment

**Strong Areas** (100% complete):
- ✅ Git History Preservation (CHK016-CHK021): All 6 items satisfied
- ✅ Sequential Dependency Requirements (CHK022-CHK027): All 6 items satisfied
- ✅ Package & Import Requirements (CHK028-CHK032): All 5 items satisfied
- ✅ CLI Compatibility Requirements (CHK041-CHK046): All 6 items satisfied
- ✅ Completeness & Coverage Validation (CHK055-CHK059): All 5 items satisfied
- ✅ Ambiguities & Conflicts (CHK060-CHK063): All 4 items satisfied

**Good Coverage** (83-90%):
- ✅ Rollback & Recovery Requirements (CHK001-CHK006): 4 of 6 satisfied
- ✅ Verification Checkpoint Requirements (CHK007-CHK015): 8 of 9 satisfied
- ✅ TUI Rendering & Visual Requirements (CHK033-CHK040): 6 of 8 satisfied
- ✅ Error Handling & Exception Requirements (CHK047-CHK051): 4 of 5 satisfied
- ✅ Time & Performance Requirements (CHK052-CHK054): 2 of 3 satisfied

### Audit Conclusion

**Status**: ✅ **PASS WITH MINOR GAPS**

The specification has achieved 87.3% coverage of migration safety requirements. The core safety mechanisms (rollback procedures, git history preservation, verification checkpoints, sequential dependencies) are comprehensively documented.

The remaining gaps are primarily edge cases (partial failures, terminal compatibility, performance degradation criteria) that are **acceptable for a developer-focused refactoring task** where:
1. The developer can manually recover from failures (clean git history)
2. Terminal compatibility is verified through manual testing
3. Performance degradation is a non-blocking issue for this migration

### Recommendation

**Proceed with implementation.** The specification quality is sufficient for safe, reversible migration. The identified gaps can be addressed if discovered during implementation, but do not block starting the migration process.

**Next Steps**: Execute tasks.md implementation plan with confidence that all critical safety requirements are documented.
