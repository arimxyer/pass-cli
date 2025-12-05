#!/usr/bin/env bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

# Parse arguments
JSON_MODE=false
DRY_RUN=false
LIST_MODE=false
COMPLETED_MODE=false
SPEC_NUMBERS=()

for arg in "$@"; do
    case "$arg" in
        --json|-j)
            JSON_MODE=true
            ;;
        --list|-l)
            LIST_MODE=true
            ;;
        --dry-run|-n)
            DRY_RUN=true
            ;;
        --completed|-c)
            COMPLETED_MODE=true
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS] [SPEC_NUMBERS...]"
            echo ""
            echo "Archive completed or old spec directories to specs/archive/"
            echo ""
            echo "Options:"
            echo "  --json, -j         Output in JSON format"
            echo "  --list, -l         List all specs with their status"
            echo "  --dry-run, -n      Show what would be archived without moving files"
            echo "  --completed, -c    Archive all completed specs"
            echo "  --help, -h         Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 --list                 # List all specs with status"
            echo "  $0 001 003                # Archive specs 001 and 003"
            echo "  $0 --dry-run 001 003      # Preview archiving 001 and 003"
            echo "  $0 --completed            # Archive all completed specs"
            exit 0
            ;;
        [0-9][0-9][0-9])
            SPEC_NUMBERS+=("$arg")
            ;;
        *)
            echo "Error: Unknown argument '$arg'" >&2
            echo "Use --help for usage information" >&2
            exit 1
            ;;
    esac
done

# Get repository paths
REPO_ROOT=$(get_repo_root)
SPECS_DIR="$REPO_ROOT/specs"
ARCHIVE_DIR="$SPECS_DIR/archive"

# Ensure specs directory exists
if [[ ! -d "$SPECS_DIR" ]]; then
    if $JSON_MODE; then
        echo '{"error":"specs directory not found"}'
    else
        echo "Error: specs directory not found at $SPECS_DIR" >&2
    fi
    exit 1
fi

# Function to check if a spec is completed by parsing tasks.md
is_spec_completed() {
    local spec_dir="$1"
    local tasks_file="$spec_dir/tasks.md"

    if [[ ! -f "$tasks_file" ]]; then
        # No tasks file means not completed
        return 1
    fi

    # Extract all task lines (lines that start with - [ ] or - [X])
    local total_tasks=$(grep -E '^\s*-\s+\[[Xx ]\]' "$tasks_file" 2>/dev/null | wc -l)
    local completed_tasks=$(grep -E '^\s*-\s+\[[Xx]\]' "$tasks_file" 2>/dev/null | wc -l)

    # Spec is completed if it has tasks and all are marked as complete
    if [[ $total_tasks -gt 0 && $completed_tasks -eq $total_tasks ]]; then
        return 0
    else
        return 1
    fi
}

# Function to get spec info
get_spec_info() {
    local spec_dir="$1"
    local spec_name=$(basename "$spec_dir")
    local spec_number=""
    local spec_slug=""
    local status="IN PROGRESS"

    # Extract number and slug from spec name (format: 001-spec-slug)
    if [[ "$spec_name" =~ ^([0-9]{3})-(.+)$ ]]; then
        spec_number="${BASH_REMATCH[1]}"
        spec_slug="${BASH_REMATCH[2]}"
    else
        spec_number="???"
        spec_slug="$spec_name"
    fi

    # Check if completed
    if is_spec_completed "$spec_dir"; then
        status="COMPLETED"
    fi

    if $JSON_MODE; then
        printf '{"number":"%s","slug":"%s","name":"%s","status":"%s","path":"%s"}' \
            "$spec_number" "$spec_slug" "$spec_name" "$status" "$spec_dir"
    else
        printf "%s - %s [%s]" "$spec_number" "$spec_slug" "$status"
    fi
}

