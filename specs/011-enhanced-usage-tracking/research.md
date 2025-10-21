# Research: Enhanced Usage Tracking CLI

**Feature**: Enhanced Usage Tracking CLI
**Date**: 2025-10-20
**Status**: Complete (all decisions made via clarification session)

## Overview

This feature exposes existing usage tracking data through CLI commands. All technical decisions were resolved during the specification clarification phase (Questions 1-5 in spec.md). No additional research required - existing infrastructure handles data collection and storage.

---

## Technical Decision 1: Large Output Handling

**Question**: How should `usage` command handle credentials with 50+ usage locations?

**Decision**: Default to 20 most recent locations, add `--limit N` flag (0 = unlimited)

**Rationale**:
- Sensible defaults prevent overwhelming users
- `--limit` flag provides customization for power users
- Common CLI pattern (similar to `git log`, `docker ps`)
- Performance: Default truncation keeps response times predictable

**Alternatives Considered**:
- Display all without limit: Risk poor UX for rare edge case
- Fixed limit with no flag: Inflexible, forces users to pipe to `head`
- Pagination: Too complex for read-only display

**Implementation**:
- FR-004a: `usage` command supports `--limit N` flag
- FR-004b: Display "... and N more locations (use --limit 0 to see all)" when truncated
- Default: 20 locations (recent-first sort)

**Source**: Clarification Q1, Gemini consensus

---

## Technical Decision 2: Flag Composition Behavior

**Question**: What happens when `--by-project` and `--location` are specified together?

**Decision**: Combine intelligently - filter by location first, then group results by project (orthogonal operations)

**Rationale**:
- `--location`: Filter (which credentials) - WHERE clause
- `--by-project`: Display mode (how to show them) - GROUP BY clause
- Orthogonal operations compose naturally
- Unix philosophy: combining filters is expected behavior
- Real use case: "Show credentials from THIS directory, organized by project"

**Alternatives Considered**:
- Mutual exclusion (error): Too restrictive, forces artificial choice
- Last flag wins: Surprising, violates least astonishment principle
- Sequential application: Confusing, `--by-project` isn't a filter

**Implementation**:
- FR-012a: Apply location filter first, then group by project
- Example: `list --location /home/user/work --by-project --recursive`

**Source**: Clarification Q2, Gemini consensus

---

## Technical Decision 3: Deleted Path Display Strategy

**Question**: How to display usage locations when directory path no longer exists (deleted/moved)?

**Decision**: Format-specific behavior - table hides deleted paths (clean UX), JSON includes all paths with `path_exists` field (complete data)

**Rationale**:
- **Table format** (human users): Hide deleted paths for clean, focused output on current usage
- **JSON format** (scripts): Include all paths with `"path_exists": false` for complete historical data
- Leverages existing `--format` distinction (already treats table vs. JSON differently per FR-016/FR-017)
- No feature creep: Uses existing format infrastructure
- Respects read-only design: Preserves all data without modification

**Alternatives Considered**:
- Display all with markers: Clutters interactive output
- Hide all: Scripts lose historical data
- Add --include-deleted flag: Unnecessary complexity

**Implementation**:
- FR-018: Table format hides non-existent paths
- FR-019: JSON format includes `path_exists` boolean field
- Helper function: `pathExists(path string) bool` for filesystem check at display time

**Source**: Clarification Q3, Gemini consensus

---

## Technical Decision 4: Cross-Machine Path Handling

**Question**: How to handle credentials accessed from network/cloud-synced paths with different absolute paths across machines?

**Decision**: No special handling - different machines are different usage contexts (expected behavior)

**Rationale**:
- Different machines ARE different contexts (accurate representation)
- Git repositories already unified via `--by-project` flag (machine-independent view)
- Cannot reliably normalize non-git projects
- Consistent with existing assumption: "Absolute paths are platform-specific - display as-is"
- Zero complexity: No normalization heuristics needed

**Alternatives Considered**:
- Path normalization: Complex, brittle, many false positives
- Git repo + relative path matching: Doesn't work for non-git projects

**Implementation**:
- Display all paths exactly as recorded
- Users wanting machine-independent view use `list --by-project` (groups by repo name)
- Example: Windows `C:\Users\Alice\Dropbox\project` and Linux `/home/bob/Dropbox/project` appear separately (accurate - different machines, different users)

**Source**: Clarification Q4, Gemini consensus

---

## Technical Decision 5: Renamed Repository Handling

**Question**: How should `--by-project` behave when git repository was renamed?

**Decision**: Use historical git repo name as recorded at access time (immutable historical record)

