---
title: "Keychain Setup"
weight: 2
toc: true
---

Configure OS keychain integration to store your master password securely and eliminate repeated password prompts.

## Keychain Integration

pass-cli can store your master password in your OS keychain for convenience, eliminating the need to type it for every operation.

### Enable Keychain Integration

If you didn't enable keychain during initialization, you can enable it anytime:

```bash
$ pass-cli keychain enable

Master password: ••••••••••••
[OK] Keychain integration enabled for vault at /home/user/.pass-cli/vault.enc

Future commands will not prompt for password when keychain is available.
```

### Check Keychain Status

View the current keychain integration status:

```bash
$ pass-cli keychain status

Keychain Status for /home/user/.pass-cli/vault.enc:

[PASS] System Keychain:        Available (macOS Keychain)
[PASS] Password Stored:        Yes
[PASS] Backend:                keychain
[PASS] Vault Configuration:    Keychain enabled

[PASS] Keychain integration is properly configured.
Your vault password is securely stored in the system keychain.
Future commands will not prompt for password.
```

**If keychain is not enabled:**
```bash
$ pass-cli keychain status

Keychain Status for /home/user/.pass-cli/vault.enc:

[PASS] System Keychain:        Available (Windows Credential Manager)
[FAIL] Password Stored:        No
[PASS] Vault Configuration:    Keychain not enabled

The system keychain is available but no password is stored for this vault.
Suggestion: Enable keychain integration with 'pass-cli keychain enable'
```

### Disable Keychain Integration

To remove your master password from the keychain, use your operating system's credential manager:

**Windows**:
```powershell
# Open Credential Manager → Windows Credentials
# Find and delete entry named "pass-cli"
cmdkey /delete:pass-cli
```

**macOS**:
```bash
# Open Keychain Access app, search for "pass-cli", delete the entry
# Or use command line:
security delete-generic-password -s "pass-cli" -a "$USER"
```

**Linux**:
```bash
# Use your desktop environment's credential manager
# Or use secret-tool:
secret-tool clear service pass-cli
```

After removing the keychain entry, you'll need to enter your master password for each operation.

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

## Health Checks

Verify your pass-cli installation is working correctly:

```bash
$ pass-cli doctor

Health Check Results
====================

[PASS] Version: v1.2.3 (up to date)
[PASS] Vault: vault.enc accessible (600 permissions)
[PASS] Config: Valid configuration
[PASS] Keychain: Integration active
[PASS] Backup: 3 backup files found

Overall Status: HEALTHY
```

See [Health Checks](../05-operations/health-checks) for detailed health check documentation.

### Common First-Time Issues

#### Keychain Access Denied (macOS)

**Symptom**:
```
⚠ Keychain: Access denied by OS
```
