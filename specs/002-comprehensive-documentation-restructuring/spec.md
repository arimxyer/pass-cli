# Feature Specification: Comprehensive Documentation Restructuring

**Feature Branch**: `002-comprehensive-documentation-restructuring`
**Created**: 2025-11-12
**Status**: Draft
**Input**: User description: "Comprehensive documentation restructuring into task-based 6-section architecture with focused, scannable documents averaging 300 lines. Split oversized files (cli-reference: 2040 lines, troubleshooting: 1404 lines), eliminate 15-20% content redundancy, and reorganize into user-journey-optimized structure."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - New User Quick Start (Priority: P1)

A developer discovers pass-cli and wants to install it and add their first credential within 5 minutes.

**Why this priority**: First-time user experience directly impacts adoption. Currently, new users must wade through 1,357 lines across installation.md and first-steps.md to complete basic setup.

**Independent Test**: Can be fully tested by following the quick-start documentation from a fresh system to successfully storing and retrieving a credential, delivering immediate value for new users.

**Acceptance Scenarios**:

1. **Given** a developer on any platform, **When** they navigate to getting-started documentation, **Then** they find a quick-install guide under 100 lines with only package manager instructions
2. **Given** pass-cli is installed, **When** they follow the quick-start guide, **Then** they can init vault, add a credential, and retrieve it using less than 200 lines of documentation
3. **Given** a new user wants comprehensive installation options, **When** they need manual installation or building from source, **Then** they find a separate manual-install document without quick-start clutter

---

### User Story 2 - Daily Command Reference Lookup (Priority: P1)

An existing user needs to quickly look up command syntax for get, add, update, or delete operations without scrolling through unrelated content.

**Why this priority**: Daily users represent majority of traffic. Currently forced to scan 2,040-line cli-reference.md mixing commands, TUI guide, config, troubleshooting, and examples.

**Independent Test**: Can be tested by timing how long it takes to find specific command syntax in restructured common-commands.md vs current cli-reference.md, demonstrating improved efficiency.

**Acceptance Scenarios**:

1. **Given** a user needs command syntax, **When** they open the command reference, **Then** they find all CLI commands in a focused 600-line document without TUI/config/troubleshooting content
2. **Given** a user wants common command examples, **When** they access common-commands documentation, **Then** they see get/add/list/update/delete in a focused 200-line guide
3. **Given** a user needs TUI documentation, **When** they search for TUI guide, **Then** they find a separate 400-line tui-guide.md not mixed with CLI commands

---

### User Story 3 - Troubleshooting Error Resolution (Priority: P2)

A user encounters a vault corruption error and needs to find recovery steps without reading 1,404 lines of mixed troubleshooting content.

**Why this priority**: Users experiencing problems need fast resolution. Current 1,404-line troubleshooting.md mixes installation, vault, keychain, TUI, and platform issues making problems hard to locate.

**Independent Test**: Can be tested by giving users specific error scenarios and measuring time-to-solution in category-specific troubleshooting docs vs monolithic document.

**Acceptance Scenarios**:

1. **Given** a user has a vault access error, **When** they navigate to troubleshooting, **Then** they find a dedicated troubleshooting-vault.md (350 lines) covering only vault issues
2. **Given** a user has keychain integration problems, **When** they search troubleshooting, **Then** they find platform-specific keychain troubleshooting (250 lines) separate from other issues
3. **Given** a user has a general question, **When** they check FAQ, **Then** they find a consolidated faq.md (200 lines) with all FAQ content from multiple docs in one location

---

### User Story 4 - Advanced Feature Discovery (Priority: P2)

A power user wants to explore usage tracking, backup/restore, and keychain integration without information scattered across multiple documents.

**Why this priority**: Power users generate engagement and evangelism. Current structure scatters advanced features (usage tracking in cli-reference lines 1581-1721, keychain in first-steps.md and cli-reference.md).

**Independent Test**: Can be tested by providing power users with feature discovery tasks and measuring completion rates and satisfaction with dedicated guides vs scattered content.

**Acceptance Scenarios**:

1. **Given** a user wants usage tracking, **When** they search for usage documentation, **Then** they find a dedicated usage-tracking.md (200 lines) explaining multi-location tracking
2. **Given** a user wants keychain integration, **When** they look for keychain setup, **Then** they find a complete keychain-setup.md (150 lines) consolidating all keychain content
3. **Given** a user wants to automate pass-cli, **When** they search for scripting, **Then** they find a scripting-guide.md (300 lines) with quiet mode, JSON output, and examples

---

### User Story 5 - Security Audit and Operations (Priority: P3)

A security engineer needs to review pass-cli's security architecture and incident response procedures for compliance approval.

**Why this priority**: Enterprise adoption requires security documentation. Current security.md (750 lines) mixes technical architecture with operational procedures.

**Independent Test**: Can be tested by security professionals reviewing separated security-architecture.md and security-operations.md for completeness and clarity compared to monolithic security.md.

