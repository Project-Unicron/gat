#!/usr/bin/env pwsh
# Basic tests for gat

# Import test utilities
. (Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) "test_utils.ps1")

# Test if gat executable runs
Test-Report -TestName "Basic: Executable runs" -TestBlock {
    $result = Run-Gat -Arguments @("--help")
    Assert ($result.ExitCode -eq 0) "Gat executable failed to run"
}

# Test help command
Test-Report -TestName "Basic: Help command works" -TestBlock {
    $result = Run-Gat -Arguments @("--help")
    Assert ($result.ExitCode -eq 0) "Help command failed"
    Assert ($result.StdOut -match "GitHub Account Tool") "Help output doesn't contain expected text"
}

# Test various subcommand help
Test-Report -TestName "Basic: Add command help works" -TestBlock {
    $result = Run-Gat -Arguments @("add", "--help")
    Assert ($result.ExitCode -eq 0) "Add help command failed"
    Assert ($result.StdOut -match "add") "Add help output doesn't contain command name"
}

Test-Report -TestName "Basic: Switch command help works" -TestBlock {
    $result = Run-Gat -Arguments @("switch", "--help")
    Assert ($result.ExitCode -eq 0) "Switch help command failed"
    Assert ($result.StdOut -match "switch") "Switch help output doesn't contain command name"
}

Test-Report -TestName "Basic: Platforms command help works" -TestBlock {
    $result = Run-Gat -Arguments @("platforms", "--help")
    Assert ($result.ExitCode -eq 0) "Platforms help command failed"
    Assert ($result.StdOut -match "platforms") "Platforms help output doesn't contain command name"
}

# Test platforms list
Test-Report -TestName "Basic: Platforms list works" -TestBlock {
    $result = Run-Gat -Arguments @("platforms", "list")
    Assert ($result.ExitCode -eq 0) "Platforms list command failed"
    Assert ($result.StdOut -match "github") "Platforms list output doesn't show github"
    Assert ($result.StdOut -match "gitlab") "Platforms list output doesn't show gitlab"
} 