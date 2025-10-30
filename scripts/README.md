# Scripts

Automation scripts for Pass-CLI development and release management.

## update-version.sh / update-version.ps1

Automated version and date updater for documentation and package manifests.

### Usage

**Bash (Linux, macOS, Git Bash on Windows):**
```bash
./scripts/update-version.sh v0.1.0
```

**PowerShell (Windows - if issues occur, use Git Bash instead):**
```powershell
.\scripts\update-version.ps1 v0.1.0
```

### What It Updates

1. **Documentation Version Footers**
   - `docs/USAGE.md`
   - `docs/SECURITY.md`
   - `docs/MIGRATION.md`
   - `docs/TROUBLESHOOTING.md`
   - `docs/KNOWN_LIMITATIONS.md`
   - `docs/GETTING_STARTED.md`
   - `docs/INSTALLATION.md`
   - `docs/DOCTOR_COMMAND.md`

2. **Last Updated Dates**
   - Updates all "Last Updated: Month YYYY" dates to current month/year

3. **Package Manifests**
   - `homebrew/pass-cli.rb` - version number
   - `scoop/pass-cli.json` - version number

### When to Use

Run this script before creating a new release tag:

1. **Before release**: Update all versions/dates
2. **Review changes**: `git diff`
3. **Update CHANGELOG.md**: Add release notes manually
4. **Commit**: `git commit -m "chore: bump version to vX.X.X"`
5. **Tag**: `git tag -a vX.X.X -m "Release vX.X.X"`
6. **Push**: `git push origin main --tags`

### What It Doesn't Update

- Package manifest URLs and checksums (handled by GoReleaser automatically)
- CHANGELOG.md (manual step - add your release notes)
- go.mod (not needed - GoReleaser reads from git tags)

### Example Output

```
=== Pass-CLI Version Update Tool ===
New version: v0.1.0
Update date: January 2025

Updating documentation files...
  ✓ USAGE.md
  ✓ SECURITY.md
  ✓ MIGRATION.md
  ✓ TROUBLESHOOTING.md
  ✓ KNOWN_LIMITATIONS.md
  ✓ GETTING_STARTED.md
  ✓ INSTALLATION.md
  ✓ DOCTOR_COMMAND.md

Updating package manifests...
  ✓ homebrew/pass-cli.rb
  ✓ scoop/pass-cli.json

=== Update Complete ===
Updated 10 files
```

### Troubleshooting

**"Must run from project root directory"**
- Run from the repository root (where `go.mod` exists)

**"You have uncommitted changes"**
- Commit or stash your changes first
- Or type `y` to continue anyway

**PowerShell execution policy error**
- Run: `powershell -ExecutionPolicy Bypass -File .\scripts\update-version.ps1 v0.1.0`
- Or use Git Bash with the `.sh` version

## See Also

- [Release Process](../docs/RELEASE.md) - Full release workflow
- [CI/CD Pipeline](../docs/CI-CD.md) - Automated workflows
