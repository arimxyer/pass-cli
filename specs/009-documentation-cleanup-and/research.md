# Research: Documentation Lifecycle Best Practices

**Feature**: Documentation Cleanup and Archival
**Phase**: 0 (Research & Discovery)
**Date**: 2025-10-15

## Research Questions

1. What are industry best practices for documentation lifecycle policies?
2. How do successful open-source projects manage documentation retention?
3. What tools exist for automated documentation quality checking?
4. What criteria should guide archive vs. delete decisions?

## Findings

### 1. Documentation Lifecycle Policy Structure

**Research Sources**: Analysis of Kubernetes, Rust, Python, and Go documentation governance

**Key Findings**:

**Standard Policy Sections**:
1. **Purpose & Scope**: Why the policy exists, what docs it covers
2. **Document Classification**: Types (user guides, API docs, specs, examples, RFCs, ADRs)
3. **Retention Rules**: How long each type is kept
4. **Archival Triggers**: When to archive (version sunset, feature removal, content superseded)
5. **Decision Authority**: Who approves archival (solo maintainer vs. committee)
6. **Process Workflow**: Steps for archiving (review → decision → commit with rationale → PR)

**Decision**: Adopt this 6-section structure for `docs/DOCUMENTATION_LIFECYCLE.md`

**Rationale**: Industry-standard structure ensures policy is comprehensive and follows familiar patterns for contributors who've worked on other projects.

**Alternatives Considered**:
- Minimal 2-section policy (rules + process): Rejected—too vague, leads to inconsistent decisions
- Embed policy in CONTRIBUTING.md: Rejected—clutters contribution guidelines, policy warrants dedicated document per spec clarification

---

### 2. Retention Rules by Document Type

**Research Sources**: Linux kernel docs, Django, Rails, Node.js documentation practices

**Key Findings**:

| Document Type | Retention Policy | Rationale |
|---------------|------------------|-----------|
| **Specs** | Indefinite | Preserves design rationale and decision history (per FR-008 clarification) |
| **User Guides** | Until content superseded | Active content; update in place or archive when feature removed |
| **API Docs** | Until API version unsupported | Tied to versioned APIs; deprecation schedule drives archival |
| **Examples** | Until framework/language obsolete | Practical value as long as code runs; archive when breaking changes make unmaintainable |
| **Troubleshooting** | Until problem no longer occurs | Keep while issue is possible; archive when underlying cause removed |
| **Migration Guides** | 2-3 versions after target | Users need time to upgrade; archive once user base migrated |

**Decision**: Adopt type-specific retention rules matching table above, with specs retained indefinitely per clarification

**Rationale**: Different documentation types have different lifespans. User guides need frequent updates while specs provide permanent design history.

**Alternatives Considered**:
- Universal retention period (e.g., 1 year): Rejected—doesn't account for different doc purposes
- No formal retention rules: Rejected—leads to inconsistent cleanup decisions

---

### 3. Archive vs. Delete Decision Criteria

**Research Sources**: Google's documentation style guide, Write the Docs community practices, GitHub's own docs evolution

**Key Findings**:

**Delete Permanently When**:
- Content is factually incorrect and misleading
- Duplicate of canonical source with no unique information
- Temporary/experimental docs that were never finalized
- Outdated tutorials that would harm users if followed
- Git history provides sufficient context if recovery needed

**Reasons to Delete**:
1. Reduces noise in search results
2. Prevents user confusion from contradictory information
3. Simplifies maintenance surface
4. Git history preserves for audit/recovery

