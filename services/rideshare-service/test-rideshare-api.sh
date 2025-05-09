#!/bin/bash
# test-rideshare-api.sh - Comprehensive test script for the Rideshare API Service
# Tests database connectivity and key API endpoints

set -e  # Exit on error

# Text formatting
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Configuration
API_HOST="localhost"
API_PORT="8080"
API_BASE="http://$API_HOST:$API_PORT/api"
TIMEOUT=5  # Seconds to wait for service to start

# Log messages
log_success() {
  echo -e "${GREEN}✓ $1${NC}"
}

log_error() {
  echo -e "${RED}✗ $1${NC}"
  exit 1
}

log_info() {
  echo -e "${YELLOW}$1${NC}"
}

log_header() {
  echo -e "\n${BOLD}$1${NC}"
  echo "----------------------------------------"
}

# Make sure required utilities are available
check_requirements() {
  log_header "Checking Requirements"
  
  for cmd in go curl jq; do
    if ! command -v $cmd &> /dev/null; then
      log_error "$cmd is required but not installed"
    else
      log_success "$cmd is installed"
    fi
  done
}

# Build the service
build_service() {
  log_header "Building Rideshare Service"
  
  # Get dependencies
  log_info "Downloading dependencies..."
  go mod download || log_error "Failed to download dependencies"
  log_success "Dependencies downloaded"
  
  # Build the service
  log_info "Building service..."
  go build -o rideshare-api ./cmd/server || log_error "Build failed"
  log_success "Service built successfully"
}

# Start the service in background
check_existing_process() {
  ps -W | grep -i "rideshare-api" > /dev/null
  return $?
}

start_service() {
  log_header "Starting Rideshare Service"
  
  # Check if service is already running using our function
  if check_existing_process; then
    log_info "Service is already running, stopping it first"
    taskkill /F /IM rideshare-api.exe 2>/dev/null || log_info "No existing service found"
  fi
  
  # Start the service
  log_info "Starting service..."
  ./rideshare-api > rideshare-api.log 2>&1 &
  PID=$!
  
  # Check if process is running
  if ! ps | grep $PID > /dev/null; then
    log_error "Failed to start service"
  fi
  
  # Store PID for later cleanup
  echo $PID > .api.pid
  log_success "Service started with PID $PID"

  # Wait for service to be ready
  log_info "Waiting for service to be ready (max $TIMEOUT seconds)..."
  for i in $(seq 1 $TIMEOUT); do
    if curl -s "$API_BASE/health" > /dev/null; then
      log_success "Service is ready!"
      return 0
    fi
    sleep 1
  done
  
  log_error "Service failed to become ready in $TIMEOUT seconds"
}

stop_service() {
  log_header "Stopping Rideshare Service"
  
  if [ -f .api.pid ]; then
    PID=$(cat .api.pid)
    if ps | grep $PID > /dev/null; then
      log_info "Stopping service with PID $PID..."
      kill $PID
      sleep 1
      log_success "Service stopped"
    else
      log_info "Service with PID $PID is not running"
    fi
    rm .api.pid
  else
    log_info "No PID file found, service may not be running"
    # Try to kill any running instances
    taskkill /F /IM rideshare-api.exe 2>/dev/null || true
  fi
}
# Test health endpoint
test_health() {
  log_header "Testing Health Endpoint"
  
  response=$(curl -s "$API_BASE/health")
  if [ $? -ne 0 ]; then
    log_error "Failed to connect to health endpoint"
  fi
  
  # Check if database field exists in response
  if echo $response | jq -e '.database' > /dev/null; then
    db_status=$(echo $response | jq -r '.database')
    if [ "$db_status" == "ok" ]; then
      log_success "Database connection is healthy"
    else
      log_error "Database connection is not healthy: $db_status"
    fi
  else
    log_error "Health response doesn't contain database status"
  fi
  
  log_success "Health endpoint test passed"
}

# Register a test user
register_user() {
  log_header "Testing User Registration"
  
  # Generate random email to avoid conflicts
  random_suffix=$RANDOM
  test_email="test_user_$random_suffix@example.com"
  test_password="Password123!"
  
  # Register user
  log_info "Registering test user with email: $test_email"
  response=$(curl -s -X POST "$API_BASE/users/register" \
    -H "Content-Type: application/json" \
    -d '{
      "email": "'$test_email'",
      "password": "'$test_password'",
      "firstName": "Test",
      "lastName": "User",
      "phoneNumber": "1234567890",
      "dateOfBirth": "1990-01-01"
    }')
  
  # Check response
  if [ $? -ne 0 ]; then
    log_error "Failed to register user"
  fi
  
  # Extract user ID and token
  if echo $response | jq -e '.userId' > /dev/null; then
    user_id=$(echo $response | jq -r '.userId')
    token=$(echo $response | jq -r '.token')
    
    # Store for later use
    echo $user_id > .test_user_id
    echo $token > .test_token
    
    log_success "User registered successfully with ID: $user_id"
    return 0
  else
    error_msg=$(echo $response | jq -r '.error' 2>/dev/null || echo "Unknown error")
    log_error "Failed to register user: $error_msg"
  fi
}

