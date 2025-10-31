# Data Model: Documentation Review and Production Release Preparation

**Feature**: Documentation Review and Production Release Preparation
**Branch**: `008-review-and-update`
**Date**: 2025-01-14

## Overview

This data model defines the entities, attributes, and validation rules for tracking documentation accuracy during the production release preparation process. Since this is a documentation-only feature, the "data model" represents the validation tracking structure rather than application data.

## Entity: Documentation File Validation Record

### Purpose
Track validation status, issues, and acceptance criteria for each documentation file during the review process.

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_path` | String (absolute path) | Yes | Absolute path to documentation file (e.g., `R:\Test-Projects\pass-cli\docs\SECURITY.md`) |
| `validation_status` | Enum | Yes | Current validation state: `PENDING`, `IN_PROGRESS`, `PASS`, `FAIL` |
| `issues_found` | List<ValidationIssue> | Yes | List of validation failures discovered during review (empty for PASS status) |
| `last_updated` | Timestamp | Yes | ISO 8601 timestamp of last file modification |
| `validated_by` | Enum | Yes | Validation method: `MANUAL_REVIEW`, `AUTOMATED_SCRIPT`, `COMMAND_EXECUTION`, `LINK_CHECK` |
| `acceptance_criteria` | List<String> | Yes | List of applicable functional requirements (e.g., `["FR-001", "FR-003", "FR-004"]`) |
| `priority` | Enum | Yes | Review priority based on user story: `P1`, `P2`, `P3` |
| `estimated_review_time` | Integer (minutes) | No | Estimated time to complete validation (for planning) |
| `actual_review_time` | Integer (minutes) | No | Actual time spent on validation (for metrics) |

### Validation Rules

1. **Status Transitions**:
   - `PENDING` → `IN_PROGRESS` (when review starts)
   - `IN_PROGRESS` → `PASS` (when all issues resolved and acceptance criteria met)
   - `IN_PROGRESS` → `FAIL` (when issues found)
   - `FAIL` → `IN_PROGRESS` (when fixes applied and re-validation starts)

2. **Completion Criteria**:
   - `validation_status = PASS` requires `issues_found = []` (empty list)
   - All 7 documentation files must have `validation_status = PASS` for feature completion
   - Each file must map to at least one functional requirement in `acceptance_criteria`

3. **Issue Severity**:
   - CRITICAL: Command execution failures, incorrect version references, missing security specs
   - HIGH: Broken links, outdated feature references, incomplete troubleshooting
   - MEDIUM: Formatting inconsistencies, minor inaccuracies
   - LOW: Typos, style inconsistencies

## Entity: Validation Issue

### Purpose
Represent a specific problem discovered during documentation validation.

### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `issue_id` | String | Yes | Unique identifier (e.g., `README-001`, `SECURITY-042`) |
| `severity` | Enum | Yes | Issue severity: `CRITICAL`, `HIGH`, `MEDIUM`, `LOW` |
| `category` | Enum | Yes | Issue type: `VERSION_REFERENCE`, `COMMAND_ERROR`, `BROKEN_LINK`, `MISSING_CONTENT`, `INACCURACY`, `FORMATTING` |
| `description` | String | Yes | Human-readable description of the issue |
| `location` | String | Yes | Line number or section heading where issue found (e.g., `Line 42`, `## Installation section`) |
| `expected_value` | String | No | What the documentation should contain (e.g., `600,000 iterations`) |
| `actual_value` | String | No | What the documentation currently contains (e.g., `100,000 iterations`) |
| `resolution_notes` | String | No | Description of how the issue was fixed |
| `resolved` | Boolean | Yes | Whether the issue has been fixed (default: `false`) |
| `related_fr` | List<String> | Yes | Functional requirements violated by this issue (e.g., `["FR-001", "FR-013"]`) |

### Validation Rules

1. **Severity Assignment**:
   - CRITICAL: Blocks production release (security specs missing, commands fail)
   - HIGH: Significantly impacts user experience (broken install instructions)
   - MEDIUM: Minor user impact (outdated examples)
   - LOW: Cosmetic issues (formatting, typos)

