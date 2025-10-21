# Branch Workflow

This document describes the branching strategy and workflow for pass-cli development.

## Branch Structure

### `main` Branch
- **Purpose**: Production-ready code only
- **Protection**: Heavily protected (see below)
- **Updates**: Only via "Promote to Main" workflow or PR
- **Default**: Yes (users cloning repo get this)
- **Tags**: All release tags are on main
- **CI**: Runs only on PRs to main

### `develop` Branch
- **Purpose**: Active development work
- **Protection**: Light protection (no deletion, no force push)
- **Updates**: Direct pushes allowed for daily work
- **CI**: Runs on every push to develop
- **Specs/Planning**: All specs, notes, and planning docs live here

## Daily Workflow

### Starting Work

```bash
# Clone repo (if new)
git clone https://github.com/ari1110/pass-cli.git
cd pass-cli

# Switch to develop
git checkout develop

# Make sure you're up to date
git pull origin develop
```

### Making Changes

```bash
# Make changes on develop branch
git add .
git commit -m "feat: add new feature"
git push origin develop

# CI runs automatically on develop
```

### Promoting to Production

When ready to release:

1. **Verify develop is clean**
   ```bash
   git checkout develop
   git status
   # Should show "nothing to commit, working tree clean"
   ```

2. **Trigger promotion workflow**
   - Go to GitHub Actions: https://github.com/ari1110/pass-cli/actions
   - Select **"Promote to Main"** workflow (left sidebar)
   - Click **"Run workflow"** button (right side)
   - Choose options:
     - Branch: `develop` (default)
     - Run tests: `true` (recommended)
   - Click green **"Run workflow"** button

3. **Wait for merge**
   - Workflow runs tests (if enabled)
   - Merges develop → main automatically
   - Check workflow succeeds

4. **Create release**
   ```bash
   # Switch to main and pull the merge
   git checkout main
   git pull origin main

   # Tag the release
   git tag v0.x.x
   git push origin v0.x.x

   # Release workflow runs automatically
   ```

5. **Return to develop**
   ```bash
   git checkout develop
   # Continue working
   ```

## Branch Protection

### Main Branch Protection (Ruleset)

**Enabled protections:**
- ✅ Require pull request before merging
- ✅ Require status checks to pass:
  - `Lint`
  - `Unit Tests`
  - `Integration Tests (ubuntu-latest)`
  - `Integration Tests (macos-latest)`
  - `Integration Tests (windows-latest)`
  - `Security Scan`
  - `Build`
- ✅ Require branches to be up to date before merging
- ✅ Block force pushes
- ✅ Restrict deletions

**Result**: Cannot push directly to main. Must use workflow or PR.

### Develop Branch Protection (Ruleset)

**Enabled protections:**
- ✅ Block force pushes
- ✅ Restrict deletions

**Result**: Direct pushes allowed, but history preserved.

## CI/CD Pipeline

### On `develop` Push
```
develop push → CI runs (lint, unit tests, integration tests, security scan, build)
```

### On PR to `main` or `develop`
```
PR created → CI runs on PR branch
```

### On Promotion
```
"Promote to Main" workflow triggered →
  Optional: Run tests on develop →
  Merge develop → main (--no-ff) →
  Push to main
```

### On Release Tag
```
Tag pushed to main (v*) →
  Release workflow runs →
  Build binaries for all platforms →
  Create GitHub release →
  Update Homebrew tap →
  Update Scoop bucket
```

## Emergency Hotfixes

For critical production bugs:

**Option 1: Via develop (Recommended)**
```bash
git checkout develop
# Fix the bug
git commit -m "fix: critical bug"
git push origin develop
# Use "Promote to Main" workflow immediately
```

**Option 2: Direct PR to main (If develop has unreleased work)**
```bash
git checkout -b hotfix/critical-bug main
# Fix the bug
git commit -m "fix: critical bug"
git push origin hotfix/critical-bug
# Create PR: hotfix/critical-bug → main
# CI must pass before merge allowed
```

## Common Commands

### Check which branch you're on
```bash
git branch --show-current
```

### See commits on develop not in main
```bash
git log main..develop --oneline
```

### Compare file changes between branches
```bash
git diff main..develop
```

### List all branches
```bash
git branch -a
```

## Tips

1. **Always work on develop** - Don't work directly on main
2. **Commit frequently** - Small, focused commits are better
3. **Test locally** - Run `go test ./...` before pushing
4. **Use promote workflow** - Don't manually merge develop to main
5. **Keep main clean** - Only release-ready code on main

## Troubleshooting

### "Cannot push to main" error
This is expected! Use the "Promote to Main" workflow instead.

### Develop and main diverged
```bash
# Should never happen with this workflow, but if it does:
git checkout develop
git pull origin develop
git pull origin main
# Resolve conflicts if any
git push origin develop
```

### CI failing on develop
Fix the issue before promoting to main:
```bash
git checkout develop
# Fix the failing tests/linting
git commit -m "fix: resolve CI failures"
git push origin develop
# Wait for CI to pass
```

## References

- GitHub Actions: https://github.com/ari1110/pass-cli/actions
- Branch Settings: https://github.com/ari1110/pass-cli/settings/rules
- Release Workflow: `.github/workflows/release.yml`
- CI Workflow: `.github/workflows/ci.yml`
- Promote Workflow: `.github/workflows/promote-to-main.yml`
