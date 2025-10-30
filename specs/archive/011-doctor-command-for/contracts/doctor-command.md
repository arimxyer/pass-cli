# Command Contract: `pass-cli doctor`

**Command**: `doctor`
**Purpose**: Comprehensive health check for vault, configuration, and system state
**Library**: `internal/health/`
**CLI**: `cmd/doctor.go`

---

## Command Signature

```bash
pass-cli doctor [flags]
```

---

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output results as JSON instead of human-readable format |
| `--quiet` | bool | `false` | No output to stdout/stderr, exit code only (script-friendly) |
| `--verbose` | bool | `false` | Show detailed check execution logging (no secrets) |

**Flag Interactions**:
- `--json` and `--quiet` are mutually exclusive (both suppress human output, prefer `--quiet`)
- `--verbose` is ignored when `--quiet` is set
- `--verbose` works with `--json` (adds `execution_log` field to JSON)

---

## Input

**Environment**:
- `HOME` / `%USERPROFILE%`: Used to locate default vault and config paths
- `PASS_CLI_VAULT`: Optional environment variable overriding default vault location (not used for first-run detection)

**Files** (read-only, no modifications):
- `~/.pass-cli/vault`: Default vault file (checked for presence and accessibility)
- `~/.pass-cli/config.yaml`: Config file (validated for syntax and values)
- `~/.pass-cli/*.backup`: Backup files (detected and analyzed)
- System keychain: Queried for `pass-cli:*` entries

**Network** (optional):
- GitHub API: `https://api.github.com/repos/USER/pass-cli/releases/latest` (1-second timeout)

---

## Output

### Human-Readable Format (default)

**stdout**:
```
Pass-CLI Health Check Report
════════════════════════════════════════

✅ Binary Version: v1.2.3 (up to date)
✅ Vault File: Present and accessible at ~/.pass-cli/vault
⚠️  Config File: Warning - clipboard_timeout (500s) exceeds maximum (300s)
   → Recommendation: Edit ~/.pass-cli/config.yaml, set clipboard_timeout to 30

❌ Keychain: 2 orphaned entries detected
   → vault1: /home/user/old-vault (file not found)
   → vault2: /tmp/test-vault (file not found)
   → Recommendation: Run 'pass-cli keychain cleanup' to remove orphaned entries

✅ Backup Files: No abandoned backups

Summary: 3 checks passed, 1 warning, 1 error
```

**stderr**: None (errors are reported in stdout with ❌ prefix)

**Color Scheme**:
- ✅ Green (`color.FgGreen`): Pass
- ⚠️  Yellow (`color.FgYellow`): Warning
- ❌ Red (`color.FgRed`): Error

---

### JSON Format (`--json`)

**stdout**:
```json
{
  "summary": {
    "passed": 3,
    "warnings": 1,
    "errors": 1,
    "exit_code": 2
  },
  "checks": [
    {
      "name": "version",
      "status": "pass",
      "message": "v1.2.3 (up to date)",
      "recommendation": "",
      "details": {
        "current": "v1.2.3",
        "latest": "v1.2.3",
        "update_url": "https://github.com/USER/pass-cli/releases/tag/v1.2.3",
        "up_to_date": true,
        "check_error": ""
      }
    },
    {
      "name": "config",
      "status": "warning",
      "message": "clipboard_timeout (500s) exceeds maximum (300s)",
      "recommendation": "Edit ~/.pass-cli/config.yaml, set clipboard_timeout to 30",
      "details": {
        "path": "/home/user/.pass-cli/config.yaml",
        "exists": true,
        "valid": true,
        "errors": [
          {
            "key": "clipboard_timeout",
            "problem": "value out of range",
            "current": "500",
            "expected": "5-300"
          }
        ],
        "unknown_keys": []
      }
    }
  ],
  "timestamp": "2025-10-21T14:32:00Z"
}
```

**stderr**: None

---

### Quiet Mode (`--quiet`)

**stdout**: (empty)
**stderr**: (empty)
**Exit Code**: 0=healthy, 1=warnings, 2=errors

---

## Exit Codes

| Code | Meaning | Scenario |
|------|---------|----------|
| `0` | Healthy | All checks passed (no warnings or errors) |
| `1` | Warnings | Non-critical issues detected (e.g., outdated version, old backups) |
| `2` | Errors | Critical issues detected (e.g., vault inaccessible, config invalid, orphaned keychain entries) |
| `3` | Reserved | Future use: Security-specific errors |

**Exit Code Priority**: If both warnings and errors exist, return `2` (errors take precedence).

---

## Health Checks Performed

### 1. Binary Version Check
- **Pass**: Current version matches latest GitHub release
- **Warning**: Newer version available on GitHub
- **Error**: Network timeout or GitHub API error (only if offline check fails)

**Details**: `VersionCheckDetails`

---

### 2. Vault File Check
- **Pass**: Vault file exists, is readable, has correct permissions (`0600` on Unix)
- **Warning**: Vault exists but has overly permissive permissions (e.g., `0644`)
- **Error**: Vault file missing or inaccessible

**Details**: `VaultCheckDetails`

---

