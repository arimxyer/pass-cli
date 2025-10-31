# Feature Specification: Password Visibility Toggle

**Feature Branch**: `003-would-it-be`
**Created**: 2025-10-09
**Status**: Implementation Complete
**Completion**: 2025-10-20 (41/42 tasks completed, password toggle via Ctrl+H implemented)
**Input**: User description: "Would it be possible for us to have the ability to apply the show/hide toggle with passwords in both forms? So that you can see what you're adding or editting?"

## Clarifications

### Session 2025-10-09

- Q: Which password fields should have the visibility toggle? → A: Add and edit forms each contain only a single password field (no password confirmation fields present)
- Q: How should users activate the toggle via keyboard? → A: Dedicated keyboard shortcut (when on the form page/modal), similar to how 'p' toggles password in the details panel
- Q: What type of visual indicator should the toggle control use? → A: Match the existing pattern used elsewhere in the application
- Q: Should the keyboard shortcut use the same key ('p') or a different key? → A: Use Ctrl+H keyboard shortcut, similar to how Ctrl+S is for saving the form
- Q: Should mouse/pointer interaction be supported? → A: Keep FR-006 - support both Ctrl+H keyboard shortcut and mouse/pointer interaction (clickable control)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Toggle Password Visibility When Adding Entries (Priority: P1)

When adding a new password entry, users need the ability to verify they've typed the password correctly before saving. This is critical because typos in password entry can lock users out of accounts.

**Why this priority**: This is the primary use case and addresses the most common pain point - preventing password typos during initial entry. Without this, users may save incorrect passwords and experience login failures.

**Independent Test**: Can be fully tested by opening the add password form, typing a password, toggling visibility on/off, and verifying the password displays correctly. Delivers immediate value by preventing typos during password creation.

**Acceptance Scenarios**:

1. **Given** the user is on the add password form with a password field, **When** the user activates the visibility toggle, **Then** the password characters change from masked (hidden) to plaintext (visible)
2. **Given** the password is currently visible in the add form, **When** the user activates the visibility toggle again, **Then** the password characters change back to masked (hidden)
3. **Given** the user is typing a new password in the add form, **When** the user toggles visibility while typing, **Then** all previously typed characters display according to the current visibility state without losing focus

---

### User Story 2 - Toggle Password Visibility When Editing Entries (Priority: P2)

When editing existing password entries, users need the ability to see what they're changing. This helps verify edits are correct before saving, especially when updating passwords for important accounts.

**Why this priority**: Slightly lower priority than adding new entries because editing is less frequent, but still critical for preventing errors when updating passwords.

**Independent Test**: Can be fully tested by opening an existing password entry in edit mode, modifying the password, toggling visibility on/off, and verifying the password displays correctly. Delivers value by preventing typos during password updates.

**Acceptance Scenarios**:

1. **Given** the user is editing an existing password entry, **When** the user activates the visibility toggle, **Then** the password characters change from masked to plaintext
2. **Given** the password is currently visible in the edit form, **When** the user activates the visibility toggle again, **Then** the password characters change back to masked
3. **Given** the user is modifying a password in the edit form, **When** the user toggles visibility while editing, **Then** the edited password displays according to the current visibility state without losing focus or cursor position

---

### User Story 3 - Persistent Visibility State Awareness (Priority: P3)

Users need clear visual feedback about whether their password is currently visible or hidden. This prevents accidental exposure in public settings and provides confidence about the current state.

**Why this priority**: This is a usability enhancement that improves the core feature but isn't essential for basic functionality. Users can still use the toggle without explicit state indicators.

**Independent Test**: Can be fully tested by toggling password visibility and observing the visual indicator (icon, label, or button state) changes to reflect current state. Delivers value by improving user confidence and preventing accidental exposure.

**Acceptance Scenarios**:

