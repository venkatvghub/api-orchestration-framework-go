# Mobile Onboarding BFF  - WireMock Integration

A comprehensive Backend for Frontend (BFF) implementation for mobile app onboarding using WireMock for local mock API development. This example demonstrates how to build a production-ready BFF layer that orchestrates multiple backend services through WireMock.

## 🚀 Features

- **Multi-screen onboarding flow** with 5 distinct screens
- **WireMock mock API integration** for realistic backend simulation
- **Real-time API orchestration** using the API Orchestration Framework
- **Mobile-optimized responses** with field selection and data transformation
- **Comprehensive error handling** with fallback mechanisms
- **Request/response validation** with detailed schemas
- **Performance monitoring** with metrics and logging
- **Complete test suite** with curl examples

## 📱 Onboarding Screens

1. **Welcome Screen** - App introduction and getting started
2. **Personal Info** - User details collection (name, email, phone)
3. **Preferences** - Theme, language, and notification settings
4. **Verification** - Email/phone verification process
5. **Completion** - Success confirmation and next steps

## 🔗 WireMock Mock APIs

This example uses WireMock for local mock API development with the following endpoints:

- **Base URL**: `http://localhost:8080` (configurable via `MOCK_API_BASE_URL` env var)
- **WireMock Admin**: [http://localhost:8082/__admin](http://localhost:8082/__admin)
- **Alternative**: Use Beeceptor cloud at `https://mobile-bff.free.beeceptor.com`

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/screens/{screenId}` | Get screen configuration |
| `POST` | `/api/screens/{screenId}/submit` | Submit screen data |
| `GET` | `/api/users/{userId}/progress` | Get onboarding progress |
| `PUT` | `/api/users/{userId}/progress` | Update progress |
| `GET` | `/api/screens/{screenId}/next` | Get next screen |
| `GET` | `/api/users/{userId}/flow` | Get user flow |
| `POST` | `/api/users/{userId}/complete` | Complete onboarding |
| `GET` | `/api/analytics/events` | Get analytics data |
| `POST` | `/api/analytics/events` | Track single event |
| `POST` | `/api/analytics/events/batch` | Track multiple events |
| `GET` | `/api/analytics/users/{userId}` | Get user analytics |

## 🎯 WireMock API Specification

### Required Headers for All Endpoints
```
Content-Type: application/json
X-API-Version: 2.0
```

### 1. GET `{baseURL}/api/screens/{screenId}`

**Query Parameters:**
- `user_id` (required): User identifier
- `device_type` (optional): Device type (ios, android, web)

**Example Request:**
```
GET http://localhost:8080/api/screens/welcome?user_id=user123&device_type=ios
Content-Type: application/json
X-API-Version: 2.0
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "title": "Welcome to Our App!",
    "description": "Let's get you started with a quick setup",
    "type": "welcome"
  }
}
```

### 2. POST `{baseURL}/api/screens/{screenId}/submit`

**Example Request:**
```
POST http://localhost:8080/api/screens/welcome/submit
Content-Type: application/json
X-API-Version: 2.0
```

**Request Body:**
```json
{
  "user_id": "user123",
  "screen_id": "welcome",
  "data": {
    "accepted_terms": true,
    "marketing_consent": false
  },
  "timestamp": "2024-01-15T10:31:00Z",
  "device_info": {
    "type": "ios",
    "platform": "mobile",
    "app_version": "1.0.0",
    "os_version": "17.0"
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "next_screen": "personal_info",
    "message": "Welcome! Let's continue with your personal information."
  }
}
```

### 3. GET `{baseURL}/api/users/{userId}/progress`

**Example Request:**
```
GET http://localhost:8080/api/users/user123/progress
Content-Type: application/json
X-API-Version: 2.0
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "current_screen": "preferences",
    "completed_steps": 2,
    "total_steps": 5,
    "percent_complete": 0.4
  }
}
```

### 4. PUT `{baseURL}/api/users/{userId}/progress`

**Example Request:**
```
PUT http://localhost:8080/api/users/user123/progress
Content-Type: application/json
X-API-Version: 2.0
```

**Request Body:**
```json
{
  "user_id": "user123",
  "screen_id": "personal_info",
  "data": {
    "completed": true
  },
  "timestamp": "2024-01-15T10:35:00Z"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "updated": true
  }
}
```

### 5. GET `{baseURL}/api/screens/{screenId}/next`

**Query Parameters:**
- `user_id` (required): User identifier

**Example Request:**
```
GET http://localhost:8080/api/screens/personal_info/next?user_id=user123
Content-Type: application/json
X-API-Version: 2.0
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "next_screen": "preferences",
    "progress": 0.4
  }
}
```

### 6. GET `{baseURL}/api/users/{userId}/flow`

**Example Request:**
```
GET http://localhost:8080/api/users/user123/flow
Content-Type: application/json
X-API-Version: 2.0
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "user_id": "user123",
    "flow_id": "mobile_onboarding_v2",
    "current_screen": "preferences",
    "completed_screens": ["welcome", "personal_info"],
    "progress": 0.4,
    "status": "in_progress",
    "started_at": "2024-01-15T10:30:00Z",
    "last_activity": "2024-01-15T10:35:00Z",
    "total_screens": 5
  }
}
```

### 7. POST `{baseURL}/api/users/{userId}/complete`

**Example Request:**
```
POST http://localhost:8080/api/users/user123/complete
Content-Type: application/json
X-API-Version: 2.0
```

**Request Body:**
```json
{
  "completion_data": {
    "feedback": "Great experience!",
    "rating": 5
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "status": "completed",
    "message": "Onboarding completed successfully!",
    "completion_time": "2024-01-15T10:39:00Z",
    "total_duration": "9 minutes"
  }
}
```

### 8. GET `{baseURL}/api/analytics/events`

**Query Parameters:**
- `user_id` (required): User identifier
- `time_range` (optional): Time range (7d, 30d, etc.)

**Example Request:**
```
GET http://localhost:8080/api/analytics/events?user_id=user123&time_range=7d
Content-Type: application/json
X-API-Version: 2.0
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "events": [
      {
        "event_type": "screen_view",
        "user_id": "user123",
        "screen_id": "welcome",
        "timestamp": "2024-01-15T10:30:00Z"
      }
    ],
    "total_events": 15,
    "time_range": "7d"
  }
}
```

### 9. POST `{baseURL}/api/analytics/events`

**Example Request:**
```
POST http://localhost:8080/api/analytics/events
Content-Type: application/json
X-API-Version: 2.0
```

**Request Body:**
```json
{
  "event_type": "screen_view",
  "user_id": "user123",
  "screen_id": "welcome",
  "properties": {
    "device_type": "ios",
    "timestamp": "2024-01-15T10:30:00Z"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Response (200 OK or 201 Created):**
```json
{
  "success": true,
  "data": {
    "tracked": true,
    "event_id": "evt_123"
  }
}
```

### 10. POST `{baseURL}/api/analytics/events/batch`

**Example Request:**
```
POST http://localhost:8080/api/analytics/events/batch
Content-Type: application/json
X-API-Version: 2.0
```

**Request Body:**
```json
{
  "events": [
    {
      "event_type": "screen_view",
      "user_id": "user123",
      "screen_id": "welcome",
      "properties": {
        "device_type": "ios"
      },
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ],
  "timestamp": "2024-01-15T10:30:00Z",
  "count": 1
}
```

**Response (200 OK or 201 Created):**
```json
{
  "success": true,
  "data": {
    "processed": 1,
    "batch_id": "batch_123"
  }
}
```

### 11. GET `{baseURL}/api/analytics/users/{userId}`

**Example Request:**
```
GET http://localhost:8080/api/analytics/users/user123
Content-Type: application/json
X-API-Version: 2.0
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "total_events": 25,
    "screen_views": 15,
    "submissions": 8,
    "time_spent_seconds": 540
  }
}
```

### Error Response Format

All endpoints return errors in this format:

**Response (4xx/5xx):**
```json
{
  "success": false,
  "error": "Error message description"
}
```

### Screen-Specific Data Examples

#### Welcome Screen (`welcome`)
```json
{
  "accepted_terms": true,
  "marketing_consent": false
}
```

#### Personal Info Screen (`personal_info`)
```json
{
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
```

#### Preferences Screen (`preferences`)
```json
{
  "notifications": true,
  "theme": "dark",
  "language": "en"
}
```

#### Verification Screen (`verification`)
```json
{
  "verification_code": "123456",
  "email": "user@example.com"
}
```

#### Completion Screen (`completion`)
```json
{
  "feedback": "Great experience!",
  "rating": 5,
  "newsletter_signup": true
}
```

## 🐳 WireMock Setup

### Using Docker Compose (Recommended)

1. **Start WireMock**
   ```bash
   docker-compose up -d wiremock
   ```

2. **Verify WireMock is running**
   ```bash
   curl http://localhost:8080/__admin/health
   ```

3. **View WireMock mappings**
   ```bash
   curl http://localhost:8080/__admin/mappings
   ```

### Manual WireMock Setup

1. **Download WireMock**
   ```bash
   wget https://repo1.maven.org/maven2/com/github/tomakehurst/wiremock-jre8-standalone/2.35.0/wiremock-jre8-standalone-2.35.0.jar
   ```

2. **Start WireMock**
   ```bash
   java -jar wiremock-jre8-standalone-2.35.0.jar --port 8082 --root-dir ./wiremock
   ```

### WireMock Admin Interface

- **Admin UI**: http://localhost:8080/__admin
- **Health Check**: http://localhost:8080/__admin/health
- **View Mappings**: http://localhost:8080/__admin/mappings
- **Reset Mappings**: `POST http://localhost:8080/__admin/reset`

### Environment Configuration

```bash
# Use local WireMock (default)
export MOCK_API_BASE_URL=http://localhost:8080

# Use Beeceptor cloud (alternative)
export MOCK_API_BASE_URL=https://mobile-bff.free.beeceptor.com
```

## 🛠 Setup Instructions

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Git
- curl (for testing)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd examples/mobile-bff
   ```

2. **Start WireMock**
   ```bash
   docker-compose up -d wiremock
   ```

3. **Verify WireMock is running**
   ```bash
   curl http://localhost:8080/__admin/health
   ```

4. **Install dependencies**
   ```bash
   go mod tidy
   ```

5. **Run the BFF server**
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080`

## 🧪 Testing the BFF Layer

### 1. Health Check

First, verify the BFF is running:

```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "2.0.0",
  "service": "mobile-onboarding-bff-v2",
  "mock_api": "integrated"
}
```

### 2. Demo Endpoints

Check available endpoints:

```bash
curl http://localhost:8080/demo/
```

**Expected Response:**
```json
{
  "message": "Mobile Onboarding BFF v2 Demo",
  "endpoints": [
    "GET /api/v1/onboarding/screens/{screenId}?user_id={userId}",
    "POST /api/v1/onboarding/screens/{screenId}/submit",
    "GET /api/v1/onboarding/flow/{userId}",
    "POST /api/v1/onboarding/flow/{userId}/complete",
    "GET /api/v1/onboarding/progress/{userId}",
    "GET /api/v1/onboarding/analytics/{userId}",
    "POST /api/v1/onboarding/analytics/events"
  ],
  "screens": ["welcome", "personal_info", "preferences", "verification", "completion"],
  "mock_api_url": "http://localhost:8080"
}
```

### 3. Complete Onboarding Flow Test

#### Step 1: Get Welcome Screen

```bash
curl -X GET "http://localhost:8080/api/v1/onboarding/screens/welcome?user_id=test123" \
  -H "Content-Type: application/json" \
  -H "X-Device-Type: ios" \
  -H "X-App-Version: 1.0.0"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "title": "Welcome to Our App!",
    "description": "Let's get you started with a quick setup",
    "type": "welcome"
  },
  "metadata": {
    "version": "2.0.0",
    "process_time": "45.123ms",
    "source": "mock_api"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Step 2: Submit Welcome Screen

```bash
curl -X POST "http://localhost:8080/api/v1/onboarding/screens/welcome/submit" \
  -H "Content-Type: application/json" \
  -H "X-Device-Type: ios" \
  -H "X-App-Version: 1.0.0" \
  -d '{
    "user_id": "test123",
    "data": {
      "accepted_terms": true,
      "marketing_consent": false
    }
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "next_screen": "personal_info",
    "message": "Welcome! Let's continue with your personal information.",
    "timestamp": "2024-01-15T10:31:00Z"
  },
  "metadata": {
    "version": "2.0.0",
    "process_time": "67.456ms",
    "source": "mock_api"
  },
  "timestamp": "2024-01-15T10:31:00Z"
}
```

#### Step 3: Get Personal Info Screen

```bash
curl -X GET "http://localhost:8080/api/v1/onboarding/screens/personal_info?user_id=test123" \
  -H "Content-Type: application/json" \
  -H "X-Device-Type: ios"
