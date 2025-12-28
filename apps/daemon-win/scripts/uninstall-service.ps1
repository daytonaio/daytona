# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Uninstall daemon-win Windows service
# Run this script as Administrator

param(
    [string]$InstallPath = "C:\daytona",
    [string]$ServiceName = "DaytonaDaemon"
)

$ErrorActionPreference = "Stop"

Write-Host "=== Daytona Windows Daemon Service Uninstaller ===" -ForegroundColor Cyan
Write-Host ""

# Check for admin privileges
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "ERROR: This script must be run as Administrator" -ForegroundColor Red
    exit 1
}

$NssmPath = Join-Path $InstallPath "nssm.exe"

# Check if service exists
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if (-not $service) {
    Write-Host "Service '$ServiceName' is not installed" -ForegroundColor Yellow
    exit 0
}

Write-Host "Stopping service..." -ForegroundColor Yellow
if (Test-Path $NssmPath) {
    & $NssmPath stop $ServiceName 2>$null
} else {
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
}
Start-Sleep -Seconds 2

Write-Host "Removing service..." -ForegroundColor Yellow
if (Test-Path $NssmPath) {
    & $NssmPath remove $ServiceName confirm
} else {
    sc.exe delete $ServiceName
}
Start-Sleep -Seconds 1

# Verify removal
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if (-not $service) {
    Write-Host ""
    Write-Host "=== Service Uninstalled Successfully ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "Note: The daemon executable and logs remain at: $InstallPath" -ForegroundColor Yellow
    Write-Host "To completely remove, delete: $InstallPath" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "WARNING: Service may still exist. Try rebooting." -ForegroundColor Yellow
}

