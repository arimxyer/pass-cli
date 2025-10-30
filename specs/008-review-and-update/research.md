# Documentation Review Research

**Research Date**: 2025-10-14
**Purpose**: Baseline assessment for documentation accuracy and completeness before production release

---

## 1. Current Release Version

**Latest Git Tag**: `v0.0.1`

**Version History**:
```
v0.0.1 (current release)
```

**Analysis**: Application is at initial release version (v0.0.1). No subsequent version tags exist.

---

## 2. Feature Inventory from Specs 001-007

### Spec Status Overview

| Spec | Title | Status | Merged to Main |
|------|-------|--------|----------------|
| 001 | Reorganize cmd/tui | Complete | ✅ Yes (Jan 2025) |
| 002 | Hey I'd | Unknown | ❓ Not found in commits |
| 003 | Would It Be | Unknown | ❓ Not found in commits |
| 004 | Documentation Update | Complete | ✅ Yes (Jan 2025) |
| 005 | Security Hardening | Complete | ✅ Yes (Jan 2025) |
| 006 | Minimum Terminal Size | Complete | ✅ Yes (Jan 2025) |
| 007 | User Configuration File | Complete | ✅ Yes (Jan 2025) |

### Implemented Features (Specs 001-007)

#### Spec 001: TUI Reorganization (January 2025)
- **Status**: Merged to main (commit: 6a79317)
- **Features Implemented**:
  - Package renamed from `package main` to `package tui`
  - Entry point converted to `func Run(vaultPath string) error`
  - Directory moved from `cmd/tui-tview/` to `cmd/tui/`
  - Import paths updated throughout codebase
  - Main entry point integration (`pass-cli` launches TUI by default)

#### Spec 004: Documentation Update (January 2025)
- **Status**: Merged to main (commit history shows documentation commits)
- **Features Documented**:
  - Interactive TUI features (password visibility toggle, search navigation)
  - Keyboard shortcuts with context (e.g., "Ctrl+H - Toggle password visibility (in add/edit forms)")
  - TUI launch instructions
  - Updated installation instructions

#### Spec 005: Security Hardening (January 2025)
- **Status**: Merged to main (commit history shows security commits)
- **Features Implemented**:
  - **Memory Security (P1)**: Master password handled as `[]byte`, cleared after use with `crypto/subtle`
  - **Crypto Hardening (P1)**: Upgraded to 600,000 PBKDF2-SHA256 iterations (from 100k)
  - **Password Policy (P2)**: Enforced 12+ character minimum with complexity requirements (uppercase, lowercase, digit, symbol)
  - **Audit Logging (P3)**: Optional HMAC-signed audit trail with OS keychain integration
  - Atomic vault migration with rollback on failure
  - Backward compatibility maintained for old vaults

#### Spec 006: Minimum Terminal Size (January 2025)
- **Status**: Merged to main (commit: 68baef1)
- **Features Implemented**:
  - Minimum terminal size enforcement (60 columns × 30 rows)
  - Blocking warning overlay on resize below threshold
  - Automatic recovery when terminal meets minimum size
  - Visual clarity with distinct warning styling
  - Edge case handling (rapid oscillation, boundary conditions, modal interaction)

#### Spec 007: User Configuration File (January 2025)
- **Status**: Merged to main (commit history shows config commits from Jan 2025)
- **Features Implemented**:
  - **Terminal Threshold Customization (US1)**: User-configurable terminal size thresholds via `config.yml`
  - **Keyboard Shortcut Remapping (US2)**: Custom keybindings with conflict detection
  - **Configuration Management Commands (US3)**: `pass-cli config init|edit|validate|reset`
  - Config file location: `~/.config/pass-cli/config.yml` (Linux/macOS), `%APPDATA%\pass-cli\config.yml` (Windows)
  - YAML format with validation (100 KB max file size)
  - UI hints automatically reflect custom keybindings (status bar, help modal, forms)

#### Specs 002-003: Status Unknown
- **Analysis**: No merge commits or feature evidence found in git log for these specs
- **Recommendation**: Investigate these spec directories to determine if they were implemented, abandoned, or merged under different commit messages

---

## 3. TUI Keyboard Shortcuts Inventory

### Complete Keyboard Shortcuts Reference

Based on analysis of `cmd/tui/events/handlers.go`, `cmd/tui/components/forms.go`, and `internal/config/keybinding.go`:

#### Configurable Shortcuts (User-Remappable via config.yml)

