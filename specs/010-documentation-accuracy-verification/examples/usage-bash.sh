#!/bin/bash

# Comprehensive bash script containing ALL executable bash code blocks from USAGE.md
# Extracted from: R:\Test-Projects\pass-cli\docs\USAGE.md
# Generated on: 2025-10-15
# Only contains executable commands, excludes output examples and prompts

set -e  # Exit on any error

echo "=== Pass-CLI Usage Examples ==="
echo "This script contains all executable bash code blocks from USAGE.md"
echo "Source file: docs/USAGE.md"
echo ""

# =============================================================================
# SECTION: Global Flag Examples (lines 36-45)
# =============================================================================
echo "=== Global Flag Examples ==="
echo "From USAGE.md Global Flag Examples section"
echo ""

# Use custom vault location
pass-cli --vault /secure/vault.enc list

# Enable verbose logging
pass-cli --verbose get github

# Get help for any command
pass-cli get --help

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: init command examples (lines 55-71)
# =============================================================================
echo "=== init Command Examples ==="
echo "From USAGE.md init command Examples section"
echo ""

# Synopsis
pass-cli init

# Initialize with default location
pass-cli init

# Initialize with custom location
pass-cli --vault /custom/path/vault.enc init

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Audit Logging Example (lines 99-102)
# =============================================================================
echo "=== Audit Logging Example ==="
echo "From USAGE.md Audit Logging section"
echo ""

# Initialize vault with audit logging
pass-cli init --enable-audit

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Audit Verification (lines 112-115)
# =============================================================================
echo "=== Audit Verification ==="
echo "From USAGE.md Audit Verification section"
echo ""

# Verify audit log integrity
pass-cli verify-audit

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: add command examples (lines 133-171)
# =============================================================================
echo "=== add Command Examples ==="
echo "From USAGE.md add command Examples section"
echo ""

# Synopsis
pass-cli add <service> [flags]

# Interactive mode (prompts for username/password)
pass-cli add github

# With username flag
pass-cli add github --username user@example.com

# With URL and notes
pass-cli add github \
  --username user@example.com \
  --url https://github.com \
  --notes "Personal account"

# With category
pass-cli add github -u user@example.com -c "Version Control"

# All flags (not recommended for password)
pass-cli add github \
  -u user@example.com \
  -p secret123 \
  --url https://github.com \
  --notes "Work account"

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: get command examples (lines 207-254)
# =============================================================================
echo "=== get Command Examples ==="
echo "From USAGE.md get command Examples section"
echo ""

# Synopsis
pass-cli get <service> [flags]

# Default: Display credential and copy to clipboard
pass-cli get github

# Quiet mode (password only, for scripts)
pass-cli get github --quiet
pass-cli get github -q

# Get specific field
pass-cli get github --field username
pass-cli get github -f url

# Quiet mode with specific field
pass-cli get github --field username --quiet

# Display without clipboard
pass-cli get github --no-clipboard

# Display with masked password
pass-cli get github --masked

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: list command examples (lines 295-324)
# =============================================================================
echo "=== list Command Examples ==="
echo "From USAGE.md list command Examples section"
echo ""

# Synopsis
pass-cli list [flags]

# List all credentials (table format)
pass-cli list

# List as JSON
pass-cli list --format json

# Simple list (service names only)
pass-cli list --format simple

# Show unused credentials (not accessed in 30 days)
pass-cli list --unused

# Show credentials not used in 90 days
pass-cli list --unused --days 90

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: update command examples (lines 361-405)
# =============================================================================
echo "=== update Command Examples ==="
echo "From USAGE.md update command Examples section"
echo ""

# Synopsis
pass-cli update <service> [flags]

# Update password (prompted)
pass-cli update github

# Update username
pass-cli update github --username newuser@example.com

# Update URL
pass-cli update github --url https://github.com/enterprise

# Update notes
pass-cli update github --notes "Updated account info"

# Update category
pass-cli update github --category "Work"

# Clear category field
pass-cli update github --clear-category

# Update multiple fields
pass-cli update github \
  --username newuser@example.com \
  --url https://github.com/enterprise \
  --notes "Corporate account"

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: delete command examples (lines 430-449)
# =============================================================================
echo "=== delete Command Examples ==="
echo "From USAGE.md delete command Examples section"
echo ""

# Synopsis
pass-cli delete <service> [flags]

# Delete with confirmation
pass-cli delete github

# Force delete (no confirmation)
pass-cli delete github --force
pass-cli delete github -f

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: generate command examples (lines 474-513)
# =============================================================================
echo "=== generate Command Examples ==="
echo "From USAGE.md generate command Examples section"
echo ""

# Synopsis
pass-cli generate [flags]

# Generate default password (20 chars, all character types)
pass-cli generate

# Custom length
pass-cli generate --length 32

# Alphanumeric only (no symbols)
pass-cli generate --no-symbols

# Digits and symbols only
pass-cli generate --no-lower --no-upper

# Letters only (no digits or symbols)
pass-cli generate --no-digits --no-symbols

# Display only (no clipboard)
pass-cli generate --no-clipboard

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: version command examples (lines 539-557)
# =============================================================================
echo "=== version Command Examples ==="
echo "From USAGE.md version command Examples section"
echo ""

# Synopsis
pass-cli version [flags]

# Show version
pass-cli version

