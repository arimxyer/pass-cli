<!--
Sync Impact Report - Constitution Update
═══════════════════════════════════════════════════════════════

Version Change: [Initial/Template] → 1.0.0

Changes Summary:
- Initial ratification of Pass-CLI constitution
- Added 7 core principles specific to password manager development
- Defined security-first architecture and testing requirements
- Established governance and compliance framework

Modified Principles:
- NEW: I. Security-First Development (NON-NEGOTIABLE)
- NEW: II. Library-First Architecture
- NEW: III. CLI Interface Standards
- NEW: IV. Test-Driven Development (NON-NEGOTIABLE)
- NEW: V. Cross-Platform Compatibility
- NEW: VI. Observability & Auditability
- NEW: VII. Simplicity & YAGNI

Added Sections:
- Security Requirements (dedicated section for password manager constraints)
- Development Workflow (quality gates and review process)

Templates Requiring Updates:
- ✅ plan-template.md: Constitution Check section references verified
- ✅ spec-template.md: Requirements and acceptance criteria alignment verified
- ✅ tasks-template.md: Task categorization aligns with principles
- ⚠ Runtime guidance (CLAUDE.md): Already contains spec adherence rules - no changes needed

Follow-up TODOs:
- None (all placeholders filled)

Last Updated: 2025-10-09
═══════════════════════════════════════════════════════════════
-->

# Pass-CLI Constitution

## Core Principles

### I. Security-First Development (NON-NEGOTIABLE)

**All development decisions prioritize security over convenience, performance, or features.**

- **Encryption Standards**: AES-256-GCM with PBKDF2-SHA256 (600,000 iterations minimum) MUST be used for all credential storage. No weaker algorithms permitted.
- **No Secret Logging**: Passwords, master passwords, API keys, or any credential data MUST NEVER be logged, printed to stdout/stderr, or written to any file outside the encrypted vault.
- **Zero-Trust Clipboard**: Clipboard operations MUST clear after 30 seconds maximum. Credentials MUST NOT be stored in clipboard history.
- **Secure Memory Handling**: Sensitive data in memory MUST be zeroed after use. No credential caching in plaintext memory.
- **System Keychain Integration**: Master passwords MUST be stored only in OS-provided secure storage (Windows Credential Manager, macOS Keychain, Linux Secret Service).
- **Threat Modeling Required**: Any new feature touching credentials MUST include threat analysis before implementation.

**Rationale**: As a password manager, Pass-CLI is a high-value target. A single security failure compromises all user credentials. This principle is non-negotiable because the entire value proposition relies on trustworthy security.

### II. Library-First Architecture

**Every feature MUST start as a standalone library with clear interfaces before CLI integration.**

- Libraries MUST be self-contained and independently testable
- Each library MUST have a single, well-defined purpose (no organizational-only libraries)
- Libraries MUST NOT depend on CLI-specific concerns (flags, output formatting, user interaction)
- Public APIs MUST be documented with usage examples
- Library changes MUST maintain backward compatibility or follow semantic versioning

**Rationale**: Separation enables reuse in GUI/TUI frontends, scripting, and testing. Forces clear architectural boundaries.

### III. CLI Interface Standards

**All functionality MUST be accessible via CLI following consistent text I/O protocols.**

- **Input**: stdin, command-line arguments, or environment variables only
- **Output**: Structured data to stdout (JSON + human-readable formats)
- **Errors**: All errors to stderr with non-zero exit codes
- **Script-Friendly Modes**: Support `--quiet`, `--field`, `--masked`, `--no-clipboard` flags for automation
- **No Interactive Prompts in Pipes**: Detect non-TTY stdin and fail fast if interaction required
- **Consistent Exit Codes**: 0=success, 1=user error, 2=system error, 3=security error

**Rationale**: CLI is the primary interface. Consistent I/O enables reliable shell scripting and integration into developer workflows.

