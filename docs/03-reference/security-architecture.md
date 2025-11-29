---
title: "Security Architecture"
weight: 3
toc: true
---
Comprehensive security architecture, cryptographic implementation, threat model, and security guarantees for Pass-CLI.

![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?path=docs&label=Last%20Updated)

## Security Overview

Pass-CLI is designed with security as the primary concern. All credentials are encrypted using industry-standard cryptography and stored locally on your machine with no cloud dependencies.

### Key Security Features

- **AES-256-GCM Encryption**: Military-grade authenticated encryption
- **PBKDF2 Key Derivation**: 600,000 iterations with SHA-256 (hardened, January 2025)
- **BIP39 Recovery Phrase**: 24-word mnemonic for vault password recovery (industry-standard)
- **System Keychain Integration**: Secure master password storage
- **Offline-First Design**: No network calls, no cloud dependencies
- **Secure Memory Handling**: Byte-based password handling with immediate zeroing
- **Password Policy Enforcement**: Complexity requirements for vault and credential passwords
- **Tamper-Evident Audit Logging**: HMAC-SHA256 signed audit trail (optional)
- **File Permission Protection**: Vault files restricted to user-only access
- **Atomic Vault Operations**: Rollback safety for vault updates

## Cryptographic Implementation

### Encryption Algorithm

**AES-256-GCM (Galois/Counter Mode)**

- **Algorithm**: Advanced Encryption Standard
- **Key Size**: 256 bits (32 bytes)
- **Mode**: GCM (Galois/Counter Mode)
- **Authentication**: Built-in GMAC authentication tag
- **Implementation**: Go standard library `crypto/aes` and `crypto/cipher`

#### Why AES-256-GCM?

1. **NIST Approved**: Recommended by NIST for classified information
2. **Authenticated Encryption**: Prevents tampering and chosen-ciphertext attacks
3. **Parallelizable**: Fast performance on modern hardware
4. **Standard**: Widely used and well-audited implementation

### Key Derivation

**PBKDF2-SHA256**

- **Algorithm**: Password-Based Key Derivation Function 2
- **Hash Function**: SHA-256
- **Iterations**: 600,000 (increased from 100,000 in January 2025)
- **Salt Length**: 32 bytes (256 bits)
- **Output Length**: 32 bytes (256 bits)
- **Implementation**: `golang.org/x/crypto/pbkdf2`
- **Performance**: ~50-100ms on modern CPUs (2023+), 500-1000ms on older hardware

#### Key Derivation Process

```text
Master Key = PBKDF2(
    password = user's master password,
    salt = unique 32-byte random salt,
    iterations = 600,000,
    hash = SHA-256,
    key_length = 32 bytes
)
```

#### Why PBKDF2?

1. **Computationally Expensive**: 600,000 iterations significantly slow down brute-force attacks
2. **Salted**: Unique salt prevents rainbow table attacks
3. **Standard**: NIST recommended for password-based cryptography
4. **Deterministic**: Same password + salt = same key

#### Migration from 100k to 600k Iterations

- **Backward Compatibility**: Vaults with 100k iterations continue to work
- **Automatic Detection**: Iteration count stored in vault metadata
- **Migration Path**: Manual migration required (export credentials, reinitialize vault, re-import)
- **See**: `docs/MIGRATION.md` for detailed upgrade instructions

### Encryption Process

#### Encrypting Credentials

1. **Generate Salt** (first time only)
   ```text
   salt = crypto/rand.Read(32 bytes)
   ```

2. **Derive Encryption Key**
   ```text
   key = PBKDF2(master_password, salt, 600000, SHA256, 32)
   ```

3. **Generate Nonce**
   ```text
   nonce = crypto/rand.Read(12 bytes)  // Per-encryption unique
   ```

4. **Encrypt Data**
   ```text
   ciphertext = AES-256-GCM.Encrypt(
       plaintext = JSON(credentials),
       key = derived_key,
       nonce = nonce,
       additional_data = nil
   )
   ```

5. **Combine Components**
   ```text
   vault_file = nonce || ciphertext || auth_tag
   ```

#### Decrypting Credentials

