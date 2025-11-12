---
title: "Migration"
weight: 4
bookToc: true
---

# Migration Guide
Guide for upgrading Pass-CLI vaults and adapting to security hardening changes (January 2025 release).

![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?path=docs&label=Last%20Updated)


## Table of Contents

- [Overview](#overview)
- [What Changed](#what-changed)
- [Migration Scenarios](#migration-scenarios)
- [Step-by-Step Migration](#step-by-step-migration)
- [Backward Compatibility](#backward-compatibility)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

## Overview

The January 2025 security hardening release introduces several important changes:

1. **Vault Location Configuration**: `--vault` flag removed, use config file instead
2. **Increased PBKDF2 Iterations**: 100,000 → 600,000 (6x stronger)
3. **Password Policy Enforcement**: New complexity requirements for all passwords
4. **Audit Logging**: Optional tamper-evident logging with HMAC signatures
5. **Atomic Vault Operations**: Improved rollback safety

**Good News**: Existing vaults continue to work without migration. You can upgrade at your own pace.

## What Changed

### 1. Vault Location Configuration (Breaking Change)

**The `--vault` flag and `PASS_CLI_VAULT` environment variable have been removed.**

**Before** (Old way - no longer works):
```bash
pass-cli --vault /custom/path/vault.enc init
pass-cli --vault /custom/path/vault.enc add github
export PASS_CLI_VAULT=/custom/path/vault.enc
pass-cli get github
```

**After** (New way - use config file):
```bash
# Set custom vault location in config file
echo "vault_path: /custom/path/vault.enc" > ~/.pass-cli/config.yml

# Now use commands without --vault flag
pass-cli init
pass-cli add github
pass-cli get github
```

**Why the change?**
- **Simplicity**: One consistent way to configure vault location
- **Less error-prone**: No conflict between flag, env var, and config
- **Better UX**: Users don't need to remember to add `--vault` to every command

**Migration Steps:**
1. If you use a **default vault location** (`~/.pass-cli/vault.enc`): Nothing to do! Everything works as-is.
2. If you use a **custom vault location**:
   - Create or edit `~/.pass-cli/config.yml`
   - Add: `vault_path: /your/custom/path/vault.enc`
   - Remove `--vault` flag from scripts and commands
   - Remove `PASS_CLI_VAULT` from your environment

**Path Expansion Support:**
- Environment variables: `vault_path: $HOME/.pass-cli/vault.enc`
- Tilde expansion: `vault_path: ~/my-vault.enc`
- Relative paths: `vault_path: vault.enc` (resolved relative to home directory)
- Absolute paths: `vault_path: /custom/absolute/path/vault.enc`

### 2. PBKDF2 Iterations (Crypto Hardening)

| Version | Iterations | Unlock Time (Modern CPU) | Security Benefit |
|---------|------------|--------------------------|------------------|
| **Old** | 100,000 | ~15-20ms | Baseline |
| **New** | 600,000 | ~50-100ms | 6x brute-force resistance |

**Impact**: Vault unlock is slightly slower (~30-80ms slower) but significantly more secure.

**Security Rationale**: The increase from 100,000 to 600,000 iterations aligns with current industry standards:
- **OWASP**: Recommends 600,000+ iterations for PBKDF2-SHA256 (2023 guidance)
- **NIST SP 800-132**: Recommends iteration counts that result in ≥100ms processing time
- **Brute-Force Resistance**: 6x computational cost for attackers attempting password cracking

### 2. Password Policy Enforcement

**New Requirements** (enforced for all passwords):
- Minimum 12 characters (was 8)
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special symbol (!@#$%^&*()-_=+[]{}|;:,.<>?)

**Impact**: Weak passwords will be rejected when creating/updating credentials.

### 3. Audit Logging (Optional Feature)

**New Feature**: Tamper-evident audit trail for vault operations.

- Opt-in via `--enable-audit` flag
- HMAC-SHA256 signatures for tamper detection
- Keys stored in OS keychain
- Auto-rotation at 10MB with 7-day retention

**Impact**: No impact unless you opt-in to enable audit logging.

## Migration Scenarios

### Scenario 1: No Migration (Recommended for Most Users)

**Who**: Users satisfied with current vault security.

**Action**: None required. Your vault continues to work with 100k iterations.

**Pros**:
- Zero downtime
- No password changes required
- Vault remains compatible with older Pass-CLI versions

**Cons**:
- Lower brute-force resistance (still secure, but not optimal)

### Scenario 2: Migrate to 600k Iterations

**Who**: Users wanting maximum security.

**Action**: Re-initialize vault with new master password or use migration command (future feature).

**Pros**:
- 6x stronger brute-force resistance
- Future-proof security posture

**Cons**:
- Slightly slower vault unlock (~30-80ms)
- Requires password policy compliance

### Scenario 3: Enable Audit Logging

**Who**: Users in regulated industries or requiring compliance audit trails.

**Action**: Enable audit logging on existing or new vault.

**Pros**:
- Tamper-evident audit trail
- Compliance-ready logging
- Forensic investigation support

**Cons**:
- Additional disk space (~10MB per rotation)
- Requires OS keychain for HMAC keys

## Step-by-Step Migration

### Option A: Fresh Vault (Recommended)

**Best for**: Small vaults (< 50 credentials) or users wanting a clean start.

**Steps**:

1. **Backup current vault**:
   ```bash
   # Backup your vault
   cp ~/.pass-cli/vault.enc ~/backup/vault-old-$(date +%Y%m%d).enc

   # Export credentials (optional)
   pass-cli list --format json > ~/backup/credentials-$(date +%Y%m%d).json
   ```

2. **Initialize new vault**:
   ```bash
   # Create new vault (automatically uses 600k iterations)
   pass-cli init

   # Or with audit logging enabled
   pass-cli init --enable-audit
   ```

3. **Re-add credentials**:
   ```bash
   # Interactive mode (recommended for password policy compliance)
   pass-cli add service1
   pass-cli add service2

   # Or generate password separately, then add credential
   pass-cli generate  # Copy generated password
   pass-cli add service1 --username user@example.com  # Paste when prompted
   ```

4. **Verify migration**:
   ```bash
   # List all credentials
   pass-cli list

   # Test accessing a credential
   pass-cli get service1
   ```

5. **Delete old vault** (after verification):
   ```bash
   rm ~/backup/vault-old-*.enc
   ```

**Time Required**: ~5-10 minutes for 20 credentials.

### Option B: In-Place Migration (Future Feature)

> **⚠️ WARNING**: This feature is **NOT YET IMPLEMENTED**. The `pass-cli migrate` command does not currently exist. Use Option A (Manual Migration) instead.

**Status**: Not yet implemented. Planned for future release.

**Planned Command** (for future reference only):
```bash
# Future: Migrate vault to 600k iterations in-place
pass-cli migrate --iterations 600000

# Future: Migrate with audit logging enabled
pass-cli migrate --iterations 600000 --enable-audit
```

**Expected Behavior**:
- Reads existing vault with current iteration count
- Re-encrypts all credentials with 600k iterations
- Creates backup automatically
- Atomic operation (rollback on failure)

### Option C: Hybrid Approach (Keep Old Vault)

**Best for**: Users wanting to test 600k iterations before full migration.

**Steps**:

1. **Create new vault in custom location**:
   ```bash
   # Create config for new vault
   mkdir -p ~/.pass-cli
   echo "vault_path: ~/.pass-cli/vault-new.enc" > ~/.pass-cli/config-new.yml

   # Initialize new vault
   pass-cli --config ~/.pass-cli/config-new.yml init --enable-audit
   ```

2. **Add new credentials to new vault**:
   ```bash
   pass-cli --config ~/.pass-cli/config-new.yml add newservice
   ```

3. **Keep old vault for existing credentials**:
   ```bash
   # Use default config (points to ~/.pass-cli/vault.enc)
   pass-cli get oldservice
   ```

4. **Switch to new vault when ready**:
   ```bash
   # Backup old vault
   mv ~/.pass-cli/vault.enc ~/.pass-cli/vault-old-backup.enc

   # Promote new vault to default
   mv ~/.pass-cli/vault-new.enc ~/.pass-cli/vault.enc

   # Update default config (or remove custom config file)
   rm ~/.pass-cli/config-new.yml
   ```

## Backward Compatibility

### Vault File Format

**100k Iteration Vaults**:
- ✅ Fully supported
- ✅ Auto-detected by iteration count in metadata
- ✅ No performance degradation
- ✅ Can be used alongside 600k vaults

**600k Iteration Vaults**:
- ⚠️ **Not compatible with older Pass-CLI versions** (pre-January 2025)
- ✅ Auto-detected by iteration count in metadata
- ✅ Future-proof format

### Password Policy

**Existing Credentials**:
- ✅ Old passwords (not meeting policy) remain valid
- ⚠️ Policy enforced only when creating/updating credentials
- ✅ No forced password changes

**New/Updated Credentials**:
- ⚠️ Must meet new policy requirements
- ✅ Real-time validation with helpful error messages
- ✅ TUI shows password strength indicator

### Cross-Version Compatibility Matrix

| Vault Type | Pass-CLI (Old) | Pass-CLI (Jan 2025) |
|------------|----------------|---------------------|
| 100k iterations | ✅ Read/Write | ✅ Read/Write |
| 600k iterations | ❌ Incompatible | ✅ Read/Write |
| With audit logging | ❌ Incompatible | ✅ Read/Write |

## Troubleshooting

### Problem: "Password does not meet requirements"

**Symptom**: Error when creating/updating credentials.

**Solution**:
```bash
# Ensure password meets policy:
# - 12+ characters
# - Uppercase + lowercase + digit + symbol

# Good examples:
MySecureP@ss2025!
Correct-Horse-Battery-29!
Admin#2025$Password

# Or generate a policy-compliant password
pass-cli generate  # Automatically meets policy requirements
```

### Problem: Vault unlock is slower after upgrade

**Symptom**: Vault unlock takes 50-100ms instead of 15-20ms.

**Explanation**: This is expected behavior with 600k iterations. The slowdown is intentional for security.

**Solution**: No action needed. Performance is within normal range.

**Benchmark**:
- Modern CPU (2023+): 50-100ms
- Mid-range CPU (2018-2022): 200-500ms
- Older CPU (2015-2017): 500-1000ms

### Problem: Cannot downgrade to older Pass-CLI version

**Symptom**: "Invalid vault format" error when using old Pass-CLI with new vault.

**Solution**:
1. Keep backup of old vault before migration
2. Or create new vault with old Pass-CLI version
3. Or upgrade to latest Pass-CLI version

### Problem: Audit log verification fails

**Symptom**: `pass-cli verify-audit` reports HMAC verification failures.

**Causes**:
- Audit log file manually edited (tampering detected)
- Audit key deleted from OS keychain
- Audit log file corrupted

**Solution**:
```bash
# Check if audit key exists in keychain
# If missing, audit logging needs to be re-enabled

# Backup corrupted log
mv ~/.pass-cli/audit.log ~/.pass-cli/audit.log.corrupted

# Start fresh audit log (requires vault re-init with --enable-audit)
pass-cli init --enable-audit
```

### Problem: "Vault file corrupted" after migration

**Symptom**: Cannot unlock vault after re-initialization.

**Solution**:
```bash
# Restore from backup
cp ~/backup/vault-old-*.enc ~/.pass-cli/vault.enc

# Verify restoration
pass-cli list

# Retry migration more carefully
```

## FAQ

### Q: Do I have to migrate?

**A**: No. Existing vaults with 100k iterations continue to work indefinitely. Migration is optional for users wanting stronger security.

### Q: Will migration delete my credentials?

**A**: No. Migration is non-destructive. Always creates backup before changes. Credentials are preserved.

### Q: How long does migration take?

**A**: Depends on vault size:
- Small vault (< 20 credentials): 5-10 minutes
- Medium vault (20-100 credentials): 15-30 minutes
- Large vault (100+ credentials): 30-60 minutes

Time includes manual re-entry of credentials. Future in-place migration will be automatic (seconds).

### Q: Can I migrate back to 100k iterations?

**A**: Technically yes (create new vault), but not recommended. Forward migration only makes sense for security.

### Q: Does audit logging slow down vault operations?

**A**: Minimal impact (~1-2ms per operation). Audit logging uses asynchronous writes and graceful degradation.

### Q: What if I forget my master password after migration?

**A**: Same as before: vault is unrecoverable. Keep master password backups secure. No backdoor or recovery mechanism.

### Q: Are audit logs encrypted?

**A**: Audit logs are **not encrypted** (they contain service names, not passwords). Logs are **tamper-evident** via HMAC signatures. If encryption is required, use full-disk encryption.

### Q: Can I disable audit logging after enabling?

**A**: Yes, but audit logs remain on disk. You can manually delete old logs. Future releases may add a command to disable audit logging cleanly.

### Q: Will old Pass-CLI versions work with migrated vaults?

**A**: No. 600k iteration vaults require January 2025+ Pass-CLI versions. Keep old vault backup if you need old version compatibility.

### Q: Is there a tool to convert vault format?

**A**: Not yet. Currently requires manual re-initialization. In-place migration planned for future release.

## Best Practices

### Before Migration

1. ✅ Backup vault file: `cp ~/.pass-cli/vault.enc ~/backup/`
2. ✅ Export credentials: `pass-cli list --format json > backup.json`
3. ✅ Test new Pass-CLI version with test vault first
4. ✅ Read this migration guide completely

### During Migration

1. ✅ Use `pass-cli generate` command for policy-compliant passwords
2. ✅ Verify each credential after adding
3. ✅ Test vault unlock multiple times
4. ✅ Enable audit logging for compliance needs

### After Migration

1. ✅ Verify all credentials accessible
2. ✅ Test credential retrieval in scripts
3. ✅ Update documentation/runbooks with new requirements
4. ✅ Delete old vault backup after 30-day grace period
5. ✅ Run `pass-cli verify-audit` monthly (if audit logging enabled)

## Support

- **Documentation**: [docs/SECURITY.md](SECURITY.md), [docs/USAGE.md](USAGE.md)
- **Issues**: [GitHub Issues](https://github.com/ari1110/pass-cli/issues)

