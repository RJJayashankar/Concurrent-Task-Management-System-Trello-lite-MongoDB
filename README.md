# Concurrent-Task-Management-System-Trello-lite-MongoDB
Creating the concurrent task managament system using Go and MongoDB


# Trello Lite

A minimal Trello-like backend in Go using MongoDB. Includes JWT-based auth, RBAC, background workers, and aggregated endpoints.

## Features
- User signup & login with JWT ([`utils.GenerateJWT`](trello-lite/utils/jwt.go), [`utils.JwtKey`](trello-lite/utils/jwt.go))
- Role-based access control via middleware ([`middleware.AuthMiddleware`](trello-lite/middleware/auth.go))
- CRUD for projects and tasks with aggregation pipelines ([`handlers.CreateProjectHandler`](trello-lite/handlers/project-handler.go), [`handlers.CreateTaskHandler`](trello-lite/handlers/task_handler.go))
- Search, update, delete task flows ([`handlers.SearchTaskHandler`](trello-lite/handlers/task_handler.go), [`handlers.UpdateTaskStatusHandler`](trello-lite/handlers/task_handler.go), [`handlers.DeleteTaskHandler`](trello-lite/handlers/task_handler.go))
- Background overdue-task scanner ([`workers.StartOverdueScanner`](trello-lite/workers/task_worker.go))
- Standardized JSON responses ([`utils.SendSuccess`](trello-lite/utils/response.go), [`utils.SendError`](trello-lite/utils/response.go))
- MongoDB index creation on startup ([`databases.ConnectDB`](trello-lite/databases/mongodb.go), [`databases.CreateIndexes`](trello-lite/databases/mongodb.go))

## Repo layout (key files)
- [main.go](trello-lite/main.go) — server routes and startup
- [trello-lite/databases/mongodb.go](trello-lite/databases/mongodb.go) — MongoDB connection & indexes (`databases.GetCollection`)
- [trello-lite/middleware/auth.go](trello-lite/middleware/auth.go) — JWT auth middleware
- [trello-lite/handlers/project-handler.go](trello-lite/handlers/project-handler.go)
- [trello-lite/handlers/task_handler.go](trello-lite/handlers/task_handler.go)
- [trello-lite/handlers/user_handler.go](trello-lite/handlers/user_handler.go)
- [trello-lite/models/user.go](trello-lite/models/user.go), [trello-lite/models/project.go](trello-lite/models/project.go), [trello-lite/models/tasks.go](trello-lite/models/tasks.go)
- [trello-lite/utils/jwt.go](trello-lite/utils/jwt.go), [trello-lite/utils/response.go](trello-lite/utils/response.go)
- [trello-lite/workers/task_worker.go](trello-lite/workers/task_worker.go)

# Trello API Documentation

A comprehensive REST API for managing users, projects, and tasks in a Trello-like application.

## Base URL
```
http://localhost:8080
```

