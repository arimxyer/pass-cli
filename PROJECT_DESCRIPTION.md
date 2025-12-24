# Pass-CLI Project Description

## Project Overview

**Pass-CLI** is a secure, cross-platform command-line password manager designed specifically for developers who need fast, reliable, and script-friendly access to credentials without cloud dependencies. The project emphasizes local-first data storage, military-grade encryption, and seamless integration with developer workflows.

## Purpose & Problem Solved

**Problem**: Developers need secure credential management that integrates seamlessly with their command-line workflows, supports automation and scripting, and doesn't rely on cloud services or third-party infrastructure.

**Solution**: Pass-CLI provides a single-binary, offline-first password manager with:
- Local encrypted storage (no cloud sync required)
- Script-friendly output modes for CI/CD integration
- OS-native keychain integration for convenience
- Terminal User Interface (TUI) for visual management
- Cross-platform support (Windows, macOS, Linux)

## Core Features

### Security
- **AES-256-GCM Encryption**: Military-grade encryption with authenticated encryption mode
- **PBKDF2 Key Derivation**: 600,000 iterations for hardened password hashing
- **BIP39 Recovery Phrase**: 24-word mnemonic for vault password recovery (same standard as hardware wallets)
- **OS Keychain Integration**: Secure master password storage in Windows Credential Manager, macOS Keychain, or Linux Secret Service
- **Tamper-Evident Audit Logging**: HMAC-SHA256 signed audit trail for all vault operations
- **Password Policy Enforcement**: Complexity requirements for both vault and credential passwords

### Usability
- **Interactive TUI Mode**: Visual credential management with keyboard shortcuts and responsive layout
- **CLI Mode**: Traditional command-line interface for scripting and automation
- **Script-Friendly Outputs**: Clean output modes (`--quiet`, `--field`, `--masked`) for shell integration
- **Usage Tracking**: Automatic tracking of where credentials are used across projects
- **Health Checks**: Built-in `doctor` command for vault verification and troubleshooting

### Data Management
- **Atomic Vault Operations**: Crash-safe save operations with automatic rollback
- **Backup & Restore**: Manual and automatic backup creation with integrity verification
- **Vault Migration**: Tools for migrating between different encryption standards
- **Offline First**: No network calls, works completely offline

## Technology Stack

### Core Technologies
- **Language**: Go 1.25.1
- **CLI Framework**: Cobra (command structure and parsing)
- **Configuration**: Viper (YAML-based configuration management)
- **TUI Framework**: rivo/tview (Terminal User Interface)
- **Encryption**: Go crypto stdlib (AES-GCM, PBKDF2-SHA256)
- **Keychain**: zalando/go-keyring (OS credential manager integration)
- **Recovery**: tyler-smith/go-bip39 (BIP39 mnemonic phrase generation)

### Architecture
- **Library-First Design**: Business logic in `internal/` packages, CLI commands as thin wrappers
- **Layered Architecture**:
  - `cmd/`: Command-line interface and TUI entry points
  - `internal/vault/`: Vault operations and credential management
  - `internal/crypto/`: Encryption, decryption, key derivation
  - `internal/keychain/`: OS keychain integration
  - `internal/security/`: Audit logging and tamper detection
  - `internal/storage/`: File operations and atomic saves
  - `internal/config/`: Configuration management
  - `internal/health/`: Health checks and diagnostics

## Project Structure

```
pass-cli/
├── cmd/                      # CLI commands (Cobra-based)
│   ├── tui/                  # TUI components (rivo/tview)
│   │   ├── components/       # UI components (sidebar, detail, header)
│   │   ├── layout/           # Layout management
│   │   ├── models/           # Data models and state
│   │   ├── events/           # Event handling
│   │   └── styles/           # Visual styling
│   ├── root.go               # Root command and global flags
│   ├── init.go               # Vault initialization
│   ├── add.go                # Add credential
│   ├── get.go                # Retrieve credential
│   ├── update.go             # Update credential
│   ├── delete.go             # Delete credential
│   ├── list.go               # List credentials
│   ├── generate.go           # Generate password
│   ├── keychain.go           # Keychain commands
│   ├── vault.go              # Vault commands
│   ├── doctor.go             # System diagnostics
│   └── version.go            # Version information
├── internal/                 # Internal library packages
│   ├── vault/                # Vault operations
│   ├── crypto/               # Encryption/decryption
│   ├── keychain/             # OS keychain integration
│   ├── security/             # Audit logging
│   ├── storage/              # File operations
│   ├── config/               # Configuration handling
│   └── health/               # Health checks
├── test/                     # Integration and unit tests
├── docs/                     # Documentation
├── specs/                    # Feature specifications
└── main.go                   # Application entry point
```

## Key Workflows

### First-Time Setup
1. User runs `pass-cli init`
2. System generates 24-word BIP39 recovery phrase
3. User sets master password (with policy enforcement)
4. Vault created with AES-256-GCM encryption
5. Optional: Enable OS keychain integration for auto-unlock

### Adding Credentials
1. User runs `pass-cli add <name>`
2. System prompts for username and password
3. Credential encrypted and stored in vault
4. Optional: Add URL, notes, and usage location
5. Audit log entry created (if enabled)

