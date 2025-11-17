---
title: "Basic Workflows"
weight: 1
toc: true
---

Common daily operations with pass-cli: listing, updating, deleting, and generating credentials.

## Retrieving Credentials

Retrieve it:

```bash
$ pass-cli get github

Username: your-github-username
Password: your-github-password

Last accessed: 2025-10-21 14:23:45
Access count: 1
```

## Basic Operations

### List All Credentials

```bash
pass-cli list
```

Output:
```
Stored Credentials
==================

github
  Username: your-github-username
  Last accessed: 2025-10-21 14:23:45
  Access count: 1

aws
  Username: your-aws-username
  Last accessed: 2025-10-20 09:15:22
  Access count: 5

Total: 2 credentials
```

### Update a Credential

```bash
pass-cli update github
```

Prompts for new username and password.

### Delete a Credential

```bash
pass-cli delete github
```

Prompts for confirmation before deletion.

### Generate a Strong Password

Pass-CLI can generate strong, random passwords in multiple ways:

**Standalone Generation:**
```bash
$ pass-cli generate

Generated password (16 characters):
kR9$mN2@pL5#wQ8!

# Generate with custom length
$ pass-cli generate --length 24

# Copy to clipboard (macOS)
$ pass-cli generate | pbcopy
```

**Generate During Add:**
```bash
# Generate password while adding credential (16 characters)
$ pass-cli add github -u myuser --generate

# Generate with custom length
$ pass-cli add github -u myuser --generate --gen-length 32

✓ Generated 32-character password
✓ Credential 'github' added successfully
```

**Generate During Update (Password Rotation):**
```bash
# Generate new password while updating (16 characters)
$ pass-cli update github --generate

# Generate with custom length
$ pass-cli update github --generate --gen-length 24

✓ Generated 24-character password
✓ Credential 'github' updated successfully
```

**TUI Generation:**
When adding or editing credentials in the TUI, press `Ctrl+G` in the password field to generate a random password instantly.

## Keychain Integration

pass-cli can store your master password in your OS keychain for convenience, eliminating the need to type it for every operation.

### Enable Keychain Integration

If you didn't enable keychain during initialization, you can enable it anytime:

```bash
$ pass-cli keychain enable

Master password: ••••••••••••
✅ Keychain integration enabled for vault at /home/user/.pass-cli/vault.enc

Future commands will not prompt for password when keychain is available.
```

### Check Keychain Status

View the current keychain integration status:

```bash
$ pass-cli keychain status

Keychain Status for /home/user/.pass-cli/vault.enc:

✓ System Keychain:        Available (macOS Keychain)
✓ Password Stored:        Yes
✓ Backend:                keychain
✓ Vault Configuration:    Keychain enabled

✓ Keychain integration is properly configured.
Your vault password is securely stored in the system keychain.
Future commands will not prompt for password.
```

**If keychain is not enabled:**
```bash
$ pass-cli keychain status

Keychain Status for /home/user/.pass-cli/vault.enc:

✓ System Keychain:        Available (Windows Credential Manager)
✗ Password Stored:        No
✓ Vault Configuration:    Keychain not enabled

The system keychain is available but no password is stored for this vault.
Suggestion: Enable keychain integration with 'pass-cli keychain enable'
```

### Disable Keychain Integration

To remove your master password from the keychain, use your operating system's credential manager:

**Windows**: `cmdkey /delete:pass-cli` (or use Credential Manager GUI)
**macOS**: `security delete-generic-password -s "pass-cli" -a "$USER"` (or use Keychain Access app)
**Linux**: `secret-tool clear service pass-cli` (or use your DE's credential manager)

After removing the keychain entry, you'll need to enter your master password for each operation.

See [Keychain Setup](keychain-setup) for detailed platform-specific instructions.

### Platform-Specific Backends

pass-cli integrates with your operating system's secure credential storage:

- **Windows**: Windows Credential Manager
- **macOS**: macOS Keychain
- **Linux**: Secret Service API (gnome-keyring/kwallet)

### TUI Auto-Unlock

When keychain integration is enabled, the TUI (Terminal User Interface) automatically unlocks your vault without prompting for a password:

```bash
pass-cli tui  # Opens directly, no password prompt
```

## Script-Friendly Usage

### Quiet Mode

Suppress prompts and output only the credential value:

```bash
# Get only the password field
export DB_PASSWORD=$(pass-cli get database --quiet --field password)

# Get only the username field
export DB_USER=$(pass-cli get database --quiet --field username)

# Use in scripts
#!/bin/bash
API_KEY=$(pass-cli get api-service --quiet --field password)
curl -H "Authorization: Bearer $API_KEY" https://api.example.com/data
```

### JSON Output

```bash
$ pass-cli get github --json

{
  "key": "github",
  "username": "your-github-username",
  "password": "your-github-password",
  "metadata": {
    "last_accessed": "2025-10-21T14:23:45Z",
    "access_count": 1
  }
}
```

Process with `jq`:

```bash
# Extract specific field
pass-cli get github --json | jq -r '.username'

# Use in scripts
username=$(pass-cli get github --json | jq -r '.username')
password=$(pass-cli get github --json | jq -r '.password')
```

## Health Checks

Verify your pass-cli installation is working correctly:

```bash
$ pass-cli doctor

Health Check Results
====================

✓ Version: v1.2.3 (up to date)
✓ Vault: vault.enc accessible (600 permissions)
✓ Config: Valid configuration
✓ Keychain: Integration active
✓ Backup: 3 backup files found

Overall Status: HEALTHY
```

See [Health Checks](../05-operations/health-checks) for detailed health check documentation.

### Common First-Time Issues

#### Keychain Access Denied (macOS)

**Symptom**:
```
⚠ Keychain: Access denied by OS
```