**Rationale**:
- Immutable historical record (append-only log philosophy)
- No filesystem I/O needed (no git state checking)
- Handles deleted paths (can't check current git config for deleted directories)
- Performance: Zero overhead for historical lookups
- Accurate: Shows what repo was called when credential was accessed

**Alternatives Considered**:
- Use current repo name: Requires filesystem access, breaks for deleted paths
- Group by git remote URL: More stable but slower, requires git operations

**Trade-off Accepted**: Same project may appear under multiple names if renamed. This is accurate historical timeline (shows project evolution), not a bug.

**Implementation**:
- FR-008a: `list --by-project` uses historical GitRepository field value
- No git state checking during display
- Example: "my-old-project" and "my-new-project" appear as separate groups if repo renamed

**Source**: Clarification Q5, Gemini consensus

---

## Existing Infrastructure Analysis

### UsageRecord Data Structure

**Location**: `internal/vault/vault.go:29-36`

**Fields**:
```go
type UsageRecord struct {
    Location      string         // Absolute path where accessed
    GitRepository string         // Git repo name at access time
    FieldCounts   map[string]int // Field-level access (password:5, username:2)
    Timestamp     time.Time      // Last access time
    AccessCount   int            // Total access count
}
```

**Storage**: Per-credential map in `Credential.UsageRecords` field (line 50)

**Collection**: Automatic during `get`, `update` operations in VaultService

**Verdict**: ✅ All required data already collected and persisted

### TUI Display Reference

**Location**: `cmd/tui/components/detail.go:340-395`

**Functionality**:
- Formats UsageRecord data for visual display
- Shows location, repo name, field counts, timestamps
- Human-readable relative time formatting

**Verdict**: ✅ Can adapt this logic for CLI table output

### Existing Formatting Patterns

**JSON Output**: `cmd/list.go` already supports `--format json`
**Table Output**: `cmd/list.go` uses tabwriter for aligned columns
**Simple Output**: `cmd/list.go` supports `--format simple` (newline-separated)

**Verdict**: ✅ Reuse existing formatting infrastructure

---

## Technology Stack Confirmation

**Language**: Go 1.21+ (existing codebase)
**Dependencies**: Standard library only (no new external dependencies)
**Testing**: Go testing framework (`go test`)
**Build**: Existing `go build` infrastructure
**Platforms**: Windows, macOS, Linux (no platform-specific code needed)

**Verdict**: ✅ No new technologies, all infrastructure exists

---

## Performance Analysis

**SC-001**: Under 3 seconds for 50 locations
- **Data Source**: In-memory (vault already loaded)
- **Operations**: Iteration + string formatting
- **I/O**: Zero (except optional `pathExists` filesystem check)
- **Verdict**: ✅ Easily achievable

**SC-003/SC-004**: Under 2 seconds for 100 credentials
- **Data Source**: In-memory
- **Operations**: Filtering (string matching), grouping (map operations)
- **I/O**: Zero
- **Verdict**: ✅ Easily achievable

---

## Best Practices Applied

### CLI Design Patterns (from research + existing codebase)

1. **Output Formatting**:
   - Default: Human-readable table
   - `--format json`: Machine-readable (idempotent, parseable)
   - `--format simple`: Newline-separated (for piping to other tools)

2. **Flag Naming**:
   - `--by-project`: Descriptive action
   - `--location <path>`: Requires argument
   - `--recursive`: Boolean modifier
   - `--limit N`: Numeric parameter

3. **Error Handling**:
   - Non-existent credential: Clear error message to stderr, exit code 1
   - Empty results: Informative message (not an error)
   - Invalid flags: Usage help to stderr, exit code 1

4. **Script-Friendly Design**:
   - JSON output is well-formed (no pretty-printing required)
   - No ANSI colors in non-TTY output
   - Deterministic ordering (sort by timestamp, repo name)

### Testing Patterns (from Constitution Principle IV)

1. **Unit Tests**: Format helpers, filtering logic
2. **Integration Tests**: Full command execution with real vault files
3. **Contract Tests**: JSON schema stability
4. **Edge Case Tests**: Empty data, deleted paths, large datasets

---

## Risk Assessment

### Security Risks

**Risk Level**: LOW

- ✅ Read-only feature (no credential modification)
- ✅ No credential data displayed (only metadata)
- ✅ Existing vault locking/unlocking preserved
- ✅ No new encryption/decryption code

### Compatibility Risks

**Risk Level**: LOW

- ✅ New command (`usage`) - no conflict potential
- ✅ New flags additive to `list` - backward compatible
- ✅ No changes to vault file format
- ✅ No changes to existing command behavior

### Performance Risks

**Risk Level**: LOW

- ✅ In-memory operations only
- ✅ Optional filesystem checks (only for deleted path detection)
- ✅ Success criteria easily met (under 2-3 seconds)

---

## Dependencies & Integrations

### Internal Dependencies

- `internal/vault/vault.go`: VaultService, UsageRecord struct
- `cmd/tui/components/detail.go`: Reference formatting logic
- `cmd/list.go`: Existing command to extend with new flags
- `cmd/helpers.go`: Common CLI utilities (likely readPassword, formatOutput)

### External Dependencies

**None** - Standard library only:
- `fmt`: Output formatting
- `os`: Filesystem checks (pathExists)
- `time`: Timestamp formatting
- `encoding/json`: JSON marshaling
- `github.com/spf13/cobra`: CLI framework (already in use)

---

## Summary

**Research Status**: ✅ COMPLETE

**Key Findings**:
1. All infrastructure exists (UsageRecord, TUI display, formatting patterns)
2. No new dependencies needed (stdlib only)
3. All design decisions made via clarification (Q1-Q5)
4. Constitution compliance verified (all gates pass)
5. Performance targets easily achievable (in-memory operations)

**Next Phase**: Generate data-model.md, contracts/commands.md, quickstart.md

**References**:
- Clarifications: spec.md lines 8-16 (Session 2025-10-20)
- Constitution: .specify/memory/constitution.md
- Existing code: internal/vault/vault.go:29-36, 50; cmd/tui/components/detail.go:340-395