# Login with the test user
login_user() {
  log_header "Testing User Login"
  
  # Get the test email
  if [ ! -f .test_user_id ]; then
    log_error "No test user found. Run register_user first."
  fi
  
  random_suffix=$RANDOM
  test_email="test_user_$random_suffix@example.com"
  test_password="Password123!"
  
  # Re-register to ensure we have a fresh user
  register_user
  
  # Login
  log_info "Logging in as $test_email"
  response=$(curl -s -X POST "$API_BASE/users/login" \
    -H "Content-Type: application/json" \
    -d '{
      "email": "'$test_email'",
      "password": "'$test_password'"
    }')
  
  # Check response
  if [ $? -ne 0 ]; then
    log_error "Failed to login"
  fi
  
  # Extract token
  if echo $response | jq -e '.token' > /dev/null; then
    token=$(echo $response | jq -r '.token')
    
    # Store token for later use
    echo $token > .test_token
    
    log_success "Login successful"
    return 0
  else
    error_msg=$(echo $response | jq -r '.error' 2>/dev/null || echo "Unknown error")
    log_error "Failed to login: $error_msg"
  fi
}

# Create a vehicle for the test user
create_vehicle() {
  log_header "Testing Vehicle Creation"
  
  # Check if we have a token
  if [ ! -f .test_token ]; then
    log_error "No authentication token found. Run login_user first."
  fi
  
  token=$(cat .test_token)
  
  # Create vehicle
  log_info "Creating a test vehicle"
  raw_response=$(curl -s -X POST "$API_BASE/vehicles" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
    -d '{
      "make": "Tesla",
      "model": "Model 3",
      "year": 2023,
      "color": "Blue",
      "licensePlate": "TEST123",
      "capacity": 4
    }')
  
  # Debug: Print raw response
  echo -e "${YELLOW}DEBUG: Raw vehicle creation response: '$raw_response'${NC}"
  
  # Check response
  if [ $? -ne 0 ]; then
    log_error "Failed to create vehicle"
  fi
  
  # Extract vehicle ID
  if echo $raw_response | jq -e '.vehicleId' > /dev/null; then
    vehicle_id=$(echo $raw_response | jq -r '.vehicleId')
    
    # Store for later use
    echo $vehicle_id > .test_vehicle_id
    
    log_success "Vehicle created successfully with ID: $vehicle_id"
    return 0
  else
    error_msg=$(echo $raw_response | jq -r '.error' 2>/dev/null || echo "Unknown error")
    log_error "Failed to create vehicle: $error_msg"
  fi
}

# Get vehicles for the test user
get_vehicles() {
  log_header "Testing Vehicle Retrieval"
  
  # Check if we have a token
  if [ ! -f .test_token ]; then
    log_error "No authentication token found. Run login_user first."
  fi
  
  token=$(cat .test_token)
  
  # Get vehicles
  log_info "Fetching user vehicles"
  response=$(curl -s -X GET "$API_BASE/vehicles" \
    -H "Authorization: Bearer $token")
  
  # Check response
  if [ $? -ne 0 ]; then
    log_error "Failed to get vehicles"
  fi
  
  # Check if we got an array
  if echo $response | jq -e 'length' > /dev/null; then
    vehicle_count=$(echo $response | jq 'length')
    log_success "Successfully retrieved $vehicle_count vehicles"
    return 0
  else
    error_msg=$(echo $response | jq -r '.error' 2>/dev/null || echo "Unknown error")
    log_error "Failed to get vehicles: $error_msg"
  fi
}

