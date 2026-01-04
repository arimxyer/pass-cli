#!/bin/bash
# Cleanup script to remove pass-cli test keyring entries
# These are left over from integration tests that didn't clean up properly

set -e

echo "=== Pass-CLI Test Keyring Cleanup ==="
echo ""

# Count entries before cleanup
BEFORE=$(secret-tool search --all service pass-cli 2>&1 | grep -c "attribute.username" || echo 0)
echo "Found $BEFORE total pass-cli entries"

# Delete all vault-test entries (test pollution)
echo ""
echo "Deleting test entries (vault-test-*)..."
DELETED=0

for username in $(secret-tool search --all service pass-cli 2>&1 | grep "attribute.username = master-password-vault-test" | sed 's/attribute.username = //'); do
    if secret-tool clear service pass-cli username "$username" 2>/dev/null; then
        echo "  ✓ Deleted: $username"
        ((DELETED++)) || true
    else
        echo "  ✗ Failed: $username"
    fi
done

# Delete test audit entries
echo ""
echo "Deleting test audit entries..."
for username in "test-vault-id" "001"; do
    if secret-tool clear service pass-cli-audit username "$username" 2>/dev/null; then
        echo "  ✓ Deleted audit: $username"
        ((DELETED++)) || true
    fi
done

# Count entries after cleanup
AFTER=$(secret-tool search --all service pass-cli 2>&1 | grep -c "attribute.username" || echo 0)

echo ""
echo "=== Summary ==="
echo "Deleted: $DELETED entries"
echo "Remaining pass-cli entries: $AFTER"
echo ""
echo "Your real entries should remain:"
echo "  - master-password-_pass-cli (your vault password)"
echo "  - .pass-cli in pass-cli-audit (your audit key)"
