# Verification Procedures: Documentation Accuracy Testing

**Feature**: Documentation Accuracy Verification and Remediation
**Phase**: 1 (Design - Test Procedures)
**Date**: 2025-10-15

**Purpose**: Repeatable test procedures for each of the 10 verification categories. Each procedure follows Given/When/Then format and produces Pass/Fail results for audit report.

---

## Test Environment Setup

**Prerequisites**:
- Pass-CLI built from current main branch
- Test vault created: `~/.pass-cli-test/vault.enc`
- Test credentials added:
  ```bash
  export PASS_VAULT=~/.pass-cli-test/vault.enc
  pass-cli init
  pass-cli add testservice -u test@example.com -p TestPass123!
  pass-cli add github -u user@example.com
  ```

**Cleanup** (after audit):
```bash
rm -rf ~/.pass-cli-test
```

---

## Category 1: CLI Interface Verification

### Procedure

**For each command** listed in USAGE.md and README.md:

1. **Execute**: `pass-cli [command] --help`
2. **Capture**: Full help output
3. **Compare** against documented flags table in USAGE.md
4. **Verify**:
   - Flag name matches exactly (e.g., `--generate` vs. `--gen`)
   - Flag type matches (bool vs. string vs. int)
   - Short flag matches (e.g., `-u` for `--username`)
   - Description matches or is reasonably similar
   - Default values match (if shown)

**Commands to Verify** (14 total):
- `init`, `add`, `get`, `list`, `update`, `delete`, `generate`, `change-password`, `version`, `verify-audit`, `config`, `config init`, `config edit`, `config validate`, `config reset`, `tui`

**Test Template** (per command):

```
Given USAGE.md documents `pass-cli add` with flags table
When I run `pass-cli add --help`
Then output shows flags: --username/-u, --password/-p, --category/-c, --url, --notes
And flag types match documented types
And descriptions match documented text
```

**Pass Criteria**: All documented flags exist with matching names, types, and short flags.

