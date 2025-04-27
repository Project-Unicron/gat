# 🧪 gat Test Suite

This directory contains automated tests for the GitHub Account Tool (gat).

## 🏃‍♂️ Running Tests

### All Tests

To run all tests, execute the PowerShell script:

```powershell
.\run_all_tests.ps1
```

This will:
1. Back up your existing gat configuration (if any)
2. Run all test scripts in order
3. Restore your original configuration
4. Display a summary of test results

### Individual Tests

You can also run individual test files:

```powershell
.\test_01_basic.ps1
.\test_02_profiles.ps1
# etc.
```

## 📋 Test Organization

Tests are organized by logical functionality:

1. **Basic Tests** (`test_01_basic.ps1`): Basic functionality, help commands, etc.
2. **Profile Tests** (`test_02_profiles.ps1`): Creating, listing, switching, and managing profiles
3. **Platform Tests** (`test_03_platforms.ps1`): Platform management functionality
4. **Doctor Tests** (`test_04_doctor.ps1`): Testing the doctor command

## ➕ Adding New Tests

To add a new test:

1. Create a new file named `test_XX_name.ps1` where `XX` is a number (to control test order)
2. Import test utilities: `. (Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) "test_utils.ps1")`
3. Create test cases using the `Test-Report` function

Example:

```powershell
#!/usr/bin/env pwsh
# Description of your test

# Import test utilities
. (Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) "test_utils.ps1")

# Test case
Test-Report -TestName "Feature: Test scenario" -TestBlock {
    $result = Run-Gat -Arguments @("command", "subcommand", "--flag")
    Assert ($result.ExitCode -eq 0) "Command failed"
    Assert ($result.StdOut -match "expected output") "Output doesn't contain expected text"
}
```

## 🧰 Utility Functions

The `test_utils.ps1` file provides several helper functions:

- `Run-Gat`: Run a gat command and capture output
- `Assert`: Check a condition and throw an error if it fails
- `Create-TestProfile`: Create a profile for testing
- `Remove-TestProfile`: Remove a test profile
- `Switch-TestProfile`: Switch to a specific profile
- `Register-TestPlatform`: Register a custom platform
- `Test-Report`: Run a test and report results

## ⚠️ Notes

- Tests create temporary profiles and platforms that should be automatically cleaned up
- If tests are interrupted, your original configuration will be restored from backup
- Some tests may fail if Git is not properly configured on your system 