2. **Resolution Tracking**:
   - Issue cannot be marked `resolved = true` without `resolution_notes`
   - All CRITICAL and HIGH issues must be resolved before file status = `PASS`
   - MEDIUM and LOW issues can be deferred with justification

## Relationships

### Documentation File → Functional Requirements (Many-to-Many)

Each documentation file validates multiple functional requirements, and each functional requirement may be validated across multiple files.

**Examples**:
- `README.md` → `[FR-001, FR-002, FR-015]`
- `SECURITY.md` → `[FR-001, FR-003, FR-004, FR-005, FR-013, FR-014]`
- `FR-001` (version accuracy) → `[README.md, SECURITY.md, USAGE.md, INSTALLATION.md, TROUBLESHOOTING.md, MIGRATION.md, KNOWN_LIMITATIONS.md]`

### Documentation File → User Stories (Many-to-Many, Traceability)

Documentation files support specific user stories through their acceptance scenarios.

**Examples**:
- `README.md` → User Story 1 (New User Onboarding Success - P1)
- `SECURITY.md` → User Story 2 (Production Deployment Confidence - P1)
- `USAGE.md` → User Story 3 (CLI vs TUI Mode Clarity - P2)
- `TROUBLESHOOTING.md` → User Story 5 (Troubleshooting Self-Service - P3)

### Documentation File → CLI Commands Referenced (One-to-Many)

Each documentation file references zero or more CLI commands that must be validated.

**Examples**:
- `USAGE.md` → `["pass-cli init", "pass-cli add", "pass-cli get", "pass-cli list", "pass-cli update", "pass-cli delete", "pass-cli generate", "pass-cli version"]`
- `README.md` → `["pass-cli init", "pass-cli add github", "pass-cli get github", "pass-cli get github --copy"]`
- `INSTALLATION.md` → `["brew tap ari1110/homebrew-tap", "brew install pass-cli", "scoop bucket add pass-cli ...", "scoop install pass-cli"]`

### Validation Issue → Documentation File (Many-to-One)

Each validation issue belongs to exactly one documentation file.

## Data Instances (Examples)

### Example 1: README.md Validation Record

```json
{
  "file_path": "R:\\Test-Projects\\pass-cli\\README.md",
  "validation_status": "IN_PROGRESS",
  "issues_found": [
    {
      "issue_id": "README-001",
      "severity": "HIGH",
      "category": "MISSING_CONTENT",
      "description": "TUI keyboard shortcuts incomplete (only 6 shown, should be 20+)",
      "location": "Line 98-106 (## Key TUI Shortcuts table)",
      "expected_value": "Table with 20+ keyboard shortcuts including custom keybindings",
      "actual_value": "Table with only 6 shortcuts (Ctrl+H, /, s, i, ?, q)",
      "resolution_notes": null,
      "resolved": false,
      "related_fr": ["FR-007"]
    },
    {
      "issue_id": "README-002",
      "severity": "MEDIUM",
      "category": "MISSING_CONTENT",
      "description": "No documentation for config.yml feature (spec 007)",
      "location": "## Advanced Usage section",
      "expected_value": "Section documenting config.yml location, format, and keybinding customization",
      "actual_value": "Section missing entirely",
      "resolution_notes": null,
      "resolved": false,
      "related_fr": ["FR-007", "FR-010"]
    }
  ],
  "last_updated": "2025-01-14T10:30:00Z",
  "validated_by": "MANUAL_REVIEW",
  "acceptance_criteria": ["FR-001", "FR-002", "FR-015"],
  "priority": "P1",
  "estimated_review_time": 30,
  "actual_review_time": null
}
```

### Example 2: SECURITY.md Validation Record (PASS)

```json
{
  "file_path": "R:\\Test-Projects\\pass-cli\\docs\\SECURITY.md",
  "validation_status": "PASS",
  "issues_found": [],
  "last_updated": "2025-01-13T14:22:00Z",
  "validated_by": "MANUAL_REVIEW",
  "acceptance_criteria": ["FR-001", "FR-003", "FR-004", "FR-005", "FR-013", "FR-014"],
  "priority": "P1",
  "estimated_review_time": 45,
  "actual_review_time": 52
}
```

### Example 3: USAGE.md Validation Record (FAIL)

