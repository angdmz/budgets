# Budget Management System

A complete Budget Management System with multi-user support, SSO authentication, and group-based data isolation.

## Features

- **Multi-user system** with Google OAuth SSO authentication
- **Group-based data isolation** - users can only see data from groups they belong to
- **Budgets** with expected vs actual expenses tracking
- **Encrypted monetary data** - all money values are encrypted at rest using Fernet
- **RESTful JSON API** with OpenAPI/Swagger documentation
- **Production-ready architecture** with clean separation of concerns

## Architecture

### Backend (Go)
- **Framework**: GinGonic
- **Database**: PostgreSQL with pgx driver
- **Authentication**: JWT tokens with Google OAuth
- **Documentation**: Swagger/OpenAPI via swaggo

### Migrations Layer (Python)
- **ORM**: SQLAlchemy
- **Migration tool**: Alembic
- **Encryption**: cryptography (Fernet)

## Project Structure

```
budgets/
├── docker-compose.yml          # Docker orchestration
├── .env.example                # Environment variables template
├── core/                       # Go backend
│   ├── main.go                 # Application entry point
│   ├── Dockerfile              # Multi-stage Dockerfile (base, test, builder, production)
│   ├── internal/
│   │   ├── config/             # Configuration management
│   │   ├── database/           # Database connection & transactions
│   │   ├── domain/             # Domain models & errors
│   │   ├── encryption/         # Fernet encryption utilities
│   │   ├── handler/            # HTTP handlers with Swagger annotations
│   │   ├── middleware/         # Auth middleware
│   │   ├── repository/         # Data access layer
│   │   ├── secrets/            # Secrets provider abstraction
│   │   ├── server/             # Server setup & routing
│   │   └── service/            # Business logic layer
│   └── docs/                   # Generated Swagger docs
└── migrations/                 # Python migrations
    ├── Dockerfile
    ├── alembic.ini
    ├── env.py
    ├── entities.py             # SQLAlchemy models
    ├── requirements.txt
    ├── encryption/             # Encryption utilities
    ├── secrets/                # Secrets provider abstraction
    └── versions/               # Alembic migration files
```

---

## Quick Start

### Prerequisites

- Docker and Docker Compose (or Podman with podman-compose)

### Step 1: Clone and Configure

```bash
git clone <repository-url>
cd budgets

# Copy environment template
cp .env.example .env
```

### Step 2: Create Secrets Files

The system uses Docker secrets for sensitive data. Create the secrets directory and files:

```bash
mkdir -p secrets

# Database password
echo -n "your-secure-db-password" > secrets/db_password.txt

# Generate Fernet encryption key
docker run --rm python:3.13-slim sh -c "pip install -q cryptography && python -c \"from cryptography.fernet import Fernet; print(Fernet.generate_key().decode(), end='')\"" > secrets/encryption_key.txt

# JWT secret (at least 32 characters)
echo -n "your-jwt-secret-at-least-32-characters-long" > secrets/jwt_secret.txt

# Google OAuth secret (optional, can be empty)
echo -n "" > secrets/google_client_secret.txt
```

### Step 3: Configure Google OAuth (Optional)

If you want SSO authentication:

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Navigate to **APIs & Services** > **Credentials**
4. Create **OAuth 2.0 Client ID** (Web application)
5. Add `http://localhost:8080/api/v1/auth/google/callback` as authorized redirect URI
6. Set `GOOGLE_CLIENT_ID` in your `.env` file
7. Add the client secret to `secrets/google_client_secret.txt`

### Step 4: Start the System

```bash
docker-compose up --build
```

This will:
1. Start PostgreSQL database (waits for health check)
2. Run database migrations automatically
3. Start the Go API server

### Step 5: Access the API

- **API Base URL**: http://localhost:8080/api/v1
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health

---

## Docker Compose Commands

### Starting Services

```bash
# Build and start all services (foreground)
docker-compose up --build

# Build and start all services (background/detached)
docker-compose up --build -d

# Start without rebuilding (if already built)
docker-compose up -d
```

### Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (deletes database data)
docker-compose down -v

# Stop and remove everything including orphan containers
docker-compose down -v --remove-orphans
```

### Viewing Logs

```bash
# View logs from all services
docker-compose logs

# Follow logs in real-time
docker-compose logs -f

# View logs for a specific service
docker-compose logs api
docker-compose logs db
docker-compose logs migrations
```

### Running Tests

```bash
# Run Go tests using the test profile
docker-compose --profile test run --rm test
```

### Database Operations

```bash
# Start only the database
docker-compose up -d db

# Run migrations manually
docker-compose run --rm migrations

# Connect to database with psql
docker-compose exec db psql -U postgres -d budgets
```

### Migration Operations

All migration operations are available as dedicated docker-compose services:

```bash
# Create a new migration (autogenerate from model changes)
docker-compose --profile tools run --rm migrate-create "your migration description"

# Apply all pending migrations
docker-compose --profile tools run --rm migrate-upgrade

# Upgrade to a specific revision
docker-compose --profile tools run --rm migrate-upgrade <revision>

