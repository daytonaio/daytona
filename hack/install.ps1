# PowerShell script to download and install Daytona binary

# Determine architecture
$architecture = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "arm64" }

# Define version and download URL
$version = if ($env:DAYTONA_SERVER_VERSION) { $env:DAYTONA_SERVER_VERSION } else { "latest" }
$baseUrl = if ($env:DAYTONA_SERVER_DOWNLOAD_URL) { $env:DAYTONA_SERVER_DOWNLOAD_URL } else { "https://download.daytona.io/daytona" }
$destination = if ($env:DAYTONA_PATH) { $env:DAYTONA_PATH } else { "$env:APPDATA\bin\daytona" }
$downloadUrl = "$baseUrl/$version/daytona-windows-$architecture.exe"

Write-Host "Installing Daytona..."

Write-Host ""  # Empty line

# Display installation directory
if ($env:DAYTONA_PATH) {
    Write-Host "Using custom installation directory: $destination"
}
else {
    Write-Host "Default installation directory: $destination"
    Write-Host "You can override this by setting the DAYTONA_PATH environment variable."
}
Write-Host ""  # Empty line
# Create destination directory if it doesn't exist
try {
    if (!(Test-Path -Path $destination)) {
        Write-Host "Creating installation directory at $destination"
        New-Item -ItemType Directory -Force -Path $destination -ErrorAction Stop | Out-Null
        Write-Host ""  # Empty line
    }
}
catch {
    Write-Error "Failed to create installation directory: $_"
    exit 1
}
# File to download
$outputFile = "$destination\daytona.exe"

# Download the file with progress using Invoke-WebRequest
try {
    Write-Host "Downloading Daytona binary from $downloadUrl"

    Invoke-WebRequest -Uri $downloadUrl -OutFile $outputFile -UseBasicParsing -ErrorAction Stop

    Write-Host ""  # Empty line
    Write-Host "Download complete!"
}
catch {
    Write-Error "Failed to download Daytona binary: $_"
    exit 1
}
Write-Host ""  # Empty line
# Set executable permissions
try {
    Write-Host "Setting executable permissions for Daytona binary..."
    Set-ItemProperty -Path $outputFile -Name IsReadOnly -Value $false
    [System.IO.File]::SetAttributes($outputFile, 'Normal')
}
catch {
    Write-Error "Failed to set executable permissions: $_"
    exit 1
}
Write-Host ""  # Empty line
# Add to PATH if not already present
try {
    if (-not ($env:Path -split ';' | ForEach-Object { $_.TrimEnd('\') } | Where-Object { $_ -eq $destination })) {
        Write-Host "Adding $destination to PATH..."
        [System.Environment]::SetEnvironmentVariable("Path", "$env:Path;$destination", [System.EnvironmentVariableTarget]::User)
        Write-Host "PATH updated successfully!"
    }
}
catch {
    Write-Error "Failed to update PATH: $_"
    exit 1
}

Write-Host ""  # Empty line

# Confirm installation
if (Test-Path $outputFile) {
    Write-Host "Daytona has been successfully installed to $destination"
    Write-Host "You can now use 'daytona' from the command line."
}
else {
    Write-Error "Daytona installation failed."
    exit 1
}
Write-Host ""  # Empty line
