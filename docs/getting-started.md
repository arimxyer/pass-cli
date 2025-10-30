# Getting Started with pass-cli

This guide will walk you through your first steps with pass-cli, from installation to managing your first credentials.

## Table of Contents

- [Installation](#installation)
- [First-Run Guided Initialization](#first-run-guided-initialization)
- [Manual Initialization](#manual-initialization)
- [Your First Credential](#your-first-credential)
- [Basic Operations](#basic-operations)
- [Health Checks](#health-checks)
- [Next Steps](#next-steps)

## Installation

See [INSTALLATION.md](INSTALLATION.md) for platform-specific installation instructions (Homebrew, Scoop, or binary download).

After installation, verify pass-cli is available:

```bash
pass-cli version
```

## First-Run Guided Initialization

The easiest way to get started is to let pass-cli guide you through the setup process automatically.

### How It Works

When you run any vault-requiring command (`add`, `get`, `list`, `update`, `delete`) for the first time without an existing vault, pass-cli automatically detects this and offers to guide you through initialization.

### Example: First-Run Experience

```bash
$ pass-cli list

╔════════════════════════════════════════════════════════════╗
║                  Welcome to pass-cli!                       ║
╚════════════════════════════════════════════════════════════╝

This appears to be your first time using pass-cli.
Would you like to create a new vault? (y/n): y

Great! Let's set up your secure password vault.

┌────────────────────────────────────────────────────────────┐
│ Step 1: Master Password                                    │
└────────────────────────────────────────────────────────────┘

Your master password encrypts all credentials in your vault.

Password Requirements:
  • Minimum 12 characters
  • At least one uppercase letter
  • At least one lowercase letter
  • At least one number
  • At least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

Enter master password: ••••••••••••••
Confirm master password: ••••••••••••••

✓ Password meets all requirements

┌────────────────────────────────────────────────────────────┐
│ Step 2: Keychain Integration                               │
└────────────────────────────────────────────────────────────┘

Store your master password in OS keychain for convenience?
(Windows Credential Manager, macOS Keychain, Linux Secret Service)

Benefits:
  ✓ No need to type password for every operation
  ✓ Secure OS-level storage
  ✓ Can be disabled later with --no-keychain

Enable keychain storage? (y/n): y

✓ Master password stored in keychain

┌────────────────────────────────────────────────────────────┐
│ Step 3: Audit Logging                                      │
└────────────────────────────────────────────────────────────┘

Enable audit logging to track all vault operations?

Benefits:
  ✓ Security audit trail
  ✓ Tamper-evident with HMAC signatures
  ✓ Track all add/get/update/delete operations

Enable audit logging? (y/n): y

✓ Audit logging enabled

┌────────────────────────────────────────────────────────────┐
│ Setup Complete!                                            │
└────────────────────────────────────────────────────────────┘

Your vault has been created at:
  /home/user/.pass-cli/vault.enc

Next steps:
  • Add your first credential: pass-cli add github
  • List credentials: pass-cli list
  • Check vault health: pass-cli doctor

No credentials found.
```

### When Guided Initialization Is NOT Triggered

Guided initialization is **skipped** in these scenarios:

1. **Vault already exists**: If you have an existing vault at the default location, no prompt appears
2. **Custom vault flag used**: If you specify `--vault /path/to/vault.enc`, initialization is skipped (you're explicitly managing vault location)
3. **Non-TTY environment**: If running in a script or pipe (stdin is not a terminal), initialization is skipped to avoid blocking
4. **Commands that don't require vault**: Commands like `version`, `doctor`, `help` don't trigger initialization

### Non-TTY Behavior

If you run pass-cli in a script or CI/CD environment without a vault:

```bash
$ echo "list" | pass-cli list
Error: vault file not found: /home/user/.pass-cli/vault.enc

Run 'pass-cli init' to create a new vault.
```

This prevents scripts from hanging while waiting for interactive input.

## Manual Initialization

If you prefer to initialize your vault explicitly, or if you're in a non-interactive environment, use the `init` command:

```bash
pass-cli init
```

This provides the same setup process as guided initialization, but you invoke it explicitly.

### Advanced Options

#### Custom Vault Location

```bash
pass-cli init --vault /custom/path/vault.enc
```

Creates a vault at a custom location. You'll need to use `--vault` flag with all future commands:

```bash
pass-cli add github --vault /custom/path/vault.enc
pass-cli get github --vault /custom/path/vault.enc
```

#### Skip Keychain Integration

```bash
pass-cli init --no-keychain
```

Creates a vault without storing the master password in OS keychain. You'll need to enter your password for each operation.

#### Disable Audit Logging

```bash
pass-cli init --no-audit
```

Creates a vault without audit logging enabled (not recommended for production use).

## Your First Credential

After initialization (automatic or manual), add your first credential:

```bash
$ pass-cli add github

Enter username: your-github-username
Enter password: ••••••••••••
Confirm password: ••••••••••••

✓ Credential 'github' added successfully
```

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

```bash
$ pass-cli generate

Generated password (16 characters):
kR9$mN2@pL5#wQ8!

# Generate with custom length
$ pass-cli generate --length 24

# Copy to clipboard (macOS)
$ pass-cli generate | pbcopy

# Use with 'add' command
$ pass-cli add myservice --password "$(pass-cli generate)"
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

See [doctor-command.md](doctor-command.md) for detailed health check documentation.

### Common First-Time Issues

#### Keychain Access Denied (macOS)

**Symptom**:
```
⚠ Keychain: Access denied by OS
```

**Solution**: Grant Terminal/iTerm2 keychain access:
1. System Preferences → Security & Privacy → Privacy
2. Select "Keychain" from left sidebar
3. Check the box next to your terminal application

#### Keychain Unavailable (Linux)

**Symptom**:
```
⚠ Keychain: OS keychain not available
```

**Solution**: Install Secret Service backend:

```bash
# Debian/Ubuntu
sudo apt-get install libsecret-1-0

# Fedora/RHEL
sudo dnf install libsecret

# Arch Linux
sudo pacman -S libsecret
```

#### Permission Issues (All Platforms)

**Symptom**:
```
✗ Vault: Vault file has insecure permissions
```

**Solution**:

```bash
# Linux/macOS
chmod 600 ~/.pass-cli/vault.enc

# Windows (PowerShell)
icacls "$env:USERPROFILE\.pass-cli\vault.enc" /inheritance:r /grant:r "$env:USERNAME:F"
```

## Changing Your Master Password

If you need to change your master password:

```bash
pass-cli change-password
```

Enter current password, then your new password. All credentials are re-encrypted with the new password.

## Backup and Recovery

pass-cli automatically creates backup files when you modify credentials:

```bash
$ ls ~/.pass-cli/
vault.enc
vault.enc.backup.1  # Most recent backup
vault.enc.backup.2  # Second most recent
vault.enc.backup.3  # Oldest backup
```

To restore from backup:

```bash
cp ~/.pass-cli/vault.enc.backup.1 ~/.pass-cli/vault.enc
```

## Configuration

Customize pass-cli behavior with `~/.pass-cli/config.json`:

```json
{
  "vault_path": "/home/user/.pass-cli/vault.enc",
  "keychain_enabled": true,
  "audit_enabled": true,
  "backup_count": 3,
  "password_policy": {
    "min_length": 12,
    "require_uppercase": true,
    "require_lowercase": true,
    "require_number": true,
    "require_special": true
  }
}
```

See [USAGE.md](USAGE.md) for complete configuration documentation.

## Next Steps

Now that you have pass-cli set up:

1. **Import existing credentials**: Migrate from other password managers
   - See [MIGRATION.md](MIGRATION.md) for migration guides

2. **Set up shell integration**: Add password retrieval to your shell scripts
   ```bash
   # .bashrc / .zshrc
   alias dbpass='pass-cli get database --quiet --field password'
   ```

3. **Configure backups**: Ensure vault backups are included in your backup strategy
   ```bash
   # Include ~/.pass-cli/ in your backup scripts
   rsync -av ~/.pass-cli/ /backup/location/pass/
   ```

4. **Review security best practices**: Learn about secure password management
   - See [SECURITY.md](SECURITY.md) for security guidelines

5. **Explore advanced features**:
   - Audit log verification: `pass-cli verify-audit`
   - Custom password generation: `pass-cli generate --length 32 --no-special`
   - Multiple vault support: `--vault` flag for different vaults

## Getting Help

- **Command help**: `pass-cli help` or `pass-cli <command> --help`
- **Troubleshooting**: See [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- **Security concerns**: See [SECURITY.md](SECURITY.md)
- **Report issues**: https://github.com/ari1110/pass-cli/issues

## Quick Reference

```bash
# Initialization
pass-cli init                    # Create new vault (manual)
# Or just run any command - guided init will trigger automatically

# Add credential
pass-cli add <key>

# Retrieve credential
pass-cli get <key>
pass-cli get <key> --quiet --field password  # Script-friendly

# List all credentials
pass-cli list

# Update credential
pass-cli update <key>

# Delete credential
pass-cli delete <key>

# Generate password
pass-cli generate

# Health check
pass-cli doctor

# Change master password
pass-cli change-password

# Verify audit log
pass-cli verify-audit

# Version info
pass-cli version
```

Welcome to secure, simple password management with pass-cli!
