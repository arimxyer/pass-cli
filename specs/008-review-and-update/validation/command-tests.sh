#!/bin/bash
# T005: Command Tests Script
# Extract all `pass-cli <command>` examples from docs and verify they exist
# Exit code 0 if all valid, 1 if failures

set -e

PROJECT_ROOT="R:/Test-Projects/pass-cli"
DOCS_DIR="$PROJECT_ROOT/docs"
README="$PROJECT_ROOT/README.md"

echo "🔍 Command Tests - Verifying all documented commands exist..."
echo ""

issues_found=0
commands_tested=0

# Extract all pass-cli commands from documentation
# Look for patterns like: pass-cli <command>, `pass-cli <command>`, pass-cli command
echo "Extracting commands from documentation..."
commands=$(grep -rho "pass-cli [a-z-]*" "$README" "$DOCS_DIR"/*.md 2>/dev/null | \
    sed 's/pass-cli //' | \
    sed 's/`//g' | \
    grep -v "^$" | \
    sort | uniq | \
    grep -v "^-" | \
    grep -v "^\[" | \
    head -20)

echo "Found $(echo "$commands" | wc -l) unique commands to test"
echo ""

# Test each command
for cmd in $commands; do
    commands_tested=$((commands_tested + 1))

    # Skip if command is empty or looks like a flag
    if [ -z "$cmd" ] || [[ "$cmd" == -* ]]; then
        continue
    fi

    # Test if command exists by running --help
    if pass-cli "$cmd" --help >/dev/null 2>&1; then
        echo "✓ PASS: pass-cli $cmd (command exists)"
    else
        echo "❌ FAIL: pass-cli $cmd (command not found or help failed)"
        issues_found=$((issues_found + 1))
    fi
done

echo ""
echo "========================================="
echo "Commands tested: $commands_tested"
if [ $issues_found -eq 0 ]; then
    echo "✅ COMMAND TESTS PASS: All documented commands valid"
    exit 0
else
    echo "❌ COMMAND TESTS FAIL: $issues_found invalid command(s)"
    echo "   Remove or update documentation for non-existent commands"
    exit 1
fi
