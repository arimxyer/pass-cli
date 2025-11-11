# Usage Guide
Complete reference for all Pass-CLI commands, flags, and features.

![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?path=docs&label=Last%20Updated)

## Table of Contents

- [Global Options](#global-options)
- [Commands](#commands)
  - [init](#init---initialize-vault)
  - [add](#add---add-credential)
  - [get](#get---retrieve-credential)
  - [list](#list---list-credentials)
  - [update](#update---update-credential)
  - [delete](#delete---delete-credential)
  - [generate](#generate---generate-password)
  - [keychain](#keychain---manage-keychain-integration)
  - [vault](#vault---manage-vault-files)
  - [usage](#usage---view-credential-usage-history)
  - [version](#version---show-version)
- [Output Modes](#output-modes)
- [Script Integration](#script-integration)
- [Environment Variables](#environment-variables)
- [Configuration](#configuration)
- [Usage Tracking](#usage-tracking)
- [Best Practices](#best-practices)

## Global Options

Available for all commands:

| Flag | Description | Example |
|------|-------------|---------|
| `--verbose` | Enable verbose output | `--verbose` |
| `--help`, `-h` | Show help | `--help` |

### Global Flag Examples

```bash
# Enable verbose logging
pass-cli --verbose get github

# Get help for any command
pass-cli get --help
```

### Custom Vault Location

To use a custom vault location, configure it in your config file (`~/.pass-cli/config.yml`):

```yaml
vault_path: /custom/path/vault.enc
```

See [Configuration](#configuration) section for details on path expansion (environment variables, tilde, relative paths).

## Commands

### init - Initialize Vault

Create a new password vault.

#### Synopsis

```bash
pass-cli init
```

#### Description

Creates a new encrypted vault at `~/.pass-cli/vault.enc` and stores the master password in your system keychain. You will be prompted to create a master password.

#### Examples

```bash
# Initialize with default location
pass-cli init

# For custom vault location, configure in config file first:
# Edit ~/.pass-cli/config.yml and add: vault_path: /custom/path/vault.enc
# Then run: pass-cli init
```

#### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--enable-audit` | bool | Enable tamper-evident audit logging |
| `--use-keychain` | bool | Store master password in OS keychain |

#### Password Policy (January 2025)

All master passwords must meet complexity requirements:
- **Minimum Length**: 12 characters
- **Uppercase**: At least one uppercase letter (A-Z)
- **Lowercase**: At least one lowercase letter (a-z)
- **Digit**: At least one digit (0-9)
- **Symbol**: At least one special symbol (!@#$%^&*()-_=+[]{}|;:,.<>?)

**Examples**:
- ✅ `MySecureP@ssw0rd2025!` (meets all requirements)
- ✅ `Correct-Horse-Battery-29!` (meets all requirements)
- ❌ `password123` (too short, no uppercase, no symbol)
- ❌ `MyPassword` (no digit, no symbol)

#### Audit Logging (Optional)

Enable audit logging to record vault operations with HMAC signatures:

```bash
# Initialize vault with audit logging
pass-cli init --enable-audit
```

**Audit features**:
- **Tamper-Evident**: HMAC-SHA256 signatures prevent log modification
- **Privacy**: Service names logged, passwords NEVER logged
- **Key Storage**: HMAC keys stored in OS keychain (separate from vault)
- **Auto-Rotation**: Logs rotate at 10MB with 7-day retention
- **Graceful Degradation**: Operations continue if logging fails

**Verification**:
```bash
# Verify audit log integrity
pass-cli verify-audit
```

#### Notes

- Master password must meet complexity requirements (12+ chars, uppercase, lowercase, digit, symbol)
- Strong passwords (20+ characters) recommended for master password
- Master password is stored in OS keychain for convenience
- Vault file is created with restricted permissions (0600)
- Audit logging is opt-in (disabled by default)

---

### add - Add Credential

Add a new credential to the vault.

#### Synopsis

```bash
pass-cli add <service> [flags]
```

#### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--username` | `-u` | string | Username for the credential |
| `--password` | `-p` | string | Password (not recommended, use prompt) |
| `--generate` | `-g` | bool | Generate a random secure password |
| `--gen-length` | | int | Length of generated password (default: 16) |
| `--category` | `-c` | string | Category for organizing credentials (e.g., 'Cloud', 'Databases') |
| `--url` | | string | Service URL |
| `--notes` | | string | Additional notes |

#### Examples

```bash
# Interactive mode (prompts for username/password)
pass-cli add github

# With username flag
pass-cli add github --username user@example.com

# With URL and notes
pass-cli add github \
  --username user@example.com \
  --url https://github.com \
  --notes "Personal account"

# With category
pass-cli add github -u user@example.com -c "Version Control"

# Generate random password (16 characters)
pass-cli add github -u user@example.com --generate

# Generate random password with custom length
pass-cli add github -u user@example.com --generate --gen-length 24

# Generate password with other metadata
pass-cli add github \
  -u user@example.com \
  --generate \
  --gen-length 20 \
  --url https://github.com \
  --notes "Work account"

# All flags (not recommended for password)
pass-cli add github \
  -u user@example.com \
  -p secret123 \
  --url https://github.com \
  --notes "Work account"
```

#### Interactive Prompts

When not using flags, you'll be prompted:

```
Enter username: user@example.com
Enter password: ******* (hidden input)
Enter URL (optional): https://github.com
Enter notes (optional): Personal account
```

#### Password Policy

Credential passwords must meet the same complexity requirements as master passwords:
- Minimum 12 characters with uppercase, lowercase, digit, and symbol
- TUI mode shows real-time strength indicator
- Generated passwords automatically meet policy requirements

#### Notes

- Service names must be unique
- Password input is hidden by default
- Passing password via `-p` flag is insecure (visible in shell history)
- Use `pass-cli generate` to create strong random passwords that meet policy requirements
- Usage tracking begins when credential is first accessed

---

### get - Retrieve Credential

Retrieve a credential from the vault.

#### Synopsis

```bash
pass-cli get <service> [flags]
```

#### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--quiet` | `-q` | bool | Output password only (for scripts) |
| `--field` | `-f` | string | Extract specific field |
| `--no-clipboard` | | bool | Skip clipboard copy |
| `--masked` | | bool | Display password as asterisks |

#### Field Options

For `--field` flag:
- `username` - User's username
- `password` - User's password
- `url` - Service URL
- `notes` - Additional notes
- `service` - Service name
- `created` - Creation timestamp
- `modified` - Last modified timestamp
- `accessed` - Last accessed timestamp

#### Examples

```bash
# Default: Display credential and copy to clipboard
pass-cli get github

# Quiet mode (password only, for scripts)
pass-cli get github --quiet
pass-cli get github -q

# Get specific field
pass-cli get github --field username
pass-cli get github -f url

# Quiet mode with specific field
pass-cli get github --field username --quiet

# Display without clipboard
pass-cli get github --no-clipboard

# Display with masked password
pass-cli get github --masked
```

#### Output Examples

**Default output:**
```
Service:  github
Username: user@example.com
Password: mySecretPassword123!
URL:      https://github.com
Notes:    Personal account

✓ Password copied to clipboard (will clear in 5 seconds)
```

**Quiet mode:**
```bash
$ pass-cli get github --quiet
mySecretPassword123!
```

**Field extraction:**
```bash
$ pass-cli get github --field username --quiet
user@example.com
```

#### Notes

- Clipboard auto-clears after 5 seconds
- Usage tracking records current directory
- Accessing a credential updates the "last accessed" timestamp

---

### list - List Credentials

List all credentials in the vault.

#### Synopsis

```bash
pass-cli list [flags]
```

#### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--format` | string | Output format: table, json, simple (default: table) |
| `--unused` | bool | Show only unused credentials |
| `--days` | int | Days threshold for unused (default: 30) |

#### Examples

```bash
# List all credentials (table format)
pass-cli list

# List as JSON
pass-cli list --format json

# Simple list (service names only)
pass-cli list --format simple

# Show unused credentials (not accessed in 30 days)
pass-cli list --unused

# Show credentials not used in 90 days
pass-cli list --unused --days 90
```

#### Output Examples

**Table format (default):**
```
+----------+----------------------+---------------------+
| SERVICE  | USERNAME             | LAST ACCESSED       |
+----------+----------------------+---------------------+
| github   | user@example.com     | 2025-01-20 14:22    |
| aws-prod | admin@company.com    | 2025-01-18 09:15    |
| database | dbuser               | 2025-01-15 16:30    |
+----------+----------------------+---------------------+
```

**Simple format:**
```
github
aws-prod
database
```

#### Notes

- Passwords are never shown in list output
- Table format is best for human viewing
- JSON format is best for parsing
- Simple format is best for shell scripts

---

### update - Update Credential

Update an existing credential.

#### Synopsis

```bash
pass-cli update <service> [flags]
```

#### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--username` | `-u` | string | New username |
| `--password` | `-p` | string | New password (not recommended) |
| `--generate` | `-g` | bool | Generate a random secure password |
| `--gen-length` | | int | Length of generated password (default: 16) |
| `--category` | | string | New category |
| `--url` | | string | New URL |
| `--notes` | | string | New notes |
| `--clear-category` | | bool | Clear category field to empty |
| `--clear-notes` | | bool | Clear notes field to empty |
| `--clear-url` | | bool | Clear URL field to empty |
| `--force` | `-f` | bool | Skip confirmation prompt |

#### Examples

```bash
# Update password (prompted)
pass-cli update github

# Update username
pass-cli update github --username newuser@example.com

# Update URL
pass-cli update github --url https://github.com/enterprise

# Update notes
pass-cli update github --notes "Updated account info"

# Update category
pass-cli update github --category "Work"

# Generate new random password (16 characters)
pass-cli update github --generate

# Generate new password with custom length
pass-cli update github --generate --gen-length 32

# Generate password and update other fields
pass-cli update github \
  --generate \
  --gen-length 24 \
  --notes "Password rotated on 2025-11-11"

# Clear category field
pass-cli update github --clear-category

# Update multiple fields
pass-cli update github \
  --username newuser@example.com \
  --url https://github.com/enterprise \
  --notes "Corporate account"
```

#### Interactive Mode

If no flags provided, prompts for password:

```
Enter new password (leave blank to keep current): *******
Password updated successfully!
```

#### Notes

- At least one field must be updated
- Updating password clears usage history
- Original values preserved if not specified

---

### delete - Delete Credential

Delete a credential from the vault.

#### Synopsis

```bash
pass-cli delete <service> [flags]
```

#### Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--force` | `-f` | bool | Skip confirmation prompt |

#### Examples

```bash
# Delete with confirmation
pass-cli delete github

# Force delete (no confirmation)
pass-cli delete github --force
pass-cli delete github -f
```

#### Interactive Confirmation

Without `--force`:

```
Are you sure you want to delete 'github'? (yes/no): yes
Credential 'github' deleted successfully!
```

#### Notes

- Deletion is permanent (no undo)
- Confirmation required unless using `--force`
- Credential completely removed from vault

---

### generate - Generate Password

Generate a cryptographically secure password.

#### Synopsis

```bash
pass-cli generate [flags]
```

#### Aliases

`gen`, `pwd`

#### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--length` | int | Password length (8-128, default: 20) |
| `--no-lower` | bool | Exclude lowercase letters |
| `--no-upper` | bool | Exclude uppercase letters |
| `--no-digits` | bool | Exclude digits |
| `--no-symbols` | bool | Exclude symbols |
| `--no-clipboard` | bool | Skip clipboard copy |

#### Examples

```bash
# Generate default password (20 chars, all character types)
pass-cli generate

# Custom length
pass-cli generate --length 32

# Alphanumeric only (no symbols)
pass-cli generate --no-symbols

# Digits and symbols only
pass-cli generate --no-lower --no-upper

# Letters only (no digits or symbols)
pass-cli generate --no-digits --no-symbols

# Display only (no clipboard)
pass-cli generate --no-clipboard
```

#### Character Sets

Default character sets:
- Lowercase: `a-z`
- Uppercase: `A-Z`
- Digits: `0-9`
- Symbols: `!@#$%^&*()_+-=[]{}|;:,.<>?`

#### Notes

- Uses `crypto/rand` for cryptographic randomness
- At least one character set must be enabled
- Minimum length: 8 characters
- Maximum length: 128 characters
- Clipboard auto-clears after 5 seconds

---

### version - Show Version

Display version information.

#### Synopsis

```bash
pass-cli version [flags]
```

#### Flags

| Flag | Type | Description |
|------|------|-------------|
| `--verbose` | bool | Show detailed version info |

#### Examples

```bash
# Show version
pass-cli version

# Verbose version info
pass-cli version --verbose
```

#### Output Examples

**Default:**
```
pass-cli version X.Y.Z
```

**Verbose:**
```
pass-cli version X.Y.Z
  commit: abc123f
  built:  2025-01-20T10:30:00Z
  go:     go1.25.1
```

### keychain - Manage Keychain Integration

Manage system keychain integration for storing vault master passwords.

#### Synopsis

```bash
pass-cli keychain <subcommand>
```

#### Subcommands

##### keychain enable

Enable keychain integration for an existing vault by storing the master password in the system keychain.

**Synopsis:**
```bash
pass-cli keychain enable [flags]
```

**Description:**
Stores your vault master password in the OS keychain (Windows Credential Manager, macOS Keychain, or Linux Secret Service). Future commands will not prompt for password when keychain is available. This is useful for vaults created without the `--use-keychain` flag.

**Flags:**

| Flag | Type | Description |
|------|------|-------------|
| `--force` | bool | Overwrite existing keychain entry if already enabled |

**Examples:**
```bash
# Enable keychain for default vault
pass-cli keychain enable

# For custom vault location, configure vault_path in ~/.pass-cli/config.yml

# Force overwrite existing keychain entry
pass-cli keychain enable --force
```

**Output:**
```
Master password: ********

✅ Keychain integration enabled for vault at /home/user/.pass-cli/vault.enc

Future commands will not prompt for password when keychain is available.
```

##### keychain status

Display keychain integration status for the current vault.

**Synopsis:**
```bash
pass-cli keychain status [flags]
```

**Description:**
Shows keychain availability, password storage status, and backend name. This is a read-only operation that doesn't require unlocking the vault.

**Examples:**
```bash
# Check keychain status for default vault
pass-cli keychain status

# For custom vault location, configure vault_path in ~/.pass-cli/config.yml
```

**Output Examples:**

**When keychain is enabled:**
```
Keychain Status for /home/user/.pass-cli/vault.enc:

✓ System Keychain:        Available (keychain)
✓ Password Stored:        Yes
✓ Backend:                keychain

Your vault password is securely stored in the system keychain.
Future commands will not prompt for password.
```

**When keychain is available but not enabled:**
```
Keychain Status for /home/user/.pass-cli/vault.enc:

✓ System Keychain:        Available (wincred)
✗ Password Stored:        No

The system keychain is available but no password is stored for this vault.
Run 'pass-cli keychain enable' to store your password and skip future prompts.
```

**When keychain is not available:**
```
Keychain Status for /home/user/.pass-cli/vault.enc:

✗ System Keychain:        Not available on this platform
✗ Password Stored:        N/A

System keychain is not accessible. You will be prompted for password on each command.
```

#### Platform Support

| Platform | Backend | Service Name |
|----------|---------|--------------|
| Windows | wincred | Windows Credential Manager |
| macOS | keychain | Keychain Access |
| Linux | gnome-keyring/kwallet | Secret Service API |

### vault - Manage Vault Files

Manage pass-cli vault files and their lifecycle.

#### Synopsis

```bash
pass-cli vault <subcommand>
```

#### Subcommands

##### vault remove

Permanently delete a vault file and its associated keychain entry.

**Synopsis:**
```bash
pass-cli vault remove <path> [flags]
```

**Description:**
Permanently deletes:
1. The vault file from disk
2. The master password from the system keychain
3. Any orphaned keychain entries

**⚠️ WARNING:** This operation is irreversible. All stored credentials will be lost. Ensure you have backups before proceeding.

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `<path>` | Yes | Path to vault file to remove |

**Flags:**

| Flag | Type | Description |
|------|------|-------------|
| `-y`, `--yes` | bool | Skip confirmation prompt (for automation) |
| `-f`, `--force` | bool | Force removal even if vault appears in use |

**Examples:**
```bash
# Remove vault with confirmation prompt
pass-cli vault remove /path/to/vault.enc

# Remove vault without confirmation (for automation)
pass-cli vault remove /path/to/vault.enc --yes

# Force removal even if file appears in use
pass-cli vault remove /path/to/vault.enc --force
```

**Output:**
```
⚠️  WARNING: This will permanently delete the vault and all stored credentials.
Are you sure you want to remove /home/user/.pass-cli/vault.enc? (y/n): y

✅ Vault removed successfully:
   • Vault file deleted
   • Keychain entry removed
   • Orphaned entries cleaned up
```

**Use Cases:**
- Decommissioning a vault that's no longer needed
- Cleaning up test vaults
- Removing vaults with forgotten passwords (data loss)
- Automated testing and CI/CD cleanup

### usage - View Credential Usage History

Display detailed usage history for a specific credential across all locations.

#### Synopsis

```bash
pass-cli usage <service> [flags]
```

**Description:**
Shows where and when a credential was accessed, including:
- Location paths (working directories)
- Git repository context
- Last access timestamps
- Access counts per location
- Field-level usage breakdown (which fields were accessed)

This helps you understand which projects use specific credentials, audit access patterns, and identify unused credentials.

#### Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<service>` | Yes | Name of the credential to view usage for |

#### Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `--format` | string | Output format: table, json, simple | table |
| `--limit` | int | Maximum number of locations to display (0 = unlimited) | 20 |

#### Examples

```bash
# View usage history (default: table format, 20 most recent locations)
pass-cli usage github

# View all locations (no limit)
pass-cli usage aws --limit 0

# View only 5 most recent locations
pass-cli usage postgres --limit 5

# JSON output for scripting
pass-cli usage heroku --format json

# Simple format (location paths only)
pass-cli usage redis --format simple
```

#### Output Examples

**Table Format (default):**
```
Credential: github
Total Access Count: 47
Locations: 3

Location                              Repository      Last Used        Count   Fields
─────────────────────────────────────────────────────────────────────────────────────
/home/user/projects/web-app           my-web-app      2 hours ago      25      password:20, username:5
/home/user/projects/api               my-api          5 days ago       15      password:15
/home/user/scripts                    -               2 weeks ago      7       password:7
```

**JSON Format:**
```json
{
  "credential": "github",
  "total_count": 47,
  "locations": [
    {
      "path": "/home/user/projects/web-app",
      "repository": "my-web-app",
      "last_accessed": "2025-10-30T14:25:00Z",
      "count": 25,
      "fields": {
        "password": 20,
        "username": 5
      },
      "path_exists": true
    }
  ]
}
```

**Simple Format:**
```
/home/user/projects/web-app
/home/user/projects/api
/home/user/scripts
```

#### Use Cases

- **Audit before rotation:** Check which projects use a credential before changing it
- **Discover project dependencies:** Find all projects that depend on specific credentials
- **Track access patterns:** Monitor when and where credentials are being used
- **Find unused credentials:** Identify credentials that haven't been accessed recently
- **Multi-project organization:** Understand how credentials are organized across projects

#### Related Commands

- `pass-cli list --location <path>` - Filter credentials by location
- `pass-cli list --by-project` - Group credentials by repository
- `pass-cli list --unused --days 30` - Find credentials not used in 30 days

---

## Output Modes

Pass-CLI supports multiple output modes for different use cases.

### Human-Readable (Default)

Formatted tables and colored output for terminal viewing.

```bash
pass-cli get github
# Service:  github
# Username: user@example.com
# Password: ****** (or full password)
```

### Quiet Mode

Single-line output, perfect for scripts.

```bash
pass-cli get github --quiet
# mySecretPassword123!

pass-cli get github --field username --quiet
# user@example.com
```

### Simple Mode (List Only)

Service names only, one per line.

```bash
pass-cli list --format simple
# github
# aws-prod
# database
```

## Script Integration

### Bash/Zsh Examples

**Export to environment variable:**

```bash
#!/bin/bash

# Export password
export SERVICE_PASSWORD=$(pass-cli get testservice --quiet)

# Export specific field
export SERVICE_USER=$(pass-cli get testservice --field username --quiet)

# Use in command
mysql -u "$(pass-cli get testservice -f username -q)" \
      -p"$(pass-cli get testservice -q)" \
      mydb
```

**Conditional execution:**

```bash
# Check if credential exists
if pass-cli get testservice --quiet &>/dev/null; then
    echo "Credential exists"
    export API_KEY=$(pass-cli get testservice --quiet)
else
    echo "Credential not found"
    exit 1
fi
```

**Loop through credentials:**

```bash
# Process all credentials
for service in $(pass-cli list --format simple); do
    echo "Processing $service..."
    username=$(pass-cli get "$service" --field username --quiet)
    echo "  Username: $username"
done
```

### PowerShell Examples

**Export to environment variable:**

```powershell
# Export password
$env:SERVICE_PASSWORD = pass-cli get testservice --quiet

# Export specific field
$env:SERVICE_USER = pass-cli get testservice --field username --quiet

# Use in command
$apiKey = pass-cli get github --quiet
Invoke-RestMethod -Uri "https://api.github.com" -Headers @{
    "Authorization" = "Bearer $apiKey"
}
```

**Conditional execution:**

```powershell
# Check if credential exists
try {
    $password = pass-cli get testservice --quiet 2>$null
    Write-Host "Credential exists"
    $env:API_KEY = $password
} catch {
    Write-Host "Credential not found"
    exit 1
}
```

### Python Examples

```python
import subprocess

# Get password only
result = subprocess.run(
    ['pass-cli', 'get', 'github', '--quiet'],
    capture_output=True,
    text=True,
    check=True
)
password = result.stdout.strip()

# Get specific field
result = subprocess.run(
    ['pass-cli', 'get', 'github', '--field', 'username', '--quiet'],
    capture_output=True,
    text=True,
    check=True
)
username = result.stdout.strip()
```

### Makefile Examples

```makefile
.PHONY: deploy
deploy:
	@export AWS_KEY=$$(pass-cli get aws --quiet --field username); \
	export AWS_SECRET=$$(pass-cli get aws --quiet); \
	./deploy.sh

.PHONY: test-db
test-db:
	@DB_URL="postgres://$$(pass-cli get testdb -f username -q):$$(pass-cli get testdb -q)@localhost/testdb" \
	go test ./...
```

## Environment Variables

### PASS_CLI_VERBOSE

Enable verbose logging.

```bash
# Bash
export PASS_CLI_VERBOSE=1
pass-cli get github

# PowerShell
$env:PASS_CLI_VERBOSE = "1"
pass-cli get github
```

**Note**: To use a custom vault location, configure `vault_path` in the config file (`~/.pass-cli/config.yml`) instead of using environment variables. See [Configuration](#configuration) section.

## Configuration

**Configuration Location** (added January 2025):
- **All platforms**: `~/.pass-cli/config.yml`

**Management Commands**:
```bash
# Initialize default config
pass-cli config init

# Edit config in default editor
pass-cli config edit

# Validate config syntax
pass-cli config validate

# Reset to defaults
pass-cli config reset
```

### Example Configuration

```yaml
# Custom vault location (optional)
vault_path: /custom/path/vault.enc  # Supports env vars ($HOME), tilde (~), relative, absolute paths

# TUI theme (optional)
theme: "dracula"  # Valid values: dracula, nord, gruvbox, monokai (default: dracula)

# Terminal display thresholds (TUI mode)
terminal:
  # Enable terminal size warnings (default: true)
  warning_enabled: true
  min_width: 60   # Minimum columns (default: 60)
  min_height: 30  # Minimum rows (default: 30)
  # Detail panel positioning (default: auto)
  detail_position: "auto"  # Valid values: auto, right, bottom
  # Width threshold for auto positioning (default: 120)
  detail_auto_threshold: 120  # Range: 80-500

# Custom keyboard shortcuts (TUI mode)
keybindings:
  quit: "q"                  # Quit application
  add_credential: "a"        # Add new credential
  edit_credential: "e"       # Edit credential
  delete_credential: "d"     # Delete credential
  toggle_detail: "i"         # Toggle detail panel
  toggle_sidebar: "s"        # Toggle sidebar
  help: "?"                  # Show help modal
  search: "/"                # Activate search
  confirm: "enter"           # Confirm actions in forms
  cancel: "esc"                # Cancel actions in forms

# Supported key formats for keybindings:
# - Single letters: a-z
# - Numbers: 0-9
# - Function keys: f1-f12
# Modifiers: ctrl+, alt+, shift+
# Examples: ctrl+q, alt+a, shift+f1
```

### Vault Path Configuration

The `vault_path` config field supports flexible path formats:

**Environment Variables (Unix):**
```yaml
vault_path: $HOME/.pass-cli/vault.enc
vault_path: $HOME/secure/vault.enc
```

**Environment Variables (Windows):**
```yaml
vault_path: %USERPROFILE%\Documents\vault.enc
```

**Tilde Expansion:**
```yaml
vault_path: ~/Dropbox/vault.enc
vault_path: ~/.pass-cli/vault.enc
```

**Relative Paths** (resolved relative to home directory):
```yaml
vault_path: vault.enc  # Resolved to $HOME/vault.enc
```

**Absolute Paths:**
```yaml
vault_path: /custom/absolute/path/vault.enc
```

If `vault_path` is not specified, defaults to `~/.pass-cli/vault.enc`.

### Keybinding Customization

**Configurable Actions**:
- `quit`, `add_credential`, `edit_credential`, `delete_credential`
- `toggle_detail`, `toggle_sidebar`, `help`, `search`

**Hardcoded Shortcuts** (cannot be changed):
- Navigation: Tab, Shift+Tab, ↑/↓, Enter, Esc
- Forms: Ctrl+P, Ctrl+S, Ctrl+C
- Detail view: p, c

**Validation**:
- Duplicate key assignments rejected (conflict detection)
- Unknown actions rejected
- Invalid config shows warning modal, app continues with defaults
- UI hints automatically update to reflect custom keybindings

### Configuration Priority

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file
4. Built-in defaults (lowest priority)

## TUI Mode

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
```
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

## Usage Tracking

Pass-CLI automatically tracks where and when credentials are accessed, enabling powerful organization and discovery features.

### How It Works

Usage data is recorded automatically whenever you access a credential:

```bash
# Access from project directory
cd ~/projects/my-app
pass-cli get testservice

# Usage tracking captures:
# - Location (absolute path to current directory)
# - Git repository (if in a git repo)
# - Timestamp and access count
# - Which fields were accessed
```

### Commands

#### View Detailed Usage: `pass-cli usage <service>`

See all locations where a credential has been accessed:

```bash
# View usage history
pass-cli usage github

# JSON output for scripting
pass-cli usage github --format json

# Limit results
pass-cli usage github --limit 10
```

**Output shows**:
- Location paths where credential was accessed
- Git repository name (if applicable)
- Last access timestamp from each location
- Access count from each location
- Field-level usage (which fields accessed)

#### Group by Project: `pass-cli list --by-project`

Organize credentials by git repository context:

```bash
# Group all credentials by repository
pass-cli list --by-project

# JSON output
pass-cli list --by-project --format json

# Simple format (one line per project)
pass-cli list --by-project --format simple
```

**Output shows**:
- Credentials grouped by git repository
- Ungrouped section for non-git-tracked credentials

#### Filter by Location: `pass-cli list --location <path>`

Find credentials used in a specific directory:

```bash
# Show credentials from current directory
pass-cli list --location .

# Specific path
pass-cli list --location /home/user/projects/web-app

# Include subdirectories
pass-cli list --location /home/user/projects --recursive

# Combine with project grouping
pass-cli list --location ~/work --by-project --recursive
```

### Organizing Credentials by Context

Pass-CLI uses a **single-vault model** where one vault contains all your credentials, organized by usage context rather than separate vaults per project.

**Benefits**:
- **Discover credentials by location**: See which credentials are used in each project
- **Cross-project visibility**: Understand credential reuse across projects
- **Machine-independent organization**: `--by-project` groups by git repo name (works across different machines)
- **Location-aware access**: `--location` filters by directory path (machine-specific)

**Example workflow**:

```bash
# Start working on a project
cd ~/projects/web-app

# Discover which credentials are used here
pass-cli list --location .

# Or see project overview
pass-cli list --by-project

# View detailed usage for a specific credential
pass-cli usage github
```

### Use Cases

- **Credential Auditing**: Before rotating credentials, see all locations where they're used
- **Project Onboarding**: New team member discovers project credentials via `--location`
- **Cross-Project Analysis**: Identify shared credentials with `--by-project`
- **Cleanup**: Find unused credentials with `--unused --days 90`
- **Script Integration**: JSON output for automated credential analysis

### Examples

**Audit before rotation**:
```bash
# See all locations using aws-prod credential
pass-cli usage aws-prod

# Export for team review
pass-cli usage aws-prod --format json > aws-audit.json
```

**Discover project credentials**:
```bash
# Which credentials does this project use?
cd ~/projects/new-project
pass-cli list --location . --recursive
```

**Multi-project workflow**:
```bash
# Overview of all projects and their credentials
pass-cli list --by-project

# Filter to work directory only
pass-cli list --location ~/work --by-project --recursive
```

## Best Practices

### Security

1. **Never pass passwords via flags** - Use prompts or `--generate`
2. **Use quiet mode in scripts** - Prevents logging sensitive data
3. **Clear shell history** - When testing commands with passwords
4. **Use strong master passwords** - 20+ characters recommended

### Workflow

1. **Generate passwords** - Use `--generate` for new credentials
2. **Update regularly** - Rotate credentials periodically
3. **Track usage** - Review unused credentials monthly
4. **Backup vault** - Copy `~/.pass-cli/vault.enc` regularly

### Scripting

1. **Always use `--quiet`** - Clean output for variables
2. **Check exit codes** - Handle errors properly
3. **Use `--field`** - Extract exactly what you need
4. **Redirect stderr** - Control error output

### Examples

**Good:**
```bash
export API_KEY=$(pass-cli get service --quiet 2>/dev/null)
if [ -z "$API_KEY" ]; then
    echo "Failed to get credential" >&2
    exit 1
fi
```

**Bad:**
```bash
# Don't do this - exposes password in process list
pass-cli add service --password mySecretPassword
```

## Common Patterns

### CI/CD Pipeline

```bash
# Retrieve deployment credentials
export DEPLOY_KEY=$(pass-cli get production --quiet)
export DB_PASSWORD=$(pass-cli get prod-db --quiet)

# Run deployment
./deploy.sh
```

### Local Development

```bash
# Set up environment from credentials
export DB_HOST=$(pass-cli get dev-db --field url --quiet)
export DB_USER=$(pass-cli get dev-db --field username --quiet)
export DB_PASS=$(pass-cli get dev-db --quiet)

# Start development server
npm run dev
```

### Credential Rotation

```bash
# Generate new password
NEW_PWD=$(pass-cli generate --length 32 --quiet)

# Update service
pass-cli update testservice --password "$NEW_PWD"

# Use new password
echo "$NEW_PWD" | some-service-update-command
```

## Troubleshooting

### Usage Tracking FAQ

#### Why is my usage data empty?

**Problem**: Running `pass-cli usage <service>` shows "No usage history available"

**Causes**:
- Credential was added but never accessed via `get` command
- Credential was only accessed before usage tracking was implemented
- Usage data cleared or vault migrated

**Solution**:
```bash
# Access the credential once to generate usage data
pass-cli get <service>

# Usage data will now be available
pass-cli usage <service>
```

#### Why doesn't `--location` show any credentials?

**Problem**: Running `pass-cli list --location <path>` returns no results

**Causes**:
- Path doesn't match exactly (unless using `--recursive`)
- Credentials haven't been accessed from that location
- Path is specified incorrectly (relative vs. absolute)

**Solutions**:
```bash
# Use current directory
pass-cli list --location .

# Include subdirectories
pass-cli list --location /path/to/project --recursive

# Check what locations exist
pass-cli list --by-project
pass-cli usage <service>
```

#### Why do I see different paths on different machines?

**Expected Behavior**: Location paths are absolute and machine-specific (e.g., `/home/user/project` on Linux, `C:\Users\user\project` on Windows)

**Solution**: Use `--by-project` for machine-independent view (groups by git repository name):
```bash
# Machine-independent view
pass-cli list --by-project
```

#### What does "Ungrouped" mean in `--by-project` output?

**Explanation**: Credentials accessed from non-git directories are grouped under "Ungrouped"

**Solution**:
- Access credentials from within git repositories to enable project grouping
- Or continue using location-based filtering with `--location`

#### How do I clean up old usage data?

**Problem**: Want to remove usage data for deleted projects or old locations

**Current Limitation**: Usage data is append-only and cannot be selectively deleted

**Workaround**:
- Use `--location` to filter to current directories
- Historical data doesn't affect performance

#### Why does usage show deleted directories?

**Table Format**: Deleted paths are automatically hidden for clean output

**JSON Format**: All paths shown with `path_exists: false` field for complete data

```bash
# Table format (hides deleted)
pass-cli usage <service>

# JSON format (shows all with path_exists field)
pass-cli usage <service> --format json
```

### Doctor and First-Run FAQ

#### How do I know if my vault is healthy?

**Solution**: Run the `doctor` command to check vault health:

```bash
pass-cli doctor
```

The doctor command performs 5 comprehensive health checks:
1. **Version Check**: Compares installed version with latest release
2. **Vault Check**: Verifies file accessibility, permissions, and integrity
3. **Config Check**: Validates configuration syntax and settings
4. **Keychain Check**: Tests OS keychain integration status
5. **Backup Check**: Verifies backup files exist and are accessible

**Exit Codes**:
- `0` = All checks passed (HEALTHY)
- `1` = Warnings detected (review recommended)
- `2` = Errors detected (action required)

**Example Output**:
```
Health Check Results
====================

✓ Version: v1.2.3 (up to date)
✓ Vault: vault.enc accessible (600 permissions)
✓ Config: Valid configuration
✓ Keychain: Integration active
✓ Backup: 3 backup files found

Overall Status: HEALTHY
```

See [DOCTOR_COMMAND.md](DOCTOR_COMMAND.md) for detailed documentation and troubleshooting.

#### Why does doctor report orphaned keychain entries?

**Symptom**: Doctor reports "⚠ Keychain: Orphaned entry detected"

**Causes**:
- Vault file was deleted/moved but keychain entry remains
- Vault path changed but old keychain entry wasn't cleaned up
- Multiple vaults were created and old entries weren't removed

**Impact**: Low - orphaned entries don't affect current vault operations, but clutter the keychain

**Solutions**:

**Option 1: Clean up manually** (macOS):
```bash
# Open Keychain Access
open -a "Keychain Access"

# Search for "pass-cli"
# Delete old/orphaned entries
```

**Option 2: Clean up manually** (Windows):
```powershell
# Open Credential Manager
control /name Microsoft.CredentialManager

# Navigate to "Windows Credentials"
# Remove old "pass-cli" entries
```

**Option 3: Clean up manually** (Linux):
```bash
# List all pass-cli keychain entries
secret-tool search service pass-cli

# Delete specific entry
secret-tool clear service pass-cli vault /old/path/vault.enc
```

**Prevention**: When deleting or moving vaults, remove the keychain entry first:
```bash
# Before deleting vault
pass-cli change-password --no-keychain  # Disables keychain
# OR manually remove from OS keychain
```

#### What if first-run detection doesn't trigger?

**Expected Behavior**: When running vault-requiring commands (`add`, `get`, `list`, `update`, `delete`) for the first time without an existing vault, pass-cli offers guided initialization.

**Scenarios where first-run detection is skipped**:

1. **Vault already exists**:
   ```bash
   # Check if vault exists
   ls ~/.pass-cli/vault.enc
   ```
   **Solution**: First-run detection is not needed - your vault is already set up.

2. **Custom vault configured**:
   ```bash
   # If vault_path is configured in config file
   # First-run detection uses that location
   ```
   **Solution**: Configuration is respected - first-run detection will create vault at configured location

3. **Non-TTY environment** (scripts, pipes, CI/CD):
   ```bash
   # This environment doesn't support interactive prompts
   echo "list" | pass-cli list
   ```
   **Solution**: Initialize vault manually in interactive session first, or use `pass-cli init` explicitly:
   ```bash
   # In CI/CD or scripts (configure vault_path in config file first)
   pass-cli init < password-input.txt
   ```

4. **Command doesn't require vault**:
   ```bash
   # These commands don't trigger first-run detection
   pass-cli version
   pass-cli doctor
   pass-cli help
   ```
   **Solution**: Run a vault-requiring command: `pass-cli list` or `pass-cli init`

**Manual initialization**: If first-run detection doesn't trigger and you need to create a vault:
```bash
pass-cli init
```

This provides the same guided setup as automatic first-run detection.

**Troubleshooting**: If first-run detection should trigger but doesn't:
```bash
# Verify no vault exists
ls ~/.pass-cli/vault.enc

# Check if running in TTY
tty  # Should show /dev/pts/X or similar, not "not a tty"

# Try explicit init
pass-cli init
```

See [GETTING_STARTED.md](GETTING_STARTED.md) for complete first-run documentation.

## Getting Help

- Run any command with `--help` flag
- See [README](../README.md) for overview
- Check [Troubleshooting Guide](TROUBLESHOOTING.md) for common issues
- Visit [GitHub Issues](https://github.com/ari1110/pass-cli/issues)

