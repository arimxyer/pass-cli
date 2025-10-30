#!/usr/bin/env bash
set -euo pipefail

# update-version.sh - Update version numbers and dates across documentation
#
# Usage: ./scripts/update-version.sh <new-version>
# Example: ./scripts/update-version.sh v0.1.0

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if version argument provided
if [ $# -eq 0 ]; then
    echo -e "${RED}Error: Version argument required${NC}"
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.1.0"
    exit 1
fi

NEW_VERSION="$1"
# Strip 'v' prefix if present for certain contexts
VERSION_NO_V="${NEW_VERSION#v}"

# Get current date in "Month YYYY" format (e.g., "January 2025")
CURRENT_DATE=$(date +"%B %Y")

echo -e "${BLUE}=== Pass-CLI Version Update Tool ===${NC}"
echo -e "New version: ${GREEN}${NEW_VERSION}${NC}"
echo -e "Update date: ${GREEN}${CURRENT_DATE}${NC}"
echo ""

# Check if we're in the project root
if [ ! -f "go.mod" ] || [ ! -d "docs" ]; then
    echo -e "${RED}Error: Must run from project root directory${NC}"
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD -- 2>/dev/null; then
    echo -e "${YELLOW}Warning: You have uncommitted changes${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo -e "${BLUE}Updating documentation files...${NC}"

# Counter for changes
UPDATED_FILES=0

# Function to update a file
update_file() {
    local file="$1"
    local description="$2"

    if [ ! -f "$file" ]; then
        echo -e "${YELLOW}  Skipping $file (not found)${NC}"
        return
    fi

    local temp_file="${file}.tmp"
    local changes=0

    # Update "Documentation Version: vX.X.X"
    if grep -q "Documentation Version:" "$file" 2>/dev/null; then
        sed "s/\*\*Documentation Version\*\*: v[0-9]\+\.[0-9]\+\.[0-9]\+/**Documentation Version**: ${NEW_VERSION}/" "$file" > "$temp_file"
        if ! cmp -s "$file" "$temp_file"; then
            mv "$temp_file" "$file"
            ((changes++))
        else
            rm "$temp_file"
        fi
    fi

    # Update "Last Updated: Month YYYY"
    if grep -q "Last Updated:" "$file" 2>/dev/null; then
        sed "s/\*\*Last Updated\*\*: [A-Za-z]\+ [0-9]\+/**Last Updated**: ${CURRENT_DATE}/" "$file" > "$temp_file"
        if ! cmp -s "$file" "$temp_file"; then
            mv "$temp_file" "$file"
            ((changes++))
        else
            rm "$temp_file"
        fi
    fi

    if [ $changes -gt 0 ]; then
        echo -e "  ${GREEN}✓${NC} $description"
        ((UPDATED_FILES++))
    fi
}

# Update all documentation files with version footers
update_file "docs/USAGE.md" "USAGE.md"
update_file "docs/SECURITY.md" "SECURITY.md"
update_file "docs/MIGRATION.md" "MIGRATION.md"
update_file "docs/TROUBLESHOOTING.md" "TROUBLESHOOTING.md"
update_file "docs/KNOWN_LIMITATIONS.md" "KNOWN_LIMITATIONS.md"
update_file "docs/GETTING_STARTED.md" "GETTING_STARTED.md"
update_file "docs/INSTALLATION.md" "INSTALLATION.md"
update_file "docs/DOCTOR_COMMAND.md" "DOCTOR_COMMAND.md"

echo ""
echo -e "${BLUE}Updating package manifests...${NC}"

# Update Homebrew formula
if [ -f "homebrew/pass-cli.rb" ]; then
    sed -i.bak "s/version \"[0-9]\+\.[0-9]\+\.[0-9]\+\"/version \"${VERSION_NO_V}\"/" homebrew/pass-cli.rb
    rm homebrew/pass-cli.rb.bak
    echo -e "  ${GREEN}✓${NC} homebrew/pass-cli.rb"
    ((UPDATED_FILES++))
fi

# Update Scoop manifest
if [ -f "scoop/pass-cli.json" ]; then
    sed -i.bak "s/\"version\": \"[0-9]\+\.[0-9]\+\.[0-9]\+\"/\"version\": \"${VERSION_NO_V}\"/" scoop/pass-cli.json
    rm scoop/pass-cli.json.bak
    echo -e "  ${GREEN}✓${NC} scoop/pass-cli.json"
    ((UPDATED_FILES++))
fi

echo ""
echo -e "${GREEN}=== Update Complete ===${NC}"
echo -e "Updated ${UPDATED_FILES} files"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Review changes: git diff"
echo "  2. Update CHANGELOG.md with release notes"
echo "  3. Commit changes: git add -A && git commit -m \"chore: bump version to ${NEW_VERSION}\""
echo "  4. Create tag: git tag -a ${NEW_VERSION} -m \"Release ${NEW_VERSION}\""
echo "  5. Push: git push origin main --tags"
echo ""
echo -e "${BLUE}Note:${NC} Package manifest URLs and checksums will be updated automatically by GoReleaser"
