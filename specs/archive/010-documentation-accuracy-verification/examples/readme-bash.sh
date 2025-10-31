#!/bin/bash
# Extracted bash code blocks from README.md for verification
# This file contains ALL bash code examples from README.md

# Installation examples
echo "# Installation examples from README.md"
echo "brew tap ari1110/homebrew-tap"
echo "brew install pass-cli"

echo "# Manual installation"
echo "tar -xzf pass-cli_*_<os>_<arch>.tar.gz"
echo "sudo mv pass-cli /usr/local/bin/"

# First steps examples
echo "# First steps examples from README.md"
echo "pass-cli init"
echo "pass-cli add github"
echo "pass-cli get github"
echo "export API_KEY=\$(pass-cli get myservice --quiet --field password)"

# TUI launch examples
echo "# TUI launch examples from README.md"
echo "pass-cli"
echo "pass-cli list"
echo "pass-cli get github"

# Initialize vault examples
echo "# Initialize vault examples from README.md"
echo "pass-cli init"

# Add credentials examples
echo "# Add credentials examples from README.md"
echo "pass-cli add myservice"
echo "pass-cli add github --url https://github.com --notes \"Personal account\""
echo "pass-cli generate"
echo "pass-cli add newservice"

# Retrieve credentials examples
echo "# Retrieve credentials examples from README.md"
echo "pass-cli get myservice"
echo "pass-cli get myservice --copy"
echo "pass-cli get myservice --quiet"
echo "pass-cli get myservice --field username"
echo "pass-cli get myservice --masked"

# List credentials examples
echo "# List credentials examples from README.md"
echo "pass-cli list"
echo "pass-cli list --unused"

# Update credentials examples
echo "# Update credentials examples from README.md"
echo "pass-cli update myservice"
echo "pass-cli update myservice --username newuser@example.com"
echo "pass-cli update myservice --url https://new-url.com"
echo "pass-cli update myservice --notes \"Updated notes\""

# Delete credentials examples
echo "# Delete credentials examples from README.md"
echo "pass-cli delete myservice"
echo "pass-cli delete myservice --force"

# Generate passwords examples
echo "# Generate passwords examples from README.md"
echo "pass-cli generate"
echo "pass-cli generate --length 32"
echo "pass-cli generate --no-symbols"

# Version information examples
echo "# Version information examples from README.md"
echo "pass-cli version"
echo "pass-cli version --verbose"

# Audit logging examples
echo "# Audit logging examples from README.md"
echo "pass-cli init --enable-audit"
echo "pass-cli verify-audit"

# Script integration examples
echo "# Script integration examples from README.md"
echo "export API_KEY=\$(pass-cli get openai --quiet --field password)"
echo "curl -H \"Authorization: Bearer \$(pass-cli get github --quiet)\" https://api.github.com/user"
echo "if pass-cli get myservice --quiet > /dev/null 2>&1; then echo \"Credential exists\"; fi"

# Usage tracking examples
echo "# Usage tracking examples from README.md"
echo "cd ~/project-a && pass-cli get database"
echo "cd ~/project-b && pass-cli get database"
echo "pass-cli list --unused --days 30"

# Advanced usage examples
echo "# Advanced usage examples from README.md"
echo "pass-cli --vault /path/to/custom/vault.enc list"
echo "export PASS_CLI_VAULT=/path/to/custom/vault.enc"
echo "pass-cli list"
echo "pass-cli --verbose get myservice"

# Configuration examples
echo "# Configuration examples from README.md"
echo "pass-cli config init"
echo "pass-cli config edit"
echo "pass-cli config validate"
echo "pass-cli config reset"

# Build from source examples
echo "# Build from source examples from README.md"
echo "git clone https://github.com/ari1110/pass-cli.git"
echo "cd pass-cli"
echo "go build -o pass-cli ."
echo "make build"
echo "make test"
echo "make test-coverage"

# Running tests examples
echo "# Running tests examples from README.md"
echo "go test ./..."
echo "go test -cover ./..."
echo "go test -tags=integration ./test/"
echo "make test-all"

# Code quality examples
echo "# Code quality examples from README.md"
echo "make lint"
echo "make security-scan"
echo "make fmt"

# Backup examples
echo "# Backup examples from README.md"
echo "cp ~/.pass-cli/vault.enc ~/backup/vault-\$(date +%Y%m%d).enc"