### 3. Config File Check
- **Pass**: Config file exists, valid YAML, all values in expected ranges
- **Warning**: Unknown keys detected (possible typos), or values outside recommended ranges
- **Error**: Config file unparsable (invalid YAML syntax)

**Details**: `ConfigCheckDetails`

---

### 4. Keychain Status Check
- **Pass**: Keychain available, master password stored for current vault, no orphaned entries
- **Warning**: Keychain available but no password stored for current vault
- **Error**: Orphaned keychain entries detected (entries for deleted vaults)

**Details**: `KeychainCheckDetails`

**Orphan Detection Logic**:
1. Query keychain for all entries with `pass-cli:` prefix
2. Extract vault path from entry key (e.g., `pass-cli:/home/user/vault` → `/home/user/vault`)
3. Check if vault file exists at extracted path
4. If file doesn't exist, mark entry as orphaned

---

### 5. Backup File Check
- **Pass**: No backup files found, or only recent backups (<24 hours)
- **Warning**: Old backup files detected (>24 hours), may indicate interrupted operation
- **Error**: (Not used for backup checks, only warnings)

**Details**: `BackupCheckDetails`

---

## Error Handling

### Network Errors (Version Check)
- **Behavior**: Skip version check gracefully, report as "Unable to check for updates (offline)"
- **Status**: `pass` (not `error`)
- **Details**: `check_error` field contains network error message

### Keychain Access Denied
- **Behavior**: Report as warning, continue other checks
- **Status**: `warning`
- **Message**: "Keychain unavailable (access denied)"

### Config File Missing
- **Behavior**: Report as warning (config is optional)
- **Status**: `warning`
- **Message**: "Config file not found (using defaults)"

---

## Implementation Notes

### Timeout Handling
- Version check: 1-second timeout via `http.Client{Timeout: 1 * time.Second}`
- Other checks: No timeout (all local filesystem/keychain operations)

### Parallel Execution
- All checks run sequentially (no need for parallelism with <5s total time)
- Future optimization: Run checks in parallel with `sync.WaitGroup` if performance becomes an issue

### No Vault Decryption
- Doctor NEVER requires master password
- Vault checks only verify file metadata (existence, permissions, size)
- No vault contents are read or logged

### Security Guarantees
- No credential logging (vault contents, master password, keychain entries)
- Verbose mode only logs check names and execution times (no sensitive data)
- Audit log entry: `{"event": "doctor", "timestamp": "...", "status": "completed"}` (no check results logged)

---

## Testing Contracts

### Unit Tests (`internal/health/`)
- `TestVersionCheck_UpToDate`: Current == Latest → Pass
- `TestVersionCheck_UpdateAvailable`: Current < Latest → Warning
- `TestVersionCheck_NetworkTimeout`: Offline → Pass (with check_error)
- `TestVaultCheck_Exists`: Vault present, readable → Pass
- `TestVaultCheck_Missing`: No vault file → Error
- `TestConfigCheck_Valid`: Valid YAML, all keys correct → Pass
- `TestConfigCheck_InvalidValue`: clipboard_timeout=500 → Warning
- `TestKeychainCheck_Orphaned`: Keychain entry for deleted vault → Error
- `TestBackupCheck_OldBackup`: Backup >24h old → Warning

### Integration Tests (`test/doctor_test.go`)
- `TestDoctorCommand_Healthy`: All checks pass → Exit 0
- `TestDoctorCommand_JSON`: Output valid JSON, verify schema
- `TestDoctorCommand_Quiet`: No stdout/stderr, verify exit code
- `TestDoctorCommand_Offline`: Network unavailable → Version check skipped
- `TestDoctorCommand_NoVault`: Vault missing → Exit 2

---

## Example Usage

### Check vault health (human-readable)
```bash
$ pass-cli doctor
Pass-CLI Health Check Report
════════════════════════════════════════
✅ Binary Version: v1.2.3 (up to date)
✅ Vault File: Present and accessible
...
Summary: 5 checks passed, 0 warnings, 0 errors
```

### Script-friendly health check
```bash
$ pass-cli doctor --quiet
$ echo $?
0  # Healthy
```

### JSON output for monitoring
```bash
$ pass-cli doctor --json | jq '.summary.exit_code'
0
```

### Verbose mode (debugging)
```bash
$ pass-cli doctor --verbose
[14:32:00] Running version check...
[14:32:01] Version check completed (1.2s)
[14:32:01] Running vault check...
...
```

---

## Dependencies

**Library Dependencies**:
- `net/http`: GitHub API version check
- `os`, `path/filepath`: File existence and path handling
- `internal/keychain`: Cross-platform keychain access
- `internal/config`: Config file validation
- `time`: Backup file age calculation

**CLI Dependencies**:
- `github.com/spf13/cobra`: Command framework
- `github.com/fatih/color`: Colored output
- `encoding/json`: JSON output formatting

---

## Contract Validation

✅ Follows Constitution Principle III (CLI Interface Standards):
- Structured JSON output
- Consistent exit codes
- Script-friendly `--quiet` mode
- No credentials in output

✅ Follows Constitution Principle II (Library-First):
- All logic in `internal/health/`
- CLI is thin wrapper over library

✅ Follows Constitution Principle I (Security-First):
- No vault decryption required
- No credential logging
- No secrets in verbose mode
