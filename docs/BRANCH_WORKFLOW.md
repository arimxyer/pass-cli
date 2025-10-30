# Branch Workflow

This document describes the branching strategy and workflow for pass-cli development.

## Branch Structure

### `main` Branch
- **Purpose**: Production-ready code
- **Protection**: Requires PR and passing CI checks
- **Updates**: Via pull requests from develop
- **Default**: Yes (users cloning repo get this)
- **Tags**: All release tags are created on main

### `develop` Branch
- **Purpose**: Active development work
- **Protection**: Block force pushes and deletions
- **Updates**: Direct pushes allowed for daily work
- **CI**: Runs on every push to develop
- **Content**: All files including specs, planning docs, and development notes

**Note**: Both branches now contain the same files. Development files (specs/, .specify/, etc.) exist on both branches to simplify the workflow. Release artifacts remain clean via GoReleaser configuration.

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
# Work on develop branch
git add .
git commit -m "feat: add new feature"
git push origin develop

# CI runs automatically
```

### Promoting to Production

When develop is stable and ready for release:

1. **Verify develop is clean and CI passing**
   ```bash
   git checkout develop
   git status
   # Should show "nothing to commit, working tree clean"
   # Check GitHub Actions to confirm CI is green
   ```

2. **Create PR from develop to main**
   - Go to GitHub: https://github.com/ari1110/pass-cli/compare/main...develop
   - Click "Create pull request"
   - Title: "chore: promote develop to main"
   - Description: List key changes being promoted
   - CI will run on the PR

3. **Review and merge**
   - Wait for CI checks to pass
   - Review changes in PR
   - Click "Squash and merge" or "Merge pull request"

4. **Sync develop with main**
   ```bash
   git checkout develop
   git pull origin develop
   git merge main
   git push origin develop
   ```

5. **Create release**
   ```bash
   git checkout main
   git pull origin main

   # Tag the release
   git tag v0.x.x
   git push origin v0.x.x

   # Release workflow runs automatically
   ```

## Branch Protection

### Main Branch Protection

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
- ✅ Block force pushes
- ✅ Restrict deletions
- ✅ Repository admins can bypass (for emergency fixes)

**Result**: Cannot push directly to main. Must use pull requests.

### Develop Branch Protection

**Enabled protections:**
- ✅ Block force pushes
- ✅ Restrict deletions

**Result**: Direct pushes allowed, but history preserved.

## CI/CD Pipeline

### On `develop` Push
```
develop push → CI runs (lint, unit tests, integration tests, security scan, build)
```

**Smart filtering**: CI skips test jobs when only non-code files change (.github/, docs/, specs/, etc.)

### On PR to `main` or `develop`
```
PR created → CI runs on PR branch → Must pass before merge allowed
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
git pull origin develop
# Fix the bug
git commit -m "fix: critical bug"
git push origin develop
# Create PR to main immediately
```

**Option 2: Direct PR to main (If develop has unreleased work)**
```bash
git checkout -b hotfix/critical-bug main
# Fix the bug
git commit -m "fix: critical bug"
git push origin hotfix/critical-bug
# Create PR: hotfix/critical-bug → main
# After merge, sync to develop:
git checkout develop
git merge main
git push origin develop
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

### Sync develop with main (after promoting)
```bash
git checkout develop
git merge main
git push origin develop
```

## Tips

1. **Always work on develop** - Don't work directly on main
2. **Commit frequently** - Small, focused commits are better
3. **Test locally** - Run `go test ./...` before pushing
4. **Create PRs for promotion** - Manual PR from develop → main
5. **Sync after promotion** - Merge main → develop after successful promotion
6. **Keep commits clean** - Use descriptive commit messages

## Troubleshooting

### "Cannot push to main" error
This is expected! Create a pull request from develop → main instead.

### Develop and main diverged
This can happen if you forget to sync develop after promotion:
```bash
git checkout develop
git pull origin develop
git merge main
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
# Wait for CI to pass on GitHub Actions
```

### PR has merge conflicts
```bash
# Update develop with main's changes
git checkout develop
git merge main
# Resolve conflicts
git push origin develop
# PR should now be conflict-free
```

## References

- GitHub Actions: https://github.com/ari1110/pass-cli/actions
- Branch Settings: https://github.com/ari1110/pass-cli/settings/rules
- Release Workflow: `.github/workflows/release.yml`
- CI Workflow: `.github/workflows/ci.yml`