# Rollback last migration
docker-compose --profile tools run --rm migrate-downgrade

# Rollback multiple migrations
docker-compose --profile tools run --rm migrate-downgrade -2

# View migration history
docker-compose --profile tools run --rm migrate-history
```

### Rebuilding Services

```bash
# Rebuild a specific service
docker-compose build api
docker-compose build migrations

# Rebuild all services (no cache)
docker-compose build --no-cache

# Rebuild and restart a specific service
docker-compose up --build -d api
```

---

## Environment Variables

Create a `.env` file in the project root with these variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_USERNAME` | PostgreSQL username | `postgres` | No |
| `DB_PASSWORD` | PostgreSQL password | `postgres` | No |
| `DB_NAME` | Database name | `budgets` | No |
| `DB_PORT` | Database port (host mapping) | `5432` | No |
| `API_PORT` | API server port (host mapping) | `8080` | No |
| `SERVER_ENV` | Environment mode | `development` | No |
| `ENCRYPTION_KEY` | Fernet encryption key | - | **Yes** |
| `JWT_SECRET` | JWT signing secret | `change-me-in-production` | No |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | - | No |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | - | No |
| `SECRETS_PROVIDER` | Secrets source (`env`/`aws`/`localstack`) | `env` | No |

### Example `.env` file

```bash
# Database
DB_USERNAME=postgres
DB_PASSWORD=mysecretpassword
DB_NAME=budgets
DB_PORT=5432

# API
API_PORT=8080
SERVER_ENV=development

# Security (REQUIRED - generate with command above)
ENCRYPTION_KEY=your-generated-fernet-key-here
JWT_SECRET=your-jwt-secret-change-in-production

# Google OAuth (optional)
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Secrets Provider
SECRETS_PROVIDER=env
```

---

## API Endpoints

All endpoints are prefixed with `/api/v1`.

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/auth/google/login` | Initiate Google OAuth login |
| GET | `/auth/google/callback` | OAuth callback (returns JWT) |
| GET | `/auth/me` | Get current user info (requires auth) |

### Groups

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/groups` | Create a new group |
| GET | `/groups` | List user's groups |
| GET | `/groups/:id` | Get group by ID |
| PUT | `/groups/:id` | Update group |
| DELETE | `/groups/:id` | Soft delete group |

### Categories

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/groups/:id/categories` | Create category in group |
| GET | `/groups/:id/categories` | List group categories |
| PUT | `/categories/:id` | Update category |
| DELETE | `/categories/:id` | Soft delete category |

### Budgets

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/groups/:id/budgets` | Create budget in group |
| GET | `/groups/:id/budgets` | List group budgets |
| GET | `/budgets/:id` | Get budget by ID |
| PUT | `/budgets/:id` | Update budget |
| DELETE | `/budgets/:id` | Soft delete budget |

### Expected Expenses

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/budgets/:id/expected-expenses` | Create expected expense |
| GET | `/budgets/:id/expected-expenses` | List expected expenses |
| PUT | `/expected-expenses/:id` | Update expected expense |
| DELETE | `/expected-expenses/:id` | Soft delete expected expense |

### Actual Expenses

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/budgets/:id/actual-expenses` | Create actual expense |
| GET | `/budgets/:id/actual-expenses` | List actual expenses |
| PUT | `/actual-expenses/:id` | Update actual expense |
| DELETE | `/actual-expenses/:id` | Soft delete actual expense |

---

## Troubleshooting

### Database connection issues

```bash
# Check if database is running
docker-compose ps

# Check database logs
docker-compose logs db

# Restart database
docker-compose restart db
```

### Migrations failing

```bash
# Check migration logs
docker-compose logs migrations

# Run migrations manually with verbose output
docker-compose run --rm migrations alembic upgrade head --sql
```

### API not starting

```bash
# Check API logs
docker-compose logs api

# Rebuild API container
docker-compose build --no-cache api
docker-compose up -d api
```

### Port already in use

Change the port mapping in your `.env` file:

```bash
API_PORT=3000  # Change from 8080
DB_PORT=5433   # Change from 5432
```

### Reset everything

```bash
# Stop all containers and remove volumes
docker-compose down -v --remove-orphans

# Remove all images
docker-compose down --rmi all

# Start fresh
docker-compose up --build
```

---

## Security

- **Encryption at rest**: All monetary values are encrypted using Fernet symmetric encryption
- **Cross-language compatibility**: Encryption works identically in Python (migrations) and Go (API)
- **JWT authentication**: Stateless authentication with configurable expiration
- **Group-based authorization**: Users can only access data from groups they belong to
- **Soft deletion**: All deletes are soft deletes (revoked_at timestamp) for audit trail

## Secrets Management

The system supports multiple secrets providers:

- **env**: Environment variables (default, recommended for development)
- **aws**: AWS Secrets Manager (recommended for production)
- **localstack**: LocalStack for local AWS development/testing

Configure via `SECRETS_PROVIDER` environment variable.

---

## License

MIT License - see LICENSE file for details