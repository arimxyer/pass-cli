---
title: "TUI Guide"
weight: 5
toc: true
---

Pass-CLI includes an interactive Terminal User Interface (TUI) for visual credential management. The TUI provides an alternative to CLI commands with visual navigation, real-time search, and keyboard-driven workflows.

### Launching TUI Mode

```bash
# Launch TUI (no arguments)
pass-cli

# TUI opens automatically when no subcommand is provided
```

The TUI launches immediately and displays:
- **Left sidebar**: Category navigation (auto-hides on narrow terminals)
- **Center table**: Credential list with service name, username, last accessed time
- **Right panel**: Credential details with password, URL, notes, usage locations
- **Bottom status bar**: Context-aware keyboard shortcuts and status messages

### TUI vs CLI Mode

Pass-CLI operates in two modes:

| Mode | Activation | Use Case |
|------|------------|----------|
| **TUI Mode** | Run `pass-cli` with no arguments | Interactive browsing, visual credential management |
| **CLI Mode** | Run `pass-cli <command>` with explicit subcommand | Scripts, automation, quick single operations |

**Examples**:
```bash
# TUI Mode
pass-cli                        # Opens interactive interface

# CLI Mode
pass-cli list                   # Outputs credential table to stdout
pass-cli get github --quiet     # Outputs password only (script-friendly)
pass-cli add newcred            # Interactive prompts for credential data
```

Both modes access the same encrypted vault file (`~/.pass-cli/vault.enc`).

### TUI Keyboard Shortcuts

#### Navigation

| Shortcut | Action | Context |
|----------|--------|---------|
| `Tab` | Next component | Any view |
| `Shift+Tab` | Previous component | Any view |
| `↑` / `↓` | Navigate lists | Table, sidebar |
| `Enter` | Select credential / View details | Table |

#### Actions

| Shortcut | Action | Context |
|----------|--------|---------|
| `n` | New credential (opens add form) | Main view |
| `e` | Edit selected credential | Main view (credential selected) |
| `d` | Delete selected credential | Main view (credential selected) |
| `p` | Toggle password visibility | Detail panel |
| `c` | Copy password to clipboard | Detail panel |
| `u` | Copy username to clipboard | Detail panel |
| `l` | Copy URL to clipboard | Detail panel |
| `n` | Copy notes to clipboard | Detail panel |
| `t` | Copy TOTP code to clipboard | Detail panel |

#### View Controls

| Shortcut | Action | Context |
|----------|--------|---------|
| `i` | Toggle detail panel (Auto/Hide/Show) | Main view |
| `s` | Toggle sidebar (Auto/Hide/Show) | Main view |
| `/` | Activate search mode | Main view |

#### Forms (Add/Edit)

| Shortcut | Action | Context |
|----------|--------|---------|
| `Ctrl+S` | Save form | Add/edit forms |
| `Ctrl+P` | Toggle password visibility | Add/edit forms |
| `Ctrl+G` | Generate random password | Add/edit forms (password field) |
| `Tab` | Next field | Forms |
| `Shift+Tab` | Previous field | Forms |
| `Esc` | Cancel / Close form | Forms |

#### General

| Shortcut | Action | Context |
|----------|--------|---------|
| `?` | Show help modal | Any time |
| `q` | Quit application | Main view |
| `Esc` | Close modal / Cancel search | Modals, search mode |
| `Ctrl+C` | Quit application | Any time |

