# Develop Branch Deprecated

**Date**: 2025-10-31
**Status**: ⚠️ DEPRECATED

## Notice

The `develop` branch has been **deprecated** as of October 31, 2025.

Pass-CLI has transitioned to a **main-only workflow** for simplified development and CI/CD.

## What Changed

### Old Workflow (deprecated)
- Work on `develop` branch
- Merge `develop` → `main` via PR
- CI runs twice (once on develop, once on PR)

### New Workflow (current)
- Create feature branch from `main` (e.g., `feat/my-feature`)
- Create PR to `main`
- CI runs once on PR
- Merge to `main`
- Tag releases on `main`

## Migration Guide

### If You Have Local Develop Branch

```bash
# Switch to main
git checkout main
git pull origin main

# Delete local develop branch
git branch -d develop

# For any work on develop, create feature branches from main instead
git checkout -b feat/my-feature main
```

### If You Have Open PRs to Develop

1. Close the existing PR to `develop`
2. Rebase your branch on `main`:
   ```bash
   git checkout your-feature-branch
   git fetch origin
   git rebase origin/main
   git push origin your-feature-branch --force-with-lease
   ```
3. Create new PR targeting `main`

## Why This Change?

1. **Eliminates duplicate CI runs** - CI now runs once per change instead of twice
2. **Fixes merge timing issues** - No more merge-base detection problems with paths-filter
3. **Simpler workflow** - Standard GitHub flow (feature → PR → main → tag)
4. **Industry standard** - Aligns with most open-source projects
5. **Better branch protection** - All changes require PR review and CI to pass

## Current Workflow

See [BRANCH_WORKFLOW.md](docs/BRANCH_WORKFLOW.md) for complete documentation.

**Quick Start**:
```bash
# Create feature branch
git checkout -b feat/my-feature

# Make changes and commit
git commit -m "feat: add feature"

# Push and create PR
git push origin feat/my-feature
# Create PR via GitHub to 'main'
```

## Questions?

Refer to [docs/BRANCH_WORKFLOW.md](docs/BRANCH_WORKFLOW.md) for the complete workflow guide.

---

**Note**: The `develop` branch will remain accessible for historical reference but should not be used for new development.