## Table of Contents
- [Authentication](#authentication)
- [API Endpoints](#api-endpoints)
  - [User Management](#user-management)
  - [Project Management](#project-management)
  - [Task Management](#task-management)
  - [System Operations](#system-operations)
- [Setup & Installation](#setup--installation)
- [Usage Examples](#usage-examples)
- [Error Handling](#error-handling)

## Authentication

This API uses custom header-based authentication:

| Header | Description | Required |
|--------|-------------|----------|
| `Role` | User role (e.g., admin, user) | Yes |
| `User-ID` | Unique user identifier | Yes |

Include these headers in all authenticated requests.

## API Endpoints

### User Management

#### 1. User Signup
Create a new user account.

**Endpoint:** `POST /signup`

**Request Body:**
```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "success": true,
  "userId": "string",
  "message": "User created successfully"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

---

#### 2. User Login
Authenticate a user and retrieve session information.

**Endpoint:** `GET /login`

**Query Parameters:**
- `username` or `email` (string, required)
- `password` (string, required)

**Response:**
```json
{
  "success": true,
  "userId": "string",
  "role": "string",
  "token": "string"
}
```

**cURL Example:**
```bash
curl -X GET "http://localhost:8080/login?email=john@example.com&password=securepass123"
```

---

#### 3. Get All Users
Retrieve a list of all registered users.

**Endpoint:** `POST /getallusers`

**Headers:**
- `Role`: admin
- `User-ID`: string

**Response:**
```json
{
  "users": [
    {
      "userId": "string",
      "username": "string",
      "email": "string",
      "role": "string"
    }
  ]
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/getallusers \
  -H "Role: admin" \
  -H "User-ID: user123"
```

---

### Project Management

#### 4. Create Project
Create a new project.

**Endpoint:** `POST /project/create`

**Headers:**
- `Role`: string
- `User-ID`: string

**Request Body:**
```json
{
  "projectName": "string",
  "description": "string",
  "startDate": "string",
  "endDate": "string"
}
```

**Response:**
```json
{
  "success": true,
  "projectId": "string",
  "message": "Project created successfully"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/project/create \
  -H "Content-Type: application/json" \
  -H "Role: admin" \
  -H "User-ID: user123" \
  -d '{
    "projectName": "Website Redesign",
    "description": "Complete redesign of company website",
    "startDate": "2024-01-01",
    "endDate": "2024-06-30"
  }'
```

---

#### 5. Get Projects
Retrieve all projects for the authenticated user.

**Endpoint:** `GET /getProject`

**Headers:**
- `Role`: string
- `User-ID`: string

**Response:**
```json
{
  "projects": [
    {
      "projectId": "string",
      "projectName": "string",
      "description": "string",
      "startDate": "string",
      "endDate": "string",
      "status": "string"
    }
  ]
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8080/getProject \
  -H "Role: user" \
  -H "User-ID: user123"
```

---

### Task Management

#### 6. Get Tasks
Retrieve all tasks.

**Endpoint:** `GET /tasks` (Note: Collection shows this as "Get Tasks")

**Headers:**
- `Role`: string
- `User-ID`: string

**Response:**
```json
{
  "tasks": [
    {
      "taskId": "string",
      "taskName": "string",
      "description": "string",
      "status": "string",
      "assignedTo": "string",
      "projectId": "string",
      "dueDate": "string"
    }
  ]
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8080/tasks \
  -H "Role: user" \
  -H "User-ID: user123"
```

---

#### 7. Create Task
Create a new task within a project.

**Endpoint:** `POST /task/create`

**Headers:**
- `Role`: string
- `User-ID`: string

**Request Body:**
```json
{
  "taskName": "string",
  "description": "string",
  "projectId": "string",
  "assignedTo": "string",
  "dueDate": "string",
  "priority": "string"
}
```

**Response:**
```json
{
  "success": true,
  "taskId": "string",
  "message": "Task created successfully"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/task/create \
  -H "Content-Type: application/json" \
  -H "Role: user" \
  -H "User-ID: user123" \
  -d '{
    "taskName": "Design homepage mockup",
    "description": "Create initial design mockup for homepage",
    "projectId": "proj123",
    "assignedTo": "user456",
    "dueDate": "2024-02-15",
    "priority": "high"
  }'
```

---

#### 8. Update Task
Update an existing task's details.

**Endpoint:** `POST /task/update`

**Headers:**
- `Role`: string
- `User-ID`: string

**Request Body:**
```json
{
  "taskId": "string",
  "taskName": "string",
  "description": "string",
  "status": "string",
  "dueDate": "string",
  "priority": "string"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Task updated successfully"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/task/update \
  -H "Content-Type: application/json" \
  -H "Role: user" \
  -H "User-ID: user123" \
  -d '{
    "taskId": "task789",
    "status": "in-progress",
    "priority": "urgent"
  }'
```

---

#### 9. Delete Task
Delete a task by ID.

**Endpoint:** `DELETE /task/delete`

**Headers:**
- `Role`: string
- `User-ID`: string

**Query Parameters:**
- `id` (string, required) - Task ID to delete

**Response:**
```json
{
  "success": true,
  "message": "Task deleted successfully"
}
```

**cURL Example:**
```bash
curl -X DELETE "http://localhost:8080/task/delete?id=0001" \
  -H "Role: admin" \
  -H "User-ID: user123"
```

---

#### 10. Update Task Owner
Reassign a task to a different user.

**Endpoint:** `POST /taskOwnerUpdate`

**Headers:**
- `Role`: string
- `User-ID`: string

**Request Body:**
```json
{
  "taskId": "string",
  "newOwnerId": "string"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Task owner updated successfully"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/taskOwnerUpdate \
  -H "Content-Type: application/json" \
  -H "Role: admin" \
  -H "User-ID: user123" \
  -d '{
    "taskId": "task789",
    "newOwnerId": "user456"
  }'
```

---

### System Operations

#### 11. Get Everything
Retrieve all system data (users, projects, tasks).

**Endpoint:** `GET /everything`

**Headers:**
- `Role`: admin
- `User-ID`: string

**Response:**
```json
{
  "users": [...],
  "projects": [...],
  "tasks": [...]
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8080/everything \
  -H "Role: admin" \
  -H "User-ID: user123"
```

---

## Setup & Installation

### Prerequisites
- Node.js (v14 or higher)
- npm or yarn
- Database (MongoDB/PostgreSQL/MySQL)

### Installation Steps

1. **Clone the repository:**
```bash
git clone <repository-url>
cd trello-api
```

2. **Install dependencies:**
```bash
npm install
```

3. **Configure environment variables:**
Create a `.env` file in the root directory:
```env
PORT=8080
DATABASE_URL=your_database_connection_string
JWT_SECRET=your_jwt_secret
```

4. **Start the server:**
```bash
npm start
```

The API will be available at `http://localhost:8080`

---

## Usage Examples

### Complete Workflow Example

1. **Register a new user:**
```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "email": "alice@example.com", "password": "pass123"}'
```

2. **Login:**
```bash
curl -X GET "http://localhost:8080/login?email=alice@example.com&password=pass123"
```

3. **Create a project:**
```bash
curl -X POST http://localhost:8080/project/create \
  -H "Content-Type: application/json" \
  -H "Role: user" \
  -H "User-ID: alice123" \
  -d '{"projectName": "Mobile App", "description": "New mobile application"}'
```

4. **Create a task:**
```bash
curl -X POST http://localhost:8080/task/create \
  -H "Content-Type: application/json" \
  -H "Role: user" \
  -H "User-ID: alice123" \
  -d '{"taskName": "Setup project", "projectId": "proj123"}'
```

5. **Get all tasks:**
```bash
curl -X GET http://localhost:8080/tasks \
  -H "Role: user" \
  -H "User-ID: alice123"
```

---

## Error Handling

### Common HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid authentication |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error |

### Error Response Format
```json
{
  "success": false,
  "error": "Error message description",
  "code": "ERROR_CODE"
}
```

### Common Errors

**Missing Authentication Headers:**
```json
{
  "success": false,
  "error": "Role and User-ID headers are required",
  "code": "AUTH_REQUIRED"
}
```

**Invalid Permissions:**
```json
{
  "success": false,
  "error": "Insufficient permissions to perform this action",
  "code": "FORBIDDEN"
}
```

**Resource Not Found:**
```json
{
  "success": false,
  "error": "Task not found",
  "code": "NOT_FOUND"
}
```

---

## Role-Based Access Control

### Roles

| Role | Permissions |
|------|-------------|
| `admin` | Full access to all endpoints |
| `user` | Can manage own projects and tasks |
| `viewer` | Read-only access |



Note: Many routes are protected by [`middleware.AuthMiddleware`](trello-lite/middleware/auth.go). Provide `Authorization: Bearer <token>` header.

## Requirements
- Go (see module: [trello-lite/go.mod](trello-lite/go.mod))
- MongoDB running at mongodb://localhost:27017
- Recommended: run `go mod tidy` inside `trello-lite/`

## Run
1. Start MongoDB.
2. From repo root:
```sh
cd trello-lite
go mod tidy
go run .
```
The server listens on :8080 (see [trello-lite/main.go](trello-lite/main.go)).

## Notes
- Database name: `Trello_lite` (see [`databases.GetCollection`](trello-lite/databases/mongodb.go)).
- Index creation runs automatically on startup via [`databases.CreateIndexes`](trello-lite/databases/mongodb.go).
- Responses use standardized format via [`utils.SendSuccess`](trello-lite/utils/response.go) and [`utils.SendError`](trello-lite/utils/response.go).


