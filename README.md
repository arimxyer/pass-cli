# Pass-CLI

> A secure, cross-platform command-line password manager designed for developers

**Version**: v0.0.1 | **Status**: Production Ready | **Last Updated**: October 2025

Pass-CLI is a fast, secure password and API key manager that stores credentials locally with AES-256-GCM encryption. Built for developers who need quick, script-friendly access to credentials without cloud dependencies.

## ✨ Key Features

- **🔒 Military-Grade Encryption**: AES-256-GCM with hardened PBKDF2 key derivation (600,000 iterations)
- **🔐 System Keychain Integration**: Seamless integration with Windows Credential Manager, macOS Keychain, and Linux Secret Service
- **🛡️ Password Policy Enforcement**: Complexity requirements for vault and credential passwords
- **📝 Tamper-Evident Audit Logging**: Optional HMAC-signed audit trail for vault operations
- **⚡ Lightning Fast**: Sub-100ms credential retrieval, ~50-100ms vault unlock on modern CPUs
- **🖥️ Cross-Platform**: Single binary for Windows, macOS (Intel/ARM), and Linux (amd64/arm64)
- **📋 Clipboard Support**: Automatic credential copying with security timeouts
- **🔑 Password Generation**: Cryptographically secure random passwords
- **📊 Usage Tracking**: Automatic tracking of where credentials are used
- **🤖 Script-Friendly**: Clean output modes for shell integration (`--quiet`, `--field`, `--masked`, `--no-clipboard`)
- **🔌 Offline First**: No cloud dependencies, works completely offline
- **📦 Easy Installation**: Available via Homebrew and Scoop

## 🚀 Quick Start

### Installation

#### macOS / Linux (Homebrew)

```bash
# Add tap and install
brew tap ari1110/homebrew-tap
brew install pass-cli
```

#### Windows (Scoop)

```powershell
# Add bucket and install
scoop bucket add pass-cli https://github.com/ari1110/scoop-bucket
scoop install pass-cli
```

#### Manual Installation

