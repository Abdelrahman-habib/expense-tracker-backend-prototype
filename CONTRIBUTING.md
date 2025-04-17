# Contributing to Expense Tracker

First off, thank you for considering contributing to Expense Tracker! It's people like you that make this project better.

## Project Structure

Our project follows a standard Go project layout with specific conventions:

```
.
├── cmd/                    # Main applications
├── internal/              # Private application code
│   ├── db/               # Database related code
│   │   ├── sql/         # SQL files
│   │   │   ├── migrations/  # Goose migration files
│   │   │   └── queries/     # SQLC query files
│   ├── feature/          # Each feature has its own package
│   │   ├── types/       # Feature-specific types and interfaces
│   │   ├── service/     # Business logic implementation
│   │   ├── handlers/    # HTTP handlers
│   │   ├── repository/  # Database operations
│   │   ├── integration/ # Integration tests
│   │   └── routes/      # Route definitions
│   ├── errors/          # Custom error types and handling
│   └── payloads/        # Response payload structures
├── pkg/                  # Public library code
│   ├── handlers/        # Base handler functionality
│   └── context/         # Context utilities
└── config/              # Configuration files
```

## Layer Architecture

Our codebase follows a clean architecture pattern with distinct layers:

1. **Repository Layer** (`repository/`)

   - Handles database operations using SQLC-generated code
   - Implements data access patterns
   - Returns domain models

2. **Service Layer** (`service/`)

   - Contains business logic
   - Orchestrates repository calls
   - Handles domain-specific operations
   - Implements validation rules (usually kept minimal as we already validate the request in the handler layers)

3. **Handler Layer** (`handlers/`)

   - HTTP request handling
   - Request validation
   - Response formatting
   - Uses chi router and render package
   - Implements swagger documentation

4. **Routes Layer** (`routes/`)
   - Defines API endpoints
   - Configures middleware
   - Groups related endpoints

## Dependency Injection

We use constructor-based dependency injection:

```go
// Handler level
type ProjectHandler struct {
    *handlers.BaseHandler
    service ProjectService
}

func NewProjectHandler(logger *zap.Logger, service ProjectService) *ProjectHandler {
    return &ProjectHandler{
        BaseHandler: handlers.NewBaseHandler(logger),
        service: service,
    }
}

// Service level
type ProjectService struct {
    repo ProjectRepository
}

func NewProjectService(repo ProjectRepository) *ProjectService {
    return &ProjectService{
        repo: repo,
    }
}
```

## Database Management

### Migrations

1. **Location**: `internal/db/sql/migrations/`
2. **Naming Convention**: `YYYYMMDDHHMMSS_description.sql`
3. **Format**:

```sql
-- +goose Up
CREATE TABLE "users" (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS "users";
```

### SQLC Queries

1. **Location**: `internal/db/sql/queries/`
2. **Naming Convention**: `<feature>.sql`
3. **Format**:

```sql
-- name: GetContact :one
SELECT * FROM contacts
WHERE contact_id = $1 AND user_id = $2 LIMIT 1;

-- name: CreateContact :one
INSERT INTO contacts (
    user_id, name, email
) VALUES ($1, $2, $3)
RETURNING *;
```

## API Documentation

We use swaggo/swag for API documentation. Documentation is generated from code annotations:

### Type Definitions

```go
// @Description Project information including details and status
type Project struct {
    ProjectID uuid.UUID `json:"projectId" example:"123e4567-e89b-12d3-a456-426614174000" format:"uuid"`
    Name      string    `json:"name" example:"My Project" minLength:"1" maxLength:"255"`
    Status    string    `json:"status" example:"ongoing" enums:"ongoing,completed,canceled"`
}
```

### Handler Documentation

```go
// @Summary Update a project
// @Description Updates an existing project
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "project ID" format(uuid)
// @Success 200 {object} payloads.Response{data=types.Project}
// @Failure 400 {object} errors.ErrorResponse
// @Router /projects/{id} [put]
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request)
```