1. **Given** the password is currently hidden, **When** the user views the toggle control, **Then** the control displays an indicator suggesting "show" action (e.g., "Show Password" or eye icon)
2. **Given** the password is currently visible, **When** the user views the toggle control, **Then** the control displays an indicator suggesting "hide" action (e.g., "Hide Password" or eye-slash icon)
3. **Given** the user switches between add and edit forms, **When** the user returns to a form, **Then** the visibility state resets to hidden (default secure state)

---

### Edge Cases

- What happens when the user toggles visibility on an empty password field? (Toggle should still work but show no characters)
- What happens when the password field contains special characters, emojis, or multi-byte characters? (All characters should display correctly in visible mode)
- What happens when the user copies text while password is visible? (Copy function should work normally)
- What happens when the user cancels the form after toggling visibility? (No state should persist, form should close normally)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a visibility toggle control for the password field in the add password form
- **FR-002**: System MUST provide a visibility toggle control for the password field in the edit password form
- **FR-003**: System MUST display password characters as plaintext when visibility is enabled
- **FR-004**: System MUST display password characters as masked (e.g., asterisks or dots) when visibility is disabled
- **FR-005**: System MUST allow users to activate the visibility toggle via Ctrl+H keyboard shortcut when focus is on the add or edit form
- **FR-006**: ~~System MUST allow users to activate the visibility toggle via mouse/pointer interaction with a clickable control~~ **[DEFERRED]** - Mouse/pointer activation deferred to future iteration due to tview Form API limitations requiring custom rendering (see research.md Section 4). MVP implements keyboard-only via Ctrl+H.
  - **Future Acceptance Criteria (when un-deferred)**:
    - Clickable widget (button, icon, or label region) positioned adjacent to password field
    - Mouse click on widget toggles visibility (same behavior as Ctrl+H)
    - Visual feedback on hover (if terminal supports mouse events)
    - Widget state indicates current visibility (show/hide icon or text label)
    - Implementation requires custom Form rendering or separate clickable tview primitive (estimated 4-6 hours additional work per research.md)
- **FR-007**: System MUST provide clear visual feedback indicating the current visibility state (visible vs. hidden) using patterns consistent with existing application UI
- **FR-008**: System MUST maintain cursor position (character index) when toggling visibility while user is typing, regardless of whether cursor is at the beginning, middle, or end of the password field
- **FR-009**: System MUST default password fields to hidden state when forms are first opened
- **FR-010**: System MUST reset password visibility to hidden when navigating away from forms

### Non-Functional Requirements

- **NFR-001**: Password visibility toggle MUST use tview's UTF-8 character handling for all password character types (special characters, Unicode, multi-byte).
  - **Testability Note**: Display correctness (glyph rendering, character width) is terminal-dependent and outside our control. We test correct API usage (`SetMaskCharacter(0)` preserves all runes) but cannot test visual rendering.
  - **Acceptance Criteria**: Integration test T174 verifies Unicode passwords (e.g., "测试🔐emoji") toggle between masked and visible states without data loss. Terminal rendering quality is user's terminal responsibility.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can toggle password visibility in under 1 second from any point in the password entry workflow
- **SC-002**: 100% of password fields in add and edit forms support visibility toggling
- **SC-003**: Password visibility toggle is accessible via keyboard (Ctrl+H) without losing focus or cursor position
- **SC-004**: Users can type a 16-character password, toggle visibility to verify correctness, toggle back to masked, and save successfully - measured by completing manual test scenario without re-typing due to typos (demonstrates visibility toggle enables error-free entry)
- **SC-005**: Password fields default to hidden state on form load, providing security by default
- **SC-006**: Toggle state changes are reflected immediately (under 100ms) with clear visual feedback
- **SC-007**: ~~Mouse/pointer interaction success criteria~~ **[DEFERRED]** - Pending FR-006 implementation (deferred per research.md Section 4). MVP success criteria focus on keyboard-only interaction (SC-001 through SC-006).