### Retrieving Credentials
1. User runs `pass-cli get <name>`
2. System unlocks vault (prompts for password or uses keychain)
3. Credential decrypted and displayed
4. Usage tracking updated
5. Audit log entry created (if enabled)

### Script Integration
```bash
# Export credential to environment variable
export API_KEY=$(pass-cli get myservice --quiet --field password)

# Use in CI/CD pipelines
pass-cli get github --field password | docker login -u user --password-stdin
```

## Development Practices

### Code Quality
- **Testing**: Unit tests in `internal/`, integration tests in `test/`
- **Linting**: golangci-lint, gosec (security scanner), go vet
- **Coverage**: Target 60%+ coverage on critical packages
- **Security**: govulncheck for dependency vulnerabilities

### Contribution Workflow
1. Fork repository
2. Create feature branch from `main`
3. Implement changes following Go best practices
4. Add tests (unit and integration)
5. Run pre-commit checks (fmt, vet, lint, test, security scan)
6. Submit pull request with clear description

### Commit Standards
- Format: `<type>: <description>` (feat, fix, docs, refactor, test, chore)
- Include detailed body for non-trivial changes
- Reference issue/spec when applicable
- Footer: `Generated with Claude Code\n\nCo-Authored-By: Claude <noreply@anthropic.com>`

## Distribution

### Package Managers
- **Homebrew** (macOS/Linux): `brew tap arimxyer/homebrew-tap && brew install pass-cli`
- **Scoop** (Windows): `scoop bucket add arimxyer https://github.com/arimxyer/scoop-bucket && scoop install pass-cli`
- **Manual**: Download binary from GitHub Releases

### Supported Platforms
- macOS (Intel x64, Apple Silicon ARM64)
- Linux (amd64, arm64)
- Windows (x64)

## Documentation

### User Documentation
- **Getting Started**: Quick start guide, installation instructions
- **Command Reference**: Complete CLI command documentation
- **TUI Guide**: Terminal UI keyboard shortcuts and features
- **Security Architecture**: Encryption details, best practices
- **Recovery Guide**: BIP39 recovery phrase usage
- **Troubleshooting**: FAQ and common issues

### Developer Documentation
- **Architecture**: System design and component interaction
- **Contributing Guide**: Development workflow and standards
- **Testing Guide**: Unit and integration test patterns
- **CI/CD Integration**: GitHub Actions and pipeline examples
- **Branch Workflow**: Git workflow for contributors

Full documentation available at: https://arimxyer.github.io/pass-cli/

## Roadmap

### Completed Features
- ✅ AES-256-GCM encryption with PBKDF2
- ✅ BIP39 recovery phrase support
- ✅ OS keychain integration (Windows/macOS/Linux)
- ✅ Interactive TUI mode
- ✅ Atomic vault operations with crash safety
- ✅ Manual backup and restore
- ✅ Password policy enforcement
- ✅ Tamper-evident audit logging
- ✅ Health check diagnostics
- ✅ Usage tracking across projects

### Planned Features
- 🔄 TOTP/2FA support (store TOTP secrets, generate codes)
- 🔄 Advanced search and filtering
- 🔄 Credential groups/categories
- 🔄 Import from other password managers

## Community & Support

- **GitHub Repository**: https://github.com/arimxyer/pass-cli
- **Issue Tracker**: https://github.com/arimxyer/pass-cli/issues
- **Discussions**: https://github.com/arimxyer/pass-cli/discussions
- **License**: Apache License 2.0

## Target Audience

### Primary Users
- **DevOps Engineers**: CI/CD pipeline credential management
- **Software Developers**: API keys, database credentials, service passwords
- **System Administrators**: Server and infrastructure credentials
- **Security-Conscious Users**: Offline-first, local-only credential storage

### Use Cases
1. **Local Development**: Store database credentials, API keys for local dev
2. **CI/CD Pipelines**: Script-friendly credential retrieval in automated workflows
3. **Server Administration**: Manage SSH keys, sudo passwords, service credentials
4. **Multi-Project Workflows**: Track which credentials are used where
5. **Offline Security**: Work with sensitive credentials without internet connectivity

## Unique Value Propositions

1. **Developer-First Design**: Built by developers, for developers
2. **No Cloud Lock-In**: Completely offline, no third-party dependencies
3. **Script-Friendly**: Clean output modes designed for shell integration
4. **Single Binary**: No complex installation, just download and run
5. **Cross-Platform**: Same experience on Windows, macOS, and Linux
6. **Military-Grade Security**: Industry-standard encryption (AES-256-GCM, PBKDF2)
7. **Recovery Safety Net**: BIP39 recovery phrase prevents permanent lockout
8. **OS Integration**: Native keychain support for convenience without sacrificing security
9. **Tamper Detection**: Audit logging with HMAC signatures
10. **Open Source**: Transparent security, community-driven development

---

**Project Status**: Active Development
**Current Version**: Check [GitHub Releases](https://github.com/arimxyer/pass-cli/releases)
**Last Updated**: 2025-12-24
