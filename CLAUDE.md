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

## Active Technologies
- Go 1.21+ (existing codebase) (011-vault-and-keychain)
- Encrypted vault files (vault.enc) with JSON structure, OS-native keychain for master passwords (per-vault entries using service name format "pass-cli:/absolute/path/to/vault.enc" per spec.md FR-003) (011-keychain-lifecycle-management)

## Recent Changes

- 011-vault-and-keychain: Added Go 1.21+ (existing codebase)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->

**Last updated**: 2025-10-20