```

#### Step 4: Submit Personal Info

```bash
curl -X POST "http://localhost:8080/api/v1/onboarding/screens/personal_info/submit" \
  -H "Content-Type: application/json" \
  -H "X-Device-Type: ios" \
  -d '{
    "user_id": "test123",
    "data": {
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@example.com",
      "phone": "+1234567890"
    }
  }'
```

#### Step 5: Get User Progress

```bash
curl -X GET "http://localhost:8080/api/v1/onboarding/progress/test123" \
  -H "Content-Type: application/json"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "current_screen": "preferences",
    "completed_steps": 2,
    "total_steps": 5,
    "percent_complete": 0.4
  },
  "timestamp": "2024-01-15T10:35:00Z"
}
```

#### Step 6: Get Onboarding Flow

```bash
curl -X GET "http://localhost:8080/api/v1/onboarding/flow/test123" \
  -H "Content-Type: application/json"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "user_id": "test123",
    "flow_id": "mobile_onboarding_v2",
    "current_screen": "preferences",
    "completed_screens": ["welcome", "personal_info"],
    "progress": 0.4,
    "status": "in_progress",
    "started_at": "2024-01-15T10:30:00Z",
    "last_activity": "2024-01-15T10:35:00Z",
    "total_screens": 5
  },
  "timestamp": "2024-01-15T10:35:00Z"
}
```

### 4. Analytics Testing

#### Track an Event

```bash
curl -X POST "http://localhost:8080/api/v1/onboarding/analytics/events" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "screen_view",
    "user_id": "test123",
    "screen_id": "welcome",
    "properties": {
      "device_type": "ios",
      "session_id": "sess_123"
    }
  }'
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "tracked": true,
    "event_id": "evt_ABC123"
  },
  "timestamp": "2024-01-15T10:36:00Z"
}
```

#### Get User Analytics

```bash
curl -X GET "http://localhost:8080/api/v1/onboarding/analytics/test123?time_range=7d" \
  -H "Content-Type: application/json"
