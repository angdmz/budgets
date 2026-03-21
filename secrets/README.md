# Secrets Directory

This directory contains Docker secrets files. These files are mounted into containers at `/run/secrets/`.

## Required Files

Create the following files with your secret values (one value per file, no trailing newline):

### `db_password.txt`
Database password for PostgreSQL.

```bash
echo -n "your-secure-db-password" > db_password.txt
```

### `encryption_key.txt`
Fernet encryption key for encrypting sensitive data. Generate with:

```bash
docker run --rm python:3.13-slim sh -c "pip install -q cryptography && python -c \"from cryptography.fernet import Fernet; print(Fernet.generate_key().decode(), end='')\"" > encryption_key.txt
```

### `jwt_secret.txt`
Secret key for signing JWT tokens.

```bash
echo -n "your-jwt-secret-at-least-32-chars" > jwt_secret.txt
```

### `auth0_client_secret.txt`
Auth0 client secret for authentication. Get this from your Auth0 dashboard.

```bash
echo -n "your-auth0-client-secret" > auth0_client_secret.txt
```

**Note**: This is optional for the backend but recommended. The frontend uses Auth0 public client (SPA) which doesn't require a client secret.

## Security Notes

- **Never commit these files to version control** - they are ignored by `.gitignore`
- Use strong, randomly generated values for production
- Rotate secrets periodically
- In production, consider using a proper secrets manager (AWS Secrets Manager, HashiCorp Vault, etc.)
