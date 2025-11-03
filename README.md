# Offline Agent Tooling

Local/offline-first agent tools and tests for testing local automation and tooling.

See `scripts/command-runner.ps1` to run the curated command sets and `scripts/commands-tests.txt` for the test suite.

## Try it locally

Run the curated test set (PowerShell on Windows):

```powershell
pwsh -NoProfile -File scripts/command-runner.ps1 -Tests
```

To run the full commands queue (be careful — may contain long-running tasks):

```powershell
pwsh -NoProfile -File scripts/command-runner.ps1
```

## Run tests and generate reports locally

Python (recommended inside a virtualenv):

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
python -m pip install -r requirements-dev.txt
pytest tests/python --junitxml=test-reports/python-junit.xml
```

PowerShell / Pester (Windows / pwsh):

```powershell
# Install Pester if you don't have it (run as current user)
Install-Module -Name Pester -Scope CurrentUser -Force -AllowClobber
# Run Pester and produce NUnit XML
Invoke-Pester -Script tests/ps/pester -OutputFormat NUnitXml -OutputFile test-reports/pester-nunit.xml
```

Test reports will be available under the `test-reports/` directory and are uploaded by CI as artifacts.

## Safe PR orchestration (use repo tooling)

Use the included orchestrator script to create branches and PRs safely. The script defaults to a DryRun (no remote actions). It will never relax branch protection. Run it from the repo root:

```powershell
# Dry run (shows plan, no remote actions)
pwsh .github\scripts\pr_orchestrator.ps1 -Title 'chore: update docs' -Body 'Add docs'

# Execute (create branch, push, open PR) — must pass -DryRun:$false
pwsh .github\scripts\pr_orchestrator.ps1 -Title 'chore: update docs' -Body 'Add docs' -DryRun:$false
```

If you want an automated merge attempt after CI passes, use `-AutoMerge` and the script will prompt for explicit confirmation. The script will not perform admin bypasses or auto-relax protection.

This is the preferred flow instead of running ad-hoc GH/git commands in the terminal when you want repeatable, auditable changes.
