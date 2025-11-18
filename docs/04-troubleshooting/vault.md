---
title: "Vault Issues"
weight: 2
toc: true
---

Solutions for vault access, corruption, and recovery issues.

## Vault Access Issues

### "Invalid Master Password" Error

**Symptom**: Password rejected when accessing vault

**Cause**: Incorrect password or vault corruption

**Solutions**:

1. **Verify password**
   - Check caps lock
   - Try typing slowly
   - Copy-paste if stored elsewhere

2. **Check keychain entry**
   ```bash
   # macOS: View in Keychain Access
   # Linux: View in Seahorse/KWallet
   # Windows: View in Credential Manager
   ```

3. **Restore from backup**
   ```bash
   # If vault is corrupted
   cp ~/.pass-cli/vault.enc.backup ~/.pass-cli/vault.enc
   ```

4. **Try manual backup**
   ```bash
   # If you have manual backup
   cp ~/backups/vault-20250120.enc ~/.pass-cli/vault.enc
   ```

---

### "Vault File Corrupted" Error

**Symptom**: Cannot decrypt vault, corruption detected

**Cause**: File corruption from crash or disk error

**Solutions**:

1. **Restore automatic backup**
   ```bash
   # Check if backup exists
   ls -la ~/.pass-cli/vault.enc.backup

   # Restore
   cp ~/.pass-cli/vault.enc.backup ~/.pass-cli/vault.enc

   # Test access
   pass-cli list
   ```

2. **Restore manual backup**
   ```bash
   # List available backups
   ls -la ~/backups/vault-*.enc

   # Restore most recent
   cp ~/backups/vault-20250120.enc ~/.pass-cli/vault.enc
   ```

3. **If no backup available**
   ```bash
   # Unfortunately, corrupted vault without backup is unrecoverable
   # Initialize new vault
   mv ~/.pass-cli/vault.enc ~/.pass-cli/vault.enc.corrupted
   pass-cli init
   # Re-add credentials manually
   ```

---

### "Permission Denied" Reading Vault

**Symptom**: Cannot read vault file

**Cause**: File permission issues

**Solutions**:

**macOS/Linux:**
```bash
# Check permissions
ls -la ~/.pass-cli/vault.enc

# Fix permissions (should be 0600)
chmod 600 ~/.pass-cli/vault.enc

# Fix ownership
sudo chown $USER:$USER ~/.pass-cli/vault.enc
```

**Windows:**
```powershell
# Check ACL
Get-Acl "$env:USERPROFILE\.pass-cli\vault.enc" | Format-List

# Reset permissions to current user
$acl = Get-Acl "$env:USERPROFILE\.pass-cli\vault.enc"
$acl.SetAccessRuleProtection($true, $false)
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule(
    $env:USERNAME, "FullControl", "Allow"
)
$acl.AddAccessRule($rule)
Set-Acl "$env:USERPROFILE\.pass-cli\vault.enc" $acl
```

---

## Vault Recovery

### Forgot Master Password

**Symptom**: Cannot remember master password

**Options**:

1. **Use BIP39 Recovery Phrase** (if enabled during init)
   ```bash
   # Recover access with recovery phrase
   pass-cli change-password --recover
   ```
   
   **Requirements**:
   - Recovery phrase must have been enabled during `pass-cli init`
   - You must have your 24-word recovery phrase written down
   - You'll need to provide 6 random words from the phrase
   
   **Process**:
   1. Run `pass-cli change-password --recover`
   2. System prompts for 6 random words from your 24-word phrase
   3. If verification succeeds, you can set a new master password
   4. Vault is re-encrypted with new password
   
   For detailed recovery instructions, see [Recovery Phrase Guide](../../02-guides/recovery-phrase.md).

2. **Check keychain** (if still accessible)
   - macOS: Keychain Access → search "pass-cli"
   - Linux: Seahorse → search "pass-cli"
   - Windows: Credential Manager → search "pass-cli"

3. **Try to remember**
   - Try common variations
   - Check password manager if stored there
   - Check secure notes

4. **If recovery phrase not available or wasn't enabled**
   ```bash
   # Vault is unrecoverable without master password or recovery phrase
   # Start fresh
   mv ~/.pass-cli/vault.enc ~/.pass-cli/vault.enc.lost
   pass-cli init
   # Re-add credentials from services
   ```

