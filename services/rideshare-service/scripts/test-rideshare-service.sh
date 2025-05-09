#!/bin/bash
# test-rideshare-service.sh
# A comprehensive test script for the rideshare service API

# Configuration
BASE_URL="http://localhost:8080"
TEST_USER_ID=""
TEST_RIDE_ID=""
TEST_TOKEN=""
VEHICLE_ID=""

# Color output helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

function print_success() {
    echo -e "${GREEN}$1${NC}"
}

function print_error() {
    echo -e "${RED}ERROR: $1${NC}"
}

function print_info() {
    echo -e "${CYAN}$1${NC}"
}

function print_separator() {
    echo -e "${GRAY}------------------------------------------------------${NC}"
}

# Test runner
function test_endpoint() {
    local name=$1
    local endpoint=$2
    local method=${3:-"GET"}
    local body=$4
    local headers=$5
    
    echo -n "Testing: $name... "
    
    # Prepare request
    local curl_cmd="curl -s -X $method"
    
    # Add headers
    if [ -n "$headers" ]; then
        for header in $headers; do
            curl_cmd="$curl_cmd -H '$header'"
        done
    fi
    
    # Add content type header for POST/PUT requests
    if [ "$method" = "POST" ] || [ "$method" = "PUT" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json'"
    fi
    
    # Add body if present
    if [ -n "$body" ]; then
        curl_cmd="$curl_cmd -d '$body'"
    fi
    
    # Add URL
    curl_cmd="$curl_cmd ${BASE_URL}${endpoint}"
    
    # Execute the request and capture ONLY the response (no other output)
    local response=$(eval $curl_cmd)
    local exit_code=$?
    
    # Check for error in execution
    if [ $exit_code -ne 0 ]; then
        print_error "Request failed with exit code $exit_code"
        return 1
    fi
    
    # Check for error in response
    if echo "$response" | grep -q "error"; then
        print_error "API returned error: $response"
        return 1
    fi
    
    # Success
    print_success "PASSED"
    echo "$response"
    # Return the raw response for extraction
    echo "$response"
}

# Begin testing
print_info "STARTING RIDESHARE SERVICE API TESTS"
print_info "$(date)"
print_separator

# 1. Test database connections
print_info "TESTING DATABASE CONNECTIONS"

# Check primary database
if docker-compose exec -T db-primary psql -U postgres -c "\l" 2>&1 | grep -q "rideshare"; then
    print_success "Primary database connection: SUCCESS"
else
    print_error "Primary database connection: FAILED"
fi

# Check replica database
if docker-compose exec -T db-replica1 psql -U postgres -c "\l" 2>&1 | grep -q "rideshare"; then
    print_success "Replica database connection: SUCCESS"
else
    print_error "Replica database connection: FAILED"
fi
print_separator

# 2. Test Health Endpoint
print_info "TESTING BASIC ENDPOINTS"
test_endpoint "Health Check" "/api/health"

# 3. User Registration and Authentication
print_info "TESTING USER MANAGEMENT"
email="test.user$RANDOM@example.com"
password="Password123!"

# Register a user
response=$(test_endpoint "User Registration" "/api/users/register" "POST" \
    "{\"email\":\"$email\",\"password\":\"$password\",\"firstName\":\"Test\",\"lastName\":\"User\",\"phoneNumber\":\"555-123-4567\",\"dateOfBirth\":\"1990-01-01\"}")

# Extract user ID
TEST_USER_ID=$(echo $response | grep -o '"userId":"[^"]*' | sed 's/"userId":"//')

if [ -n "$TEST_USER_ID" ]; then
    print_success "Got User ID: $TEST_USER_ID"
else
    print_error "Failed to get User ID"
fi

# Login
response=$(test_endpoint "User Login" "/api/users/login" "POST" \
    "{\"email\":\"$email\",\"password\":\"$password\"}")

# Extract token
TEST_TOKEN=$(echo $response | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -n "$TEST_TOKEN" ]; then
    print_success "Got Auth Token: $TEST_TOKEN"
else
    print_error "Failed to get Auth Token"
fi

# 4. Vehicle Management
print_info "TESTING VEHICLE MANAGEMENT"
if [ -n "$TEST_TOKEN" ]; then
    # Add a vehicle
    echo "Sending vehicle creation request..."
    
    # Make a direct curl call to avoid output mixing
    raw_response=$(curl -s -X POST \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TEST_TOKEN" \
      -d "{\"make\":\"Toyota\",\"model\":\"Camry\",\"year\":2022,\"color\":\"Blue\",\"licensePlate\":\"TEST-$RANDOM\",\"capacity\":4}" \
      ${BASE_URL}/api/vehicles)
    
    echo "Raw API response: $raw_response"
    
    # Extract vehicle ID
    VEHICLE_ID=$(echo "$raw_response" | grep -o '"vehicleId":"[^"]*' | sed 's/"vehicleId":"//')
    
    if [ -n "$VEHICLE_ID" ]; then
        print_success "Got Vehicle ID: $VEHICLE_ID"
    else
        print_error "Failed to get Vehicle ID"
    fi
    
    # Now test getting user vehicles
    test_endpoint "Get User Vehicles" "/api/vehicles" "GET" "" "Authorization: Bearer $TEST_TOKEN"
else
    print_error "Skipping vehicle tests: No authentication token available"
fi

# 5. Ride Management
print_info "TESTING RIDE MANAGEMENT"

if [ -n "$TEST_TOKEN" ] && [ -n "$VEHICLE_ID" ]; then
    # Create a ride
    future_date=$(date -d "+1 day" -Iseconds)
    future_date_plus_hour=$(date -d "+1 day +1 hour" -Iseconds)
    
    response=$(test_endpoint "Create Ride" "/api/rides" "POST" \
        "{\"hostId\":\"$TEST_USER_ID\",\"vehicleId\":\"$VEHICLE_ID\",\"originAddress\":\"Central Park, New York\",\"originLatitude\":40.7812,\"originLongitude\":-73.9665,\"destinationAddress\":\"Times Square, New York\",\"destinationLatitude\":40.7580,\"destinationLongitude\":-73.9855,\"departureTime\":\"$future_date\",\"estimatedArrivalTime\":\"$future_date_plus_hour\",\"maxPassengers\":3,\"availableSeats\":3,\"pricePerSeat\":15.00,\"description\":\"Test ride\"}" \
        "Authorization: Bearer $TEST_TOKEN")
    
    # Extract ride ID
    TEST_RIDE_ID=$(echo $response | grep -o '"rideId":"[^"]*' | sed 's/"rideId":"//')
    
    if [ -n "$TEST_RIDE_ID" ]; then
        print_success "Got Ride ID: $TEST_RIDE_ID"
    else
        print_error "Failed to get Ride ID"
    fi
    
    # Find nearby rides
    test_endpoint "Find Nearby Rides" "/api/rides/nearby?lat=40.7812&lon=-73.9665&destLat=40.7580&destLon=-73.9855&radius=10000"
    
    # Get ride by ID
    if [ -n "$TEST_RIDE_ID" ]; then
        test_endpoint "Get Ride by ID" "/api/rides/$TEST_RIDE_ID"
    fi
else
    print_error "Skipping ride tests: No authentication token or vehicle ID available"
fi

# 6. Ride Request Management
if [ -n "$TEST_TOKEN" ] && [ -n "$TEST_RIDE_ID" ]; then
    print_info "TESTING RIDE REQUESTS"
    
    # Create a secondary test user for requesting rides
    email2="rider.test$RANDOM@example.com"
    password2="Password123!"
    
    # Register rider
    response=$(test_endpoint "Register Rider" "/api/users/register" "POST" \
        "{\"email\":\"$email2\",\"password\":\"$password2\",\"firstName\":\"Test\",\"lastName\":\"Rider\",\"phoneNumber\":\"555-987-6543\",\"dateOfBirth\":\"1992-03-15\"}")
    
    # Extract second user ID
    TEST_USER_ID2=$(echo $response | grep -o '"userId":"[^"]*' | sed 's/"userId":"//')
    
    if [ -n "$TEST_USER_ID2" ]; then
        print_success "Got Rider ID: $TEST_USER_ID2"
    else
        print_error "Failed to get Rider ID"
    fi
    
    # Login as rider
    response=$(test_endpoint "Login Rider" "/api/users/login" "POST" \
        "{\"email\":\"$email2\",\"password\":\"$password2\"}")
    
    # Extract second token
    TEST_TOKEN2=$(echo $response | grep -o '"token":"[^"]*' | sed 's/"token":"//')
    
    if [ -n "$TEST_TOKEN2" ]; then
        print_success "Got Rider Auth Token: $TEST_TOKEN2"
        
        # Request a ride
        response=$(test_endpoint "Request Ride" "/api/rides/$TEST_RIDE_ID/requests" "POST" \
            "{\"riderId\":\"$TEST_USER_ID2\",\"pickupAddress\":\"Columbia University, New York\",\"pickupLatitude\":40.8075,\"pickupLongitude\":-73.9626,\"seatsRequested\":1,\"message\":\"Test ride request\"}" \
            "Authorization: Bearer $TEST_TOKEN2")
        
        # Extract request ID
        REQUEST_ID=$(echo $response | grep -o '"requestId":"[^"]*' | sed 's/"requestId":"//')
        
        if [ -n "$REQUEST_ID" ]; then
            print_success "Got Request ID: $REQUEST_ID"
            
            # Accept ride request (as driver)
            test_endpoint "Accept Ride Request" "/api/rides/$TEST_RIDE_ID/requests/$REQUEST_ID" "PUT" \
                "{\"status\":\"accepted\"}" \
                "Authorization: Bearer $TEST_TOKEN"
        else
            print_error "Failed to get Request ID"
        fi
    else
        print_error "Failed to get Rider Auth Token"
    fi
else
    print_error "Skipping ride request tests: No authentication token or ride ID available"
fi

# 7. Test cleanup - optional
if [ -n "$TEST_TOKEN" ] && [ -n "$TEST_RIDE_ID" ]; then
    print_info "CLEANING UP TEST DATA"
    
    test_endpoint "Delete Test Ride" "/api/rides/$TEST_RIDE_ID" "DELETE" \
        "" "Authorization: Bearer $TEST_TOKEN"
fi

print_separator
print_info "TESTS COMPLETED: $(date)"