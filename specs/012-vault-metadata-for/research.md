# Research: Vault Metadata for Audit Logging

**Feature**: 012-vault-metadata-for
**Date**: 2025-10-20
**Purpose**: Resolve technical decisions for metadata file implementation

## Research Tasks

### 1. Metadata File Patterns in CLIs

**Question**: How do other command-line tools handle sidecar metadata files alongside primary data files?

**Findings**:
- **Git**: Uses `.git` directory for metadata, but our use case is file-level, not directory-level
- **Docker**: Uses JSON manifests (`manifest.json`) alongside image layers - similar pattern to our needs
- **Terraform**: Uses `.terraform.tfstate` alongside `.tf` files - good analog for vault.enc → vault.meta
- **VS Code**: Uses `.vscode/settings.json` for workspace metadata - centralized directory approach

**Pattern Identified**: Most tools use one of two patterns:
1. **Sidecar file** (same directory, related extension): `file.ext` → `file.meta` or `file.ext.meta`
2. **Centralized metadata directory**: `~/.app/metadata/<hash-of-file-path>`

**Recommendation**: Sidecar file pattern (`vault.enc` → `vault.meta`) for simplicity and discoverability.

---

### 2. JSON Schema Versioning

**Question**: What are best practices for versioning JSON configuration files?

**Findings**:
- **JSON Schema** spec recommends `$schema` URL for formal schemas, integer `version` for simple configs
- **Kubernetes** uses `apiVersion` (string like "v1", "v1beta1") - semantic but complex
- **Docker Compose** uses integer `version` field (1, 2, 3) - simple and effective
- **OpenAPI** uses `openapi: "3.0.0"` (semantic versioning string)

**Forward/Backward Compatibility Strategies**:
- **Additive changes**: New fields can be added without breaking old readers (JSON unmarshaling ignores unknown fields in Go)
- **Breaking changes**: Increment version number, implement migration logic or dual-format support
- **Unknown version handling**: Log warning, attempt best-effort parsing, fall back to safe defaults

**Recommendation**: Integer `version` field (currently `1`), additive changes only for minor updates, version increment for breaking changes.

---

### 3. Atomic File Writes

**Question**: How to ensure atomic writes for metadata updates to prevent corruption?

**Findings**:
- **POSIX**: `rename()` syscall is atomic if source and dest are on same filesystem
- **Windows**: `MoveFileEx()` with `MOVEFILE_REPLACE_EXISTING` is atomic (Go `os.Rename` uses this)
- **Go stdlib**: `os.Rename(src, dst)` provides atomic replacement on modern filesystems

**Write Pattern**:
```go
// 1. Write to temporary file
tmpPath := filepath.Join(dir, ".vault.meta.tmp")
err := ioutil.WriteFile(tmpPath, data, 0644)

// 2. Atomic rename (replaces existing file)
err = os.Rename(tmpPath, metadataPath)
```

**Edge Cases**:
- **Crash before rename**: Temporary file left behind (harmless, cleanup on next run)
- **Crash during rename**: Atomicity guarantees no partial write (either old or new file intact)
- **Concurrent writes**: Last writer wins (acceptable for single-vault model; metadata changes are rare)

**Recommendation**: Write to temp file + rename pattern for atomic updates.

---

### 4. Audit Log Self-Discovery

**Question**: How can the system auto-detect audit log configuration from the filesystem?

**Findings**:
- **Convention over configuration**: If `audit.log` exists in vault directory, assume audit was enabled
- **Heuristic validity**: User explicitly ran `--enable-audit` at some point to create audit.log
- **Safety**: Self-discovery is best-effort; won't create audit.log spontaneously (only if metadata OR previous audit setup exists)

**Fallback Logic**:
```go
// Pseudocode for fallback in NewVaultService
meta, err := LoadMetadata(vaultPath)
if meta == nil || err != nil {
    // No metadata or corrupted - try self-discovery
    auditLogPath := filepath.Join(filepath.Dir(vaultPath), "audit.log")
    if fileExists(auditLogPath) {
        // Found audit.log, enable audit at best-effort
        vaultService.EnableAudit(auditLogPath, vaultPath)
    }
}
```

**Limitations**:
- Can't distinguish "audit disabled" from "never had audit" when metadata missing
- Assumes audit.log belongs to this vault (valid for single-vault-per-directory model)
- Won't retroactively create audit.log if none exists (safe default)

**Recommendation**: Implement self-discovery as fallback when metadata missing/corrupted (P3 user story requirement).

---

## Decision Summary

| Decision Point | Choice | Rationale |
|---------------|--------|-----------|
| **File format** | JSON | Standard library, human-readable, versioned |
| **File location** | Same directory as vault.enc | Simple discovery, atomic cleanup |
| **Versioning** | Integer `version` field | Simple, matches Docker Compose pattern |
| **Atomic writes** | Temp file + rename | POSIX/Windows atomic guarantee |
| **Fallback strategy** | Metadata-first + self-discovery | Resilience (P3 user story) |
| **Metadata permissions** | 0644 (world-readable) | No sensitive data (FR-015), follows config file norms |

---

## Implementation Notes

**No External Dependencies**:
- All functionality uses Go standard library (`encoding/json`, `os`, `path/filepath`, `time`)
- Aligns with Constitution Principle VII (Simplicity & YAGNI)

**Cross-Platform Considerations**:
- `filepath.Join` for path construction (handles `/` vs `\`)
- `filepath.Abs` to ensure absolute paths in metadata
- `os.Rename` atomic behavior verified on Windows/macOS/Linux

**Security Review**:
- ✅ No sensitive data in metadata (vault path, boolean, log path only)
- ✅ Metadata file is plaintext and world-readable (by design, no secrets)
- ✅ Audit log itself is integrity-protected with HMAC (existing feature)

---

## References

- [Go os.Rename documentation](https://pkg.go.dev/os#Rename) - Atomic file replacement
- [JSON Schema versioning](https://json-schema.org/understanding-json-schema/reference/schema.html#schema) - Versioning best practices
- [Docker Compose version field](https://docs.docker.com/compose/compose-file/04-version-and-name/) - Integer versioning example
- [Terraform state file](https://www.terraform.io/docs/language/state/index.html) - Sidecar metadata pattern

**Research complete**. All unknowns from Technical Context resolved.
