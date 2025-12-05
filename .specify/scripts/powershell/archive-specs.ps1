#!/usr/bin/env pwsh

# Stop on errors
$ErrorActionPreference = "Stop"

# Source common functions
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
. "$ScriptDir\common.ps1"

# Parse arguments
$JsonMode = $false
$DryRun = $false
$ListMode = $false
$CompletedMode = $false
$SpecNumbers = @()

foreach ($arg in $args) {
    switch -Regex ($arg) {
        '^(--json|-j)$' {
            $JsonMode = $true
        }
        '^(--list|-l)$' {
            $ListMode = $true
        }
        '^(--dry-run|-n)$' {
            $DryRun = $true
        }
        '^(--completed|-c)$' {
            $CompletedMode = $true
        }
        '^(--help|-h)$' {
            Write-Host "Usage: $($MyInvocation.MyCommand.Name) [OPTIONS] [SPEC_NUMBERS...]"
            Write-Host ""
            Write-Host "Archive completed or old spec directories to specs/archive/"
            Write-Host ""
            Write-Host "Options:"
            Write-Host "  --json, -j         Output in JSON format"
            Write-Host "  --list, -l         List all specs with their status"
            Write-Host "  --dry-run, -n      Show what would be archived without moving files"
            Write-Host "  --completed, -c    Archive all completed specs"
            Write-Host "  --help, -h         Show this help message"
            Write-Host ""
            Write-Host "Examples:"
            Write-Host "  $($MyInvocation.MyCommand.Name) --list                 # List all specs with status"
            Write-Host "  $($MyInvocation.MyCommand.Name) 001 003                # Archive specs 001 and 003"
            Write-Host "  $($MyInvocation.MyCommand.Name) --dry-run 001 003      # Preview archiving 001 and 003"
            Write-Host "  $($MyInvocation.MyCommand.Name) --completed            # Archive all completed specs"
            exit 0
        }
        '^[0-9]{3}$' {
            $SpecNumbers += $arg
        }
        default {
            Write-Error "Error: Unknown argument '$arg'. Use --help for usage information"
            exit 1
        }
    }
}

# Get repository paths
$RepoRoot = Get-RepoRoot
$SpecsDir = Join-Path $RepoRoot "specs"
$ArchiveDir = Join-Path $SpecsDir "archive"

# Ensure specs directory exists
if (-not (Test-Path $SpecsDir -PathType Container)) {
    if ($JsonMode) {
        Write-Host '{"error":"specs directory not found"}'
    } else {
        Write-Error "Error: specs directory not found at $SpecsDir"
    }
    exit 1
}

# Function to check if a spec is completed by parsing tasks.md
function Test-SpecCompleted {
    param([string]$SpecDir)

    $TasksFile = Join-Path $SpecDir "tasks.md"

    if (-not (Test-Path $TasksFile)) {
        return $false
    }

    # Extract all task lines (lines that start with - [ ] or - [X])
    $Content = Get-Content $TasksFile -Raw -ErrorAction SilentlyContinue
    if (-not $Content) {
        return $false
    }

    $AllTasks = [regex]::Matches($Content, '^\s*-\s+\[[Xx ]\]', 'Multiline')
    $CompletedTasks = [regex]::Matches($Content, '^\s*-\s+\[[Xx]\]', 'Multiline')

    $TotalTasks = $AllTasks.Count
    $CompletedCount = $CompletedTasks.Count

    # Spec is completed if it has tasks and all are marked as complete
    return ($TotalTasks -gt 0 -and $CompletedCount -eq $TotalTasks)
}

