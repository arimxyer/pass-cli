# Audit Checklist: Documentation Review

**Feature**: Documentation Cleanup and Archival
**Phase**: 1 (Design)
**Date**: 2025-10-15

## Purpose

Systematic checklist for auditing existing Pass-CLI documentation to identify obsolete content, duplicates, and broken links. Use this checklist during implementation to ensure comprehensive coverage.

## Pre-Audit Setup

**Before starting audit, complete these preparatory steps**:

- [ ] Policy document created and committed (`docs/DOCUMENTATION_LIFECYCLE.md`)
- [ ] Baseline metrics captured:
  - Total doc file count: `find docs/ specs/ -name "*.md" | wc -l`
  - Total line count: `find docs/ specs/ -name "*.md" -exec wc -l {} + | tail -1`
  - Spec count: `ls -d specs/*/ | wc -l`

**Tooling Verification**:
- [ ] `ripgrep` available (`rg --version`)
- [ ] Git status clean (commit work-in-progress before audit)
- [ ] Current branch: `009-documentation-cleanup-and`

---

## Phase 1: docs/archive/ Directory Audit

**Objective**: Review all files in `docs/archive/` to determine if truly obsolete or should remain

### Step 1.1: Inventory Archive Contents

```bash
ls -lh docs/archive/
```

- [ ] List all files in docs/archive/ with last-modified dates
- [ ] Total files in archive: ______

### Step 1.2: Per-File Evaluation

For EACH file in `docs/archive/`, evaluate against retention rules:

**File**: `docs/archive/<filename>`
- [ ] Document type: [Spec / User Guide / API Doc / Example / Troubleshooting / Migration / Other]
- [ ] Last meaningful update: [date from git log]
- [ ] Archival trigger: [Feature removed / Content superseded / Duplicate / Link rot / Never finalized]
- [ ] Unique information not captured elsewhere: [Yes/No - if Yes, list what]
- [ ] Decision: [Keep / Delete / Consolidate into <target>]
- [ ] Rationale: [1-2 sentences explaining decision]

**Commit after each deletion**:
```bash
git rm docs/archive/<filename>
git commit -m "docs: archive <filename> - <rationale>"
```

### Step 1.3: Archive Summary

- [ ] Files reviewed: ______
- [ ] Files kept: ______ (with documented justification)
- [ ] Files deleted: ______
- [ ] Files consolidated: ______ (merged into: ______)

---

## Phase 2: specs/ Directory Audit

**Objective**: Identify old spec artifacts and process per retention policy (indefinite for specs unless created in error)

### Step 2.1: Inventory All Specs

```bash
ls -d specs/*/
```

- [ ] Total spec directories: ______
- [ ] List all spec numbers and feature names

### Step 2.2: Categorize Specs by Feature Status

For EACH spec directory, determine feature status:

**Spec**: `specs/<###>-<feature-name>/`
- [ ] Feature status: [Active / Deprecated / Removed / Never Implemented / Created in Error]
- [ ] Evidence: [Reference to code, PR, or commit]
- [ ] Retention decision: [Keep (default) / Delete (only if created in error)]
- [ ] Rationale: [if deleting, explain why exception applies]

**Note**: Default is KEEP per FR-008. Only delete if spec was created in error (wrong repo, duplicate, etc.)

### Step 2.3: Specs Summary

- [ ] Specs reviewed: ______
- [ ] Specs for active features: ______ (keep)
- [ ] Specs for removed features: ______ (keep per retention rule)
- [ ] Specs created in error: ______ (delete with justification)

---

## Phase 3: Duplicate Content Detection

**Objective**: Identify and consolidate overlapping documentation

### Step 3.1: Keyword-Based Duplicate Search

Run keyword searches for common topics:

**Installation Documentation**:
```bash
rg -i "installation" docs/ --files-with-matches
```
- [ ] Files found: ______
- [ ] Review for duplication: [List files]
- [ ] Consolidation decision: [Keep all / Merge into <canonical> / Cross-reference only]

**Usage/Getting Started**:
```bash
rg -i "getting started|usage|quick start" docs/ --files-with-matches
```
- [ ] Files found: ______
- [ ] Consolidation decision: ______

**Configuration**:
```bash
rg -i "configuration|config|settings" docs/ --files-with-matches
```
- [ ] Files found: ______
- [ ] Consolidation decision: ______

**Troubleshooting**:
```bash
rg -i "troubleshoot|common issues|faq" docs/ --files-with-matches
```
- [ ] Files found: ______
- [ ] Consolidation decision: ______

**Security**:
```bash
rg -i "security|encryption|vault" docs/ --files-with-matches
```
- [ ] Files found: ______
- [ ] Consolidation decision: ______

### Step 3.2: Section Header Analysis

Check for overlapping section structures:

```bash
rg "^##\s+" docs/ -N | sort | uniq -c | sort -rn | head -20
```

- [ ] Common H2 headers identified: ______
- [ ] Files with identical section structures (likely duplicates): ______

### Step 3.3: Manual Review of Flagged Duplicates

For each duplicate content area:

**Topic**: ______
**Files**: ______
- [ ] Read all files to confirm true duplication
- [ ] Select canonical source: ______ (most comprehensive)
- [ ] Unique content in duplicates: [List what needs merging]
- [ ] Consolidation action: [Merge + delete / Cross-reference only / Split by audience]

**Commit after consolidation**:
```bash
git rm <duplicate-files>
git commit -m "docs: consolidate <topic> into <canonical> - reduces duplicate maintenance"
```

### Step 3.4: Duplication Summary

