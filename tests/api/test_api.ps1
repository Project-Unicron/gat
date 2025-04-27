#!/usr/bin/env pwsh
# API tests for gat

# Import test utilities
. (Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) ".." "test_utils.ps1")

# Start the API server
$gatProcess = $null

function Start-ApiServer {
    $gatPath = Get-Command "gat" | Select-Object -ExpandProperty Source
    if (-not $gatPath) {
        $gatPath = Join-Path $env:WORKSPACE_ROOT "gat.exe"
    }
    
    # Start the gat serve process
    $gatProcess = Start-Process -FilePath $gatPath -ArgumentList "serve", "--port", "9998" -PassThru -NoNewWindow
    
    # Give it a moment to start up
    Start-Sleep -Seconds 1
    
    return $gatProcess
}

function Stop-ApiServer {
    param (
        [Parameter(Mandatory=$true)]
        [System.Diagnostics.Process]$Process
    )
    
    if (-not $Process.HasExited) {
        $Process.Kill()
        $Process.WaitForExit()
    }
}

function Test-ApiEndpoint {
    param (
        [Parameter(Mandatory=$true)]
        [string]$Endpoint,
        
        [Parameter(Mandatory=$true)]
        [string]$ExpectedPattern
    )
    
    $url = "http://localhost:9998$Endpoint"
    $response = Invoke-WebRequest -Uri $url -UseBasicParsing
    
    if ($response.StatusCode -ne 200) {
        return $false, "Status code was $($response.StatusCode), expected 200"
    }
    
    if ($response.Content -notmatch $ExpectedPattern) {
        return $false, "Response did not contain expected pattern '$ExpectedPattern'"
    }
    
    return $true, "Endpoint $Endpoint is working correctly"
}

# Run tests
Test-Report -TestName "API: Server starts" -TestBlock {
    try {
        $proc = Start-ApiServer
        Assert ($proc -ne $null) "Failed to start API server"
        Assert (-not $proc.HasExited) "API server process has exited prematurely"
    }
    finally {
        if ($proc) {
            Stop-ApiServer -Process $proc
        }
    }
}

Test-Report -TestName "API: Health check" -TestBlock {
    try {
        $proc = Start-ApiServer
        $result, $message = Test-ApiEndpoint -Endpoint "/ping" -ExpectedPattern "pong"
        Assert $result $message
    }
    finally {
        if ($proc) {
            Stop-ApiServer -Process $proc
        }
    }
}

Test-Report -TestName "API: Profiles endpoint" -TestBlock {
    try {
        $proc = Start-ApiServer
        $result, $message = Test-ApiEndpoint -Endpoint "/profiles" -ExpectedPattern '"profiles":'
        Assert $result $message
    }
    finally {
        if ($proc) {
            Stop-ApiServer -Process $proc
        }
    }
}

Test-Report -TestName "API: Platforms endpoint" -TestBlock {
    try {
        $proc = Start-ApiServer
        $result, $message = Test-ApiEndpoint -Endpoint "/platforms" -ExpectedPattern '"platforms":'
        Assert $result $message
    }
    finally {
        if ($proc) {
            Stop-ApiServer -Process $proc
        }
    }
}

Test-Report -TestName "API: GraphQL endpoint" -TestBlock {
    try {
        $proc = Start-ApiServer
        
        $query = '{"query":"{ platforms { id name } }"}'
        $headers = @{
            "Content-Type" = "application/json"
        }
        
        $url = "http://localhost:9998/graphql"
        $response = Invoke-WebRequest -Uri $url -Method POST -Body $query -Headers $headers -UseBasicParsing
        
        Assert ($response.StatusCode -eq 200) "GraphQL request failed with status code $($response.StatusCode)"
        Assert ($response.Content -match 'github') "GraphQL response did not contain expected platform 'github'"
    }
    finally {
        if ($proc) {
            Stop-ApiServer -Process $proc
        }
    }
} 