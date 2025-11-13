---
title: "Quick Install"
weight: 1
bookToc: true
---

# Quick Install Guide
Fast installation using package managers for Pass-CLI across all supported platforms.

![Version](https://img.shields.io/github/v/release/ari1110/pass-cli?label=Version) ![Last Updated](https://img.shields.io/github/last-commit/ari1110/pass-cli?path=docs&label=Last%20Updated)

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

