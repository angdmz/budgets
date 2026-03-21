# Budget Management System - Backend (Go)

## Overview

The backend is a RESTful API built with Go and the GinGonic framework. It provides secure, encrypted budget management functionality with Auth0 authentication.

## Technology Stack

- **Language**: Go 1.21+
- **Framework**: GinGonic
- **Database**: PostgreSQL (via pgx)
- **Authentication**: Auth0 (RS256 JWT with JWKS)
- **Encryption**: Fernet (compatible with Python)
- **API Documentation**: Swagger/OpenAPI
- **Testing**: Testify

## Architecture

```
internal/
├── handler/          # HTTP request handlers
├── service/          # Business logic layer
├── repository/       # Data access layer
├── domain/           # Domain models and types
├── middleware/       # Auth0, config injection
├── encryption/       # Fernet encryption utilities
├── secrets/          # Secrets provider abstraction
├── config/           # Configuration management
├── database/         # Database connection and transactions
└── server/           # Server setup and routing
```

## Features

### Authentication & Authorization
- Auth0 JWT validation with JWKS
- RS256 signature verification
- Group-based data isolation
- No cross-group data access

### Data Security
- Fernet encryption for all monetary values
- Encrypted at rest in database
- Python-compatible encryption format
- Secrets management abstraction (Docker, AWS, LocalStack)

### API Endpoints (27 total)

**Groups** (5 endpoints)
- `POST /api/v1/groups` - Create group
- `GET /api/v1/groups` - List user's groups
- `GET /api/v1/groups/:id` - Get group details
- `PUT /api/v1/groups/:id` - Update group
- `DELETE /api/v1/groups/:id` - Delete group (soft delete)

**Categories** (4 endpoints)
- `POST /api/v1/groups/:id/categories` - Create category
- `GET /api/v1/groups/:id/categories` - List categories
- `PUT /api/v1/categories/:id` - Update category
- `DELETE /api/v1/categories/:id` - Delete category

**Budgets** (5 endpoints)
- `POST /api/v1/groups/:id/budgets` - Create budget
- `GET /api/v1/groups/:id/budgets` - List budgets
- `GET /api/v1/budgets/:id` - Get budget details
- `PUT /api/v1/budgets/:id` - Update budget
- `DELETE /api/v1/budgets/:id` - Delete budget

**Expected Expenses** (4 endpoints)
- `POST /api/v1/budgets/:id/expected-expenses` - Create expected expense
- `GET /api/v1/budgets/:id/expected-expenses` - List expected expenses
- `PUT /api/v1/expected-expenses/:id` - Update expected expense
- `DELETE /api/v1/expected-expenses/:id` - Delete expected expense

**Actual Expenses** (4 endpoints)
- `POST /api/v1/budgets/:id/actual-expenses` - Create actual expense
- `GET /api/v1/budgets/:id/actual-expenses` - List actual expenses
- `PUT /api/v1/actual-expenses/:id` - Update actual expense
- `DELETE /api/v1/actual-expenses/:id` - Delete actual expense

**System** (2 endpoints)
- `GET /health` - Health check
- `GET /swagger/index.html` - API documentation

## Setup & Development

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 16
- Docker (optional, for containerized development)

### Environment Variables

Create a `.env` file or set these variables:

```bash
# Server
SERVER_PORT=8080
SERVER_ENV=development

# Database
DB_HOSTNAME=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_NAME=budgets
DB_SSLMODE=disable

# Auth0
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.budget.local
AUTH0_CLIENT_ID=your-client-id

# Secrets Provider (env, docker, aws, localstack)
SECRETS_PROVIDER=env
```

### Secrets

The application requires these secrets (see `/secrets/README.md`):

- `db_password` - Database password
- `encryption_key` - Fernet encryption key
- `jwt_secret` - JWT signing secret
- `auth0_client_secret` - Auth0 client secret

**For development with environment variables:**
```bash
export ENCRYPTION_KEY="your-fernet-key"
export JWT_SECRET="your-jwt-secret"
export DB_PASSWORD="your-db-password"
export AUTH0_CLIENT_SECRET="your-auth0-secret"
```

### Running Locally

**1. Install dependencies:**
```bash
go mod download
```

**2. Run database migrations:**
```bash
cd ../migrations
python -m alembic upgrade head
```

**3. Generate Swagger documentation:**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init
```

**4. Run the server:**
```bash
go run main.go
```

The API will be available at `http://localhost:8080`

### Running with Docker

**Build the image:**
```bash
docker build -t budget-api .
```

**Run the container:**
```bash
docker run -p 8080:8080 \
  -e AUTH0_DOMAIN=your-tenant.auth0.com \
  -e AUTH0_AUDIENCE=https://api.budget.local \
  -e AUTH0_CLIENT_ID=your-client-id \
  budget-api
```

### Running Tests

**Unit tests:**
```bash
go test ./...
```

**Integration tests:**
```bash
docker-compose --profile test up test
```

**With coverage:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Configuration

### Secrets Providers

The application supports multiple secrets providers:

**1. Environment Variables** (`SECRETS_PROVIDER=env`)
```bash
export ENCRYPTION_KEY="..."
export JWT_SECRET="..."
```

