#!/usr/bin/env pwsh
# Profile management tests for gat

# Import test utilities
. (Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) "test_utils.ps1")

# Clean up any existing test profiles to start fresh
try {
    Remove-TestProfile -Name "test_profile" | Out-Null
    Remove-TestProfile -Name "test_profile2" | Out-Null
    Remove-TestProfile -Name "invalid_profile" | Out-Null
} catch {
    # Ignore errors if profiles don't exist
}

# Create a profile
Test-Report -TestName "Profiles: Create profile" -TestBlock {
    $result = Create-TestProfile -Name "test_profile" -Username "testuser" -Email "test@example.com"
    Assert ($result.ExitCode -eq 0) "Failed to create test profile"
    Assert ($result.StdOut -match "Added profile: test_profile") "Profile creation output doesn't contain expected text"
}

# List profiles
Test-Report -TestName "Profiles: List profiles" -TestBlock {
    $result = Run-Gat -Arguments @("list")
    Assert ($result.ExitCode -eq 0) "List command failed"
    Assert ($result.StdOut -match "test_profile") "Profile list doesn't show the test profile"
    Assert ($result.StdOut -match "testuser") "Profile list doesn't show the test username"
}

# Status command
Test-Report -TestName "Profiles: Check status" -TestBlock {
    $result = Run-Gat -Arguments @("status")
    Assert ($result.ExitCode -eq 0) "Status command failed"
    # Check for either "Current Profile" or "No active profile"
    Assert ($result.StdOut -match "No active profile" -or $result.StdOut -match "Current Profile") "Status output doesn't show profile info"
}

# Switch to profile
Test-Report -TestName "Profiles: Switch to profile" -TestBlock {
    $result = Switch-TestProfile -Name "test_profile"
    
    # The command might fail in test environment if git is not in PATH,
    # but we should still check that the command attempted to switch profiles
    if ($result.ExitCode -ne 0) {
        # Check if the error is just about git not being found
        Assert ($result.StdErr -match "git.*executable file not found" -or $result.StdOut -match "Switching to .* profile") "Failed to switch to test profile"
    } else {
        Assert ($result.StdOut -match "Switched to profile") "Switch output doesn't contain expected text"
    }
}

# Create a second profile
Test-Report -TestName "Profiles: Create second profile" -TestBlock {
    $result = Create-TestProfile -Name "test_profile2" -Username "testuser2" -Email "test2@example.com" -Platform "gitlab"
    Assert ($result.ExitCode -eq 0) "Failed to create second test profile"
    Assert ($result.StdOut -match "Added profile: test_profile2") "Second profile creation output doesn't contain expected text"
}

# Overwrite existing profile
Test-Report -TestName "Profiles: Overwrite existing profile" -TestBlock {
    $result = Create-TestProfile -Name "test_profile" -Username "updated_user" -Email "updated@example.com" -Overwrite
    Assert ($result.ExitCode -eq 0) "Failed to overwrite test profile"
    Assert ($result.StdOut -match "Added profile: test_profile") "Profile overwrite output doesn't contain expected text"
    
    # Verify the profile was updated
    $listResult = Run-Gat -Arguments @("list")
    Assert ($listResult.StdOut -match "updated_user") "Updated username not found in profile list"
}

# Try to create profile with invalid username (should fail)
Test-Report -TestName "Profiles: Reject invalid username" -TestBlock {
    # Use a name with invalid characters - the @ symbol is not allowed in our regex
    $result = Create-TestProfile -Name "invalid_profile" -Username "user-@invalid" -Email "test@example.com"
    # The add command accepts the invalid username, but the switch command should reject it
    $switchResult = Run-Gat -Arguments @("switch", "invalid_profile")
    Assert ($switchResult.ExitCode -ne 0) "Should have failed when switching to profile with invalid username"
    Assert ($switchResult.StdErr -match "invalid GitHub username format") "Error message doesn't mention invalid username format"
}

# Test switch with protocol flags
Test-Report -TestName "Profiles: Switch with protocols" -TestBlock {
    # First with SSH
    $sshResult = Switch-TestProfile -Name "test_profile" -SSH
    
    # We might get failures due to git not being in PATH, but the command should still run
    if ($sshResult.ExitCode -ne 0) {
        Assert ($sshResult.StdErr -match "git.*executable file not found" -or $sshResult.StdOut -match "Switching to") "Failed to switch to test profile with SSH"
    } else {
        Assert ($sshResult.StdOut -match "SSH") "SSH protocol not mentioned in output"
    }
    
    # Then with HTTPS
    $httpsResult = Switch-TestProfile -Name "test_profile" -HTTPS
    
    if ($httpsResult.ExitCode -ne 0) {
        Assert ($httpsResult.StdErr -match "git.*executable file not found" -or $httpsResult.StdOut -match "Switching to") "Failed to switch to test profile with HTTPS"
    } else {
        Assert ($httpsResult.StdOut -match "HTTPS") "HTTPS protocol not mentioned in output"
    }
}

# Remove profiles
Test-Report -TestName "Profiles: Remove profiles" -TestBlock {
    # Remove first profile
    $removeResult1 = Remove-TestProfile -Name "test_profile"
    Assert ($removeResult1.ExitCode -eq 0) "Failed to remove test_profile"
    Assert ($removeResult1.StdOut -match "Profile.*destroyed") "Remove output doesn't contain expected text"
    
    # Remove second profile
    $removeResult2 = Remove-TestProfile -Name "test_profile2"
    Assert ($removeResult2.ExitCode -eq 0) "Failed to remove test_profile2"
    
    # Verify profiles are gone
    $listResult = Run-Gat -Arguments @("list")
    Assert (-not ($listResult.StdOut -match "test_profile")) "test_profile still appears in profile list after removal"
    Assert (-not ($listResult.StdOut -match "test_profile2")) "test_profile2 still appears in profile list after removal"
} 