#!/bin/bash
# T008: Run All Validation Script
# Execute all 4 validation scripts in parallel and generate summary report
# Exit code 0 only if all scripts pass

set +e  # Don't exit on first error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================="
echo "Documentation Validation Suite"
echo "========================================="
echo ""

# Track individual script results
version_audit_exit=0
command_tests_exit=0
link_check_exit=0
cross_ref_exit=0

# Run all validation scripts
echo "Running all validation scripts..."
echo ""

# Make scripts executable
chmod +x "$SCRIPT_DIR/version-audit.sh" 2>/dev/null || true
chmod +x "$SCRIPT_DIR/command-tests.sh" 2>/dev/null || true
chmod +x "$SCRIPT_DIR/link-check.sh" 2>/dev/null || true
chmod +x "$SCRIPT_DIR/cross-reference-check.sh" 2>/dev/null || true

# Run scripts (capturing exit codes)
echo "1. Version Audit"
echo "---"
bash "$SCRIPT_DIR/version-audit.sh"
version_audit_exit=$?
echo ""

echo "2. Command Tests"
echo "---"
bash "$SCRIPT_DIR/command-tests.sh"
command_tests_exit=$?
echo ""

echo "3. Link Check"
echo "---"
bash "$SCRIPT_DIR/link-check.sh"
link_check_exit=$?
echo ""

echo "4. Cross-Reference Check"
echo "---"
bash "$SCRIPT_DIR/cross-reference-check.sh"
cross_ref_exit=$?
echo ""

# Generate summary report
echo "========================================="
echo "VALIDATION SUMMARY REPORT"
echo "========================================="
echo ""

total_failed=0

if [ $version_audit_exit -eq 0 ]; then
    echo "✅ Version Audit: PASS"
else
    echo "❌ Version Audit: FAIL"
    total_failed=$((total_failed + 1))
fi

if [ $command_tests_exit -eq 0 ]; then
    echo "✅ Command Tests: PASS"
else
    echo "❌ Command Tests: FAIL"
    total_failed=$((total_failed + 1))
fi

if [ $link_check_exit -eq 0 ]; then
    echo "✅ Link Check: PASS"
else
    echo "❌ Link Check: FAIL"
    total_failed=$((total_failed + 1))
fi

if [ $cross_ref_exit -eq 0 ]; then
    echo "✅ Cross-Reference Check: PASS"
else
    echo "❌ Cross-Reference Check: FAIL"
    total_failed=$((total_failed + 1))
fi

echo ""
echo "========================================="

if [ $total_failed -eq 0 ]; then
    echo "✅ ALL VALIDATION SCRIPTS PASSED"
    echo "   Documentation is ready for manual review"
    exit 0
else
    echo "❌ VALIDATION FAILED: $total_failed script(s) failed"
    echo "   Please review errors above and fix issues"
    echo "   Then re-run this validation suite"
    exit 1
fi