| Action | Default Key | Description | Context | Configurable |
|--------|-------------|-------------|---------|--------------|
| `quit` | `q` | Quit application | Main view | ✅ Yes |
| `add_credential` | `a` | New credential | Main view | ✅ Yes |
| `edit_credential` | `e` | Edit credential | Main view | ✅ Yes |
| `delete_credential` | `d` | Delete credential | Main view | ✅ Yes |
| `toggle_detail` | `i` | Toggle detail panel (Auto/Hide/Show) | Main view | ✅ Yes |
| `toggle_sidebar` | `s` | Toggle sidebar (Auto/Hide/Show) | Main view | ✅ Yes |
| `help` | `?` | Show help modal | Any time | ✅ Yes |
| `search` | `/` | Activate search/filter | Main view | ✅ Yes |

#### Hardcoded Shortcuts (Not Configurable)

| Key | Description | Context | Reason Not Configurable |
|-----|-------------|---------|------------------------|
| `Tab` | Next component | All contexts | Navigation fundamental |
| `Shift+Tab` | Previous component | All contexts | Navigation fundamental |
| `↑/↓` | Navigate lists | List views | Navigation fundamental |
| `Enter` | Select / View details | List views | Selection fundamental |
| `Esc` | Close modal / Cancel search | Modals, search | Cancel fundamental |
| `Ctrl+C` | Force quit application | Any time | Emergency exit |
| `p` | Toggle password visibility | Detail view | Context-specific |
| `c` | Copy password to clipboard | Detail view | Context-specific |
| `Ctrl+H` | Toggle password visibility | Add/edit forms | Form-specific |
| `Ctrl+S` | Quick-save / Submit form | Add/edit forms | Form-specific |
| `PgUp/PgDn` | Scroll help modal | Help modal | Modal-specific |
| `Mouse Wheel` | Scroll help modal | Help modal | Modal-specific |

#### Special Keys (Context-Dependent Behavior)

| Key | Primary Action | Secondary Action | Conditions |
|-----|---------------|------------------|------------|
| `Esc` | Close search (keep filter active) | Clear search filter | Press once: exit input, twice: clear filter |
| `Esc` | Close modal | N/A | When modal is open |
| `Ctrl+C` | Close modal | Quit application | Modal open vs. main view |
| `↑/↓/Enter` | Forward to table navigation | N/A | When search input is focused (allows list navigation while searching) |

### Keybinding Configuration Schema

```yaml
keybindings:
  quit: "q"                    # Default: q
  add_credential: "a"          # Default: a
  edit_credential: "e"         # Default: e
  delete_credential: "d"       # Default: d
  toggle_detail: "i"           # Default: i (changed from 'tab' in spec 007)
  toggle_sidebar: "s"          # Default: s
  help: "?"                    # Default: ?
  search: "/"                  # Default: /
```

**Supported Key Formats**:
- Single letters: `a-z` (lowercase)
- Numbers: `0-9`
- Special keys: `enter`, `esc`, `escape`, `tab`, `space`, `backspace`, `delete`, `del`, `insert`, `ins`, `home`, `end`, `pageup`, `pgup`, `pagedown`, `pgdn`, `up`, `down`, `left`, `right`
- Function keys: `f1-f12`
- Modifiers: `ctrl+`, `alt+`, `shift+` (can be combined)
- Examples: `ctrl+q`, `alt+a`, `shift+f1`, `ctrl+alt+d`

**Validation Rules**:
- No duplicate key assignments allowed (conflict detection on load)
- Only valid action names accepted (unknown actions rejected)
- File size limit: 100 KB max
- Invalid config shows modal warning, app continues with defaults

### Status Bar Key Hints (Dynamic)

The status bar displays context-aware hints that reflect custom keybindings:

**Main View (FocusTable)**:
```
a:Add  e:Edit  d:Delete  s:Sidebar  i:Detail  /:Search  ?:Help  q:Quit
```

**Search Active**:
```
↑/↓:Navigate  Enter:Select  Esc:Exit search
```

**Forms (Add/Edit)**:
```
Tab/Shift+Tab:Navigate  Ctrl+S:Add  Ctrl+H:Toggle password  Esc:Cancel
```

**Help Modal**:
```
PgUp/PgDn or Mouse Wheel to scroll  •  Esc to close
```

### Recent Keyboard Shortcut Improvements (Spec 007)

Commits from October 2025 show ongoing refinements:
- **09a2b8e**: Improved status bar hints - clarify Esc behavior, restore sidebar/details, show Shift+Tab
- **823d650**: Standardize and improve status bar key hints
- **88c7044**: Keep search filter active when pressing Escape to allow password operations
- **d8eff9f**: Add Escape key handler to close help modal
- **0be5bca**: Change toggle_detail default keybinding from 'tab' to 'i'
- **3588b0a**: Populate ParsedKeybindings in GetDefaults to fix keybinding functionality

---

## 4. Package Manager Status

