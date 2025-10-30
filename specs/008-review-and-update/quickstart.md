# Quickstart Guide: Documentation Review and Production Release Preparation

**Feature**: Documentation Review and Production Release Preparation
**Branch**: `008-review-and-update`
**Date**: 2025-01-14

## Purpose

This quickstart provides the minimal steps to validate Pass-CLI documentation accuracy for production release. Follow these steps to ensure all documentation meets the quality standards defined in the specification.

## Prerequisites

1. **Pass-CLI binary installed** (current release: v0.0.1 or later)
   ```bash
   pass-cli version
   # Expected output: pass-cli version v0.0.1 (or higher)
   ```

2. **Repository cloned and feature branch checked out**
   ```bash
   git clone https://github.com/ari1110/pass-cli.git
   cd pass-cli
   git checkout 008-review-and-update
   ```

3. **Tools installed** (for automated validation):
   - `grep`, `sed`, `awk` (standard on Unix, available via Git Bash on Windows)
   - `markdown-link-check` (optional): `npm install -g markdown-link-check`
   - Bash shell (Git Bash for Windows users)

## Quick Validation (5 Minutes)

Run automated validation scripts to identify critical issues:

```bash
# Navigate to validation scripts directory
cd specs/008-review-and-update/validation

# 1. Check for outdated version references (SC-007)
./version-audit.sh
# Expected: Zero outdated iteration counts (100k), zero version inconsistencies

# 2. Validate CLI command examples (SC-003)
./command-tests.sh
# Expected: 100% of documented commands execute successfully

# 3. Check external links (optional)
./link-check.sh
# Expected: All links return HTTP 200

# 4. Verify cross-references
./cross-reference-check.sh
# Expected: All internal document references resolve
```

**If all scripts PASS**: Proceed to manual validation

**If any script FAILS**: Review script output, fix issues in documentation files, re-run scripts

## Manual Validation (30-60 Minutes)

### Step 1: README.md Review (10 minutes)

**File**: `R:\Test-Projects\pass-cli\README.md`

**Checklist**:
- [ ] Quick Start section demonstrates: init → add → get → copy workflow
- [ ] All Quick Start commands execute successfully
- [ ] Walkthrough completable in ≤10 minutes (time yourself)
- [ ] TUI keyboard shortcuts table includes 20+ shortcuts (not just 6)
- [ ] Feature roadmap marks completed features with `[x]`
- [ ] Configuration file documentation (config.yml) present
- [ ] Current release version mentioned

**Test**:
```bash
# Execute Quick Start commands
pass-cli init                          # Initialize vault (prompts for password)
pass-cli add github                    # Add credential (prompts for username/password)
pass-cli get github                    # Retrieve credential
pass-cli get github --copy             # Copy to clipboard
pass-cli delete github --force         # Clean up test credential
```

**Acceptance**: All commands execute without errors, workflow completes in ≤10 minutes

---

### Step 2: SECURITY.md Review (15 minutes)

**File**: `R:\Test-Projects\pass-cli\docs\SECURITY.md`

**Checklist**:
- [ ] AES-256-GCM algorithm explicitly documented
- [ ] PBKDF2-SHA256 with **600,000 iterations** (not 100,000)
- [ ] 32-byte salt, 12-byte nonce documented
- [ ] Random sources: `crypto/rand` (Windows `CryptGenRandom`, Unix `/dev/urandom`)
- [ ] NIST compliance references (SP 800-38D, SP 800-132)
- [ ] Password policy: 12+ chars, uppercase, lowercase, digit, symbol
- [ ] Audit logging: HMAC-SHA256, `--enable-audit` flag, `pass-cli verify-audit` command
- [ ] Migration path to 600k iterations documented, references MIGRATION.md
- [ ] TUI security warnings (shoulder surfing, screen recording, shared terminals)

**Test**:
```bash
# Verify iteration count references
grep -c "600,000\|600k" docs/SECURITY.md
# Expected: >0 (should appear multiple times)

# Ensure no outdated iteration counts (except in migration context)
grep "100,000\|100k" docs/SECURITY.md
# Expected: Only in migration/historical context, not as current spec
```

**Acceptance**: Security professional can identify all crypto parameters (algorithm, key size, iterations, nonce) without reading source code (SC-002)