**Acceptance Scenarios**:

1. **Given** a security auditor reviews architecture, **When** they read security documentation, **Then** they find technical crypto details and threat model in security-architecture.md (500 lines) without operational clutter
2. **Given** an ops engineer needs incident response, **When** they search security operations, **Then** they find best practices and incident response in security-operations.md (250 lines) separate from technical specs
3. **Given** a security team needs health monitoring, **When** they look for diagnostics, **Then** they find health-checks.md in Operations section (moved from Development) as it's user-facing

---

### Edge Cases

- What happens when a doc split creates orphaned content that doesn't fit new structure? → Document in Assumptions as "miscellaneous content" requiring case-by-case decisions
- How do we handle cross-references between split docs? → Use Hugo `relref` shortcodes for all internal links, verified during final testing
- What if users bookmark old document URLs? → Hugo handles file moves gracefully; old structure still in git history
- How do we prevent re-accumulation of content bloat? → Out of scope (could be future doc lifecycle policy)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Documentation MUST be reorganized into 6 top-level sections: Getting Started, Guides, Reference, Troubleshooting, Operations, Development
- **FR-002**: cli-reference.md (2,040 lines) MUST be split into 5 separate documents: command-reference.md (600 lines), tui-guide.md (400 lines), scripting-guide.md (300 lines), configuration.md (250 lines), usage-tracking.md (200 lines)
- **FR-003**: troubleshooting.md (1,404 lines) MUST be split into 5 separate documents: troubleshooting-installation.md (300 lines), troubleshooting-vault.md (350 lines), troubleshooting-keychain.md (250 lines), troubleshooting-tui.md (300 lines), faq.md (200 lines)
- **FR-004**: installation.md (708 lines) MUST be split into 3 documents: quick-install.md (100 lines), manual-install.md (400 lines), uninstall.md (100 lines)
- **FR-005**: first-steps.md (649 lines) MUST be split into 3 documents: quick-start.md (200 lines), basic-workflows.md (250 lines), keychain-setup.md (150 lines)
- **FR-006**: security.md (750 lines) MUST be split into 2 documents: security-architecture.md (500 lines), security-operations.md (250 lines)
- **FR-007**: Duplicate content (keychain across 3 docs, FAQ across 4 docs, configuration across 2 docs) MUST be consolidated into single-source-of-truth documents
- **FR-008**: All documentation files MUST use git mv during reorganization to preserve file history
- **FR-009**: All internal documentation links MUST be updated to Hugo relref shortcode format pointing to new locations
- **FR-010**: Hugo site homepage (docs/README.md and docs/_index.md) MUST be updated with new quick links to reorganized structure
- **FR-011**: Root README.md MUST be updated with documentation links pointing to new file paths
- **FR-012**: All section _index.md files MUST be created/updated with accurate descriptions and weights for new structure
- **FR-013**: Front matter (title, weight, bookToc) MUST be verified/updated for all moved/split documents
- **FR-014**: doctor-command.md MUST be relocated from Development section to new Operations section
- **FR-015**: Final structure MUST contain 29 documentation files organized across 6 sections

### Key Entities

- **Documentation File**: A markdown file representing a single-topic guide, reference, or troubleshooting document. Key attributes: file path, line count, topic focus, target audience (new user/daily user/power user/developer/ops), section (getting-started/guides/reference/troubleshooting/operations/development)
- **Documentation Section**: A top-level category containing related docs. Key attributes: section name, weight (order), number of files, total content (words/lines)
- **Content Fragment**: A reusable piece of documentation content. Key attributes: topic (keychain/FAQ/config), source files (where duplicated), canonical location (single-source-of-truth destination)
- **Internal Link**: A reference from one doc to another. Key attributes: source file, target file, link format (Hugo relref required)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Average documentation file length reduced from 450 lines to 300 lines or less
- **SC-002**: Longest single documentation file reduced from 2,040 lines to under 700 lines
- **SC-003**: Total documentation content reduced from ~33,000 words to ~27,000 words through deduplication (15-20% reduction)
- **SC-004**: Duplicated content occurrences reduced from 15-20% to under 5% (measured by common content appearing in multiple files)
- **SC-005**: Time to find answer for common tasks (install, basic commands, troubleshooting) reduced by 40% (measured by user testing or page length proxy)
- **SC-006**: All 29 documentation files successfully render in Hugo site without broken links
- **SC-007**: Users can navigate from homepage to any common task (install, add credential, troubleshoot error) in 3 clicks or less
- **SC-008**: Documentation structure aligns with user journey priorities: new users find quick-start in Getting Started section, daily users find commands in Guides section, troubleshooters find categorized solutions in Troubleshooting section
- **SC-009**: Hugo site builds successfully with zero broken internal links after restructuring
- **SC-010**: All file moves preserve git history (verified by git log --follow on moved files)
