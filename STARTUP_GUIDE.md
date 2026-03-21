# Budget Management System - Startup Guide

## Quick Start (Full Stack)

### Prerequisites
- Docker and Docker Compose installed
- Auth0 account (free tier works)

### Step 1: Configure Auth0

1. **Create Auth0 Application** (Single Page Application)
   - Go to https://manage.auth0.com
   - Create new application → Single Page Application
   - Name: "Budget Manager"
   - Note your:
     - Domain: `your-tenant.auth0.com`
     - Client ID: `abc123...`

2. **Configure Application Settings**
   - Allowed Callback URLs: `http://localhost/app, http://localhost/admin`
   - Allowed Logout URLs: `http://localhost/app, http://localhost/admin, http://localhost`
   - Allowed Web Origins: `http://localhost`
   - Save changes

3. **Create Auth0 API**
   - Go to Applications → APIs → Create API
   - Name: "Budget Manager API"
   - Identifier: `https://api.budget.local`
   - Signing Algorithm: RS256
   - Enable RBAC: Yes
   - Add Permissions in Token: Yes

### Step 2: Update Environment Variables

Edit `.env` file:

```bash
# Auth0 Configuration
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.budget.local
AUTH0_CLIENT_ID=your-client-id-from-auth0
```

### Step 3: Verify Secrets

Check that all secret files exist:

```bash
ls -la secrets/
# Should show:
# - db_password.txt
# - encryption_key.txt
# - jwt_secret.txt
# - auth0_client_secret.txt
```

If missing, generate them:

```bash
docker-compose --profile setup up
```

### Step 4: Start All Services

```bash
# Build and start everything
docker-compose up --build

# Or run in background
docker-compose up --build -d
```

This will start:
- PostgreSQL database (port 5432)
- Database migrations
- Go API backend (internal port 8080)
- Next.js landing page (internal port 3000)
- React main app (internal port 3001)
- React admin app (internal port 3002)
- Nginx gateway (port 80)

### Step 5: Access the System

Open your browser:

- **Landing Page**: http://localhost/
- **Main Application**: http://localhost/app
- **Admin Panel**: http://localhost/admin
- **API Documentation**: http://localhost/swagger/index.html
- **Health Check**: http://localhost/health

### Step 6: First Login

1. Go to http://localhost/app
2. Click "Get Started" or navigate to the app
3. You'll be redirected to Auth0 login
4. Sign up with Google or email
5. After authentication, you'll be redirected back to the app

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         Nginx (Port 80)                      │
│  Routes:                                                     │
│  /          → Landing Page (Next.js)                        │
│  /app       → Main Application (React)                      │
│  /admin     → Admin Panel (React)                           │
│  /api       → Backend API (Go)                              │
│  /swagger   → API Documentation                             │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
┌───────▼────────┐   ┌────────▼────────┐   ┌───────▼────────┐
│  Landing Page  │   │   Main App      │   │   Admin App    │
│   (Next.js)    │   │   (React)       │   │   (React)      │
│   Port 3000    │   │   Port 3001     │   │   Port 3002    │
└────────────────┘   └─────────────────┘   └────────────────┘
                              │
                     ┌────────▼────────┐
                     │   API Backend   │
                     │      (Go)       │
                     │   Port 8080     │
                     └────────┬────────┘
                              │
                     ┌────────▼────────┐
                     │   PostgreSQL    │
                     │   Port 5432     │
                     └─────────────────┘
```

## Development Mode

To run services individually for development:

### Backend Only
```bash
docker-compose up db migrations api
```

### Frontend Development (without Docker)

**Landing Page:**
```bash
cd landing
npm install
npm run dev
# Access at http://localhost:3000
```

**Main App:**
```bash
cd app
npm install

# Create .env.local
cat > .env.local << EOF
VITE_AUTH0_DOMAIN=your-tenant.auth0.com
VITE_AUTH0_CLIENT_ID=your-client-id
VITE_AUTH0_AUDIENCE=https://api.budget.local
VITE_API_URL=http://localhost:8080/api/v1
EOF

