---
title: "Recovery Phrase"
weight: 3
toc: true
---

Complete guide to using BIP39 recovery phrases to recover vault access if you forget your master password.

![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?path=docs&label=Last%20Updated)

## Overview

Pass-CLI's BIP39 recovery feature generates a 24-word recovery phrase when you create your vault. If you ever forget your master password, you can reset it using just 6 words from your recovery phrase.

**Key Benefits**:
- [OK] **Industry Standard**: Uses BIP39 (same as hardware wallets)
- [OK] **Secure**: 6 words = 73.8 quintillion combinations
- [OK] **Fast**: Recover in under 30 seconds
- [OK] **Optional**: Can skip with `--no-recovery` flag if you use keychain integration

## Setting Up Recovery

### During Vault Initialization

When you run `pass-cli init`, recovery is **enabled by default**:

```bash
$ pass-cli init
Enter master password: ****
Confirm master password: ****

[PASS] Vault created

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Recovery Phrase Setup
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Write down these 24 words in order:

 1. abandon    7. device    13. hover    19. spatial
 2. ability    8. diagram   14. hurdle   20. sphere
 3. about      9. dial      15. hybrid   21. spike
 4. above     10. diamond   16. icon     22. spin
 5. absent    11. diary     17. idea     23. spirit
 6. absorb    12. diesel    18. identify 24. split

⚠  WARNINGS:
   • Anyone with this phrase can access your vault
   • Store offline (write on paper, use a safe)
   • Recovery requires 6 random words from this list

Advanced: Add passphrase protection? (y/N): n

Verify your backup? (Y/n): y

Enter word #7: device
[PASS] (1/3)

Enter word #18: identify
[PASS] (2/3)

Enter word #22: spin
[PASS] (3/3)

[PASS] Recovery phrase verified
[PASS] Vault initialized successfully
```

### Skipping Recovery Phrase

If you prefer to rely solely on keychain integration or have another backup strategy:

```bash
# Skip recovery phrase generation
pass-cli init --no-recovery
```

**Warning**: If you skip recovery phrase generation, you cannot recover vault access if you forget your master password.

### What to Do Next

**CRITICAL**: Write down your 24-word phrase **on paper** (not digitally). Store it securely:

**Secure Storage** (Recommended):
- [OK] Physical safe or lockbox
- [OK] Safety deposit box at bank
- [OK] Fireproof/waterproof document safe at home
- [OK] Split across multiple secure locations (advanced)

**Insecure Storage** (Avoid):
- [ERROR] Digital notes apps (Apple Notes, Google Keep, etc.)
- [ERROR] Cloud storage (Dropbox, Google Drive, iCloud)
- [ERROR] Email or messaging apps
- [ERROR] Screenshots or photos on phone
- [ERROR] Password managers (defeats the purpose)

**Keep your phrase offline**. If someone gets your phrase, they can access your vault.

## Recovering Your Vault

### When to Use Recovery

Use recovery if:
- [OK] You forgot your master password
- [OK] You have your 24-word recovery phrase
- [OK] Recovery was enabled during `pass-cli init`

**Note**: If keychain is enabled and accessible, you don't need recovery. Your master password is stored securely in your OS keychain.

### Recovery Steps

#### Step 1: Run Recovery Command

```bash
pass-cli change-password --recover
```

#### Step 2: Answer Challenge Questions

You'll be asked for 6 random words from your 24-word phrase:

```bash
🔐 Vault Recovery
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
You will be asked for 6 words from your 24-word phrase.
Have your recovery phrase ready.

Enter word #18: identify
[PASS] (1/6)

Enter word #3: about
[PASS] (2/6)

Enter word #22: spin
[PASS] (3/6)

Enter word #7: device
[PASS] (4/6)

Enter word #11: diary
[PASS] (5/6)

Enter word #15: hybrid
[PASS] (6/6)

[PASS] Recovery phrase verified
[PASS] Vault unlocked
```

#### Step 3: Set New Master Password

```bash
Enter new master password: ****
Confirm new master password: ****

[PASS] Master password changed successfully
Your vault has been re-encrypted with the new password.
```

#### Step 4: Done!

Your vault is now accessible with your new master password. If keychain integration is enabled, the new password is automatically stored in your OS keychain.

### Recovery Tips

- **Use your written phrase**: Look at the paper where you wrote down the 24 words
- **Position numbers matter**: "word #7" refers to the 7th word in your original list
- **Order is randomized**: The system asks for words in random order each time
- **Typos caught immediately**: Invalid words (not in BIP39 wordlist) are rejected instantly
- **6 attempts maximum**: After 6 failed attempts, you must wait before trying again

## Security Best Practices

### Secure Storage

