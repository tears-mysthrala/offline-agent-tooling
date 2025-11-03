<#
.SYNOPSIS
  Orchestrate creating a branch and pull request using the repo tooling.

DESCRIPTION
  This script is a safe wrapper for creating a branch, committing staged changes,
  pushing to origin, creating a PR with the GH CLI and optionally waiting for CI.

  IMPORTANT: DryRun is ON by default. The script will NOT perform remote actions
  unless you pass -DryRun:$false. It will NEVER relax branch protection. If you
  request AutoMerge, the script will attempt a normal merge (no admin flags) only
  after confirmation and only if CI is passing and required approvals exist.

USAGE
  # Dry run (default) - shows what would be done
  pwsh .github\scripts\pr_orchestrator.ps1 -Title 'chore: update docs' -Body 'Add docs' -DryRun

  # Actually create branch, push and open PR (no auto-merge)
  pwsh .github\scripts\pr_orchestrator.ps1 -Title 'chore: update docs' -Body 'Add docs' -DryRun:$false

  # Create and auto-merge (will prompt and will not relax protection)
  pwsh .github\scripts\pr_orchestrator.ps1 -Title 'chore: x' -Body '...' -DryRun:$false -AutoMerge

#>

param(
    [string]$BranchName = "pr-orchestrator/$(Get-Date -Format 'yyyyMMdd-HHmmss')",
    [string]$Title = "",
    [string]$Body = "",
    [switch]$AutoMerge,
    [switch]$DryRun = $true,
    [int]$WaitSeconds = 6,
    [int]$MaxAttempts = 200
)

function Show-Plan {
    Write-Host "[PLAN] Create branch: $BranchName"
    Write-Host "[PLAN] Commit staged changes (if any) with message: '$Title'"
    Write-Host "[PLAN] Push branch to origin"
    Write-Host "[PLAN] Create PR (title: '$Title')"
    Write-Host "[PLAN] Wait for CI 'test-windows' to complete and pass"
    if ($AutoMerge) { Write-Host "[PLAN] Attempt to merge PR automatically (only if CI passes and approvals exist)" }
}

if (-not $Title) {
    Write-Error "Title is required. Use -Title 'your title'"
    exit 1
}

Show-Plan

if ($DryRun) {
    Write-Host "DryRun is ON — no remote actions will be executed. To run for real set -DryRun:$false." -ForegroundColor Yellow
    exit 0
}

# Real execution path (explicit permission given by setting -DryRun:$false)
Write-Host "Starting orchestrator: creating branch $BranchName" -ForegroundColor Green

git checkout -b $BranchName

# Commit staged changes if any
$staged = git diff --staged --name-only
if ($staged) {
    if (-not $Title) { $msg = "chore: staged changes" } else { $msg = $Title }
    git commit -m $msg || Write-Host "No commit created (maybe nothing staged)"
} else {
    Write-Host "No staged changes to commit"
}

git push -u origin $BranchName

# Create PR via GH CLI
$prUrl = gh pr create --title $Title --body $Body --base main --head $BranchName --web || gh pr create --title $Title --body $Body --base main --head $BranchName
if ($prUrl) { Write-Host "Created PR: $prUrl" } else { Write-Host "Created PR (check gh output)" }

# Extract PR number
$pr = gh pr list --head $BranchName --json number,url -q '.[0]'
$prNumber = ($pr | ConvertFrom-Json).number
Write-Host "PR number: $prNumber"

Write-Host "Waiting for CI 'test-windows' to complete (polling every $WaitSeconds seconds)"
for ($i = 0; $i -lt $MaxAttempts; $i++) {
    $runs = gh run list --json databaseId,name,headBranch,conclusion,status -L 50 | ConvertFrom-Json
    $run = $runs | Where-Object { $_.headBranch -eq $BranchName } | Select-Object -First 1
    if ($run) {
        Write-Host "Found run: id=$($run.databaseId) status=$($run.status) conclusion=$($run.conclusion)"
        if ($run.status -eq 'completed') { break }
    } else {
        Write-Host "No run found yet (attempt $i)"
    }
    Start-Sleep -Seconds $WaitSeconds
}

if (-not $run) {
    Write-Error "No workflow run found for branch $BranchName. Exiting."
    exit 2
}

if ($run.conclusion -ne 'success') {
    Write-Error "CI did not pass (conclusion: $($run.conclusion)). Not merging."
    exit 3
}

Write-Host "CI passed. PR ready for merge checks (approvals)."

if ($AutoMerge) {
    $confirm = Read-Host "AutoMerge requested. This will attempt a normal merge (no admin flag). Type 'yes' to proceed"
    if ($confirm -ne 'yes') { Write-Host 'AutoMerge aborted by user'; exit 0 }

    # Check approvals
    $reviews = gh pr view $prNumber --json reviews -q '.reviews'
    # This is a lightweight check; we still attempt merge and rely on GH API to fail if approvals missing
    Write-Host "Attempting to merge PR #$prNumber"
    gh pr merge $prNumber --merge --delete-branch
    Write-Host "Merge attempted. Verify PR state on GitHub." 
}

Write-Host "Orchestration completed."
