# Implementation Plan: Vault Metadata for Audit Logging

**Branch**: `012-vault-metadata-for` | **Date**: 2025-10-20 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/012-vault-metadata-for/spec.md`

## Summary

Enable audit logging for keychain lifecycle operations that don't unlock the vault (`keychain status`, `vault remove`) by storing audit configuration in a plaintext metadata file (`vault.meta`) alongside the encrypted vault. This allows VaultService to initialize audit logging without requiring vault decryption, ensuring complete FR-015 compliance from spec 011. The system includes fallback self-discovery for resilience when metadata is missing or corrupted.

**Technical Approach**: Introduce a `VaultMetadata` struct and associated persistence functions that read/write JSON metadata files during vault initialization. Update VaultService constructor to load metadata (if present) and initialize audit logging before any vault operations. Implement hybrid approach with metadata-first and fallback self-discovery for maximum reliability.

## Technical Context

**Language/Version**: Go 1.21+ (existing codebase)
**Primary Dependencies**: Go standard library (`encoding/json`, `os`, `path/filepath`)
**Storage**: File-based (vault.meta JSON files, existing audit.log)
**Testing**: Go test framework (`testing` package, existing test infrastructure)
**Target Platform**: Cross-platform (Windows, macOS, Linux)
**Project Type**: Single project (CLI application)
**Performance Goals**: <50ms for metadata file operations (creation/read/write)
**Constraints**: Zero crashes on corrupted metadata, backward compatible with existing vaults, no sensitive data in plaintext
**Scale/Scope**: Single-vault model, metadata file <1KB, minimal performance overhead

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Security-First Development (NON-NEGOTIABLE)
- ✅ **No Secret Logging**: Metadata contains only non-sensitive audit config (vault path, boolean flags, log paths)
- ✅ **Forbidden Operations**: No credentials in metadata file (FR-015 explicit requirement)
- ✅ **Threat Modeling**: Feature enhances security by enabling audit trail for destructive operations (vault deletion)
- ⚠️ **Security Risk Assessment**: Plaintext metadata file could reveal audit log location to attacker with filesystem access
  - **Mitigation**: Metadata only reveals what's already discoverable (vault.enc and audit.log in same directory). No credential exposure. Audit log itself is integrity-protected with HMAC.
  - **Justification**: Audit capability for non-unlocking operations requires *some* external hint. Metadata is minimal disclosure.

### II. Library-First Architecture
- ✅ **Self-Contained Library**: New `internal/vault/metadata.go` with standalone VaultMetadata type
- ✅ **Single Purpose**: Metadata persistence separated from vault crypto and audit logging
- ✅ **No CLI Dependencies**: Metadata functions operate on paths and structs, no command-line concerns
- ✅ **Documented APIs**: Public functions `LoadMetadata()`, `SaveMetadata()`, `DeleteMetadata()` with clear contracts

### III. CLI Interface Standards
- ✅ **No Changes Required**: Feature is transparent to CLI (metadata managed internally by VaultService)
- ✅ **Exit Codes**: Metadata errors degrade gracefully (warnings logged, operation continues via fallback)
- ✅ **Script-Friendly**: Existing commands (`keychain status`, `vault remove`) gain audit capability without API changes

### IV. Test-Driven Development (NON-NEGOTIABLE)
- ✅ **TDD Workflow**: Tests written first for metadata CRUD operations, integration tests for audit logging with/without metadata
- ✅ **Test Coverage**: Minimum 80% for new `metadata.go` module
- ✅ **Test Types**:
  - Unit: metadata save/load/delete, JSON marshaling, version validation
  - Integration: VaultService initialization with/without metadata, fallback self-discovery
  - Security: verify no sensitive data in metadata file, graceful handling of corrupted files
- ✅ **Security Test Cases**: Verify FR-015 (no secrets in metadata), SC-006 (zero crashes on corruption)

### V. Cross-Platform Compatibility
- ✅ **Path Handling**: Use `filepath.Join`, `filepath.Abs` for all metadata file operations
- ✅ **Platform-Specific**: No platform-specific code needed (JSON format, standard library file I/O)
- ✅ **Testing Matrix**: Integration tests run on Windows, macOS, Linux (existing CI)

### VI. Observability & Auditability
- ✅ **Audit Trail**: Feature enables audit logging for `keychain status` and `vault remove` (FR-007, FR-008)
- ✅ **Structured Logging**: Warnings logged when metadata missing/corrupted (FR-009, FR-012, FR-016)
- ✅ **No Credential Logging**: Metadata only contains paths and boolean flags (FR-015)

### VII. Simplicity & YAGNI
- ✅ **Concrete User Need**: Addresses real FR-015 compliance gap from spec 011
- ✅ **Standard Library**: Uses only `encoding/json` and `os` packages (no external dependencies)
- ✅ **Flat Architecture**: Single new file `metadata.go` in existing `internal/vault` package
- ✅ **Direct Solution**: Straightforward JSON persistence, no abstractions or frameworks

**GATE RESULT**: ✅ PASSED (with justification for plaintext metadata security trade-off)

## Project Structure

### Documentation (this feature)

```
specs/012-vault-metadata-for/
├── plan.md              # This file
├── research.md          # Phase 0: Best practices for metadata patterns
├── data-model.md        # Phase 1: VaultMetadata structure
├── quickstart.md        # Phase 1: How to use metadata feature
├── contracts/           # Phase 1: VaultMetadata JSON schema
│   └── vault-metadata-schema.json
└── tasks.md             # Phase 2: (created by /speckit.tasks, not this command)
```

### Source Code (repository root)

```
pass-cli/
├── cmd/                      # CLI commands (existing)
│   ├── keychain_status.go    # [MODIFIED] Use metadata for audit
│   └── vault_remove.go       # [MODIFIED] Use metadata for audit
├── internal/                 # Internal library packages
│   ├── vault/                # Vault operations
│   │   ├── vault.go          # [MODIFIED] Load metadata in constructor
│   │   ├── metadata.go       # [NEW] Metadata persistence
│   │   └── metadata_test.go  # [NEW] Metadata unit tests
│   └── security/             # Audit logging (existing)
│       └── audit.go          # [UNCHANGED] No changes needed
└── test/                     # Integration tests
    ├── vault_metadata_test.go # [NEW] Metadata integration tests
    ├── keychain_status_test.go # [MODIFIED] Verify audit with metadata
    └── vault_remove_test.go   # [MODIFIED] Verify audit with metadata
