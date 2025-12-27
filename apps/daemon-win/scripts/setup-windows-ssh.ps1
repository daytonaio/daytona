# Copyright 2025 Daytona Platforms Inc.
# SPDX-License-Identifier: AGPL-3.0

# Setup script to enable OpenSSH Server on Windows
# Run this script on the Windows VM with Administrator privileges:
#   PowerShell -ExecutionPolicy Bypass -File setup-windows-ssh.ps1

Write-Host "=== Setting up OpenSSH Server on Windows ===" -ForegroundColor Cyan

# Check if running as Administrator
$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Host "Error: This script must be run as Administrator" -ForegroundColor Red
    Write-Host "Right-click PowerShell and select 'Run as Administrator'" -ForegroundColor Yellow
    exit 1
}

# Install OpenSSH Server if not present
Write-Host "Checking for OpenSSH Server..." -ForegroundColor Yellow
$sshServer = Get-WindowsCapability -Online | Where-Object Name -like 'OpenSSH.Server*'

if ($sshServer.State -eq "NotPresent") {
    Write-Host "Installing OpenSSH Server..." -ForegroundColor Yellow
    Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
    Write-Host "OpenSSH Server installed." -ForegroundColor Green
} else {
    Write-Host "OpenSSH Server is already installed." -ForegroundColor Green
}

# Start and configure the SSH service
Write-Host "Configuring SSH service..." -ForegroundColor Yellow
Start-Service sshd
Set-Service -Name sshd -StartupType 'Automatic'

# Configure firewall rule for SSH
Write-Host "Configuring firewall..." -ForegroundColor Yellow
$existingRule = Get-NetFirewallRule -Name "OpenSSH-Server-In-TCP" -ErrorAction SilentlyContinue
if (-not $existingRule) {
    New-NetFirewallRule -Name "OpenSSH-Server-In-TCP" -DisplayName "OpenSSH Server (sshd)" `
        -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22
    Write-Host "Firewall rule created." -ForegroundColor Green
} else {
    Write-Host "Firewall rule already exists." -ForegroundColor Green
}

# Set PowerShell as default shell for SSH (optional but useful)
Write-Host "Setting PowerShell as default SSH shell..." -ForegroundColor Yellow
New-ItemProperty -Path "HKLM:\SOFTWARE\OpenSSH" -Name DefaultShell `
    -Value "C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe" `
    -PropertyType String -Force | Out-Null

# Create daytona user if it doesn't exist
$daytonaUser = Get-LocalUser -Name "daytona" -ErrorAction SilentlyContinue
if (-not $daytonaUser) {
    Write-Host "Creating 'daytona' user..." -ForegroundColor Yellow
    $password = Read-Host "Enter password for 'daytona' user" -AsSecureString
    New-LocalUser -Name "daytona" -Password $password -FullName "Daytona" -Description "Daytona development user"
    Add-LocalGroupMember -Group "Administrators" -Member "daytona"
    Write-Host "'daytona' user created and added to Administrators group." -ForegroundColor Green
} else {
    Write-Host "'daytona' user already exists." -ForegroundColor Green
}

# Create deployment directory
$deployPath = "C:\daytona"
if (-not (Test-Path $deployPath)) {
    Write-Host "Creating deployment directory: $deployPath" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $deployPath -Force | Out-Null
    Write-Host "Deployment directory created." -ForegroundColor Green
} else {
    Write-Host "Deployment directory already exists." -ForegroundColor Green
}

# Display connection info
Write-Host ""
Write-Host "=== Setup Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "SSH Server Status:" -ForegroundColor Cyan
Get-Service sshd | Format-Table Name, Status, StartType -AutoSize

$ipAddress = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.InterfaceAlias -notlike "*Loopback*" } | Select-Object -First 1).IPAddress
Write-Host "Windows IP Address: $ipAddress" -ForegroundColor Cyan
Write-Host ""
Write-Host "To connect from Linux devcontainer:" -ForegroundColor Yellow
Write-Host "  ssh -J h1001.blinkbox.dev daytona@$ipAddress" -ForegroundColor White
Write-Host ""
Write-Host "To deploy daemon-win.exe:" -ForegroundColor Yellow
Write-Host "  WIN_VM_IP=$ipAddress nx deploy daemon-win" -ForegroundColor White


