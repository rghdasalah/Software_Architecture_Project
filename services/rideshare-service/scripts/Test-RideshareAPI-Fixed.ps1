# PowerShell script to run the Rideshare API test script on Windows
# Simple version with minimal formatting to avoid parsing issues

# Check if Git Bash is available
$gitBashPath = ""
if (Test-Path "C:\Program Files\Git\bin\bash.exe") {
    $gitBashPath = "C:\Program Files\Git\bin\bash.exe"
    Write-Host "Git Bash found at: $gitBashPath" -ForegroundColor Green
}
elseif (Test-Path "C:\Program Files (x86)\Git\bin\bash.exe") {
    $gitBashPath = "C:\Program Files (x86)\Git\bin\bash.exe"
    Write-Host "Git Bash found at: $gitBashPath" -ForegroundColor Green
}
else {
    Write-Host "Git Bash not found in standard locations" -ForegroundColor Yellow
}

# Check for Go
$goInstalled = $null -ne (Get-Command go -ErrorAction SilentlyContinue)
if ($goInstalled) {
    Write-Host "Go is installed" -ForegroundColor Green
} 
else {
    Write-Host "Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from https://golang.org/dl/"
    exit 1
}

# Function to run the test script using Git Bash
function RunBashScript {
    if ($gitBashPath) {
        Write-Host "Running test script using Git Bash..." -ForegroundColor Yellow
        & $gitBashPath -c "cd '$PSScriptRoot' && ./test-rideshare-api.sh"
        return $LASTEXITCODE
    } 
    else {
        Write-Host "Git Bash not found. Cannot run the bash script directly." -ForegroundColor Red
        return 1
    }
}

# Function to run the service directly with PowerShell
function RunServiceWithPowerShell {
    Write-Host "`nBuilding and Running Rideshare API with PowerShell" -ForegroundColor Cyan
    Write-Host "------------------------------------"
    
    # Build the service
    Write-Host "Building service..." -ForegroundColor Yellow
    Set-Location $PSScriptRoot
    & go build -o rideshare-api.exe ./cmd/server
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed" -ForegroundColor Red
        return 1
    }
    Write-Host "Build succeeded" -ForegroundColor Green
    
    # Check if the port is already in use
    $portInUse = $null -ne (Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue)
    if ($portInUse) {
        Write-Host "Port 8080 is already in use. You may need to stop existing services." -ForegroundColor Yellow
    }
    
    # Start the service
    Write-Host "Starting service (press Ctrl+C to stop)..." -ForegroundColor Yellow
    Write-Host "Manual testing required with tools like curl, Postman, or a web browser"
    Write-Host "API base URL: http://localhost:8080/api"
    Write-Host ""
    Write-Host "Example endpoints:"
    Write-Host "  - Health check:  GET http://localhost:8080/api/health"
    Write-Host "  - Register:      POST http://localhost:8080/api/users/register"
    Write-Host "  - Login:         POST http://localhost:8080/api/users/login"
    
    # Run the service in the foreground
    & .\rideshare-api.exe
    
    return 0
}

# Main script execution
Write-Host "`nRIDESHARE API TEST SCRIPT" -ForegroundColor Cyan
Write-Host "============================="

$option = ""
if (-not $gitBashPath) {
    Write-Host "Git Bash not found. Using PowerShell to run tests."
    $option = "2"
} 
else {
    Write-Host "Select an option:"
    Write-Host "1) Run full automated test suite using Git Bash"
    Write-Host "2) Build and run service for manual testing with PowerShell"
    $option = Read-Host "Option"
}

switch ($option) {
    "1" { 
        $result = RunBashScript 
        if ($result -eq 0) {
            Write-Host "All tests completed successfully!" -ForegroundColor Green
        } 
        else {
            Write-Host "Some tests failed. Check the output for details." -ForegroundColor Red
        }
    }
    "2" { 
        RunServiceWithPowerShell 
    }
    default { 
        Write-Host "Invalid option selected." -ForegroundColor Red
        exit 1
    }
}