---

### Step 3: USAGE.md Review (20 minutes)

**File**: `R:\Test-Projects\pass-cli\docs\USAGE.md`

**Checklist**:
- [ ] "TUI vs CLI Mode" section clearly explains `pass-cli` (TUI) vs `pass-cli <command>` (CLI)
- [ ] TUI Keyboard Shortcuts table documents 20+ shortcuts
- [ ] Shortcuts match implementation: n, e, d, p, c, i, s, /, ?, q, Ctrl+H, Ctrl+S, Ctrl+C, Tab, Shift+Tab, etc.
- [ ] Custom keybinding documentation references config.yml
- [ ] All documented CLI commands execute successfully
- [ ] File paths accurate: `~/.pass-cli/vault.enc`, `~/.pass-cli/config.yaml`, `~/.pass-cli/audit.log`

**Test**:
```bash
# Count TUI shortcuts documented
grep -A 100 "TUI Keyboard Shortcuts" docs/USAGE.md | grep -c "^| "
# Expected: ≥20 rows (excluding header)

# Execute a sample of documented commands
pass-cli list --help
pass-cli get --help
pass-cli add --help
pass-cli generate --help
# Expected: All commands return help text without errors
```

**Acceptance**: 100% of documented commands execute successfully (SC-003), all 20+ TUI shortcuts documented (SC-006)

---

### Step 4: INSTALLATION.md Review (10 minutes)

**File**: `R:\Test-Projects\pass-cli\docs\INSTALLATION.md`

**Checklist**:
- [ ] Homebrew tap command: `brew tap ari1110/homebrew-tap`
- [ ] Homebrew install command: `brew install pass-cli`
- [ ] Scoop bucket command: `scoop bucket add pass-cli https://github.com/ari1110/scoop-bucket`
- [ ] Scoop install command: `scoop install pass-cli`
- [ ] Manual installation download links point to current release (v0.0.1 or later)
- [ ] Checksum verification instructions accurate
- [ ] Build from source: Go 1.25+ requirement, `go build` or `make build` commands

**Test** (if applicable platform):
```bash
# macOS/Linux: Check Homebrew info
brew info pass-cli
# Expected: Shows pass-cli package, current version

# Windows: Check Scoop info
scoop info pass-cli
# Expected: Shows pass-cli package, current version
```

**Acceptance**: All installation methods complete successfully (SC-005)

---

### Step 5: TROUBLESHOOTING.md Review (10 minutes)

**File**: `R:\Test-Projects\pass-cli\docs\TROUBLESHOOTING.md`

**Checklist**:
- [ ] TUI-specific issues covered: rendering artifacts, keyboard shortcuts not working, black screen, search not filtering, Ctrl+H toggle, sidebar/detail panel visibility
- [ ] Platform-specific solutions: Windows (Credential Manager), macOS (Keychain), Linux (Secret Service, D-Bus)
- [ ] Each issue has actionable solution with specific commands
- [ ] Solutions findable in ≤5 minutes via search/scan

**Test**:
```bash
# Time yourself finding solution for "TUI rendering" issue
# Method: grep or manual scan
grep -i "rendering" docs/TROUBLESHOOTING.md
# Expected: Find solution in ≤5 minutes
```

**Acceptance**: Common issues findable and solvable in ≤5 minutes (SC-004)

---

### Step 6: MIGRATION.md Review (5 minutes)

**File**: `R:\Test-Projects\pass-cli\docs\MIGRATION.md`

**Checklist**:
- [ ] Documents PBKDF2 iteration upgrade from 100,000 to 600,000
- [ ] Explains security rationale (OWASP recommendations)
- [ ] Provides migration command or process
- [ ] Explains backward compatibility (old vaults work, but should upgrade)
- [ ] Performance impact documented (~50-100ms increase)
- [ ] Consistent with SECURITY.md iteration counts

**Test**:
```bash
# Verify migration command exists (if documented)
pass-cli migrate --help
# OR
pass-cli --help | grep -i migrate

# Cross-reference with SECURITY.md
diff <(grep "600,000\|600k" docs/SECURITY.md) <(grep "600,000\|600k" docs/MIGRATION.md)
# Expected: Iteration counts consistent across files
```

**Acceptance**: Migration path clearly documented, cross-references consistent

