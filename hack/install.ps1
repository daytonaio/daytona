# PowerShell script to download and install Daytona binary

# Determine architecture
$architecture = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "arm64" }

# Define version and download URL
$version = if ($env:DAYTONA_SERVER_VERSION) { $env:DAYTONA_SERVER_VERSION } else { "latest" }
$baseUrl = if ($env:DAYTONA_SERVER_DOWNLOAD_URL) { $env:DAYTONA_SERVER_DOWNLOAD_URL } else { "https://download.daytona.io/daytona" }
$destination = if ($env:DAYTONA_PATH) { $env:DAYTONA_PATH } else { "$env:APPDATA\bin\daytona" }
$downloadUrl = "$baseUrl/$version/daytona-windows-$architecture.exe"

# Create destination directory if it doesn't exist
if (!(Test-Path -Path $destination)) {
    New-Item -ItemType Directory -Force -Path $destination | Out-Null
}

# File to download
$outputFile = "$destination\daytona.exe"

# Download the file with progress using Invoke-WebRequest
try {
    Write-Host "Downloading Daytona binary from $downloadUrl..." -ForegroundColor Yellow

    Invoke-WebRequest -Uri $downloadUrl -OutFile $outputFile -UseBasicParsing -ErrorAction Stop

    Write-Host "Download complete!" -ForegroundColor Green
} catch {
    Write-Error "Failed to download Daytona binary: $_"
    exit 1
}

# Set executable permissions
try {
    Set-ItemProperty -Path $outputFile -Name IsReadOnly -Value $false
    [System.IO.File]::SetAttributes($outputFile, 'Normal')
} catch {
    Write-Error "Failed to set executable permissions: $_"
    exit 1
}

# Add to PATH
$env:Path += ";$destination"
[System.Environment]::SetEnvironmentVariable("Path", $env:Path, [System.EnvironmentVariableTarget]::User)

# Confirm installation
if (Test-Path $outputFile) {
    Write-Host "Daytona successfully installed to $destination!" -ForegroundColor Green
    Write-Host "You can now use 'daytona' from the command line."
} else {
    Write-Error "Daytona installation failed."
    exit 1
}
