# ⚠️ PROTOTYPE VERSION - WORK IN PROGRESS ⚠️

**This is a prototype version of the Expense Tracker Backend.**

- This codebase is still under active development
- Features may be incomplete or subject to change
- Not recommended for production use
- API endpoints and data structures may be modified
- Documentation may be outdated

---

# Expense Tracker API

A robust Go-based API service for managing expenses, projects, wallets, and contacts with authentication and authorization.

## Features

- User authentication and authorization
- Project management
- Wallet tracking
- Contact management
- Tag-based organization
- OAuth integration (Google)
- RESTful API design
- Swagger documentation

## Getting Started

### Prerequisites

- Go 1.22 or newer
- PostgreSQL 15+
- Docker and Docker Compose

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/expense-tracker-backend-prototype.git
cd expense-tracker-backend-prototype
```

2. Set up the database:

```bash
make docker-run
make db-up
```

3. Generate SQLC code:

```bash
make sqlc
```

4. Run the application:

```bash
make run
```

### Development

The project uses Air for live reloading during development:

```bash
make watch
```

### Testing

Run the test suite:

```bash
make test        # Unit tests
make itest       # Integration tests
```

### Documentation

API documentation is available in Swagger format at `/docs/swagger.json` when the server is running.

## Project Structure

```
expense-tracker-backend-prototype/
├── cmd/                    # Main application entry point
├── config/                # Configuration files
├── internal/             # Private application code
│   ├── app/             # Application setup
│   ├── auth/            # Authentication and authorization
│   ├── contacts/        # Contact management
│   ├── core/            # Core utilities and types
│   ├── db/              # Database operations
│   ├── projects/        # Project management
│   ├── server/          # Server configuration
│   ├── tags/            # Tag management
│   ├── users/           # User management
│   ├── utils/           # Utility functions
│   ├── validate/        # Custom validators
│   └── wallets/         # Wallet management
├── pkg/                  # Public library code
├── scripts/             # Development scripts
└── sqlc.yaml           # SQLC configuration
```

## Architecture

The project follows a clean architecture pattern with distinct layers:

1. **Repository Layer**: Handles database operations using SQLC
2. **Service Layer**: Contains business logic and validation
3. **Handler Layer**: HTTP request handling and response formatting
4. **Routes Layer**: API endpoint definitions and middleware

## Database Management

### Migrations

Migrations are managed using Goose:

```bash
make db-up    # Apply migrations
make db-down  # Rollback migrations
```

### SQLC

SQLC is used for type-safe database operations:

```bash
make sqlc    # Generate database code
```

## Make Commands

- `make all` - Run build with tests
- `make build` - Build the application
- `make run` - Run the application
- `make docker-run` - Start database container
- `make docker-down` - Stop database container
- `make test` - Run unit tests
- `make itest` - Run integration tests
- `make watch` - Live reload during development
- `make clean` - Clean build artifacts

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
