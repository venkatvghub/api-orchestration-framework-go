#!/bin/bash

# Mobile Onboarding BFF - Comprehensive Test Script
# This script tests all endpoints in the mobile-bff example

set -e

# Configuration
BASE_URL="http://localhost:8080"
USER_ID="test_$(date +%s)"
DEVICE_TYPE="ios"
APP_VERSION="1.0.0"
PLATFORM="mobile"
OS_VERSION="17.0"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_test() {
    echo -e "${YELLOW}Test: $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Test function with error handling
run_test() {
    local test_name="$1"
    local curl_command="$2"
    
    print_test "$test_name"
    echo "Command: $curl_command"
    echo "Response:"
    
    if eval "$curl_command"; then
        print_success "$test_name completed"
    else
        print_error "$test_name failed"
    fi
    echo ""
}

print_header "Mobile Onboarding BFF Test Suite"
echo "Base URL: $BASE_URL"
echo "Test User ID: $USER_ID"
echo "Device Type: $DEVICE_TYPE"
echo ""

# 1. HEALTH AND SYSTEM CHECKS
print_header "1. Health and System Checks"

run_test "Health Check" \
"curl -s -X GET '$BASE_URL/health' \
  -H 'Content-Type: application/json' | jq ."

run_test "Metrics Endpoint" \
"curl -s -X GET '$BASE_URL/metrics' \
  -H 'Content-Type: application/json'"

# 2. DEMO ENDPOINTS
print_header "2. Demo Endpoints"

run_test "Demo Overview" \
"curl -s -X GET '$BASE_URL/demo/' \
  -H 'Content-Type: application/json' | jq ."

run_test "Demo Test for Welcome Screen" \
"curl -s -X GET '$BASE_URL/demo/test/welcome' \
  -H 'Content-Type: application/json' | jq ."

run_test "Debug Context" \
"curl -s -X GET '$BASE_URL/demo/debug/context' \
  -H 'Content-Type: application/json' | jq ."

# 3. ONBOARDING SCREEN ENDPOINTS
print_header "3. Onboarding Screen Endpoints"

# Get Welcome Screen
run_test "Get Welcome Screen" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/screens/welcome?user_id=$USER_ID' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' \
  -H 'X-Platform: $PLATFORM' \
  -H 'X-OS-Version: $OS_VERSION' | jq ."

# Submit Welcome Screen
run_test "Submit Welcome Screen" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/screens/welcome/submit' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' \
  -H 'X-Platform: $PLATFORM' \
  -H 'X-OS-Version: $OS_VERSION' \
  -d '{
    \"user_id\": \"$USER_ID\",
    \"data\": {
      \"accepted_terms\": true,
      \"marketing_consent\": false
    }
  }' | jq ."

# Get Personal Info Screen
run_test "Get Personal Info Screen" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/screens/personal_info?user_id=$USER_ID' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' | jq ."

# Submit Personal Info Screen
run_test "Submit Personal Info Screen" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/screens/personal_info/submit' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' \
  -d '{
    \"user_id\": \"$USER_ID\",
    \"data\": {
      \"first_name\": \"John\",
      \"last_name\": \"Doe\",
      \"email\": \"john.doe@example.com\",
      \"phone\": \"+1234567890\"
    }
  }' | jq ."

# Get Preferences Screen
run_test "Get Preferences Screen" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/screens/preferences?user_id=$USER_ID' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' | jq ."

# Submit Preferences Screen
run_test "Submit Preferences Screen" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/screens/preferences/submit' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' \
  -d '{
    \"user_id\": \"$USER_ID\",
    \"data\": {
      \"notifications\": true,
      \"theme\": \"dark\",
      \"language\": \"en\"
    }
  }' | jq ."

# Get Verification Screen
run_test "Get Verification Screen" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/screens/verification?user_id=$USER_ID' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' | jq ."

# Submit Verification Screen
run_test "Submit Verification Screen" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/screens/verification/submit' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' \
  -d '{
    \"user_id\": \"$USER_ID\",
    \"data\": {
      \"verification_code\": \"123456\",
      \"email\": \"john.doe@example.com\"
    }
  }' | jq ."

# Get Completion Screen
run_test "Get Completion Screen" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/screens/completion?user_id=$USER_ID' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' | jq ."

# Submit Completion Screen
run_test "Submit Completion Screen" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/screens/completion/submit' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -H 'X-App-Version: $APP_VERSION' \
  -d '{
    \"user_id\": \"$USER_ID\",
    \"data\": {
      \"feedback\": \"Great experience!\",
      \"rating\": 5,
      \"newsletter_signup\": true
    }
  }' | jq ."

# 4. FLOW MANAGEMENT ENDPOINTS
print_header "4. Flow Management Endpoints"

# Get Onboarding Flow
run_test "Get Onboarding Flow" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/flow/$USER_ID' \
  -H 'Content-Type: application/json' | jq ."

# Complete Onboarding Flow
run_test "Complete Onboarding Flow" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/flow/$USER_ID/complete' \
  -H 'Content-Type: application/json' \
  -d '{
    \"completion_data\": {
      \"feedback\": \"Excellent onboarding experience!\",
      \"rating\": 5,
      \"suggestions\": \"Keep it simple and intuitive\"
    }
  }' | jq ."

# 5. PROGRESS TRACKING
print_header "5. Progress Tracking"

# Get User Progress
run_test "Get User Progress" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/progress/$USER_ID' \
  -H 'Content-Type: application/json' | jq ."

# 6. ANALYTICS ENDPOINTS
print_header "6. Analytics Endpoints"

