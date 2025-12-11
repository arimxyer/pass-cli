# Cleanup orphaned pass-cli keychain entries from integration tests
# Usage: powershell -ExecutionPolicy Bypass -File scripts/cleanup-keychain.ps1

$keepPattern = "master-password-_pass-cli$"

Write-Host "Scanning for pass-cli keychain entries..." -ForegroundColor Cyan

# Get all entries
$output = cmdkey /list
$entries = $output | Select-String "target=pass-cli"

$toKeep = @()
$toDelete = @()

foreach ($entry in $entries) {
    if ($entry -match "target=([^\s]+)") {
        $target = $matches[1]
        if ($target -match $keepPattern) {
            $toKeep += $target
        } else {
            $toDelete += $target
        }
    }
}

Write-Host "`nFound $($entries.Count) total pass-cli entries"
Write-Host "Keeping: $($toKeep.Count)" -ForegroundColor Green
foreach ($t in $toKeep) {
    Write-Host "  - $t" -ForegroundColor Green
}

Write-Host "`nTo delete: $($toDelete.Count)" -ForegroundColor Yellow

if ($toDelete.Count -eq 0) {
    Write-Host "Nothing to clean up!" -ForegroundColor Green
    exit 0
}

$confirm = Read-Host "`nDelete $($toDelete.Count) entries? (y/N)"
if ($confirm -ne "y") {
    Write-Host "Aborted." -ForegroundColor Red
    exit 1
}

$deleted = 0
$failed = 0

foreach ($target in $toDelete) {
    Write-Host "Deleting: $target" -ForegroundColor Gray
    $result = cmdkey /delete:$target 2>&1
    if ($LASTEXITCODE -eq 0) {
        $deleted++
    } else {
        Write-Host "  Failed: $result" -ForegroundColor Red
        $failed++
    }
}

Write-Host "`nDeleted: $deleted" -ForegroundColor Green
if ($failed -gt 0) {
    Write-Host "Failed: $failed" -ForegroundColor Red
}