npm run dev
# Access at http://localhost:3001
```

**Admin App:**
```bash
cd admin
npm install
npm run dev
# Access at http://localhost:3002
```

## Troubleshooting

### Issue: "Failed to get encryption key"

**Solution**: Ensure secrets are generated:
```bash
docker-compose --profile setup up
```

### Issue: "Invalid token" or Auth errors

**Solution**: 
1. Check Auth0 configuration matches `.env`
2. Verify Auth0 callback URLs include your domain
3. Check browser console for specific Auth0 errors

### Issue: "Connection refused" to database

**Solution**: 
1. Ensure database is running: `docker-compose ps`
2. Check database logs: `docker-compose logs db`
3. Restart services: `docker-compose restart`

### Issue: Nginx returns 502 Bad Gateway

**Solution**:
1. Check if all services are running: `docker-compose ps`
2. Check service logs: `docker-compose logs [service-name]`
3. Ensure services are on the same network

### Issue: CORS errors in browser

**Solution**: CORS is configured in Nginx. If you see CORS errors:
1. Check Nginx configuration in `nginx/nginx.conf`
2. Restart Nginx: `docker-compose restart nginx`

## Useful Commands

```bash
# View all running services
docker-compose ps

# View logs for specific service
docker-compose logs -f api
docker-compose logs -f landing
docker-compose logs -f app

# Restart a service
docker-compose restart api

# Rebuild a specific service
docker-compose up --build api

# Stop all services
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v

# Run tests
docker-compose --profile test up test

# Generate new secrets
docker-compose --profile setup up
```

## API Endpoints

All API endpoints require Authentication (Bearer token from Auth0).

### Groups
- `POST /api/v1/groups` - Create group
- `GET /api/v1/groups` - List groups
- `GET /api/v1/groups/:id` - Get group
- `PUT /api/v1/groups/:id` - Update group
- `DELETE /api/v1/groups/:id` - Delete group

### Budgets
- `POST /api/v1/groups/:id/budgets` - Create budget
- `GET /api/v1/groups/:id/budgets` - List budgets
- `GET /api/v1/budgets/:id` - Get budget
- `PUT /api/v1/budgets/:id` - Update budget
- `DELETE /api/v1/budgets/:id` - Delete budget

### Categories
- `POST /api/v1/groups/:id/categories` - Create category
- `GET /api/v1/groups/:id/categories` - List categories
- `PUT /api/v1/categories/:id` - Update category
- `DELETE /api/v1/categories/:id` - Delete category

### Expenses
- `POST /api/v1/budgets/:id/expected-expenses` - Create expected expense
- `GET /api/v1/budgets/:id/expected-expenses` - List expected expenses
- `POST /api/v1/budgets/:id/actual-expenses` - Create actual expense
- `GET /api/v1/budgets/:id/actual-expenses` - List actual expenses
- `PUT /api/v1/expected-expenses/:id` - Update expected expense
- `PUT /api/v1/actual-expenses/:id` - Update actual expense
- `DELETE /api/v1/expected-expenses/:id` - Delete expected expense
- `DELETE /api/v1/actual-expenses/:id` - Delete actual expense

## Security Features

✅ **Auth0 Integration** - Enterprise-grade authentication
✅ **JWT Validation** - RS256 signed tokens with JWKS
✅ **Encrypted Data** - All monetary values encrypted at rest (Fernet)
✅ **Group Isolation** - Users can only access their groups
✅ **Safe Error Handling** - No internal errors exposed in production
✅ **Required Secrets** - No default passwords or keys
✅ **HTTPS Ready** - Configure SSL in Nginx for production

## Production Deployment

For production deployment:

1. **Use HTTPS**: Configure SSL certificates in Nginx
2. **Update Auth0**: Add production URLs to Auth0 configuration
3. **Environment Variables**: Set `SERVER_ENV=production`
4. **Secrets Management**: Use AWS Secrets Manager or similar
5. **Database**: Use managed PostgreSQL (AWS RDS, etc.)
6. **Monitoring**: Add logging and monitoring solutions

## Next Steps

1. ✅ System is running
2. Create your first group
3. Add categories for your expenses
4. Create a budget for the current month
5. Start tracking expenses!

## Support

For issues or questions:
- Check the logs: `docker-compose logs -f`
- Review the API documentation: http://localhost/swagger/index.html
- Check Auth0 dashboard for authentication issues

## File Structure

```
budgets/
├── core/                    # Go backend
│   ├── internal/
│   │   ├── handler/        # HTTP handlers
│   │   ├── service/        # Business logic
│   │   ├── repository/     # Data access
│   │   ├── middleware/     # Auth0 middleware
│   │   └── domain/         # Domain models
│   └── main.go
├── migrations/              # Python database migrations
├── landing/                 # Next.js landing page
├── app/                     # React main application
├── admin/                   # React admin panel
├── nginx/                   # Nginx gateway
├── secrets/                 # Secret files (gitignored)
├── docker-compose.yml       # Service orchestration
└── .env                     # Environment variables
```

---

**You're all set! 🎉**

The Budget Management System is now running with full Auth0 integration, encrypted data storage, and a complete frontend stack.
