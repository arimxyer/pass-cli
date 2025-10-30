# update-version.ps1 - Update version numbers and dates across documentation
param([Parameter(Mandatory=$true)][string]$NewVersion)

$VersionNoV = $NewVersion -replace '^v', ''
$CurrentDate = Get-Date -Format "MMMM yyyy"

Write-Host "=== Pass-CLI Version Update Tool ===" -ForegroundColor Blue
Write-Host "New version: $NewVersion" -ForegroundColor Green
Write-Host "Update date: $CurrentDate" -ForegroundColor Green
Write-Host ""

if (-not (Test-Path "go.mod") -or -not (Test-Path "docs")) {
    Write-Host "Error: Must run from project root directory" -ForegroundColor Red
    exit 1
}

$gitStatus = git status --porcelain 2>$null
if ($gitStatus) {
    Write-Host "Warning: You have uncommitted changes" -ForegroundColor Yellow
    $response = Read-Host "Continue anyway? (y/N)"
    if ($response -ne 'y' -and $response -ne 'Y') { exit 1 }
}

Write-Host "Updating documentation files..." -ForegroundColor Blue

$UpdatedFiles = 0

function Update-DocFile {
    param([string]$Path, [string]$Desc)
    if (-not (Test-Path $Path)) { return }

    $content = Get-Content $Path -Raw
    $original = $content

    $content = $content -replace '\*\*Documentation Version\*\*: v[\d\.]+', "**Documentation Version**: $NewVersion"
    $content = $content -replace '\*\*Last Updated\*\*: \w+ \d+', "**Last Updated**: $CurrentDate"

    if ($content -ne $original) {
        $content | Set-Content $Path -NoNewline
        Write-Host "  ✓ $Desc" -ForegroundColor Green
        $script:UpdatedFiles++
    }
}

Update-DocFile "docs\USAGE.md" "USAGE.md"
Update-DocFile "docs\SECURITY.md" "SECURITY.md"
Update-DocFile "docs\MIGRATION.md" "MIGRATION.md"
Update-DocFile "docs\TROUBLESHOOTING.md" "TROUBLESHOOTING.md"
Update-DocFile "docs\KNOWN_LIMITATIONS.md" "KNOWN_LIMITATIONS.md"
Update-DocFile "docs\GETTING_STARTED.md" "GETTING_STARTED.md"
Update-DocFile "docs\INSTALLATION.md" "INSTALLATION.md"
Update-DocFile "docs\DOCTOR_COMMAND.md" "DOCTOR_COMMAND.md"

Write-Host ""
Write-Host "Updating package manifests..." -ForegroundColor Blue

if (Test-Path "homebrew\pass-cli.rb") {
    $content = Get-Content "homebrew\pass-cli.rb" -Raw
    $newContent = $content -replace 'version "[\d\.]+"', ('version "' + $VersionNoV + '"')
    if ($content -ne $newContent) {
        $newContent | Set-Content "homebrew\pass-cli.rb" -NoNewline
        Write-Host "  ✓ homebrew/pass-cli.rb" -ForegroundColor Green
        $UpdatedFiles++
    }
}

if (Test-Path "scoop\pass-cli.json") {
    $content = Get-Content "scoop\pass-cli.json" -Raw
    $newContent = $content -replace '"version": "[\d\.]+"', ('"version": "' + $VersionNoV + '"')
    if ($content -ne $newContent) {
        $newContent | Set-Content "scoop\pass-cli.json" -NoNewline
        Write-Host "  ✓ scoop/pass-cli.json" -ForegroundColor Green
        $UpdatedFiles++
    }
}

Write-Host ""
Write-Host "=== Update Complete ===" -ForegroundColor Green
Write-Host "Updated $UpdatedFiles files"
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Review changes: git diff"
Write-Host "  2. Update CHANGELOG.md with release notes"
Write-Host "  3. Commit: git add -A; git commit -m 'chore: bump version to $NewVersion'"
Write-Host "  4. Tag: git tag -a $NewVersion -m 'Release $NewVersion'"
Write-Host "  5. Push: git push origin main --tags"
Write-Host ""
Write-Host "Note: Package URLs/checksums updated automatically by GoReleaser" -ForegroundColor Blue