```

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "total_events": 25,
    "screen_views": 15,
    "submissions": 8,
    "time_spent_seconds": 540
  },
  "timestamp": "2024-01-15T10:37:00Z"
}
```

### 5. Error Testing

#### Test Invalid Screen ID

```bash
curl -X GET "http://localhost:8080/api/v1/onboarding/screens/invalid?user_id=test123" \
  -H "Content-Type: application/json" \
  -H "X-Device-Type: ios"
```

#### Test Missing User ID

```bash
curl -X GET "http://localhost:8080/api/v1/onboarding/screens/welcome" \
  -H "Content-Type: application/json" \
  -H "X-Device-Type: ios"
```

#### Test Invalid JSON

```bash
curl -X POST "http://localhost:8080/api/v1/onboarding/screens/welcome/submit" \
  -H "Content-Type: application/json" \
  -H "X-Device-Type: ios" \
  -d '{"invalid": json}'
```

### 6. Load Testing

#### Install hey (if not already installed)

```bash
go install github.com/rakyll/hey@latest
```

#### Test Welcome Screen Endpoint

```bash
hey -n 100 -c 10 -H "Content-Type: application/json" -H "X-Device-Type: ios" \
  "http://localhost:8080/api/v1/onboarding/screens/welcome?user_id=load_test"
```

