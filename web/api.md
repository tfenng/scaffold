# API Documentation

Base URL: `http://localhost:8080`

## User Endpoints

### Create User

**POST** `/users`

Request:
```json
{
  "email": "user@example.com",
  "name": "User Name",
  "used_name": "Old Name",
  "company": "Company Name",
  "birth": "1990-01-15"
}
```

Response (201):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "User Name",
  "used_name": "Old Name",
  "company": "Company Name",
  "birth": "1990-01-15",
  "created_at": "2026-02-28T12:00:00Z",
  "updated_at": "2026-02-28T12:00:00Z"
}
```

---

### Get User

**GET** `/users/:id`

Response (200):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "User Name",
  "used_name": "Old Name",
  "company": "Company Name",
  "birth": "1990-01-15",
  "created_at": "2026-02-28T12:00:00Z",
  "updated_at": "2026-02-28T12:00:00Z"
}
```

---

### List Users

**GET** `/users`

Query Parameters:
| Parameter | Type | Description |
|-----------|------|-------------|
| email | string | Filter by exact email |
| name_like | string | Fuzzy search by name |
| page | int | Page number (default: 1) |
| page_size | int | Page size (default: 20, max: 200) |

Example:
```
GET /users?page=1&page_size=20&name_like=john
```

Response (200):
```json
{
  "items": [
    {
      "id": 1,
      "email": "user@example.com",
      "name": "User Name",
      "used_name": "Old Name",
      "company": "Company Name",
      "birth": "1990-01-15",
      "created_at": "2026-02-28T12:00:00Z",
      "updated_at": "2026-02-28T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

---

### Update User

**PUT** `/users/:id`

Request:
```json
{
  "name": "New Name",
  "used_name": "Old Name",
  "company": "New Company",
  "birth": "1990-01-15"
}
```

Response (200):
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "New Name",
  "used_name": "Old Name",
  "company": "New Company",
  "birth": "1990-01-15",
  "created_at": "2026-02-28T12:00:00Z",
  "updated_at": "2026-02-28T13:00:00Z"
}
```

---

### Delete User

**DELETE** `/users/:id`

Response: 204 No Content

---

## Error Responses

Error responses follow this format:
```json
{
  "code": "NOT_FOUND",
  "message": "user not found"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| INVALID_ARGUMENT | 400 | Invalid input parameters |
| NOT_FOUND | 404 | Resource not found |
| CONFLICT | 409 | Resource conflict (e.g., duplicate email) |
| INTERNAL | 500 | Internal server error |
