### 1. Communication Standards

**Be concise and direct**:
- Avoid preamble like "Great!", "Sure!", "Let me help"
- State facts and actions clearly
- Only explain when complexity requires it

**When reporting progress**:
- Use file paths with line numbers: `cmd/tui/model.go:54`
- Show before/after for changes
- Confirm completion, don't elaborate unless asked

### 2. Committing Work During Specs

**Commit frequently and often** when working through spec tasks:

**When to commit**:
- After completing each task
- After completing each phase of a spec
- After any significant milestone or working state
- Before switching to a different task
- When you update relevant spec documentation

**Commit message format**:
```
<type>: <description>

<body explaining changes>

<phase reference if applicable>

Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>
```

**Examples**:
```
feat: Integrate tview view implementations into Model struct

- Update Model view field types to tview variants
- Update NewModel() to use tview view constructors
- Fix view method calls for tview compatibility

Phase 1 of tview-migration-remediation spec.

Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>
```

**Why commit frequently**:
- Enables easy rollback to working states
- Provides clear audit trail of implementation
- Allows atomic changes that can be reviewed independently
- Demonstrates systematic progress through spec tasks

### 3. Accuracy and Transparency (CRITICAL)

**Accurate assessments and transparency are the #1 priority in this repository.**

**NEVER**:
- Claim a task is complete when it's only partially done
- Mark a task as completed if tests are failing
- Skip steps in a task to save time
- Take shortcuts that deviate from the spec
- Implement differently than the spec describes
- Ignore acceptance criteria in spec
- Hide errors or issues you encounter

**ALWAYS**:
- Report the actual state of work, not aspirational state
- If you discover incomplete work, STOP and document the gap
- If you cannot complete a task, explain why clearly
- If a spec has errors, surface them immediately
- Follow the spec exactly as written - no interpretation
- Execute all steps in a task, even if they seem redundant
- Test thoroughly before marking tasks complete

**If a spec exists, you MUST follow it with NO QUESTIONS ASKED, ONLY EXECUTION:**

The spec represents deliberate planning and design. Thoroughness and time was taken to create the spec related documentation, so thoroughness and time should be taken when implementing the spec.

**No shortcuts. No deviations. No assumptions.**

If you think the spec is wrong, unclear, or could be improved:
1. **STOP implementation**
2. Document the specific issue
3. Ask the user for clarification or correction
4. Wait for spec update and approval
5. THEN continue implementation

**Do not reinterpret, optimize, or "improve" the spec on your own.** Execute it exactly as written.

### 4. Handling Errors and Blockers

**When compilation fails**:
1. Read the error message carefully
2. Identify which layer is affected
3. Check if it's a type mismatch (common during migration)
4. Fix at the source, not with workarounds

**When tests fail**:
1. Run individually: `go test -v ./path/to/package -run TestName`
2. Check if test needs updating for current framework/patterns
3. Fix implementation OR update test (whichever is wrong)

**When stuck on a task**:
1. Re-read the task field
2. Check the spec-docs for existing code to reference
3. Read the relevant spec-docs to understand acceptance criteria
4. Search codebase for similar patterns (use Grep, Glob, Search, or MCP-server tools if relevant)

**When discovering incomplete work**:
1. **STOP immediately** - Don't continue building on broken foundation
2. Document the gap (what was claimed vs. what exists)
3. Create remediation plan
4. Get user approval before proceeding

---

## Summary: Your Responsibilities

**ALWAYS**:
- Read docs before creating specs
- Follow specs exactly as written - NO shortcuts, NO deviations
- Report accurate state of work - transparency is #1 priority
- Commit frequently (after each task, phase, milestone)
- Update related documents when changing dependencies/frameworks
- Update related documents when changing file organization
- Respect architectural layers (never mix)
- Write tests for all new code
- Follow security standards strictly
- Be concise and direct in communication
- Update task checkboxes as you progress

**NEVER**:
- Take shortcuts or skip steps in spec tasks
- Mark tasks complete if tests are failing, or shortcuts were taken
- Reinterpret or "improve" specs on your own
- Mix architectural layers
- Store sensitive data insecurely
- Skip testing or security checks
- Work on multiple specs simultaneously
- Claim work is done when it's only partial

**Critical Rule**: If a spec exists, follow it exactly. No questions asked, only execution.

**When in doubt**: Read the spec docs, check existing patterns, and ask the user a question if the information isn't 100% clear according to the existing documentation.

---

# pass-cli Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-20

## Active Technologies
- Go 1.25.1 (existing codebase) + Go standard library (`encoding/json`, `os`, `path/filepath`, `net/http`) (011-doctor-command-for)
- Cobra (CLI framework), go-keyring (keychain), Viper (config), term (TTY detection) (011-doctor-command-for)
- File-based (vault.meta JSON files, existing audit.log) (012-vault-metadata-for)
- Go 1.21+ (existing codebase) (011-keychain-lifecycle-management)

