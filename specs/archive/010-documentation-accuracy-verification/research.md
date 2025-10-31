# Research: Documentation Verification Methodology

**Feature**: Documentation Accuracy Verification and Remediation
**Phase**: 0 (Research & Discovery)
**Date**: 2025-10-15

## Research Questions

### Q1: CLI Verification Approach

**Question**: What is the most reliable method to extract and compare documented CLI flags against actual cobra command definitions?

**Options Evaluated**:

1. **AST Parsing** (Go's `go/parser` and `go/ast`):
   - Pros: Programmatic, can extract struct tags and flag definitions directly
   - Cons: Complex, requires maintaining parsing logic, brittle to code structure changes
   - Verdict: Over-engineered for documentation audit

2. **Execute `--help` and Parse Output**:
   - Pros: Validates actual runtime behavior, matches user experience exactly
   - Cons: Requires building binary, output format may vary
   - Verdict: **SELECTED** - Most reliable, tests actual CLI behavior

3. **Manual Inspection with Checklist**:
   - Pros: Simple, no tooling required
   - Cons: Error-prone, not reproducible, time-consuming
   - Verdict: Acceptable for small audits, but systematic approach preferred

**Decision**: **Option B** - Execute `pass-cli [command] --help` and compare output against documented flag tables. This approach verifies the actual user-facing behavior rather than code structure. Create structured checklist documenting expected vs. actual for each command.

**Rationale**: Users experience the CLI through `--help` output, not source code. Verifying against runtime behavior ensures documentation matches what users actually see. This also catches discrepancies in flag descriptions, default values, and usage examples.

---

### Q2: Code Example Testing Strategy

**Question**: How should code examples be extracted and tested systematically?

**Approach**:

1. **Extraction Pattern**:
   - Markdown code blocks with language tags: ` ```bash `, ` ```powershell `, ` ```sh `
   - Manual review required for inline code (`` `pass-cli add myservice` ``)
   - Exclude output examples (not executable)

2. **Test Vault Setup**:
   - Create temporary test vault: `~/.pass-cli-test/vault.enc`
   - Pre-populate with test credentials: `testservice` (user: test@example.com, password: TestPass123!)
   - Use `--vault` flag to isolate from user's actual vault

3. **Platform Handling**:
   - Bash examples: Execute on Git Bash (Windows) / native bash (macOS/Linux)
   - PowerShell examples: Execute only on Windows
   - Mark platform-specific examples with annotations in audit report

4. **Validation**:
   - Command executes with exit code 0
   - Output format matches documented example (if output shown)
   - Side effects documented (e.g., "credential added" → verify with `pass-cli list`)

**Decision**: Manual extraction with structured test log. For each example:
1. Copy code block to test script
2. Execute in clean test vault environment
3. Record: Pass/Fail + actual output/error
4. Flag discrepancies in audit report

**Rationale**: Automation would require sophisticated markdown parsing and output comparison logic. Manual testing with structured logging provides sufficient rigor for one-time audit while remaining simple (Principle VII: Simplicity & YAGNI).

---

### Q3: Discrepancy Categorization & Severity

**Question**: What severity/priority levels should be assigned to different discrepancy types?

**Severity Levels Defined**:

| Severity | Definition | Examples | Remediation Priority |
|----------|------------|----------|---------------------|
| **Critical** | Documentation claims feature exists, but it's completely broken or non-existent. User attempts fail immediately. | `--generate` flag documented but doesn't exist; command documented but not in codebase | P1 - Fix immediately |
| **High** | Documentation is incorrect in a way that causes user confusion or wasted time. Feature exists but behaves differently. | Flag exists but has different name/type; incorrect default values; wrong syntax examples | P1 - Fix immediately |
| **Medium** | Feature exists but is undocumented or incompletely documented. User can discover through `--help` but docs don't guide them. | `--category` flag missing from USAGE.md; valid config fields undocumented | P2 - Fix in batch |
| **Low** | Cosmetic issues that don't affect functionality. Metadata staleness, minor formatting issues. | "Last Updated: January 2025" when actually October; broken internal links; typos in descriptions | P3 - Fix in batch |

**Decision**: Use 4-level severity model above. Critical/High issues remediated immediately (individual commits), Medium/Low batched by file.

**Rationale**: Aligns with user impact. Critical/High issues directly cause user failure or frustration. Medium issues are discoverable workarounds. Low issues are polish. This prioritization ensures high-impact fixes aren't delayed by low-priority cleanup.

---

### Q4: Audit Report Format

**Question**: What structure enables efficient remediation tracking?

**Structure Decision**: **Per-File Grouping** with category tags

**Format**:

```markdown
# Documentation Accuracy Audit Report

**Audit Date**: 2025-10-15
**Scope**: 7 primary docs + all docs/ subdirectory files
**Methodology**: Manual verification per verification-procedures.md
**Total Discrepancies**: [count]

## Summary Statistics

| Category | Total | Critical | High | Medium | Low |
|----------|-------|----------|------|--------|-----|
| CLI Interface | X | X | X | X | X |
| Code Examples | X | X | X | X | X |
| (... 10 categories) | X | X | X | X | X |

| File | Discrepancies | Status |
|------|---------------|--------|
| README.md | X | In Progress / Fixed / Verified |
| docs/USAGE.md | X | ... |

## Discrepancy Details

### README.md (X discrepancies)

#### DISC-001 [CLI/Critical] Non-existent `--generate` flag documented

- **Location**: README.md:158
- **Documented**: `pass-cli add newservice --generate`
- **Actual**: Flag does not exist in cmd/add.go
- **Remediation**: Remove flag from example, document separate `pass-cli generate` command
- **Status**: ❌ Open / ✅ Fixed / ✓ Verified
- **Commit**: [hash once fixed]

(Repeat for each discrepancy)
```

**Alternative Considered**: Per-category grouping
- Pros: Easy to see all CLI issues together
- Cons: Hard to track per-file completion, scattered fixes across report
- Rejected: File-based grouping aligns with remediation workflow (fix all issues in README.md, commit, move to next file)

**Decision Rationale**: Per-file grouping matches remediation workflow. Maintainer can work through one file at a time, mark all discrepancies fixed, commit, and move to next file. Category tags enable cross-cutting analysis via grep/search.

---

## Best Practices Review

**Reviewed Sources**:
- Spec 008 (documentation review methodology): Focused on content quality, not accuracy verification
- Spec 009 (documentation cleanup): Focused on obsolescence and duplication, not factual correctness
- Industry practice: W3C spec reviews, OpenAPI schema validation, API documentation testing

**Key Learnings**:

1. **Schema-Based Validation**: OpenAPI uses machine-readable schemas to validate API documentation. For CLI, `--help` output is analogous schema.

2. **Example-Driven Testing**: W3C specs include test suites that execute documented examples. We should test every code block that claims to work.

3. **Version Tracking**: Documentation should reference implementation version. We lack this (all docs say "v0.0.1" but features evolved). **Recommendation**: Accept current state (all v0.0.1) since no releases yet, but flag for future when versioned releases begin.

4. **Continuous Validation**: One-time audits drift. **Recommendation**: Document this in CONTRIBUTING.md - any new feature MUST update docs AND verify accuracy before merge.

---

## Methodology Summary

**Verification Approach** (per category):

1. **CLI Interface**: Execute `pass-cli [cmd] --help`, compare against USAGE.md/README.md tables
2. **Code Examples**: Extract bash/PowerShell blocks, execute in test vault, verify exit codes/output
3. **File Paths**: Grep for path references, verify against internal/config and platform conventions
4. **Configuration**: Extract YAML examples, validate against internal/config validation rules
5. **Feature Claims**: Manual testing per feature (audit logging, keychain, password policy, TUI shortcuts)
6. **Architecture**: Compare documented package descriptions against `ls internal/` and package godoc
7. **Metadata**: Grep for version/date references, cross-check against `git tag` and `git log`
8. **Output Examples**: Execute commands, compare actual output against documented samples
9. **Cross-References**: Parse markdown links, verify anchors/files exist
10. **Behavioral Claims**: Identify "when X, then Y" statements, verify via testing

**Tooling**:
- Standard shell tools (grep, find, ls)
- `pass-cli` binary (for --help and example execution)
- Git (for version/date verification)
- Text editor (for YAML validation against code inspection)

**No custom automation**: Aligns with Principle VII (Simplicity). One-time audit doesn't justify tooling investment.

---

## Decisions Summary

| Research Question | Decision | Rationale |
|-------------------|----------|-----------|
| CLI Verification Approach | Execute `--help` and compare output | Matches user experience, validates runtime behavior |
| Code Example Testing | Manual extraction + test vault execution | Simple, sufficient rigor for one-time audit |
| Discrepancy Severity Levels | 4 levels (Critical/High/Medium/Low) | Aligns with user impact, guides remediation priority |
| Audit Report Format | Per-file grouping with category tags | Matches remediation workflow (fix file by file) |
| Tooling | Standard shell tools + pass-cli binary | No custom automation (Simplicity principle) |

---

**Next**: Create `verification-procedures.md` with detailed Given/When/Then test procedures for each category.