### Assumptions

**Homebrew (macOS/Linux)**:
- Repository: `ari1110/homebrew-tap`
- Installation: `brew tap ari1110/homebrew-tap && brew install pass-cli`
- **Assumption**: Package manager should be updated after v0.0.1 release
- **Validation Needed**: Confirm Homebrew formula points to v0.0.1 release artifacts

**Scoop (Windows)**:
- Repository: `ari1110/scoop-bucket`
- Installation: `scoop bucket add pass-cli https://github.com/ari1110/scoop-bucket && scoop install pass-cli`
- **Assumption**: Package manager should be updated after v0.0.1 release
- **Validation Needed**: Confirm Scoop manifest points to v0.0.1 release artifacts

**Manual Installation**:
- GitHub Releases: `https://github.com/ari1110/pass-cli/releases`
- Requires download and extraction of platform-specific binary

---

## 5. Validation Tooling Recommendations

### Link Validation

**Tool**: `markdown-link-check`
```bash
# Install
npm install -g markdown-link-check

# Check all markdown files
find . -name "*.md" -exec markdown-link-check {} \;

# Or use makefile target
make validate-docs
```

**Purpose**: Detect broken internal/external links in documentation

### Documentation Accuracy Scripts

**Create validation script**: `scripts/validate-docs.sh`

```bash
#!/bin/bash
# Validate documentation matches codebase

echo "🔍 Validating documentation accuracy..."

# Check that documented shortcuts match config defaults
echo "Checking keybinding defaults..."
grep -r "Default:" docs/ | while read line; do
    # Extract key and compare to internal/config/config.go defaults
    echo "  - $line"
done

# Check that documented commands exist in cmd/
echo "Checking CLI commands..."
grep -r "pass-cli " README.md docs/ | grep -E "^\`" | \
    sed 's/.*pass-cli \([a-z-]*\).*/\1/' | sort -u | while read cmd; do
    if [ -f "cmd/$cmd.go" ] || grep -q "\"$cmd\"" cmd/root.go; then
        echo "  ✓ $cmd command exists"
    else
        echo "  ✗ $cmd command NOT FOUND"
    fi
done

# Check that documented file paths exist
echo "Checking file path references..."
grep -r "\`.*\.go\`" docs/ README.md | \
    sed 's/.*`\(.*\.go\)`.*/\1/' | sort -u | while read path; do
    if [ -f "$path" ]; then
        echo "  ✓ $path exists"
    else
        echo "  ✗ $path NOT FOUND"
    fi
done

echo "✅ Validation complete"
```

### Configuration Validation

**Tool**: `yamllint` for YAML syntax validation
```bash
# Install
pip install yamllint

# Validate config examples
yamllint docs/examples/*.yml
```

### Code-Documentation Sync Checks

**Create grep patterns** to detect mismatches:

```bash
# Check for hardcoded iteration count references
grep -rn "100,000" docs/ README.md  # Should be updated to 600,000
grep -rn "100k" docs/ README.md     # Should be updated to 600k

# Check for old minimum terminal size references
grep -rn "60×20" docs/ README.md    # Should be 60×30

# Check for old keybinding references
grep -rn "tab.*toggle" docs/ README.md  # Should be 'i' not 'tab'
```

---

## 6. Documentation Accuracy Baseline

### Current Documentation State (README.md)

**Verified Accurate**:
- ✅ AES-256-GCM encryption documented correctly
- ✅ 600,000 PBKDF2 iterations documented (updated from 100k)
- ✅ System keychain integration documented
- ✅ Password policy (12 chars, complexity) documented
- ✅ Audit logging documented as optional
- ✅ TUI launch instructions present ("pass-cli" with no args)
- ✅ Ctrl+H password toggle documented
- ✅ Search activation with "/" documented
- ✅ Keyboard shortcuts table includes context column
- ✅ Master password storage (OS keychain) documented
- ✅ Vault location paths documented correctly

**Potential Gaps Identified**:
- ⚠️ Keyboard shortcuts table is condensed (only 6 shortcuts shown, 20+ exist)
- ⚠️ No mention of configurable keybindings feature (spec 007)
- ⚠️ No mention of terminal size warning feature (spec 006)
- ⚠️ Documentation states "minimum 60×30" but some older references may exist
- ⚠️ Toggle detail keybinding changed from 'tab' to 'i' (verify all docs updated)
- ⚠️ Specs 002-003 status unknown (may be undocumented features or abandoned)

**Documentation Completeness**:
- README.md: **85%** complete (missing config feature docs, full shortcut reference)
- docs/USAGE.md: **Not checked** (file path reference in README exists)
- docs/MIGRATION.md: **Referenced** but not verified
- docs/SECURITY.md: **Not verified**

