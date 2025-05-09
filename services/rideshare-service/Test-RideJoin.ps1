#!/usr/bin/env pwsh
# Test script for ride join API endpoints

# Base URL for the API
$baseUrl = "http://localhost:8080/api"
$token = ""

# Step 1: Create a test rider user
Write-Host "Step 1: Creating test rider user..." -ForegroundColor Cyan
$riderData = @{
    email = "rider@example.com"
    password = "password123"
    firstName = "Test"
    lastName = "Rider"
    phoneNumber = "+1234567890"
    dateOfBirth = "1990-01-01"
}
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method Post -Body ($riderData | ConvertTo-Json) -ContentType "application/json" -ErrorAction Stop
    $riderId = $response.user.id
    Write-Host "Created rider with ID: $riderId" -ForegroundColor Green
} 
catch {
    Write-Host "Error creating rider:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 2: Login as the rider to get token
Write-Host "Step 2: Logging in as the rider..." -ForegroundColor Cyan
$loginData = @{
    email = $riderData.email
    password = $riderData.password
}
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/login" -Method Post -Body ($loginData | ConvertTo-Json) -ContentType "application/json" -ErrorAction Stop
    $riderToken = $response.token
    Write-Host "Got rider token: $($riderToken.Substring(0, 20))..." -ForegroundColor Green
}
catch {
    Write-Host "Error logging in as rider:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 3: Create a test driver user
Write-Host "Step 3: Creating test driver user..." -ForegroundColor Cyan
$driverData = @{
    email = "driver@example.com"
    password = "password123"
    firstName = "Test"
    lastName = "Driver"
    phoneNumber = "+1987654321"
    dateOfBirth = "1985-01-01"
}
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/register" -Method Post -Body ($driverData | ConvertTo-Json) -ContentType "application/json" -ErrorAction Stop
    $driverId = $response.user.id
    Write-Host "Created driver with ID: $driverId" -ForegroundColor Green
} 
catch {
    Write-Host "Error creating driver:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 4: Login as the driver to get token
Write-Host "Step 4: Logging in as the driver..." -ForegroundColor Cyan
$loginData = @{
    email = $driverData.email
    password = $driverData.password
}
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/login" -Method Post -Body ($loginData | ConvertTo-Json) -ContentType "application/json" -ErrorAction Stop
    $driverToken = $response.token
    Write-Host "Got driver token: $($driverToken.Substring(0, 20))..." -ForegroundColor Green
}
catch {
    Write-Host "Error logging in as driver:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 5: Create a vehicle for the driver
Write-Host "Step 5: Creating a vehicle for the driver..." -ForegroundColor Cyan
$headers = @{ Authorization = "Bearer $driverToken" }
$vehicleData = @{
    make = "Toyota"
    model = "Camry"
    year = 2022
    color = "Blue"
    licensePlate = "TEST123"
    capacity = 4
}
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/vehicles" -Method Post -Headers $headers -Body ($vehicleData | ConvertTo-Json) -ContentType "application/json" -ErrorAction Stop
    $vehicleId = $response.vehicleId
    Write-Host "Created vehicle with ID: $vehicleId" -ForegroundColor Green
}
catch {
    Write-Host "Error creating vehicle:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 6: Create a ride with the driver
Write-Host "Step 6: Creating a ride as the driver..." -ForegroundColor Cyan
$departureTime = [DateTime]::Now.AddHours(2).ToString("yyyy-MM-ddTHH:mm:sszzz")
$arrivalTime = [DateTime]::Now.AddHours(3).ToString("yyyy-MM-ddTHH:mm:sszzz")

$rideData = @{
    vehicleId = $vehicleId
    originAddress = "123 Start St, City"
    originLatitude = 34.052235
    originLongitude = -118.243683
    destinationAddress = "456 End Ave, City"
    destinationLatitude = 34.052235
    destinationLongitude = -118.343683
    departureTime = $departureTime
    estimatedArrivalTime = $arrivalTime
    maxPassengers = 3
    availableSeats = 3
    pricePerSeat = 15.00
    description = "Test ride"
    isPetsAllowed = $true
    isSmokingAllowed = $false
}
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/rides" -Method Post -Headers $headers -Body ($rideData | ConvertTo-Json) -ContentType "application/json" -ErrorAction Stop
    $rideId = $response.id
    Write-Host "Created ride with ID: $rideId" -ForegroundColor Green
} 
catch {
    Write-Host "Error creating ride:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    Write-Host $_.Exception.Response.Content
    exit
}

# Step 7: Request to join the ride as the rider
Write-Host "Step 7: Requesting to join the ride as the rider..." -ForegroundColor Cyan
$headers = @{ Authorization = "Bearer $riderToken" }
$joinRequestData = @{
    pickupAddress = "123 Pickup St, City"
    pickupLatitude = 34.052235
    pickupLongitude = -118.253683
    seatsRequested = 2
    message = "Can I join this ride?"
}
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/rides/$rideId/join" -Method Post -Headers $headers -Body ($joinRequestData | ConvertTo-Json) -ContentType "application/json" -ErrorAction Stop
    $requestId = $response.requestId
    Write-Host "Created join request with ID: $requestId" -ForegroundColor Green
}
catch {
    Write-Host "Error requesting to join ride:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 8: View rider's requests
Write-Host "Step 8: Viewing rider's requests..." -ForegroundColor Cyan
$headers = @{ Authorization = "Bearer $riderToken" }
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/requests" -Method Get -Headers $headers -ErrorAction Stop
    Write-Host "Rider's requests:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3 | Write-Host
}
catch {
    Write-Host "Error getting rider's requests:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 9: Driver views join requests for their ride
Write-Host "Step 9: Driver viewing join requests for their ride..." -ForegroundColor Cyan
$headers = @{ Authorization = "Bearer $driverToken" }
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/rides/$rideId/requests" -Method Get -Headers $headers -ErrorAction Stop
    Write-Host "Ride join requests:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3 | Write-Host
}
catch {
    Write-Host "Error getting join requests:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 10: Driver accepts the join request
Write-Host "Step 10: Driver accepting the join request..." -ForegroundColor Cyan
$acceptData = @{
    status = "accepted"
}
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/rides/$rideId/requests/$requestId" -Method Put -Headers $headers -Body ($acceptData | ConvertTo-Json) -ContentType "application/json" -ErrorAction Stop
    Write-Host "Join request accepted successfully" -ForegroundColor Green
}
catch {
    Write-Host "Error accepting join request:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 11: View ride passengers
Write-Host "Step 11: Viewing ride passengers..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/rides/$rideId/passengers" -Method Get -Headers $headers -ErrorAction Stop
    Write-Host "Ride passengers:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3 | Write-Host
}
catch {
    Write-Host "Error getting ride passengers:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

# Step 12: Verify available seats were updated
Write-Host "Step 12: Verifying available seats were updated..." -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/rides/$rideId" -Method Get -Headers $headers -ErrorAction Stop
    Write-Host "Ride details after join:" -ForegroundColor Green
    Write-Host "Available seats: $($response.availableSeats) (should be 1)"
    if ($response.availableSeats -eq 1) {
        Write-Host "✓ Seats updated correctly" -ForegroundColor Green
    }
    else {
        Write-Host "✗ Seats not updated correctly" -ForegroundColor Red
    }
}
catch {
    Write-Host "Error getting updated ride:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    exit
}

Write-Host "All tests completed!" -ForegroundColor Green