## Project Structure

```
pass-cli/
├── cmd/                      # CLI commands (Cobra-based)
│   ├── tui/                  # TUI components (charmbracelet/bubbletea)
│   ├── root.go               # Root command and global flags
│   ├── init.go               # Vault initialization
│   ├── add.go                # Add credential
│   ├── get.go                # Retrieve credential
│   ├── update.go             # Update credential
│   ├── delete.go             # Delete credential
│   ├── list.go               # List credentials
│   ├── generate.go           # Generate password
│   ├── change_password.go    # Change vault master password
│   ├── verify_audit.go       # Verify audit log integrity
│   ├── config.go             # Configuration management
│   ├── version.go            # Version information
│   └── helpers.go            # Shared command helpers
├── internal/                 # Internal library packages
│   ├── vault/                # Vault operations and credential management
│   ├── crypto/               # Encryption/decryption (AES-GCM, password clearing)
│   ├── keychain/             # OS keychain integration (Windows/macOS/Linux)
│   ├── security/             # Audit logging with HMAC signatures
│   ├── storage/              # File operations
│   └── config/               # Configuration handling
├── test/                     # Integration and unit tests
│   ├── tui/                  # TUI integration tests
│   ├── unit/                 # Unit tests
│   ├── integration_test.go   # CLI integration tests
│   ├── keychain_integration_test.go
│   └── tui_integration_test.go
├── specs/                    # Feature specifications (Speckit framework)
├── docs/                     # Documentation
├── .specify/                 # Speckit framework configuration
├── main.go                   # Application entry point
└── go.mod                    # Go module dependencies
```

**Architecture**: Library-first design (Constitution Principle II). CLI commands (`cmd/`) are thin wrappers that delegate to `internal/` packages. Single-vault model with multi-location usage tracking per credential.

## Commands

### Building
```bash
go build -o pass-cli .              # Build binary
go install .                        # Install to GOPATH
```

### Testing
```bash
go test ./...                       # Run all tests
go test -race ./...                 # Run with race detection
go test -v -tags=integration -timeout 5m ./test  # Integration tests
go test -coverprofile=coverage.out ./...         # Coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Code Quality
```bash
go fmt ./...                        # Format code
go vet ./...                        # Static analysis
golangci-lint run                   # Linting (comprehensive)
gosec ./...                         # Security scanning
govulncheck ./...                   # Vulnerability checking
```

### Pre-Commit Checks
```bash
go fmt ./...
go vet ./...
golangci-lint run
go test -race ./...
gosec ./...
```

## Code Style

**General Principles**:
- Follow Go best practices and idioms (Effective Go, Go Code Review Comments)
- Library-first architecture: business logic in `internal/`, CLI in `cmd/`
- Security-first: No credentials logged, memory cleared with `crypto.ClearBytes()`, audit logging for all operations
- Test-driven development (TDD): Write tests before implementation

**Naming Conventions**:
- Packages: lowercase, no underscores (e.g., `keychain`, `vault`, `crypto`)
- Exported types: PascalCase (e.g., `VaultService`, `UsageRecord`)
- Unexported functions: camelCase (e.g., `readPassword`, `logAudit`)
- Error variables: Prefix with `Err` (e.g., `ErrVaultLocked`, `ErrCredentialNotFound`)

**Password Handling**:
- Use `[]byte` type (never `string`)
- Apply `defer crypto.ClearBytes(password)` immediately after allocation
- Example pattern:
  ```go
  password, err := readPassword()
  if err != nil { return err }
  defer crypto.ClearBytes(password)  // CRITICAL: clear on all paths
  ```

**Error Handling**:
- Wrap errors with context: `fmt.Errorf("failed to unlock vault: %w", err)`
- Platform-specific error messages for keychain operations (Windows/macOS/Linux)
- Graceful degradation (e.g., keychain unavailable should not crash)

**Testing**:
- Unit tests: `internal/` packages
- Integration tests: `test/` directory with real vault files and keychain operations
- Security tests: Verify audit logging, password clearing, no credential leakage
- Test tags: `-tags=integration` for integration tests

**Commit Messages**:
- Format: `<type>: <description>` (e.g., `feat:`, `fix:`, `docs:`, `refactor:`)
- Include body for non-trivial changes
- Reference spec phases when implementing specs
- Footer: `Generated with Claude Code\n\nCo-Authored-By: Claude <noreply@anthropic.com>`

## Recent Changes
- 011-doctor-command-for: Added Go 1.25.1 + stdlib (net/http for version check), Cobra/go-keyring/Viper/term dependencies
- 012-vault-metadata-for: Added Go 1.21+ (existing codebase) + Go standard library (`encoding/json`, `os`, `path/filepath`)
- 011-keychain-lifecycle-management: Added Go 1.21+ (existing codebase)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