```json
{
  "file_path": "R:\\Test-Projects\\pass-cli\\docs\\USAGE.md",
  "validation_status": "FAIL",
  "issues_found": [
    {
      "issue_id": "USAGE-001",
      "severity": "CRITICAL",
      "category": "COMMAND_ERROR",
      "description": "Documented command flag does not exist in current release",
      "location": "Line 234 (## generate - Generate Password section)",
      "expected_value": "Flag should exist and execute successfully",
      "actual_value": "Command 'pass-cli generate --custom-charset abc123' returns error: unknown flag",
      "resolution_notes": null,
      "resolved": false,
      "related_fr": ["FR-011"]
    },
    {
      "issue_id": "USAGE-002",
      "severity": "HIGH",
      "category": "INACCURACY",
      "description": "Toggle detail panel shortcut outdated",
      "location": "Line 859 (TUI Keyboard Shortcuts table)",
      "expected_value": "'i' key to toggle detail panel",
      "actual_value": "Documentation shows 'Tab' key",
      "resolution_notes": null,
      "resolved": false,
      "related_fr": ["FR-007"]
    }
  ],
  "last_updated": "2025-01-10T09:15:00Z",
  "validated_by": "COMMAND_EXECUTION",
  "acceptance_criteria": ["FR-001", "FR-006", "FR-007", "FR-011", "FR-012"],
  "priority": "P1",
  "estimated_review_time": 60,
  "actual_review_time": null
}
```

## State Transitions

```
PENDING
   ↓ (Review starts)
IN_PROGRESS
   ↓ (Issues found)
FAIL
   ↓ (Fixes applied, re-validation)
IN_PROGRESS
   ↓ (All issues resolved)
PASS
```

## Validation Workflow

1. **Initialize**: Create validation record for each of 7 documentation files with `status = PENDING`
2. **Prioritize**: Review P1 files first (README, SECURITY, USAGE), then P2 (INSTALLATION, TROUBLESHOOTING), then P3 (MIGRATION, KNOWN_LIMITATIONS)
3. **Validate**: For each file:
   - Set `status = IN_PROGRESS`
   - Run applicable validation methods (manual review, command execution, link check)
   - Record issues in `issues_found` list
   - Set `status = FAIL` if issues found, `status = PASS` if no issues
4. **Fix**: For each FAIL status:
   - Update documentation file to resolve issues
   - Mark issues as `resolved = true` with `resolution_notes`
   - Re-validate (return to step 3)
5. **Complete**: When all 7 files have `status = PASS`, documentation review is complete

## Metrics Tracked

- **Total issues found**: Count across all validation records
- **Issues by severity**: Count of CRITICAL, HIGH, MEDIUM, LOW
- **Issues by category**: Count of VERSION_REFERENCE, COMMAND_ERROR, BROKEN_LINK, etc.
- **Time per file**: `actual_review_time` for efficiency analysis
- **Pass rate**: Percentage of files passing on first validation
- **Re-validation cycles**: Number of FAIL → IN_PROGRESS → PASS cycles per file

## Success Criteria Mapping

| Success Criterion | Validation Method | Data Model Field |
|-------------------|-------------------|------------------|
| SC-001: 10-minute onboarding | Manual walkthrough | README.md issues related to FR-002 |
| SC-002: Crypto params identifiable | Manual review | SECURITY.md issues related to FR-003, FR-004, FR-005 |
| SC-003: 100% command accuracy | Command execution | `category = COMMAND_ERROR` issues count must be 0 |
| SC-004: 5-minute troubleshooting | Manual search | TROUBLESHOOTING.md issues related to FR-009 |
| SC-005: Installation methods work | Platform testing | INSTALLATION.md issues related to FR-008 |
| SC-006: 20+ shortcuts documented | Shortcut verification | USAGE.md issues related to FR-007 |
| SC-007: Zero outdated references | Automated grep | `category = VERSION_REFERENCE` issues count must be 0 |
| SC-008: Script examples run | Script execution | `category = COMMAND_ERROR` issues in script examples |

---

**Note**: This data model is conceptual for tracking purposes. Implementation may use spreadsheet, JSON files, or issue tracking system rather than formal database.
