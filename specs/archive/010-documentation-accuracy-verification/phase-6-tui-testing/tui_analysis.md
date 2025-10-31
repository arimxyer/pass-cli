# TUI Keyboard Shortcuts Analysis

## Comparison: README.md vs Actual TUI Implementation

### README.md Documentation (Lines 97-128)

**Configurable Shortcuts (8):**
| Shortcut | Action | Context |
|----------|--------|---------|
| `q` | Quit application | Any time |
| `a` | Add new credential | Main view |
| `e` | Edit credential | Main view |
| `d` | Delete credential | Main view |
| `i` | Toggle detail panel (Auto/Hide/Show) | Main view |
| `s` | Toggle sidebar (Auto/Hide/Show) | Main view |
| `?` | Show help modal | Any time |
| `/` | Activate search/filter | Main view |

**Hardcoded Shortcuts (11):**
| Shortcut | Action | Context |
|----------|--------|---------|
| `Tab` | Next component | All views |
| `Shift+Tab` | Previous component | All views |
| `↑/↓` | Navigate lists | List views |
| `Enter` | Select / View details | List views |
| `Esc` | Close modal / Exit search | Modals, search |
| `Ctrl+C` | Force quit application | Any time |
| `p` | Copy password to clipboard | Detail view |
| `c` | Copy username to clipboard | Detail view |
| `Ctrl+H` | Toggle password visibility | Add/edit forms |
| `Ctrl+S` | Quick-save / Submit form | Add/edit forms |
| `PgUp/PgDn` | Scroll help modal | Help modal |

### Actual TUI Help (from `pass-cli tui --help`)

**Documented Shortcuts (9):**
| Shortcut | Action |
|----------|--------|
| `n` | New credential |
| `e` | Edit credential |
| `d` | Delete credential |
| `c` | Copy password |
| `p` | Toggle password visibility |
| `/` | Search/filter |
| `?` | Show help |
| `Tab` | Cycle focus between panels |
| `q` | Quit |

## Discrepancies Found

### 1. Different Shortcut Keys
- **README says:** `a` for Add new credential
- **TUI help says:** `n` for New credential

### 2. Missing Features in TUI Help
The README.md documents these features that are NOT mentioned in the TUI help:
- `i` - Toggle detail panel
- `s` - Toggle sidebar
- `Shift+Tab` - Previous component
- `↑/↓` - Navigate lists
- `Enter` - Select/View details
- `Esc` - Close modal/Exit search
- `Ctrl+C` - Force quit application
- `c` - Copy username (different meaning in TUI help)
- `Ctrl+H` - Toggle password visibility (different meaning in TUI help)
- `Ctrl+S` - Quick-save/Submit form
- `PgUp/PgDn` - Scroll help modal

### 3. Inconsistent Functionality
- **README says:** `c` copies username, `p` copies password
- **TUI help says:** `c` copies password, `p` toggles password visibility

### 4. Missing Documentation
The TUI help doesn't mention:
- Panel toggling (detail/sidebar)
- Advanced navigation (Shift+Tab, arrow keys)
- Form-specific shortcuts (Ctrl+H, Ctrl+S)
- Force quit (Ctrl+C)

## Test Environment Setup

✅ **Test Vault Created Successfully**
- Location: `~/.pass-cli/test_vault.enc`
- Master Password: `TestMasterP@ss123`
- Test Credentials:
  - `testservice` (testuser/TestPassword123!)
  - `github` (githubuser/GitHubPass123!)

## Manual Testing Required

To verify actual TUI behavior:
1. Launch: `./pass-cli.exe --vault ~/.pass-cli/test_vault.enc tui`
2. Enter master password: `TestMasterP@ss123`
3. Test each keyboard shortcut and document actual behavior

## Priority Issues to Fix

1. **High Priority:** Standardize shortcut keys between README and implementation
2. **Medium Priority:** Update TUI help to include all documented features
3. **Low Priority:** Consider consistency with common TUI patterns