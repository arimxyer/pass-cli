# Quickstart: cmd/tui Directory Reorganization

**Feature**: Reorganize cmd/tui Directory Structure
**Branch**: `001-reorganize-cmd-tui`
**Estimated Time**: 1-2 hours

## Prerequisites

- ✅ Working TUI on `pre-reorg-tui` branch (confirmed no black screen)
- ✅ Git installed and configured
- ✅ Go 1.25.1 toolchain installed
- ✅ Terminal capable of rendering tview TUI
- ✅ Test vault available at `test-vault/vault.enc` with known password

## Migration Overview

This guide walks through the systematic migration of the TUI package from `cmd/tui-tview/` to `cmd/tui/` in four atomic steps, with verification checkpoints after each step.

**Step Sequence**:
1. Package rename (`package main` → `package tui`)
2. Import path updates (`cmd/tui-tview/*` → `cmd/tui/*`)
3. Directory move (`git mv cmd/tui-tview cmd/tui`)
4. Main integration (update `main.go` to call `tui.Run()`)

## Step-by-Step Instructions

### Step 0: Setup Baseline

1. **Checkout working baseline**:
   ```bash
   git checkout pre-reorg-tui
   git pull origin pre-reorg-tui
   ```

2. **Verify baseline works**:
   ```bash
   go build -o pass-cli-baseline.exe .
   ./pass-cli-baseline.exe --vault test-vault/vault.enc
   ```
   - **Expected**: TUI renders completely, no black screen
   - **Test password**: `test123` (or your test vault password)
   - **Visual check**: Sidebar visible, credentials listed, navigation works
   - Press `q` to exit

3. **Create feature branch**:
   ```bash
   git checkout -b 001-reorganize-cmd-tui
   ```

**Checkpoint 0**: ✅ Baseline working, feature branch created

---

### Step 1: Package Rename and Export

**Goal**: Change TUI package from standalone executable to importable library

1. **Update package declaration**:

   Find all `.go` files in `cmd/tui-tview/`:
   ```bash
   rg "^package main$" cmd/tui-tview/ --type go -l
   ```

   For each file, change:
   ```go
   package main
   ```
   to:
   ```go
   package tui
   ```

2. **Rename main() to Run()**:

   Edit `cmd/tui-tview/main.go`:

   **Find**:
   ```go
   func main() {
       // 1. Get default vault path
       vaultPath := getDefaultVaultPath()

       // ... rest of main()
   }
   ```

   **Replace with**:
   ```go
   // Run starts the TUI application (exported for main.go to call)
   // If vaultPath is empty, uses the default vault location
   func Run(vaultPath string) error {
       // 1. Get vault path (use provided path or default)
       if vaultPath == "" {
           vaultPath = getDefaultVaultPath()
       }

       // ... rest of function (same as before)

       // Return error instead of os.Exit
       return nil
   }
   ```

   **Key changes**:
   - Function name: `main()` → `Run(vaultPath string) error`
   - Add `vaultPath` parameter and handling
   - Return errors instead of calling `os.Exit(1)`
   - Keep `LaunchTUI()` unchanged

3. **Verify compilation**:
   ```bash
   go build ./...
   ```
   - **Expected**: Compiles successfully with no errors
   - **If errors**: Check that ALL .go files in cmd/tui-tview/ use `package tui`

4. **Test TUI rendering** (temporary test build):

   Temporarily add a test wrapper to `cmd/tui-tview/main.go`:
   ```go
   // Temporary test wrapper - will be removed after migration
   func main() {
       if err := Run(""); err != nil {
           fmt.Fprintf(os.Stderr, "Error: %v\n", err)
           os.Exit(1)
       }
   }
   ```

   Build and run:
   ```bash
   go build -o pass-cli-step1.exe cmd/tui-tview/
   ./pass-cli-step1.exe --vault test-vault/vault.enc
   ```
   - **Expected**: TUI renders completely, no black screen
   - **Visual check**: All features work (navigation, forms, detail view)
   - Press `q` to exit

5. **Remove test wrapper** (after verification):
   Delete the temporary `func main()` wrapper added above.

6. **Commit**:
   ```bash
   git add cmd/tui-tview/
   git commit -m "refactor: Change TUI package from main to tui

- Update package declaration in all TUI files
- Rename main() to Run(vaultPath string) error
- Add vaultPath parameter handling
- Keep LaunchTUI() unchanged

Step 1 of 4 for cmd/tui reorganization.

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"
   ```

**Checkpoint 1**: ✅ Package renamed, Run() exported, TUI verified working

---

### Step 2: Import Path Updates

