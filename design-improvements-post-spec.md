# Design Improvements - Spec 002 Post-Spec Backlog

**Date:** 2025-11-04
**Status:** Active - Collecting Improvements
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
