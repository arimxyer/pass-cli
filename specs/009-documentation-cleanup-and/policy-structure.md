# Policy Structure: Documentation Lifecycle

**Feature**: Documentation Cleanup and Archival
**Phase**: 1 (Design)
**Date**: 2025-10-15

## Purpose

This document defines the structure and content template for `docs/DOCUMENTATION_LIFECYCLE.md`, the repository's documentation governance policy.

## Policy Document Template

Based on Phase 0 research (industry best practices from Kubernetes, Rust, Python), the policy document should follow this 6-section structure:

---

### Template: DOCUMENTATION_LIFECYCLE.md

```markdown
# Documentation Lifecycle Policy

**Version**: 1.0.0
**Effective**: [Date]
**Owner**: [Repository Maintainers]

## 1. Purpose & Scope

**Purpose**: Define retention periods, archival triggers, and decision workflows for Pass-CLI documentation to maintain high-quality, current documentation while preserving historical design context.

**Scope**: This policy covers all documentation in:
- `docs/` (user-facing guides, installation, usage, troubleshooting)
- `specs/` (feature specifications, design documents, ADRs)
- Root documentation (README.md, CONTRIBUTING.md, SECURITY.md)
- Inline code documentation (Go doc comments)

**Out of Scope**: This policy does NOT cover:
- Source code (covered by version control practices)
- Issue/PR templates (covered by GitHub configuration)
- CI/CD configuration docs (covered by development workflow)

---

## 2. Document Classification

All documentation falls into one of these types:

| Type | Examples | Retention Period | Rationale |
|------|----------|------------------|-----------|
| **Specifications** | specs/###-feature/spec.md, plan.md, tasks.md | Indefinite | Preserves complete design history and decision rationale. Minimal storage impact for markdown files. |
| **User Guides** | INSTALLATION.md, USAGE.md, docs/*.md | Until content superseded or feature removed | Active content updated in place. Archive when feature deprecated/removed. |
| **API Documentation** | Go doc comments, library README files | Until API version unsupported | Tied to code versions. Deprecation schedule drives archival. |
| **Examples** | examples/, code snippets in guides | Until unmaintainable | Keep while code runs on supported versions. Archive when breaking changes render obsolete. |
| **Troubleshooting** | TROUBLESHOOTING.md, KNOWN_LIMITATIONS.md | Until problem no longer possible | Keep while issue can occur. Archive when underlying cause fixed or feature removed. |
| **Migration Guides** | MIGRATION.md, version upgrade guides | 2-3 versions after target version | Users need time to upgrade. Archive once user base migrated (track via telemetry/issues). |
| **Governance** | CONTRIBUTING.md, this policy, constitution.md | Permanent | Foundational documents. Update in place via amendment process. |

---

## 3. Retention Rules

### Rule 1: Indefinite Retention for Specs

All feature specifications (specs/###-feature/) MUST be retained indefinitely to preserve:
- Design decision rationale
- Acceptance criteria and validation approach
- Historical context for future maintenance

**Exception**: Specs may be deleted ONLY if:
1. Feature was never implemented AND
2. Spec was created in error (duplicate, wrong repo, etc.)

**Process**: Maintainer judgment with rationale in commit message.

### Rule 2: Active Content Updates

User guides, troubleshooting docs, and governance documents MUST be updated in place rather than versioned/archived. These serve current users and must reflect current state.

**Archival Trigger**: Content becomes obsolete when:
- Feature is deprecated (mark deprecated, archive 1 version later)
- Feature is removed (archive immediately)
- Content is fully superseded by better documentation (consolidate, then archive old)

### Rule 3: Version-Tied Documentation

API docs and migration guides tied to specific versions follow version lifecycle:

- **API Docs**: Archive when version reaches end-of-life
- **Migration Guides**: Archive 2-3 versions after target (when 95%+ users migrated)

**Tracking**: Monitor usage metrics, support tickets, and community discussion to gauge migration completion.

### Rule 4: Git History as Archive

Deleted documentation MUST have git history preserved. Use permanent deletion (not move to docs/archive/) because:
- Git history provides complete audit trail
- `git log --follow` tracks file across renames
- Recovery via `git show <commit>:<path>` when needed
- Reduces search noise and maintenance burden

---

## 4. Archival Triggers

Documentation should be archived/removed when ANY of these conditions occur:

1. **Feature Lifecycle**:
   - Feature deprecated → Mark docs deprecated
   - Feature removed → Archive docs immediately

2. **Content Quality**:
   - Factually incorrect and misleading → Delete immediately
   - Duplicate of canonical source with no unique info → Consolidate + delete
   - Superseded by better doc → Consolidate unique content, delete original

3. **Maintainability**:
   - Examples/code no longer runs on supported versions → Archive
   - Content requires frequent corrections due to staleness → Update or archive
   - Link rot (>50% broken links, unfixable) → Update or archive

4. **Relevance**:
   - Problem no longer occurs (troubleshooting) → Archive
   - Migration complete (guide no longer needed) → Archive
   - Experimental doc never finalized → Delete

---

## 5. Decision Authority

**Solo Maintainer**: Ari (primary repository owner)

**Decision Process**:
1. Identify documentation meeting archival triggers
2. Review for unique information not captured elsewhere
3. Decide: Update, Consolidate, or Delete
4. Document rationale in git commit message
5. Commit with format: `docs: archive <file> - <rationale>`

**Escalation**: For governance docs (this policy, CONTRIBUTING.md, constitution.md), discuss changes in GitHub Issue before committing.

**Review Frequency**: Quarterly documentation audit (every 3 months) to identify obsolete content.

---

## 6. Process Workflow

### Archiving User Guides / Troubleshooting

1. **Identify**: Run quarterly audit or flag during regular maintenance
2. **Evaluate**: Check against retention rules (Section 3)
3. **Extract Unique Content**: Copy any information not in canonical source
4. **Delete**: `git rm <file>`
5. **Commit**: Detailed commit message with rationale
6. **Update Links**: Fix references in other docs
7. **PR**: Submit as atomic change (one doc per PR when possible)

**Commit Message Format**:
```
docs: archive <filename> - <rationale>