**Note**: Configurable shortcuts (a, e, d, i, s, ?, /, q) can be customized via the config file (see [Configuration](#configuration) section for paths). Navigation shortcuts (Tab, arrows, Enter, Esc, Ctrl+P, Ctrl+S, Ctrl+C) are hardcoded and cannot be changed.

### Search & Filter

Press `/` to activate search mode. An input field appears at the top of the credential table.

**Search Behavior**:
- **Case-insensitive**: "git" matches "GitHub", "gitlab", "digit"
- **Substring matching**: Query can appear anywhere in field
- **Searchable fields**: Service name, username, URL, category (Notes field excluded)
- **Real-time filtering**: Results update as you type
- **Navigation**: Use `↑`/`↓` arrow keys to navigate filtered results

**Examples**:
```bash
# Search for GitHub credentials
/
github      # Type query → only GitHub credentials shown

# Search by category
/
dev         # Shows credentials in "Development" category

# Clear search
Esc         # Exits search mode, shows all credentials
```

**When searching**:
- Newly added credentials matching the query appear immediately in results
- Selection preserved if selected credential matches search
- Empty results show message: "No credentials match your search"

### Password Visibility Toggle

In add and edit forms, press `Ctrl+P` to toggle between masked and visible passwords.

**Use Cases**:
- Verify password spelling before saving
- Check for typos when editing existing passwords
- Confirm generated passwords meet requirements

**Behavior**:
- **Default state**: Password masked (asterisks: `******`)
- **After `Ctrl+P`**: Password visible (plaintext), label shows `[VISIBLE]`
- **After `Ctrl+P` again**: Password masked again
- **On form close**: Visibility resets to masked (secure default)
- **Cursor position**: Preserved when toggling (no text loss)

**Examples**:
```bash
# In add form
n                              # Open new credential form
Type: SecureP@ssw0rd!         # Password shows as ******
Ctrl+P                         # Password shows: SecureP@ssw0rd!
Ctrl+P                         # Password shows as ******
Ctrl+S                         # Save (password saved correctly)

# In edit form
e                              # Open edit form for selected credential
Focus password field           # Existing password loads (masked)
Ctrl+P                         # View current password
Type new password              # Update password
Ctrl+P                         # Mask again to verify asterisks
Ctrl+S                         # Save changes
```

**Security Note**: Password visibility is per-form. Switching between add and edit forms resets visibility to masked.

### Layout Controls

The TUI layout adapts to terminal size with manual override controls.

#### Detail Panel Toggle (`i` key)

Cycles through three states:
1. **Auto (responsive)**: Shows on wide terminals (>100 cols), hides on narrow
2. **Force Hide**: Always hidden regardless of terminal width
3. **Force Show**: Always visible regardless of terminal width

Status bar displays current state when toggling:
- "Detail Panel: Auto (responsive)"
- "Detail Panel: Hidden"
- "Detail Panel: Visible"

**Use Cases**:
- Hide detail panel to focus on credential list
- Force show on narrow terminal to view credential details
- Return to auto mode for responsive behavior

#### Sidebar Toggle (`s` key)

Cycles through three states:
1. **Auto (responsive)**: Shows on wide terminals (>80 cols), hides on narrow
2. **Force Hide**: Always hidden regardless of terminal width
3. **Force Show**: Always visible regardless of terminal width

Status bar displays current state when toggling:
- "Sidebar: Auto (responsive)"
- "Sidebar: Hidden"
- "Sidebar: Visible"

**Use Cases**:
- Hide sidebar to maximize table width
- Force show on narrow terminal to access category navigation
- Return to auto mode for responsive behavior

**Manual overrides persist** until user changes them or application restarts.

### Usage Location Display

The detail panel shows where each credential has been accessed.

**Information Displayed**:
- **File path**: Absolute path to working directory where `pass-cli get` was executed
- **Access count**: Number of times credential accessed from that location
- **Timestamp**: Hybrid format (relative for recent, absolute for old)
  - Recent (within 7 days): "2 hours ago", "3 days ago"
  - Older: "2025-09-15", "2024-12-01"
- **Git repository** (if available): Repository name extracted from working directory
- **Line number** (if available): File path with line number (e.g., `/path/file.go:42`)

**Display Format**:
```text
Usage Locations:
  /home/user/projects/web-app
    Accessed: 12 times
    Last: 2 hours ago
    Repo: web-app

  /home/user/projects/api-server/src/config.go:156
    Accessed: 5 times
    Last: 2025-09-20
    Repo: api-server
```

**Empty State**: If credential has never been accessed, shows: "No usage recorded"

**Sorting**: Locations sorted by most recent access timestamp descending.

**Use Cases**:
- Audit which projects use which credentials
- Identify stale credentials not accessed recently
- Track credential usage patterns across repositories
- Understand credential dependencies for project cleanup

### Exiting TUI Mode

Press `q` or `Ctrl+C` at any time to quit the TUI and return to shell.

**Note**: If a modal is open (add form, edit form, help), pressing `q` or `Esc` closes the modal instead of quitting. Press `q` again from main view to quit application.

## TUI Configuration

The TUI appearance and behavior can be customized via `~/.pass-cli/config.yml`.

### Theme Configuration

Pass-CLI TUI supports multiple color themes. Available themes:

#### Dracula (Default)
Dark theme with vibrant purples, pinks, and cyans. Perfect for low-light environments.
- **Background**: Deep dark purple (#282a36)
- **Accents**: Cyan, pink, purple
- **Status**: Green (success), red (error), yellow (warning)

#### Nord
Cool, bluish theme inspired by arctic ice and polar nights.
- **Background**: Dark blue-gray (#2e3440)
- **Accents**: Frost blues and teals
- **Status**: Muted greens, reds, and yellows

#### Gruvbox
Warm, retro theme with earthy tones and high contrast.
- **Background**: Dark gray-brown (#282828)
- **Accents**: Warm aqua, yellow, orange
- **Status**: Vibrant greens, reds, yellows

#### Monokai
Vibrant, colorful theme popular in code editors.
- **Background**: Very dark gray (#272822)
- **Accents**: Bright cyan, purple, yellow
- **Status**: Neon greens, hot pinks, bright yellows

**Configuration:**
```yaml
# Valid themes: dracula, nord, gruvbox, monokai
theme: "nord"
```

**Changing Themes:**
1. Edit config file:
   ```bash
   # macOS/Linux
   nano ~/.pass-cli/config.yml

   # Windows (PowerShell)
   notepad $env:USERPROFILE\.pass-cli\config.yml
   ```

2. Set theme:
   ```yaml
   theme: "nord"  # or dracula, gruvbox, monokai
   ```

3. Restart TUI:
   ```bash
   pass-cli
   ```

**Validation**: If you specify an invalid theme name, Pass-CLI will show a warning, fall back to Dracula, and continue running normally.

### Detail Panel Configuration

The detail panel position can adapt to terminal width for optimal viewing experience.

**Configuration Options:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `detail_position` | string | `"auto"` | Detail panel positioning: `auto`, `right`, or `bottom` |
| `detail_auto_threshold` | int | `120` | Width threshold (columns) for auto-positioning (80-500) |

**Position Modes:**

**Auto Mode** (`detail_position: "auto"`)
- Terminal ≥ threshold: Detail panel on right (traditional horizontal layout)
- Terminal 80-119: Detail panel on bottom (vertical layout)
- Terminal < 80: Detail panel hidden

**Right Mode** (`detail_position: "right"`)
- Terminal ≥ 120: Detail panel on right
- Terminal < 120: Detail panel hidden
- Best for wide terminals, traditional layout

**Bottom Mode** (`detail_position: "bottom"`)
- Terminal ≥ 80: Detail panel always on bottom
- Terminal < 80: Detail panel hidden
- Best for narrow terminals, maximizes horizontal space

**Configuration Example:**
```yaml
terminal:
  warning_enabled: true
  min_width: 60
  min_height: 30
  detail_position: "auto"          # or "right", "bottom"
  detail_auto_threshold: 120       # Width threshold for auto mode
```

**Use Cases:**
- **Auto mode**: Best for users who frequently resize terminal or use different displays
- **Right mode**: Best for users with consistently wide terminals (≥120 columns)
- **Bottom mode**: Best for users who prefer vertical layouts or narrow terminals

**Threshold Tuning**: Adjust `detail_auto_threshold` based on your display preferences. Lower values (80-100) switch to vertical layout sooner, higher values (120-150) prefer horizontal layout longer.

## TUI Best Practices

1. **Use `/` search for large vaults** - Faster than scrolling through 50+ credentials
2. **Press `?` to see all shortcuts** - Built-in help always available
3. **Toggle detail panel (`i`) on narrow terminals** - Maximize table visibility
4. **Use `Ctrl+P` to verify passwords** - Catch typos before saving
5. **Check usage locations before deleting** - Understand credential dependencies
6. **Press `c` to copy passwords** - Clipboard auto-clears after 5 seconds

## TUI Troubleshooting

**Problem**: TUI doesn't launch, shows "command not found"
**Solution**: Ensure you're running `pass-cli` with no arguments. If you pass any argument (even invalid), it attempts CLI mode.

**Problem**: Ctrl+P does nothing in forms
**Solution**: Ensure you're in add or edit form, not the main view. Password toggle only works in forms.

**Problem**: Search key `/` types "/" character instead of activating search
**Solution**: Ensure focus is on the main view (table/sidebar), not inside a form or modal. Press `Esc` to close any open modal first.

**Problem**: Sidebar doesn't appear
**Solution**: Press `s` to toggle sidebar. On narrow terminals (<80 cols), sidebar auto-hides in responsive mode. Press `s` twice to force show.

**Problem**: Usage locations not showing
**Solution**: Usage locations only appear after you've accessed credentials via `pass-cli get <service>` from different working directories. New credentials won't have usage data until first access.