### IV. Test-Driven Development (NON-NEGOTIABLE)

**Test-first development is mandatory for all new features and bug fixes.**

- **Red-Green-Refactor**: Tests MUST be written → User approved → Tests MUST fail → Then implement → Tests pass → Refactor
- **No Implementation Without Tests**: Pull requests without corresponding tests WILL be rejected
- **Test Coverage Gates**: Minimum 80% code coverage for all non-trivial packages
- **Test Types Required**:
  - **Unit Tests**: All library functions, especially cryptographic operations
  - **Integration Tests**: CLI commands end-to-end with real vault files
  - **Contract Tests**: Public API stability (prevents breaking changes)
- **Security Test Cases**: Every security principle MUST have verification tests (e.g., clipboard clearing, memory zeroing)

**Rationale**: Password managers cannot afford regressions. TDD ensures security guarantees are tested and maintained. This is non-negotiable because untested security code is untrustworthy code.

### V. Cross-Platform Compatibility

**Pass-CLI MUST function identically across Windows, macOS, and Linux.**

- **Single Binary**: Compile to standalone executables for each platform (no runtime dependencies)
- **Platform-Specific Code**: Isolate in dedicated packages (e.g., `keychain_windows.go`, `keychain_darwin.go`)
- **Path Handling**: Use `filepath.Join` and OS-agnostic path operations throughout
- **Testing Matrix**: CI MUST run tests on Windows, macOS (Intel + ARM), and Linux (amd64 + arm64)
- **Home Directory Handling**: Respect `%USERPROFILE%` (Windows) and `$HOME` (Unix) conventions
- **Line Endings**: Handle CRLF and LF gracefully in all file operations

**Rationale**: Developers use diverse platforms. Broken platform support fragments the user base and undermines trust.

### VI. Observability & Auditability

**System behavior MUST be observable and auditable without compromising security.**

- **Structured Logging**: Use leveled logging (DEBUG, INFO, WARN, ERROR) for non-sensitive operations
- **Audit Trail**: Log vault access events (init, add, get, update, delete) with timestamps to `~/.pass-cli/audit.log`
- **No Credential Logging**: Audit logs MUST contain only operation types, timestamps, and credential names—NEVER the credentials themselves
- **Verbose Mode**: Support `--verbose` flag for debugging (MUST NOT output secrets even in verbose mode)
- **Usage Tracking**: Track credential access by working directory for usage analytics (helps users identify unused credentials)

**Rationale**: Users need visibility into system behavior for debugging and security audits. Logging discipline ensures observability without leaking secrets.

### VII. Simplicity & YAGNI

**Start simple. Add complexity only when justified by concrete user needs.**

- **No Speculative Features**: Implement only what users explicitly request or specs require
- **Prefer Standard Library**: Minimize external dependencies (reduces supply chain risk)
- **Flat Architecture**: Avoid deep package hierarchies or abstract layers without clear purpose
- **Direct Solutions**: Prefer straightforward implementations over "clever" optimizations
- **Delete Dead Code**: Remove unused features, packages, or code paths immediately

**Rationale**: Complexity is the enemy of security. Every line of code is a potential vulnerability. Simplicity improves auditability, maintainability, and trust.

---

## Security Requirements

**This section supplements Principle I with specific constraints for password manager development.**

### Forbidden Operations

The following operations are FORBIDDEN and MUST be blocked in code review:

- ❌ Logging any variable named `password`, `masterPassword`, `apiKey`, `secret`, `credential`, `vault`, or similar
- ❌ Writing credentials to temporary files (use in-memory buffers only)
- ❌ Transmitting credentials over network (Pass-CLI is offline-first)
- ❌ Storing credentials in environment variables (except user's own controlled `export` in scripts)
- ❌ Using weak cryptography (MD5, SHA1 for hashing, DES, 3DES, RC4 for encryption)
- ❌ Hardcoding encryption keys, salts, or IVs in source code

### Security Review Checklist

Before merging any PR touching cryptographic or credential-handling code:

- [ ] No secrets logged or printed
- [ ] Memory containing secrets zeroed after use
- [ ] Error messages do not leak sensitive information
- [ ] Clipboard auto-clears within 30 seconds
- [ ] Vault file permissions restrict access to owner only (`0600` on Unix, equivalent ACLs on Windows)
- [ ] All cryptographic operations use approved algorithms (AES-256-GCM, PBKDF2-SHA256)
- [ ] Tests verify security guarantees (e.g., clipboard clearing, vault file permissions)

---

## Development Workflow

### Quality Gates

**All code MUST pass these gates before merging:**

1. **Tests Pass**: `go test ./...` returns zero exit code on all target platforms
2. **Linting Clean**: `golangci-lint run` reports zero issues
3. **Security Scan**: `gosec ./...` reports zero high/medium vulnerabilities
4. **Coverage Threshold**: Minimum 80% code coverage for new code
5. **Constitution Compliance**: Reviewer verifies adherence to all principles above

### Code Review Requirements

- **Security-Critical Code**: Requires two approvals (one MUST be from a security-focused reviewer)
- **Cryptographic Changes**: Requires external security audit before merge
- **API Changes**: Requires contract test updates demonstrating backward compatibility or versioning
- **Cross-Platform Code**: Reviewer MUST verify CI passes on all platforms

### Branch Strategy

- **Main Branch**: Protected, always deployable, MUST pass all quality gates
- **Feature Branches**: Named `###-feature-name` (e.g., `001-add-tui`, `002-import-from-bitwarden`)
- **Hotfix Branches**: Named `hotfix/description` for security or critical bugs

### Commit Discipline

- **Commit Frequently**: After each task, phase, or working checkpoint (as per CLAUDE.md guidelines)
- **Atomic Commits**: Each commit MUST represent a single logical change
- **Conventional Commits**: Use `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:` prefixes
- **Security Fixes**: Prefix with `security:` and include CVE/issue reference if applicable

---

## Governance

### Amendment Process

This constitution is the authoritative source for Pass-CLI development practices. **All other practices and guidelines are subordinate to these principles.**

**To amend this constitution:**

1. Propose change with detailed justification in GitHub Issue
2. Draft amendment with version bump rationale (MAJOR/MINOR/PATCH)
3. Update all dependent templates (plan, spec, tasks, commands)
4. Obtain approval from project maintainers
5. Commit with message: `docs: amend constitution to vX.Y.Z (summary of changes)`
6. Update all active specs and in-flight work to comply with new version

### Version Bump Rules

- **MAJOR**: Backward-incompatible governance changes, principle removals, or redefinitions that invalidate existing specs
- **MINOR**: New principles added, materially expanded guidance, new sections added
- **PATCH**: Clarifications, wording improvements, typo fixes, non-semantic refinements

### Compliance Reviews

- **Per-PR Review**: Every pull request MUST verify compliance with applicable principles (reviewers check constitution during review)
- **Quarterly Audit**: Maintainers conduct full codebase audit against constitution every quarter
- **Spec Reviews**: All spec documents (plan.md, spec.md, tasks.md) MUST reference constitution principles in acceptance criteria

### Complexity Justification

Any violation of Principle VII (Simplicity) MUST be justified in the relevant plan.md under "Complexity Tracking." Unjustified complexity WILL be rejected in code review.

### Runtime Guidance

For day-to-day development workflow, communication standards, and detailed implementation guidelines, refer to [CLAUDE.md](../../CLAUDE.md). That file provides runtime guidance for AI assistants and human developers working on Pass-CLI features.

**Precedence**: Constitution > CLAUDE.md > Other docs. In case of conflict, this constitution takes precedence.

---

**Version**: 1.1.0 | **Ratified**: 2025-10-09 | **Last Amended**: 2025-10-11
