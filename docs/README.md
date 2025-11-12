---
title: "pass-cli Documentation"
---

# pass-cli Documentation

![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?path=docs&label=Last%20Updated)

Welcome to the **pass-cli** documentation. pass-cli is a secure, cross-platform CLI password and API key manager designed for developers.

## Quick Links

- [Getting Started]({{< relref "01-getting-started/first-steps" >}}) - First-time setup and initialization
- [Installation]({{< relref "01-getting-started/installation" >}}) - Installation instructions for all platforms
- [Usage Guide]({{< relref "02-usage/cli-reference" >}}) - Complete command reference
- [Backup & Restore Guide]({{< relref "03-guides/backup-restore" >}}) - Manual vault backup management
- [Security]({{< relref "04-reference/security" >}}) - Security features and best practices
- [Troubleshooting]({{< relref "04-reference/troubleshooting" >}}) - Common issues and solutions

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

See [Contributing Guide]({{< relref "05-development/contributing" >}}) for developer documentation and contribution guidelines.
