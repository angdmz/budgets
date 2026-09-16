# Database Migrations (Python + Alembic)

## Overview

Alembic-based database migration tool for the Budget Management System. Manages PostgreSQL schema changes, including encrypted columns for monetary values using Fernet encryption (compatible with the Go backend).

## Technology Stack

- **Language**: Python 3.12
- **Migration Framework**: Alembic
- **ORM**: SQLAlchemy (for model definitions in `entities.py`)
- **Encryption**: Fernet (cryptography library, compatible with Go backend)

## Project Structure

```
migrations/
├── env.py                    # Alembic environment config (loads DB connection + secrets)
├── entities.py               # SQLAlchemy models and metadata
├── alembic.ini               # Alembic configuration
├── requirements.txt          # Python dependencies
├── Dockerfile                # Container definition
├── secrets/                  # Secrets provider abstraction
│   ├── factory.py            # Provider selection (env, docker, aws, localstack)
│   ├── env_provider.py       # Environment variable secrets provider
│   ├── docker_provider.py    # Docker secrets provider
│   ├── aws_provider.py       # AWS Secrets Manager provider
│   └── localstack_provider.py # LocalStack provider
├── encryption/               # Encryption utilities for migrations
│   ├── encrypted_money.py   # Encrypted monetary value type
│   └── encrypted_string.py  # Encrypted string type
└── versions/                 # Migration scripts
```

## Environment Variables

The migrations service reads from the root `.env` file (via `env_file: .env` in `docker-compose.yml`). Configuration is loaded in `env.py` → `load_database_config()`:

| Variable | Default | Used In | Purpose |
|----------|---------|---------|---------|
| `DB_HOSTNAME` | `localhost` | `env.py:39` → `db_config["hostname"]` | PostgreSQL server hostname. In Docker Compose, set to `db` (the compose service name). |
| `DB_PORT` | `5432` | `env.py:40` → `db_config["port"]` | PostgreSQL server port. |
| `DB_USERNAME` | `postgres` | `env.py:41` → `db_config["username"]` | PostgreSQL database username. |
| `DB_NAME` | `budgets` | `env.py:43` → `db_config["name"]` | PostgreSQL database name. |
| `DB_URL_PREFIX` | `postgresql` | `env.py:38` → `db_config["url_prefix"]` | SQLAlchemy driver prefix for the connection URL. |
| `DB_SCHEMA` | `public` | `env.py:44` → `db_config["schema"]` | PostgreSQL schema name. |
| `SECRETS_PROVIDER` | `env` | `secrets/factory.py:44` → `get_secrets_provider()` | Selects the secrets provider: `env`, `docker`, `aws`, `localstack`. |
| `SECRETS_PREFIX` | _(empty)_ | `secrets/factory.py:49` | Prefix prepended to secret keys when using the `env` provider. |
| `AWS_REGION` | `us-east-1` | `secrets/factory.py:54` | AWS region. Used when `SECRETS_PROVIDER=aws` or `localstack`. |
| `AWS_SECRET_NAME` | _(empty)_ | `secrets/factory.py:55` | AWS Secrets Manager secret name. Used when `SECRETS_PROVIDER=aws` or `localstack`. |
| `LOCALSTACK_ENDPOINT` | `http://localhost:4566` | `secrets/factory.py:65` | LocalStack endpoint URL. Used when `SECRETS_PROVIDER=localstack`. |
| `DOCKER_SECRETS_PATH` | `/run/secrets` | `secrets/factory.py:73` | Docker secrets directory. Used when `SECRETS_PROVIDER=docker`. |

### Secrets

The migrations service requires two secrets, retrieved through the same secrets provider abstraction as the backend:

| Secret Key | Retrieved In | Used For |
|------------|--------------|----------|
| `db_password` | `env.py:28` → `provider.get_secret("db_password")` | PostgreSQL database password. Falls back to `DB_PASSWORD` env var if not found via the secrets provider. Used to build the SQLAlchemy connection URL. |
| `encryption_key` | `env.py:33` → `provider.get_secret("encryption_key")` | Fernet encryption key. Used by `entities.py:_get_encryptor()` to encrypt/decrypt monetary values during migrations that involve encrypted columns. Falls back to `ENCRYPTION_KEY` env var if not found via the secrets provider. |

**For development with environment variables (`SECRETS_PROVIDER=env`):**
```bash
export DB_PASSWORD="your-db-password"
export ENCRYPTION_KEY="your-fernet-key"
```

**For Docker Compose (`SECRETS_PROVIDER=docker`):**
Secrets are mounted at `/run/secrets/db_password` and `/run/secrets/encryption_key` via Docker secrets.

## Running Migrations

### With Docker Compose (recommended)

```bash
# Apply all pending migrations
docker-compose run --rm migrations

# Create a new migration (autogenerate)
docker-compose --profile tools run --rm migrate-create "your migration description"

# Rollback last migration
docker-compose --profile tools run --rm migrate-downgrade

# View migration history
docker-compose --profile tools run --rm migrate-history
```

### Locally

```bash
cd migrations
pip install -r requirements.txt
alembic upgrade head
```

## License

See LICENSE file in repository root.