---

### Step 7: KNOWN_LIMITATIONS.md Review (5 minutes)

**File**: `R:\Test-Projects\pass-cli\docs\KNOWN_LIMITATIONS.md`

**Checklist**:
- [ ] All listed limitations exist in current release
- [ ] No limitations that have been resolved (e.g., no "no TUI mode" if TUI exists)
- [ ] Review specs 001-007 to identify features that may have resolved old limitations

**Test**:
```bash
# Verify each limitation by testing
# Example: If doc says "No import from other password managers"
pass-cli import --help
# Expected: Should fail if limitation is accurate, or exist if limitation is outdated
```

**Acceptance**: All listed limitations current, no resolved limitations remain

---

## Fix Issues and Re-validate

### If Validation Failures Found:

1. **Document Issues**:
   - Create issue list in `specs/008-review-and-update/validation/issues.md`
   - Format: `[FILE]-[NUMBER]: [Severity] - [Description]`
   - Example: `README-001: HIGH - TUI shortcuts incomplete (only 6 shown, should be 20+)`

2. **Prioritize Fixes**:
   - CRITICAL: Command errors, missing security specs, broken install instructions
   - HIGH: Outdated version references, incomplete troubleshooting, broken links
   - MEDIUM: Formatting issues, minor inaccuracies
   - LOW: Typos, style inconsistencies

3. **Apply Fixes**:
   - Update documentation files to resolve issues
   - Test fixes (re-run commands, verify links, etc.)

4. **Re-run Validation**:
   - Execute automated scripts again: `./version-audit.sh`, `./command-tests.sh`, etc.
   - Re-test manual checklists for affected files
   - Repeat until all validation status = PASS

## Final Sign-off Checklist

Before marking documentation review complete:

- [ ] All 7 documentation files reviewed (README, SECURITY, USAGE, INSTALLATION, TROUBLESHOOTING, MIGRATION, KNOWN_LIMITATIONS)
- [ ] All automated validation scripts PASS (version-audit, command-tests, link-check, cross-reference)
- [ ] All manual validation checklists complete (no unchecked items)
- [ ] All CRITICAL and HIGH severity issues resolved
- [ ] Success criteria verified:
  - [ ] SC-001: New user onboarding in ≤10 minutes
  - [ ] SC-002: Crypto params identifiable from SECURITY.md
  - [ ] SC-003: 100% CLI command accuracy
  - [ ] SC-004: Troubleshooting solutions findable in ≤5 minutes
  - [ ] SC-005: All installation methods work
  - [ ] SC-006: 20+ TUI shortcuts documented
  - [ ] SC-007: Zero outdated version references
  - [ ] SC-008: Script examples run successfully

## Time Estimates

| Task | Estimated Time |
|------|----------------|
| Automated validation (scripts) | 5 minutes |
| README.md manual review | 10 minutes |
| SECURITY.md manual review | 15 minutes |
| USAGE.md manual review | 20 minutes |
| INSTALLATION.md manual review | 10 minutes |
| TROUBLESHOOTING.md manual review | 10 minutes |
| MIGRATION.md manual review | 5 minutes |
| KNOWN_LIMITATIONS.md manual review | 5 minutes |
| **Total (first pass)** | **80 minutes** |
| Issue fixes + re-validation | 60-120 minutes |
| **Grand Total** | **2-3 hours** |

## Success Indicators

✅ **Documentation ready for production release when**:
- All automated scripts return exit code 0 (PASS)
- All manual checklists 100% complete
- Zero CRITICAL or HIGH severity issues remain
- All 8 success criteria verified

🚫 **Block production release if**:
- Any CRITICAL issue unresolved (command errors, missing security specs)
- More than 3 HIGH severity issues unresolved
- Any success criterion fails (e.g., onboarding takes >10 minutes, outdated iteration counts remain)

## Next Steps

After successful validation:
1. Commit documentation updates to feature branch
2. Create pull request to main branch
3. Request review from project maintainers
4. Merge after approval
5. Tag release and update package managers (Homebrew, Scoop)

---

**Questions or Issues?**

- Review `data-model.md` for validation tracking structure
- Review `contracts/documentation-validation-schema.md` for detailed validation requirements
- Review `research.md` for current release version and feature inventory
