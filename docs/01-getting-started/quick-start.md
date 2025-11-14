---
title: "Quick Start Guide"
weight: 3
toc: true
---

This 5-minute guide will walk you through initializing your vault and storing your first credential.

![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?path=docs&label=Last%20Updated)

## Installation

See [Quick Install]({{< relref "quick-install" >}}) for platform-specific installation instructions (Homebrew, Scoop) or [Manual Installation]({{< relref "manual-install" >}}) for binary download.

After installation, verify pass-cli is available:

```bash
pass-cli version
```

## First-Run Guided Initialization

The easiest way to get started is to let pass-cli guide you through the setup process automatically.

### How It Works

When you run any vault-requiring command (`add`, `get`, `list`, `update`, `delete`) for the first time without an existing vault, pass-cli automatically detects this and offers to guide you through initialization.

### Example: First-Run Experience

```bash
$ pass-cli list

╔════════════════════════════════════════════════════════════╗
║                  Welcome to pass-cli!                      ║
╚════════════════════════════════════════════════════════════╝

This appears to be your first time using pass-cli.
Would you like to create a new vault? (y/n): y

Great! Let's set up your secure password vault.

┌────────────────────────────────────────────────────────────┐
│ Step 1: Master Password                                    │
└────────────────────────────────────────────────────────────┘

Your master password encrypts all credentials in your vault.

Password Requirements:
  • Minimum 12 characters
  • At least one uppercase letter
  • At least one lowercase letter
  • At least one number
  • At least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

Enter master password: ••••••••••••••
Confirm master password: ••••••••••••••

✓ Password meets all requirements

┌────────────────────────────────────────────────────────────┐
│ Step 2: Keychain Integration                               │
└────────────────────────────────────────────────────────────┘

Store your master password in OS keychain for convenience?
(Windows Credential Manager, macOS Keychain, Linux Secret Service)

Benefits:
  ✓ No need to type password for every operation
  ✓ Secure OS-level storage
  ✓ Can be disabled later (see Keychain Setup guide)

Enable keychain storage? (y/n): y

✓ Master password stored in keychain

┌────────────────────────────────────────────────────────────┐
│ Step 3: Audit Logging                                      │
└────────────────────────────────────────────────────────────┘

Enable audit logging to track all vault operations?

Benefits:
  ✓ Security audit trail
  ✓ Tamper-evident with HMAC signatures
  ✓ Track all add/get/update/delete operations

Enable audit logging? (y/n): y

✓ Audit logging enabled

┌────────────────────────────────────────────────────────────┐
│ Setup Complete!                                            │
└────────────────────────────────────────────────────────────┘

Your vault has been created at:
  /home/user/.pass-cli/vault.enc

Next steps:
  • Add your first credential: pass-cli add github
  • List credentials: pass-cli list
  • Check vault health: pass-cli doctor

No credentials found.
```

### When Guided Initialization Is NOT Triggered

Guided initialization is **skipped** in these scenarios:

1. **Vault already exists**: If you have an existing vault at the default or configured location, no prompt appears
2. **Custom vault configured**: If you've configured `vault_path` in your config file, initialization uses that location automatically
3. **Non-TTY environment**: If running in a script or pipe (stdin is not a terminal), initialization is skipped to avoid blocking
4. **Commands that don't require vault**: Commands like `version`, `doctor`, `help` don't trigger initialization

### Non-TTY Behavior

If you run pass-cli in a script or CI/CD environment without a vault:

```bash
$ echo "list" | pass-cli list
Error: vault file not found: /home/user/.pass-cli/vault.enc

Run 'pass-cli init' to create a new vault.
```

This prevents scripts from hanging while waiting for interactive input.

## Manual Initialization

If you prefer to initialize your vault explicitly, or if you're in a non-interactive environment, use the `init` command:

```bash
pass-cli init
```

This provides the same setup process as guided initialization, but you invoke it explicitly.

### Advanced Options

#### Custom Vault Location

To use a custom vault location, configure it in your config file before initialization:

```bash
# Edit config file
echo "vault_path: /custom/path/vault.enc" > ~/.pass-cli/config.yml

# Then initialize
pass-cli init
```

Future commands will automatically use the configured vault location:

```bash
pass-cli add github
pass-cli get github
```

#### Skip Keychain Integration

To initialize a vault without storing the master password in the OS keychain, simply don't use the `--use-keychain` flag:

```bash
pass-cli init
```

During the interactive setup, answer "n" when asked about keychain storage. You'll need to enter your password for each operation.

#### Skip Audit Logging

Audit logging is disabled by default. To enable it during initialization, use:

```bash
pass-cli init --enable-audit
```

If you omit this flag, your vault will be created without audit logging.

## Your First Credential

After initialization (automatic or manual), add your first credential:

```bash
$ pass-cli add github

Enter username: your-github-username
Enter password: ••••••••••••
Confirm password: ••••••••••••

✓ Credential 'github' added successfully
```
