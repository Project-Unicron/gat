#!/usr/bin/env pwsh
# Improved test runner for gat

# Variables
$TestsDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent $TestsDir
$GatExe = Join-Path $RootDir "gat.exe"

# Ensure gat.exe exists
if (-not (Test-Path $GatExe)) {
    Write-Host "❌ gat.exe not found at $GatExe" -ForegroundColor Red
    Write-Host "Please build the project first with 'go build'" -ForegroundColor Yellow
    exit 1
}

Write-Host "Starting test runner..." -ForegroundColor Green

# Define test files in order
$TestFiles = @(
    "test_01_basic.ps1", 
    "test_02_profiles.ps1", 
    "test_03_platforms.ps1", 
    "test_04_doctor.ps1"
)

$FailedTests = 0
$PassedTests = 0

# Run each test
foreach ($TestFile in $TestFiles) {
    $TestPath = Join-Path $TestsDir $TestFile
    
    if (Test-Path $TestPath) {
        Write-Host "Running test: $TestFile" -ForegroundColor Cyan
        try {
            # Set error action to continue so script doesn't stop on test failures
            $ErrorActionPreference = "Continue"
            
            # Execute the test file
            & $TestPath
            
            # Check if PowerShell reported any errors
            if ($LASTEXITCODE -ne 0) {
                Write-Host "❌ Test file $TestFile exited with code $LASTEXITCODE" -ForegroundColor Red
                $FailedTests++
            } else {
                $PassedTests++
            }
        } catch {
            $errorMsg = $_.Exception.Message
            Write-Host "❌ Exception running test $TestFile`: $errorMsg" -ForegroundColor Red
            $FailedTests++
        } finally {
            # Reset error action preference
            $ErrorActionPreference = "Stop"
            
            # Add separator for clarity
            Write-Host "-----------------------------------------" -ForegroundColor DarkGray
        }
    } else {
        Write-Host "⚠️ Test file not found: $TestFile" -ForegroundColor Yellow
    }
}

# Summary
Write-Host "Test Summary:" -ForegroundColor White
Write-Host "  Passed: $PassedTests" -ForegroundColor Green
if ($FailedTests -gt 0) {
    Write-Host "  Failed: $FailedTests" -ForegroundColor Red
} else {
    Write-Host "  Failed: $FailedTests" -ForegroundColor Green
}
Write-Host "All tests completed" -ForegroundColor Cyan

# Return non-zero exit code if any tests failed
if ($FailedTests -gt 0) {
    exit 1
} else {
    exit 0
} 