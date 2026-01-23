# Instructions for Creating PR to awesome-tuis

## Overview
This repository now contains a patch file that adds pass-cli and models to the awesome-tuis list.

## Quick Start

### Option 1: Apply the Patch File
```bash
# 1. Fork rothgar/awesome-tuis on GitHub (if not already done)
#    Go to: https://github.com/rothgar/awesome-tuis
#    Click "Fork"

# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/awesome-tuis.git
cd awesome-tuis

# 3. Create a new branch
git checkout -b add-pass-cli-and-models

# 4. Apply the patch
git am /path/to/0001-Add-pass-cli-and-models-to-awesome-tuis-list.patch

# 5. Push to your fork
git push origin add-pass-cli-and-models

# 6. Create PR on GitHub
#    Go to: https://github.com/rothgar/awesome-tuis
#    Click "New Pull Request"
#    Select your fork and branch
```

### Option 2: Manual Edit
If you prefer to make the changes manually:

1. Fork and clone rothgar/awesome-tuis
2. Edit README.md and add these two lines:

**In Development section (after line 148 - after mitmproxy):**
```markdown
- [models](https://github.com/arimxyer/models) Super fast CLI and TUI for browsing AI models. Quickly look up context windows, pricing, capabilities, and more for 2000+ models across 75+ providers.
```

**In Miscellaneous section (after line 507 - after packemon):**
```markdown
- [pass-cli](https://github.com/arimxyer/pass-cli) A secure, cross-platform, always-free, and open-source CLI and TUI password manager for folks who live in the command line.
```

3. Commit, push, and create PR

## What's Being Added

### pass-cli (Miscellaneous section)
- **Description:** A secure, cross-platform, always-free, and open-source CLI and TUI password manager for folks who live in the command line
- **Location:** After "packemon", before "PesterExplorer"
- **Section:** Miscellaneous (security/password management tools)

### models (Development section)
- **Description:** Super fast CLI and TUI for browsing AI models with context windows, pricing, and capabilities
- **Location:** After "mitmproxy", before "nap"
- **Section:** Development (AI development tools)

## PR Title Suggestion
```
Add pass-cli and models TUI projects
```

## PR Description Suggestion
```
This PR adds two terminal user interface projects:

1. **pass-cli** - A secure, cross-platform CLI and TUI password manager for command-line users
   - Added to Miscellaneous section
   - Repository: https://github.com/arimxyer/pass-cli

2. **models** - A fast CLI and TUI for browsing AI models across 75+ providers
   - Added to Development section  
   - Repository: https://github.com/arimxyer/models

Both entries follow the existing format and are alphabetically sorted within their sections.
```
