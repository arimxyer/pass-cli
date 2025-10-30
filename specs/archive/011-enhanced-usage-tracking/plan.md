# Implementation Plan: Enhanced Usage Tracking CLI

**Branch**: `011-enhanced-usage-tracking` | **Date**: 2025-10-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/011-enhanced-usage-tracking/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Expose existing multi-location usage tracking through CLI commands to enable single-vault organization by context. Add `usage` command to show detailed credential usage across locations (similar to TUI detail view), and extend `list` command with `--by-project` and `--location` flags. Infrastructure already exists (UsageRecord struct in `internal/vault/vault.go`, TUI display in `cmd/tui/components/detail.go`). This feature provides CLI access to rich usage data for script-friendly automation and better visibility into pass-cli's context-based organization model.

## Technical Context

**Language/Version**: Go 1.21+ (existing codebase)
**Primary Dependencies**: Standard library only (no new external dependencies)
**Storage**: Existing vault file format (`vault.enc`), UsageRecord already persisted in vault data structure
**Testing**: Go testing framework (`go test`), existing test infrastructure in `test/` and `internal/*/`
**Target Platform**: Cross-platform (Windows, macOS, Linux) - consistent with existing pass-cli architecture
**Project Type**: Single CLI binary project with library-first architecture
**Performance Goals**: Display usage data in under 3 seconds for 50+ locations (SC-001), list operations in under 2 seconds for 100 credentials (SC-003, SC-004)
**Constraints**:
- Read-only feature (no modification of usage history per spec Out of Scope)
- Must maintain backward compatibility with existing `list` command
- Must work in CLI-only environments (no TUI dependencies)
- Format-specific behavior (table vs. JSON output)
**Scale/Scope**:
- 3 user stories (independently testable)
- 1 new command (`usage`), 2 new flags on existing command (`list --by-project`, `list --location`)
- 23 functional requirements total
- Existing infrastructure handles data collection, feature only exposes via CLI

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Security-First Development (NON-NEGOTIABLE)
- ✅ **PASS**: Feature is read-only display of metadata (locations, timestamps, counts)
- ✅ **PASS**: No credential data displayed (only service names, paths, repo names, access metadata)
- ✅ **PASS**: No new credential handling, storage, or encryption code
- ✅ **PASS**: Existing vault locking/unlocking behavior preserved (FR-015)
- **Risk Level**: LOW - Pure read/display feature, no security-critical operations

### Principle II: Library-First Architecture
- ✅ **PASS**: Existing library functions in `internal/vault/vault.go` handle data access
- ✅ **PASS**: CLI commands will be thin wrappers calling vault service methods
- ✅ **PASS**: No new libraries needed (leverages existing VaultService, UsageRecord)
- **Approach**: New CLI commands (`cmd/usage.go`, updates to `cmd/list.go`) delegate to existing `internal/vault` package

### Principle III: CLI Interface Standards
- ✅ **PASS**: Commands follow existing pass-cli patterns (flags, output formats)
- ✅ **PASS**: Script-friendly JSON output supported (FR-013, all commands support `--format json`)
- ✅ **PASS**: Errors to stderr with non-zero exit codes (FR-006)
- ✅ **PASS**: No interactive prompts (read-only display)
- **Implementation**: Consistent with existing `list`, `get`, `verify-audit` commands

### Principle IV: Test-Driven Development (NON-NEGOTIABLE)
- ✅ **PASS**: Spec includes 10 acceptance scenarios for User Story 1 (P1 MVP)
- ✅ **PASS**: Spec includes 5 acceptance scenarios for User Story 2 (P2)
- ✅ **PASS**: Spec includes 5 acceptance scenarios for User Story 3 (P3)
- ✅ **PASS**: 20 total acceptance scenarios to drive TDD
- ✅ **PASS**: Success criteria are measurable (SC-001 through SC-006)
- **Commitment**: Tests will be written first for each user story, implementation follows

### Principle V: Cross-Platform Compatibility
- ✅ **PASS**: No platform-specific code needed (pure Go stdlib)
- ✅ **PASS**: Path handling already cross-platform in existing codebase
- ✅ **PASS**: Clarification Q3 confirms: "Display paths as-is without normalization" (per existing assumptions)
- **Platform Notes**: Absolute paths are platform-specific (Windows vs. Unix) - feature displays them as-is

### Principle VI: Observability & Auditability
- ✅ **PASS**: Feature enhances observability (exposes usage tracking data via CLI)
- ✅ **PASS**: No sensitive data displayed (only metadata: locations, timestamps, counts)
- ✅ **PASS**: Read-only operations don't require additional audit logging
- **Value**: This feature IS the observability improvement (makes usage data accessible for analysis)