#### Test Submission Endpoint

```bash
hey -n 50 -c 5 -H "Content-Type: application/json" -H "X-Device-Type: ios" \
  -d '{"user_id":"load_test","data":{"accepted_terms":true}}' \
  -m POST "http://localhost:8080/api/v1/onboarding/screens/welcome/submit"
```

### 7. WireMock Direct Testing

You can also test WireMock directly to verify mock responses:

#### Test Screen Endpoint

```bash
curl -H "Content-Type: application/json" -H "X-API-Version: 2.0" \
  "http://localhost:8080/api/screens/welcome?user_id=test123&device_type=ios"
```

#### Test Submission Endpoint

```bash
curl -X POST -H "Content-Type: application/json" -H "X-API-Version: 2.0" \
  "http://localhost:8080/api/screens/welcome/submit" \
  -d '{"user_id":"test123","data":{"accepted_terms":true}}'
```

### 8. Automated Test Script

Create a test script for complete flow testing:

```bash
#!/bin/bash
# test_onboarding_flow.sh

USER_ID="test_$(date +%s)"
BASE_URL="http://localhost:8080"

echo "Testing complete onboarding flow for user: $USER_ID"

# Test 1: Welcome screen
echo "1. Getting welcome screen..."
curl -s -X GET "$BASE_URL/api/v1/onboarding/screens/welcome?user_id=$USER_ID" \
  -H "Content-Type: application/json" -H "X-Device-Type: ios" | jq .

# Test 2: Submit welcome
echo "2. Submitting welcome screen..."
curl -s -X POST "$BASE_URL/api/v1/onboarding/screens/welcome/submit" \
  -H "Content-Type: application/json" -H "X-Device-Type: ios" \
  -d "{\"user_id\":\"$USER_ID\",\"data\":{\"accepted_terms\":true}}" | jq .

# Test 3: Check progress
echo "3. Checking progress..."
curl -s -X GET "$BASE_URL/api/v1/onboarding/progress/$USER_ID" \
  -H "Content-Type: application/json" | jq .

# Test 4: Get flow
echo "4. Getting flow..."
curl -s -X GET "$BASE_URL/api/v1/onboarding/flow/$USER_ID" \
  -H "Content-Type: application/json" | jq .

echo "Flow test completed for user: $USER_ID"
```

Make it executable and run:

```bash
chmod +x test_onboarding_flow.sh
./test_onboarding_flow.sh
```

### 9. Monitoring and Debugging

#### Check Application Logs

The application logs will show detailed information about requests and responses.

#### Check WireMock Logs

```bash
docker-compose logs wiremock
```

#### View WireMock Admin Interface

Open in browser: http://localhost:8080/__admin

#### Reset WireMock State

```bash
curl -X POST http://localhost:8080/__admin/reset
```

### 10. Performance Metrics

The BFF exposes metrics at:

```bash
curl http://localhost:8080/metrics
```

This provides Prometheus-compatible metrics for monitoring request rates, response times, and error rates.

## 🔧 Configuration

### Environment Variables

```