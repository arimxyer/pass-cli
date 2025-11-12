# Vault Backup & Restore Guide

Complete guide to backing up and restoring your pass-cli vault, including automatic backups, manual backups, and disaster recovery procedures.

## Table of Contents

- [Overview](#overview)
- [Automatic Backups](#automatic-backups)
- [Manual Backups](#manual-backups)
- [Restoring from Backup](#restoring-from-backup)
- [Viewing Backup Status](#viewing-backup-status)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)

## Overview

Pass-CLI provides two types of vault backups:

1. **Automatic Backups**: Created automatically before each vault modification
2. **Manual Backups**: Created on-demand with timestamped filenames

Both backup types:
- Use the same AES-256-GCM encryption as your vault
- Include full integrity verification
- Require no password to create (vault must exist)
- Can be restored to replace a corrupted or lost vault

### Backup File Locations

**Default vault location**:
- Windows: `%USERPROFILE%\.pass-cli\vault.enc`
- macOS/Linux: `~/.pass-cli/vault.enc`

**Backup files**:
- Automatic: `vault.enc.backup` (same directory as vault)
- Manual: `vault.enc.YYYYMMDD-HHMMSS.manual.backup` (timestamped)

## Automatic Backups

Automatic backups are created before every vault modification (add, update, delete operations). The system maintains one automatic backup file that is overwritten with each new modification.

### How Automatic Backups Work

```bash
# When you modify the vault:
pass-cli add github --username myuser --password mypass

# Automatic backup created at:
# ~/.pass-cli/vault.enc.backup
```

**Backup Lifecycle**:
1. Before saving changes, the current vault is copied to `vault.enc.backup`
2. Changes are saved to the main vault file
3. Previous automatic backup is replaced

**Benefits**:
- No action required - completely automatic
- Always have the last known good state
- Protects against accidental deletions or corrupted saves

## Manual Backups

Manual backups create timestamped copies of your vault that persist alongside automatic backups. Use these for important checkpoints or before major changes.

### Creating Manual Backups

```bash
# Create a manual backup
pass-cli vault backup create

# Output:
# ✅ Backup created successfully
#
# Backup: /home/user/.pass-cli/vault.enc.20251111-143022.manual.backup
# Size: 2.5 KB
# Created: 2025-11-11 14:30:22
#
# You can restore from this backup with: pass-cli vault backup restore
```

### With Verbose Output

```bash
# See detailed progress
pass-cli vault backup create --verbose

# Output includes:
# [VERBOSE] Vault path: /home/user/.pass-cli/vault.enc
# [VERBOSE] Generating timestamped backup filename...
# [VERBOSE] Backup created at: /home/user/.pass-cli/vault.enc.20251111-143022.manual.backup
# [VERBOSE] Verifying backup integrity...
# [VERBOSE] Backup verified successfully
# [VERBOSE] Backup creation completed
```

### When to Create Manual Backups

- **Before bulk operations**: Importing or deleting many credentials
- **Before experiments**: Testing new workflows or scripts
- **Periodic checkpoints**: Weekly or monthly snapshots
- **Before system changes**: OS upgrades, migrations, or hardware changes
- **Before major updates**: Upgrading pass-cli to a new version

## Restoring from Backup

The restore command automatically selects the newest valid backup (manual or automatic) and replaces your current vault.

### Basic Restore

```bash
# Restore with confirmation prompt
pass-cli vault backup restore

# Output:
# ⚠️  Warning: This will overwrite your current vault with the backup.
#
# Backup to restore:
#   File: /home/user/.pass-cli/vault.enc.20251111-143022.manual.backup
#   Type: manual
#   Modified: 2025-11-11 14:30:22
#
# Are you sure you want to continue? (y/n):
```

### Restore Without Confirmation

```bash
# Use --force to skip confirmation (useful for scripts)
pass-cli vault backup restore --force

# Output:
# ✅ Vault restored successfully from backup
#
# Restored from: /home/user/.pass-cli/vault.enc.20251111-143022.manual.backup
# Backup type: manual
#
# You can now unlock your vault with your master password.
```

### Dry Run Mode

Preview which backup would be restored without making changes:

```bash
pass-cli vault backup restore --dry-run

# Output:
# Dry-run mode: No changes will be made
#
# Would restore from:
#   Backup: /home/user/.pass-cli/vault.enc.20251111-143022.manual.backup
#   Type: manual
#   Size: 2.5 KB
#   Modified: 2025-11-11 14:30:22
#
# To actually restore, run without --dry-run flag.
```

### Restore with Verbose Output

```bash
pass-cli vault backup restore --force --verbose

# Shows detailed progress:
# [VERBOSE] Vault path: /home/user/.pass-cli/vault.enc
# [VERBOSE] Searching for backups...
# [VERBOSE] Found backup: /home/user/.pass-cli/vault.enc.20251111-143022.manual.backup
# [VERBOSE] Backup type: manual
# [VERBOSE] Backup size: 2.5 KB
# [VERBOSE] Backup modified: 2025-11-11 14:30:22
# [VERBOSE] Starting restore operation...
# [VERBOSE] Backup copied to vault location
# [VERBOSE] Verifying vault file permissions...
# [VERBOSE] Vault permissions set to 0600
# [VERBOSE] Restore operation completed
```

## Viewing Backup Status

Use the `info` command to see all available backups, their status, and disk usage.

### Basic Info

```bash
pass-cli vault backup info

# Output:
# 📦 Vault Backup Status
#
# Automatic Backup:
# ✓ 2 hours ago, 2.5 KB
#
# Manual Backups (3 total):
#
# 1. ✓ 2 hours ago, 2.5 KB
# 2. ✓ 1 day ago, 2.4 KB
# 3. ✓ 1 week ago, 2.3 KB
#
# Total backup size: 7.2 KB
#
# ✓ Restore priority: manual backup (2 hours ago)
```

### Info with Verbose Output

```bash
pass-cli vault backup info --verbose

# Shows full paths:
# 📦 Vault Backup Status
#
# Automatic Backup:
# ✓ 2 hours ago, 2.5 KB
#    Path: /home/user/.pass-cli/vault.enc.backup
#    Modified: 2025-11-11 14:30:22
#
# Manual Backups (3 total):
#
# 1. ✓ 2 hours ago, 2.5 KB
#    Path: /home/user/.pass-cli/vault.enc.20251111-143022.manual.backup
#    Modified: 2025-11-11 14:30:22
# ...
```

### Understanding Info Output

- **✓** = Backup passed integrity verification
- **⚠️** = Backup is corrupted or invalid
- **Restore priority** = Which backup will be used for restore
- **Total backup size** = Combined size of all backups

### Warnings in Info Output

```bash
# Old backup warning (>30 days)
⚠️  Warning: Backup is 45 days old. Consider creating a fresh backup.

# Too many manual backups (>5)
⚠️  Warning: 12 manual backups found. Consider removing old backups to free disk space.
```

## Best Practices

### Backup Strategy

1. **Rely on automatic backups** for daily protection
2. **Create manual backups** before:
   - Major changes (bulk operations, migrations)
   - Version upgrades
   - System changes
3. **Test restore** periodically to verify backups work
4. **Store backups externally** for disaster recovery:
   ```bash
   # Copy backup to external location
   cp ~/.pass-cli/vault.enc.backup ~/Dropbox/pass-cli-backup/
   ```

### Backup Retention

Manual backups accumulate over time. Clean up old backups periodically:

```bash
# List all manual backups
ls -lh ~/.pass-cli/*.manual.backup

# Remove backups older than 90 days (example)
find ~/.pass-cli -name "*.manual.backup" -mtime +90 -delete
```

### External Backup Storage

For critical vaults, store backups in multiple locations:

```bash
# Encrypted cloud storage (recommended)
cp ~/.pass-cli/vault.enc.backup ~/Dropbox/backups/

# USB drive
cp ~/.pass-cli/vault.enc.backup /media/usb/pass-cli-backup/

# Network storage
cp ~/.pass-cli/vault.enc.backup /mnt/nas/backups/
```

**Note**: Backup files are already encrypted with AES-256-GCM. They're safe to store on cloud services or external media.

### Backup Verification

Verify backups are valid before relying on them:

```bash
# Check backup status
pass-cli vault backup info

# Look for:
# - ✓ (checkmark) = valid backup
# - ⚠️ (warning) = corrupted backup

# Test restore (dry run)
pass-cli vault backup restore --dry-run
```

## Troubleshooting

### No Backups Found

```bash
Error: no backup available

No backup files found. Create a backup with: pass-cli vault backup create
```

**Solution**: Create your first backup:
```bash
pass-cli vault backup create
```

### Corrupted Backup

```bash
# Info shows corrupted backup
⚠️ 2 hours ago, 2.5 KB
```

**Causes**:
- Incomplete backup (disk full, interrupted operation)
- File system corruption
- Manual modification of backup file

**Solution**:
1. Check if other backups are available: `pass-cli vault backup info`
2. Create a new backup from current vault: `pass-cli vault backup create`
3. If all backups are corrupted, focus on recovering the main vault with `pass-cli doctor`

### Permission Denied

```bash
Error: permission denied creating backup

Check directory permissions for: /home/user/.pass-cli
```

**Solution**:
```bash
# Fix directory permissions
chmod 700 ~/.pass-cli

# Fix vault file permissions
chmod 600 ~/.pass-cli/vault.enc
```

### Backup Too Large

Backups are the same size as your vault. If backups are consuming too much disk space:

```bash
# Check backup sizes
pass-cli vault backup info

# Remove old manual backups
rm ~/.pass-cli/vault.enc.20241001-*.manual.backup
```

### Restore Fails

```bash
Error: failed to restore from backup: <reason>
```

**Common causes**:
1. **Backup file missing**: Check if backup exists
2. **Corrupted backup**: Try another backup
3. **Permission error**: Fix file permissions (see above)

**Recovery steps**:
```bash
# 1. View all available backups
pass-cli vault backup info

# 2. Try dry-run to see which backup would be used
pass-cli vault backup restore --dry-run

# 3. If automatic backup is corrupted, newest manual backup will be used
pass-cli vault backup restore --force
```

## Advanced Usage

### Scripting with Backup Commands

```bash
#!/bin/bash
# Automated backup script

# Create daily backup with error handling
if pass-cli vault backup create &>/dev/null; then
    echo "✓ Backup created: $(date)"
else
    echo "✗ Backup failed: $(date)" >&2
    exit 1
fi

# Copy to external storage
cp ~/.pass-cli/vault.enc.*.manual.backup /mnt/nas/backups/

# Clean up backups older than 30 days
find ~/.pass-cli -name "*.manual.backup" -mtime +30 -delete
```

### Pre-Migration Backup

Before migrating to a new system:

```bash
# 1. Create final backup on old system
pass-cli vault backup create

# 2. Copy vault and all backups
tar czf pass-cli-migration.tar.gz ~/.pass-cli/

# 3. Transfer to new system
scp pass-cli-migration.tar.gz newhost:~/

# 4. Extract on new system
ssh newhost 'tar xzf pass-cli-migration.tar.gz'

# 5. Verify on new system
ssh newhost 'pass-cli vault backup info'
```

### Backup Rotation Script

Automated backup rotation to maintain 7 daily, 4 weekly backups:

```bash
#!/bin/bash
# backup-rotation.sh

VAULT_DIR="$HOME/.pass-cli"
BACKUP_DIR="$HOME/vault-backups"

# Create backup directory
mkdir -p "$BACKUP_DIR"/{daily,weekly}

# Create new backup
pass-cli vault backup create

# Get newest manual backup
NEWEST=$(ls -t "$VAULT_DIR"/*.manual.backup 2>/dev/null | head -1)

if [ -n "$NEWEST" ]; then
    # Daily backup
    cp "$NEWEST" "$BACKUP_DIR/daily/backup-$(date +%Y%m%d).enc"

    # Weekly backup (Sundays)
    if [ "$(date +%u)" = "7" ]; then
        cp "$NEWEST" "$BACKUP_DIR/weekly/backup-$(date +%Y-W%V).enc"
    fi

    # Rotate: keep last 7 daily backups
    ls -t "$BACKUP_DIR"/daily/*.enc | tail -n +8 | xargs rm -f

    # Rotate: keep last 4 weekly backups
    ls -t "$BACKUP_DIR"/weekly/*.enc | tail -n +5 | xargs rm -f
fi
```

### Continuous Backup Monitoring

Monitor backup health with cron:

```bash
# Add to crontab: check backup health daily
0 2 * * * pass-cli vault backup info | grep -q "⚠️" && echo "Warning: Backup issue detected" | mail -s "pass-cli backup alert" admin@example.com
```

## See Also

- [Security Documentation](../SECURITY.md) - Encryption details and security best practices
- [Troubleshooting Guide](../TROUBLESHOOTING.md) - General troubleshooting for pass-cli
- [Doctor Command](../DOCTOR_COMMAND.md) - Vault health checks and diagnostics
- [Getting Started](../GETTING_STARTED.md) - First-time setup and basic workflows