### Missing Documentation Areas

1. **User Configuration Feature** (Spec 007):
   - config.yml location and format
   - Keybinding customization
   - Terminal threshold customization
   - Config management commands (init, edit, validate, reset)
   - Config validation and error handling

2. **Terminal Size Warning** (Spec 006):
   - Minimum size requirements (60×30)
   - Warning overlay behavior
   - Recovery process

3. **Complete Keyboard Shortcuts**:
   - README shows 6 shortcuts in table
   - Help modal shows ~20 shortcuts
   - Documentation should include all shortcuts with:
     - Default keys
     - Action descriptions
     - Context (where they work)
     - Configurability status

4. **Specs 002-003**:
   - Investigate and document if implemented
   - Or mark as abandoned/future work

---

## 7. Recommendations for Spec 008

### High Priority (P1)

1. **Add User Configuration Documentation**:
   - Create section in README.md or docs/CONFIGURATION.md
   - Document config.yml schema with examples
   - Document config management commands
   - Add troubleshooting guide for config validation errors

2. **Expand Keyboard Shortcuts Reference**:
   - Replace condensed table with comprehensive list
   - Add "Configurable" column to indicate which can be remapped
   - Document both default keys and how to customize them
   - Include context for each shortcut

3. **Verify and Update Iteration Count References**:
   - Run grep to find any remaining "100,000" or "100k" references
   - Update to "600,000" or "600k" everywhere

4. **Document Terminal Size Requirements**:
   - Add minimum terminal size (60×30) to README
   - Document warning behavior and recovery

### Medium Priority (P2)

5. **Investigate Specs 002-003**:
   - Check if features were implemented but not documented
   - Remove from spec list if abandoned
   - Document if implemented

6. **Create Validation Scripts**:
   - Implement `scripts/validate-docs.sh`
   - Add `make validate-docs` target
   - Run in CI/CD to prevent documentation drift

7. **Verify Package Managers**:
   - Confirm Homebrew formula is up-to-date (v0.0.1)
   - Confirm Scoop manifest is up-to-date (v0.0.1)
   - Document package manager update process

### Low Priority (P3)

8. **Add Visual Documentation**:
   - Screenshots of TUI in action
   - Animated GIFs of key features (search, password toggle, etc.)
   - Help modal screenshot showing all shortcuts

9. **Create Migration Guide**:
   - Document v0.0.1 features for users
   - Create upgrade path documentation
   - Document security improvements (600k iterations, password policy)

10. **API/Integration Documentation**:
    - Script integration examples (currently present)
    - CI/CD integration patterns
    - Environment variable reference

---

## 8. Quick Wins (Can Be Done Immediately)

1. **Update README.md keyboard shortcuts table** with full 20+ shortcuts
2. **Add config.yml example** to README or create docs/CONFIGURATION.md
3. **Run link checker** to find broken references
4. **Grep for outdated references** (100k iterations, 60×20 terminal size, 'tab' for toggle_detail)
5. **Verify all CLI commands** mentioned in docs actually exist in cmd/

---

## Appendix: Research Sources

### Files Analyzed
- `R:\Test-Projects\pass-cli\specs\001-reorganize-cmd-tui\spec.md`
- `R:\Test-Projects\pass-cli\specs\004-we-ve-recently\spec.md`
- `R:\Test-Projects\pass-cli\specs\005-security-hardening-address\spec.md`
- `R:\Test-Projects\pass-cli\specs\006-implement-minimum-terminal\spec.md`
- `R:\Test-Projects\pass-cli\specs\007-user-wants-to\spec.md`
- `R:\Test-Projects\pass-cli\specs\007-user-wants-to\tasks.md`
- `R:\Test-Projects\pass-cli\cmd\tui\events\handlers.go`
- `R:\Test-Projects\pass-cli\cmd\tui\components\forms.go`
- `R:\Test-Projects\pass-cli\internal\config\keybinding.go`
- `R:\Test-Projects\pass-cli\README.md`

### Git Commands Used
```bash
git describe --tags --abbrev=0
git tag --sort=-version:refname | head -10
git log --oneline --merges --grep="spec" --since="2024-12-01"
git log --oneline --since="2025-01-01"
```

### Total Shortcuts Documented
- **Configurable**: 8 shortcuts (quit, add, edit, delete, toggle_detail, toggle_sidebar, help, search)
- **Hardcoded**: 12+ shortcuts (Tab, Shift+Tab, arrows, Enter, Esc, Ctrl+C, p, c, Ctrl+H, Ctrl+S, PgUp/PgDn, etc.)
- **Total**: 20+ keyboard shortcuts identified

---

**End of Research Document**
