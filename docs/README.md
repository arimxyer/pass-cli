---
title: "pass-cli Documentation"
---

![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?path=docs&label=Last%20Updated)

Welcome to the **pass-cli** documentation. pass-cli is a secure, cross-platform CLI password and API key manager designed for developers.

## Quick Links

- [Quick Start]({{< relref "01-getting-started/quick-start" >}}) - First-time setup and initialization (5 minutes)
- [Quick Install]({{< relref "01-getting-started/quick-install" >}}) - Installation instructions for all platforms
- [Command Reference]({{< relref "03-reference/command-reference" >}}) - Complete command reference
- [Backup & Restore Guide]({{< relref "02-guides/backup-restore" >}}) - Manual vault backup management
- [Security Architecture]({{< relref "03-reference/security-architecture" >}}) - Security features and cryptography
- [Troubleshooting]({{< relref "04-troubleshooting/_index" >}}) - Common issues and solutions by category

## Features

- **Strong Encryption**: AES-256-GCM with PBKDF2 key derivation (600,000 iterations)
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **Keychain Integration**: Optional OS keychain support for automatic unlocking
- **Interactive TUI**: Beautiful terminal UI built with tview
- **Clipboard Support**: Secure clipboard integration with auto-clear
- **Usage Tracking**: Per-credential usage statistics
- **Audit Logging**: HMAC-signed audit logs for all operations
- **Manual Backups**: Create and restore vault backups on demand

## Getting Help

- **GitHub Issues**: [Report bugs or request features](https://github.com/ari1110/pass-cli/issues)
- **GitHub Discussions**: [Ask questions and share ideas](https://github.com/ari1110/pass-cli/discussions)
- **Documentation**: You're reading it!

## Contributing

See [Contributing Guide]({{< relref "06-development/contributing" >}}) for developer documentation and contribution guidelines.
