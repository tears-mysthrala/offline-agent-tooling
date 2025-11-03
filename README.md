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