1. **Load Master Password** from system keychain
2. **Read Vault File** and extract salt, nonce, ciphertext
3. **Derive Key** using PBKDF2 with stored salt
4. **Decrypt and Verify**
   ```text
   plaintext = AES-256-GCM.Decrypt(
       ciphertext,
       key,
       nonce
   )
   ```
5. **Parse JSON** to access credentials

### Random Number Generation

All random values use `crypto/rand`, which provides cryptographically secure random numbers from the operating system:

- **Windows**: `CryptGenRandom`
- **macOS/Linux**: `/dev/urandom`

Used for:
- Salt generation
- Nonce generation
- Password generation

## Master Password Management

### System Keychain Integration

Pass-CLI integrates with your operating system's secure credential storage to save your master password.

#### Windows - Credential Manager

- **Location**: Windows Credential Manager
- **Storage**: Encrypted by Windows using DPAPI
- **Access**: Protected by user's Windows login
- **Implementation**: `github.com/zalando/go-keyring`

**Viewing in Windows:**
1. Open Control Panel
2. User Accounts → Credential Manager
3. Windows Credentials
4. Look for "pass-cli" entry

#### macOS - Keychain

- **Location**: macOS Keychain (login keychain)
- **Storage**: Encrypted by macOS keychain services
- **Access**: Protected by user's macOS login password
- **Implementation**: `github.com/zalando/go-keyring`

**Viewing on macOS:**
1. Open Keychain Access app
2. Search for "pass-cli"
3. Double-click to view (requires password)

#### Linux - Secret Service

- **Backend**: GNOME Keyring, KWallet, or compatible
- **Protocol**: freedesktop.org Secret Service API
- **Storage**: Encrypted by keyring daemon
- **Access**: Protected by keyring password
- **Implementation**: `github.com/zalando/go-keyring`

**Viewing on Linux (GNOME):**
1. Open Seahorse (Passwords and Keys)
2. Login keyring
3. Search for "pass-cli"

### Master Password Requirements

**Since January 2025** - Password policy enforced for both vault and credential passwords:

