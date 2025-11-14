# Quick Start: BIP39 Mnemonic Recovery

**Feature**: Vault Recovery with 24-Word Phrase
**Version**: 1.0.0
**Audience**: pass-cli users

## Overview

Never lose access to your vault again. The BIP39 recovery feature generates a 24-word recovery phrase when you create your vault. If you ever forget your master password (and keychain is disabled), you can reset it using just 6 words from your recovery phrase.

**Key Benefits**:
- ✅ **Industry Standard**: Uses BIP39 (same as hardware wallets)
- ✅ **Secure**: 6 words = 73.8 quintillion combinations
- ✅ **Fast**: Recover in under 30 seconds
- ✅ **Optional**: Can skip if you use keychain integration

---

## Table of Contents

1. [Setting Up Recovery](#1-setting-up-recovery)
2. [Recovering Your Vault](#2-recovering-your-vault)
3. [Security Best Practices](#3-security-best-practices)
4. [Advanced: Passphrase Protection](#4-advanced-passphrase-protection)
5. [FAQ](#5-faq)

---

## 1. Setting Up Recovery

### During Vault Initialization

When you run `pass-cli init`, recovery is enabled by default:

```bash
$ pass-cli init
Enter master password: ****
Confirm master password: ****

✓ Vault created

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
✓ (1/3)

Enter word #18: identify
✓ (2/3)

Enter word #22: spin
✓ (3/3)

✓ Recovery phrase verified
✓ Vault initialized successfully
```

### What to Do Next

**CRITICAL**: Write down your 24-word phrase **on paper** (not digitally). Store it securely:
- ✅ Physical safe or lockbox
- ✅ Safety deposit box
- ✅ Secure home location (fireproof safe)
- ❌ Digital notes app
- ❌ Cloud storage
- ❌ Email

**Keep your phrase offline**. If someone gets your phrase, they can access your vault.

---

## 2. Recovering Your Vault

### When to Use Recovery

Use recovery if:
- ✅ You forgot your master password
- ✅ Keychain integration is not enabled
- ✅ You have your 24-word recovery phrase

**Note**: If keychain is enabled, you don't need recovery. Your master password is stored securely in your OS keychain.

### Recovery Steps

1. **Run the recovery command**:

```bash
$ pass-cli change-password --recover
```

2. **Enter 6 random words** (in randomized order):

```bash
Vault Recovery
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
You will be asked for 6 words from your 24-word phrase.
Have your recovery phrase ready.

Enter word #18: identify
✓ (1/6)

Enter word #3: about
✓ (2/6)

Enter word #22: spin
✓ (3/6)

Enter word #7: device
✓ (4/6)

Enter word #11: diary
✓ (5/6)

Enter word #15: hybrid
✓ (6/6)

✓ Recovery phrase verified
✓ Vault unlocked
```

3. **Set a new master password**:

```bash
Enter new master password: ****
Confirm new master password: ****

✓ Master password changed successfully
```

4. **Done!** Your vault is unlocked with your new password.

### Tips for Recovery

- **Use your written phrase**: Look at the paper where you wrote down the 24 words.
- **Position numbers matter**: The system asks for "word #7", "word #18", etc. These refer to the position in your original 24-word list.
- **Order is randomized**: The system asks for words in random order each time (prevents memorization).
- **Typos caught immediately**: If you enter an invalid word (not in the BIP39 wordlist), you'll get instant feedback.

---

## 3. Security Best Practices

### Do's ✅

- **Write on paper**: Use pen and paper (preferably archival-quality).
- **Store offline**: Physical safe, safety deposit box, or secure home location.
- **Keep separate from vault**: Don't store recovery phrase on the same device as your vault.
- **Test recovery once**: After setup, optionally test recovery in a safe environment to verify you wrote down the phrase correctly.
- **Update if compromised**: If someone sees your phrase, immediately recover your vault and generate a new phrase (future feature).

### Don'ts ❌

- **Never type into untrusted devices**: Only enter your recovery phrase into pass-cli on trusted computers.
- **Never store digitally**: No photos, screenshots, cloud notes, password managers, or emails.
- **Never share**: Your recovery phrase is equivalent to your master password. Never share it.
- **Never memorize only**: Human memory is fallible. Always have a physical backup.
- **Don't lose it**: Without recovery phrase AND master password, your vault is unrecoverable.

### Passphrase Protection (Optional)

For advanced users, you can add a **25th word** (passphrase) during setup. This adds extra security:

- **Benefit**: Even if someone finds your 24 words, they need the passphrase too.
- **Risk**: If you lose the passphrase, you can't recover your vault.

See [Section 4](#4-advanced-passphrase-protection) for details.

---

## 4. Advanced: Passphrase Protection

### What is a Passphrase?

A **BIP39 passphrase** (sometimes called the "25th word") is an additional secret you can add to your recovery phrase. It's like a second password that works alongside your 24 words.

### How to Enable

During `pass-cli init`:

```bash
Advanced: Add passphrase protection? (y/N): y

⚠  PASSPHRASE NOTICE:
   This adds a 25th word for extra security.
   • Increases security (attacker needs phrase + passphrase)
   • YOU MUST REMEMBER/STORE THIS SEPARATELY
   • Losing the passphrase = losing vault access

Enter passphrase: ****
Confirm passphrase: ****

✓ Passphrase set
```

### During Recovery

If you enabled a passphrase, the system will detect it automatically:

```bash
Enter word #15: hybrid
✓ (6/6)

Enter recovery passphrase: ****

✓ Recovery phrase verified
✓ Vault unlocked
```

### Security Trade-offs

| Without Passphrase | With Passphrase |
|--------------------|-----------------|
| **Security**: 24 words (2^256 entropy) | **Security**: 24 words + passphrase (additional entropy) |
| **Recovery**: Need 24-word phrase only | **Recovery**: Need 24-word phrase **AND** passphrase |
| **Risk**: If phrase compromised, vault compromised | **Risk**: If phrase OR passphrase lost, vault unrecoverable |

**Recommendation**: Only use passphrase if you:
- Understand the security trade-offs
- Have a secure way to store the passphrase separately from the 24 words
- Are comfortable with the added complexity

**For most users**: The 24-word phrase alone provides sufficient security.

---

## 5. FAQ

### Q: Can I skip recovery setup?

**A**: Yes. Use `pass-cli init --no-recovery` to skip. However, if you forget your master password and keychain is disabled, your vault will be unrecoverable.

```bash
$ pass-cli init --no-recovery
```

### Q: How secure is the 6-word challenge?

**A**: Very secure. 6 words from a 2,048-word list = 2^66 combinations (73.8 quintillion). Even with GPU clusters, brute force would take millions of years.

### Q: Why only 6 words instead of all 24?

**A**: Security + UX balance. 6 words provides 2^66 entropy (secure against all known attacks) while being faster to enter than 24 words. The "puzzle piece" approach: you provide 6 words, the system unlocks the encrypted 18 remaining words, then reconstructs the full 24-word phrase to unlock your vault.

### Q: What if I skip verification during init?

**A**: You can, but it's risky. If you wrote down a word incorrectly, you won't discover it until you try to recover (potentially too late).

```bash
Verify your backup? (Y/n): n

⚠  WARNING: Skipping verification. Ensure you have written down
   all 24 words correctly before continuing.
```

### Q: Can I use my recovery phrase with other tools?

**A**: Yes! Your 24-word phrase is a standard BIP39 mnemonic. You can:
- Verify it with online BIP39 tools (e.g., iancoleman.io/bip39) - **USE OFFLINE ONLY**
- Import it into hardware wallets (for verification, not recommended for actual use)
- Generate the same seed with any BIP39-compatible tool

**Warning**: Never enter your recovery phrase into untrusted online tools. Use offline/air-gapped verification only.

### Q: What if I lose my recovery phrase?

**A**: If you lose your recovery phrase AND forget your master password (and keychain is disabled), your vault is **permanently unrecoverable**. There is no backdoor. This is by design (security-first).

**Prevention**: Store your recovery phrase securely (see [Section 3](#3-security-best-practices)).

### Q: Can I change my recovery phrase?

**A**: Not in v1.0. Future versions may support recovery phrase rotation (generate new 24 words). For now, if your phrase is compromised, you must:
1. Recover your vault with current phrase
2. Delete vault
3. Re-initialize with `pass-cli init` (generates new phrase)
4. Re-add credentials

### Q: Why are words asked in random order?

**A**: Security + verification. Random order:
- Prevents muscle memory (forces you to reference your written phrase)
- Verifies you actually have the phrase written down (not just memorized positions)
- Adds minor additional security (observers can't deduce full phrase from watching)

### Q: What happens if I enter a wrong word?

**A**: If you enter a word that's not in the BIP39 wordlist, you'll get immediate feedback:

```bash
Enter word #7: devic
✗ Invalid word. Try again.

Enter word #7: device
✓ (1/6)
```

If you enter a valid word but it's the wrong word for that position, decryption will fail after all 6 words are entered:

```bash
✗ Recovery failed: Incorrect recovery words

Try again? (Y/n): y
```

### Q: Can I test recovery without changing my password?

**A**: Not in v1.0 (future feature). To test recovery:
1. Create a test vault (`pass-cli init --vault-path /tmp/test-vault.enc`)
2. Write down recovery phrase
3. Test `pass-cli change-password --recover --vault-path /tmp/test-vault.enc`
4. Verify it works
5. Delete test vault

### Q: Is the recovery phrase the same as my master password?

**A**: No. They are separate:
- **Master Password**: Unlocks your vault during normal usage (`pass-cli get`, `pass-cli add`, etc.)
- **Recovery Phrase**: Allows password reset if you forget master password

Think of it like:
- Master password = front door key (daily use)
- Recovery phrase = locksmith's master key (emergency use)

### Q: What if someone steals my recovery phrase?

**A**: They can access your vault. Immediately:
1. Recover your vault (if they haven't already)
2. Change your master password to a new one
3. Generate a new recovery phrase (re-initialize vault)
4. Update all credentials stored in vault (assume they may be compromised)

**Prevention**: Store recovery phrase as securely as you'd store the credentials themselves.

---

## Quick Reference

### Commands

| Command | Description |
|---------|-------------|
| `pass-cli init` | Create vault with recovery (default) |
| `pass-cli init --no-recovery` | Create vault without recovery |
| `pass-cli change-password --recover` | Recover vault using recovery phrase |

### Key Concepts

| Concept | Description |
|---------|-------------|
| **24-Word Phrase** | BIP39 mnemonic generated during vault creation. Write down and store securely. |
| **6-Word Challenge** | During recovery, you provide 6 randomly-selected words from your 24-word phrase. |
| **Passphrase (25th Word)** | Optional additional secret for extra security. Store separately from 24 words. |
| **Verification** | During init, system asks for 3 random words to confirm you wrote down the phrase correctly. |

### Security Checklist

- [ ] Written down all 24 words on paper
- [ ] Stored paper in secure location (safe, lockbox, etc.)
- [ ] Verified backup during initialization (or tested recovery later)
- [ ] Never typed recovery phrase into untrusted devices
- [ ] Never stored recovery phrase digitally (photos, notes, cloud)
- [ ] If using passphrase: stored it separately from 24 words

---

## Getting Help

**Documentation**: See full docs at `docs/features/recovery.md`
**Issues**: Report bugs at [GitHub Issues](https://github.com/yourusername/pass-cli/issues)
**Security**: Report security issues privately to security@yourproject.com

---

**Remember**: Your recovery phrase is the last line of defense. Treat it as carefully as you'd treat the passwords it protects.
