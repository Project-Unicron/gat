#!/usr/bin/env pwsh
# Switch command tests for gat v2

# Import test utilities
. (Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) "test_utils.ps1")

# --- Test Setup ---
# Clean up any existing test profiles/repos
Remove-TestProfile -Name "switch_ssh_test" | Out-Null
Remove-TestProfile -Name "switch_https_test" | Out-Null
Remove-TestProfile -Name "switch_ssh_nokey" | Out-Null
Remove-TestProfile -Name "switch_https_notoken" | Out-Null
if (Test-Path $TestGitRepoPath) { Remove-Item -Recurse -Force $TestGitRepoPath }

# Create dummy SSH key (ensure cleanup first)
$dummySSHKeyPath = Join-Path $TempDir "switch_test_ssh_key"
if (Test-Path $dummySSHKeyPath) { Remove-Item -Path $dummySSHKeyPath -Force }
if (Test-Path "$dummySSHKeyPath.pub") { Remove-Item -Path "$dummySSHKeyPath.pub" -Force }
New-Item -Path $dummySSHKeyPath -ItemType File | Out-Null
New-Item -Path "$dummySSHKeyPath.pub" -ItemType File | Out-Null

# Create profiles for testing
Create-TestProfile -Name "switch_ssh_test" -Username "sshuser" -Email "ssh@test.com" -AuthMethod ssh -SSHIdentity $dummySSHKeyPath | Out-Null
Create-TestProfile -Name "switch_https_test" -Username "httpsuser" -Email "https@test.com" -AuthMethod https -Token "https_token" | Out-Null
Create-TestProfile -Name "switch_ssh_nokey" -Username "sshnokey" -Email "sshnokey@test.com" -AuthMethod ssh -SSHIdentity "~/nonexistent/key" | Out-Null
# Pass an explicit empty string for the token argument here to avoid potential issues
Create-TestProfile -Name "switch_https_notoken" -Username "httpsnotoken" -Email "httpsnotoken@test.com" -AuthMethod https -Token "" | Out-Null

# Create a dummy Git repository for testing remote rewrite
Initialize-TestGitRepo

# --- Tests ---

# Test switch to SSH profile
Test-Report -TestName "Switch: Basic SSH switch" -TestBlock {
    $result = Switch-TestProfile -Name "switch_ssh_test"
    Assert ($result.ExitCode -eq 0) "Switch command failed for SSH profile"
    Assert ($result.StdOut -match "Switched successfully to profile: switch_ssh_test") "Success message missing"
    Assert ($result.StdOut -match "Handling SSH Configuration") "SSH handling message missing"
    # Basic check for agent interaction messages
    Assert ($result.StdOut -match "Starting ssh-agent" -or $result.StdOut -match "Clearing existing SSH identities" -or $result.StdOut -match "Adding SSH identity") "SSH Agent message missing"
    Assert ($result.StdOut -match "Handling Git Remote URL") "Remote handling message missing"
    Assert ($result.StdOut -match "Remote 'origin' set to use SSH") "Remote rewrite to SSH message missing"

    # Verify remote URL was rewritten
    $remoteUrl = Get-TestGitRemoteUrl
    Assert ($remoteUrl -match "^git@github-switch_ssh_test:") "Remote URL was not rewritten to profile-specific SSH format"
}

# Test switch to HTTPS profile
Test-Report -TestName "Switch: Basic HTTPS switch" -TestBlock {
    $result = Switch-TestProfile -Name "switch_https_test"
    Assert ($result.ExitCode -eq 0) "Switch command failed for HTTPS profile"
    Assert ($result.StdOut -match "Switched successfully to profile: switch_https_test") "Success message missing"
    Assert ($result.StdOut -match "Handling HTTPS Configuration") "HTTPS handling message missing"
    Assert ($result.StdOut -match "Git credentials updated") "Credential update message missing"
    Assert ($result.StdOut -match "Handling Git Remote URL") "Remote handling message missing"
    Assert ($result.StdOut -match "Remote 'origin' set to use HTTPS") "Remote rewrite to HTTPS message missing"

    # Verify remote URL was rewritten
    $remoteUrl = Get-TestGitRemoteUrl
    Assert ($remoteUrl -match "^https://github.com/") "Remote URL was not rewritten to HTTPS format"
}

# Test switch with missing SSH key file (should warn but succeed)
Test-Report -TestName "Switch: SSH profile with missing key" -TestBlock {
    $result = Switch-TestProfile -Name "switch_ssh_nokey"
    Assert ($result.ExitCode -eq 0) "Switch command failed for SSH profile with missing key"
    Assert ($result.StdOut -match "Switched successfully to profile: switch_ssh_nokey") "Success message missing"
    Assert ($result.StdOut -match "SSH identity file not found") "Warning about missing SSH key missing"
}

# Test switch with missing HTTPS token (should warn but succeed)
Test-Report -TestName "Switch: HTTPS profile with missing token" -TestBlock {
    $result = Switch-TestProfile -Name "switch_https_notoken"
    Assert ($result.ExitCode -eq 0) "Switch command failed for HTTPS profile with missing token"
    Assert ($result.StdOut -match "Switched successfully to profile: switch_https_notoken") "Success message missing"
    Assert ($result.StdOut -match "uses HTTPS but has no token configured") "Warning about missing HTTPS token missing"
}

# Test switch with Dry Run
Test-Report -TestName "Switch: Dry run" -TestBlock {
    $result = Switch-TestProfile -Name "switch_ssh_test" -DryRun
    Assert ($result.ExitCode -eq 0) "Dry run switch command failed"
    Assert ($result.StdOut -match "Dry run mode enabled") "Dry run message missing"
    Assert ($result.StdOut -match "Would set Git User: sshuser") "Dry run output for user incorrect"
    Assert ($result.StdOut -match "Would manage SSH Key") "Dry run output for SSH key incorrect"
    Assert ($result.StdOut -match "Would ensure remote uses: SSH") "Dry run output for remote incorrect"
    # Verify no actual change happened (check remote URL)
    $remoteUrl = Get-TestGitRemoteUrl
    # Should still be HTTPS from the previous test run
    Assert ($remoteUrl -match "^https://github.com/") "Remote URL was changed during dry run"
}

# --- Test Cleanup ---
Remove-TestProfile -Name "switch_ssh_test" | Out-Null
Remove-TestProfile -Name "switch_https_test" | Out-Null
Remove-TestProfile -Name "switch_ssh_nokey" | Out-Null
Remove-TestProfile -Name "switch_https_notoken" | Out-Null
# Ensure cleanup uses -Force
if (Test-Path $dummySSHKeyPath) { Remove-Item -Path $dummySSHKeyPath -Force }
if (Test-Path "$dummySSHKeyPath.pub") { Remove-Item -Path "$dummySSHKeyPath.pub" -Force }
if (Test-Path $TestGitRepoPath) { Remove-Item -Recurse -Force $TestGitRepoPath } 