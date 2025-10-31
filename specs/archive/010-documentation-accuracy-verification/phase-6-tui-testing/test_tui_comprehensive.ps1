# PowerShell script to test TUI functionality
# This script will test TUI mode and document findings

Write-Host "=== Pass-CLI TUI Mode Test ===" -ForegroundColor Green
Write-Host ""

# Test variables
$PassCliPath = ".\pass-cli.exe"
$TestVault = "$env:USERPROFILE\.pass-cli\test_vault.enc"
$MasterPassword = "TestMasterP@ss123"

Write-Host "Test Configuration:" -ForegroundColor Yellow
Write-Host "- Binary: $PassCliPath"
Write-Host "- Vault: $TestVault"
Write-Host "- Master Password: $MasterPassword"
Write-Host ""

# Test 1: Verify TUI command exists
Write-Host "Test 1: Verify TUI command exists" -ForegroundColor Cyan
try {
    $result = & $PassCliPath tui --help
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ TUI command is available" -ForegroundColor Green
    } else {
        Write-Host "✗ TUI command failed" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Error running TUI command: $_" -ForegroundColor Red
}
Write-Host ""

# Test 2: Test basic TUI launch (with timeout)
Write-Host "Test 2: Test TUI launch (this will timeout after 3 seconds)" -ForegroundColor Cyan
try {
    $process = Start-Process -FilePath $PassCliPath -ArgumentList @("--vault", $TestVault, "tui") -PassThru -RedirectStandardInput "test_input.txt" -RedirectStandardOutput "tui_output.txt" -RedirectStandardError "tui_error.txt"

    # Send master password
    $inputText = $MasterPassword + "`r`n"
    $inputBytes = [System.Text.Encoding]::UTF8.GetBytes($inputText)
    $process.StandardInput.Write($inputBytes, 0, $inputBytes.Length)
    $process.StandardInput.Close()

    # Wait for 3 seconds then terminate
    Start-Sleep -Seconds 3
    if (!$process.HasExited) {
        $process.Kill()
        Write-Host "✓ TUI launched successfully (terminated after timeout)" -ForegroundColor Green
    } else {
        Write-Host "⚠ TUI terminated early" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Error launching TUI: $_" -ForegroundColor Red
}
Write-Host ""

# Test 3: Check TUI source code for keyboard shortcuts
Write-Host "Test 3: Analyzing TUI keyboard shortcuts from source code" -ForegroundColor Cyan

# Look for TUI-related files
$tuiFiles = Get-ChildItem -Path "." -Recurse -Filter "*.go" | Where-Object { $_.Name -match "(tui|ui)" -or $_.FullName -match "(tui|ui)" }

if ($tuiFiles.Count -gt 0) {
    Write-Host "Found TUI-related files:" -ForegroundColor Green
    foreach ($file in $tuiFiles) {
        Write-Host "  - $($file.FullName)"
    }

    # Search for keyboard shortcuts in TUI files
    Write-Host "`nKeyboard shortcuts found in source:" -ForegroundColor Green
    foreach ($file in $tuiFiles) {
        $content = Get-Content $file.FullName -Raw
        if ($content -match "Ctrl\+H|ctrl\+h") {
            Write-Host "  ✓ Ctrl+H found in $($file.Name)" -ForegroundColor Green
        }
        if ($content -match "Ctrl\+C|ctrl\+c") {
            Write-Host "  ✓ Ctrl+C found in $($file.Name)" -ForegroundColor Green
        }
        if ($content -match "Tab|`\t") {
            Write-Host "  ✓ Tab navigation found in $($file.Name)" -ForegroundColor Green
        }
        if ($content -match "arrow|Arrow|keyup|keydown") {
            Write-Host "  ✓ Arrow key navigation found in $($file.Name)" -ForegroundColor Green
        }
        if ($content -match "Escape|Esc|\\x1b") {
            Write-Host "  ✓ Escape key found in $($file.Name)" -ForegroundColor Green
        }
    }
} else {
    Write-Host "⚠ No TUI-related Go files found" -ForegroundColor Yellow
}
Write-Host ""

# Test 4: Manual testing instructions
Write-Host "Test 4: Manual Testing Instructions" -ForegroundColor Cyan
Write-Host "To manually test TUI functionality:" -ForegroundColor Yellow
Write-Host "1. Run: .\pass-cli.exe --vault $TestVault tui"
Write-Host "2. Enter master password when prompted: $MasterPassword"
Write-Host "3. Test these keyboard shortcuts according to README.md:"
Write-Host ""
Write-Host "Configurable shortcuts:" -ForegroundColor White
Write-Host "  - q: Quit application"
Write-Host "  - a: Add new credential"
Write-Host "  - e: Edit credential"
Write-Host "  - d: Delete credential"
Write-Host "  - i: Toggle detail panel"
Write-Host "  - s: Toggle sidebar"
Write-Host "  - ?: Show help modal"
Write-Host "  - /: Activate search/filter"
Write-Host ""
Write-Host "Hardcoded shortcuts:" -ForegroundColor White
Write-Host "  - Tab: Next component"
Write-Host "  - Shift+Tab: Previous component"
Write-Host "  - Arrow Up/Down: Navigate lists"
Write-Host "  - Enter: Select/View details"
Write-Host "  - Esc: Close modal/Exit search"
Write-Host "  - Ctrl+C: Force quit application"
Write-Host "  - p: Copy password (detail view)"
Write-Host "  - c: Copy username (detail view)"
Write-Host "  - Ctrl+H: Toggle password visibility (forms)"
Write-Host "  - Ctrl+S: Quick-save/Submit form"
Write-Host "  - PgUp/PgDn: Scroll help modal"
Write-Host ""

Write-Host "=== Expected TUI Behavior from README.md ===" -ForegroundColor Green
Write-Host "- Visual navigation with arrow keys and Tab"
Write-Host "- Interactive forms with visual feedback"
Write-Host "- Password visibility toggle with Ctrl+H in forms"
Write-Host "- Search & filter with / key"
Write-Host "- Help modal with ? key"
Write-Host "- Responsive layout with sidebar and detail panel"
Write-Host "- Minimum terminal size: 60 columns × 30 rows"
Write-Host ""

Write-Host "=== Test Complete ===" -ForegroundColor Green