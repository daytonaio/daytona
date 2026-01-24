# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Setup DNS for *.proxy.localhost -> 127.0.0.1 on Windows

Write-Host "Setting up DNS for *.proxy.localhost on Windows..." -ForegroundColor Green

# Check if running as Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "This script requires Administrator privileges. Please run as Administrator." -ForegroundColor Red
    exit 1
}

# Option 1: Use Acrylic DNS Proxy (recommended for wildcard support)
Write-Host "`nOption 1: Install Acrylic DNS Proxy for wildcard domain support" -ForegroundColor Cyan
Write-Host "This is the recommended approach for proper *.proxy.localhost resolution`n"

$installAcrylic = Read-Host "Do you want to install Acrylic DNS Proxy? (Y/N)"

if ($installAcrylic -eq 'Y' -or $installAcrylic -eq 'y') {
    # Check if Chocolatey is installed
    if (Get-Command choco -ErrorAction SilentlyContinue) {
        Write-Host "Installing Acrylic DNS Proxy via Chocolatey..." -ForegroundColor Yellow
        choco install acrylic-dns-proxy -y
        
        # Configure Acrylic
        $acrylicConfigPath = "C:\Program Files (x86)\Acrylic DNS Proxy\AcrylicHosts.txt"
        if (Test-Path $acrylicConfigPath) {
            Add-Content -Path $acrylicConfigPath -Value "`n# Daytona proxy domains"
            Add-Content -Path $acrylicConfigPath -Value "127.0.0.1 *.proxy.localhost"
            
            # Restart Acrylic service
            Restart-Service "Acrylic DNS Proxy" -ErrorAction SilentlyContinue
            
            Write-Host "`nAcrylic DNS Proxy configured successfully!" -ForegroundColor Green
            Write-Host "You may need to change your DNS settings to 127.0.0.1" -ForegroundColor Yellow
        }
    } else {
        Write-Host "Chocolatey is not installed. Please install from https://chocolatey.org/" -ForegroundColor Red
        Write-Host "Or download Acrylic DNS Proxy manually from https://mayakron.altervista.org/support/acrylic/Home.htm" -ForegroundColor Yellow
    }
} else {
    # Option 2: Manual hosts file (limited, no wildcards)
    Write-Host "`nOption 2: Add common proxy ports to hosts file (limited solution)" -ForegroundColor Cyan
    Write-Host "Note: This won't support all wildcard domains, only specific entries`n" -ForegroundColor Yellow
    
    $hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"
    
    # Common proxy ports that might be used
    $proxyEntries = @(
        "127.0.0.1 proxy.localhost",
        "127.0.0.1 3000.proxy.localhost",
        "127.0.0.1 4000.proxy.localhost",
        "127.0.0.1 8080.proxy.localhost",
        "127.0.0.1 8000.proxy.localhost"
    )
    
    Write-Host "Adding entries to hosts file..." -ForegroundColor Yellow
    $hostsContent = Get-Content $hostsPath
    
    foreach ($entry in $proxyEntries) {
        if ($hostsContent -notcontains $entry) {
            Add-Content -Path $hostsPath -Value $entry
            Write-Host "Added: $entry" -ForegroundColor Green
        }
    }
    
    Write-Host "`nHosts file updated. Note: This only covers specific ports, not all *.proxy.localhost domains" -ForegroundColor Yellow
}

Write-Host "`n=== DNS Setup Complete ===" -ForegroundColor Green
Write-Host "Test with: nslookup 2280-test.proxy.localhost" -ForegroundColor Cyan
Write-Host "Or: ping 3000.proxy.localhost" -ForegroundColor Cyan