**Goal**: Update all imports from `cmd/tui-tview/*` to `cmd/tui/*`

1. **Find all occurrences**:
   ```bash
   rg "cmd/tui-tview" --type go -l
   ```
   - **Expected**: List of files with old import paths

2. **Automated replacement**:
   ```bash
   # Windows PowerShell
   Get-ChildItem -Recurse -Filter *.go | ForEach-Object {
       (Get-Content $_.FullName) -replace 'cmd/tui-tview', 'cmd/tui' | Set-Content $_.FullName
   }

   # Or use editor find/replace:
   # Find:    cmd/tui-tview
   # Replace: cmd/tui
   # Files:   *.go
   ```

3. **Verify no missed occurrences**:
   ```bash
   rg "tui-tview" --type go
   ```
   - **Expected**: No results (all occurrences replaced)

4. **Verify compilation**:
   ```bash
   go build ./...
   ```
   - **Expected**: Compiles successfully
   - **If errors**: Check for any remaining `tui-tview` references

5. **Commit**:
   ```bash
   git add .
   git commit -m "refactor: Update import paths from cmd/tui-tview to cmd/tui

- Replace all occurrences of pass-cli/cmd/tui-tview with pass-cli/cmd/tui
- Verified no missed occurrences
- All code compiles successfully

Step 2 of 4 for cmd/tui reorganization.

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"
   ```

**Checkpoint 2**: ✅ Import paths updated, code compiles

---

### Step 3: Directory Migration

**Goal**: Physically move directory from `cmd/tui-tview/` to `cmd/tui/` while preserving git history

1. **Move directory**:
   ```bash
   git mv cmd/tui-tview cmd/tui
   ```

2. **Verify git tracked the rename**:
   ```bash
   git status
   ```
   - **Expected**: Shows `renamed: cmd/tui-tview/ -> cmd/tui/`

3. **Verify git history preservation**:
   ```bash
   git log --follow --oneline cmd/tui/main.go | head -5
   ```
   - **Expected**: Shows commit history from before the move

4. **Verify compilation**:
   ```bash
   go build ./...
   ```
   - **Expected**: Compiles successfully using new directory path

5. **Commit**:
   ```bash
   git commit -m "refactor: Move directory from cmd/tui-tview to cmd/tui

- Use git mv to preserve file history
- Verified git tracks rename correctly
- All code compiles with new directory structure

Step 3 of 4 for cmd/tui reorganization.

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"
   ```

**Checkpoint 3**: ✅ Directory moved, git history preserved, code compiles

---

### Step 4: Main Entry Point Integration

**Goal**: Integrate TUI launch into main.go for default TUI behavior

1. **Read current main.go**:
   ```bash
   cat main.go
   ```
   - **Current**: Only calls `cmd.Execute()` (CLI commands only)

2. **Update main.go**:

   **Replace entire content with**:
   ```go
   package main

   import (
       "fmt"
       "os"

       "pass-cli/cmd"
       "pass-cli/cmd/tui"
   )

   func main() {
       // Default to TUI if no subcommand provided
       shouldUseTUI := true
       vaultPath := ""

       // Parse args to detect subcommands or flags
       for i := 1; i < len(os.Args); i++ {
           arg := os.Args[i]

           // Check for help or version flags
           if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
               shouldUseTUI = false
               break
           }

           // Check for subcommands (any non-flag argument)
           if arg != "" && arg[0] != '-' {
               shouldUseTUI = false
               break
           }

           // Extract vault path if provided
           if arg == "--vault" && i+1 < len(os.Args) {
               vaultPath = os.Args[i+1]
               i++ // Skip next arg (vault path value)
           }
       }

       // Route to TUI or CLI
       if shouldUseTUI {
           if err := tui.Run(vaultPath); err != nil {
               fmt.Fprintf(os.Stderr, "Error: %v\n", err)
               os.Exit(1)
           }
       } else {
           cmd.Execute()
       }
   }
   ```

3. **Verify compilation**:
   ```bash
   go build -o pass-cli.exe .
   ```
   - **Expected**: Compiles successfully

4. **Test TUI launch** (no arguments):
   ```bash
   ./pass-cli.exe --vault test-vault/vault.enc
   ```
   - **Expected**: TUI renders completely, no black screen
   - **Visual check**: All features work
   - Press `q` to exit

5. **Test CLI commands still work**:
   ```bash
   ./pass-cli.exe list --vault test-vault/vault.enc
   ./pass-cli.exe --help
   ./pass-cli.exe --version
   ```
   - **Expected**: CLI commands execute normally (no TUI)

