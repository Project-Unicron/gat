#!/usr/bin/env pwsh
# Profile management tests for gat

# Import test utilities
. (Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) "test_utils.ps1")

# Clean up any existing test profiles to start fresh
try {
    Remove-TestProfile -Name "test_profile" | Out-Null
    Remove-TestProfile -Name "test_profile2" | Out-Null
    Remove-TestProfile -Name "invalid_profile" | Out-Null
    Remove-TestProfile -Name "update_test" | Out-Null
} catch {
    # Ignore errors if profiles don't exist
}

# Create a profile (Defaults to auth_method https)
Test-Report -TestName "Profiles: Create profile (HTTPS)" -TestBlock {
    $result = Create-TestProfile -Name "test_profile" -Username "testuser" -Email "test@example.com"
    Assert ($result.ExitCode -eq 0) "Failed to create test profile (HTTPS)"
    Assert ($result.StdOut -match "Added/Updated profile: test_profile") "Profile creation output doesn't contain expected text"
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

# Switch to profile (test basic switch, relies on config load validation)
Test-Report -TestName "Profiles: Switch to profile" -TestBlock {
    $result = Switch-TestProfile -Name "test_profile"
    # Basic check that the command runs and mentions switching
    Assert ($result.ExitCode -eq 0) "Failed to switch to test profile"
    Assert ($result.StdOut -match "Switched successfully to profile") "Switch output doesn't contain expected text"
}

# Create a second profile (Specify SSH)
Test-Report -TestName "Profiles: Create second profile (SSH)" -TestBlock {
    # Create a dummy SSH key file for testing
    $dummySSHKeyPath = Join-Path $TempDir "test_ssh_key"
    New-Item -Path $dummySSHKeyPath -ItemType File | Out-Null
    New-Item -Path "$dummySSHKeyPath.pub" -ItemType File | Out-Null
    
    $result = Create-TestProfile -Name "test_profile2" -Username "testuser2" -Email "test2@example.com" -Platform "gitlab" -SSHIdentity $dummySSHKeyPath -AuthMethod ssh
    Assert ($result.ExitCode -eq 0) "Failed to create second test profile (SSH)"
    Assert ($result.StdOut -match "Added/Updated profile: test_profile2") "Second profile creation output doesn't contain expected text"
}

# Overwrite existing profile
Test-Report -TestName "Profiles: Overwrite existing profile" -TestBlock {
    # Overwrite with HTTPS again
    $result = Create-TestProfile -Name "test_profile" -Username "updated-user" -Email "updated@example.com" -AuthMethod https -Overwrite
    Assert ($result.ExitCode -eq 0) "Failed to overwrite test profile"
    Assert ($result.StdOut -match "Added/Updated profile: test_profile") "Profile overwrite output doesn't contain expected text"
    
    # Verify the profile was updated
    $listResult = Run-Gat -Arguments @("list")
    Assert ($listResult.StdOut -match "updated-user") "Updated username not found in profile list"
    Assert ($listResult.StdOut -match "Auth Method: https") "Auth method line 'Auth Method: https' not found in profile list after overwrite"
}

# Try to create profile with invalid username (should fail)
# Validation now happens primarily on load, but add should still reject clearly invalid format.
Test-Report -TestName "Profiles: Reject invalid username" -TestBlock {
    # Create profile with invalid username 
    # Username "user-@invalid" is invalid due to special chars AND starting/ending with hyphen
    $result = Create-TestProfile -Name "invalid_user_profile" -Username "user-@invalid" -Email "test@example.com" -AuthMethod https -Overwrite
    Assert ($result.ExitCode -ne 0) "Add command should fail for invalid username format"
    Assert ($result.StdErr -match "invalid username format") "Add command error message doesn't mention invalid username format"
    
    # Since add failed, no cleanup needed for this profile name via gat remove.
    # The config file should not have been modified.
}

# Try to create profile with invalid auth_method (add should fail)
Test-Report -TestName "Profiles: Reject invalid auth_method" -TestBlock {
    $result = Run-Gat -Arguments @("add", "invalid_auth", "--username", "gooduser", "--email", "good@email.com", "--auth-method", "ftp", "--overwrite")
    Assert ($result.ExitCode -ne 0) "Add command should fail for invalid auth_method"
    Assert ($result.StdErr -match "invalid auth_method") "Error message doesn't mention invalid auth_method"
}

# Test removing profile that requires specific auth method
Test-Report -TestName "Profiles: Remove profile with auth_method" -TestBlock {
    # Add a temporary profile with ssh
    $dummySSHKeyPath = Join-Path $TempDir "temp_ssh_key"
    New-Item -Path $dummySSHKeyPath -ItemType File | Out-Null
    New-Item -Path "$dummySSHKeyPath.pub" -ItemType File | Out-Null
    Create-TestProfile -Name "temp_ssh_profile" -Username "tempuser" -Email "temp@example.com" -AuthMethod ssh -SSHIdentity $dummySSHKeyPath -Overwrite | Out-Null

    # Remove it
    $removeResult = Remove-TestProfile -Name "temp_ssh_profile"
    Assert ($removeResult.ExitCode -eq 0) "Failed to remove profile with auth_method 'ssh'"
    Assert ($removeResult.StdOut -match "destroyed") "Remove output missing expected text"

    # Verify removal
    $listResult = Run-Gat -Arguments @("list")
    Assert (-not ($listResult.StdOut -match "temp_ssh_profile")) "Profile temp_ssh_profile still appears after removal"
}

# Remove profiles (Clean up remaining test_profile2)
Test-Report -TestName "Profiles: Final Cleanup" -TestBlock {
    # Remove second profile (created with SSH)
    $removeResult = Remove-TestProfile -Name "test_profile2"
    Assert ($removeResult.ExitCode -eq 0) "Failed to remove test_profile2"
    
    # Verify profiles are gone
    $listResult = Run-Gat -Arguments @("list")
    Assert (-not ($listResult.StdOut -match "test_profile2")) "test_profile2 still appears in profile list after removal"
}

# Test updating profile fields with add --overwrite
Test-Report -TestName "Profiles: Update fields with add --overwrite" -TestBlock {
    # 1. Create initial profile
    $initialResult = Create-TestProfile -Name "update_test" -Username "user1" -Email "email1@test.com" -Platform "github" -AuthMethod "https" -Token "token1"
    Assert ($initialResult.ExitCode -eq 0) "Failed to create initial profile for update test"

    # 2. Update only username
    $updateUserResult = Run-Gat -Arguments @("add", "update_test", "--username", "user2", "--overwrite")
    Assert ($updateUserResult.ExitCode -eq 0) "Failed to update username with add --overwrite"
    Assert ($updateUserResult.StdOut -match "Added/Updated profile: update_test") "Update output message incorrect"

    # 3. Verify username updated, others preserved
    $listResult1 = Run-Gat -Arguments @("list")
    $lines1 = $listResult1.StdOut -split '(?:\r?\n)'
    $startMatch1 = $lines1 | Select-String -Pattern 'update_test' -List | Select-Object -First 1
    $profileBlock1 = ""
    if ($startMatch1) {
        $startIndex1 = $startMatch1.LineNumber - 1 # LineNumber is 1-based, array is 0-based
        $profileBlockLines1 = New-Object System.Collections.Generic.List[string]
        $profileBlockLines1.Add($lines1[$startIndex1])
        for ($i = $startIndex1 + 1; $i -lt $lines1.Length; $i++) {
            if ($lines1[$i] -match '^\s+') {
                $profileBlockLines1.Add($lines1[$i])
            } else {
                break # Stop when a non-indented line is found
            }
        }
        $profileBlock1 = $profileBlockLines1 -join "`n"
    }

    Assert ($profileBlock1 -ne "") "Could not extract profile block for update_test in step 3"
    Assert ($profileBlock1 -match "Username: user2") "Username was not updated to user2"
    Assert ($profileBlock1 -match "Email: email1@test.com") "Email was not preserved after username update"
    Assert ($profileBlock1 -match "Platform: GitHub") "Platform was not preserved after username update"
    Assert ($profileBlock1 -match "Auth Method: https") "AuthMethod was not preserved after username update"
    # Token is not listed, implicitly verified by AuthMethod remaining https

    # 4. Update email and add SSH key (should switch auth method)
    $dummySSHKeyPathUpdate = Join-Path $TempDir "update_test_ssh_key"
    New-Item -Path $dummySSHKeyPathUpdate -ItemType File -Force | Out-Null # Force creation/overwrite
    $updateEmailSSHResult = Run-Gat -Arguments @("add", "update_test", "--email", "email2@test.com", "--ssh-identity", $dummySSHKeyPathUpdate, "--overwrite")
    Assert ($updateEmailSSHResult.ExitCode -eq 0) "Failed to update email/ssh with add --overwrite"
    Assert ($updateEmailSSHResult.StdOut -match "Added/Updated profile: update_test") "Update output message incorrect for email/ssh"

    # 5. Verify email/ssh updated, username preserved, auth method switched
    $listResult2 = Run-Gat -Arguments @("list")
    $lines2 = $listResult2.StdOut -split '(?:\r?\n)'
    $startMatch2 = $lines2 | Select-String -Pattern 'update_test' -List | Select-Object -First 1
    $profileBlock2 = ""
    if ($startMatch2) {
        $startIndex2 = $startMatch2.LineNumber - 1 # LineNumber is 1-based, array is 0-based
        $profileBlockLines2 = New-Object System.Collections.Generic.List[string]
        $profileBlockLines2.Add($lines2[$startIndex2])
        for ($i = $startIndex2 + 1; $i -lt $lines2.Length; $i++) {
            if ($lines2[$i] -match '^\s+') {
                $profileBlockLines2.Add($lines2[$i])
            } else {
                break # Stop when a non-indented line is found
            }
        }
        $profileBlock2 = $profileBlockLines2 -join "`n"
    }

    Assert ($profileBlock2 -ne "") "Could not extract profile block for update_test in step 5"
    Assert ($profileBlock2 -match "Username: user2") "Username was not preserved after email/ssh update"
    Assert ($profileBlock2 -match "Email: email2@test.com") "Email was not updated to email2"
    Assert ($profileBlock2 -match "Auth Method: ssh") "AuthMethod did not switch to ssh after adding key"
    Assert ($profileBlock2 -match "SSH Key: .*update_test_ssh_key") "SSH Key path was not updated"
    # Token should still be present internally but auth method is now ssh

    # 6. Cleanup
    $cleanupResult = Remove-TestProfile -Name "update_test"
    Assert ($cleanupResult.ExitCode -eq 0) "Failed to cleanup update_test profile"
    if (Test-Path $dummySSHKeyPathUpdate) { Remove-Item $dummySSHKeyPathUpdate -Force }
}

# Remove dummy SSH keys if they exist
$dummySSHKeyPath1 = Join-Path $TempDir "test_ssh_key"
if (Test-Path $dummySSHKeyPath1) { Remove-Item $dummySSHKeyPath1 }
if (Test-Path "$dummySSHKeyPath1.pub") { Remove-Item "$dummySSHKeyPath1.pub" }
$dummySSHKeyPath2 = Join-Path $TempDir "temp_ssh_key"
if (Test-Path $dummySSHKeyPath2) { Remove-Item $dummySSHKeyPath2 }
if (Test-Path "$dummySSHKeyPath2.pub") { Remove-Item "$dummySSHKeyPath2.pub" } 