### Principle VII: Simplicity & YAGNI
- ✅ **PASS**: No new data structures (reuses existing UsageRecord)
- ✅ **PASS**: No new dependencies (stdlib only)
- ✅ **PASS**: Leverages existing formatting patterns from TUI (`cmd/tui/components/detail.go`)
- ✅ **PASS**: Clarifications resolved to avoid complexity:
  - Q1: Simple `--limit N` flag (no pagination complexity)
  - Q2: Flags compose naturally (no complex precedence rules)
  - Q3: Format-specific behavior (no new display modes)
  - Q4: No path normalization (display as-is)
  - Q5: No git state checking (use historical data)
- **Complexity Score**: LOW - Pure display feature with existing infrastructure

**Constitution Verdict**: ✅ **ALL GATES PASS** - No violations, no complexity justifications needed

## Project Structure

### Documentation (this feature)

```
specs/011-enhanced-usage-tracking/
├── spec.md              # Feature specification (complete)
├── checklists/
│   └── requirements.md  # Quality validation checklist (complete)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (to be generated)
├── data-model.md        # Phase 1 output (to be generated)
├── quickstart.md        # Phase 1 output (to be generated)
├── contracts/           # Phase 1 output (to be generated)
│   └── commands.md      # CLI command contracts
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

**Existing Structure (No Changes)**:
```
pass-cli/
├── cmd/                      # CLI commands (thin wrappers)
│   ├── usage.go              # NEW: usage command
│   ├── list.go               # MODIFY: add --by-project, --location flags
│   └── [other commands]
├── internal/                 # Library packages
│   ├── vault/
│   │   ├── vault.go          # EXISTING: VaultService with UsageRecord (lines 29-36, 50)
│   │   └── [other files]
│   └── [other packages]
├── test/                     # Integration tests
│   ├── usage_test.go         # NEW: integration tests for usage command
│   ├── list_test.go          # MODIFY: tests for new list flags
│   └── [other tests]
└── [other dirs]
```

**Structure Decision**: Single CLI binary project following existing pass-cli architecture. No structural changes needed - feature adds one new command file (`cmd/usage.go`) and modifies existing `cmd/list.go`. All business logic already exists in `internal/vault/vault.go` (VaultService methods, UsageRecord struct). Tests follow existing pattern in `test/` directory.

## Complexity Tracking

*No complexity justifications needed - all Constitution checks passed.*

## Phase 0: Research & Technical Decisions

**Purpose**: Resolve unknowns from Technical Context and make key implementation decisions.

### Research Tasks

**No research tasks needed** - Technical context is fully specified:

1. ✅ **Existing Infrastructure Confirmed**:
   - UsageRecord struct exists at `internal/vault/vault.go:29-36`
   - TUI display logic exists at `cmd/tui/components/detail.go:340-395`
   - Vault service methods exist for data access

2. ✅ **Technology Stack Confirmed**:
   - Go 1.21+ (existing)
   - Standard library only (no new dependencies)
   - Existing test infrastructure (`go test`)

3. ✅ **Design Decisions Made via Clarifications**:
   - Decision 1 (Q1): Default output limit of 20 locations, `--limit N` flag
   - Decision 2 (Q2): Flags compose naturally (filter + display mode)
   - Decision 3 (Q3): Format-specific behavior (table hides deleted, JSON shows all)
   - Decision 4 (Q4): No path normalization (display as-is, expected behavior)
   - Decision 5 (Q5): Use historical git repo names (immutable record)

### Research Output: research.md

Since no unknowns exist and all design decisions were made during clarification phase, `research.md` will document the pre-existing decisions and reference existing codebase infrastructure.

**Generated**: See `research.md` for consolidated design decisions from clarification session.

## Phase 1: Design Artifacts

**Purpose**: Generate data model, API contracts, and quickstart guide.

### 1. Data Model (data-model.md)

**Entities**:
- **UsageRecord** (existing): Already defined in `internal/vault/vault.go:29-36`, contains:
  - Location (string): Absolute path where credential accessed
  - GitRepository (string): Git repo name at time of access
  - FieldCounts (map[string]int): Field-level access counts (password, username, etc.)
  - Timestamp (time.Time): Last access time
  - AccessCount (int): Total access count

**No new entities needed** - feature exposes existing data structure via CLI.

**Generated**: See `data-model.md` for UsageRecord structure and field definitions.

### 2. API Contracts (contracts/)

**CLI Command Contracts**:
- `usage <service>` command signature, flags, output formats
- `list --by-project` flag behavior, output structure
- `list --location <path>` flag behavior, output structure
- JSON output schemas for all commands

**Generated**: See `contracts/commands.md` for command specifications.

### 3. Quickstart Guide (quickstart.md)

**User-facing examples**:
- View detailed usage: `pass-cli usage github`
- Group by project: `pass-cli list --by-project`
- Filter by location: `pass-cli list --location /home/user/work --recursive`
- JSON output: `pass-cli usage github --format json | jq`
- Limit output: `pass-cli usage aws --limit 10`

**Generated**: See `quickstart.md` for command examples and expected outputs.

### 4. Agent Context Update

**Action**: Run `.specify/scripts/powershell/update-agent-context.ps1 -AgentType claude`

**Updates**: No new technologies to add - using existing Go stdlib and test infrastructure.

## Phase 2: Task Breakdown (Not Generated by /speckit.plan)

**Note**: Phase 2 (tasks.md generation) is handled by the `/speckit.tasks` command, not `/speckit.plan`.

**Next Step After This Plan**: Run `/speckit.tasks` to generate detailed task breakdown from:
- 3 user stories (P1, P2, P3)
- 20 acceptance scenarios
- 23 functional requirements
- TDD approach (tests first, implementation second)

## Architecture Decisions

### Decision 1: Command Implementation Pattern

**Choice**: Follow existing pass-cli command pattern (thin CLI wrapper → library call)

**Rationale**:
- Consistent with Principle II (Library-First Architecture)
- Reference: `cmd/list.go`, `cmd/get.go`, `cmd/verify_audit.go`
- All business logic stays in `internal/vault/vault.go`

**Example Pattern**:
```go
// cmd/usage.go
func runUsage(cmd *cobra.Command, args []string) error {
    vaultService := vault.New(vaultPath)
    usageData := vaultService.GetCredentialUsage(serviceName)
    formatOutput(usageData, formatFlag)
}
```

### Decision 2: Output Formatting Strategy

**Choice**: Reuse existing formatting patterns from TUI and list command

**Rationale**:
- TUI already displays usage data beautifully (`cmd/tui/components/detail.go:340-395`)
- `list` command already has table/JSON/simple formatters
- Leverage existing `--format` flag infrastructure

**Reference**:
- `cmd/list.go`: Table formatting, JSON marshaling
- `cmd/tui/components/detail.go`: Usage display logic (copy relevant functions)

### Decision 3: Path Existence Checking (Clarification Q3)

**Choice**: Check filesystem only during display, not during data storage

**Rationale**:
- Read-only feature (no modification of UsageRecord)
- Performance: Single filesystem check at display time vs. continuous polling
- Format-specific: Table hides deleted (better UX), JSON shows all (complete data)

**Implementation**: Add helper function `pathExists(path string) bool` called during output formatting.

### Decision 4: Flag Composition (Clarification Q2)

**Choice**: `--by-project` and `--location` are orthogonal (filter + display mode)

**Rationale**:
- `--location`: Filters which credentials (WHERE clause)
- `--by-project`: Changes display grouping (GROUP BY clause)
- Natural SQL-like composition: "SELECT credentials WHERE location LIKE '/path%' GROUP BY git_repo"

**Implementation**: Apply location filter first, then group remaining results by project.

### Decision 5: Historical Data Integrity (Clarifications Q4, Q5)

**Choice**: Display usage data exactly as recorded (no normalization, no git state checking)

**Rationale**:
- Immutable historical record (append-only log philosophy)
- No filesystem I/O for deleted paths or renamed repos
- Performance: Zero overhead for historical lookups
- Accuracy: Shows what happened when it happened

**Trade-off Accepted**: Same credential may appear under old/new repo names if renamed. This is accurate historical timeline, not a bug.

## Success Criteria Validation

**SC-001**: Users can view detailed credential usage history in under 3 seconds for credentials with up to 50 usage locations
- **Implementation**: No I/O beyond vault load (data already in memory), simple iteration + formatting

**SC-002**: 100% of usage data displayed in TUI detail view is accessible via `pass-cli usage` command
- **Implementation**: Copy display logic from `cmd/tui/components/detail.go:340-395`, adapt for CLI table format

**SC-003**: Users can successfully filter credentials by project context using `list --by-project` with results displayed in under 2 seconds for vaults with up to 100 credentials
- **Implementation**: In-memory grouping by GitRepository field, no filesystem I/O

**SC-004**: Users can successfully filter credentials by location path using `list --location` with results displayed in under 2 seconds
- **Implementation**: In-memory string matching on Location field, optional `--recursive` for path prefix matching

**SC-005**: All new commands support script-friendly JSON output that can be parsed by standard JSON tools (jq, Python json module, etc.)
- **Implementation**: Reuse existing JSON marshaling from `list` command, ensure well-formed schemas

**SC-006**: Users can discover project-specific credentials without prior knowledge of credential names (via `list --by-project` or `list --location`)
- **Implementation**: Display grouped/filtered output even when user doesn't know specific service names

## Next Steps

1. ✅ **Phase 0 Complete**: Generate `research.md` documenting design decisions from clarification session
2. ✅ **Phase 1 Complete**: Generate `data-model.md`, `contracts/commands.md`, `quickstart.md`
3. ⏳ **Phase 2 Pending**: Run `/speckit.tasks` to generate TDD task breakdown
4. ⏳ **Implementation**: Follow tasks.md with test-first approach per Constitution Principle IV

---

**Plan Status**: ✅ READY FOR TASK GENERATION
**Constitution Compliance**: ✅ ALL GATES PASSED
**Research Required**: ✅ NONE (all decisions made via clarifications)
**Next Command**: `/speckit.tasks`