# Create a ride
create_ride() {
  log_header "Testing Ride Creation"
  
  # Check if we have token and vehicle
  if [ ! -f .test_token ] || [ ! -f .test_vehicle_id ]; then
    log_error "Missing token or vehicle ID. Run create_vehicle first."
  fi
  
  token=$(cat .test_token)
  vehicle_id=$(cat .test_vehicle_id)
  
  # Create ride
  log_info "Creating a test ride"
  
  # Calculate departure time (1 hour from now)
  departure_time=$(date -u -d "+1 hour" "+%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -v+1H -u "+%Y-%m-%dT%H:%M:%SZ")
  
  # Calculate arrival time (2 hours from now)
  arrival_time=$(date -u -d "+2 hours" "+%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -v+2H -u "+%Y-%m-%dT%H:%M:%SZ")
  
  response=$(curl -s -X POST "$API_BASE/rides" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
    -d '{
      "vehicleId": "'$vehicle_id'",
      "originAddress": "123 Main St",
      "originLatitude": 37.7749,
      "originLongitude": -122.4194,
      "destinationAddress": "456 Market St",
      "destinationLatitude": 37.7922,
      "destinationLongitude": -122.3984,
      "departureTime": "'$departure_time'",
      "estimatedArrivalTime": "'$arrival_time'",
      "maxPassengers": 3,
      "availableSeats": 3,
      "pricePerSeat": 15.50,
      "description": "Test ride",
      "isPetsAllowed": false,
      "isSmokingAllowed": false
    }')
  
  # Check response
  if [ $? -ne 0 ]; then
    log_error "Failed to create ride"
  fi
  
  # Extract ride ID
  if echo $response | jq -e '.id' > /dev/null; then
    ride_id=$(echo $response | jq -r '.id')
    
    # Store for later use
    echo $ride_id > .test_ride_id
    
    log_success "Ride created successfully with ID: $ride_id"
    return 0
  else
    error_msg=$(echo $response | jq -r '.error' 2>/dev/null || echo "Unknown error")
    log_error "Failed to create ride: $error_msg"
  fi
}

# Get a specific ride
get_ride() {
  log_header "Testing Ride Retrieval"
  
  # Check if we have a ride ID
  if [ ! -f .test_ride_id ]; then
    log_error "No test ride found. Run create_ride first."
  fi
  
  ride_id=$(cat .test_ride_id)
  
  # Get ride
  log_info "Fetching ride with ID: $ride_id"
  response=$(curl -s -X GET "$API_BASE/rides/$ride_id")
  
  # Check response
  if [ $? -ne 0 ]; then
    log_error "Failed to get ride"
  fi
  
  # Check if we got the right ride
  if echo $response | jq -e '.id' > /dev/null; then
    retrieved_id=$(echo $response | jq -r '.id')
    if [ "$retrieved_id" == "$ride_id" ]; then
      log_success "Successfully retrieved ride with ID: $ride_id"
      return 0
    else
      log_error "Retrieved wrong ride ID: expected $ride_id, got $retrieved_id"
    fi
  else
    error_msg=$(echo $response | jq -r '.error' 2>/dev/null || echo "Unknown error")
    log_error "Failed to get ride: $error_msg"
  fi
}

# Search for nearby rides
search_nearby_rides() {
  log_header "Testing Nearby Rides Search"
  
  # Search for rides
  log_info "Searching for nearby rides"
  response=$(curl -s -X GET "$API_BASE/rides/nearby?lat=37.7749&lon=-122.4194&destLat=37.7922&destLon=-122.3984")
  
  # Check response
  if [ $? -ne 0 ]; then
    log_error "Failed to search for nearby rides"
  fi
  
  # Check if we got an array
  if echo $response | jq -e 'length >= 0' > /dev/null; then
    ride_count=$(echo $response | jq 'length')
    log_success "Successfully retrieved $ride_count nearby rides"
    return 0
  else
    error_msg=$(echo $response | jq -r '.error' 2>/dev/null || echo "Unknown error")
    log_error "Failed to search for nearby rides: $error_msg"
  fi
}

# Clean up test data
cleanup() {
  log_header "Cleaning Up"
  
  # Remove temporary files
  for file in .test_user_id .test_token .test_vehicle_id .test_ride_id .api.pid; do
    if [ -f $file ]; then
      rm $file
      log_success "Removed $file"
    fi
  done
}

# Run all tests
run_tests() {
  log_header "Running All Tests"
  
  test_health
  register_user
  login_user
  create_vehicle
  get_vehicles
  create_ride
  get_ride
  search_nearby_rides
  
  log_success "All tests completed successfully!"
}

# Main function
main() {
  echo -e "${BOLD}RIDESHARE API TEST SCRIPT${NC}"
  echo "============================="
  
  # Ensure cleanup happens on exit
  trap stop_service EXIT
  
  check_requirements
  build_service
  start_service
  run_tests
  
  cleanup
}

# Execute main function
main