- **Minimum Length**: 12 characters (enforced)
- **Uppercase Letter**: At least one required
- **Lowercase Letter**: At least one required
- **Digit**: At least one required
- **Special Symbol**: At least one required (!@#$%^&*()-_=+[]{}|;:,.<>?)
- **Recommended Length**: 20+ characters for master password
- **Strength Indicator**: Real-time feedback in TUI mode

### Master Password Security

**What Pass-CLI Does:**
- [OK] Stores master password in system keychain
- [OK] Clears password from memory after use
- [OK] Never writes password to disk in plaintext
- [OK] Never logs password

**What You Should Do:**
- [OK] Use a unique master password (not reused elsewhere)
- [OK] Make it strong (20+ characters or passphrase)
- [OK] Store backup securely (password manager, safe place)
- [OK] Save your BIP39 recovery phrase offline (paper, safe)
- [ERROR] Don't share your master password
- [ERROR] Don't write it in plaintext files

### BIP39 Recovery Phrase

Pass-CLI supports optional BIP39 recovery phrases to recover vault access if you forget your master password. This feature uses the industry-standard BIP39 mnemonic specification (same as hardware wallets).

#### How It Works

**During Initialization:**
1. Generate 24-word BIP39 mnemonic phrase
2. Derive recovery key from mnemonic using PBKDF2
3. Encrypt recovery key with master password
4. Store encrypted recovery key in vault metadata

**During Recovery:**
1. User provides 6 random words from their 24-word phrase (challenge)
2. System verifies the words match the stored mnemonic
3. Derive recovery key from complete mnemonic
4. Decrypt vault metadata to verify recovery key
5. Allow user to set new master password

#### Security Properties

- **Challenge-Response**: 6 random words = 2^66 possible combinations (~73.8 quintillion)
- **Offline Storage**: Recovery phrase should be written on paper, not stored digitally
- **Optional Feature**: Can be skipped during initialization with `--no-recovery` flag
- **Passphrase Protection**: Optional 25th word for additional security
- **No Backdoor**: Recovery phrase is user-generated and user-stored only

#### Commands

```bash
# Initialize vault with recovery phrase (default)
pass-cli init

# Initialize vault without recovery phrase
pass-cli init --no-recovery

# Recover access if password forgotten
pass-cli change-password --recover
```

#### Storage Recommendations

**Secure Storage** (Recommended):
- [OK] Write on paper and store in physical safe
- [OK] Safety deposit box
- [OK] Fireproof/waterproof document safe
- [OK] Split across multiple secure locations (advanced)

**Insecure Storage** (Avoid):
- [ERROR] Digital notes apps
- [ERROR] Cloud storage (Dropbox, Google Drive)
- [ERROR] Email or messaging apps
- [ERROR] Screenshots or photos
- [ERROR] Password managers (defeats the purpose)

**Important**: Anyone with your 24-word phrase can access your vault. Protect it as carefully as your master password.

For detailed recovery procedures, see [Recovery Phrase Guide](../02-guides/recovery-phrase.md).

## Data Storage Security

### Vault File Location

- **Windows**: `%USERPROFILE%\.pass-cli\vault.enc`
- **macOS/Linux**: `~/.pass-cli/vault.enc`

### File Permissions

Vault files are created with restricted permissions:

- **Unix (macOS/Linux)**: `0600` (owner read/write only)
- **Windows**: ACL restricting to current user

### Vault File Structure

```text
+------------------+
| Salt (32 bytes)  |  ← PBKDF2 salt
+------------------+
| Nonce (12 bytes) |  ← AES-GCM nonce
+------------------+
| Ciphertext       |  ← Encrypted credentials (variable length)
+------------------+
| Auth Tag         |  ← GCM authentication tag (16 bytes)
+------------------+
```

### Atomic Writes

Vault updates use atomic write operations to prevent corruption:

1. Write to temporary file (`.vault.enc.tmp`)
2. Sync to disk (`fsync`)
3. Rename to actual vault file (atomic operation)
4. Delete temporary file on error

This ensures:
- No partial writes
- No corruption on crash
- Previous vault preserved on error

### Backup Strategy

**Automatic Backup Files** (since atomic save implementation):

Before each vault save operation, pass-cli creates an N-1 backup:
1. New vault data written to temporary file (`vault.enc.tmp.TIMESTAMP.RANDOM`)
2. Temporary file verified (decryption test)
3. Current vault renamed to `vault.enc.backup` (N-1 generation)
4. Temporary file renamed to `vault.enc` (becomes current)
5. Backup removed after next successful unlock (confirms new vault works)

**Security Implications**:

- [WARNING] **Backup files contain unencrypted vault structure**: `vault.enc.backup` is AES-256-GCM encrypted (same as vault), but still sensitive
- [OK] **File permissions**: Backup automatically inherits vault permissions (0600 - owner read/write only)
- [WARNING] **Temporary files**: `vault.enc.tmp.*` files may remain if process crashes (cleaned up automatically on next save)
- [OK] **Automatic cleanup**: Backup removed after successful unlock, minimizing exposure window
- [WARNING] **Contains N-1 state**: Backup has previous vault version (not current), may contain deleted credentials

**Manual Backup Recommendations**:

```bash
# Create timestamped backups (recommended)
cp ~/.pass-cli/vault.enc ~/backups/vault-$(date +%Y%m%d).enc

# Set correct permissions on manual backups
chmod 600 ~/backups/vault-*.enc

# Store backups on encrypted drive or secure location
# Do NOT store in cloud storage without additional encryption
```

**What Files May Exist**:
- `vault.enc` - Current encrypted vault (always present when unlocked)
- `vault.enc.backup` - Previous vault state (present between saves, removed after unlock)
- `vault.enc.tmp.YYYYMMDD-HHMMSS.XXXXXX` - Orphaned temp files from crashes (auto-cleaned)

### Audit Logging (Optional)

**Since January 2025** - Tamper-evident audit trail for vault operations:

- **Opt-In**: Disabled by default, enable with `--enable-audit` flag
- **HMAC Signatures**: HMAC-SHA256 signatures for tamper detection
- **Key Storage**: Audit HMAC keys stored in OS keychain (separate from vault)
- **Events Logged**: Vault unlock/lock, password changes, credential operations
- **Privacy**: Service names logged, passwords NEVER logged
- **Rotation**: Automatic log rotation at 10MB, 7-day retention
- **Verification**: `pass-cli verify-audit` command to check log integrity
- **Graceful Degradation**: Operations continue even if audit logging fails

**Audit Log Location**:
- **Default**: Same directory as vault (e.g., `~/.pass-cli/audit.log`)
- **Custom**: Set `PASS_AUDIT_LOG` environment variable

**Enable Audit Logging**:
```bash
# New vault with audit logging
pass-cli init --enable-audit

# Verify audit log integrity
pass-cli verify-audit
```

**Audit Log Entry Example**:
```json
{
  "timestamp": "2025-01-13T10:30:45.123Z",
  "event_type": "credential_access",
  "outcome": "success",
  "credential_name": "github.com",
  "hmac_signature": "a1b2c3..."
}
```

## Threat Model

### What Pass-CLI Protects Against

[OK] **Offline Attacks**
- Vault file encryption protects against offline brute-force
- PBKDF2 slows down password cracking (600,000 iterations)
- No plaintext credentials stored anywhere

[OK] **File System Compromise**
- Encrypted vault remains secure even if file is stolen
- File permissions prevent unauthorized local access

[OK] **Process Memory Dumps**
- Sensitive data cleared from memory after use
- Master password not kept in memory permanently

[OK] **Accidental Disclosure**
- No cloud storage = no cloud breach risk
- No network calls = no network interception

[OK] **Unauthorized Local Access**
- System keychain protects master password
- File permissions restrict vault access

### What Pass-CLI Does NOT Protect Against

[ERROR] **Malware on Your Machine**
- Keyloggers can capture master password when entered
- Memory scrapers can extract decrypted credentials
- Root/admin access bypasses file permissions

[ERROR] **Physical Access Attacks**
- Attacker with physical access can copy vault file
- Vault encryption is only protection (strong password essential)

[ERROR] **Side-Channel Attacks**
- Timing attacks, power analysis not mitigated
- Not designed for hostile multi-user systems

[ERROR] **Weak Master Passwords**
- PBKDF2 slows attacks but doesn't prevent them
- Short/common passwords can be brute-forced

[ERROR] **Social Engineering**
- Cannot protect against phishing for master password
- User education essential

[ERROR] **TUI Display Security (Interactive Mode)**
- Shoulder surfing: Credentials visible on screen in TUI mode
- Screen recording: TUI displays service names and details
- Password visibility toggle: `Ctrl+P` shows plaintext passwords
- Shared terminals: Other users may see credential list

## Security Guarantees

### What We Guarantee

1. **Confidentiality**: Credentials encrypted with AES-256-GCM
2. **Integrity**: Authentication tag prevents tampering
3. **Forward Secrecy**: Unique nonce per encryption
4. **Secure Defaults**: No insecure configuration options

### What We Cannot Guarantee

1. **Availability**: Forgot password without recovery phrase = lost vault
2. **Zero-Knowledge**: Master password accessible via keychain
3. **Perfect Security**: Subject to implementation bugs

## Limitations

### Known Limitations

1. **Master Password Recovery**: Optional BIP39 recovery phrase
   - If recovery phrase was enabled during init, you can recover access with `pass-cli change-password --recover`
   - If recovery phrase was skipped (`--no-recovery`), vault is unrecoverable without the master password
   - If you lose both master password AND recovery phrase, vault is unrecoverable
   - No backdoor or master key exists

2. **Keychain Dependency**
   - Master password security depends on OS keychain
   - Compromise of OS account = compromise of master password

3. **Single-User Design**
   - Not designed for multi-user systems
   - File permissions rely on OS access controls

4. **No Network Security**
   - Offline-only design
   - No secure sharing mechanism

5. **Memory Security**
   - Go garbage collector may leave memory traces
   - Sensitive data cleared but not guaranteed wiped

### Out of Scope

- [FAIL] Cloud synchronization
- [FAIL] Multi-user support
- [FAIL] Hardware security module (HSM) integration
- [FAIL] Biometric authentication
- [FAIL] Two-factor authentication for master password

