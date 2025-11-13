```
                              ╔═══╗
                              ║   ║
                            ╔═╩═══╩═╗
                            ║ ┌───┐ ║
                            ║ │ ● │ ║
                            ╚═╧═══╧═╝

 ██████╗   █████╗  ███████╗ ███████╗       █████╗ ██╗     ██╗
 ██╔══██╗ ██╔══██╗ ██╔════╝ ██╔════╝      ██╔═══╝ ██║     ██║
 ██████╔╝ ███████║ ███████╗ ███████╗█████╗██║     ██║     ██║
 ██╔═══╝  ██╔══██║ ╚════██║ ╚════██║╚════╝██║     ██║     ██║
 ██║      ██║  ██║ ███████║ ███████║      ╚██████╗███████╗██║
 ╚═╝      ╚═╝  ╚═╝ ╚══════╝ ╚══════╝       ╚═════╝╚══════╝╚═╝
```

> A secure, cross-platform command-line password manager designed for developers

![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?label=Last%20Updated)

Pass-CLI is a fast, secure password and API key manager that stores credentials locally with AES-256-GCM encryption. Built for developers who need quick, script-friendly access to credentials without cloud dependencies.

📚 **[View Full Documentation](https://ari1110.github.io/pass-cli/)** | [Getting Started](docs/01-getting-started/quick-start.md) | [Installation](docs/01-getting-started/quick-install.md) | [Usage Guide](docs/03-reference/command-reference.md)

## Key Features

- **Military-Grade Encryption**: AES-256-GCM with hardened PBKDF2 key derivation (600,000 iterations)
- **System Keychain Integration**: Windows Credential Manager, macOS Keychain, Linux Secret Service
- **Password Policy Enforcement**: Complexity requirements for vault and credential passwords
- **Vault Backup & Restore**: Manual backups with integrity verification and automatic backup during vault operations
- **Tamper-Evident Audit Logging**: Optional HMAC-signed audit trail for vault operations
- **Health Checks**: Built-in `doctor` command for vault verification and troubleshooting
- **Cross-Platform**: Single binary for Windows, macOS (Intel/ARM), and Linux (amd64/arm64)
- **Script-Friendly**: Clean output modes (`--quiet`, `--field`, `--masked`) for shell integration
- **Usage Tracking**: Automatic tracking of where credentials are used across projects
- **Offline First**: No cloud dependencies, works completely offline
- **Interactive TUI**: Terminal UI for visual credential management

## Quick Start

### Installation

**macOS / Linux (Homebrew)**:
```bash
brew tap ari1110/homebrew-tap
brew install pass-cli
```

**Windows (Scoop)**:
```powershell
scoop bucket add pass-cli https://github.com/ari1110/scoop-bucket
scoop install pass-cli
```

For manual installation and other methods, see [docs/01-getting-started/manual-install.md](docs/01-getting-started/manual-install.md).

### Getting Started

```bash
# Initialize vault (guided setup on first use)
pass-cli init

# Add your first credential
pass-cli add github
# Enter username and password when prompted

# Retrieve a credential
pass-cli get github

# List all credentials
pass-cli list

# Use in scripts (quiet mode)
export API_KEY=$(pass-cli get myservice --quiet --field password)
```

For detailed usage and examples, see [docs/01-getting-started/quick-start.md](docs/01-getting-started/quick-start.md).

## Interactive TUI Mode

Pass-CLI includes a Terminal User Interface for visual credential management:

```bash
# Launch TUI mode (no arguments)
pass-cli

# CLI commands work with explicit subcommands
pass-cli list
```

**Key Features**:
- Visual navigation with arrow keys and Tab
- Interactive forms for adding/editing credentials
- Password visibility toggle with `Ctrl+P`
- Search and filter with `/`
- Customizable keyboard shortcuts
- Responsive layout (requires 60x30 minimum terminal size)

Press `?` in TUI mode to see all keyboard shortcuts. For complete TUI documentation and configuration, see [docs/03-reference/command-reference.md](docs/03-reference/command-reference.md).

## Core Commands

```bash
# Initialize vault
pass-cli init

# Add credential
pass-cli add github --url https://github.com --notes "Personal account"

# Get credential (formatted display)
pass-cli get github

# Get credential (script-friendly)
pass-cli get github --quiet --field password

# List all credentials
pass-cli list

# Update credential
pass-cli update github --username newuser@example.com

# Delete credential
pass-cli delete github

# Generate password
pass-cli generate --length 32

# Remove vault
pass-cli vault remove

# Create manual vault backup
pass-cli vault backup create

# Restore vault from backup
pass-cli vault backup restore

# View backup status
pass-cli vault backup info

# Health check
pass-cli doctor
```

For complete command reference, flags, and examples, see [docs/03-reference/command-reference.md](docs/03-reference/command-reference.md).

## Security

**Encryption**:
- AES-256-GCM with PBKDF2-SHA256 key derivation (600,000 iterations)
- Unique salt per vault, unique IV per credential
- Built-in authentication tag prevents tampering

**Password Policy**:
- Minimum 12 characters with uppercase, lowercase, digit, and special symbol requirements
- Enforced for both vault and credential passwords

**Keychain Integration**:
- Master password stored in OS keychain (Windows Credential Manager, macOS Keychain, Linux Secret Service)
- Automatic unlock when needed
- Enable/disable anytime: `pass-cli keychain enable` / `pass-cli keychain disable`
- Check status: `pass-cli keychain status`
- TUI auto-unlocks when keychain is enabled

**Audit Logging** (Optional):
- Tamper-evident HMAC-SHA256 signed audit trail
- Enable with `pass-cli init --enable-audit`

**Vault Location**:
- Windows: `%USERPROFILE%\.pass-cli\vault.enc`
- macOS/Linux: `~/.pass-cli/vault.enc`

For complete security details, best practices, and migration guides, see [docs/03-reference/security-architecture.md](docs/03-reference/security-architecture.md).

## Documentation

**Essential Guides**:
- [Getting Started](docs/01-getting-started/quick-start.md) - First-time setup and basic workflows
- [Usage Guide](docs/03-reference/command-reference.md) - Complete command reference, TUI shortcuts, configuration
- [Installation](docs/01-getting-started/quick-install.md) - All installation methods and package managers
- [Security](docs/03-reference/security-architecture.md) - Encryption details, best practices, migration guides
- [Troubleshooting](docs/04-reference/troubleshooting.md) - Common issues and solutions

**Additional Resources**:
- [Doctor Command](docs/05-development/doctor-command.md) - Health check diagnostics
- [CI/CD Integration](docs/05-development/ci-cd.md) - GitHub Actions and pipeline examples
- [Branch Workflow](docs/05-development/branch-workflow.md) - Git workflow for contributors

## Building from Source

**Prerequisites**: Go 1.25 or later

```bash
# Clone and build
git clone https://github.com/ari1110/pass-cli.git
cd pass-cli
go build -o pass-cli .

# Run tests
go test ./...
```

For testing guidelines, see [test/README.md](test/README.md). For Git workflow, see [docs/05-development/branch-workflow.md](docs/05-development/branch-workflow.md).

## FAQ

### How is this different from `pass` (the standard Unix password manager)?

Pass-CLI offers system keychain integration (no GPG required), built-in clipboard support, usage tracking, cross-platform Windows support, script-friendly output modes (`--quiet`, `--field`, `--masked`), and single binary distribution.

### Is my data stored in the cloud?

No. Pass-CLI stores everything locally on your machine. There are no cloud dependencies or network calls.

### How do I backup my vault?

The vault is a single file. Simply copy it:
```bash
cp ~/.pass-cli/vault.enc ~/backup/vault-$(date +%Y%m%d).enc
```

### What happens if I forget my master password?

Unfortunately, there's no way to recover your vault without the master password. The encryption is designed to be unbreakable. Keep your master password safe.

For more questions and troubleshooting, see [docs/04-reference/troubleshooting.md](docs/04-reference/troubleshooting.md).

## Contributing

Contributions are welcome! See [docs/05-development/branch-workflow.md](docs/05-development/branch-workflow.md) for Git workflow and contribution guidelines.

## License

This project is licensed under the MIT License.

## Links

- **Releases**: [GitHub Releases](https://github.com/ari1110/pass-cli/releases)
- **Issues**: [GitHub Issues](https://github.com/ari1110/pass-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ari1110/pass-cli/discussions)

---

Made with ❤️ by developers, for developers.
