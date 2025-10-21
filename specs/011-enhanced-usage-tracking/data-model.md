# Data Model: Enhanced Usage Tracking CLI

**Feature**: Enhanced Usage Tracking CLI
**Date**: 2025-10-20
**Status**: Existing data structures (no new entities)

## Overview

This feature exposes existing usage tracking data through CLI commands. **No new data structures are needed** - all required entities already exist in the codebase at `internal/vault/vault.go`.

---

## Entity: UsageRecord (Existing)

**Source**: `internal/vault/vault.go:29-36`

**Purpose**: Tracks credential access events across different locations (directories, projects, machines)

**Definition**:
```go
type UsageRecord struct {
    Location      string
    GitRepository string
    FieldCounts   map[string]int
    Timestamp     time.Time
    AccessCount   int
}
```

### Fields

#### Location (string)

**Type**: Absolute path (platform-specific)
**Purpose**: Directory where credential was accessed
**Examples**:
- Unix: `/home/user/projects/my-app`
- Windows: `C:\Users\Alice\Projects\MyApp`
- Network: `\\server\share\project`

**Validation**:
- Not empty
- Absolute path (not relative)
- Stored as-is without normalization

**Usage**:
- Display in `usage` command output
- Filter target for `list --location <path>`
- Checked for existence during table output (FR-018)