**Prevention**:
- [OK] Enable recovery phrase during init (it's enabled by default)
- [OK] Write recovery phrase on paper and store in safe
- [OK] Write master password in secure location
- [OK] Store master password in another password manager
- [OK] Keep backup of master password

---

### Vault File Deleted

**Symptom**: Vault file missing

**Solutions**:

1. **Check trash/recycle bin**

2. **Restore from backup**
   ```bash
   # Automatic backup
   cp ~/.pass-cli/vault.enc.backup ~/.pass-cli/vault.enc

   # Manual backup
   cp ~/backups/vault-*.enc ~/.pass-cli/vault.enc
   ```

3. **Restore from cloud sync** (if syncing vault)
   ```bash
   # From Dropbox, Google Drive, etc.
   cp ~/Dropbox/vault.enc ~/.pass-cli/vault.enc
   ```

4. **If no backup**
   ```bash
   # Unfortunately, must start over
   pass-cli init
   ```

---

### Corrupt Vault Recovery

**Symptom**: Vault fails to decrypt or shows corruption errors

**Solutions**:

1. **Try automatic backup**
   ```bash
   cp ~/.pass-cli/vault.enc.backup ~/.pass-cli/vault.enc
   pass-cli list
   ```

2. **Try older backups**
   ```bash
   # List backups by date
   ls -lt ~/backups/vault-*.enc

   # Try each, newest first
   cp ~/backups/vault-20250120.enc ~/.pass-cli/vault.enc
   pass-cli list
   ```

3. **Attempt partial recovery** (advanced)
   ```bash
   # Examine vault file
   hexdump -C ~/.pass-cli/vault.enc | head -n 20

   # Check file size
   ls -la ~/.pass-cli/vault.enc

   # If file is obviously truncated or wrong size
   # Recovery likely impossible, use backup
   ```

---

### Vault Save Interrupted / Orphaned Temporary Files

**Symptom**: Save operation interrupted (crash, power loss, kill signal) or temporary files (`vault.enc.tmp.*`) left behind

**Cause**: Atomic save process interrupted mid-operation

**Background**: Pass-CLI uses an atomic save pattern to protect against corruption:
1. Writes to temporary file (`vault.enc.tmp.TIMESTAMP.RANDOM`)
2. Verifies temp file is decryptable
3. Renames current vault to backup (`vault.enc.backup`)
4. Renames temp file to vault (`vault.enc`)

If interrupted, your vault remains safe at step 1-2 (old vault unchanged) or step 3-4 (backup exists).

**Solutions**:

1. **Check vault status after interruption**
   ```bash
   # Check which files exist
   ls -la ~/.pass-cli/

   # Files you may see:
   # vault.enc                  - Current vault (may be old or new)
   # vault.enc.backup           - Previous vault state
   # vault.enc.tmp.20250109-*   - Orphaned temp file from crash
   ```

2. **Verify current vault is accessible**
   ```bash
   # Try to unlock vault
   pass-cli list

   # If this works, vault is fine
   # Orphaned temp files can be safely deleted
   ```

3. **If current vault fails to unlock**
   ```bash
   # Check if backup exists
   ls -la ~/.pass-cli/vault.enc.backup

   # Restore from backup (contains N-1 state)
   cp ~/.pass-cli/vault.enc.backup ~/.pass-cli/vault.enc

   # Test restored vault
   pass-cli list

   # You may have lost your most recent changes,
   # but vault is recovered
   ```

4. **Clean up orphaned temporary files**
   ```bash
   # List temp files
   ls -la ~/.pass-cli/vault.enc.tmp.*

   # Remove orphaned temp files (safe after vault verified)
   rm ~/.pass-cli/vault.enc.tmp.*

   # Note: pass-cli automatically cleans these up
   # on next save operation
   ```

5. **Identify what was lost**
   ```bash
   # If you had to restore from backup, you lost:
   # - Most recent credential addition/update/deletion
   # - Most recent password change (if in progress)

   # Check audit log if enabled
   tail -20 ~/.pass-cli/audit.log

   # Last successful operation shown in audit log
   ```

**Recovery Decision Matrix**:
- `vault.enc` accessible → Keep it, delete temp files
- `vault.enc` corrupted, `vault.enc.backup` exists → Restore backup
- Both corrupted, `vault.enc.tmp.*` exists → Try unlocking temp file
- All corrupted → Use external backup or reinitialize

**Prevention**:
- Regular backups to external location
- Don't force-kill pass-cli during save operations
- Ensure sufficient disk space before operations
- Use UPS for desktop systems

---

### Understanding Vault Backup Lifecycle

**Q: When is `vault.enc.backup` created and removed?**

**A: Backup lifecycle with atomic save**:

1. **Backup Created**: During save operation
   ```bash
   # Before save:
   vault.enc (old data)

   # During save:
   vault.enc.tmp.20250109-143022.a1b2c3 (new data being written)

   # After verification passes:
   vault.enc → renamed to → vault.enc.backup (old data)
   vault.enc.tmp.* → renamed to → vault.enc (new data)
   ```

2. **Backup Removed**: After next successful unlock
   ```bash
   # On next vault unlock:
   # - Vault decrypts successfully
   # - Backup is removed (no longer needed)
   # - Only vault.enc remains
   ```

3. **Why this pattern?**
   - Backup proves old vault was valid
   - If new vault fails to decrypt, backup still available
   - After successful unlock, new vault proven good
   - Backup cleanup saves disk space

**Implications**:
- Backup contains N-1 generation (one save ago)
- If save fails, backup has your last good state
- If vault becomes corrupted between saves, backup may help
- Backup removed after successful unlock (not kept permanently)

**Manual Backup Recommended**:
```bash
# Backup before risky operations
cp ~/.pass-cli/vault.enc ~/backups/vault-$(date +%Y%m%d).enc

# Automated daily backup
echo '0 0 * * * cp ~/.pass-cli/vault.enc ~/backups/vault-$(date +\%Y\%m\%d).enc' | crontab -
```

---

