# Nginx Gateway

## Overview

Nginx serves as the reverse proxy gateway for the Budget Management System. It routes traffic to the landing page, main app, admin panel, API backend, and Swagger documentation. In production, it handles TLS termination and HTTP-to-HTTPS redirection.

## Configuration Files

- **`nginx.dev.conf`** — Development configuration (HTTP only, `localhost`). Used when `NGINX_CONF=./nginx/nginx.dev.conf`.
- **`nginx.conf`** — Production configuration (HTTP redirect + TLS on port 443, `miplatita.ar`). Used when `NGINX_CONF=./nginx/nginx.conf`.

## Environment Variables

Nginx itself does not read environment variables at runtime. However, the following variables in the root `.env` file control how the Nginx container is configured in `docker-compose.yml`:

| Variable | Default | Used In | Purpose |
|----------|---------|---------|---------|
| `NGINX_CONF` | `./nginx/nginx.dev.conf` | `docker-compose.yml:157` → `volumes` | Path to the Nginx config file mounted into the container at `/etc/nginx/nginx.conf`. Use `./nginx/nginx.dev.conf` for development (HTTP only) or `./nginx/nginx.conf` for production (TLS). |
| `TLS_CERT_DIR` | `./nginx/ssl` | `docker-compose.yml:159` → `volumes` | Directory containing TLS certificates (`fullchain.pem` and `privkey.pem`), mounted at `/etc/nginx/ssl` inside the container. Only used in production config (`nginx.conf`). |

### How TLS Certificates Are Used

In the production config (`nginx.conf:76-77`), Nginx loads certificates from the mounted directory:
```
ssl_certificate /etc/nginx/ssl/fullchain.pem;
ssl_certificate_key /etc/nginx/ssl/privkey.pem;
```

The `TLS_CERT_DIR` env var controls which host directory is mounted to `/etc/nginx/ssl`. For production, set it to your Let's Encrypt or custom certificate directory (e.g., `/etc/letsencrypt/live/miplatita.ar`).

## Routing

### Development (`nginx.dev.conf`)

All traffic on port 80:

| Path | Upstream | Port |
|------|---------|------|
| `/` | `landing_backend` (Next.js) | 3000 |
| `/app` | `app_backend` (React) | 3001 |
| `/admin` | `admin_backend` (React) | 3002 |
| `/api` | `api_backend` (Go) | 8080 |
| `/swagger` | `api_backend` (Go) | 8080 |
| `/health` | `api_backend` (Go) | 8080 |

### Production (`nginx.conf`)

- Port 80: Redirects all traffic to HTTPS (except `/health` which is proxied directly).
- Port 443: TLS termination with same routing as development, plus security headers (HSTS, X-Frame-Options, etc.).

## Docker

```bash
# Build
docker build -t budget-nginx .

# Run (development)
docker run -p 8000:80 -v ./nginx/nginx.dev.conf:/etc/nginx/nginx.conf:ro budget-nginx

# Run (production with TLS)
docker run -p 8000:80 -p 8443:443 \
  -v ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro \
  -v /etc/letsencrypt/live/miplatita.ar:/etc/nginx/ssl:ro \
  budget-nginx
```

In Docker Compose, ports 8000 (host) → 80 (container) and 8443 (host) → 443 (container) are mapped. EC2 iptables NAT forwards :80→:8000 and :443→:8443 to the rootless Podman container.

## License

See LICENSE file in repository root.
