---
title: "FAQ"
weight: 5
bookToc: true
---

# Frequently Asked Questions

Common questions about pass-cli features, usage, and troubleshooting.

## Frequently Asked Questions

### General Questions

**Q: Where is my vault stored?**

A:
- Windows: `%USERPROFILE%\.pass-cli\vault.enc`
- macOS/Linux: `~/.pass-cli/vault.enc`

---

**Q: How do I backup my vault?**

A:
```bash
# Simple copy
cp ~/.pass-cli/vault.enc ~/backups/vault-$(date +%Y%m%d).enc

# Automated daily backup (cron)
0 0 * * * cp ~/.pass-cli/vault.enc ~/backups/vault-$(date +%Y%m%d).enc
```

---

**Q: Can I sync my vault across machines?**

A: Yes, but carefully:
```bash
# Option 1: Symlink to cloud storage
ln -s ~/Dropbox/vault.enc ~/.pass-cli/vault.enc

# Option 2: Configure vault_path to point to cloud storage
echo "vault_path: ~/Dropbox/vault.enc" > ~/.pass-cli/config.yml

# Option 3: Manual copy (requires keeping in sync)
cp ~/.pass-cli/vault.enc ~/Dropbox/vault.enc

# Warning: Conflicts if editing on multiple machines simultaneously
```

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

---

**Q: Is my data sent to the cloud?**

A: No. Pass-CLI:
- ✅ Works completely offline
- ✅ Never makes network calls
- ✅ Stores everything locally
- ✅ No telemetry or tracking

---

**Q: What happens if I lose my vault file?**

A:
- If you have backup: Restore from backup
- If no backup: All credentials lost, must start over
- Prevention: Regular backups essential

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

A: See [Security Architecture]({{< relref "../03-reference/security-architecture" >}}) for full details:
- AES-256-GCM encryption
- PBKDF2 key derivation (600,000 iterations as of January 2025)
- System keychain integration
- No cloud dependencies
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