6. **Commit**:
   ```bash
   git add main.go
   git commit -m "feat: Integrate TUI launch into main.go entry point

- Add default TUI routing when no subcommand provided
- Parse --vault flag for TUI mode
- Preserve CLI routing for subcommands and flags
- Verified TUI renders correctly and CLI commands work

Step 4 of 4 for cmd/tui reorganization - MIGRATION COMPLETE.

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>"
   ```

**Checkpoint 4**: ✅ Main integration complete, TUI launches by default, CLI still works

---

## Final Verification

### Comprehensive Testing

1. **TUI Mode** (default behavior):
   ```bash
   ./pass-cli.exe --vault test-vault/vault.enc
   ```
   - ✅ TUI renders with no black screen
   - ✅ Sidebar shows categories
   - ✅ Credentials listed in table
   - ✅ Detail view shows credential details
   - ✅ Navigation works (arrow keys, Tab, Enter)
   - ✅ Forms work (Ctrl+A to add credential)
   - ✅ Delete confirmation works
   - ✅ Password masking toggle works

2. **CLI Mode** (subcommands):
   ```bash
   ./pass-cli.exe list --vault test-vault/vault.enc
   ./pass-cli.exe get aws-production --vault test-vault/vault.enc
   ./pass-cli.exe add test-service --vault test-vault/vault.enc
   ./pass-cli.exe --help
   ./pass-cli.exe --version
   ```
   - ✅ All CLI commands execute normally
   - ✅ No TUI interference

3. **Edge Cases**:
   ```bash
   ./pass-cli.exe --vault test-vault/vault.enc --help
   ./pass-cli.exe -h
   ./pass-cli.exe -v
   ```
   - ✅ Help and version flags work (CLI, not TUI)

### Success Criteria Verification

Check against spec success criteria:

- ✅ **SC-001**: TUI renders completely with no black screen or visual corruption
- ✅ **SC-002**: All existing TUI features function identically to pre-migration state
- ✅ **SC-003**: Project compiles successfully after each migration step
- ✅ **SC-004**: Complete migration performed and verified in under 2 hours
- ✅ **SC-005**: Zero new compiler errors or warnings introduced
- ✅ **SC-006**: TUI launches in under 3 seconds

---

## Rollback Procedures

If issues are discovered:

### Rollback to Previous Step

```bash
# Rollback Step 4 (main integration)
git reset --hard HEAD~1

# Rollback Step 3 (directory move)
git reset --hard HEAD~2

# Rollback Step 2 (import updates)
git reset --hard HEAD~3

# Rollback Step 1 (package rename)
git reset --hard HEAD~4

# Complete rollback to baseline
git reset --hard pre-reorg-tui
```

### Resume from Checkpoint

If you need to stop and resume:

```bash
# Check current step
git log --oneline -4

# Resume from last commit
git status  # Verify clean working tree
# Continue from next step in this guide
```

---

## Troubleshooting

### Issue: Black screen after migration

**Symptoms**: TUI shows black screen, no rendering

**Diagnosis**:
1. Check for `QueueUpdateDraw()` calls before `app.Run()` in `cmd/tui/main.go`
2. Check for `nav.SetFocus()` calls before `app.Run()`
3. Check for `fmt.Println()` after password input

**Fix**:
- Remove `QueueUpdateDraw()` wrappers from callbacks before `app.Run()`
- Remove `nav.SetFocus()` call before `app.Run()`
- Remove `fmt.Println()` after `gopass.GetPasswdMasked()`

### Issue: Import errors after Step 2

**Symptoms**: Compiler errors about missing packages

**Diagnosis**:
```bash
rg "tui-tview" --type go
```

**Fix**: Missed occurrences - rerun find/replace from Step 2

### Issue: CLI commands launch TUI

**Symptoms**: `./pass-cli.exe list` opens TUI instead of listing

**Diagnosis**: Argument parsing logic in main.go incorrect

**Fix**: Verify the `shouldUseTUI` logic checks for non-flag arguments

---

## Next Steps

After successful migration:

1. **Push to remote**:
   ```bash
   git push -u origin 001-reorganize-cmd-tui
   ```

2. **Create pull request**:
   - Target branch: `main`
   - Title: "Reorganize cmd/tui directory structure"
   - Description: Link to spec.md and plan.md

3. **Run CI/CD**:
   - GitHub Actions will test cross-platform builds
   - Verify all tests pass

4. **Merge to main**:
   - After approval and CI success
   - Delete feature branch after merge

---

**Migration Complete!** 🎉

The TUI package is now cleanly organized at `cmd/tui/` with proper integration into the main entry point, while preserving all existing functionality.
