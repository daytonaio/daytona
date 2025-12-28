# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Install daemon-win as a Windows service using NSSM
# Run this script as Administrator

param(
    [string]$InstallPath = "C:\daytona",
    [string]$ServiceName = "DaytonaDaemon",
    [string]$NssmUrl = "https://nssm.cc/release/nssm-2.24.zip"
)

Write-Host "=== Daytona Windows Daemon Service Installer ===" -ForegroundColor Cyan
Write-Host ""

# Check for admin privileges
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "ERROR: This script must be run as Administrator" -ForegroundColor Red
    exit 1
}

# Paths
$DaemonExe = Join-Path $InstallPath "daemon-win.exe"
$NssmPath = Join-Path $InstallPath "nssm.exe"
$LogPath = Join-Path $InstallPath "logs"

# Check if daemon exists
if (-not (Test-Path $DaemonExe)) {
    Write-Host "ERROR: Daemon not found at $DaemonExe" -ForegroundColor Red
    Write-Host "Please deploy the daemon first using 'nx deploy daemon-win'" -ForegroundColor Yellow
    exit 1
}

# Create logs directory
if (-not (Test-Path $LogPath)) {
    New-Item -ItemType Directory -Path $LogPath -Force | Out-Null
    Write-Host "Created logs directory: $LogPath" -ForegroundColor Green
}

# Download NSSM if not present
if (-not (Test-Path $NssmPath)) {
    Write-Host "Downloading NSSM..." -ForegroundColor Yellow
    $TempZip = Join-Path $env:TEMP "nssm.zip"
    $TempDir = Join-Path $env:TEMP "nssm-extract"
    
    try {
        # Download
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        Invoke-WebRequest -Uri $NssmUrl -OutFile $TempZip -UseBasicParsing
        
        # Extract
        if (Test-Path $TempDir) { Remove-Item -Recurse -Force $TempDir }
        Expand-Archive -Path $TempZip -DestinationPath $TempDir -Force
        
        # Find and copy the 64-bit nssm.exe
        $NssmSource = Get-ChildItem -Path $TempDir -Recurse -Filter "nssm.exe" | 
                      Where-Object { $_.DirectoryName -like "*win64*" } | 
                      Select-Object -First 1
        
        if ($NssmSource) {
            Copy-Item -Path $NssmSource.FullName -Destination $NssmPath -Force
            Write-Host "NSSM installed to: $NssmPath" -ForegroundColor Green
        } else {
            throw "Could not find nssm.exe in downloaded archive"
        }
        
        # Cleanup
        Remove-Item -Path $TempZip -Force -ErrorAction SilentlyContinue
        Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
    catch {
        Write-Host "ERROR: Failed to download/install NSSM: $_" -ForegroundColor Red
        Write-Host "You can manually download NSSM from https://nssm.cc/download" -ForegroundColor Yellow
        exit 1
    }
}

# Stop and remove existing service if present
$existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($existingService) {
    Write-Host "Stopping existing service..." -ForegroundColor Yellow
    & $NssmPath stop $ServiceName 2>$null | Out-Null
    Start-Sleep -Seconds 2
    Write-Host "Removing existing service..." -ForegroundColor Yellow
    & $NssmPath remove $ServiceName confirm 2>$null | Out-Null
    Start-Sleep -Seconds 1
}

Write-Host ""
Write-Host "Installing service..." -ForegroundColor Cyan

# Install the service
& $NssmPath install $ServiceName $DaemonExe
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to install service" -ForegroundColor Red
    exit 1
}

# Configure service settings
Write-Host "Configuring service..." -ForegroundColor Yellow

# Set display name and description
& $NssmPath set $ServiceName DisplayName "Daytona Daemon"
& $NssmPath set $ServiceName Description "Daytona Windows Daemon - Provides toolbox API for file system, process, and git operations"

# Set working directory
& $NssmPath set $ServiceName AppDirectory $InstallPath

# Configure logging
$StdoutLog = Join-Path $LogPath "daemon-stdout.log"
$StderrLog = Join-Path $LogPath "daemon-stderr.log"
& $NssmPath set $ServiceName AppStdout $StdoutLog
& $NssmPath set $ServiceName AppStderr $StderrLog
& $NssmPath set $ServiceName AppStdoutCreationDisposition 4  # Append
& $NssmPath set $ServiceName AppStderrCreationDisposition 4  # Append
& $NssmPath set $ServiceName AppRotateFiles 1
& $NssmPath set $ServiceName AppRotateOnline 1
& $NssmPath set $ServiceName AppRotateBytes 10485760  # 10MB

# Configure restart on failure
& $NssmPath set $ServiceName AppExit Default Restart
& $NssmPath set $ServiceName AppRestartDelay 5000  # 5 seconds delay before restart

# Set startup type to automatic
& $NssmPath set $ServiceName Start SERVICE_AUTO_START

# Set service to run as LocalSystem (or configure specific user if needed)
& $NssmPath set $ServiceName ObjectName LocalSystem

Write-Host ""
Write-Host "Starting service..." -ForegroundColor Cyan

# Start the service
& $NssmPath start $ServiceName
Start-Sleep -Seconds 2

# Check service status
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($service -and $service.Status -eq "Running") {
    Write-Host ""
    Write-Host "=== Installation Complete ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "Service Name:    $ServiceName" -ForegroundColor White
    Write-Host "Status:          Running" -ForegroundColor Green
    Write-Host "Startup Type:    Automatic" -ForegroundColor White
    Write-Host "Executable:      $DaemonExe" -ForegroundColor White
    Write-Host "Logs:            $LogPath" -ForegroundColor White
    Write-Host ""
    Write-Host "The daemon will automatically:" -ForegroundColor Cyan
    Write-Host "  - Start on system boot" -ForegroundColor White
    Write-Host "  - Restart if it crashes (after 5 second delay)" -ForegroundColor White
    Write-Host ""
    Write-Host "Useful commands:" -ForegroundColor Yellow
    Write-Host "  Check status:  Get-Service $ServiceName" -ForegroundColor Gray
    Write-Host "  Stop service:  Stop-Service $ServiceName" -ForegroundColor Gray
    Write-Host "  Start service: Start-Service $ServiceName" -ForegroundColor Gray
    Write-Host "  View logs:     Get-Content $StdoutLog -Tail 50" -ForegroundColor Gray
    Write-Host "  Uninstall:     .\uninstall-service.ps1" -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "WARNING: Service may not have started correctly" -ForegroundColor Yellow
    Write-Host "Check logs at: $LogPath" -ForegroundColor Yellow
    Write-Host "Or run: Get-Service $ServiceName" -ForegroundColor Yellow
}

