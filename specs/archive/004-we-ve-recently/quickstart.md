# Quickstart: Documentation Update Implementation

**Branch**: `004-we-ve-recently` | **Date**: 2025-10-11 | **Phase**: 1 (Design)

## Goal

Update README.md and docs/USAGE.md to document TUI interactive features added in specs 001-003.

## Prerequisites

- Completed specs 001, 002, and 003 (TUI features implemented)
- Access to `research.md` (verified feature list and keyboard shortcuts)
- Text editor for documentation files

## Implementation Steps

### Step 1: Add TUI Mode Section to README.md

**Location**: After "Quick Start" section, before "Usage" section

**Content to Add**:
```markdown
## 🎨 Interactive TUI Mode

Pass-CLI includes an interactive Terminal User Interface (TUI) for visual credential management.

### Launching TUI Mode

```bash
# Launch TUI (no arguments)
pass-cli

# CLI commands still work with explicit subcommands
pass-cli list
pass-cli get github
```

### TUI Features

- **Visual Navigation**: Browse credentials with arrow keys and Tab
- **Interactive Forms**: Add/edit credentials with visual feedback
- **Password Visibility Toggle**: Press `Ctrl+H` in forms to verify passwords
- **Search & Filter**: Press `/` to search, `Esc` to clear
- **Keyboard Shortcuts**: Press `?` to see all available shortcuts
- **Responsive Layout**: Sidebar and detail panel adapt to terminal size

### Key TUI Shortcuts

| Shortcut | Action | Context |
|----------|--------|---------|
| `Ctrl+H` | Toggle password visibility | Add/edit forms |
| `/` | Activate search mode | Main view |
| `s` | Toggle sidebar | Main view |
| `i` | Toggle detail panel | Main view |
| `?` | Show help | Any time |
| `q` | Quit | Any time |

See [full keyboard shortcuts reference](docs/USAGE.md#tui-keyboard-shortcuts) for complete list.
```

**Acceptance**: Users can find TUI launch instructions and understand interactive vs CLI mode.

---

### Step 2: Add TUI Section to docs/USAGE.md

**Location**: After "Configuration" section, before "Best Practices" section

**Content to Add**: See `contracts/tui-usage-section.md` for complete content.

**Key Subsections**:
1. Launching TUI Mode
2. TUI vs CLI Mode
3. Complete Keyboard Shortcuts Reference (organized by context)
4. Search & Filter
5. Password Visibility Toggle
6. Layout Controls (sidebar, detail panel)
7. Usage Location Display

**Acceptance**: All TUI features from specs 001-003 are documented with usage examples.

---

### Step 3: Update Installation/Build Instructions (if needed)

**Check**: Do current build instructions work?

```bash
# Test these commands from README.md
go build -o pass-cli .
make build  # If Makefile exists
```

**If outdated**: Update with current working commands.

**Acceptance**: Build instructions execute successfully on clean checkout.

---

### Step 4: Verify File Paths

**Check**: All file paths referenced in documentation exist.

**Common References to Verify**:
- `~/.pass-cli/vault.enc` (vault location)
- `~/.pass-cli/config.yaml` (config file)
- `~/.pass-cli/audit.log` (audit log, if mentioned)
- `cmd/tui/` (source code references)

**Acceptance**: Zero broken file path references.

---

### Step 5: Manual Verification

**Test each documented feature**:

1. **TUI Launch**:
   ```bash
   pass-cli  # Should launch TUI
   ```

2. **Ctrl+H Toggle**:
   - Press `n` to open add form
   - Type password
   - Press `Ctrl+H` → password visible
   - Press `Ctrl+H` → password masked

3. **Search**:
   - Press `/` → search input appears
   - Type "git" → filters results
   - Press `Esc` → clears search

4. **Sidebar Toggle**:
   - Press `s` → sidebar hides
   - Press `s` → sidebar shows
   - Press `s` → sidebar auto (responsive)

5. **Usage Locations**:
   - Select credential
   - Check detail panel shows "Usage Locations" section

**Acceptance**: Every documented feature works as described.

---

## Before/After Comparison

### Before (Current State)

**README.md**: No mention of TUI mode. Users only see CLI commands.

**docs/USAGE.md**: No keyboard shortcuts, no interactive features documented.

### After (Target State)

**README.md**:
- TUI launch instructions in Quick Start
- Feature highlights with keyboard shortcuts table
- Link to full shortcuts reference

**docs/USAGE.md**:
- Complete "TUI Mode" section (1000+ words)
- Organized keyboard shortcuts by context
- Usage examples for all interactive features
- Search, toggle, and navigation documentation

---

## Rollback Plan

If documentation changes cause confusion:

1. Check git history: `git log README.md docs/USAGE.md`
2. Revert specific file: `git checkout HEAD~1 README.md`
3. Re-evaluate changes and update incrementally

---

## Testing Checklist

- [ ] TUI launches when running `pass-cli` with no args
- [ ] `Ctrl+H` toggles password visibility in add form
- [ ] `Ctrl+H` toggles password visibility in edit form
- [ ] `/` key activates search mode
- [ ] `Esc` exits search mode
- [ ] `s` key toggles sidebar (3 states)
- [ ] `i` key toggles detail panel (3 states)
- [ ] `?` key shows help modal
- [ ] `q` or `Ctrl+C` quits TUI
- [ ] Usage locations visible in detail panel
- [ ] All file paths in docs exist
- [ ] All examples are copy-pasteable

---

## Common Issues

**Issue**: "TUI doesn't launch"
**Fix**: Ensure `main.go` calls `tui.Run()` when no args provided (implemented in spec 001)

**Issue**: "Ctrl+H doesn't work"
**Fix**: Verify forms.go implements `SetInputCapture` with `tcell.KeyCtrlH` handler (implemented in spec 003)

**Issue**: "Search doesn't filter"
**Fix**: Verify AppState implements search filtering logic (implemented in spec 002)

---

## Next Phase

After completing this quickstart implementation → Run `/speckit.tasks` to generate tasks.md with atomic implementation steps.