**2. Docker Secrets** (`SECRETS_PROVIDER=docker`)
- Reads from `/run/secrets/`
- Used in Docker Compose

**3. AWS Secrets Manager** (`SECRETS_PROVIDER=aws`)
```bash
export AWS_REGION=us-east-1
export AWS_SECRET_PREFIX=budget-app/
```

**4. LocalStack** (`SECRETS_PROVIDER=localstack`)
- For local AWS simulation

### Database Configuration

The application uses connection pooling with these defaults:
- Max open connections: 25
- Max idle connections: 5
- Connection max lifetime: 300 seconds

Adjust via environment variables:
```bash
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME_SECONDS=600
```

## API Documentation

### Swagger UI
Access interactive API documentation at:
```
http://localhost:8080/swagger/index.html
```

### Authentication
All endpoints (except `/health` and `/swagger`) require a Bearer token:

```bash
curl -H "Authorization: Bearer YOUR_AUTH0_TOKEN" \
  http://localhost:8080/api/v1/groups
```

### Example Requests

**Create a group:**
```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Family Budget",
    "description": "Our household budget"
  }'
```

**Create a budget:**
```bash
curl -X POST http://localhost:8080/api/v1/groups/{group_id}/budgets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "March 2024",
    "start_date": "2024-03-01",
    "end_date": "2024-03-31"
  }'
```

**Add an expense:**
```bash
curl -X POST http://localhost:8080/api/v1/budgets/{budget_id}/actual-expenses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Groceries",
    "amount": {
      "amount": "150.50",
      "currency": "USD"
    },
    "expense_date": "2024-03-15"
  }'
```

## Error Handling

The API uses environment-aware error handling:

**Development** (`SERVER_ENV=development`):
- Detailed error messages
- Stack traces included
- Internal errors exposed

**Production** (`SERVER_ENV=production`):
- Generic error messages
- No internal details exposed
- Errors logged server-side

### Error Response Format

```json
{
  "error": "error_code",
  "message": "Human-readable error message"
}
```

Common error codes:
- `invalid_request` - Validation error
- `unauthorized` - Missing or invalid token
- `forbidden` - Insufficient permissions
- `not_found` - Resource not found
- `internal_error` - Server error

## Security

### Authentication Flow
1. Frontend obtains Auth0 JWT token
2. Token sent in `Authorization: Bearer {token}` header
3. Backend validates token signature using Auth0 JWKS
4. User identity extracted from `sub` claim
5. Request processed with user context

### Data Encryption
- All monetary values encrypted with Fernet
- Encryption key stored securely (never in code)
- Encrypted format: `{"amount": "100.50", "currency": "USD"}`
- Compatible with Python Fernet implementation

### Authorization
- Group-based isolation
- Users can only access their own groups
- Participant association required
- No cross-group queries possible

## Performance

### Database Queries
- O(1) queries per endpoint
- Proper indexing on foreign keys
- Connection pooling enabled
- Prepared statements used

### Caching
- No application-level caching (stateless)
- Database query results not cached
- Consider adding Redis for production

## Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok"
}
```

### Logging
- Structured logging with context
- Request ID tracking
- Error logging with stack traces
- Performance metrics logged

## Troubleshooting

### Common Issues

**1. "Failed to get encryption key"**
- Ensure `ENCRYPTION_KEY` is set or secret file exists
- Check secrets provider configuration

**2. "Invalid token"**
- Verify Auth0 configuration matches frontend
- Check token expiration
- Ensure JWKS is accessible

**3. "Database connection failed"**
- Verify PostgreSQL is running
- Check connection string
- Ensure database exists

**4. "Port already in use"**
- Change `SERVER_PORT` environment variable
- Kill existing process on port 8080

### Debug Mode

Enable debug logging:
```bash
export GIN_MODE=debug
```

## Contributing

### Code Style
- Follow Go conventions
- Use `gofmt` for formatting
- Run `golint` before committing
- Write tests for new features

### Adding New Endpoints

1. Define handler in `internal/handler/`
2. Add business logic in `internal/service/`
3. Add data access in `internal/repository/`
4. Register route in `internal/server/server.go`
5. Add Swagger annotations
6. Write tests
7. Update API documentation

## Production Deployment

### Checklist
- [ ] Set `SERVER_ENV=production`
- [ ] Use managed PostgreSQL (AWS RDS, etc.)
- [ ] Use AWS Secrets Manager for secrets
- [ ] Enable HTTPS
- [ ] Configure monitoring (Datadog, New Relic)
- [ ] Set up log aggregation
- [ ] Configure backups
- [ ] Load testing completed
- [ ] Security audit performed

### Docker Production Build

```bash
docker build --target production -t budget-api:latest .
```

### Environment Variables for Production

```bash
SERVER_ENV=production
SERVER_PORT=8080
DB_HOSTNAME=prod-db.example.com
DB_SSLMODE=require
SECRETS_PROVIDER=aws
AWS_REGION=us-east-1
```

## License

See LICENSE file in repository root.

## Support

For issues or questions:
- Check Swagger documentation: `/swagger/index.html`
- Review error logs
- Check Auth0 dashboard for authentication issues