**Decision**: Default to permanent deletion with rationale in commit message, relying on git history for recovery (per Clarification #1)

**Rationale**: Git history provides complete audit trail and recovery mechanism. Keeping obsolete docs in active repo increases maintenance burden and confuses users.

**Alternatives Considered**:
- Move to docs/archive/obsolete/: Rejected—users still find in searches, no benefit over git history
- External archive system: Rejected—adds complexity, unnecessary for markdown files

---

### 4. Link Checking & Validation Tools

**Research Sources**: Awesome-markdown list, documentation tooling surveys, GitHub Actions marketplace

**Key Findings**:

**Recommended Tools**:
1. **markdown-link-check** (npm)
   - Pros: Simple, widely used, GitHub Actions integration
   - Cons: Requires Node.js

2. **lychee** (Rust)
   - Pros: Fast, comprehensive (checks markdown + HTML + text files), parallel
   - Cons: Another toolchain dependency

3. **Built-in ripgrep + manual validation**
   - Pros: Already available, no new dependencies
   - Cons: Manual process, no automated link validation

**Decision**: Use ripgrep for initial duplicate detection, manual validation for links (fits Constitution Principle VII: prefer standard tooling, avoid unnecessary dependencies)

**Rationale**: Documentation cleanup is one-time work. Adding npm/Rust tool dependencies for single-use validation doesn't justify complexity. Manual validation with existing tools (ripgrep, grep) aligns with simplicity principle.

**Alternatives Considered**:
- markdown-link-check in CI: Rejected for one-time cleanup—good for ongoing validation but overkill for this feature
- lychee: Rejected—Rust toolchain not currently in project

---

### 5. Duplicate Content Detection Strategy

**Research Sources**: Content deduplication practices from technical writing communities

**Key Findings**:

**Detection Approaches**:
1. **Keyword/Title Search**: `rg -i "installation" docs/` finds overlapping topics
2. **Section Header Analysis**: Compare H2/H3 headings across files
3. **Manual Review**: Read through docs to identify conceptual overlap

**Consolidation Patterns**:
- **Pattern A**: Keep most comprehensive doc, add "See also" links from others
- **Pattern B**: Merge into single canonical doc, delete duplicates
- **Pattern C**: Split if docs serve different audiences (quick start vs. deep dive)

**Decision**: Use ripgrep for keyword search → manual review → Pattern B (merge + delete) for true duplicates, Pattern A (cross-reference) for different audience levels

**Rationale**: Automated deduplication risks losing nuanced differences. Manual review with keyword-based hints balances efficiency and accuracy.

**Alternatives Considered**:
- Semantic similarity tools (NLP-based): Rejected—massive overkill, adds dependencies
- Leave duplicates with warnings: Rejected—doesn't achieve spec goal of reducing maintenance burden

---

### 6. CONTRIBUTING.md Integration Patterns

**Research Sources**: React, Vue, Angular, Svelte documentation governance sections

**Key Findings**:

**Common Patterns**:
1. **Inline Policy**: Entire policy embedded in CONTRIBUTING.md
2. **Summary + Link**: 2-3 sentence overview with link to full policy doc
3. **New Section**: "Documentation Governance" section with policy details
4. **Separate GOVERNANCE.md**: Link from CONTRIBUTING.md to standalone governance doc

**Decision**: Pattern #2 (Summary + Link) per Clarification #2—add brief summary to CONTRIBUTING.md linking to docs/DOCUMENTATION_LIFECYCLE.md

**Rationale**: CONTRIBUTING.md already long; inline policy would clutter. Summary provides visibility, link provides details for those who need it.

**Alternatives Considered**:
- Full inline policy: Rejected—clutters CONTRIBUTING.md, violates single-source-of-truth
- Link only: Rejected—less discoverable, summary aids skimming

---

## Implementation Recommendations

Based on research findings:

1. **Policy Document Structure**: Use 6-section industry-standard format (Purpose, Classification, Retention, Triggers, Authority, Process)

2. **Retention Rules**: Implement type-specific rules with indefinite spec retention (per FR-008)

3. **Cleanup Process**:
   - Use ripgrep for duplicate detection keyword searches
   - Manual review for quality assessment
   - Delete obsolete docs permanently (git history preserves)
   - Document rationale in commit messages

4. **Link Validation**: Manual checking with existing tools (no new dependencies)

5. **CONTRIBUTING.md Integration**: Add 2-3 sentence summary + link to full policy

6. **Success Metrics**: Track before/after counts:
   - Total doc file count
   - Duplicate topic areas (pre-consolidation vs. post)
   - Broken link count (pre-fix vs. post-fix)

---

## Open Questions

None—all technical unknowns resolved through research.

## Next Phase

Proceed to Phase 1 (Design) to create:
- `policy-structure.md`: Template for DOCUMENTATION_LIFECYCLE.md with 6 sections
- `audit-checklist.md`: Systematic checklist for reviewing existing documentation
