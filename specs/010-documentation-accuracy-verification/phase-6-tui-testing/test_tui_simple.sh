#!/bin/bash

# Simple TUI test using timeout
# This will test basic TUI functionality

echo "=== Testing TUI Mode ==="

# Test 1: Check if TUI launches
echo "Test 1: Launching TUI mode..."
timeout 5s ./pass-cli.exe --help > /dev/null 2>&1
if [ $? -eq 124 ]; then
    echo "✓ TUI command is available"
else
    echo "✗ TUI command failed"
fi

# Test 2: Try explicit tui subcommand with timeout
echo "Test 2: Testing explicit tui subcommand..."
echo "TestMasterP@ss123" | timeout 3s ./pass-cli.exe tui > /dev/null 2>&1
if [ $? -eq 124 ]; then
    echo "✓ TUI subcommand launches (timed out as expected)"
else
    echo "⚠ TUI subcommand behavior needs investigation"
fi

# Test 3: Check if TUI is default behavior
echo "Test 3: Testing default TUI behavior (no arguments)..."
echo "TestMasterP@ss123" | timeout 3s ./pass-cli.exe > /dev/null 2>&1
if [ $? -eq 124 ]; then
    echo "✓ Default behavior is TUI (timed out as expected)"
else
    echo "⚠ Default behavior may not be TUI"
fi

echo ""
echo "=== Manual Testing Instructions ==="
echo "To manually test TUI mode:"
echo "1. Run: ./pass-cli.exe"
echo "2. Enter master password: TestMasterP@ss123"
echo "3. Test these keyboard shortcuts:"
echo "   - Ctrl+H: Toggle password visibility (in forms)"
echo "   - Ctrl+C: Quit application"
echo "   - ?: Show help modal"
echo "   - a: Add new credential"
echo "   - e: Edit credential"
echo "   - d: Delete credential"
echo "   - Tab: Navigate between components"
echo "   - Arrow keys: Navigate lists"
echo "   - Enter: Select items"
echo "   - Esc: Close modals/cancel"
echo ""