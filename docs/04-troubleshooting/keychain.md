---
title: "Keychain Issues"
weight: 3
bookToc: true
---

# Keychain Troubleshooting

Solutions for OS keychain integration and master password storage issues.

## Keychain Access Issues

### "Failed to Retrieve Master Password" Error

**Symptom**: Cannot get master password from keychain

**Cause**: Master password not stored or keychain locked

**Solutions**:

1. **Verify keychain entry exists**

   **macOS:**
   ```bash
   # Check Keychain Access app
   # Search for "pass-cli" entry
   ```

   **Linux:**
   ```bash
   # Check with Seahorse (GNOME) or KWalletManager (KDE)
   ```

   **Windows:**
   ```powershell
   # Check Credential Manager
   # Control Panel → User Accounts → Credential Manager
   # Windows Credentials → Look for "pass-cli"
   ```

2. **Reinitialize vault** (will prompt for password again)
   ```bash
   # Backup vault first
   cp ~/.pass-cli/vault.enc ~/vault-backup.enc

   # Reinitialize (keeps vault but updates keychain)
   pass-cli init
   ```

---

### Keychain Access Denied

**Symptom**: "Access denied" when accessing keychain

**Cause**: Keychain locked or permission issues

**Solutions**:

**macOS:**
```bash
# Unlock keychain
security unlock-keychain ~/Library/Keychains/login.keychain-db

# Grant access to pass-cli
# Will prompt when pass-cli runs - click "Always Allow"
```

**Linux (GNOME):**
```bash
# Unlock keyring
# Will prompt for keyring password when pass-cli runs

# If keyring password is different from login password
# Open Seahorse → Right-click Login → Change Password
```

**Windows:**
```powershell
# Ensure running as correct user
whoami

# Credential Manager uses Windows login credentials
# Ensure logged in as user who created vault
```

---

### "Secret Service Not Available" (Linux)

**Symptom**: Cannot access secret service on Linux

**Cause**: Secret service daemon not running

**Solutions**:

**GNOME:**
```bash
# Install GNOME Keyring
sudo apt install gnome-keyring  # Ubuntu/Debian
sudo dnf install gnome-keyring  # Fedora

# Start daemon
gnome-keyring-daemon --start --components=secrets

# Add to session startup
# Add to ~/.profile or ~/.bash_profile:
eval $(gnome-keyring-daemon --start --components=secrets)
```

**KDE:**
```bash
# Install KWallet
sudo apt install kwalletmanager  # Ubuntu/Debian

# Start KWallet
kwalletd5 &
```

**Alternative: File-based password** (less secure)
```bash
# Store password in encrypted file (not recommended)
# Use vault without keychain integration
# Enter password each time
```

---

