#!/bin/bash
# T006: Link Check Script
# Validate HTTP/HTTPS links and internal file links
# Exit code 0 if all links valid, 1 if broken links

set -e

PROJECT_ROOT="R:/Test-Projects/pass-cli"
DOCS_DIR="$PROJECT_ROOT/docs"
README="$PROJECT_ROOT/README.md"

echo "🔍 Link Check - Validating internal and external links..."
echo ""

issues_found=0

# Check if markdown-link-check is available
if command -v markdown-link-check >/dev/null 2>&1; then
    echo "Using markdown-link-check for validation..."
    echo ""

    # Check README.md
    if markdown-link-check "$README" --quiet; then
        echo "✓ PASS: README.md links valid"
    else
        echo "❌ FAIL: README.md has broken links"
        issues_found=$((issues_found + 1))
    fi

    # Check all docs/*.md files
    for file in "$DOCS_DIR"/*.md; do
        if [ -f "$file" ]; then
            filename=$(basename "$file")
            if markdown-link-check "$file" --quiet; then
                echo "✓ PASS: $filename links valid"
            else
                echo "❌ FAIL: $filename has broken links"
                issues_found=$((issues_found + 1))
            fi
        fi
    done
else
    echo "⚠️  WARNING: markdown-link-check not installed"
    echo "   Install with: npm install -g markdown-link-check"
    echo "   Falling back to basic internal link validation..."
    echo ""

    # Basic internal link validation
    # Look for [text](file.md) patterns and verify files exist
    for doc in "$README" "$DOCS_DIR"/*.md; do
        if [ -f "$doc" ]; then
            filename=$(basename "$doc")
            echo "Checking internal links in $filename..."

            # Extract markdown links to .md files
            internal_links=$(grep -oh "\[.*\]([^)]*\.md[^)]*)" "$doc" 2>/dev/null | sed 's/.*(\([^)]*\))/\1/' || true)

            for link in $internal_links; do
                # Remove anchors
                link_file=$(echo "$link" | sed 's/#.*//')

                # Check if file exists (relative to docs or project root)
                if [ -f "$DOCS_DIR/$link_file" ] || [ -f "$PROJECT_ROOT/$link_file" ]; then
                    echo "  ✓ PASS: $link_file exists"
                else
                    echo "  ❌ FAIL: $link_file not found"
                    issues_found=$((issues_found + 1))
                fi
            done
        fi
    done
fi

echo ""
echo "========================================="
if [ $issues_found -eq 0 ]; then
    echo "✅ LINK CHECK PASS: All links valid"
    exit 0
else
    echo "❌ LINK CHECK FAIL: $issues_found broken link(s)"
    echo "   Please fix or remove broken links"
    echo "   For 404 errors, consider archive.org fallback URLs (edge case E1)"
    exit 1
fi