## Response Handling

We use a standardized response format:

```go
type Response struct {
    Status  int         `json:"status"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Meta    struct {
        Query     string `json:"query,omitempty"`
        Limit     int32  `json:"limit,omitempty"`
        Count     int    `json:"count,omitempty"`
        NextToken string `json:"next_token,omitempty"`
    } `json:"meta"`
}
```

### Response Helpers

```go
// Success responses
h.Respond(w, r, payloads.OK(data))
h.Respond(w, r, payloads.Created(data))
h.Respond(w, r, payloads.Updated(data))

// Error responses
h.RespondError(w, r, errors.ErrInvalidRequest(err))
h.RespondError(w, r, errors.ErrNotFound())
```

## Error Handling

1. **Custom Error Types**: Define domain-specific errors in `internal/errors/`
2. **Error Wrapping**: Use error wrapping for context
3. **HTTP Status Codes**: Map errors to appropriate status codes
4. **Logging**: Include relevant context in error logs

## Validation

1. **Request Validation**: Use ozzo-validation
2. **Swagger Annotations**: Include validation rules in swagger docs
3. **Custom Validators**: Implement domain-specific validation rules

Example:

```go
func (c *ProjectCreatePayload) Bind(r *http.Request) error {
    return validation.Errors{
        "name": validation.Validate(c.Name,
            validation.Required,
            validation.Length(1, MaxNameLength)),
        "status": validation.Validate(c.Status,
            validation.Required,
            validation.In("ongoing", "completed", "canceled")),
    }.Filter()
}
```

## Testing

1. **Integration Tests**: Use testcontainers for database testing
2. **Unit Tests**: Mock dependencies using interfaces
3. **Test Helpers**: Utilize test fixtures and factories

## Development Workflow

1. Create a new branch
2. Implement changes following the architecture
3. Add appropriate tests
4. Update swagger annotations
5. Generate swagger docs using swag
6. Create pull request

## Running the Project

1. **Database Migrations**:

```bash
make db-up
```

```bash
make db-down
```

or simply

```bash
make goose <command>
```

2. **Generate SQLC Code**:

```bash
make sqlc
```

3. **Generate Swagger Docs**:

```bash
make docs
```

4. **Run Tests**:

```bash
make test
make itest
```

## Questions?

Feel free to open an issue for any questions about contributing. We're here to help!

## Commit Guidelines

1. **Commit Messages**

   - Use the present tense ("Add feature" not "Added feature")
   - Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
   - Structure messages as:

     ```
     <type>(<scope>): <subject>

     <body>

     <footer>
     ```

   - Types: feat, fix, docs, style, refactor, test, chore
   - Reference issues in footer

2. **Pull Requests**
   - Use the PR template
   - Include test coverage
   - Update documentation
   - Add to CHANGELOG.md if applicable

## Creating New Features

1. **Use the Feature Creation Script**

   ```bash
   ./scripts/create_feature.sh <feature_name>
   ```

2. **Feature Checklist**
   - [ ] Types defined
   - [ ] Handlers implemented
   - [ ] Routes configured
   - [ ] Tests written
   - [ ] Documentation updated
   - [ ] Migrations created (if needed)
   - [ ] Configuration updated (if needed)

## Code Review Process

1. **Review Checklist**

   - Code follows project structure
   - Tests are included and passing
   - Documentation is updated
   - No security vulnerabilities
   - Performance considerations addressed
   - Error handling is appropriate

2. **Review Response Time**
   - Initial response within 2 business days
   - Follow-up reviews within 1 business day
   - Mark PR as draft if not ready for review

## Security

1. **Security Considerations**
   - No credentials in code
   - Use environment variables for sensitive data
   - Implement proper input validation
   - Follow secure coding practices
   - Report security issues privately

## Performance

1. **Performance Guidelines**
   - Use appropriate database indexes
   - Implement pagination for list endpoints
   - Consider caching where appropriate
   - Monitor query performance