**Notes**:
- Platform-specific (Windows uses `\`, Unix uses `/`)
- May no longer exist (directory deleted/moved)
- Same logical location may have different paths across machines

---

#### GitRepository (string)

**Type**: Git repository name (historical)
**Purpose**: Git repository name at the time credential was accessed
**Examples**:
- `my-web-app`
- `client-api-service`
- `internal-tools`

**Validation**:
- May be empty (accessed outside git repository)
- Historical value (not updated if repo renamed)

**Usage**:
- Grouping key for `list --by-project`
- Displayed alongside Location in `usage` command
- Used to show project context

**Notes**:
- Immutable historical record (per research.md Decision 5)
- Same project may appear under different names if renamed
- Empty for credentials accessed outside git repositories

---

#### FieldCounts (map[string]int)

**Type**: Map of field names to access counts
**Purpose**: Track which credential fields were accessed and how many times
**Examples**:
```go
{
    "password": 5,
    "username": 2,
    "url": 1
}
```

**Validation**:
- Keys: Valid credential field names (password, username, url, notes, category)
- Values: Non-negative integers

**Usage**:
- Displayed in `usage` command as field-level breakdown
- Shows which parts of credential are actively used
- Helps users understand credential value

**Notes**:
- Cumulative counts (incremented on each access)
- Empty map = credential never accessed at this location

---

#### Timestamp (time.Time)

**Type**: Go `time.Time` (UTC)
**Purpose**: Last access time for this location
**Format**:
- Table output: Human-readable relative time (FR-016) - "2 hours ago", "3 days ago"
- JSON output: ISO 8601 (FR-017) - "2025-10-20T15:30:00Z"

**Validation**:
- Not zero value
- UTC timezone

**Usage**:
- Sorting (most recent first)
- Display in `usage` command
- Filter criteria for `list --unused --days N`

**Notes**:
- Updated on every access
- Used to determine "recent" vs. "stale" credentials

---

#### AccessCount (int)

**Type**: Non-negative integer
**Purpose**: Total number of times credential accessed from this location

**Validation**:
- >= 0 (zero = never accessed, but shouldn't exist)
- Typically >= 1 for existing records

**Usage**:
- Displayed in `usage` command
- Indicates frequency of use at this location

**Notes**:
- Cumulative counter
- Separate from sum of FieldCounts (may access multiple fields per operation)

---

## Entity: Credential (Existing - Partial)

**Source**: `internal/vault/vault.go:50`

**Relevant Field**:
```go
type Credential struct {
    Service string
    // ... other fields ...
    UsageRecords map[string]UsageRecord  // Key: Location (absolute path)
}
```

### UsageRecords (map[string]UsageRecord)

**Type**: Map from location path to UsageRecord
**Purpose**: Store usage history across multiple locations for a single credential

**Key**: Location (absolute path string)
**Value**: UsageRecord struct

**Example**:
```go
UsageRecords: map[string]UsageRecord{
    "/home/user/projects/web-app": UsageRecord{
        Location: "/home/user/projects/web-app",
        GitRepository: "my-web-app",
        FieldCounts: {"password": 5, "username": 2},
        Timestamp: time.Date(2025, 10, 20, 15, 30, 0, 0, time.UTC),
        AccessCount: 7,
    },
    "/home/user/projects/api": UsageRecord{
        Location: "/home/user/projects/api",
        GitRepository: "my-api",
        FieldCounts: {"password": 3},
        Timestamp: time.Date(2025, 10, 15, 10, 0, 0, 0, time.UTC),
        AccessCount: 3,
    },
}
```

**Usage**:
- Iterated by `usage` command to display all locations
- Filtered by `list --location` to match specific paths
- Grouped by `GitRepository` field for `list --by-project`

**Notes**:
- Map size = number of distinct locations where credential accessed
- Empty map = credential never accessed (handled gracefully per FR-014)

---

## Data Relationships

```
Credential (1) ──── has many ──── UsageRecord (N)
                                     │
                                     ├── Location (path)
                                     ├── GitRepository (name)
                                     ├── FieldCounts (map)
                                     ├── Timestamp (time)
                                     └── AccessCount (int)
```

**Cardinality**: One-to-Many
- One credential can be accessed from many locations
- Each location has one UsageRecord per credential
- Keyed by Location (absolute path)

---

## Data Lifecycle

### Creation

**Trigger**: First access to credential from a new location
**Operation**: Create new UsageRecord in Credential.UsageRecords map
**Fields Initialized**:
- Location: Current working directory (absolute path)
- GitRepository: Detected from `.git` directory (if present)
- FieldCounts: Empty map
- Timestamp: Current time (UTC)
- AccessCount: 0 (incremented immediately)

### Update

**Trigger**: Subsequent access to credential from existing location
**Operation**: Update existing UsageRecord in map
**Fields Updated**:
- FieldCounts: Increment count for accessed field(s)
- Timestamp: Current time (UTC)
- AccessCount: Increment by 1

**Fields NOT Updated**:
- Location: Immutable (map key)
- GitRepository: Immutable historical value (per research.md Decision 5)

### Deletion

**Trigger**: **NEVER** (read-only feature per spec Out of Scope)
**Operation**: Not supported
**Note**: UsageRecords are append-only historical data

---

## Data Access Patterns

### Pattern 1: Get All Usage for One Credential

**Operation**: `pass-cli usage <service>`
**Data Access**:
1. Load vault → Unlock → Get credential by service name
2. Iterate `Credential.UsageRecords` map
3. Sort by `Timestamp` (descending - most recent first)
4. Apply `--limit N` truncation (default 20)
5. Format for output (table/JSON)

**Performance**: O(N log N) where N = number of locations (typically < 50)

### Pattern 2: Filter Credentials by Location

**Operation**: `pass-cli list --location /path/to/project`
**Data Access**:
1. Load vault → Unlock → Get all credentials
2. For each credential, check if any UsageRecord.Location matches path
3. Apply `--recursive` flag (prefix matching vs. exact match)
4. Collect matching credentials
5. Format for output

**Performance**: O(C * L) where C = credentials, L = avg locations per credential

### Pattern 3: Group Credentials by Project

**Operation**: `pass-cli list --by-project`
**Data Access**:
1. Load vault → Unlock → Get all credentials
2. For each credential, collect all unique GitRepository values from UsageRecords
3. Group credentials by GitRepository
4. Sort groups by repository name
5. Format for output (show repo name + credential list)

**Performance**: O(C * L) where C = credentials, L = avg locations per credential

### Pattern 4: Combined Filter + Group

**Operation**: `pass-cli list --location /path --by-project`
**Data Access**:
1. Apply Pattern 2 (filter by location) → subset of credentials
2. Apply Pattern 3 (group by project) on filtered subset
3. Format for output

**Performance**: O(C * L) same as Pattern 2/3

---

## Data Constraints

### Validation Rules

1. **UsageRecord.Location**:
   - MUST be non-empty
   - MUST be absolute path (not relative)
   - MAY no longer exist on filesystem

2. **UsageRecord.GitRepository**:
   - MAY be empty (valid for non-git locations)
   - MUST NOT be updated if repository renamed

3. **UsageRecord.FieldCounts**:
   - Keys MUST be valid credential fields
   - Values MUST be non-negative
   - MAY be empty map

4. **UsageRecord.Timestamp**:
   - MUST NOT be zero value
   - MUST be in UTC timezone

5. **UsageRecord.AccessCount**:
   - MUST be >= 0
   - SHOULD be >= 1 for existing records

### Consistency Rules

1. **Map Key Consistency**: `Credential.UsageRecords` map key MUST equal `UsageRecord.Location` value
2. **Uniqueness**: One UsageRecord per (Credential, Location) pair
3. **Historical Integrity**: GitRepository values are immutable (never updated)

---

## Display Formats

### Table Format (Human-Readable)

**usage Command**:
```
Location                              Repository      Last Used        Count   Fields
────────────────────────────────────────────────────────────────────────────────────
/home/user/projects/web-app          my-web-app      2 hours ago      7       password:5, username:2
/home/user/projects/api               my-api          5 days ago       3       password:3
```

**list --by-project**:
```
my-web-app (2 credentials):
  github
  aws-dev

my-api (1 credential):
  heroku

Ungrouped (1 credential):
  local-db
```

### JSON Format (Machine-Readable)

**usage Command**:
```json
{
  "service": "github",
  "usage_locations": [
    {
      "location": "/home/user/projects/web-app",
      "git_repository": "my-web-app",
      "path_exists": true,
      "last_access": "2025-10-20T15:30:00Z",
      "access_count": 7,
      "field_counts": {
        "password": 5,
        "username": 2
      }
    }
  ]
}
```

**list --by-project**:
```json
{
  "projects": {
    "my-web-app": ["github", "aws-dev"],
    "my-api": ["heroku"],
    "Ungrouped": ["local-db"]
  }
}
```

---

## Storage & Persistence

**Storage Format**: JSON (encrypted) in `vault.enc` file
**Persistence**: Automatic via existing VaultService save operations
**Encryption**: AES-256-GCM (existing vault encryption)

**Note**: UsageRecords are part of Credential struct, saved whenever vault is saved. No separate storage needed for this feature.

---

## Summary

**Entities**: 1 existing (UsageRecord)
**New Entities**: 0 (feature uses existing data structures)
**Data Changes**: 0 (read-only feature)
**Storage Changes**: 0 (uses existing vault file format)

**Key Points**:
- All data already collected and persisted
- Feature only adds CLI access to existing data
- No schema changes or migrations needed
- Immutable historical record (read-only)

**References**:
- Source: `internal/vault/vault.go:29-36` (UsageRecord struct)
- Source: `internal/vault/vault.go:50` (Credential.UsageRecords field)
- Research: `research.md` (design decisions)
- Spec: `spec.md` (functional requirements FR-001 through FR-019)
