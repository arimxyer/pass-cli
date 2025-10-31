#!/bin/bash
# T007: Cross-Reference Check Script
# Parse cross-references between files and verify referenced content exists
# Exit code 0 if all valid, 1 if broken refs

set -e

PROJECT_ROOT="R:/Test-Projects/pass-cli"
DOCS_DIR="$PROJECT_ROOT/docs"
README="$PROJECT_ROOT/README.md"

echo "🔍 Cross-Reference Check - Verifying internal document references..."
echo ""

issues_found=0

# Check SECURITY.md ↔ MIGRATION.md consistency (iteration counts)
echo "Checking SECURITY.md ↔ MIGRATION.md consistency..."
if [ -f "$DOCS_DIR/SECURITY.md" ] && [ -f "$DOCS_DIR/MIGRATION.md" ]; then
    security_600k=$(grep -c "600,000\|600k" "$DOCS_DIR/SECURITY.md" || echo "0")
    migration_600k=$(grep -c "600,000\|600k" "$DOCS_DIR/MIGRATION.md" || echo "0")

    if [ "$security_600k" -gt 0 ] && [ "$migration_600k" -gt 0 ]; then
        echo "✓ PASS: Both SECURITY.md and MIGRATION.md reference 600k iterations"
    else
        echo "❌ FAIL: Iteration count mismatch between SECURITY.md and MIGRATION.md"
        echo "   SECURITY.md: $security_600k references to 600k"
        echo "   MIGRATION.md: $migration_600k references to 600k"
        issues_found=$((issues_found + 1))
    fi
else
    echo "⚠️  WARNING: SECURITY.md or MIGRATION.md not found"
    issues_found=$((issues_found + 1))
fi
echo ""

# Check README.md → USAGE.md shortcut references
echo "Checking README.md → USAGE.md cross-references..."
if grep -q "docs/USAGE.md\|USAGE.md" "$README" 2>/dev/null; then
    if [ -f "$DOCS_DIR/USAGE.md" ]; then
        echo "✓ PASS: README.md references USAGE.md and file exists"
    else
        echo "❌ FAIL: README.md references USAGE.md but file not found"
        issues_found=$((issues_found + 1))
    fi
else
    echo "ℹ️  INFO: README.md does not reference USAGE.md"
fi
echo ""

# Check README.md → INSTALLATION.md references
echo "Checking README.md → INSTALLATION.md cross-references..."
if grep -q "docs/INSTALLATION.md\|INSTALLATION.md" "$README" 2>/dev/null; then
    if [ -f "$DOCS_DIR/INSTALLATION.md" ]; then
        echo "✓ PASS: README.md references INSTALLATION.md and file exists"
    else
        echo "❌ FAIL: README.md references INSTALLATION.md but file not found"
        issues_found=$((issues_found + 1))
    fi
else
    echo "ℹ️  INFO: README.md does not reference INSTALLATION.md"
fi
echo ""

# Check for "See <file>" patterns and verify files exist
echo "Checking 'See <file>' cross-references in all documentation..."
for doc in "$README" "$DOCS_DIR"/*.md; do
    if [ -f "$doc" ]; then
        filename=$(basename "$doc")

        # Extract "See FILENAME.md" patterns
        see_refs=$(grep -oh "See [A-Z_]*\.md" "$doc" 2>/dev/null || true)

        for ref in $see_refs; do
            ref_file=$(echo "$ref" | sed 's/See //')

            if [ -f "$DOCS_DIR/$ref_file" ] || [ -f "$PROJECT_ROOT/$ref_file" ]; then
                echo "  ✓ PASS: $filename references $ref_file (exists)"
            else
                echo "  ❌ FAIL: $filename references $ref_file (not found)"
                issues_found=$((issues_found + 1))
            fi
        done
    fi
done

echo ""
echo "========================================="
if [ $issues_found -eq 0 ]; then
    echo "✅ CROSS-REFERENCE CHECK PASS: All references valid"
    exit 0
else
    echo "❌ CROSS-REFERENCE CHECK FAIL: $issues_found broken reference(s)"
    echo "   Please fix or remove broken cross-references"
    exit 1
fi
