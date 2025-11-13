---
title: "Command Reference"
weight: 1
bookToc: true
---

# Command Reference
Complete reference for all pass-cli commands and their options.

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
| `--gen-length` | | int | Length of generated password (default: 20) |
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
| `--gen-length` | | int | Length of generated password (default: 20) |
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

##### vault backup

Manage vault backups for disaster recovery.

**Synopsis:**
```bash
pass-cli vault backup <subcommand>
```

###### vault backup create

Create a timestamped manual backup of the vault.

**Synopsis:**
```bash
pass-cli vault backup create [flags]
```

**Description:**
Creates a manual backup with naming pattern `vault.enc.YYYYMMDD-HHMMSS.manual.backup`. Works without requiring the master password (no vault unlock needed).

**Flags:**

| Flag | Type | Description |
|------|------|-------------|
| `-v`, `--verbose` | bool | Show detailed operation progress |

**Examples:**
```bash
# Create manual backup
pass-cli vault backup create

# Create backup with verbose output
pass-cli vault backup create --verbose
```

**Output:**
```
✅ Manual backup created successfully:
   /home/user/.pass-cli/vault.enc.20251112-143022.manual.backup
   Size: 2.45 MB
```

###### vault backup restore

Restore vault from the most recent backup.

**Synopsis:**
```bash
pass-cli vault backup restore [flags]
```

**Description:**
Automatically selects the newest valid backup (automatic or manual) and restores it. Considers both `vault.enc.backup` (automatic) and `vault.enc.*.manual.backup` files.

**⚠️ WARNING:** This command overwrites your current vault file with the backup.

**Flags:**

| Flag | Type | Description |
|------|------|-------------|
| `-f`, `--force` | bool | Skip confirmation prompt |
| `-v`, `--verbose` | bool | Show detailed operation progress |
| `--dry-run` | bool | Preview which backup would be restored (no changes) |

**Examples:**
```bash
# Restore from newest backup (with confirmation)
pass-cli vault backup restore

# Restore without confirmation
pass-cli vault backup restore --force

# Preview which backup would be restored
pass-cli vault backup restore --dry-run

# Restore with detailed progress
pass-cli vault backup restore --verbose
```

**Output:**
```
Found backup: /home/user/.pass-cli/vault.enc.20251112-143022.manual.backup
Backup age: 2 hours ago
Size: 2.45 MB

⚠️  This will overwrite your current vault file.
Are you sure you want to restore from this backup? (y/n): y

✅ Vault restored successfully from backup
```

###### vault backup info

View backup status and information.

**Synopsis:**
```bash
pass-cli vault backup info [flags]
```

**Description:**
Displays all available backups with metadata:
- Backup type (automatic or manual)
- File size and creation time
- Age with warnings for backups >30 days old
- Which backup would be used for restore
- Disk space usage alerts (>5 manual backups)

**Flags:**

| Flag | Type | Description |
|------|------|-------------|
| `-v`, `--verbose` | bool | Show detailed backup information |

**Examples:**
```bash
# View all backups
pass-cli vault backup info

# View with detailed information
pass-cli vault backup info --verbose
```

**Output:**
```
📦 Backup Status for: /home/user/.pass-cli/vault.enc

Automatic Backup:
  ✅ vault.enc.backup
     Size: 2.45 MB
     Created: 1 day ago (2025-11-11 14:30:22)

Manual Backups:
  ✅ vault.enc.20251112-143022.manual.backup ← Would be used for restore
     Size: 2.45 MB
     Created: 2 hours ago (2025-11-12 14:30:22)

  ✅ vault.enc.20251110-091545.manual.backup
     Size: 2.40 MB
     Created: 2 days ago (2025-11-10 09:15:45)

Total backups: 3
Total disk space: 7.30 MB
```

**See Also:**
- {{< relref "../02-guides/backup-restore" >}} - Comprehensive backup guide

**Use Cases:**
- Decommissioning a vault that's no longer needed
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

See {{< relref "../05-operations/health-checks" >}} for detailed documentation and troubleshooting.

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

See {{< relref "../01-getting-started/quick-start" >}} for complete first-run documentation.

## Getting Help

- Run any command with `--help` flag
- See [pass-cli Documentation](https://ari1110.github.io/pass-cli/) for overview
- Check {{< relref "../04-troubleshooting/_index" >}} for common issues
- Visit [GitHub Issues](https://github.com/ari1110/pass-cli/issues)

