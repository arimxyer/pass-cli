#!/bin/bash
# T004: Version Audit Script
# Grep all docs for version references and flag inconsistencies
# Exit code 0 if clean, 1 if issues found

set -e

PROJECT_ROOT="R:/Test-Projects/pass-cli"
DOCS_DIR="$PROJECT_ROOT/docs"
README="$PROJECT_ROOT/README.md"

echo "🔍 Version Audit - Checking documentation for version consistency..."
echo ""

issues_found=0

# Check for outdated PBKDF2 iteration counts
echo "Checking for outdated iteration counts (100,000 or 100k)..."
if grep -rn "100,000\|100k" "$README" "$DOCS_DIR"/*.md 2>/dev/null | grep -v "upgrade from\|migration\|backward"; then
    echo "❌ FAIL: Found outdated iteration count references (100k instead of 600k)"
    echo "   Files should reference 600,000 or 600k iterations for PBKDF2-SHA256"
    issues_found=$((issues_found + 1))
else
    echo "✓ PASS: No outdated iteration count references found"
fi
echo ""

# Check for version inconsistencies
echo "Checking for version references..."
version_refs=$(grep -rho "v[0-9]\+\.[0-9]\+\.[0-9]\+" "$README" "$DOCS_DIR"/*.md 2>/dev/null | sort | uniq)
version_count=$(echo "$version_refs" | wc -l)

if [ "$version_count" -gt 1 ]; then
    echo "❌ FAIL: Multiple version references found"
    echo "   Versions found: $version_refs"
    echo "   All documentation should reference the same current release version"
    issues_found=$((issues_found + 1))
else
    echo "✓ PASS: Version references consistent (found: $version_refs)"
fi
echo ""

# Check for outdated terminal size references (60×20 instead of 60×30)
echo "Checking for outdated terminal size references (60×20 instead of 60×30)..."
if grep -rn "60[×x]20\|60 x 20" "$README" "$DOCS_DIR"/*.md 2>/dev/null; then
    echo "❌ FAIL: Found outdated terminal size references (60×20)"
    echo "   Minimum terminal size should be 60×30 (spec 006)"
    issues_found=$((issues_found + 1))
else
    echo "✓ PASS: No outdated terminal size references found"
fi
echo ""

# Check for 'tab' keybinding for toggle_detail (changed to 'i' in spec 007)
echo "Checking for outdated toggle_detail keybinding references..."
if grep -rni "tab.*toggle.*detail\|toggle.*detail.*tab" "$README" "$DOCS_DIR"/*.md 2>/dev/null | grep -v "Shift+Tab"; then
    echo "❌ FAIL: Found outdated toggle_detail keybinding reference ('tab')"
    echo "   toggle_detail keybinding changed from 'tab' to 'i' in spec 007"
    issues_found=$((issues_found + 1))
else
    echo "✓ PASS: No outdated toggle_detail keybinding references found"
fi
echo ""

# Summary
echo "========================================="
if [ $issues_found -eq 0 ]; then
    echo "✅ VERSION AUDIT PASS: No issues found"
    exit 0
else
    echo "❌ VERSION AUDIT FAIL: $issues_found issue(s) found"
    echo "   Please review and update documentation files"
    exit 1
fi