# Function to get spec info
function Get-SpecInfo {
    param([string]$SpecDir)

    $SpecName = Split-Path $SpecDir -Leaf
    $SpecNumber = "???"
    $SpecSlug = $SpecName
    $Status = "IN PROGRESS"

    # Extract number and slug from spec name (format: 001-spec-slug)
    if ($SpecName -match '^([0-9]{3})-(.+)$') {
        $SpecNumber = $Matches[1]
        $SpecSlug = $Matches[2]
    }

    # Check if completed
    if (Test-SpecCompleted $SpecDir) {
        $Status = "COMPLETED"
    }

    if ($JsonMode) {
        $JsonObj = @{
            number = $SpecNumber
            slug = $SpecSlug
            name = $SpecName
            status = $Status
            path = $SpecDir
        } | ConvertTo-Json -Compress
        return $JsonObj
    } else {
        return "$SpecNumber - $SpecSlug [$Status]"
    }
}

# Function to list all specs
function Show-SpecList {
    $Specs = @()

    if (Test-Path $SpecsDir -PathType Container) {
        $Specs = Get-ChildItem $SpecsDir -Directory | Where-Object { $_.Name -ne "archive" }
    }

    if ($Specs.Count -eq 0) {
        if ($JsonMode) {
            Write-Host '{"specs":[]}'
        } else {
            Write-Host "No specs found in $SpecsDir"
        }
        return
    }

    if ($JsonMode) {
        $SpecsArray = @()
        foreach ($spec in $Specs) {
            $info = Get-SpecInfo $spec.FullName
            $SpecsArray += ($info | ConvertFrom-Json)
        }
        $result = @{
            specs = $SpecsArray
            archive_location = $ArchiveDir
        } | ConvertTo-Json -Depth 10 -Compress
        Write-Host $result
    } else {
        Write-Host "Available Specs:"
        Write-Host "================"
        foreach ($spec in $Specs) {
            Write-Host "  $(Get-SpecInfo $spec.FullName)"
        }
        Write-Host ""
        Write-Host "Archive location: $ArchiveDir"
        Write-Host ""
        Write-Host "Usage:"
        Write-Host "  - Archive specific specs: /speckit.archive 001 003"
        Write-Host "  - Archive all completed: /speckit.archive --completed"
        Write-Host "  - Dry run: /speckit.archive --dry-run 001 002"
    }
}

# Function to archive a spec
function Move-SpecToArchive {
    param([string]$SpecNumber)

    $SourceDir = $null

    # Find the spec directory
    $Matches = Get-ChildItem $SpecsDir -Directory | Where-Object { $_.Name -match "^$SpecNumber-" }
    if ($Matches) {
        $SourceDir = $Matches[0].FullName
    }

    if (-not $SourceDir) {
        if ($JsonMode) {
            $result = @{
                number = $SpecNumber
                status = "error"
                message = "Spec not found"
            } | ConvertTo-Json -Compress
            Write-Host $result
        } else {
            Write-Error "Error: Spec $SpecNumber not found"
        }
        return $false
    }

    $SpecName = Split-Path $SourceDir -Leaf
    $DestDir = Join-Path $ArchiveDir $SpecName

    if ($DryRun) {
        if ($JsonMode) {
            $result = @{
                number = $SpecNumber
                name = $SpecName
                status = "dry-run"
                from = $SourceDir
                to = $DestDir
            } | ConvertTo-Json -Compress
            Write-Host $result
        } else {
            $displayName = $SpecName -replace "^$SpecNumber-", ""
            Write-Host "  Would archive: $SpecNumber - $displayName → archive/$SpecName"
        }
        return $true
    }

    # Create archive directory if it doesn't exist
    if (-not (Test-Path $ArchiveDir)) {
        New-Item -ItemType Directory -Path $ArchiveDir -Force | Out-Null
    }

    # Move the spec to archive
    try {
        Move-Item -Path $SourceDir -Destination $DestDir -Force

        if ($JsonMode) {
            $result = @{
                number = $SpecNumber
                name = $SpecName
                status = "archived"
                from = $SourceDir
                to = $DestDir
            } | ConvertTo-Json -Compress
            Write-Host $result
        } else {
            $displayName = $SpecName -replace "^$SpecNumber-", ""
            Write-Host "  $SpecNumber - $displayName [ARCHIVED]"
        }
        return $true
    } catch {
        if ($JsonMode) {
            $result = @{
                number = $SpecNumber
                status = "error"
                message = "Failed to move spec: $($_.Exception.Message)"
            } | ConvertTo-Json -Compress
            Write-Host $result
        } else {
            Write-Error "Error: Failed to archive spec $SpecNumber : $($_.Exception.Message)"
        }
        return $false
    }
}

