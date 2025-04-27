#!/usr/bin/env pwsh
# Doctor command tests for gat

# Import test utilities
. (Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) "test_utils.ps1")

# Test doctor command with no profile
Test-Report -TestName "Doctor: Basic doctor command" -TestBlock {
    $result = Run-Gat -Arguments @("doctor")
    Assert ($result.ExitCode -eq 0) "Doctor command failed"
    Assert ($result.StdOut -match "Git Account Doctor") "Doctor output doesn't contain expected header"
}

# Create a test profile for further tests
Test-Report -TestName "Doctor: Setup test environment" -TestBlock {
    $result = Create-TestProfile -Name "doctor_test" -Username "doctoruser" -Email "doctor@example.com"
    Assert ($result.ExitCode -eq 0) "Failed to create test profile for doctor tests"
    
    # Switch to the test profile
    $switchResult = Switch-TestProfile -Name "doctor_test"
    Assert ($switchResult.ExitCode -eq 0) "Failed to switch to test profile"
}

# Run doctor with active profile
Test-Report -TestName "Doctor: Check with active profile" -TestBlock {
    $result = Run-Gat -Arguments @("doctor")
    Assert ($result.ExitCode -eq 0) "Doctor command failed with active profile"
    Assert ($result.StdOut -match "doctoruser") "Doctor output doesn't show the current username"
    Assert ($result.StdOut -match "doctor@example.com") "Doctor output doesn't show the current email"
}

# Clean up
Test-Report -TestName "Doctor: Clean up" -TestBlock {
    $result = Remove-TestProfile -Name "doctor_test"
    Assert ($result.ExitCode -eq 0) "Failed to remove doctor test profile"
} 