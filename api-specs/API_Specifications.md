# NapaChat API Documentation

## OpenAPI Specification

**Version:** 3.0.3  
**Title:** NapaChat  
**Version:** 1.0.0  
**Description:** RESTful endpoints for NapaChat, supporting both synchronous and asynchronous messaging.  

## Servers
- **Base URL:** `https://api.napachat.com/v1`

## Security
### Authentication Method
- **Type:** HTTP
- **Scheme:** Bearer
- **Bearer Format:** JWT

## Components
### Schemas
#### User
```yaml
id: string (uuid)
email: string (email)
role: enum (Not_Student, Student, Admin)
createdAt: string (date-time)
```
#### LoginRequest
```yaml
email: string (email)
password: string (password)
```
#### LoginResponse
```yaml
token: string
expiresIn: integer (Time in seconds until the token expires)
```

## Endpoints
### Authentication
#### Register a new user
**POST** `/auth/register`
- **Tags:** Auth
- **Request Body:**
  - **Schema:** `User`
- **Responses:**
  - `201`: User created
  - `400`: Bad Request
  - `409`: Email already registered

#### Log in to obtain JWT
**POST** `/auth/login`
- **Tags:** Auth
- **Request Body:**
  - **Schema:** `LoginRequest`
- **Responses:**
  - `200`: Returns JWT
  - `401`: Invalid credentials

### Students
#### Verify student license
**POST** `/students/verify`
- **Tags:** Students
- **Security:** BearerAuth
- **Request Body:**
  - `licenseNumber: string`
  - `institutionEmail: string`
- **Responses:**
  - `200`: Verification successful
  - `400`: Invalid license data
  - `401`: Unauthorized

### Messaging
#### Send an encrypted message (Asynchronous)
**POST** `/messages`
- **Tags:** Messages
- **Security:** BearerAuth
- **Description:** Messages are queued and delivered asynchronously.
- **Request Body:**
  - `recipientId: string`
  - `content: string`
  - `attachments: array (binary)` (Optional)
- **Responses:**
  - `201`: Message stored for delivery
  - `400`: Bad request
  - `401`: Unauthorized

#### Retrieve messages (Synchronous)
**GET** `/messages/{conversationId}`
- **Tags:** Messages
- **Security:** BearerAuth
- **Description:** Fetch stored messages in real-time.
- **Parameters:**
  - `conversationId: string`
  - `limit: integer` (Optional, for pagination)
- **Responses:**
  - `200`: List of messages
  - `401`: Unauthorized
  - `403`: Forbidden

### Contacts
#### Retrieve user contacts
**GET** `/contacts`
- **Tags:** Contacts
- **Security:** BearerAuth
- **Query Parameters:**
  - `role: Not_Student | Student` (Optional, filter contacts)
- **Responses:**
  - `200`: List of contacts
  - `401`: Unauthorized

### Institutions
#### Join an institution/hospital
**POST** `/institutions/join`
- **Tags:** Institutions
- **Security:** BearerAuth
- **Request Body:**
  - `institutionId: string`
- **Responses:**
  - `200`: Successfully joined
  - `404`: Institution not found
  - `403`: Forbidden

### Access Requests
#### Request access to a student
**POST** `/students/{studentId}/request-access`
- **Tags:** Students
- **Security:** BearerAuth
- **Path Parameters:**
  - `studentId: string`
- **Request Body:**
  - `message: string` (Optional)
- **Responses:**
  - `201`: Request sent
  - `404`: Student not found
  - `409`: Request already exists

#### Accept a not_student's request
**PUT** `/not_students/requests/{requestId}`
- **Tags:** Not_Students
- **Security:** BearerAuth
- **Path Parameters:**
  - `requestId: string`
- **Responses:**
  - `200`: Request accepted
  - `403`: Forbidden
  - `404`: Request not found

