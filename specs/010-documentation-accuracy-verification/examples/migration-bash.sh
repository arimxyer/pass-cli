#!/bin/bash

# Bash code blocks extracted from docs/MIGRATION.md
# These examples can be used for testing migration scenarios

# -------------------------------------------------------------------------
# Section: Option A: Fresh Vault (Recommended) - Backup current vault
# -------------------------------------------------------------------------
# Backup your vault
cp ~/.pass-cli/vault.enc ~/backup/vault-old-$(date +%Y%m%d).enc

# Export credentials (optional)
pass-cli list --format json > ~/backup/credentials-$(date +%Y%m%d).json

# -------------------------------------------------------------------------
# Section: Option A: Fresh Vault (Recommended) - Initialize new vault
# -------------------------------------------------------------------------
# Create new vault (automatically uses 600k iterations)
pass-cli init

# Or with audit logging enabled
pass-cli init --enable-audit

# -------------------------------------------------------------------------
# Section: Option A: Fresh Vault (Recommended) - Re-add credentials
# -------------------------------------------------------------------------
# Interactive mode (recommended for password policy compliance)
pass-cli add service1
pass-cli add service2

# Or generate password separately, then add credential
pass-cli generate  # Copy generated password
pass-cli add service1 --username user@example.com  # Paste when prompted

# -------------------------------------------------------------------------
# Section: Option A: Fresh Vault (Recommended) - Verify migration
# -------------------------------------------------------------------------
# List all credentials
pass-cli list

# Test accessing a credential
pass-cli get service1

# -------------------------------------------------------------------------
# Section: Option A: Fresh Vault (Recommended) - Delete old vault
# -------------------------------------------------------------------------
rm ~/backup/vault-old-*.enc

# -------------------------------------------------------------------------
# Section: Option B: In-Place Migration (Future Feature)
# -------------------------------------------------------------------------
# Future: Migrate vault to 600k iterations in-place
pass-cli migrate --iterations 600000

# Future: Migrate with audit logging enabled
pass-cli migrate --iterations 600000 --enable-audit

# -------------------------------------------------------------------------
# Section: Option C: Hybrid Approach (Keep Old Vault) - Create new vault
# -------------------------------------------------------------------------
pass-cli --vault ~/.pass-cli/vault-new.enc init --enable-audit

# -------------------------------------------------------------------------
# Section: Option C: Hybrid Approach (Keep Old Vault) - Add new credentials
# -------------------------------------------------------------------------
pass-cli --vault ~/.pass-cli/vault-new.enc add newservice

# -------------------------------------------------------------------------
# Section: Option C: Hybrid Approach (Keep Old Vault) - Access old vault
# -------------------------------------------------------------------------
pass-cli --vault ~/.pass-cli/vault-old.enc get oldservice

# -------------------------------------------------------------------------
# Section: Option C: Hybrid Approach (Keep Old Vault) - Switch to new vault
# -------------------------------------------------------------------------
mv ~/.pass-cli/vault-old.enc ~/.pass-cli/vault-old-backup.enc
mv ~/.pass-cli/vault-new.enc ~/.pass-cli/vault.enc

# -------------------------------------------------------------------------
# Section: Troubleshooting - Audit log verification fails
# -------------------------------------------------------------------------
# Backup corrupted log
mv ~/.pass-cli/audit.log ~/.pass-cli/audit.log.corrupted

# Start fresh audit log (requires vault re-init with --enable-audit)
pass-cli init --enable-audit

# -------------------------------------------------------------------------
# Section: Troubleshooting - Vault file corrupted after migration
# -------------------------------------------------------------------------
# Restore from backup
cp ~/backup/vault-old-*.enc ~/.pass-cli/vault.enc

# Verify restoration
pass-cli list

# Retry migration more carefully