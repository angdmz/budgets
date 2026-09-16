# Secrets Directory

This directory contains Docker secrets files. These files are mounted into containers at `/run/secrets/` via the `secrets:` section in `docker-compose.yml`.

## Required Files

Create the following files with your secret values (one value per file, no trailing newline):

### `db_password.txt`
Database password for PostgreSQL.

```bash
echo -n "your-secure-db-password" > db_password.txt
```

**Used by:**
- **`db` service** — PostgreSQL reads it via `POSTGRES_PASSWORD_FILE=/run/secrets/db_password` (see `docker-compose.yml:10`).
- **`api` service (Go backend)** — Retrieved by `config.Load()` in `core/internal/config/config.go:81` to build the database connection string.
- **`migrations` service** — Retrieved by `load_database_config()` in `migrations/env.py:28` to build the SQLAlchemy connection URL.
- **`test` service** — Same as `api` service (uses the Go backend's test target).

### `encryption_key.txt`
Fernet encryption key for encrypting sensitive data. Generate with:

```bash
docker run --rm python:3.13-slim sh -c "pip install -q cryptography && python -c \"from cryptography.fernet import Fernet; print(Fernet.generate_key().decode(), end='')\"" > encryption_key.txt
```

**Used by:**
- **`api` service (Go backend)** — Retrieved by `config.Load()` in `core/internal/config/config.go:71`. Passed to `encryption.NewEncryptor()` in `main.go:57` to encrypt/decrypt all monetary values at rest.
- **`migrations` service** — Retrieved by `load_database_config()` in `migrations/env.py:33`. Used by `entities.py:_get_encryptor()` to encrypt/decrypt monetary values during data migrations involving encrypted columns.
- **`test` service** — Same as `api` service (uses the Go backend's test target).

### `jwt_secret.txt`
Secret key for signing JWT tokens.

```bash
echo -n "your-jwt-secret-at-least-32-chars" > jwt_secret.txt
```

**Used by:**
- **`api` service (Go backend)** — Retrieved by `config.Load()` in `core/internal/config/config.go:76`. Used for internal JWT signing and verification.
- **`test` service** — Same as `api` service (uses the Go backend's test target).

### `auth0_client_secret.txt`
Auth0 client secret for authentication. Get this from your Auth0 dashboard.

```bash
echo -n "your-auth0-client-secret" > auth0_client_secret.txt
```

**Used by:**
- **`api` service (Go backend)** — Retrieved by `config.Load()` in `core/internal/config/config.go:86`. Used for server-side Auth0 Management API calls (e.g., user lifecycle management).
- **`test` service** — Same as `api` service (uses the Go backend's test target).

**Note**: This is optional for the backend but recommended. The frontend uses Auth0 public client (SPA) which doesn't require a client secret.

### `auth0_mgmt_client_secret.txt`
Auth0 Management API (Machine-to-Machine) client secret for integration tests. This is used to dynamically create and delete test users. Get this from your Auth0 dashboard (create a M2M app authorized for the Management API with `create:users` and `delete:users` scopes).

```bash
echo -n "your-auth0-m2m-client-secret" > auth0_mgmt_client_secret.txt
```

**Used by:**
- **`integration-tests` service** — Read from `/run/secrets/auth0_mgmt_client_secret` in `tests/settings.py:28-33` when `INTEGRATION_TESTS_SECRETS_PROVIDER=docker`. Combined with `AUTH0_MGMT_CLIENT_ID` to obtain a Management API token for dynamic test user creation/deletion in `tests/conftest.py`.

### `integration_tests_auth0_email.txt` and `integration_tests_auth0_password.txt`
Credentials for the test user account used in Auth0 login flow integration tests.

```bash
echo -n "test@example.com" > integration_tests_auth0_email.txt
echo -n "changeme" > integration_tests_auth0_password.txt
```

**Used by:**
- **`integration-tests` service** — Read from `/run/secrets/integration_tests_auth0_email` and `/run/secrets/integration_tests_auth0_password` in `tests/settings.py:57-66` when `INTEGRATION_TESTS_SECRETS_PROVIDER=docker`. Used by the test suite to perform Auth0 login flow tests.

## Security Notes

- **Never commit these files to version control** - they are ignored by `.gitignore`
- Use strong, randomly generated values for production
- Rotate secrets periodically
- In production, consider using a proper secrets manager (AWS Secrets Manager, HashiCorp Vault, etc.)