# Verbose version info
pass-cli version --verbose

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Output Mode Examples (lines 584-612)
# =============================================================================
echo "=== Output Mode Examples ==="
echo "From USAGE.md Output Modes section"
echo ""

# Human-Readable (Default)
pass-cli get github

# Quiet Mode
pass-cli get github --quiet
pass-cli get github --field username --quiet

# Simple Mode (List Only)
pass-cli list --format simple

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Bash/Zsh Script Integration (lines 620-657)
# =============================================================================
echo "=== Bash/Zsh Script Integration ==="
echo "From USAGE.md Script Integration section"
echo ""

# Export to environment variable
#!/bin/bash

# Export password
export DB_PASSWORD=$(pass-cli get database --quiet)

# Export specific field
export DB_USER=$(pass-cli get database --field username --quiet)

# Use in command
mysql -u "$(pass-cli get database -f username -q)" \
      -p"$(pass-cli get database -q)" \
      mydb

# Conditional execution
# Check if credential exists
if pass-cli get myservice --quiet &>/dev/null; then
    echo "Credential exists"
    export API_KEY=$(pass-cli get myservice --quiet)
else
    echo "Credential not found"
    exit 1
fi

# Loop through credentials
# Process all credentials
for service in $(pass-cli list --format simple); do
    echo "Processing $service..."
    username=$(pass-cli get "$service" --field username --quiet)
    echo "  Username: $username"
done

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Environment Variables (lines 736-753)
# =============================================================================
echo "=== Environment Variables ==="
echo "From USAGE.md Environment Variables section"
echo ""

# PASS_CLI_VAULT - Override default vault location
# Bash
export PASS_CLI_VAULT=/custom/path/vault.enc
pass-cli list

# PASS_CLI_VERBOSE - Enable verbose logging
export PASS_CLI_VERBOSE=1
pass-cli get github

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Configuration Commands (lines 762-774)
# =============================================================================
echo "=== Configuration Commands ==="
echo "From USAGE.md Configuration section"
echo ""

# Initialize default config
pass-cli config init

# Edit config in default editor
pass-cli config edit

# Validate config syntax
pass-cli config validate

# Reset to defaults
pass-cli config reset

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: TUI Mode Launch (lines 845-850)
# =============================================================================
echo "=== TUI Mode Launch ==="
echo "From USAGE.md TUI Mode section"
echo ""

# Launch TUI (no arguments)
pass-cli

# TUI opens automatically when no subcommand is provided

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: TUI vs CLI Mode Examples (lines 867-876)
# =============================================================================
echo "=== TUI vs CLI Mode Examples ==="
echo "From USAGE.md TUI vs CLI Mode section"
echo ""

# TUI Mode
pass-cli                        # Opens interactive interface

# CLI Mode
pass-cli list                   # Outputs credential table to stdout
pass-cli get github --quiet     # Outputs password only (script-friendly)
pass-cli add newcred            # Interactive prompts for credential data

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Usage Tracking Examples (lines 1111-1117)
# =============================================================================
echo "=== Usage Tracking Examples ==="
echo "From USAGE.md Usage Tracking section"
echo ""

# Access from project directory
cd ~/projects/my-app
pass-cli get database

# Usage tracking is automatic based on current directory

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Viewing Usage (lines 1127-1130)
# =============================================================================
echo "=== Viewing Usage ==="
echo "From USAGE.md Usage Tracking section"
echo ""

# List unused credentials
pass-cli list --unused --days 30

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Good Script Example (lines 1157-1165)
# =============================================================================
echo "=== Good Script Example ==="
echo "From USAGE.md Best Practices section"
echo ""

# Good:
export API_KEY=$(pass-cli get service --quiet 2>/dev/null)
if [ -z "$API_KEY" ]; then
    echo "Failed to get credential" >&2
    exit 1
fi

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Common Patterns - CI/CD Pipeline (lines 1175-1184)
# =============================================================================
echo "=== Common Patterns - CI/CD Pipeline ==="
echo "From USAGE.md Common Patterns section"
echo ""

# Retrieve deployment credentials
export DEPLOY_KEY=$(pass-cli get production --quiet)
export DB_PASSWORD=$(pass-cli get prod-db --quiet)

# Run deployment
./deploy.sh

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Common Patterns - Local Development (lines 1186-1196)
# =============================================================================
echo "=== Common Patterns - Local Development ==="
echo "From USAGE.md Common Patterns section"
echo ""

# Set up environment from credentials
export DB_HOST=$(pass-cli get dev-db --field url --quiet)
export DB_USER=$(pass-cli get dev-db --field username --quiet)
export DB_PASS=$(pass-cli get dev-db --quiet)

# Start development server
npm run dev

echo ""
echo "---"
echo ""

# =============================================================================
# SECTION: Common Patterns - Credential Rotation (lines 1198-1209)
# =============================================================================
echo "=== Common Patterns - Credential Rotation ==="
echo "From USAGE.md Common Patterns section"
echo ""

# Generate new password
NEW_PWD=$(pass-cli generate --length 32 --quiet)

# Update service
pass-cli update myservice --password "$NEW_PWD"

# Use new password
echo "$NEW_PWD" | some-service-update-command

echo ""
echo "=== End of Pass-CLI Usage Examples ==="
echo "Total sections extracted: $(grep -c "^# ===.*===" "$0")"
echo "Script completed successfully!"
echo ""
echo "Note: This script contains only executable bash commands from USAGE.md"
echo "Output examples and prompts (lines starting with \$) have been excluded"