Rationale: <detailed explanation>
- Archival trigger: <which condition from Section 4>
- Unique content: <none OR preserved in <other file>>
- Last meaningful update: <date>

Accessible via git history at commit <sha> if recovery needed.
```

### Archiving Specs

**Default**: DO NOT archive specs (Rule 1: indefinite retention)

**Exception**: If spec created in error:
```
docs: remove incorrect spec 042 - created in wrong repository

Rationale: Spec 042 documented feature for pass-cli-gui project,
mistakenly created in pass-cli repo. Feature never implemented here.
Correct spec exists at pass-cli-gui repo.
```

### Consolidating Duplicate Content

1. **Identify**: Use ripgrep keyword search: `rg -i "<topic>" docs/`
2. **Review**: Read all matches to confirm true duplication
3. **Select Canonical Source**: Choose most comprehensive doc
4. **Merge Unique Content**: Add missing info from duplicates to canonical
5. **Delete Duplicates**: `git rm <duplicate files>`
6. **Update Cross-References**: Add "See [Canonical Doc]" links where helpful
7. **Commit**: One commit per consolidated topic

**Commit Message Format**:
```
docs: consolidate <topic> documentation into <canonical file>

- Merged content from <file1>, <file2>, <file3>
- Canonical source: <chosen file>
- Deleted duplicates: <list>
- No information loss: <confirm all unique content preserved>

Reduces duplicate maintenance and improves search clarity.
```

### Link Validation & Fixing

1. **Manual Check**: Review links in edited files
2. **Update Broken Links**: Fix paths, update URLs
3. **Remove Dead Links**: If target permanently gone, remove link
4. **Document**: Note link fixes in commit message

**Commit Message Format**:
```
docs: fix broken links in <file>

- Updated <N> internal links (file moves)
- Removed <N> dead external links (404s)
- All links validated manually
```

---

## Governance

### Policy Updates

This policy follows semantic versioning:
- **MAJOR**: Backward-incompatible changes (e.g., changing retention rules)
- **MINOR**: New sections, expanded guidance
- **PATCH**: Clarifications, typo fixes

**Amendment Process**:
1. Open GitHub Issue proposing change
2. Draft amendment with version bump rationale
3. Discuss with maintainers (if multiple)
4. Commit with message: `docs: update documentation lifecycle policy to v1.X.Y`
5. Update CONTRIBUTING.md if integration text changes

### Compliance

Maintainers MUST follow this policy for all documentation changes. Violations should be caught in PR review.

**Audit Trail**: Every archival decision documented in git commit message provides complete audit trail accessible via `git log --all --grep="archive" --oneline`.

---

**Version**: 1.0.0 | **Effective**: 2025-10-15 | **Last Amended**: N/A
```

---

## Implementation Notes

**Placeholders to Fill During Implementation**:
1. `[Date]` in Effective field → actual policy ratification date
2. `[Repository Maintainers]` → Ari (or expanded list if governance changes)
3. Examples in commit message formats → real examples from cleanup work

**Customization Points**:
- Retention periods for migration guides (currently "2-3 versions") may need adjustment based on release cadence
- Review frequency (quarterly) can be tuned based on documentation velocity
- Decision authority section may expand if project governance changes

**Validation Checklist** (before committing policy to docs/DOCUMENTATION_LIFECYCLE.md):
- [ ] All 6 sections present and complete
- [ ] Retention rules cover all document types from Classification table
- [ ] Archival triggers are specific and measurable
- [ ] Process workflows include commit message templates
- [ ] Policy version and dates filled in
- [ ] No contradictions with spec requirements (FR-001 through FR-010)
- [ ] Policy aligns with clarifications (git commit decision log, CONTRIBUTING.md integration)