```

**Structure Decision**: Single project structure (pass-cli CLI application). New code isolated to `internal/vault/metadata.go` with minimal changes to existing VaultService initialization logic. Follows existing pattern of vault management in `internal/vault/` package.

## Complexity Tracking

*No constitution violations requiring justification. Feature follows established patterns and adds minimal complexity.*

| Item | Assessment |
|------|------------|
| New file (`metadata.go`) | Justified: Separates metadata persistence from vault crypto (SRP) |
| Hybrid approach (metadata + fallback) | Justified: Provides resilience per P3 user story; fallback is <10 LOC |
| JSON format for metadata | Standard library, human-readable for debugging, industry norm for config files |

---

## Phase 0: Research & Decision Log

### Research Tasks

1. **Metadata File Patterns**: How do other CLIs handle sidecar metadata files?
2. **JSON Schema Best Practices**: Versioning and forward/backward compatibility
3. **File System Atomicity**: Ensuring atomic writes for metadata updates
4. **Audit Log Self-Discovery**: Patterns for auto-detecting configuration from filesystem

### Decision Log

#### Decision 1: Metadata File Format (JSON)

**Rationale**:
- Human-readable for debugging (users can inspect `vault.meta` in emergencies)
- Standard library support (`encoding/json`)
- Easy to version and extend (add fields without breaking older readers)
- Industry standard for configuration files (Docker, Kubernetes, VS Code, etc.)

**Alternatives Considered**:
- **TOML**: More readable, but requires external dependency (`github.com/BurntSushi/toml`)
- **Binary format**: Smaller, but not human-readable; violates Principle VII (simplicity)
- **Environment variables**: Cleaner for some use cases, but doesn't persist across sessions

**Choice**: JSON (matches Assumptions in spec: "Metadata file format uses JSON")

#### Decision 2: Metadata File Location (Same Directory as Vault)

**Rationale**:
- Simple discovery: `vault.enc` → `vault.meta` (replace extension)
- No configuration needed for metadata path
- Atomic cleanup: `vault remove` deletes both files together
- Mirrors existing pattern (`vault.enc` + `audit.log` in same directory)

**Alternatives Considered**:
- **Centralized metadata directory** (`~/.pass-cli/metadata/`): Complicates multi-vault scenarios, breaks when vaults moved
- **Embedded in vault filename** (`vault.enc.meta`): Non-standard, harder to parse

**Choice**: Same directory (matches FR-001: "located in same directory as `vault.enc`")

#### Decision 3: Fallback Self-Discovery Strategy

**Rationale**:
- Resilience: Handles metadata file corruption/deletion gracefully
- Backward compatibility: Old vaults without metadata still get best-effort audit logging
- Minimal code: Check `audit.log` existence, use vault path as ID

**Implementation**:
```go
// Pseudocode for fallback logic
if metadataFile.NotExist() || metadataFile.Corrupted() {
    auditLogPath := filepath.Join(vaultDir, "audit.log")
    if fileExists(auditLogPath) {
        vaultService.EnableAudit(auditLogPath, vaultPath) // Best-effort
    }
}
```

**Alternatives Considered**:
- **No fallback**: Fail hard when metadata missing → Rejected (violates SC-005, P3 user story)
- **Environment variable override**: Useful for power users, but doesn't solve general case

**Choice**: Hybrid (metadata-first with self-discovery fallback per P3 user story)

#### Decision 4: Metadata Versioning Strategy

**Rationale**:
- Include `version` field (integer) in metadata JSON
- Current version: `1`
- Future versions can add fields without breaking older readers (forward compatibility)
- Older readers ignore unknown fields (backward compatibility via JSON unmarshaling)

**Implementation**:
```json
{
  "version": 1,
  "vault_id": "/path/to/vault.enc",
  "audit_enabled": true,
  "audit_log_path": "/path/to/audit.log",
  "created_at": "2025-10-20T10:00:00Z"
}
```

**Alternatives Considered**:
- **No versioning**: Breaks when format changes → Rejected (violates long-term maintainability)
- **Semantic versioning** (`"version": "1.0.0"`): Overkill for simple config file

**Choice**: Integer version field (matches FR-002: "version (integer)")

#### Decision 5: Atomic File Writes

**Rationale**:
- Use write-to-temp + rename pattern for atomic updates
- Prevents partial writes if process crashes during metadata save

**Implementation**:
```go
// Pseudocode
tmpFile := filepath.Join(dir, ".vault.meta.tmp")
ioutil.WriteFile(tmpFile, jsonData, 0644)
os.Rename(tmpFile, metadataPath) // Atomic on POSIX, best-effort on Windows
```

**Alternatives Considered**:
- **Direct write**: Risks corruption on crash → Rejected (violates reliability principle)
- **Write-ahead log**: Overkill for <1KB config file

**Choice**: Temp file + rename (matches Assumption: "File system supports atomic file writes")

---

## Phase 1: Design & Contracts

### Data Model

See [data-model.md](data-model.md) for complete entity specifications.

**Key Entity**: `VaultMetadata`

```go
type VaultMetadata struct {
    VaultID       string    `json:"vault_id"`        // Absolute path to vault.enc
    AuditEnabled  bool      `json:"audit_enabled"`   // Whether audit logging is enabled
    AuditLogPath  string    `json:"audit_log_path"`  // Absolute path to audit.log
    CreatedAt     time.Time `json:"created_at"`      // Metadata creation timestamp
    Version       int       `json:"version"`         // Metadata format version (currently 1)
}
```

### API Contracts

See [contracts/vault-metadata-schema.json](contracts/vault-metadata-schema.json) for JSON schema.

**New Functions** (in `internal/vault/metadata.go`):

```go
// LoadMetadata reads vault.meta file and returns VaultMetadata
// Returns (nil, nil) if file doesn't exist (not an error)
// Returns error if file exists but is corrupted/invalid
func LoadMetadata(vaultPath string) (*VaultMetadata, error)