**Failure Recording**: Create DISC-### entry for each mismatch:
- **Severity**: Critical (flag doesn't exist), High (flag name/type wrong), Medium (description differs)
- **Document**: Flag name, expected vs. actual

---

## Category 2: Code Examples Verification

### Procedure

**For each code block** in README.md, USAGE.md, MIGRATION.md, SECURITY.md with language tags ` ```bash `, ` ```sh `, ` ```powershell `:

1. **Extract** code block
2. **Modify** if needed:
   - Replace `myservice` placeholders with `testservice` (from test vault)
   - Replace file paths with test vault path if using `--vault` flag
3. **Execute** in test environment
4. **Verify**:
   - Exit code 0 (success)
   - No errors printed to stderr
   - Output format matches documented example (if output shown)
   - Side effects match claims (e.g., "adds credential" → verify with `pass-cli list`)

**Example Verification**:

```
Given README.md:158 shows code block:
  pass-cli add newservice --generate

When I execute command against test vault
Then command exits with code 0
And credential "newservice" appears in `pass-cli list`
And password meets complexity requirements
```

**Pass Criteria**: Command executes successfully with expected output/behavior.

**Failure Recording**: Create DISC-### entry:
- **Severity**: Critical (command fails), High (output doesn't match), Medium (side effect missing)
- **Document**: Code block line number, expected vs. actual behavior

**Special Cases**:
- **Platform-specific examples**: Mark as "Windows-only" or "macOS/Linux-only" in audit report
- **External dependencies**: Note if example requires `curl`, `mysql`, etc. (document dependency, skip execution)
- **Placeholder examples**: If example uses `myservice` as illustration (not meant to be literal), verify syntax only

---

## Category 3: File Paths Verification

### Procedure

**For each file path reference** in documentation (grep for `~/`, `%APPDATA%`, `/usr/`, `C:\`, `.pass-cli`, `.config`):

1. **Identify** platform (Windows/macOS/Linux)
2. **Check implementation**:
   - For config paths: inspect `internal/config/config.go` GetConfigPath()
   - For vault paths: inspect `cmd/root.go` GetVaultPath()
   - For examples: verify path format matches platform conventions
3. **Verify**:
   - Documented path matches implementation default
   - Platform-specific variants all documented (Windows, macOS, Linux)
   - Path format correct (forward vs. backslash)

**Example Verification**:

```
Given USAGE.md documents config location:
  Linux: ~/.config/pass-cli/config.yml
  macOS: ~/Library/Application Support/pass-cli/config.yml
  Windows: %APPDATA%\pass-cli\config.yml

When I inspect internal/config/config.go GetConfigPath()
Then implementation returns:
  Linux: filepath.Join(os.UserConfigDir(), "pass-cli", "config.yml")
  macOS: filepath.Join(os.UserConfigDir(), "pass-cli", "config.yml")
  Windows: filepath.Join(os.Getenv("APPDATA"), "pass-cli", "config.yml")
And documented paths match os.UserConfigDir() / APPDATA on each platform
```

**Pass Criteria**: All documented paths match implementation defaults for respective platforms.

**Failure Recording**: Create DISC-### entry:
- **Severity**: High (path incorrect), Medium (path format wrong), Low (missing platform variant)
- **Document**: Documented path, actual implementation path

---

## Category 4: Configuration Verification

### Procedure

**For each YAML configuration example** in USAGE.md, README.md:

1. **Extract** YAML block
2. **Parse** YAML (validate syntax)
3. **Compare** against internal/config structure:
   - Check `internal/config/config.go` Config struct fields
   - Check validation rules in LoadConfig() / Validate()
4. **Verify**:
   - All documented fields exist in Config struct
   - Field types match (bool vs. int vs. string)
   - Example values pass validation rules
   - No deprecated fields documented

**Example Verification**:

```
Given README.md shows example config.yml:
  terminal:
    warning_enabled: true
    min_width: 80
    min_height: 24
  keybindings:
    - key: "ctrl+c"
      action: "copy"

When I inspect internal/config/config.go
Then Config struct contains:
  type Config struct {
    Terminal struct {
      WarningEnabled bool `yaml:"warning_enabled"`
      MinWidth int `yaml:"min_width"`
      MinHeight int `yaml:"min_height"`
    }
    Keybindings []Keybinding `yaml:"keybindings"`
  }
And documented field names match YAML tags
And example values pass validation (min_width: 1-10000, min_height: 1-1000)
```

**Pass Criteria**: All documented fields exist, types match, example values valid.

**Failure Recording**: Create DISC-### entry:
- **Severity**: High (field doesn't exist), Medium (type mismatch), Low (invalid example value)
- **Document**: Field name, expected vs. actual type/validation

---

## Category 5: Feature Claims Verification

### Procedure

**For each documented feature** (audit logging, keychain integration, password policies, TUI shortcuts):

1. **Identify** feature description in docs
2. **Locate** implementation (internal/ packages)
3. **Test** actual behavior:
   - **Audit Logging**: Enable with `--enable-audit`, verify HMAC-SHA256 signatures exist in audit log file
   - **Keychain Integration**: Initialize with `--use-keychain`, verify Windows Credential Manager / macOS Keychain / Linux Secret Service contains entry
   - **Password Policy**: Attempt weak password on `pass-cli init`, verify rejection with policy message
   - **TUI Shortcuts**: Launch `pass-cli tui`, press documented shortcuts (Ctrl+H, Ctrl+C, etc.), verify behavior
4. **Verify**:
   - Feature exists and functions as documented
   - Implementation details match claims (e.g., "HMAC-SHA256" → verify actual HMAC usage in code)

**Example Verification**:

```
Given SECURITY.md documents audit logging:
  "Tamper-Evident: HMAC-SHA256 signatures prevent log modification"

When I run `pass-cli init --enable-audit`
And I inspect internal/audit implementation
Then audit.go uses crypto/hmac and crypto/sha256
And log entries contain HMAC field with 64-character hex signatures
And HMAC key stored in OS keychain (separate from vault)
```

**Pass Criteria**: Feature exists, implementation matches documented mechanism.

**Failure Recording**: Create DISC-### entry:
- **Severity**: Critical (feature doesn't exist), High (implementation differs), Medium (behavior incomplete)
- **Document**: Documented claim, actual implementation/behavior

---

## Category 6: Architecture Descriptions Verification

### Procedure

**For each architectural claim** in SECURITY.md, README.md (e.g., "modular design", "separation of concerns", package descriptions):

1. **Identify** claim (e.g., "vault operations in internal/vault")
2. **Inspect** codebase structure (`ls internal/`, check package godoc)
3. **Verify**:
   - Described packages exist
   - Package responsibilities match descriptions
   - Architectural patterns claimed (e.g., "library-first") are followed

**Example Verification**:

```
Given SECURITY.md documents:
  "Encryption operations isolated in internal/crypto package"

When I run `ls internal/`
Then internal/crypto directory exists
And crypto package contains AES-GCM, PBKDF2, HMAC implementations
And vault.go imports internal/crypto (not inline crypto operations)
```

**Pass Criteria**: Architectural claims match actual codebase structure.

**Failure Recording**: Create DISC-### entry:
- **Severity**: High (package doesn't exist), Medium (package responsibilities differ)
- **Document**: Documented architecture, actual structure

---

## Category 7: Metadata Verification

### Procedure

**For each metadata claim** (version numbers, dates, status labels):

1. **Identify** metadata (e.g., "Version: v0.0.1", "Last Updated: October 2025")
2. **Verify** against git:
   - Version: `git tag --list` (check latest tag matches documented version)
   - Dates: `git log --oneline --all -- [file]` (check last commit date matches "Last Updated")
   - Status: Compare "Production Ready" against actual release state
3. **Pass Criteria**: Metadata current as of audit date

**Example Verification**:

```
Given README.md shows "Version: v0.0.1 | Last Updated: October 2025"

When I run `git tag --list`
Then latest tag is v0.0.1 (matches)
When I run `git log --oneline -1 -- README.md`
Then commit date is October 2025 (matches)
```

**Pass Criteria**: All metadata current within acceptable tolerance (dates within same month).

**Failure Recording**: Create DISC-### entry:
- **Severity**: Low (metadata always low priority unless wildly wrong)
- **Document**: Documented metadata, actual value

---

## Category 8: Output Examples Verification

### Procedure

**For each documented command output example** (e.g., `✅ Credential added successfully!`, table formats):

1. **Identify** output claim
2. **Execute** corresponding command
3. **Compare** actual output to documented example:
   - Exact message match (for success/error messages)
   - Format match (for tables: column headers, separators, alignment)
4. **Verify**: Output format stable across platforms (Windows vs. macOS vs. Linux)

**Example Verification**:

```
Given USAGE.md shows output example:
  $ pass-cli list
  Service    Username           Last Used
  ─────────  ─────────────────  ─────────
  github     user@example.com   2 days ago
  aws        admin              Never

When I run `pass-cli list` in test vault
Then output shows table with headers: "Service", "Username", "Last Used"
And separator line uses ─ characters
And alignment matches (left-aligned text, right-aligned "Last Used")
```

**Pass Criteria**: Actual output matches documented format.

**Failure Recording**: Create DISC-### entry:
- **Severity**: High (format completely different), Medium (minor format change), Low (emoji/symbol difference)
- **Document**: Documented output, actual output

---

## Category 9: Cross-References Verification

### Procedure

**For each markdown link** in documentation:

1. **Extract** links: `[link text](target)` or `[link text](file.md#section)`
2. **Classify**:
   - Internal file reference (e.g., `docs/USAGE.md`)
   - Internal anchor reference (e.g., `#configuration`)
   - External URL (e.g., `https://github.com/...`)
3. **Verify**:
   - File references: `ls [target]` (file exists)
   - Anchor references: grep for `## Configuration` or `# Configuration` in target file
   - External URLs: (skip HTTP checks, note for manual review)
4. **Pass Criteria**: All internal references resolve

**Example Verification**:

```
Given README.md contains link:
  [USAGE.md](docs/USAGE.md#configuration)

When I check `ls docs/USAGE.md`
Then file exists
When I grep for "## Configuration" in docs/USAGE.md
Then heading exists at line 773
```

**Pass Criteria**: File exists and anchor heading found.

**Failure Recording**: Create DISC-### entry:
- **Severity**: Medium (file doesn't exist), Low (anchor doesn't exist but file valid)
- **Document**: Link target, actual state

---

## Category 10: Behavioral Descriptions Verification

### Procedure

**For each "when X, then Y" behavioral claim**:

1. **Identify** claim (e.g., "When password is weak, init rejects it")
2. **Design** test case to trigger condition X
3. **Execute** test
4. **Verify** behavior Y occurs

**Example Verification**:

```
Given USAGE.md claims:
  "When master password is weak, init displays strength indicator and warns user"

When I run `pass-cli init` with password "password123"
Then output shows "⚠  Password strength: Weak"
And init succeeds (warning, not rejection)
```

**Pass Criteria**: Claimed behavior occurs when condition met.

**Failure Recording**: Create DISC-### entry:
- **Severity**: High (behavior doesn't occur), Medium (behavior partially occurs)
- **Document**: Claim, actual behavior

---

## Verification Execution Order

**Recommended order** (efficiency + dependency):

1. **Category 7 (Metadata)**: Quick, no setup
2. **Category 9 (Cross-References)**: Quick, grep-based
3. **Category 1 (CLI Interface)**: Foundation for other tests
4. **Category 3 (File Paths)**: Code inspection
5. **Category 4 (Configuration)**: Code inspection + YAML parsing
6. **Category 6 (Architecture)**: Code inspection
7. **Category 2 (Code Examples)**: Requires test vault setup
8. **Category 8 (Output Examples)**: Uses test vault
9. **Category 5 (Feature Claims)**: Manual testing, most time-consuming
10. **Category 10 (Behavioral)**: Manual testing, integrates multiple categories

**Estimated Time**: 4-6 hours for full audit (14 commands × 10 categories × ~30 files)

---

**Next**: Execute verification procedures, populate audit-report.md with discrepancy findings.
