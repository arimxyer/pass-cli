# Data Model: Documentation Update

**Branch**: `004-we-ve-recently` | **Date**: 2025-10-11 | **Phase**: 1 (Design)

## Overview

This feature is a documentation-only task. No data structures are being created or modified. This document serves as a reference for documentation entities.

## Entities

### DocumentationFile

Represents a documentation file requiring updates.

**Attributes:**
- `path` (string) - Absolute file path (e.g., `R:\Test-Projects\pass-cli\README.md`)
- `type` (enum) - One of: `user-guide`, `api-reference`, `quickstart`
- `scope` (enum) - One of: `user-facing`, `contributor`, `internal`
- `status` (enum) - One of: `needs-update`, `in-progress`, `updated`, `verified`

**Target Files:**
- `README.md` (user-guide, user-facing)
- `docs/USAGE.md` (api-reference, user-facing)

### FeatureDocumentation

Represents a feature that needs to be documented.

**Attributes:**
- `feature_name` (string) - Human-readable feature name
- `source_spec` (string) - Spec ID where feature was implemented (e.g., "002-hey-i-d")
- `category` (enum) - One of: `tui-mode`, `keyboard-shortcut`, `interactive-feature`, `ui-control`
- `priority` (enum) - One of: `p1-critical`, `p2-high`, `p3-medium`

**Instances to Document:**
1. TUI Mode Launch (`001-reorganize-cmd-tui`, `tui-mode`, `p1-critical`)
2. Ctrl+H Password Toggle (`003-would-it-be`, `keyboard-shortcut`, `p1-critical`)
3. Sidebar Toggle (`002-hey-i-d`, `ui-control`, `p2-high`)
4. Search Mode (`002-hey-i-d`, `interactive-feature`, `p1-critical`)
5. Usage Location Display (`002-hey-i-d`, `interactive-feature`, `p3-medium`)

### KeyboardShortcut

Represents a keyboard shortcut to be documented.

**Attributes:**
- `key_combination` (string) - Key combination (e.g., "Ctrl+H", "s", "/")
- `action` (string) - What the shortcut does
- `context` (string) - Where it works (e.g., "in add/edit forms", "main TUI view")
- `availability` (enum) - One of: `global`, `tui-only`, `forms-only`, `modal-only`

**Complete List:**
- See research.md "Complete TUI Keyboard Shortcuts" section for full inventory

## Validation Rules

### Documentation Accuracy
- All documented features MUST exist in current implementation
- All documented file paths MUST exist on filesystem
- All documented keyboard shortcuts MUST be verified in code
- No speculative/planned features may be documented

### Completeness
- All features from specs 001-003 MUST be documented
- All keyboard shortcuts MUST include action + context
- All TUI modes MUST explain how to enter and exit

## State Transitions

### DocumentationFile Status Flow
```
needs-update → in-progress → updated → verified
```

- `needs-update`: File identified as outdated
- `in-progress`: Currently being edited
- `updated`: Changes written to file
- `verified`: Changes tested by following documented steps

## Non-Functional Constraints

- Documentation must use consistent formatting with existing docs
- Keyboard shortcuts must use format: `Key - Action (Context)`
- Examples must be copy-pasteable (no placeholder values)
- Cross-references between README.md and docs/USAGE.md must stay synchronized
