# Load .env variables manually
$envFile = "c:\Users\hjsgo\Projects\daytona\scripts\.env"
$apiKey = ""
$orgId = ""

Get-Content $envFile | ForEach-Object {
    if ($_ -match "DAYTONA_API_KEY=(.*)") { $apiKey = $matches[1] }
    if ($_ -match "DAYTONA_ORG_ID=(.*)") { $orgId = $matches[1] }
}

Write-Host "Creating sandbox via Direct API call..."
Write-Host "API Key: $($apiKey.Substring(0, 10))..."
Write-Host "Org ID: $orgId"

$headers = @{
    "Authorization"     = "Bearer $apiKey"
    "X-Organization-Id" = $orgId
    "Content-Type"      = "application/json"
}

$body = @{
    name = "debug-sandbox-$(Get-Date -Format 'HHmmss')"
    snapshot = "daytonaio/sandbox:0.5.0-slim"
} | ConvertTo-Json

try {
    $response = Invoke-WebRequest `
        -Uri "http://localhost:3000/api/sandbox" `
        -Method Post `
        -Headers $headers `
        -Body $body `
        -UseBasicParsing
    
    Write-Host "`n✅ Success!" -ForegroundColor Green
    Write-Host "Status: $($response.StatusCode)"
    Write-Host "Content: $($response.Content)"
}
catch {
    Write-Host "`n❌ Error Details:" -ForegroundColor Red
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $errorBody = $reader.ReadToEnd()
        Write-Host "Status: $($_.Exception.Response.StatusCode)"
        Write-Host "Body: $errorBody"
    } else {
        Write-Host "Message: $($_.Exception.Message)"
    }
}
