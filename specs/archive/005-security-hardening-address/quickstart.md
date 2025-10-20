# Security Hardening Quickstart

**Feature**: Security Hardening
**Branch**: `005-security-hardening-address`
**For**: Developers implementing and testing security improvements

## Overview

This guide helps developers quickly start working on the security hardening feature. For comprehensive design details, see `data-model.md`. For research findings, see `research.md`.

---

## Development Setup

### Prerequisites

- Go 1.21+ (requires `crypto/subtle`)
- `golangci-lint` for linting
- `gosec` for security scanning
- `delve` debugger (for memory inspection tests)

### Clone and Build

```bash
cd R:\Test-Projects\pass-cli
git checkout 005-security-hardening-address
go build -o pass-cli.exe ./cmd/pass-cli
```

---

## Running Tests

### Unit Tests (Password Validation, HMAC, Memory Clearing)

```bash
# Run all tests
go test ./...

# Run security package tests specifically
go test -v ./internal/security/...

# Run with coverage
go test -cover ./internal/security/...
```

### Integration Tests (Vault Operations with New Security)

```bash
# Test vault initialization with 600k iterations
go test -v ./internal/vault -run TestInitialize

# Test password policy enforcement
go test -v ./internal/vault -run TestPasswordPolicy

# Test migration (100k → 600k iterations)
go test -v ./internal/vault -run TestMigration
```

### Security-Specific Tests (Memory Inspection, Crypto Timing)

```bash
# Memory clearing verification (requires delve)
cd tests/security
go test -v -run TestMemoryClear

# Crypto timing benchmark (should take 500-1000ms)
go test -bench=BenchmarkDeriveKey ./internal/crypto -benchtime=5s

# Expected output: ~500-1000ms per operation
```

### Cross-Platform Test Matrix (Constitution V)

```bash
# Windows
go test ./...

# Linux (via WSL or CI)
GOOS=linux go test ./...

# macOS (via CI)
GOOS=darwin go test ./...
```

### Performance Benchmarks (Crypto Operations)

Run comprehensive benchmarks to validate crypto performance:

```bash
# Run all crypto benchmarks with extended time
cd internal/crypto
go test -bench=. -benchmem -benchtime=10s

# Example output (AMD Ryzen AI 9 HX 370):
# BenchmarkCryptoService_DeriveKey-24    156    77532462 ns/op     772 B/op    10 allocs/op
# BenchmarkCryptoService_Encrypt-24      9119680    1332 ns/op    3600 B/op     5 allocs/op
# BenchmarkCryptoService_Decrypt-24      14289338    871.0 ns/op   2304 B/op     3 allocs/op
```

**Performance Summary (600k iterations, January 2025)**:

| Operation | Duration | Throughput | Memory | Target |
|-----------|----------|------------|--------|--------|
| **DeriveKey (600k)** | ~77.5ms | 12.9 ops/sec | 772 B | 500-1000ms (FR-009) |
| **Encrypt (1KB)** | ~1.3µs | 750k ops/sec | 3600 B | < 10µs |
| **Decrypt (1KB)** | ~0.9µs | 1.1M ops/sec | 2304 B | < 10µs |

**Notes**:
- **DeriveKey faster than target**: Modern CPUs (Ryzen 9 HX 370) complete PBKDF2 faster than the 500-1000ms target. This is acceptable per Spec Assumption 1 ("faster machines may complete faster").
- **Encrypt/Decrypt performance**: AES-256-GCM is extremely fast, well under 10µs for 1KB payloads.
- **Memory efficiency**: All operations use minimal allocations (3-10 allocs per op).

**Hardware Impact**:
- **Modern CPUs (2023+)**: 50-100ms for 600k iterations
- **Mid-range CPUs (2018-2022)**: 200-500ms for 600k iterations
- **Older CPUs (2015-2017)**: 500-1000ms for 600k iterations
- **Very old CPUs (<2015)**: May exceed 1000ms (acceptable per spec)

---

## Common Development Tasks

### Adding New Password Validation Rules

Edit `internal/security/password.go`:

```go
func (p *PasswordPolicy) Validate(password []byte) error {
    // Add your validation logic here
    // Return descriptive error per FR-016

    if containsCommonPasswords(password) {
        return errors.New("password is too common")
    }

    return nil
}
```

**Test it**:
```bash
go test -v ./internal/security -run TestPasswordPolicy
```

### Enabling Audit Logging (For Testing)

```bash
# Initialize vault with audit logging enabled
pass-cli init --enable-audit

# Set custom audit log location
set PASS_AUDIT_LOG=C:\temp\test-audit.log
pass-cli init --enable-audit

# View audit log
type C:\Users\%USERNAME%\.pass-cli\audit.log
```

### Adjusting PBKDF2 Iterations (Power Users)

```bash
# Set custom iteration count (must be >= 600,000)
set PASS_CLI_ITERATIONS=1000000
pass-cli init

# Verify iteration count in vault metadata
pass-cli info --json | jq .iterations
```

### Debugging Memory Clearing

Use `delve` to inspect memory after vault operations:

```bash
dlv debug ./cmd/pass-cli

# Set breakpoint after password use
(dlv) break internal/vault/vault.go:150

# Run unlock command
(dlv) continue

# Inspect memory at masterPassword field
(dlv) print v.masterPassword
# Should be zeroed: []byte{}
```

---

## Architecture Quick Reference

### Package Structure

