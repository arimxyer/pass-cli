# Quickstart: Manual Vault Backup and Restore

**Feature**: Manual Vault Backup and Restore Commands
**Branch**: `001-add-manual-vault`
**Date**: 2025-11-11

## Overview

This guide helps developers quickly set up their environment, understand the feature scope, and begin implementation or testing.

## What This Feature Does

Adds three CLI commands to manually manage vault backups:
- `pass vault backup create` - Create timestamped manual backup
- `pass vault backup restore` - Restore vault from newest backup
- `pass vault backup info` - View backup status and history

**Key Distinction**: Exposes existing storage service backup methods via CLI. Minimal library changes required.

## Prerequisites

### Required

- Go 1.21 or higher
- Git (for version control)
- Terminal/shell access

### Recommended

- golangci-lint (for code quality checks)
- gosec (for security scanning)
- VS Code with Go extension (or preferred Go IDE)

## Quick Setup

### 1. Clone and Branch

```bash
# If not already on feature branch
git checkout 001-add-manual-vault

# Verify you're on the correct branch
git branch --show-current
# Should output: 001-add-manual-vault
```

### 2. Install Dependencies

```bash
# Install Go dependencies
go mod tidy

# Verify build works
go build -o pass-cli .

# Run existing tests to ensure baseline
go test ./...
```

### 3. Verify Existing Backup Methods

The storage service already has backup methods we'll expose:

```bash
# Check existing backup API
grep -n "func.*Backup" internal/storage/storage.go
```

Expected output shows:
- `CreateBackup() error`
- `RestoreFromBackup() error`
- `RemoveBackup() error`

## Project Structure

### Files to Create (implementation)

```
cmd/
├── vault_backup.go                  # Parent command for backup subcommands
├── vault_backup_create.go           # Create manual backup
├── vault_backup_restore.go          # Restore from backup
└── vault_backup_info.go             # View backup info

internal/storage/
├── backup.go                        # Manual backup naming logic
└── backup_test.go                   # Unit tests

test/
├── vault_backup_integration_test.go # Integration tests for all commands
└── vault_backup_info_test.go        # Info command specific tests
```

### Files to Read (context)

Before implementing, familiarize yourself with:

1. **Existing Commands** (pattern reference):
   - `cmd/vault.go` - Parent command example
   - `cmd/vault_remove.go` - Subcommand example

2. **Storage Service** (API to call):
   - `internal/storage/storage.go:425-633` - Existing backup methods

3. **Test Examples**:
   - `test/vault_remove_test.go` - Integration test pattern

## Development Workflow

### Phase 0: Research (Completed)

✅ All research documented in `research.md`

Key decisions:
- Manual backups use `vault.enc.[timestamp].manual.backup` naming
- Restore selects newest backup (automatic or manual)
- Info command shows all backups with disk usage warnings

### Phase 1: Design (Current)

✅ Data model defined in `data-model.md`
✅ CLI contracts defined in `contracts/`
✅ This quickstart guide created

### Phase 2: Implementation (Next)

Run `/speckit.tasks` to generate `tasks.md` with prioritized implementation tasks.

### TDD Workflow

Per Constitution Principle IV, follow test-driven development:

1. **Write Test First** (red)
   ```bash
   # Create test file
   touch test/vault_backup_integration_test.go

   # Write failing test
   # Run test to verify it fails
   go test ./test/vault_backup_integration_test.go
   ```

2. **Implement Feature** (green)
   ```bash
   # Create command file
   touch cmd/vault_backup_create.go

   # Implement until test passes
   go test ./test/vault_backup_integration_test.go
   ```

3. **Refactor** (refactor)
   ```bash
   # Improve code quality
   golangci-lint run cmd/vault_backup_create.go

   # Verify tests still pass
   go test ./test/vault_backup_integration_test.go
   ```

## Running Tests

### Unit Tests (library layer)

```bash
# Test backup naming logic
go test -v ./internal/storage/backup_test.go

# Test with coverage
go test -cover ./internal/storage/
```

### Integration Tests (CLI layer)

```bash
# Test all backup commands
go test -v -tags=integration ./test/vault_backup_integration_test.go

# Test specific command
go test -v -tags=integration -run TestBackupCreate ./test/
```

### Full Test Suite

```bash
# Run all tests
go test ./...

# With race detection
go test -race ./...

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Manual Testing

### Setup Test Vault

```bash
# Build the binary
go build -o pass-cli .