# Function to list all specs
list_specs() {
    local specs=()

    if [[ -d "$SPECS_DIR" ]]; then
        for dir in "$SPECS_DIR"/*; do
            if [[ -d "$dir" && $(basename "$dir") != "archive" ]]; then
                specs+=("$dir")
            fi
        done
    fi

    if [[ ${#specs[@]} -eq 0 ]]; then
        if $JSON_MODE; then
            echo '{"specs":[]}'
        else
            echo "No specs found in $SPECS_DIR"
        fi
        return
    fi

    if $JSON_MODE; then
        echo -n '{"specs":['
        local first=true
        for spec_dir in "${specs[@]}"; do
            if $first; then
                first=false
            else
                echo -n ','
            fi
            get_spec_info "$spec_dir"
        done
        printf '],"archive_location":"%s"}\n' "$ARCHIVE_DIR"
    else
        echo "Available Specs:"
        echo "================"
        for spec_dir in "${specs[@]}"; do
            echo "  $(get_spec_info "$spec_dir")"
        done
        echo ""
        echo "Archive location: $ARCHIVE_DIR"
        echo ""
        echo "Usage:"
        echo "  - Archive specific specs: speckit.archive 001 003"
        echo "  - Archive all completed: speckit.archive --completed"
        echo "  - Dry run: speckit.archive --dry-run 001 002"
    fi
}

# Function to archive a spec
archive_spec() {
    local spec_number="$1"
    local source_dir=""

    # Find the spec directory
    for dir in "$SPECS_DIR"/"$spec_number"-*; do
        if [[ -d "$dir" ]]; then
            source_dir="$dir"
            break
        fi
    done

    if [[ -z "$source_dir" ]]; then
        if $JSON_MODE; then
            printf '{"number":"%s","status":"error","message":"Spec not found"}\n' "$spec_number"
        else
            echo "Error: Spec $spec_number not found" >&2
        fi
        return 1
    fi

    local spec_name=$(basename "$source_dir")
    local dest_dir="$ARCHIVE_DIR/$spec_name"

    if $DRY_RUN; then
        if $JSON_MODE; then
            printf '{"number":"%s","name":"%s","status":"dry-run","from":"%s","to":"%s"}\n' \
                "$spec_number" "$spec_name" "$source_dir" "$dest_dir"
        else
            echo "  Would archive: $spec_number - ${spec_name#$spec_number-} → archive/$spec_name"
        fi
        return 0
    fi

    # Create archive directory if it doesn't exist
    mkdir -p "$ARCHIVE_DIR"

    # Move the spec to archive (Windows-aware)
    if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" || "$OSTYPE" == "cygwin" ]]; then
        # Windows: use robocopy for reliable directory moving
        # Convert paths to Windows format
        local win_source=$(cygpath -w "$source_dir" 2>/dev/null || echo "$source_dir")
        local win_dest=$(cygpath -w "$dest_dir" 2>/dev/null || echo "$dest_dir")

        # robocopy: //E = copy subdirectories including empty, //MOVE = move (delete source after copy)
        # Double slashes prevent MSYS path conversion
        # //NFL = no file list, //NDL = no directory list, //NJH = no job header, //NJS = no job summary, //NC = no class, //NS = no size, //NP = no progress
        robocopy "$win_source" "$win_dest" //E //MOVE //NFL //NDL //NJH //NJS //NC //NS //NP > /dev/null 2>&1

        # robocopy exit codes: 0-7 are success (0=no files, 1=files copied, 2=extra files, etc.)
        local rc=$?
        if [[ $rc -gt 7 ]]; then
            if $JSON_MODE; then
                printf '{"number":"%s","status":"error","message":"Failed to move spec (robocopy error %d)"}\n' "$spec_number" "$rc"
            else
                echo "Error: Failed to archive spec $spec_number (robocopy error $rc)" >&2
            fi
            return 1
        fi

        # Remove source directory if still exists (robocopy /MOVE should remove it, but sometimes doesn't remove empty parent)
        if [[ -d "$source_dir" ]]; then
            rmdir "$source_dir" 2>/dev/null || true
        fi
    else
        # Unix: use standard mv
        mv "$source_dir" "$dest_dir"
        if [[ $? -ne 0 ]]; then
            if $JSON_MODE; then
                printf '{"number":"%s","status":"error","message":"Failed to move spec"}\n' "$spec_number"
            else
                echo "Error: Failed to archive spec $spec_number" >&2
            fi
            return 1
        fi
    fi

    if $JSON_MODE; then
        printf '{"number":"%s","name":"%s","status":"archived","from":"%s","to":"%s"}\n' \
            "$spec_number" "$spec_name" "$source_dir" "$dest_dir"
    else
        echo "  $spec_number - ${spec_name#$spec_number-} [ARCHIVED]"
    fi
}

# Main execution logic

# If no arguments or --list flag, show list
if [[ ${#SPEC_NUMBERS[@]} -eq 0 && "$COMPLETED_MODE" = false ]] || $LIST_MODE; then
    list_specs
    exit 0
fi

# Check if on main branch (only for git repos)
if has_git; then
    CURRENT_BRANCH=$(get_current_branch)
    if [[ "$CURRENT_BRANCH" != "main" ]]; then
        if ! $JSON_MODE; then
            echo "WARNING: You are on branch '$CURRENT_BRANCH', not 'main'" >&2
            echo "" >&2
            echo "Archiving specs on a feature branch can cause inconsistencies:" >&2
            echo "  - Archive state will differ between branches" >&2
            echo "  - Potential merge conflicts when updating specs/ directory" >&2
            echo "  - Confusion about which specs exist" >&2
            echo "" >&2
            echo "Recommended: Switch to main branch first" >&2
            echo "" >&2
            read -p "Continue anyway? (yes/no): " response
            if [[ ! "$response" =~ ^[Yy][Ee][Ss]$ ]]; then
                echo "Aborted."
                exit 0
            fi
        fi
    fi
fi

# Handle --completed mode
if $COMPLETED_MODE; then
    COMPLETED_SPECS=()

    for dir in "$SPECS_DIR"/*; do
        if [[ -d "$dir" && $(basename "$dir") != "archive" ]]; then
            if is_spec_completed "$dir"; then
                spec_name=$(basename "$dir")
                if [[ "$spec_name" =~ ^([0-9]{3})- ]]; then
                    COMPLETED_SPECS+=("${BASH_REMATCH[1]}")
                fi
            fi
        fi
    done

    if [[ ${#COMPLETED_SPECS[@]} -eq 0 ]]; then
        if $JSON_MODE; then
            echo '{"message":"No completed specs found"}'
        else
            echo "No completed specs found."
        fi
        exit 0
    fi

    # Show what will be archived and ask for confirmation
    if ! $JSON_MODE && ! $DRY_RUN; then
        echo "Found completed specs ready to archive:"
        for spec_num in "${COMPLETED_SPECS[@]}"; do
            for dir in "$SPECS_DIR"/"$spec_num"-*; do
                if [[ -d "$dir" ]]; then
                    spec_name=$(basename "$dir")
                    echo "  - $spec_num - ${spec_name#$spec_num-}"
                    break
                fi
            done
        done
        echo ""
        read -p "Archive these specs? (yes/no): " response
        if [[ ! "$response" =~ ^[Yy][Ee][Ss]$ ]]; then
            echo "Aborted."
            exit 0
        fi
    fi

    SPEC_NUMBERS=("${COMPLETED_SPECS[@]}")
fi

# Archive the specified specs
if $JSON_MODE; then
    echo -n '{"archived":['
    first=true
    for spec_num in "${SPEC_NUMBERS[@]}"; do
        if $first; then
            first=false
        else
            echo -n ','
        fi
        archive_spec "$spec_num"
    done
    echo ']}'
else
    if $DRY_RUN; then
        echo "DRY RUN - No specs will be moved"
        echo "================================="
        echo "Would archive:"
    else
        echo "Archiving specs: ${SPEC_NUMBERS[*]}"
        echo ""
        echo "Archive Complete:"
        echo "================="
    fi

    for spec_num in "${SPEC_NUMBERS[@]}"; do
        archive_spec "$spec_num"
    done

    if $DRY_RUN; then
        echo ""
        echo "To perform archive, run: speckit.archive ${SPEC_NUMBERS[*]}"
    else
        echo ""
        echo "Archived to: $ARCHIVE_DIR"
    fi
fi