- [ ] Topic areas searched: ______
- [ ] Duplicate sets found: ______
- [ ] Files consolidated: ______
- [ ] Canonical sources: ______

---

## Phase 4: Link Validation

**Objective**: Identify and fix all broken internal links

### Step 4.1: Manual Link Check (Per-File)

For each documentation file, manually verify links:

**File**: ______
- [ ] Total links in file: ______
- [ ] Broken internal links: ______ (fix with correct paths)
- [ ] Broken external links: ______ (update or remove)
- [ ] Links to deleted files: ______ (remove or update to replacement)

**Fix Process**:
1. Test each link by navigating path or checking target file exists
2. Update broken links with correct paths
3. Remove dead links if target permanently gone
4. Commit fixes: `git commit -m "docs: fix broken links in <file>"`

### Step 4.2: Cross-Reference Validation

After deletions and consolidations, verify links to affected files:

```bash
rg "<deleted-filename>" docs/ specs/ README.md CONTRIBUTING.md
```

- [ ] References to deleted files: ______ (update to new canonical source or remove)
- [ ] References to consolidated files: ______ (update to canonical location)

### Step 4.3: Link Validation Summary

- [ ] Total broken links found: ______
- [ ] Links fixed: ______
- [ ] Links removed (dead): ______
- [ ] Files with link updates: ______

---

## Phase 5: CONTRIBUTING.md Integration

**Objective**: Add policy summary to CONTRIBUTING.md per FR-010 and Clarification #2

### Step 5.1: Locate Insertion Point

- [ ] Read CONTRIBUTING.md to find appropriate section
- [ ] Suggested location: [After "How to Contribute" / New "Documentation Governance" section / Other: ______]

### Step 5.2: Draft Summary Text

**Inline Summary** (2-3 sentences):
```markdown
## Documentation Governance

Pass-CLI maintains a [Documentation Lifecycle Policy](docs/DOCUMENTATION_LIFECYCLE.md)
that defines retention periods, archival triggers, and decision workflows for all
repository documentation. Contributors should consult the policy when adding new
documentation or proposing changes to existing docs. The policy ensures documentation
remains current and maintainable while preserving historical design context.
```

### Step 5.3: Integration

- [ ] Add summary text to CONTRIBUTING.md
- [ ] Verify link to `docs/DOCUMENTATION_LIFECYCLE.md` works
- [ ] Commit: `git commit -m "docs: integrate lifecycle policy into CONTRIBUTING.md"`

---

## Phase 6: Final Validation

**Objective**: Verify all success criteria met

### Step 6.1: Success Criteria Checklist

- [ ] **SC-001**: Measure documentation maintenance effort reduction (baseline vs. post-cleanup file count)
  - Before: ______ files
  - After: ______ files
  - Reduction: ______%

- [ ] **SC-002**: Zero broken internal links
  - Manual verification: All links tested
  - Automated check: `rg "\]\(" docs/ specs/ | <manual verification>`

- [ ] **SC-003**: Policy document exists at `docs/DOCUMENTATION_LIFECYCLE.md` with CONTRIBUTING.md integration
  - [ ] File exists
  - [ ] CONTRIBUTING.md has summary + link

- [ ] **SC-004**: 50%+ reduction in duplicate/overlapping documentation
  - Duplicate topic areas before: ______
  - After consolidation: ______
  - Reduction: ______%

- [ ] **SC-005**: 100% of deletions documented in commit messages
  - [ ] Review `git log --grep="archive" --oneline`
  - [ ] All commits have rationale

- [ ] **SC-006**: 40% improvement in documentation search clarity
  - Metric: Topic areas covered by multiple documents
  - Before: ______
  - After: ______
  - Improvement: ______%

- [ ] **SC-007**: Zero historical context loss
  - [ ] All deleted files accessible via git history
  - [ ] Specs retained per policy (indefinite)
  - [ ] Unique content consolidated before deletion

### Step 6.2: Post-Cleanup Metrics

Capture final metrics for comparison:

```bash
# Total doc file count
find docs/ specs/ -name "*.md" | wc -l
# Result: ______

# Total line count
find docs/ specs/ -name "*.md" -exec wc -l {} + | tail -1
# Result: ______

# Spec count (should be unchanged or minimally reduced)
ls -d specs/*/ | wc -l
# Result: ______
```

### Step 6.3: Git History Verification

- [ ] All cleanup decisions logged: `git log --oneline --grep="docs:" | wc -l`
- [ ] Sample commit messages reviewed for quality (rationale present)
- [ ] Deleted files recoverable: `git log --all --full-history -- "docs/archive/<sample-deleted-file>"`

---

## Phase 7: Final Commit & PR

### Step 7.1: Final Review

- [ ] All phases 1-6 complete
- [ ] Git status clean (no uncommitted changes)
- [ ] All success criteria met
- [ ] Changelog/release notes updated (if applicable)

### Step 7.2: Create Summary Commit (if needed)

If small fixups needed:
```bash
git commit -m "docs: finalize documentation cleanup - Phase 7 complete"
```

### Step 7.3: Prepare for Merge

- [ ] All commits follow conventional commit format
- [ ] Branch ready for PR or direct merge to main
- [ ] Documentation about cleanup logged in commit messages (no separate changelog needed)

---

## Audit Completion Summary

**Date Started**: ______
**Date Completed**: ______
**Total Time**: ______ hours

**Metrics**:
- Docs reviewed: ______
- Docs deleted: ______
- Docs consolidated: ______
- Broken links fixed: ______
- Commits created: ______

**Outcome**: All functional requirements (FR-001 through FR-010) and success criteria (SC-001 through SC-007) validated.