# Initialize test vault
./pass-cli init

# Add some test credentials
./pass-cli add test-service
# Enter dummy credentials
```

### Test Backup Commands

```bash
# Create manual backup
./pass-cli vault backup create

# View backup info
./pass-cli vault backup info

# Simulate vault corruption (CAREFUL: only in test environment)
echo "corrupted" > ~/.pass/vault.enc

# Restore from backup
./pass-cli vault backup restore

# Verify vault works
./pass-cli list
```

### Cleanup Test Data

```bash
# Remove test vault and backups
./pass-cli vault remove --force

# Or manually
rm -rf ~/.pass/
```

## Code Quality Checks

### Linting

```bash
# Run golangci-lint
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix
```

### Security Scanning

```bash
# Run gosec security scanner
gosec ./...

# Check specific directories
gosec ./cmd/ ./internal/storage/
```

### Formatting

```bash
# Format code
go fmt ./...

# Check formatting (CI)
test -z "$(go fmt ./...)"
```

## Debugging Tips

### Verbose Output

Use `--verbose` flag to see detailed operation progress:

```bash
./pass-cli vault backup create --verbose
./pass-cli vault backup restore --verbose
./pass-cli vault backup info --verbose
```

### Check Audit Logs

View audit trail for backup operations:

```bash
tail -f ~/.pass-cli/audit.log
```

### Inspect Backup Files

```bash
# List all backups
ls -lh ~/.pass/*.backup*

# Check file sizes
du -h ~/.pass/*.backup*

# Verify file permissions
stat ~/.pass/vault.enc.backup
```

## Common Issues

### Issue: Test Vault Conflicts with Production

**Solution**: Use separate directories for testing

```bash
# Set test vault path
export PASS_VAULT_DIR=/tmp/pass-test

# Run tests
go test ./test/
```

### Issue: Permission Denied Errors

**Solution**: Check vault directory permissions

```bash
# Fix permissions (Unix)
chmod 700 ~/.pass
chmod 600 ~/.pass/vault.enc*
```

### Issue: Tests Fail with "vault locked"

**Solution**: Clean up lock files from previous test runs

```bash
# Remove stale lock files
rm -f /tmp/pass-test-*/vault.lock
```

## Next Steps

1. **Generate Implementation Tasks**:
   ```bash
   # Run from repository root
   cd R:\Test-Projects\pass-cli

   # Generate tasks.md (using speckit)
   # This will be done via /speckit.tasks command
   ```

2. **Begin Implementation**:
   - Start with library layer (`internal/storage/backup.go`)
   - Then CLI commands (`cmd/vault_backup_*.go`)
   - Follow TDD: tests first, then implementation

3. **Commit Frequently**:
   - After each completed task
   - After each phase (library → CLI → tests)
   - Before switching contexts

## References

### Documentation

- **Specification**: `specs/001-add-manual-vault/spec.md`
- **Implementation Plan**: `specs/001-add-manual-vault/plan.md`
- **Research**: `specs/001-add-manual-vault/research.md`
- **Data Model**: `specs/001-add-manual-vault/data-model.md`
- **Contracts**: `specs/001-add-manual-vault/contracts/`

### Codebase

- **Storage Service**: `internal/storage/storage.go`
- **Vault Service**: `internal/vault/vault.go`
- **Existing Commands**: `cmd/vault*.go`
- **Test Examples**: `test/*_test.go`

### External

- **Cobra CLI Framework**: https://github.com/spf13/cobra
- **Go Testing**: https://pkg.go.dev/testing
- **Constitution**: `.specify/memory/constitution.md`

## Support

### Stuck? Check These First

1. **Constitution**: Review relevant principles (I-VII)
2. **CLAUDE.md**: Development workflow and commit standards
3. **Existing Code**: Similar commands for patterns
4. **Spec Documents**: Requirements and acceptance criteria

### Debug Commands

```bash
# Verify Go environment
go version
go env

# Check dependencies
go mod verify
go mod tidy

# Clean build cache
go clean -cache
go clean -testcache

# Rebuild from scratch
rm -f pass-cli pass-cli.exe
go build -v -o pass-cli .
```

## Getting Help

- **Project Issues**: Check existing GitHub issues
- **Go Documentation**: https://pkg.go.dev
- **Cobra Docs**: https://cobra.dev
- **Constitution**: `.specify/memory/constitution.md`
- **CLAUDE.md**: `CLAUDE.md` (runtime development guidance)
