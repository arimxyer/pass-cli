---
title: "Cloud Sync"
weight: 5
toc: true
---

Synchronize your vault across multiple devices using rclone. This enables seamless access to your credentials from different computers (e.g., dual-boot setups, work/home machines).

## Overview

Pass-CLI integrates with [rclone](https://rclone.org/) to sync your vault directory to cloud storage providers. The sync feature:

- **Pulls** the latest vault from the cloud before any operation (once per session)
- **Pushes** changes to the cloud after write operations (add, update, delete)
- **Works offline** - operations continue if sync fails (with warning)
- **Supports 70+ cloud providers** via rclone (Google Drive, Dropbox, OneDrive, S3, etc.)

### How Sync Works

```text
┌─────────────┐     pull (on first use)      ┌─────────────┐
│   Device A  │ ◄──────────────────────────► │   Cloud     │
│  (Windows)  │                              │  Storage    │
└─────────────┘     push (after writes)      └─────────────┘
                                                    ▲
                                                    │
┌─────────────┐     pull (on first use)             │
│   Device B  │ ◄───────────────────────────────────┘
│   (Linux)   │     push (after writes)
└─────────────┘
```

**Session-aware sync**: Pull only happens once per CLI session to avoid unnecessary network calls. Push happens immediately after every write operation.

## Prerequisites

### 1. Install rclone

**Windows (Scoop)**:
```bash
scoop install rclone
```

**Windows (Chocolatey)**:
```bash
choco install rclone
```

**macOS**:
```bash
brew install rclone
```

**Linux (Debian/Ubuntu)**:
```bash
sudo apt install rclone
```

**Linux (Arch)**:
```bash
sudo pacman -S rclone
```

Verify installation:
```bash
rclone version
```

### 2. Configure a Remote

Configure rclone with your cloud provider. Example for Google Drive:

```bash
rclone config
```

Follow the interactive prompts:
1. Choose `n` for new remote
2. Name it (e.g., `gdrive`)
3. Select your provider (e.g., `drive` for Google Drive)
4. Follow provider-specific OAuth flow
5. For Google Drive, select scope `3` (drive.file) for minimal permissions

Test the remote:
```bash
# List remote contents
rclone ls gdrive:

# Create a test directory
rclone mkdir gdrive:.pass-cli
```

## Configuration

Enable sync in your pass-cli configuration file (`~/.pass-cli/config.yml`):

```yaml
sync:
  enabled: true
  remote: "gdrive:.pass-cli"  # Format: <remote-name>:<path>
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `false` | Enable/disable sync |
| `remote` | string | `""` | rclone remote and path (e.g., `gdrive:.pass-cli`) |

### Remote Path Format

The `remote` field uses rclone's remote path format:

```yaml
# Google Drive
remote: "gdrive:.pass-cli"

# Dropbox
remote: "dropbox:Apps/pass-cli"

# OneDrive
remote: "onedrive:Documents/pass-cli"

# S3-compatible storage
remote: "s3:my-bucket/pass-cli"

# SFTP server
remote: "myserver:/home/user/.pass-cli"
```

## Usage

Once configured, sync happens automatically:

```bash
# First command in session - pulls from cloud
pass-cli list

# Subsequent reads - no sync (already pulled this session)
pass-cli get github

# Write operations - push to cloud after completion
pass-cli add newservice --username user --password pass
# Output: [sync] Pushed changes to remote
```

### Manual Sync

Sync is automatic, but you can manually trigger it by restarting your terminal session or using rclone directly:

```bash
# Manual pull
rclone sync gdrive:.pass-cli ~/.pass-cli

# Manual push
rclone sync ~/.pass-cli gdrive:.pass-cli
```

## Portable Audit Keys

When sync is enabled, pass-cli automatically uses **portable audit key derivation**. This means:

- Audit keys are derived from your master password (not stored in OS keychain)
- `verify-audit` command works on any synced device
- Audit log integrity can be verified cross-platform

The audit salt is stored in your vault's metadata file and syncs with your vault.

### How It Works

```text
Master Password + Audit Salt
         │
         ▼ PBKDF2-SHA256 (100k iterations)
         │
    Audit Key
         │
         ▼
   HMAC Signatures (audit log entries)
```

This replaces the default keychain-based audit key storage when sync is enabled.

## Dual-Boot Setup Example

Perfect for users who dual-boot between Windows and Linux:

**1. Configure rclone on both OSes** (with same remote name):

```bash
# On Windows
rclone config
# Create remote named "gdrive"

# On Linux
rclone config
# Create remote named "gdrive" (same name, same account)
```

**2. Enable sync on both OSes** (`~/.pass-cli/config.yml`):

```yaml
sync:
  enabled: true
  remote: "gdrive:.pass-cli"
```

**3. Initialize vault on one OS**:

```bash
# On Windows
pass-cli init
pass-cli add github --username myuser --password mypass
# Vault syncs to cloud
```

**4. Use on other OS**:

```bash
# On Linux - first use pulls vault from cloud
pass-cli list
# Shows: github
```

## Conflict Handling

Pass-CLI uses rclone's sync behavior which **overwrites** the destination with the source. This means:

- **Pull**: Cloud overwrites local (ensures you have latest)
- **Push**: Local overwrites cloud (your changes take precedence)

**To avoid conflicts**:
1. Always use the same device for a session
2. Don't run pass-cli simultaneously on multiple devices
3. If unsure, manually check with `rclone ls <remote>` before operations

## Offline Operation

Sync failures don't block operations:

```bash
# If offline or cloud unreachable
pass-cli add service --username user --password pass

# Output:
# Warning: sync push failed: <error details>
# Credential 'service' added successfully
```

Your local vault is updated successfully. Changes sync on next successful push.

## Troubleshooting

### Sync Not Working

1. **Verify rclone is installed**:
   ```bash
   rclone version
   ```

2. **Test remote connectivity**:
   ```bash
   rclone ls gdrive:.pass-cli
   ```

3. **Check configuration**:
   ```bash
   pass-cli config validate
   ```

4. **Verify sync is enabled**:
   ```yaml
   # ~/.pass-cli/config.yml
   sync:
     enabled: true      # Must be true
     remote: "gdrive:.pass-cli"  # Must not be empty
   ```

### "rclone not found" Warning

Install rclone and ensure it's in your PATH:

```bash
# Check if rclone is in PATH
which rclone    # Linux/macOS
where rclone    # Windows
```

### Permission Denied Errors

Ensure your rclone remote has write permissions:

```bash
# Test write access
echo "test" > /tmp/test.txt
rclone copy /tmp/test.txt gdrive:.pass-cli/
rclone ls gdrive:.pass-cli/
```

### Audit Verification Fails After Sync

If `verify-audit` fails on a synced device:

1. Ensure you're using the **same master password** on all devices
2. Check that the vault metadata file (`.meta.json`) synced correctly
3. The audit salt must be present in metadata for portable key derivation

### Slow Sync Operations

For large vaults or slow connections:

```bash
# Check what will be synced without transferring
rclone sync ~/.pass-cli gdrive:.pass-cli --dry-run

# Show transfer progress
rclone sync ~/.pass-cli gdrive:.pass-cli --progress
```

## Security Considerations

### What Gets Synced

The entire vault directory is synced, including:
- `vault.enc` - Encrypted vault (AES-256-GCM)
- `vault.enc.meta.json` - Vault metadata (audit salt, timestamps)
- `audit.log` - Audit log (HMAC-signed entries)
- Backup files (if present)

### What Stays Local

- Master password (never stored)
- Keychain entries (OS-specific)
- Session state (in-memory only)

### Cloud Storage Security

Your vault is already encrypted with AES-256-GCM before sync. The cloud provider only sees encrypted data. However:

- Use a strong master password
- Enable 2FA on your cloud account
- Consider using end-to-end encrypted providers (e.g., rclone crypt)

### Additional Encryption Layer (Optional)

For extra security, use rclone's crypt remote:

```bash
# Create encrypted remote on top of cloud storage
rclone config
# Choose 'crypt' type
# Set remote to "gdrive:.pass-cli"
# This adds another encryption layer
```

## See Also

- [Configuration Reference](../03-reference/configuration) - Full configuration options
- [Security Architecture](../03-reference/security-architecture) - Encryption and audit details
- [Backup & Restore](./backup-restore) - Local backup strategies
- [rclone Documentation](https://rclone.org/docs/) - Official rclone docs
