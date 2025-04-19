# Authentication Service

This service provides secure user authentication for the ride-sharing platform using Google OAuth2 and JWTs, with token storage in Redis. It supports login for 1,000 users and 20,000 API calls/day.

## Prerequisites

- Node.js and npm
- Docker (for Redis)
- Google OAuth2 credentials (set in `.env`)

## Setup

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/rghdasalah/Software_Architecture_Project.git
   cd Software_Architecture_Project/authentication-service
   ```

2. **Install Dependencies**:
   ```bash
   npm install
   ```

3. **Set Up Environment**:
   - Copy `.env.example` to `.env` and add credentials:
     ```bash
     cp .env.example .env
     ```
   - Edit `.env`:
     ```
     GOOGLE_CLIENT_ID=your_client_id
     GOOGLE_CLIENT_SECRET=your_client_secret
     JWT_SECRET=your_32_char_secret
     ```

4. **Run Redis**:
   ```bash
   docker run -d -p 6379:6379 --name redis redis
   ```

5. **Start the Server**:
   ```bash
   node index.js
   ```

## Usage

- **Login**: Visit `http://localhost:8082/auth/google` in a browser.
- **Response**: After Google login, `/auth/google/callback` returns:
  ```json
  { "token": "eyJhbGciOiJIUzI1Ni..." }
  ```
- **Token Storage**: JWTs are stored in Redis with a 1-hour expiry (key: `token:<userId>`).

## Testing

- **Verify Login**: Check the browser response for a JWT.
- **Check Redis**:
  ```bash
  docker exec -it redis redis-cli
  KEYS token:*
  GET token:<userId>
  ```
- **Outputs**: See `tests/jwt.txt` (JWT), `tests/redis_keys.txt` (Redis keys), and `tests/callback_response.json` (callback response).

## Notes

- Ensure `.env` is not committed (listed in `.gitignore`).

