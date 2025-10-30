# Research: Password Visibility Toggle

**Feature**: Password Visibility Toggle
**Date**: 2025-10-09
**Purpose**: Resolve technical unknowns and document implementation approach

## Research Topics

### 1. tview InputField Masking API

**Question**: How does tview's InputField handle mask character toggling while preserving cursor position?

**Findings**:

- **API Method**: `InputField.SetMaskCharacter(rune)`
  - Setting mask to `'*'` (or any non-zero rune) masks the input
  - Setting mask to `0` displays plaintext (tview convention)
- **Cursor Preservation**: tview's `SetMaskCharacter()` automatically preserves cursor position and does NOT reset the input buffer
- **Unicode Support**: Fully supports Unicode/multi-byte characters in both masked and plaintext modes (tview handles UTF-8 natively)
- **Reference**: Existing codebase uses `AddPasswordField("Password", "", 0, '*', nil)` in forms.go:91 and forms.go:331

**Decision**: Use `SetMaskCharacter(0)` for visible mode, `SetMaskCharacter('*')` for masked mode. No additional cursor management needed.

**Rationale**: Leverages tview's built-in behavior, maintains consistency with existing form code patterns.

**Alternatives Considered**:
- Custom InputField wrapper with manual cursor tracking: Rejected - unnecessary complexity, reinvents tview functionality
- Separate visible/masked InputField widgets with swap logic: Rejected - loses focus management, introduces state sync issues

---

### 2. Keyboard Shortcut Registration in tview Forms

**Question**: How to add Ctrl+H shortcut to existing forms without conflicting with built-in tview navigation (Tab, Shift+Tab, Esc)?

**Findings**:

- **API Method**: `Form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey)`
  - Intercepts key events before tview's internal form navigation
  - Returning `nil` consumes the event (stops propagation)
  - Returning `event` allows tview to process it normally
- **Existing Pattern**: forms.go:231-249 (AddForm) and forms.go:549-567 (EditForm) already use `SetInputCapture` for Ctrl+S
  - Pattern: `case tcell.KeyCtrlS: af.onAddPressed(); return nil`
- **Ctrl+H Key Code**: `tcell.KeyCtrlH` (platform-agnostic constant)
- **No Conflicts**: Ctrl+H not used by tview's default form navigation or any existing shortcuts in forms.go

**Decision**: Add `case tcell.KeyCtrlH:` to existing `SetInputCapture` handlers in both AddForm and EditForm `setupKeyboardShortcuts()` methods.

**Rationale**: Minimal change to proven pattern, no new abstraction needed, maintains consistency with Ctrl+S shortcut.

**Alternatives Considered**:
- Global app-level shortcut handler: Rejected - bypasses form context, harder to track which form is active
- Per-field SetInputCapture on password field only: Rejected - doesn't work when focus is on other fields (breaks UX)

---

### 3. Visual Indicator for Toggle State

**Question**: What UI pattern should indicate current visibility state (masked vs. visible)?

**Findings**:

- **Existing Pattern** (detail panel): cmd/tui/events/handlers.go uses 'p' key to toggle password visibility in detail view
  - Visual feedback: Password text changes from `***` to plaintext directly in TextView
  - No separate indicator widget - the password field itself is the indicator
- **Form Constraints**: tview.Form has fixed layout - difficult to add inline icons without custom Form rendering
- **tview Label Support**: InputField.SetLabel() can update label text dynamically
  - Current: `"Password"`
  - Visible state: `"Password [VISIBLE]"` or `"Password 👁"`
  - Hidden state: `"Password [HIDDEN]"` or `"Password ●"`

**Decision**: Update InputField label to indicate state:
- Hidden (default): `"Password"`
- Visible: `"Password [VISIBLE]"`

Use text-only indicator for cross-platform terminal compatibility (no emoji dependencies).

**Rationale**: Minimal UI change, consistent with existing form label patterns, accessible on all terminals, no layout disruption.

**Alternatives Considered**:
- Emoji in label (e.g., 👁/●): Rejected - may not render on all terminals, accessibility concerns
- Separate status TextView widget: Rejected - disrupts form layout, adds complexity
- Button label change: Rejected - no button for toggle (keyboard/mouse clicks directly on field)
- Help text update (keyboard hints at bottom): Considered - may add as supplemental hint (`Ctrl+H: Toggle visibility`)

---

### 4. Mouse Click Activation

**Question**: How to make password field clickable to toggle visibility?

**Findings**:

- **tview InputField Mouse Support**:
  - `InputField.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse))` available in tview v0.42.0
  - `MouseAction` types: `MouseLeftClick`, `MouseLeftDown`, `MouseLeftDoubleClick`, etc.
- **Challenge**: Standard click on InputField focuses the field for text entry (expected behavior)
  - Toggling on every click would break typing workflow
  - Need separate click target (e.g., label area or specific region)
