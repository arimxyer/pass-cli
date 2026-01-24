### 1. Communication Standards

**Be concise and direct**:
- Avoid preamble like "Great!", "Sure!", "Let me help"
- State facts and actions clearly
- Only explain when complexity requires it

**When reporting progress**:
- Use file paths with line numbers: `cmd/tui/model.go:54`
- Show before/after for changes
- Confirm completion, don't elaborate unless asked

### 2. Committing Work

**Commit frequently** - after completing tasks, milestones, or before switching context.

**Format**: `<type>: <description>` (feat:, fix:, docs:, refactor:, chore:)
**Footer**: `Co-Authored-By: Claude <noreply@anthropic.com>`

### 3. Accuracy and Transparency

**NEVER**:
- Claim a task is complete when it's only partially done
- Mark a task as completed if tests are failing
- Hide errors or issues you encounter

**ALWAYS**:
- Report the actual state of work, not aspirational state
- Test thoroughly before marking tasks complete
- If you cannot complete a task, explain why clearly

### 4. Handling Errors and Blockers

**When compilation fails**:
1. Read the error message carefully
2. Identify which layer is affected
3. Fix at the source, not with workarounds

**When tests fail**:
1. Run individually: `go test -v ./path/to/package -run TestName`
2. Check if test needs updating for current framework/patterns
3. Fix implementation OR update test (whichever is wrong)

---

# pass-cli Development Guidelines

## Active Technologies
- Cobra (CLI framework)
- Viper (configuration management)
- go-keyring (OS keychain integration)
- Go 1.21+

## Project Structure

```
pass-cli/
├── cmd/                      # CLI commands (Cobra-based)
│   ├── tui/                  # TUI components (rivo/tview)
│   ├── root.go               # Root command and global flags
│   └── ...                   # Individual command files
├── internal/                 # Internal library packages
│   ├── vault/                # Vault operations and credential management
│   ├── crypto/               # Encryption/decryption (AES-GCM, password clearing)
│   ├── keychain/             # OS keychain integration (Windows/macOS/Linux)
│   ├── security/             # Audit logging with HMAC signatures
│   ├── storage/              # File operations
│   ├── config/               # Configuration handling
│   └── health/               # Health checks for doctor command
├── test/                     # Integration and unit tests
│   ├── unit/                 # Unit tests
│   ├── integration/          # Integration tests
│   └── helpers/              # Test utilities
├── docs/                     # Documentation
├── main.go                   # Application entry point
└── go.mod                    # Go module dependencies
```

**Architecture**: Library-first design. CLI commands (`cmd/`) are thin wrappers that delegate to `internal/` packages.

## Commands

```bash
# Use mise tasks for all CLI operations
mise tasks                    # List available tasks
mise run test                 # Run unit tests
mise run test:integration     # Run integration tests
mise run lint                 # Run linter
mise run build                # Build binary
mise run git <args>           # Git operations
mise run gh <args>            # GitHub CLI operations
```

## Code Style

**Password Handling**:
- Use `[]byte` type (never `string`)
- Apply `defer crypto.ClearBytes(password)` immediately after allocation

**Error Handling**:
- Wrap errors with context: `fmt.Errorf("failed to unlock vault: %w", err)`
- Graceful degradation (e.g., keychain unavailable should not crash)

**Testing**:
- Unit tests: `internal/` packages
- Integration tests: `test/integration/` with `-tags=integration`
- Use `runtime.GOOS` for platform-specific test behavior

## Platform-Specific Gotchas

**macOS keychain**: Tests that override `HOME` env var break keychain access (tied to user session). Use `runtime.GOOS != "darwin"` checks before setting fake HOME in integration tests.

## Cobra Patterns

**Config loading order**: `cobra.OnInitialize` runs BEFORE flags are parsed. For flag-dependent config (like `--config`), use `PersistentPreRunE` instead.

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