# Main execution logic

# If no arguments or --list flag, show list
if (($SpecNumbers.Count -eq 0 -and -not $CompletedMode) -or $ListMode) {
    Show-SpecList
    exit 0
}

# Check if on main branch (only for git repos)
if (Test-HasGit) {
    $CurrentBranch = Get-CurrentBranch
    if ($CurrentBranch -ne "main") {
        if (-not $JsonMode) {
            Write-Warning "You are on branch '$CurrentBranch', not 'main'"
            Write-Host ""
            Write-Host "Archiving specs on a feature branch can cause inconsistencies:"
            Write-Host "  - Archive state will differ between branches"
            Write-Host "  - Potential merge conflicts when updating specs/ directory"
            Write-Host "  - Confusion about which specs exist"
            Write-Host ""
            Write-Host "Recommended: Switch to main branch first"
            Write-Host ""
            $response = Read-Host "Continue anyway? (yes/no)"
            if ($response -notmatch '^[Yy][Ee][Ss]$') {
                Write-Host "Aborted."
                exit 0
            }
        }
    }
}

# Handle --completed mode
if ($CompletedMode) {
    $CompletedSpecs = @()

    $AllSpecs = Get-ChildItem $SpecsDir -Directory | Where-Object { $_.Name -ne "archive" }
    foreach ($spec in $AllSpecs) {
        if (Test-SpecCompleted $spec.FullName) {
            $SpecName = $spec.Name
            if ($SpecName -match '^([0-9]{3})-') {
                $CompletedSpecs += $Matches[1]
            }
        }
    }

    if ($CompletedSpecs.Count -eq 0) {
        if ($JsonMode) {
            Write-Host '{"message":"No completed specs found"}'
        } else {
            Write-Host "No completed specs found."
        }
        exit 0
    }

    # Show what will be archived and ask for confirmation
    if (-not $JsonMode -and -not $DryRun) {
        Write-Host "Found completed specs ready to archive:"
        foreach ($specNum in $CompletedSpecs) {
            $specDir = Get-ChildItem $SpecsDir -Directory | Where-Object { $_.Name -match "^$specNum-" } | Select-Object -First 1
            if ($specDir) {
                $displayName = $specDir.Name -replace "^$specNum-", ""
                Write-Host "  - $specNum - $displayName"
            }
        }
        Write-Host ""
        $response = Read-Host "Archive these specs? (yes/no)"
        if ($response -notmatch '^[Yy][Ee][Ss]$') {
            Write-Host "Aborted."
            exit 0
        }
    }

    $SpecNumbers = $CompletedSpecs
}

# Archive the specified specs
if ($JsonMode) {
    $archivedResults = @()
    foreach ($specNum in $SpecNumbers) {
        $result = Move-SpecToArchive $specNum
        # Result is already written in JSON format by the function
    }
} else {
    if ($DryRun) {
        Write-Host "DRY RUN - No specs will be moved"
        Write-Host "================================="
        Write-Host "Would archive:"
    } else {
        Write-Host "Archiving specs: $($SpecNumbers -join ' ')"
        Write-Host ""
        Write-Host "Archive Complete:"
        Write-Host "================="
    }

    foreach ($specNum in $SpecNumbers) {
        Move-SpecToArchive $specNum | Out-Null
    }

    if ($DryRun) {
        Write-Host ""
        Write-Host "To perform archive, run: /speckit.archive $($SpecNumbers -join ' ')"
    } else {
        Write-Host ""
        Write-Host "Archived to: $ArchiveDir"
    }
}
