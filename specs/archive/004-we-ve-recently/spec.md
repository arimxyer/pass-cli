# Feature Specification: Documentation Update for Recent Application Changes

**Feature Branch**: `004-we-ve-recently`
**Created**: 2025-10-11
**Status**: Draft
**Input**: User description: "we've recently made a bunch of changes to the application, but it's been quite a while since we've updated the relevant documentation"

## Clarifications

### Session 2025-10-11

- Q: What level of detail should the updated keyboard shortcut documentation include? → A: Add context: shortcut + action + where it works (e.g., "Ctrl+H - Toggle password visibility (in add/edit forms)")
- Q: Should internal development workflow (speckit commands, CLAUDE.md) be updated? → A: No - scope is user-facing documentation only (README, docs/ files users see on GitHub). CLAUDE.md stays as-is. Testing guidelines should focus on user-facing testing documentation, not contributor workflow.
- Q: Should documentation include how to launch and use the TUI interactive mode? → A: Yes
- Q: Should docs/ARCHITECTURE.md be created for TUI internal architecture? → A: No - TUI internal architecture (components/, events/, layout/, models/, styles/) is contributor documentation, not user-facing. Out of scope.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Update Interactive TUI Features Documentation (Priority: P1)

Users need documentation of new interactive TUI features (password visibility toggle, keyboard shortcuts, form interactions) added in recent specs to understand all available functionality and maximize their productivity.

**Why this priority**: Feature documentation directly impacts user experience. Users won't discover or use features they don't know exist, reducing the application's value.

**Independent Test**: Can be fully tested by verifying each documented feature (password toggle via Ctrl+H, search navigation, statusbar shortcuts) exists in the application and behaves as documented. Delivers value by helping users discover and use new capabilities.

**Acceptance Scenarios**:

1. **Given** a user reading the documentation, **When** they look for password management features, **Then** they find documentation of the password visibility toggle (Ctrl+H) in add/edit forms
2. **Given** a user looking for keyboard shortcuts, **When** they review the documentation, **Then** they see an updated list including all recent additions (Ctrl+H, search navigation shortcuts, statusbar shortcuts)
3. **Given** a user trying to use a documented feature, **When** they follow the documentation instructions, **Then** the feature works exactly as described

---


### Edge Cases

- What happens when documentation references files that were moved or deleted? (Need to audit all file paths and ensure they're current)
- What happens when users follow outdated installation instructions? (Verify all commands still work with current dependencies)
- What happens when documentation contradicts actual behavior? (Actual behavior takes precedence; document as-implemented)
- What happens when features are documented but not yet implemented? (Remove or clearly mark as "planned" to avoid confusion)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Documentation MUST include instructions for launching and using the TUI interactive mode
- **FR-002**: Documentation MUST include all interactive features added in specs 001-003 (password visibility toggle, search navigation, statusbar shortcuts)
- **FR-003**: Documentation MUST include accurate keyboard shortcuts with their current key bindings in format: shortcut + action + context (e.g., "Ctrl+H - Toggle password visibility (in add/edit forms)")
- **FR-004**: Documentation MUST update any outdated installation instructions, build commands, or dependency information
- **FR-005**: Documentation MUST remove or clearly mark as "planned" any features that are documented but not yet implemented
- **FR-006**: Documentation MUST include file path references that match the current codebase structure
- **FR-007**: Documentation MUST maintain consistency between README.md and docs/ files visible to users

### Key Entities

- **Documentation Files**: README.md, docs/USAGE.md (user-facing documentation only)
- **Documented Features**: TUI interactive mode launch, keyboard shortcuts, interactive features (password visibility toggle, search navigation, statusbar shortcuts)
- **Code References**: File paths, component names, command examples from user perspective

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can find and follow instructions to launch TUI interactive mode in under 1 minute
- **SC-002**: 100% of user-facing features implemented in specs 001-003 are documented with usage examples
- **SC-003**: All documented keyboard shortcuts match current implementation (verified by testing each documented shortcut)
- **SC-004**: All documented file paths exist in the current codebase (zero broken references)
- **SC-005**: Users reading the documentation can discover and use password visibility toggle, search navigation, and statusbar shortcuts

## Assumptions

- The application's core CLI functionality (add, get, list, update, delete, generate) hasn't changed significantly and existing CLI documentation remains accurate
- Vault storage file is `~/.pass-cli/vault.enc` (encrypted JSON format, not `vault.json`)
- Recent changes are primarily in the TUI layer (cmd/tui) requiring documentation updates
- Documentation should reflect implemented features only, not planned/future features
- Scope is limited to user-facing documentation (README.md, docs/*.md visible on GitHub)
- Internal development documentation (CLAUDE.md, .specify/, docs/development/) is out of scope
