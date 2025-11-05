# Installation Guide


![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?path=docs&label=Last%20Updated) ![Status](https://img.shields.io/badge/Status-Production%20Ready-brightgreen)
Complete installation instructions for Pass-CLI across all supported platforms.

## Table of Contents

- [Quick Install](#quick-install)
- [Package Managers](#package-managers)
  - [Homebrew (macOS/Linux)](#homebrew-macoslinux)
  - [Scoop (Windows)](#scoop-windows)
- [Manual Installation](#manual-installation)
  - [Download Pre-built Binaries](#download-pre-built-binaries)
  - [Verify Checksums](#verify-checksums)
  - [Install Binary](#install-binary)
- [Building from Source](#building-from-source)
  - [Prerequisites](#prerequisites)
  - [Build Steps](#build-steps)
  - [Build Options](#build-options)
- [Post-Installation](#post-installation)
- [Troubleshooting](#troubleshooting)
- [Uninstallation](#uninstallation)

## Quick Install

### macOS / Linux

```bash
# Using Homebrew
brew tap ari1110/homebrew-tap
brew install pass-cli
```

### Windows

```powershell
# Using Scoop
scoop bucket add pass-cli https://github.com/ari1110/scoop-bucket
scoop install pass-cli
```

## Package Managers

Package managers provide the easiest installation method with automatic updates.

### Homebrew (macOS/Linux)

Homebrew is the recommended installation method for macOS and Linux.

#### Prerequisites

- macOS 10.15+ or Linux (any modern distribution)
- Homebrew installed ([installation instructions](https://brew.sh/))

#### Installation Steps

```bash
# Add the Pass-CLI tap
brew tap ari1110/homebrew-tap

# Install Pass-CLI
brew install pass-cli

# Verify installation
pass-cli version
```

#### Update

```bash
# Update Homebrew
brew update

# Upgrade Pass-CLI
brew upgrade pass-cli
```

#### Install Specific Version

```bash
# List available versions
brew info pass-cli

# Install specific version (if available)
brew install pass-cli@0.0.1
```

### Scoop (Windows)

Scoop is the recommended installation method for Windows.

#### Prerequisites

- Windows 10+ or Windows Server 2019+
- PowerShell 5.1+ or PowerShell Core 7+
- Scoop installed ([installation instructions](https://scoop.sh/))

#### Installation Steps

```powershell
# Add the Pass-CLI bucket
scoop bucket add pass-cli https://github.com/ari1110/scoop-bucket

# Install Pass-CLI
scoop install pass-cli

# Verify installation
pass-cli version
```

#### Update

```powershell
# Update Scoop
scoop update

# Upgrade Pass-CLI
scoop update pass-cli
```

#### Install Specific Version

```powershell
# List available versions
scoop info pass-cli

# Install specific version
scoop install pass-cli@0.0.1
```

## Manual Installation

Manual installation gives you direct control over the binary location and version.

### Download Pre-built Binaries

1. **Visit the Releases Page**

   Go to [GitHub Releases](https://github.com/ari1110/pass-cli/releases/latest)

2. **Choose Your Platform**

   Download the appropriate archive for your system:

   | Platform | Architecture | File |
   |----------|-------------|------|
   | macOS | Intel (x86_64) | `pass-cli_VERSION_darwin_amd64.tar.gz` |
   | macOS | Apple Silicon (ARM64) | `pass-cli_VERSION_darwin_arm64.tar.gz` |
   | Linux | x86_64 | `pass-cli_VERSION_linux_amd64.tar.gz` |
   | Linux | ARM64 | `pass-cli_VERSION_linux_arm64.tar.gz` |
   | Windows | x86_64 | `pass-cli_VERSION_windows_amd64.zip` |
   | Windows | ARM64 | `pass-cli_VERSION_windows_arm64.zip` |

3. **Download Checksums**

   Download `checksums.txt` from the same release page for verification.

### Verify Checksums

Verifying checksums ensures the downloaded file hasn't been tampered with.

#### macOS / Linux

```bash
# Download your platform's archive and checksums.txt
# Go to GitHub Releases and download your platform's archive
# Example for Linux amd64:
# 1. Visit: https://github.com/ari1110/pass-cli/releases/latest
# 2. Download: pass-cli_VERSION_linux_amd64.tar.gz
# 3. Download: checksums.txt

# Verify checksum
sha256sum -c checksums.txt --ignore-missing
```

Alternative using `grep`:

```bash
# Replace FILENAME with your downloaded file
FILENAME="pass-cli_X.Y.Z_linux_amd64.tar.gz"

# Extract expected checksum
EXPECTED=$(grep "$FILENAME" checksums.txt | cut -d' ' -f1)

# Calculate actual checksum
ACTUAL=$(sha256sum "$FILENAME" | cut -d' ' -f1)

# Compare
if [ "$EXPECTED" = "$ACTUAL" ]; then
    echo "Checksum verified!"
else
    echo "Checksum mismatch! Do not install."
    exit 1
fi
```

#### Windows (PowerShell)

```powershell
# After downloading from https://github.com/ari1110/pass-cli/releases/latest
# Replace with your downloaded filename
$file = "pass-cli_X.Y.Z_windows_amd64.zip"

# Extract expected checksum
$expected = (Get-Content checksums.txt | Select-String $file).ToString().Split()[0]

# Calculate actual checksum
$actual = (Get-FileHash $file -Algorithm SHA256).Hash.ToLower()

# Compare
if ($expected -eq $actual) {
    Write-Host "Checksum verified!" -ForegroundColor Green
} else {
    Write-Host "Checksum mismatch! Do not install." -ForegroundColor Red
    exit 1
}
```

### Install Binary

#### macOS / Linux

```bash
# Extract the archive
tar -xzf pass-cli_*_linux_amd64.tar.gz

# Make binary executable (should already be)
chmod +x pass-cli

# Move to a directory in PATH
sudo mv pass-cli /usr/local/bin/

# Verify installation
pass-cli version
```

Alternative user-specific installation (no sudo):

```bash
# Create local bin directory if it doesn't exist
mkdir -p ~/.local/bin

# Move binary
mv pass-cli ~/.local/bin/

# Add to PATH in ~/.bashrc or ~/.zshrc if not already there
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify installation
pass-cli version
```

#### Windows

**Using File Explorer:**

1. Extract the `.zip` file
2. Copy `pass-cli.exe` to a directory in your PATH (e.g., `C:\Program Files\pass-cli\`)
3. Or create a new directory and add it to PATH:
   - Create `C:\Tools\pass-cli\`
   - Copy `pass-cli.exe` to it
   - Add to PATH: System Properties → Environment Variables → Path → New → `C:\Tools\pass-cli`

**Using PowerShell:**

```powershell
# Extract the archive
Expand-Archive pass-cli_*_windows_amd64.zip -DestinationPath .

# Create installation directory
$installDir = "$env:LOCALAPPDATA\Programs\pass-cli"
New-Item -ItemType Directory -Force -Path $installDir

# Move binary
Move-Item pass-cli.exe $installDir\

# Add to PATH (current user)
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable(
        "Path",
        "$path;$installDir",
        "User"
    )
}

# Refresh environment (restart PowerShell or run)
$env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")

# Verify installation
pass-cli version
```

## Building from Source

Building from source gives you the latest development version and allows customization.

### Prerequisites

- **Go**: Version 1.25 or later ([download](https://golang.org/dl/))
- **Git**: For cloning the repository

Verify prerequisites:

```bash
go version    # Should show 1.25+
git --version
```

### Build Steps

#### Clone Repository

```bash
# Clone the repository
git clone https://github.com/ari1110/pass-cli.git
cd pass-cli

# Checkout specific version (optional)
git checkout v0.0.1

# Or use main branch for latest
git checkout main
```

#### Build Binary

```bash
# Build for current platform
go build -o pass-cli .

# Or with optimizations (smaller binary)
go build -ldflags="-s -w" -o pass-cli .

# Verify
./pass-cli version
```

**Build with version information:**

```bash
# Set version variables
VERSION="0.0.1"
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build with ldflags
go build -ldflags="-s -w \
  -X pass-cli/cmd.version=${VERSION} \
  -X pass-cli/cmd.commit=${COMMIT} \
  -X pass-cli/cmd.date=${DATE}" \
  -o pass-cli .

# Verify version info
./pass-cli version --verbose
```

#### Install Binary

```bash
# macOS/Linux: Move to PATH
sudo mv pass-cli /usr/local/bin/

# Or user-specific
mv pass-cli ~/.local/bin/

# Windows: Move to a directory in PATH
# Move-Item pass-cli.exe $env:LOCALAPPDATA\Programs\pass-cli\
```

### Build Options

#### Cross-Compilation

Build for different platforms:

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o pass-cli-linux-amd64 .

# Build for macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o pass-cli-darwin-amd64 .

# Build for macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o pass-cli-darwin-arm64 .

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o pass-cli-windows-amd64.exe .
```

#### Static Linking

For maximum portability:

```bash
# Static binary (no external dependencies)
CGO_ENABLED=0 go build -ldflags="-s -w" -o pass-cli .
```

#### All Platforms (using GoReleaser)

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Build for all platforms (snapshot mode)
goreleaser build --snapshot --clean

# Binaries will be in dist/
ls dist/
```

### Run Tests

```bash
# Run all tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests
go test -v -tags=integration -timeout 5m ./test
```

## Post-Installation

After installation, verify everything works:

```bash
# Check version
pass-cli version

# Check help
pass-cli --help

# Initialize a vault (creates ~/.pass-cli/)
pass-cli init

# Launch interactive TUI mode (recommended for new users)
pass-cli
# Press 'q' to quit TUI

# Or use CLI commands
pass-cli add test      # Add a test credential
pass-cli get test      # Retrieve it
pass-cli delete test --force  # Clean up test
```

**TUI vs CLI Mode:**
- **TUI Mode**: Run `pass-cli` with no arguments for interactive visual interface
  - Best for: Browsing credentials, interactive management
  - Features: Search, keyboard shortcuts, visual feedback
- **CLI Mode**: Run `pass-cli <command>` with explicit subcommand
  - Best for: Scripts, automation, quick single operations
  - Features: Quiet mode, field extraction, scriptable output

## Troubleshooting

### Command Not Found

**Symptom**: `pass-cli: command not found`

**Solutions**:

1. **Verify binary location**
   ```bash
   which pass-cli  # Should show path
   ```

2. **Check PATH**
   ```bash
   echo $PATH  # Should include installation directory
   ```

3. **Add to PATH** (if missing)
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PATH="$PATH:/path/to/pass-cli"
   source ~/.bashrc
   ```

4. **Windows**: Restart PowerShell/CMD after adding to PATH

### Permission Denied

**Symptom**: `Permission denied` when running pass-cli

**Solutions**:

```bash
# Make executable
chmod +x /path/to/pass-cli

# Or reinstall with correct permissions
sudo install -m 755 pass-cli /usr/local/bin/
```

### Checksum Mismatch

**Symptom**: Checksum verification fails

**Solutions**:

1. **Re-download** the file (may be corrupted)
2. **Verify** you downloaded the correct platform file
3. **Check** checksums.txt is for the same version
4. **Report** if problem persists (possible security issue)

### Cannot Execute Binary

**macOS Symptom**: "pass-cli cannot be opened because the developer cannot be verified"

**Solution**:

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /path/to/pass-cli

# Or in System Preferences:
# Security & Privacy → General → "Allow anyway"
```

### Go Build Fails

**Symptom**: Build errors when compiling from source

**Solutions**:

1. **Update Go**
   ```bash
   go version  # Should be 1.25+
   ```

2. **Clean module cache**
   ```bash
   go clean -modcache
   go mod download
   ```

3. **Update dependencies**
   ```bash
   go mod tidy
   go mod verify
   ```

### Homebrew Installation Fails

**Solutions**:

```bash
# Update Homebrew
brew update

# Check for conflicts
brew doctor

# Verbose installation
brew install --verbose pass-cli

# Force reinstall
brew reinstall pass-cli
```

### Scoop Installation Fails

**Solutions**:

```powershell
# Update Scoop
scoop update

# Check status
scoop status

# Force reinstall
scoop uninstall pass-cli
scoop install pass-cli

# Check logs
scoop cat pass-cli
```

## Uninstallation

### Homebrew

```bash
# Uninstall Pass-CLI
brew uninstall pass-cli

# Remove tap (optional)
brew untap ari1110/homebrew-tap

# Remove vault (if desired)
rm -rf ~/.pass-cli
```

### Scoop

```powershell
# Uninstall Pass-CLI
scoop uninstall pass-cli

# Remove bucket (optional)
scoop bucket rm pass-cli

# Remove vault (if desired)
Remove-Item -Recurse -Force ~/.pass-cli
```

### Manual Installation

```bash
# Remove binary
sudo rm /usr/local/bin/pass-cli

# Or user installation
rm ~/.local/bin/pass-cli

# Remove vault (if desired)
rm -rf ~/.pass-cli
```

### Windows Manual Installation

```powershell
# Remove binary
Remove-Item "C:\Program Files\pass-cli\pass-cli.exe"

# Remove from PATH (if manually added)
# System Properties → Environment Variables → Edit Path

# Remove vault (if desired)
Remove-Item -Recurse -Force "$env:USERPROFILE\.pass-cli"
```

### Complete Removal

To completely remove all traces of Pass-CLI:

```bash
# 1. Uninstall binary (using method above)

# 2. Remove vault
rm -rf ~/.pass-cli

# 3. Remove master password from keychain
# macOS: Open Keychain Access → Search "pass-cli" → Delete
# Linux: Use your keyring manager (Seahorse, etc.)
# Windows: Credential Manager → Remove pass-cli entries

# 4. Remove config (if exists)
rm ~/.pass-cli.yaml

# 5. Clear shell history (optional)
history -c
```

## Platform-Specific Notes

### macOS

- **Apple Silicon**: Use ARM64 version for native performance
- **Intel**: Use amd64 version
- **Keychain**: Integration is automatic
- **Homebrew**: Recommended installation method

### Linux

- **Package Managers**: Homebrew works on Linux too
- **Keychain**: Requires Secret Service (GNOME Keyring or KWallet)
- **AppArmor/SELinux**: May need profile adjustments for keychain access
- **Distribution Packages**: May become available for specific distros

### Windows

- **Scoop**: Recommended installation method
- **Credential Manager**: Integration is automatic
- **Antivirus**: May need to whitelist pass-cli.exe
- **PATH**: Requires manual setup for manual installation

## Getting Help

If you encounter issues not covered here:

1. Check the [Troubleshooting Guide](TROUBLESHOOTING.md)
2. Review [GitHub Issues](https://github.com/ari1110/pass-cli/issues)
3. Ask in [GitHub Discussions](https://github.com/ari1110/pass-cli/discussions)
4. File a [new issue](https://github.com/ari1110/pass-cli/issues/new)

## Next Steps

After installation:

- Read the [Usage Guide](USAGE.md)
- Review [Security Documentation](SECURITY.md)
- Check out [Script Integration Examples](../README.md#-script-integration)

