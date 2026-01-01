---
title: "FAQ"
weight: 5
toc: true
---

Common questions about pass-cli features, usage, and troubleshooting.

## Frequently Asked Questions

### General Questions

**Q: Where is my vault stored?**

A:
- Windows: `%USERPROFILE%\.pass-cli\vault.enc`
- macOS/Linux: `~/.pass-cli/vault.enc`

---

**Q: How do I backup my vault?**

A: Use the built-in backup commands:
```bash
# Create a manual backup
pass-cli vault backup create

# View all available backups
pass-cli vault backup info

# Restore from a backup
pass-cli vault backup restore
```

For automated backups, you can use cron with the backup command or copy the vault file directly.

---

**Q: Can I sync my vault across machines?**

A: Yes! Pass-CLI has built-in cloud sync via rclone:

```bash
# Option 1: Enable sync on existing vault
pass-cli sync enable
# Follow prompts to configure remote (e.g., gdrive:.pass-cli)

# Option 2: Connect to existing synced vault on new device
pass-cli init
# Select "Connect to existing synced vault" when prompted

# Option 3: Set up sync during new vault creation
pass-cli init
# Select "Create new vault", then enable sync when prompted
```

Sync automatically pulls on first use and pushes after write operations. See [Sync Guide](../02-guides/sync-guide.md) for full details.

**Note**: Requires [rclone](https://rclone.org) installed and configured with at least one remote.

---

**Q: How do I change my master password?**

A: Use the `change-password` command:
```bash
pass-cli change-password
```

You'll be prompted to:
1. Enter your current master password
2. Enter a new master password (must meet security requirements)
3. Confirm the new master password

The vault will be automatically re-encrypted with the new password.

**If you forgot your master password:**
```bash
# Use your BIP39 recovery phrase (if enabled during vault initialization)
pass-cli change-password --recover
```

You'll be prompted to enter 6 words from your 24-word recovery phrase to verify your identity, then you can set a new master password.

**Note**: Recovery only works if you enabled the recovery phrase when you initialized your vault. If you used `--no-recovery` during init, recovery is not possible.

---

**Q: Is my data sent to the cloud?**

A: By default, no. Pass-CLI:
- ✅ Works completely offline by default
- ✅ Stores everything locally
- ✅ No telemetry or tracking
- ✅ Optional cloud sync (you control if/when enabled)

If you enable cloud sync, your **encrypted** vault is synced to your configured rclone remote. Your master password and decrypted credentials never leave your device.

---

**Q: What happens if I lose my vault file?**

A:
- **If sync enabled**: Run `pass-cli init` and connect to your synced vault
- **If you have backup**: Restore with `pass-cli vault backup restore`
- **If no backup or sync**: All credentials lost, must start over
- **Prevention**: Enable cloud sync or create regular backups

---

### Technical Questions

**Q: Can I use Pass-CLI in scripts?**

A: Yes, designed for it:
```bash
# Use quiet mode
export API_KEY=$(pass-cli get service --quiet)

# Extract specific field
export USERNAME=$(pass-cli get service --field username --quiet)

# JSON output for parsing
pass-cli list --format json | jq '.[] | .service'
```

---

**Q: How secure is Pass-CLI?**

A: See [Security Architecture](../03-reference/security-architecture) for full details:
- AES-256-GCM encryption
- PBKDF2 key derivation (600,000 iterations)
- System keychain integration
- Local-first with optional encrypted cloud sync
- Limitations explained in security doc

---

**Q: Can multiple users share a vault?**

A: Not designed for this:
- Vault is single-user
- Master password would be shared (insecure)
- No access control mechanism
- Solution: Use separate vaults per user, each with their own config file pointing to their vault

---

**Q: What if I forget a specific credential password?**

A: Individual credentials cannot be recovered:
- Vault decrypts all-or-nothing
- If vault accessible, all credentials accessible
- If vault locked, all credentials inaccessible
- No per-credential password recovery

---

**Q: How do I check if sync is working?**

A: Use the doctor command:
```bash
pass-cli doctor
```

This shows sync status including:
- Whether sync is enabled
- rclone installation status
- Remote configuration
- Any connectivity issues

---

**Q: Sync failed - what should I do?**

A: Sync failures are non-blocking (your local vault still works). To troubleshoot:

```bash
# Check sync configuration
pass-cli doctor

# Verify rclone can reach your remote
rclone lsd your-remote:.pass-cli

# Check rclone config
rclone config

# Manual sync (if needed)
rclone sync ~/.pass-cli your-remote:.pass-cli
```

See [Sync Troubleshooting](../02-guides/sync-guide.md#troubleshooting) for more details.

---

