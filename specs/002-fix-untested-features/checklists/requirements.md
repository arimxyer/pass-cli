# Requirements Checklist: Fix Untested Features and Complete Test Coverage

## Keychain Enable (cmd/keychain_enable.go)

- [ ] FR-001: System MUST update vault metadata file (.meta.json) when `keychain enable` runs successfully
- [ ] FR-002: System MUST set metadata field indicating keychain is enabled for this vault
- [ ] FR-003: `keychain enable` MUST verify password is correct before storing in keychain
- [ ] FR-004: `keychain enable` MUST handle vaults already using keychain (idempotent or clear error)
- [ ] FR-005: `keychain enable --force` MUST re-enable keychain even if already enabled (overwrites password)
- [ ] FR-006: `keychain enable` MUST write audit log entries when vault has audit enabled

## Keychain Status (cmd/keychain_status.go)

- [ ] FR-007: `keychain status` MUST check both keychain service AND vault metadata for consistency
- [ ] FR-008: `keychain status` MUST exit with code 0 even when keychain unavailable (informational command)
- [ ] FR-009: `keychain status` MUST display platform-specific backend name (wincred/macOS Keychain/Linux Secret Service)
- [ ] FR-010: `keychain status` MUST provide actionable suggestions when keychain not enabled
- [ ] FR-011: `keychain status` MUST write audit log entry when vault has audit enabled

## Vault Remove (cmd/vault_remove.go)

- [ ] FR-012: `vault remove` MUST delete vault file, metadata file, and keychain entry in single operation
- [ ] FR-013: `vault remove` MUST clean up orphaned keychain entries even when vault file missing
- [ ] FR-014: `vault remove` MUST prompt for confirmation unless --yes flag provided
- [ ] FR-015: `vault remove` MUST write audit entries (attempt + success) when vault has audit enabled
- [ ] FR-016: `vault remove` MUST write audit entries BEFORE deleting metadata file
- [ ] FR-017: `vault remove` MUST achieve 95% success rate across 20 consecutive runs (SC-003)

## Vault Service (internal/vault/vault.go)

- [ ] FR-018: `UnlockWithKeychain()` MUST check vault metadata to determine if keychain is enabled
- [ ] FR-019: `UnlockWithKeychain()` MUST return specific error when vault doesn't use keychain
- [ ] FR-020: `UnlockWithKeychain()` MUST return specific error when keychain password missing
- [ ] FR-021: Vault initialization with `--use-keychain` MUST set keychain flag in metadata

## TUI (cmd/tui/main.go)

- [ ] FR-022: TUI MUST attempt keychain unlock if vault metadata indicates keychain usage
- [ ] FR-023: TUI MUST fall back to password prompt only when keychain unlock fails
- [ ] FR-024: TUI MUST display clear error message when keychain unlock fails (not generic "unavailable")

## Test Coverage

- [ ] TestIntegration_KeychainEnable: Unskip "2_Enable_With_Password" test (line 68)
- [ ] TestIntegration_KeychainEnable: Unskip "3_Enable_Twice_Idempotent" test (line 105)
- [ ] TestIntegration_KeychainEnable: Unskip "4_Enable_Force_Flag" test (line 134)
- [ ] TestIntegration_KeychainStatus: Unskip "2_Status_Before_Enable" test (line 64)
- [ ] TestIntegration_KeychainStatus: Unskip "3_Enable_Keychain" test (line 95)
- [ ] TestIntegration_KeychainStatus: Unskip "4_Status_After_Enable" test (line 117)
- [ ] TestIntegration_VaultRemove: Unskip "2_Remove_With_Confirmation" test (line 70)
- [ ] TestIntegration_VaultRemove: Unskip "3_Remove_With_Yes_Flag" test (line 100)
- [ ] TestIntegration_VaultRemove: Unskip "4_Remove_Orphaned_Keychain" test (line 136)
- [ ] TestIntegration_VaultRemove: Unskip "5_Success_Rate_Test" test (line 178)

## Success Criteria

- [ ] SC-001: Users can enable keychain on existing vault and immediately use TUI/CLI without password prompts
- [ ] SC-002: All 25 TODO-skipped tests are unskipped and passing
- [ ] SC-003: Vault remove command achieves 95% success rate across 20 consecutive runs
- [ ] SC-004: Keychain status accurately reports state (0 false positives/negatives)
- [ ] SC-005: CI pipeline blocks merges containing `t.Skip("TODO:` patterns
- [ ] SC-006: Integration test suite completes with 100% pass rate
- [ ] SC-007: Documentation updated with working keychain/vault command examples

## Edge Cases Handled

- [ ] Keychain enable on vault already using keychain (idempotent/clear message)
- [ ] Vault remove on vault currently open/locked (graceful handling)
- [ ] Vault remove when keychain password exists but vault file deleted (cleanup orphaned keychain)
- [ ] Vault operations when file exists but keychain password missing (fall back to password prompt)
- [ ] Vault operations when metadata indicates keychain but keychain password missing (detect mismatch, clear error)
- [ ] Vault remove cancelled at confirmation prompt (exit cleanly with no changes)
- [ ] Keychain enable when system keychain locked/unavailable (platform-specific guidance)

## Notes

- Total functional requirements: 24 (FR-001 through FR-024)
- Total test cases to unskip: 10 (covering 25 individual test scenarios)
- Priority: P1 for keychain enable + test coverage, P2 for vault remove, P3 for status improvements
- All requirements must be completed before marking spec as Done