Download the latest release for your platform from [GitHub Releases](https://github.com/ari1110/pass-cli/releases):

```bash
# Extract and move to PATH
tar -xzf pass-cli_*_<os>_<arch>.tar.gz
sudo mv pass-cli /usr/local/bin/  # macOS/Linux
# Or move pass-cli.exe to a directory in PATH (Windows)
```

### First Steps

```bash
# Initialize your vault
pass-cli init

# Add your first credential
pass-cli add github
# Enter username and password when prompted

# Retrieve it
pass-cli get github

# Use in scripts (quiet mode)
export API_KEY=$(pass-cli get myservice --quiet --field password)
```

## 🎨 Interactive TUI Mode

Pass-CLI includes an interactive Terminal User Interface (TUI) for visual credential management.

### Launching TUI Mode

```bash
# Launch TUI (no arguments)
pass-cli

# CLI commands still work with explicit subcommands
pass-cli list
pass-cli get github
```

### TUI Features

- **Visual Navigation**: Browse credentials with arrow keys and Tab
- **Interactive Forms**: Add/edit credentials with visual feedback
- **Password Visibility Toggle**: Press `Ctrl+H` in forms to verify passwords
- **Search & Filter**: Press `/` to search, `Esc` to clear
- **Keyboard Shortcuts**: Press `?` to see all available shortcuts
- **Responsive Layout**: Sidebar and detail panel adapt to terminal size
- **Minimum Terminal Size**: Requires 60 columns × 30 rows (warning overlay displays if terminal is too small)

### Key TUI Shortcuts

#### TUI Keyboard Shortcuts

**Configurable Shortcuts** (can be customized via config.yml):

| Shortcut | Action | Context |
|----------|--------|---------|
| `q` | Quit application | Any time |
| `a` | New credential | Main view |
| `e` | Edit credential | Main view |
| `d` | Delete credential | Main view |
| `i` | Toggle detail panel | Main view |
| `s` | Toggle sidebar | Main view |
| `?` | Show help modal | Any time |
| `/` | Search/filter | Main view |

**Hardcoded Shortcuts** (navigation and forms):

| Shortcut | Action | Context |
|----------|--------|---------|
| `Tab` | Next component | All views |
| `Shift+Tab` | Previous component | All views |
| `↑/↓` | Navigate lists | List views |
| `Enter` | Select / View details | List views |
| `Esc` | Close modal / Exit search | Modals, search |
| `Ctrl+C` | Force quit application | Any time |
| `c` | Copy password to clipboard | Detail view |
| `p` | Toggle password visibility | Detail view |

**Total: 16 keyboard shortcuts** (8 configurable + 8 hardcoded)

**Customization**: Some shortcuts can be customized via `~/.pass-cli/config.yaml` (see Configuration section below)

See [full keyboard shortcuts reference](docs/USAGE.md#tui-keyboard-shortcuts) for detailed context and examples.

## 📖 Usage

### Initialize Vault

```bash
# Create a new vault
pass-cli init

# The vault is stored at ~/.pass-cli/
```

### Add Credentials

```bash
# Interactive mode (prompts for username/password)
pass-cli add myservice

# With URL and notes
pass-cli add github --url https://github.com --notes "Personal account"

# Generate a strong password separately, then add credential
pass-cli generate
# Copy generated password, then:
pass-cli add newservice
# (Paste password when prompted)
```

### Retrieve Credentials

```bash
# Display credential (formatted)
pass-cli get myservice

# Copy password to clipboard
pass-cli get myservice --copy

# Quiet mode for scripts (password only)
pass-cli get myservice --quiet

# Get specific field
pass-cli get myservice --field username --quiet

# Display with masked password
pass-cli get myservice --masked
```

### List Credentials

```bash
# List all credentials
pass-cli list

# Show unused credentials
pass-cli list --unused
```

### Update Credentials

```bash
# Update password (prompted)
pass-cli update myservice

# Update specific fields
pass-cli update myservice --username newuser@example.com
pass-cli update myservice --url https://new-url.com
pass-cli update myservice --notes "Updated notes"

```

### Delete Credentials

```bash
# Delete a credential (with confirmation)
pass-cli delete myservice

# Force delete (no confirmation)
pass-cli delete myservice --force
```

### Generate Passwords

```bash
# Generate a password (default: 20 chars, alphanumeric + symbols)
pass-cli generate

# Custom length
pass-cli generate --length 32

# Alphanumeric only (no symbols)
pass-cli generate --no-symbols
```

### Version Information

```bash
# Check version
pass-cli version

# Verbose version info
pass-cli version --verbose
```

## 🔐 Security

### Encryption

- **Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Key Derivation**: PBKDF2-SHA256 with 600,000 iterations (hardened January 2025)
- **Salt**: Unique 32-byte random salt per vault
- **Authentication**: Built-in authentication tag (GCM) prevents tampering
- **IV**: Unique initialization vector per credential
- **Performance**: ~50-100ms on modern CPUs, 500-1000ms on older hardware

### Password Policy (January 2025)

All passwords (vault and credentials) must meet:
- Minimum 12 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special symbol (!@#$%^&*()-_=+[]{}|;:,.<>?)

### Audit Logging (Optional)

Enable tamper-evident audit logging:
```bash
# Initialize vault with audit logging
pass-cli init --enable-audit

# Verify audit log integrity
pass-cli verify-audit
```

Audit logs use HMAC-SHA256 signatures and store keys in OS keychain.

### Master Password Storage

Pass-CLI integrates with your operating system's secure credential storage:

- **Windows**: Windows Credential Manager
- **macOS**: Keychain
- **Linux**: Secret Service (GNOME Keyring, KWallet)

Your master password is stored securely and unlocked automatically when needed.

### Clipboard Security

When using `--copy`, the clipboard is:
1. Cleared after 30 seconds automatically
2. Only contains the password (no metadata)
3. Can be cleared immediately with Ctrl+C

### Vault Location

The encrypted vault is stored at:
- **Windows**: `%USERPROFILE%\.pass-cli\vault.enc`
- **macOS/Linux**: `~/.pass-cli/vault.enc`

### Best Practices

- ✅ Use a strong, unique master password (20+ characters, meets complexity requirements)
- ✅ Keep your vault backed up (it's just a file!)
- ✅ Use `--generate` for new passwords (automatic policy compliance)
- ✅ Regularly update credentials
- ✅ Use `--quiet` mode in scripts to avoid logging sensitive data
- ✅ Enable audit logging for compliance/security monitoring (`--enable-audit`)
- ✅ Migrate old vaults to 600k iterations (see `docs/MIGRATION.md`)
- ❌ Don't commit vault files to version control
- ❌ Don't share your master password

## 🤖 Script Integration

### Shell Integration Examples

**Bash/Zsh**:

```bash
#!/bin/bash
# Export API key for use in script
export API_KEY=$(pass-cli get openai --quiet --field password)

# Use in curl
curl -H "Authorization: Bearer $(pass-cli get github --quiet)" \
     https://api.github.com/user

# Conditional on success
if pass-cli get myservice --quiet > /dev/null 2>&1; then
    echo "Credential exists"
fi
```

**PowerShell**:

```powershell
# Store credential in variable
$apiKey = pass-cli get myservice --quiet --field password

# Use in web request
$headers = @{
    "Authorization" = "Bearer $apiKey"
}
Invoke-RestMethod -Uri "https://api.example.com" -Headers $headers

# Use with environment variable
$env:DATABASE_PASSWORD = pass-cli get postgres --quiet
```

**CI/CD Integration**:

```yaml
# GitHub Actions example
steps:
  - name: Retrieve credentials
    run: |
      export DB_PASSWORD=$(pass-cli get database --quiet)
      ./deploy.sh
```

### Output Modes

| Flag | Output | Use Case |
|------|--------|----------|
| (default) | Formatted table | Human-readable display |
| `--quiet` | Password only | Scripts, export to variables |
| `--field <name>` | Specific field | Extract username, URL, etc. |
| `--masked` | Masked password | Display password as asterisks |

## 📊 Usage Tracking

Pass-CLI automatically tracks where and when credentials are accessed. View this data through three powerful commands:

### View Detailed Credential Usage

See all locations where a credential has been accessed:

```bash
# View usage history for a credential
pass-cli usage github
```

**Output**:
```
Location                              Repository      Last Used        Count   Fields
────────────────────────────────────────────────────────────────────────────────────
/home/user/projects/web-app          my-web-app      2 hours ago      7       password:5, username:2
/home/user/projects/api               my-api          5 days ago       3       password:3
```

**JSON output for scripting**:
```bash
# Export usage data
pass-cli usage github --format json

# Count locations
pass-cli usage github --format json | jq '.usage_locations | length'
```

### Group Credentials by Project

Organize credentials by git repository context:

```bash
# Group all credentials by repository
pass-cli list --by-project
```

**Output**:
```
my-web-app (3 credentials):
  github
  aws-dev
  postgres

my-api (2 credentials):
  heroku
  redis

Ungrouped (1 credential):
  local-db
```

### Filter Credentials by Location

Find credentials used in a specific directory:

```bash
# Show credentials from current directory
pass-cli list --location .

# Show credentials from specific path
pass-cli list --location /home/user/projects/web-app

# Include subdirectories
pass-cli list --location /home/user/projects --recursive
```

### Combined Workflows

```bash
# Combine location filter with project grouping
pass-cli list --location ~/work --by-project --recursive

# Find unused credentials
pass-cli list --unused --days 30

# Script-friendly output
pass-cli list --location . --format simple | wc -l
```

**What gets tracked**:
- Location (absolute path where credential accessed)
- Git repository name (if accessed from git repo)
- Access timestamps and counts
- Field-level usage (which fields accessed)

**Use cases**:
- Audit credential usage before rotation
- Discover project-specific credentials
- Track credential access patterns
- Find unused credentials for cleanup
- Understand credential organization across projects

## 🛠️ Advanced Usage

### Custom Vault Location

```bash
# Use a custom vault location
pass-cli --vault /path/to/custom/vault.enc list

# Or set via environment variable
export PASS_CLI_VAULT=/path/to/custom/vault.enc
pass-cli list
```

### Verbose Logging

```bash
# Enable verbose output for debugging
pass-cli --verbose get myservice
```

### Configuration File

Pass-CLI supports user configuration via `config.yaml` (added in January 2025). This allows you to customize keybindings, terminal thresholds, and default settings.

**Configuration Location**:
- **Linux/macOS**: `~/.config/pass-cli/config.yml`
- **Windows**: `%APPDATA%\pass-cli\config.yml`

**Create and Edit Configuration**:

```bash
# Initialize default config
pass-cli config init

# Edit config in default editor
pass-cli config edit

# Validate config syntax
pass-cli config validate

# Reset to defaults
pass-cli config reset
```

**Example Configuration**:

```yaml
# ~/.config/pass-cli/config.yml

# Terminal display thresholds
terminal:
  min_width: 60   # Minimum columns (default: 60)
  min_height: 30  # Minimum rows (default: 30)

# Custom keyboard shortcuts (TUI mode)
keybindings:
  quit: "q"                  # Quit application
  add_credential: "n"        # Add new credential
  edit_credential: "e"       # Edit credential
  delete_credential: "d"     # Delete credential
  help: "?"                  # Show help modal
  search: "/"                # Activate search

# Supported key formats:
# - Single letters: a-z
# - Numbers: 0-9
# - Function keys: f1-f12
# - Modifiers: ctrl+, alt+, shift+
# Examples: ctrl+q, alt+a, shift+f1

# Validation: Config is validated on load
# - Duplicate keys rejected (conflict detection)
# - Unknown actions rejected
# - Invalid config shows warning modal, app continues with defaults
```

**Keybinding Customization**:
- Some TUI shortcuts can be customized via config.yaml
- Navigation shortcuts (Tab, arrows, Enter, Esc) are hardcoded and cannot be changed
- UI hints automatically update to reflect your custom keybindings in status bar and help modal

For complete configuration reference, see [docs/USAGE.md#configuration](docs/USAGE.md#configuration).

## 🏗️ Building from Source

### Prerequisites

- Go 1.25 or later

### Build

```bash
# Clone the repository
git clone https://github.com/ari1110/pass-cli.git
cd pass-cli

# Build binary
go build -o pass-cli .

# Run tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 Development

### Running Tests

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...

# Integration tests
go test -tags=integration ./test/

# All tests (unit + integration)
go test ./...
go test -v -tags=integration -timeout 5m ./test
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Security scan
gosec ./...

# Format code
go fmt ./...
```

## 🤝 Contributing

Contributions are welcome!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [docs/development/](docs/development/) for development guidelines and setup.

## 📋 Requirements

- **Operating System**: Windows 10+, macOS 10.15+, or Linux (any modern distribution)
- **Architecture**: amd64 or arm64
- **Dependencies**: None (static binary)

## 🗺️ Roadmap

- [x] Core credential management (add, get, list, update, delete)
- [x] AES-256-GCM encryption
- [x] System keychain integration
- [x] Hardened crypto (600k PBKDF2 iterations)
- [x] Password policy enforcement
- [x] Tamper-evident audit logging
- [x] Password generation
- [x] Clipboard support
- [x] Usage tracking
- [x] Atomic vault operations with rollback
- [ ] Import from other password managers
- [ ] Export functionality
- [ ] Credential sharing (encrypted)
- [ ] Two-factor authentication support
- [ ] Browser extensions

## 📄 License

This project is licensed under the MIT License.

## 🔗 Links

- **Documentation**: [docs/](docs/)
- **Releases**: [GitHub Releases](https://github.com/ari1110/pass-cli/releases)
- **Issues**: [GitHub Issues](https://github.com/ari1110/pass-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ari1110/pass-cli/discussions)

## ❓ FAQ

### How is this different from `pass` (the standard Unix password manager)?

Pass-CLI offers:
- System keychain integration (no GPG required)
- Built-in clipboard support
- Usage tracking
- Cross-platform Windows support
- Script-friendly output modes (--quiet, --field, --masked)
- Single binary distribution

### Is my data stored in the cloud?

No. Pass-CLI stores everything locally on your machine. There are no cloud dependencies or network calls.

### Can I sync my vault across machines?

The vault is a single encrypted file (`~/.pass-cli/vault.enc`). You can sync it using any file sync service (Dropbox, Google Drive, etc.), but be aware of potential conflicts if editing from multiple machines simultaneously.

### What happens if I forget my master password?

Unfortunately, there's no way to recover your vault without the master password. The encryption is designed to be unbreakable. Keep your master password safe and consider backing it up securely.

### How do I backup my vault?

Simply copy the vault file:
```bash
cp ~/.pass-cli/vault.enc ~/backup/vault-$(date +%Y%m%d).enc
```

### Can I use Pass-CLI in my company?

Yes! Pass-CLI is MIT licensed and free for commercial use. It's designed for professional developer workflows.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [go-keyring](https://github.com/zalando/go-keyring) for system keychain integration
- Inspired by the Unix `pass` password manager

## 📞 Support

- **Bug Reports**: [GitHub Issues](https://github.com/ari1110/pass-cli/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/ari1110/pass-cli/discussions)
- **Security Issues**: Email security@example.com (please don't file public issues)

---

Made with ❤️ by developers, for developers. Star ⭐ this repo if you find it useful!