// SaveMetadata writes VaultMetadata to vault.meta file atomically
// Creates parent directory if needed
// Uses temp file + rename for atomic write
func SaveMetadata(meta *VaultMetadata, vaultPath string) error

// DeleteMetadata removes vault.meta file
// Returns nil if file doesn't exist (idempotent)
func DeleteMetadata(vaultPath string) error

// MetadataPath returns the path to vault.meta for given vault.enc path
func MetadataPath(vaultPath string) string
```

**Modified Functions** (in `internal/vault/vault.go`):

```go
// NewVaultService constructor updated to:
// 1. Load metadata file (if exists)
// 2. Initialize audit logger from metadata (if audit_enabled)
// 3. Fall back to self-discovery if metadata missing
func NewVaultService(vaultPath string) (*VaultService, error)

// EnableAudit updated to:
// 1. Enable audit logger (existing behavior)
// 2. Save metadata file with audit config
func (v *VaultService) EnableAudit(auditLogPath, vaultID string) error

// Unlock updated to:
// 1. Unlock vault (existing behavior)
// 2. Check for metadata/vault audit config mismatch
// 3. Update metadata if mismatch detected (vault settings take precedence)
func (v *VaultService) Unlock(password []byte) error

// RemoveVault updated to:
// 1. Load metadata (to get audit config)
// 2. Enable audit logger from metadata
// 3. Log vault_remove_attempt
// 4. Delete vault.enc
// 5. Log vault_remove_success/failure
// 6. Delete vault.meta
func (v *VaultService) RemoveVault(force bool) (*RemoveVaultResult, error)
```

### Integration Points

1. **VaultService Constructor** (`NewVaultService`):
   - Load metadata → Initialize audit → Ready for commands

2. **Keychain Status Command** (`cmd/keychain_status.go`):
   - No changes needed (VaultService already has audit logger initialized)

3. **Vault Remove Command** (`cmd/vault_remove.go`):
   - Update to call metadata cleanup after vault deletion

4. **Vault Init Command** (`cmd/init.go`):
   - Update to create metadata when `--enable-audit` flag used

### Quickstart

See [quickstart.md](quickstart.md) for detailed usage guide.

**For Users**:
- Metadata files are created/managed automatically
- No manual interaction required
- Audit logging "just works" for all operations

**For Developers**:
```bash
# Enable audit on existing vault
pass-cli keychain enable /path/to/vault.enc
# → Creates vault.meta automatically

# Check keychain status (now audited)
pass-cli keychain status /path/to/vault.enc
# → Audit entry written via metadata

# Remove vault (now audited)
pass-cli vault remove /path/to/vault.enc --yes
# → Audit entries written, then vault.meta deleted
```

---

## Next Steps

**This plan complete**. Next command: `/speckit.tasks` to generate task breakdown.

**Remaining Phases**:
- Phase 2: Task generation (`/speckit.tasks`)
- Phase 3: Implementation (execute tasks from tasks.md)
- Phase 4: Testing & verification
