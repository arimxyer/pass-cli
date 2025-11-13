---
title: "Scripting Guide"
weight: 6
toc: true
---

Automate pass-cli with scripts using quiet mode, JSON output, and environment variable integration.

## Output Modes

Pass-CLI supports multiple output modes for different use cases.

### Human-Readable (Default)

Formatted tables and colored output for terminal viewing.

```bash
pass-cli get github
# Service:  github
# Username: user@example.com
# Password: ****** (or full password)
```

### Quiet Mode

Single-line output, perfect for scripts.

```bash
pass-cli get github --quiet
# mySecretPassword123!

pass-cli get github --field username --quiet
# user@example.com
```

### Simple Mode (List Only)

Service names only, one per line.

```bash
pass-cli list --format simple
# github
# aws-prod
# database
```

## Script Integration

### Bash/Zsh Examples

**Export to environment variable:**

```bash
#!/bin/bash

# Export password
export SERVICE_PASSWORD=$(pass-cli get testservice --quiet)

# Export specific field
export SERVICE_USER=$(pass-cli get testservice --field username --quiet)

# Use in command
mysql -u "$(pass-cli get testservice -f username -q)" \
      -p"$(pass-cli get testservice -q)" \
      mydb
```

**Conditional execution:**

```bash
# Check if credential exists
if pass-cli get testservice --quiet &>/dev/null; then
    echo "Credential exists"
    export API_KEY=$(pass-cli get testservice --quiet)
else
    echo "Credential not found"
    exit 1
fi
```

**Loop through credentials:**

```bash
# Process all credentials
for service in $(pass-cli list --format simple); do
    echo "Processing $service..."
    username=$(pass-cli get "$service" --field username --quiet)
    echo "  Username: $username"
done
```

### PowerShell Examples

**Export to environment variable:**

```powershell
# Export password
$env:SERVICE_PASSWORD = pass-cli get testservice --quiet

# Export specific field
$env:SERVICE_USER = pass-cli get testservice --field username --quiet

# Use in command
$apiKey = pass-cli get github --quiet
Invoke-RestMethod -Uri "https://api.github.com" -Headers @{
    "Authorization" = "Bearer $apiKey"
}
```

**Conditional execution:**

```powershell
# Check if credential exists
try {
    $password = pass-cli get testservice --quiet 2>$null
    Write-Host "Credential exists"
    $env:API_KEY = $password
} catch {
    Write-Host "Credential not found"
    exit 1
}
```

### Python Examples

```python
import subprocess

# Get password only
result = subprocess.run(
    ['pass-cli', 'get', 'github', '--quiet'],
    capture_output=True,
    text=True,
    check=True
)
password = result.stdout.strip()

# Get specific field
result = subprocess.run(
    ['pass-cli', 'get', 'github', '--field', 'username', '--quiet'],
    capture_output=True,
    text=True,
    check=True
)
username = result.stdout.strip()
```

### Makefile Examples

```makefile
.PHONY: deploy
deploy:
	@export AWS_KEY=$$(pass-cli get aws --quiet --field username); \
	export AWS_SECRET=$$(pass-cli get aws --quiet); \
	./deploy.sh

.PHONY: test-db
test-db:
	@DB_URL="postgres://$$(pass-cli get testdb -f username -q):$$(pass-cli get testdb -q)@localhost/testdb" \
	go test ./...
```

## Environment Variables

### PASS_CLI_VERBOSE

Enable verbose logging.

```bash
# Bash
export PASS_CLI_VERBOSE=1
pass-cli get github

# PowerShell
$env:PASS_CLI_VERBOSE = "1"
pass-cli get github
```

**Note**: To use a custom vault location, configure `vault_path` in the config file (`~/.pass-cli/config.yml`) instead of using environment variables. See [Configuration](#configuration) section.

## Best Practices

### Security

1. **Never pass passwords via flags** - Use prompts or `--generate`
2. **Use quiet mode in scripts** - Prevents logging sensitive data
3. **Clear shell history** - When testing commands with passwords
4. **Use strong master passwords** - 20+ characters recommended

### Workflow

1. **Generate passwords** - Use `--generate` for new credentials
2. **Update regularly** - Rotate credentials periodically
3. **Track usage** - Review unused credentials monthly
4. **Backup vault** - Copy `~/.pass-cli/vault.enc` regularly

### Scripting

1. **Always use `--quiet`** - Clean output for variables
2. **Check exit codes** - Handle errors properly
3. **Use `--field`** - Extract exactly what you need
4. **Redirect stderr** - Control error output

### Examples

**Good:**
```bash
export API_KEY=$(pass-cli get service --quiet 2>/dev/null)
if [ -z "$API_KEY" ]; then
    echo "Failed to get credential" >&2
    exit 1
fi
```

**Bad:**
```bash
# Don't do this - exposes password in process list
pass-cli add service --password mySecretPassword
```

## Common Patterns

### CI/CD Pipeline

```bash
# Retrieve deployment credentials
export DEPLOY_KEY=$(pass-cli get production --quiet)
export DB_PASSWORD=$(pass-cli get prod-db --quiet)

# Run deployment
./deploy.sh
```

### Local Development

```bash
# Set up environment from credentials
export DB_HOST=$(pass-cli get dev-db --field url --quiet)
export DB_USER=$(pass-cli get dev-db --field username --quiet)
export DB_PASS=$(pass-cli get dev-db --quiet)

# Start development server
npm run dev
```

### Credential Rotation

```bash
# Generate new password
NEW_PWD=$(pass-cli generate --length 32 --quiet)

# Update service
pass-cli update testservice --password "$NEW_PWD"

# Use new password
echo "$NEW_PWD" | some-service-update-command
```

