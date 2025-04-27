# Test utilities for gat tests

# Variables
$TestsDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent $TestsDir
$GatExe = Join-Path $RootDir "gat.exe"
$TempDir = Join-Path $TestsDir "temp"

# Ensure temp directory exists
if (-not (Test-Path $TempDir)) {
    New-Item -ItemType Directory -Path $TempDir | Out-Null
}

# Run gat command and return output
function Run-Gat {
    param (
        [Parameter(Mandatory=$true)]
        [string[]]$Arguments,
        
        [int]$TimeoutSeconds = 30
    )
    
    Write-Host "DEBUG: Running gat with args: $Arguments" -ForegroundColor DarkGray
    
    $processInfo = New-Object System.Diagnostics.ProcessStartInfo
    $processInfo.FileName = $GatExe
    $processInfo.RedirectStandardOutput = $true
    $processInfo.RedirectStandardError = $true
    $processInfo.UseShellExecute = $false
    $processInfo.Arguments = $Arguments
    
    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $processInfo
    $process.Start() | Out-Null
    
    # Add timeout
    if (-not $process.WaitForExit($TimeoutSeconds * 1000)) {
        Write-Host "ERROR: Process timed out after $TimeoutSeconds seconds" -ForegroundColor Red
        try {
            $process.Kill()
        } catch {
            Write-Host "WARNING: Could not kill hanging process" -ForegroundColor Yellow
        }
        return @{
            ExitCode = 99
            StdOut = "TIMEOUT AFTER $TimeoutSeconds SECONDS"
            StdErr = "Process did not complete in the allotted time"
        }
    }
    
    $stdout = $process.StandardOutput.ReadToEnd()
    $stderr = $process.StandardError.ReadToEnd()
    
    $result = @{
        ExitCode = $process.ExitCode
        StdOut = $stdout
        StdErr = $stderr
    }
    
    Write-Host "DEBUG: Exit code: $($result.ExitCode)" -ForegroundColor DarkGray
    if ($result.ExitCode -ne 0) {
        Write-Host "DEBUG: StdErr: $($result.StdErr)" -ForegroundColor Red
    }
    Write-Host "DEBUG: StdOut: $($result.StdOut)" -ForegroundColor DarkGray
    
    return $result
}

# Assert function to check conditions
function Assert {
    param (
        [Parameter(Mandatory=$true)]
        [bool]$Condition,
        
        [Parameter(Mandatory=$true)]
        [string]$Message
    )
    
    if (-not $Condition) {
        throw "Assertion failed: $Message"
    }
}

# Create a test profile
function Create-TestProfile {
    param (
        [Parameter(Mandatory=$true)]
        [string]$Name,
        
        [Parameter(Mandatory=$true)]
        [string]$Username,
        
        [Parameter(Mandatory=$true)]
        [string]$Email,
        
        [string]$Platform = "github",
        
        [string]$Token = "test_token",
        
        [switch]$SetupSSH,
        
        [switch]$Overwrite
    )
    
    $args = @(
        "add", $Name,
        "--username", $Username,
        "--email", $Email,
        "--platform", $Platform,
        "--token", $Token
    )
    
    if (-not $SetupSSH) {
        $args += "--setup-ssh=false"
    }
    
    if ($Overwrite) {
        $args += "--overwrite"
    }
    
    $result = Run-Gat -Arguments $args
    return $result
}

# Remove a test profile
function Remove-TestProfile {
    param (
        [Parameter(Mandatory=$true)]
        [string]$Name
    )
    
    $result = Run-Gat -Arguments @("remove", $Name, "--no-backup", "--force")
    return $result
}

# Switch to a profile
function Switch-TestProfile {
    param (
        [Parameter(Mandatory=$true)]
        [string]$Name,
        
        [switch]$SSH,
        
        [switch]$HTTPS,
        
        [switch]$DryRun
    )
    
    $args = @("switch", $Name)
    
    if ($SSH) {
        $args += "--ssh"
    }
    
    if ($HTTPS) {
        $args += "--https"
    }
    
    if ($DryRun) {
        $args += "--dry-run"
    }
    
    $result = Run-Gat -Arguments $args
    return $result
}

# Register a test platform
function Register-TestPlatform {
    param (
        [Parameter(Mandatory=$true)]
        [string]$ID,
        
        [Parameter(Mandatory=$true)]
        [string]$Name,
        
        [Parameter(Mandatory=$true)]
        [string]$HostName,
        
        [Parameter(Mandatory=$true)]
        [string]$SSHPrefix,
        
        [Parameter(Mandatory=$true)]
        [string]$HTTPSPrefix,
        
        [string]$SSHUser = "git",
        
        [string]$TokenScope = "",
        
        [switch]$Force
    )
    
    $args = @(
        "platforms", "register",
        "--id", $ID,
        "--name", $Name,
        "--host", $HostName,
        "--ssh-prefix", $SSHPrefix,
        "--https-prefix", $HTTPSPrefix,
        "--ssh-user", $SSHUser
    )
    
    if ($TokenScope) {
        $args += "--token-scope", $TokenScope
    }
    
    if ($Force) {
        $args += "--force"
    }
    
    $result = Run-Gat -Arguments $args
    return $result
}

# Create a test report
function Test-Report {
    param(
        [Parameter(Mandatory=$true)]
        [string]$TestName,
        
        [Parameter(Mandatory=$true)]
        [scriptblock]$TestBlock
    )
    
    $success = $false
    $message = ""
    
    try {
        & $TestBlock
        $success = $true
    } catch {
        $message = $_.Exception.Message
    }
    
    # Display results
    if ($success) {
        Write-Host "✅ $TestName - Passed" -ForegroundColor Green
    } else {
        Write-Host "❌ $TestName - Failed" -ForegroundColor Red
        Write-Host "   $message" -ForegroundColor Yellow
    }
    
    return $success
} 