- **tview Form Limitation**: Form automatically manages focus - hard to intercept label clicks without custom Form

**Decision**: **Defer mouse activation to future iteration**. Implement keyboard shortcut (Ctrl+H) only for MVP.

**Rationale**:
- Keyboard shortcut satisfies core requirement (FR-005)
- Mouse activation (FR-006) requires custom Form rendering or separate clickable widget (significant scope increase)
- User feedback from clarification session indicated willingness to keep simple ("I feel like we can keep it, and we can have both")
- Can revisit in future story if keyboard-only proves insufficient

**Alternatives Considered**:
- Add clickable Button widget next to password field: Rejected - disrupts form layout, inconsistent with other fields
- Custom Form with label click regions: Rejected - major architectural change for minor feature
- Right-click context menu on password field: Rejected - non-standard UX, discoverability issue

**Follow-up**: Update keyboard hints text to include `Ctrl+H: Toggle password visibility` so users discover the feature.

---

### 5. Best Practices for Secure Password Visibility UX

**Question**: What security best practices apply to password visibility toggles in password managers?

**Findings**:

- **Industry Standards** (1Password, Bitwarden, KeePassXC):
  - Default to masked on form load ✅
  - Auto-hide on navigation away from field/form ✅
  - Clear visual indicator of current state ✅
  - No logging of visibility state changes ✅
- **OWASP Guidelines**:
  - Password visibility should be user-controlled, not automatic
  - Toggle should be easily reversible (single action to hide again)
  - Should work with assistive technologies (screen readers)
- **Terminal-Specific Consideration**:
  - No screen lock detection in terminal (unlike GUI apps)
  - User responsible for physical security (terminal visibility)
  - Clear indicator especially important (no window focus/blur events)

**Decision**: Follow all identified best practices:
1. Default masked (`SetMaskCharacter('*')`)
2. Reset to masked when form closed/canceled (in `SetOnSubmit`/`SetOnCancel` callbacks)
3. Label indicator shows current state (`"Password"` vs `"Password [VISIBLE]"`)
4. No logging of toggle events or visibility state
5. Update keyboard hints to show Ctrl+H shortcut

**Rationale**: Aligns with constitution Principle I (Security-First), matches user expectations from mainstream password managers.

**Alternatives Considered**:
- Auto-hide after N seconds: Rejected - adds complexity, unclear UX (user loses control)
- Remember toggle state across sessions: Rejected - security risk, violates principle of defaulting to secure state

---

## Implementation Summary

**Technical Approach**:

1. **AddForm Changes** (cmd/tui/components/forms.go):
   - Add `passwordVisible bool` field to AddForm struct
   - Add `togglePasswordVisibility()` method that:
     - Flips `passwordVisible` flag
     - Calls `GetFormItem(2).(*tview.InputField).SetMaskCharacter(mask)` with `mask = 0` (visible) or `'*'` (hidden)
     - Updates label via `SetLabel()` to show `"Password [VISIBLE]"` or `"Password"`
   - Modify `setupKeyboardShortcuts()` to add `case tcell.KeyCtrlH: af.togglePasswordVisibility(); return nil`
   - Modify `addKeyboardHints()` to include `"Ctrl+H: Toggle password"` in hints text
   - Ensure `buildFormFields()` initializes with masked state (already does via `AddPasswordField(..., '*', ...)`)

2. **EditForm Changes** (cmd/tui/components/forms.go):
   - Add `passwordVisible bool` field to EditForm struct
   - Add `togglePasswordVisibility()` method (identical logic to AddForm)
   - Modify `setupKeyboardShortcuts()` to add Ctrl+H handler
   - Modify `addKeyboardHints()` to include Ctrl+H hint
   - Ensure password field resets to masked when form opens (already does via SetMaskCharacter in buildFormFieldsWithValues:332)

3. **Testing** (new files):
   - **Unit Tests** (`tests/unit/tui_forms_test.go`):
     - TestAddFormPasswordToggle: Verify mask character changes on Ctrl+H
     - TestEditFormPasswordToggle: Verify mask character changes on Ctrl+H
     - TestPasswordResetOnFormClose: Verify passwordVisible resets to false when form canceled/submitted
   - **Integration Tests** (`tests/integration/tui_password_toggle_test.go`):
     - TestTogglePreservesCursor: Verify cursor position unchanged after toggle
     - TestToggleWithUnicode: Verify multi-byte characters display correctly in both modes
     - TestKeyboardHintsIncludeToggle: Verify Ctrl+H mentioned in on-screen hints

**No External Dependencies**: All functionality uses existing tview v0.42.0 and tcell v2.9.0 APIs.

**No Data Model Changes**: Feature is purely presentational - no vault schema or credential structure changes.

**No API Contracts**: TUI-only feature, no CLI commands or library interfaces affected.

