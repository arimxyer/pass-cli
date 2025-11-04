# Design Improvements - Spec 002 Post-Spec Backlog

**Date:** 2025-11-04
**Last Updated:** 2025-11-04 (Spec 002 Implementation Complete)
**Status:** Complete - No New Improvements Identified
**Purpose:** Document functional/architectural improvements identified during spec 002 clarification that were deferred to maintain scope focus.

---

## Overview

This document captures functional and architectural improvements identified during the spec 002 clarification process. These improvements were explicitly deferred to maintain focus on fixing broken core functionality (keychain enable/status/vault remove). Items here can be promoted to full specs based on user feedback and priority.

---

## 1. Layout & Spacing

### 1.1 [Component/Area Name]
**Current State:**
- [Description]

**Proposed Change:**
- [What should change]

**Rationale:**
- [Why]

**Visual Reference:**
- [Reference]

**Affected Files:**
- [Files]

---

## 2. Typography & Text

### 2.1 [Component/Area Name]
**Current State:**
- [Description]

**Proposed Change:**
- [What should change]

**Rationale:**
- [Why]

**Visual Reference:**
- [Reference]

**Affected Files:**
- [Files]

---

## 3. Colors & Theming

### 3.1 [Component/Area Name]
**Current State:**
- [Description]

**Proposed Change:**
- [What should change]

**Rationale:**
- [Why]

**Visual Reference:**
- [Reference]

**Affected Files:**
- [Files]

---

## 4. Interactive Elements (Buttons, Inputs, etc.)

### 4.1 [Component/Area Name]
**Current State:**
- [Description]

**Proposed Change:**
- [What should change]

**Rationale:**
- [Why]

**Visual Reference:**
- [Reference]

**Affected Files:**
- [Files]

---

## 5. Component-Specific Improvements

### 5.1 [Component Name]
**Current State:**
- [Description]

**Proposed Change:**
- [What should change]

**Rationale:**
- [Why]

**Visual Reference:**
- [Reference]

**Affected Files:**
- [Files]

---

## 6. Responsive & Accessibility

### 6.1 [Issue/Area Name]
**Current State:**
- [Description]

**Proposed Change:**
- [What should change]

**Rationale:**
- [Why]

**Visual Reference:**
- [Reference]

**Affected Files:**
- [Files]

---

## 7. Other Improvements

### 7.1 [Improvement Name]
**Current State:**
- [Description]

**Proposed Change:**
- [What should change]

**Rationale:**
- [Why]

**Affected Files:**
- [Files]

---

## 8. Functional & Architectural Improvements

### 8.1 Vault Concurrency Safety

**Current State:**
- pass-cli assumes single-process vault access (single-vault architecture)
- No detection for concurrent operations:
  - TUI running while `vault remove` executed
  - Multiple CLI commands running simultaneously
  - Vault file operations during reads
- Users could experience unexpected behavior or data corruption if multiple processes access vault

**Proposed Change:**
Implement lock file mechanism to prevent concurrent vault access:
- Create `.vault.lock` file during vault operations
- Commands check for lock file before proceeding
- Handle stale locks (timeout, PID validation)
- Cleanup on process exit/crash
- Clear error messages when operations blocked by locks

**Rationale:**
- Prevents data corruption from concurrent writes
- Prevents unexpected behavior (e.g., vault removed while TUI running)
- Provides better error messages for users
- Aligns with expectations for production-grade CLI tools

**Affected Files:**
- `internal/vault/vault.go` (lock acquisition/release)
- `cmd/vault_remove.go` (check lock before removal)
- `cmd/tui/main.go` (acquire lock on vault open)
- All CLI commands that access vault (get/list/add/delete/etc.)
- New file: `internal/vault/lock.go` (lock management)

**Impact:** High (prevents data corruption scenarios)

**Effort:** Medium

**Priority:** Unassessed (waiting for user feedback)

**Dependencies:** None

**Alternative Solutions Considered:**
1. OS-level file locking (syscall.Flock) - Platform-specific, automatic cleanup
2. Advisory warnings - Detect processes but allow override with --force (less safe)

**Known Challenges:**
- Lock file cleanup after crashes
- Cross-platform file locking differences
- Performance impact of lock checks
- User experience with lock conflicts

