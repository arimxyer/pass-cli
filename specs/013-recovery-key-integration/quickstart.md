# Quickstart: Recovery Key Integration

**Feature**: 013-recovery-key-integration
**Date**: 2025-12-05

## Prerequisites

- Go 1.21+ installed
- pass-cli repository cloned
- Existing tests passing: `go test ./...`

## Development Setup

```bash
# Clone and enter the repository
cd pass-cli

# Checkout the feature branch
git checkout 013-recovery-key-integration

# Verify build works
go build -o pass-cli .

# Run existing tests to establish baseline
go test ./...
```

## Key Files Modified

| File | Change Type | Purpose |
|------|-------------|---------|
| `internal/crypto/keywrap.go` | NEW | Key wrapping functions (GenerateDEK, WrapKey, UnwrapKey) |
| `internal/storage/storage.go` | MODIFY | Add WrappedDEK fields, version 2 support, MigrateToV2 |
| `internal/vault/metadata.go` | MODIFY | Update RecoveryMetadata version semantics |
| `internal/vault/vault.go` | MODIFY | InitializeWithRecovery, RecoverWithMnemonic, MigrateToV2 |
| `cmd/init.go` | MODIFY | Calls InitializeWithRecovery for v2 vaults (default) |
| `cmd/change_password.go` | MODIFY | Unwrap DEK for recovery, re-wrap on password change |
| `cmd/vault_migrate.go` | NEW | Migration command for v1 → v2 upgrade |

## Implementation Order

### Phase 1: Core Crypto (Foundation)

1. Create `internal/crypto/keywrap.go`:
   ```go
   // Start with these functions:
   func GenerateDEK() ([]byte, error)
   func WrapKey(dek, kek []byte) (WrappedKey, error)
   func UnwrapKey(wrapped WrappedKey, kek []byte) ([]byte, error)
   ```

2. Add tests in `internal/crypto/keywrap_test.go`:
   ```go
   func TestWrapUnwrapRoundTrip(t *testing.T)
   func TestUnwrapWithWrongKEK(t *testing.T)
   func TestNonceUniqueness(t *testing.T)
   ```

### Phase 2: Storage Layer

1. Update `internal/storage/storage.go`:
   - Add `WrappedDEK` and `WrappedDEKNonce` to `VaultMetadata`
   - Add `LoadVaultWithDEK()` for version 2 unlock
   - Keep `LoadVault()` backward compatible for version 1

2. Tests: Verify version 1 and version 2 vaults load correctly

### Phase 3: Vault Integration

1. Update `internal/vault/vault.go`:
   - Modify `Initialize()` to use key wrapping when recovery enabled
   - Add `UnlockWithRecovery()` method
   - Modify `ChangePassword()` to re-wrap DEK

2. Update `internal/vault/metadata.go`:
   - Version "2" semantics for `RecoveryMetadata.EncryptedRecoveryKey`

### Phase 4: CLI Commands

1. Update `cmd/init.go`:
   - Generate DEK and wrap with both KEKs
   - Store wrapped keys in metadata

2. Update `cmd/change_password.go`:
   - Recovery path: unwrap DEK, prompt new password, re-wrap
   - Normal path: unwrap with old, re-wrap with new

### Phase 5: Migration (Optional)

1. Add migration prompt in vault unlock
2. Generate new DEK, new mnemonic
3. Atomic re-encryption

## Testing Workflow

### Unit Tests

```bash
# Run all crypto tests
go test -v ./internal/crypto/...

# Run specific key wrapping tests
go test -v ./internal/crypto/ -run TestWrap
```

### Integration Tests

```bash
# Full recovery flow test
go test -v ./test/ -run TestRecoveryFlow

# Migration test
go test -v ./test/ -run TestMigration
```

### Manual Testing

```bash
# Build fresh binary
go build -o pass-cli .

# Test new vault with recovery (v2 format - default)
rm -rf ~/.pass-cli  # Clean slate
./pass-cli init     # Create vault, writes 24-word recovery phrase
                    # Optionally add passphrase (25th word)
                    # Verify backup by entering 3 words

# Test recovery works
./pass-cli change-password --recover
# Enter 6 challenge words → should succeed and prompt for new password

# Test migration from v1 to v2
rm -rf ~/.pass-cli
./pass-cli init --no-recovery  # Create v1 vault (no recovery)
./pass-cli vault migrate       # Migrate to v2, generates new recovery phrase
./pass-cli change-password --recover  # Verify recovery works
```

## Debugging Tips

### Verify DEK Generation

```go
// Add temporary logging (REMOVE before commit!)
dek, _ := GenerateDEK()
fmt.Printf("DEK length: %d\n", len(dek)) // Should be 32
```

### Verify Wrap/Unwrap

```go
// Test round-trip
wrapped, _ := WrapKey(dek, kek)
unwrapped, err := UnwrapKey(wrapped, kek)
if !bytes.Equal(dek, unwrapped) {
    t.Error("round-trip failed")
}
```

### Check Vault Version

```bash
# Inspect vault metadata
cat ~/.pass-cli/vault.enc | jq '.metadata.version'
# Should be 2 for new key-wrapped vaults
```

## Common Issues

### "invalid key length" Error

- Ensure DEK and KEK are exactly 32 bytes
- Check PBKDF2 key derivation output length

### "decryption failed" on Recovery

- Verify `RecoveryMetadata.Version` is "2"
- Check that DEK was wrapped with recovery KEK during init
- Ensure challenge words reconstruct correct mnemonic

### Version 1 Vault Not Migrating

- Use `pass-cli vault migrate` command to migrate v1 vaults
- Must unlock vault with current password first
- Migration generates a new recovery phrase (write it down!)
- Check `VaultMetadata.Version` in vault file

## Security Checklist

Before committing:

- [ ] All `defer crypto.ClearBytes()` calls in place for DEK/KEK
- [ ] No plaintext DEK logged or written to disk
- [ ] Tests verify memory clearing
- [ ] `gosec ./...` passes
- [ ] Error messages don't leak key material

## References

- [spec.md](./spec.md) - Feature requirements
- [research.md](./research.md) - Technical decisions
- [data-model.md](./data-model.md) - Data structures
- [contracts/keywrap.md](./contracts/keywrap.md) - API contract
- [Constitution v1.2.0](../../.specify/memory/constitution.md) - Security requirements
