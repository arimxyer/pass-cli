# PowerShell script examples from Pass-CLI USAGE.md documentation
# This script contains executable PowerShell commands for testing pass-cli on Windows

# From "Script Integration" - "PowerShell Examples" - "Export to environment variable"
Write-Host "Export to environment variable examples:" -ForegroundColor Green

# Export password
$env:DB_PASSWORD = pass-cli get database --quiet

# Export specific field
$env:DB_USER = pass-cli get database --field username --quiet

# Use in command
$apiKey = pass-cli get openai --quiet
Invoke-RestMethod -Uri "https://api.openai.com" -Headers @{
    "Authorization" = "Bearer $apiKey"
}

# From "Script Integration" - "PowerShell Examples" - "Conditional execution"
Write-Host "Conditional execution examples:" -ForegroundColor Green

# Check if credential exists
try {
    $password = pass-cli get myservice --quiet 2>$null
    Write-Host "Credential exists"
    $env:API_KEY = $password
} catch {
    Write-Host "Credential not found"
    exit 1
}

Write-Host "PowerShell examples completed." -ForegroundColor Green