# Track Screen View Event
run_test "Track Screen View Event" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/analytics/events' \
  -H 'Content-Type: application/json' \
  -d '{
    \"event_type\": \"screen_view\",
    \"user_id\": \"$USER_ID\",
    \"screen_id\": \"welcome\",
    \"properties\": {
      \"device_type\": \"$DEVICE_TYPE\",
      \"session_id\": \"sess_123\",
      \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"
    }
  }' | jq ."

# Track Button Click Event
run_test "Track Button Click Event" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/analytics/events' \
  -H 'Content-Type: application/json' \
  -d '{
    \"event_type\": \"button_click\",
    \"user_id\": \"$USER_ID\",
    \"screen_id\": \"welcome\",
    \"properties\": {
      \"button_id\": \"continue\",
      \"device_type\": \"$DEVICE_TYPE\",
      \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"
    }
  }' | jq ."

# Track Form Submission Event
run_test "Track Form Submission Event" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/analytics/events' \
  -H 'Content-Type: application/json' \
  -d '{
    \"event_type\": \"form_submit\",
    \"user_id\": \"$USER_ID\",
    \"screen_id\": \"personal_info\",
    \"properties\": {
      \"form_fields\": [\"email\", \"first_name\", \"last_name\"],
      \"device_type\": \"$DEVICE_TYPE\",
      \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"
    }
  }' | jq ."

# Get User Analytics
run_test "Get User Analytics (7 days)" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/analytics/$USER_ID?time_range=7d' \
  -H 'Content-Type: application/json' | jq ."

run_test "Get User Analytics (30 days)" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/analytics/$USER_ID?time_range=30d' \
  -H 'Content-Type: application/json' | jq ."

# 7. ERROR TESTING
print_header "7. Error Testing"

# Test Invalid Screen ID
run_test "Test Invalid Screen ID" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/screens/invalid_screen?user_id=$USER_ID' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' | jq ."

# Test Missing User ID
run_test "Test Missing User ID" \
"curl -s -X GET '$BASE_URL/api/v1/onboarding/screens/welcome' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' | jq ."

# Test Invalid JSON in Submission
run_test "Test Invalid JSON Submission" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/screens/welcome/submit' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -d '{\"invalid\": json}' | jq ."

# Test Missing Required Fields
run_test "Test Missing Required Fields in Personal Info" \
"curl -s -X POST '$BASE_URL/api/v1/onboarding/screens/personal_info/submit' \
  -H 'Content-Type: application/json' \
  -H 'X-Device-Type: $DEVICE_TYPE' \
  -d '{
    \"user_id\": \"$USER_ID\",
    \"data\": {
      \"first_name\": \"John\"
    }
  }' | jq ."

# 8. LOAD TESTING EXAMPLES
print_header "8. Load Testing Examples"

echo -e "${YELLOW}Load testing examples (commented out by default):${NC}"
echo ""
echo "# Test Welcome Screen Endpoint (100 requests, 10 concurrent)"
echo "# hey -n 100 -c 10 -H \"Content-Type: application/json\" -H \"X-Device-Type: ios\" \\"
echo "#   \"$BASE_URL/api/v1/onboarding/screens/welcome?user_id=load_test\""
echo ""
echo "# Test Submission Endpoint (50 requests, 5 concurrent)"
echo "# hey -n 50 -c 5 -H \"Content-Type: application/json\" -H \"X-Device-Type: ios\" \\"
echo "#   -d '{\"user_id\":\"load_test\",\"data\":{\"accepted_terms\":true}}' \\"
echo "#   -m POST \"$BASE_URL/api/v1/onboarding/screens/welcome/submit\""

# 9. WIREMOCK DIRECT TESTING
print_header "9. WireMock Direct Testing"

echo -e "${YELLOW}Testing WireMock endpoints directly:${NC}"
echo ""

# Test WireMock Health
run_test "WireMock Health Check" \
"curl -s -X GET 'http://localhost:8082/__admin/health' | jq ."

# Test WireMock Mappings
run_test "WireMock Mappings" \
"curl -s -X GET 'http://localhost:8082/__admin/mappings' | jq ."

# Direct WireMock Screen Test
run_test "Direct WireMock Screen Test" \
"curl -s -H 'Content-Type: application/json' -H 'X-API-Version: 2.0' \
  'http://localhost:8082/api/screens/welcome?user_id=$USER_ID&device_type=$DEVICE_TYPE' | jq ."

# Direct WireMock Submission Test
run_test "Direct WireMock Submission Test" \
"curl -s -X POST -H 'Content-Type: application/json' -H 'X-API-Version: 2.0' \
  'http://localhost:8082/api/screens/welcome/submit' \
  -d '{\"user_id\":\"$USER_ID\",\"data\":{\"accepted_terms\":true}}' | jq ."

# 10. SUMMARY
print_header "10. Test Summary"

echo -e "${GREEN}✓ All tests completed for user: $USER_ID${NC}"
echo -e "${BLUE}Base URL: $BASE_URL${NC}"
echo -e "${BLUE}Device Type: $DEVICE_TYPE${NC}"
echo -e "${BLUE}App Version: $APP_VERSION${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Check application logs for detailed request/response information"
echo "2. Monitor metrics at: $BASE_URL/metrics"
echo "3. View WireMock admin interface at: http://localhost:8082/__admin"
echo "4. Reset WireMock state: curl -X POST http://localhost:8082/__admin/reset"
echo ""
echo -e "${GREEN}Test script completed successfully!${NC}" 