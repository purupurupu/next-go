# Authentication API

## Overview

認証は JWT トークンを使用して処理されます。トークンは Authorization ヘッダーで返され、後続のリクエストに含める必要があります。

## Endpoints

### User Registration

Create a new user account.

**Endpoint:** `POST /auth/sign_up`

**Request Body:**
```json
{
  "user": {
    "email": "user@example.com",
    "password": "password123",
    "password_confirmation": "password123",
    "name": "John Doe"
  }
}
```

**Success Response (200 OK):**

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiJ9...
```

**Body (UNIFIED FORMAT):**
```json
{
  "message": "Signed up successfully.",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2024-01-01T00:00:00.000Z"
  }
}
```

**Error Response (422 Unprocessable Entity):**
```json
{
  "errors": {
    "email": ["has already been taken"],
    "password": ["is too short (minimum is 6 characters)"],
    "password_confirmation": ["doesn't match Password"]
  }
}
```

### User Login

Authenticate an existing user.

**Endpoint:** `POST /auth/sign_in`

**Request Body:**
```json
{
  "user": {
    "email": "user@example.com",
    "password": "password123"
  }
}
```

**Success Response (200 OK):**

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiJ9...
```

**Body (UNIFIED FORMAT):**
```json
{
  "message": "Logged in successfully.",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2024-01-01T00:00:00.000Z"
  }
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "Invalid email or password."
}
```

### User Logout

Revoke the current JWT token.

**Endpoint:** `DELETE /auth/sign_out`

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200 OK):**
```json
{
  "message": "Logged out successfully."
}
```

**Error Response (401 Unauthorized):**
```json
{
  "error": "Couldn't find an active session."
}
```

## JWT Token Details

### Token Structure
```
eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIiwic2NwIjoidXNlciIsImF1ZCI6bnVsbCwiaWF0IjoxNjQwOTk1MjAwLCJleHAiOjE2NDEwODE2MDAsImp0aSI6IjEyMzQ1Njc4OTAifQ.signature
```

### Token Payload
```json
{
  "sub": "1",           // User ID
  "scp": "user",        // Scope
  "aud": null,          // Audience
  "iat": 1640995200,    // Issued at
  "exp": 1641081600,    // Expiration (24 hours)
  "jti": "1234567890"   // JWT ID (for revocation)
}
```

### Token Usage
Include the token in the Authorization header for authenticated requests:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiJ9...
```

### Token Expiration
- Default expiration: 24 hours
- After expiration, user must login again
- No refresh token mechanism currently implemented

### Token Revocation
- ログアウト時にトークンが無効化されます
- 無効化されたトークンは `jwt_denylists` テーブルに保存されます
- 無効化されたトークンは有効期限前でも使用できません

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 401 | Invalid email or password | Login credentials incorrect |
| 401 | Couldn't find an active session | No valid token or already logged out |
| 401 | Invalid token | Token is malformed or revoked |
| 422 | User couldn't be created successfully | Registration validation errors |

## Security Best Practices

1. **Password Requirements**
   - Minimum 6 characters
   - Should contain mix of letters, numbers, symbols (frontend validation)

2. **Token Storage**
   - Store in localStorage or sessionStorage
   - Consider HttpOnly cookies for production
   - Clear on logout

3. **Token Transmission**
   - Always use HTTPS in production
   - Never include token in URLs
   - Use Authorization header

4. **Account Security**
   - Implement account lockout after failed attempts
   - Add email verification (future enhancement)
   - Implement password reset (future enhancement)

## Frontend Integration Example

```javascript
// Authentication Client
class AuthClient {
  async login(email, password) {
    const response = await fetch('/auth/sign_in', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({
        user: { email, password }
      })
    });
    
    const data = await response.json();
    if (response.ok) {
      // Extract token from Authorization header
      const authHeader = response.headers.get('Authorization');
      const token = authHeader?.replace('Bearer ', '');
      if (token) {
        localStorage.setItem('token', token);
      }
      return data;
    }
    throw new Error(data.error);
  }
  
  async logout() {
    const token = localStorage.getItem('token');
    await fetch('/auth/sign_out', {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    localStorage.removeItem('token');
  }
}
```