---

### 8.2 Security: File Permissions Hardening

**Current State:**
- gosec security scan identified 10 issues (mostly MEDIUM severity):
  - Config files written with 0644 permissions (should be 0600)
  - Config directories created with 0755 permissions (should be 0750)
  - Integer overflow risk in keybinding conversion
  - Expected file inclusion patterns (vault/config operations)

**Proposed Change:**
- Reduce config file permissions to 0600 (owner-only read/write)
- Reduce config directory permissions to 0750 (owner full, group read/execute)
- Add bounds checking for integer conversion in keybinding parser
- Document G204 (subprocess) and G304 (file inclusion) as expected patterns

**Rationale:**
- Defense-in-depth: Config files may contain sensitive vault paths
- Reduces attack surface for multi-user systems
- Aligns with security best practices for credential management tools
- Integer overflow could cause undefined behavior in key handling

**Affected Files:**
- `cmd/config.go` (lines 105, 125, 216, 226): Change WriteFile permissions 0644 → 0600
- `internal/config/config.go` (line 103): Change MkdirAll permissions 0755 → 0750
- `internal/storage/storage.go` (line 384): Change MkdirAll permissions 0755 → 0750
- `internal/config/keybinding.go` (line 136): Add bounds check before int → int16 conversion

**Impact:** Low (security improvement, no functional changes)

**Effort:** Small (4-file change)

**Priority:** Low (pre-existing technical debt, no active exploits)

**Dependencies:** None

**Note:** gosec findings reviewed during spec 002 Phase 6 (T044). Issues are pre-existing technical debt, not introduced by spec 002 changes.

---

## Spec 002 Implementation Review

**Completion Date:** 2025-11-04

**Work Completed:**
- ✅ Phase 1-5: All user stories (US-001 through US-005)
- ✅ Phase 6: Quality gates (CI, docs, linting, security scan)
- ✅ 25 integration tests unskipped and passing
- ✅ Metadata system implemented
- ✅ Keychain enable/status commands implemented
- ✅ Vault remove command implemented with cleanup
- ✅ Documentation updated (GETTING_STARTED.md, README.md)
- ✅ CI gate added for TODO-skipped tests

**New Improvements Identified:** 1 (Security: File Permissions Hardening - see 8.2 above)

**Deferred Items:** 1 (Vault Concurrency Safety - already documented as 8.1)

**Recommendation:**
- Vault Concurrency Safety (8.1): Defer to future spec based on user feedback
- File Permissions Hardening (8.2): Low priority, can be addressed as technical debt cleanup

---

## Implementation Notes

- All changes should maintain existing functionality
- Changes should be tested, and all existing tests should continue to pass.
- Color changes should maintain sufficient contrast for accessibility

---

## Success Criteria

- [ ] All visual improvements implemented as specified
- [ ] No regressions in existing functionality
- [ ] All tests passing
- [ ] Visual consistency across all components
- [ ] Accessibility standards maintained

---

## Next Steps [IF PRE-SPEC]

1. Complete this requirements document
2. Run `/speckit.specify` to create formal spec
3. Run `/speckit.clarify` if needed
4. Run `/speckit.plan` for implementation strategy
5. Run `/speckit.tasks` to break into actionable tasks
6. Implement systematically with testing

## NEXT STEPS [IF POST-SPEC]

1. Complete this requirements document with all identified improvements
2. Analyze and assess the improvements:
   - Categorize by impact (high/medium/low)
   - Categorize by effort (small/medium/large)
   - Identify any breaking changes or risks
3. Put forth up to 3 different action plans with your recommendation:
   - **Option A:** Implement all improvements immediately
   - **Option B:** Prioritize high-impact, low-effort improvements first
   - **Option C:** Bundle improvements into a follow-up spec
   - [Your recommendation with rationale]
4. Based on chosen action plan:
   - **If implementing directly:** Create task list and proceed systematically
   - **If creating follow-up spec:** Run `/speckit.specify` with this document as input
   - **If deferring:** Archive this document for future reference
5. Implement chosen improvements with testing
6. Verify no regressions in existing functionality
7. Update any related documentation
