#!/usr/bin/env pwsh
# Run all API tests for gat

# Get the directory of this script
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Load test utilities
. (Join-Path $scriptDir "test_utils.ps1")

# Print header
Write-Host
Write-Host "🧪 Running GAT API tests..."
Write-Host "=============================="
Write-Host

# Set environment variable for tests
$env:WORKSPACE_ROOT = (Resolve-Path (Join-Path $scriptDir "..")).Path

# Get all test scripts in the api directory
$testScripts = Get-ChildItem -Path (Join-Path $scriptDir "api") -Filter "test_*.ps1"

# Check if there are any test scripts
if ($testScripts.Count -eq 0) {
    Write-Error "❌ No test scripts found in api directory!"
    exit 1
}

# Keep track of test statistics
$testCount = 0
$passedCount = 0
$failedCount = 0

# Loop through each test script and run it
foreach ($script in $testScripts) {
    $scriptPath = $script.FullName
    $scriptBaseName = $script.BaseName -replace "^test_", ""
    
    # Print script header
    Write-Host "📝 Running test script: $scriptBaseName"
    Write-Host "----------------------------------------"
    
    # Run the test script
    try {
        $results = & $scriptPath
        
        # Process results
        $localTests = 0
        $localPassed = 0
        $localFailed = 0
        
        foreach ($result in $results) {
            $localTests++
            $testCount++
            
            if ($result.Success) {
                $localPassed++
                $passedCount++
            } else {
                $localFailed++
                $failedCount++
            }
        }
        
        # Print summary for this script
        Write-Host "⚡ $localTests tests, $localPassed passed, $localFailed failed"
    }
    catch {
        Write-Error "❌ Error running test script: $_"
        $failedCount++
        $testCount++
    }
    
    Write-Host
}

# Print overall summary
Write-Host "=============================="
Write-Host "📊 Test Summary"
Write-Host "=============================="
Write-Host "Total tests:  $testCount"
Write-Host "Passed:       $passedCount"
Write-Host "Failed:       $failedCount"
Write-Host

# Set exit code based on test results
if ($failedCount -gt 0) {
    Write-Host "❌ Tests failed!" -ForegroundColor Red
    exit 1
} else {
    Write-Host "✅ All tests passed!" -ForegroundColor Green
    exit 0
} 