**Physical Security**:
- Write recovery phrase on **archival-quality paper** (acid-free, long-lasting)
- Use **permanent ink** (not pencil, which can smudge)
- Store in **fireproof/waterproof safe** or safety deposit box
- Consider **metal backup plates** for extreme durability

**Redundancy**:
- Keep **multiple copies** in separate secure locations
- **Don't** keep all copies in same building (fire/flood risk)
- **Don't** store with vault or on same device

**Access Control**:
- Only trusted family members should know where phrase is stored
- Consider **sealed envelope** with tamper-evident security
- Update beneficiaries if phrase location changes

### What Never to Do

**Never Store Digitally**:
- [ERROR] Photos or screenshots
- [ERROR] Cloud storage services
- [ERROR] Email or messaging apps
- [ERROR] Password managers
- [ERROR] Digital note-taking apps

**Never Share**:
- [ERROR] Don't tell anyone your recovery phrase
- [ERROR] Pass-CLI will never ask for your full phrase
- [ERROR] No support person needs your recovery phrase
- [ERROR] Recovery phrase = full vault access

**Never Memorize Only**:
- [ERROR] Human memory is fallible
- [ERROR] Always have physical backup
- [ERROR] Don't rely on memory alone

### Testing Your Backup

After writing down your recovery phrase:

1. **Verify you wrote it correctly** during initialization (3-word challenge)
2. **Store phrase securely** before testing recovery
3. **Optional**: Test recovery in safe environment:
   ```bash
   # Backup existing config (if any)
   cp ~/.pass-cli/config.yml ~/.pass-cli/config.yml.backup 2>/dev/null

   # Point to temporary test vault
   echo "vault_path: /tmp/test-vault.enc" > ~/.pass-cli/config.yml

   # Create test vault and test recovery
   pass-cli init
   pass-cli change-password --recover

   # Restore original config and clean up
   mv ~/.pass-cli/config.yml.backup ~/.pass-cli/config.yml 2>/dev/null || rm ~/.pass-cli/config.yml
   rm -f /tmp/test-vault.enc
   ```

## Advanced: Passphrase Protection

### What is a Passphrase?

A **BIP39 passphrase** (sometimes called the "25th word") is an additional secret you can add to your recovery phrase. It's like a second password that works alongside your 24 words.

### How to Enable

During `pass-cli init`:

```bash
Advanced: Add passphrase protection? (y/N): y

Enter passphrase (optional 25th word): ****
Confirm passphrase: ****

[PASS] Passphrase protection enabled
```

### Security Trade-offs

**Benefits**:
- [OK] Even if someone finds your 24 words, they still need the passphrase
- [OK] Plausible deniability (can have multiple vaults with same phrase + different passphrases)
- [OK] Extra layer of security

**Risks**:
- [ERROR] If you lose the passphrase, you **cannot** recover your vault
- [ERROR] Must remember/store passphrase separately from 24-word phrase
- [ERROR] More complex recovery process

**Recommendation**: Only use passphrase protection if you:
- Understand the risks
- Have secure way to store passphrase separately
- Are comfortable with added complexity

## Troubleshooting

### "Recovery phrase not enabled for this vault"

**Cause**: Vault was initialized with `--no-recovery` flag.

**Solution**: Recovery is not possible. You must remember your master password or restore from backup.

### "Invalid recovery word"

**Cause**: Word you entered is not in the BIP39 wordlist or doesn't match your phrase.

**Solutions**:
1. Check spelling carefully
2. Verify word position (word #7 = 7th word in your list)
3. Ensure you're reading from correct recovery phrase backup
4. Try typing word manually (not copy-paste)

### "Recovery verification failed"

**Cause**: Too many incorrect words entered.

**Solutions**:
1. Double-check your written recovery phrase
2. Verify you're using the correct vault
3. Ensure recovery phrase hasn't been transcribed incorrectly
4. If phrase is correct but failing, vault may be corrupted (restore from backup)

### Lost Recovery Phrase

**Unfortunately**: If you've lost your recovery phrase AND forgotten your master password, your vault is unrecoverable.

**Options**:
1. Check all secure storage locations
2. Check with trusted family members
3. If truly lost, you must reinitialize vault and re-add credentials

**Prevention**:
- Store recovery phrase in multiple secure locations
- Tell trusted family member where phrase is stored
- Include phrase location in your estate planning

## See Also

- [Security Architecture](../03-reference/security-architecture.md#bip39-recovery-phrase) - Technical details of BIP39 implementation
- [Command Reference](../03-reference/command-reference.md#change-password---change-master-password) - `change-password --recover` command
- [Keychain Setup](keychain-setup.md) - Alternative to recovery phrase (OS keychain integration)
- [Backup & Restore](backup-restore.md) - Vault backup strategies