```
internal/
├── crypto/              # PBKDF2, AES-GCM, memory clearing
│   └── crypto.go        # Refactored to accept []byte passwords
├── vault/               # Vault operations
│   └── vault.go         # Refactored: masterPassword is []byte
├── storage/             # Vault file I/O
│   └── storage.go       # Extended VaultMetadata with Iterations field
├── security/            # NEW: Security utilities
│   ├── memory.go        # ClearBytes, SecureBytes wrapper
│   ├── password.go      # PasswordPolicy validation (FR-011-017)
│   └── audit.go         # AuditLogEntry, HMAC logging (FR-019-026)

cmd/
├── cli/                 # CLI commands
│   └── helpers.go       # readPassword() returns []byte now
└── tui/                 # TUI interface
    └── components/
        └── forms.go     # Password strength indicator (tview)
```

### Key API Changes

```go
// OLD (string-based, insecure)
func readPassword() (string, error)
func DeriveKey(password string, salt []byte) ([]byte, error)
func Initialize(masterPassword string, useKeychain bool) error

// NEW (byte-based, secure)
func readPassword() ([]byte, error)
func DeriveKey(password []byte, salt []byte, iterations int) ([]byte, error)
func Initialize(masterPassword []byte, useKeychain bool) error
```

---

## Testing Checklist

Before marking a task complete, verify:

- [ ] **Unit tests pass**: `go test ./internal/security/...`
- [ ] **Integration tests pass**: `go test ./internal/vault/...`
- [ ] **Linting clean**: `golangci-lint run`
- [ ] **Security scan clean**: `gosec ./...`
- [ ] **Cross-platform (CI)**: Tests pass on Windows/macOS/Linux
- [ ] **Memory clearing verified**: Use delve to inspect memory after operations
- [ ] **Performance acceptable**: 500-1000ms key derivation (benchmark)
- [ ] **Backward compatibility**: Old vaults (100k iterations) still unlock
- [ ] **Constitution compliance**: Review against all 7 principles

---

## Troubleshooting

### "Vault file corrupted" after adding Iterations field

**Cause**: Old vaults don't have `Iterations` field in JSON

**Fix**: Backward compatibility loading (already implemented in plan)
```go
if metadata.Iterations == 0 {
    metadata.Iterations = 100000 // Legacy default
}
```

### Key derivation taking > 2 seconds

**Cause**: Iteration count too high or slow hardware

**Fix**: Check iteration count
```bash
pass-cli info --json | jq .iterations
```

Lower for development (not production):
```bash
set PASS_CLI_ITERATIONS=100000  # Development only!
```

### Memory still contains password after clearing

**Possible Causes**:
1. Go GC copied data before clearing (expected limitation, Spec Assumption 4)
2. String conversion somewhere (`string([]byte)` creates immutable copy)
3. Deferred `Clear()` not called (missing `defer`)

**Debug**:
```bash
# Use memory profiler
go test -memprofile=mem.out ./internal/vault
go tool pprof mem.out
(pprof) list masterPassword
```

### Audit log growing unbounded

**Cause**: Log rotation not triggering (FR-024)

**Fix**: Check file size threshold
```go
const MaxAuditLogSize = 10 * 1024 * 1024 // 10MB (FR-024 default)
```

Manually rotate:
```bash
move %USERPROFILE%\.pass-cli\audit.log %USERPROFILE%\.pass-cli\audit.log.old
```

---

## Implementation Phases (from research.md)

### Phase A: Memory Security Foundation (P1)
1. Extract `clearBytes` → public `ClearBytes`
2. Change `readPassword()` to return `[]byte`
3. Refactor `VaultService.masterPassword` → `[]byte`
4. Update `crypto.DeriveKey` to accept `[]byte`
5. Add deferred cleanup handlers

### Phase B: Cryptographic Hardening (P1)
6. Add `Iterations` field to `VaultMetadata`
7. Implement backward-compatible loading
8. Update `DeriveKey` with iteration parameter
9. Set new vaults to 600,000 iterations
10. Add migration in `ChangePassword`

### Phase C: Password Policy (P2)
11. Create `internal/security/password.go`
12. Implement complexity checks (FR-011-015)
13. Add strength calculation (FR-017)
14. Update vault init/change flows
15. Implement CLI/TUI strength indicators

### Phase D: Audit Logging (P3)
16. Create `internal/security/audit.go`
17. Add audit configuration (default: disabled)
18. Instrument vault operations
19. Implement log rotation
20. Add tamper detection verification

---

## Useful Commands

```bash
# Build and run
go build -o pass-cli.exe ./cmd/pass-cli && pass-cli.exe init

# Run specific test
go test -v -run TestPasswordValidation ./internal/security

# Benchmark crypto performance
go test -bench=BenchmarkDeriveKey -benchtime=10s ./internal/crypto

# Check test coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Security scan
gosec -quiet ./...

# Lint
golangci-lint run --timeout=5m

# Cross-compile for testing
GOOS=linux GOARCH=amd64 go build -o pass-cli-linux ./cmd/pass-cli
```

---

## Need Help?

- **Constitution**: See `.specify/memory/constitution.md` for project principles
- **Spec**: See `spec.md` for functional requirements
- **Research**: See `research.md` for technical decisions
- **Data Model**: See `data-model.md` for entity structures
- **Implementation Plan**: See `plan.md` for phase breakdown

---

**Happy Secure Coding!** 🔒
