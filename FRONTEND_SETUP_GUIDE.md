# Frontend Applications Setup Guide

## Overview

This guide covers the three frontend applications that have been scaffolded:

1. **Landing Page** (Next.js) - SEO-optimized marketing site
2. **Main Application** (React + Vite) - User-facing budget management app
3. **Admin Application** (React + Vite) - Internal admin panel

## Current Status

### ✅ Completed (Backend)
- Auth0 integration in backend
- JWT validation with JWKS
- All API endpoints secured with Auth0
- Nginx gateway configuration
- Docker orchestration setup

### ✅ Completed (Landing Page)
- Next.js 14 with App Router
- SEO optimization (metadata, sitemap, robots.txt)
- Responsive design with Tailwind CSS
- Hero, Features, Benefits, CTA sections
- Dockerfile ready

### 🚧 In Progress (Main App & Admin)
- Basic scaffolding created
- Package.json configured
- Vite + React + TypeScript setup
- Tailwind CSS configured
- **Needs**: Component implementation

## Next Steps to Complete

### 1. Complete Main Application (8-10 hours)

**Directory**: `/app`

**Required Components**:
```
app/src/
├── main.tsx                 # Entry point with Auth0Provider
├── App.tsx                  # Main app with routing
├── index.css                # Global styles
├── lib/
│   ├── api.ts              # Axios instance with Auth0 token
│   └── types.ts            # TypeScript types
├── components/
│   ├── Layout.tsx          # Main layout with navigation
│   ├── ProtectedRoute.tsx  # Auth guard
│   └── ui/                 # Reusable UI components
├── pages/
│   ├── Dashboard.tsx       # Main dashboard with charts
│   ├── Groups.tsx          # Groups CRUD
│   ├── Budgets.tsx         # Budgets CRUD
│   ├── Categories.tsx      # Categories CRUD
│   └── Expenses.tsx        # Expenses CRUD
└── hooks/
    ├── useAuth.ts          # Auth0 hook wrapper
    └── useApi.ts           # React Query hooks
```

**Key Features to Implement**:
- Auth0 login/logout
- Dashboard with Recharts (expected vs actual)
- Groups management
- Budget creation and tracking
- Category management
- Expense tracking (expected + actual)
- Filters (weekly, monthly, custom date range)

### 2. Complete Admin Application (4-6 hours)

**Directory**: `/admin`

**Required Components**:
```
admin/src/
├── main.tsx
├── App.tsx
├── pages/
│   ├── Users.tsx           # User management
│   ├── Groups.tsx          # Group inspection
│   ├── Budgets.tsx         # Budget overview
│   └── Expenses.tsx        # Expense inspection
└── components/
    ├── AdminLayout.tsx
    └── DataTable.tsx       # Reusable table component
```

**Key Features**:
- Admin-only access (check Auth0 roles)
- User management
- Group inspection
- Soft delete functionality
- System-wide statistics

### 3. Docker Compose Integration

Update `docker-compose.yml` to include:

```yaml
services:
  # ... existing services ...

  landing:
    build:
      context: ./landing
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
    networks:
      - budget_network

  app:
    build:
      context: ./app
      dockerfile: Dockerfile
    ports:
      - "3001:3001"
    environment:
      - VITE_AUTH0_DOMAIN=${AUTH0_DOMAIN}
      - VITE_AUTH0_CLIENT_ID=${AUTH0_CLIENT_ID}
      - VITE_AUTH0_AUDIENCE=${AUTH0_AUDIENCE}
      - VITE_API_URL=http://localhost/api
    networks:
      - budget_network

  admin:
    build:
      context: ./admin
      dockerfile: Dockerfile
    ports:
      - "3002:3002"
    environment:
      - VITE_AUTH0_DOMAIN=${AUTH0_DOMAIN}
      - VITE_AUTH0_CLIENT_ID=${AUTH0_CLIENT_ID}
      - VITE_AUTH0_AUDIENCE=${AUTH0_AUDIENCE}
      - VITE_API_URL=http://localhost/api
    networks:
      - budget_network

  nginx:
    build:
      context: ./nginx
      dockerfile: Dockerfile
    ports:
      - "80:80"
    depends_on:
      - api
      - landing
      - app
      - admin
    networks:
      - budget_network
```

## Auth0 Configuration Required

### Backend Environment Variables
```bash
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.budget.local
AUTH0_CLIENT_ID=your-backend-client-id
```

### Frontend Environment Variables
```bash
VITE_AUTH0_DOMAIN=your-tenant.auth0.com
VITE_AUTH0_CLIENT_ID=your-frontend-client-id
VITE_AUTH0_AUDIENCE=https://api.budget.local
VITE_API_URL=http://localhost/api
```

### Auth0 Application Setup

1. **Create Auth0 Application** (SPA)
   - Type: Single Page Application
   - Allowed Callback URLs: `http://localhost/app/callback`
   - Allowed Logout URLs: `http://localhost/app`
   - Allowed Web Origins: `http://localhost`

2. **Create Auth0 API**
   - Identifier: `https://api.budget.local`
   - Enable RBAC
   - Add Permissions:
     - `read:budgets`
     - `write:budgets`
     - `admin:all` (for admin app)

3. **Configure Rules/Actions**
   - Add custom claims to JWT
   - Map user roles to permissions

## Running the Full Stack

### Development Mode
```bash
# Terminal 1 - Backend
docker-compose up db migrations api

# Terminal 2 - Landing
cd landing && npm install && npm run dev

# Terminal 3 - Main App
cd app && npm install && npm run dev

# Terminal 4 - Admin
cd admin && npm install && npm run dev
```

### Production Mode
```bash
# Build and start all services
docker-compose up --build

# Access at:
# - Landing: http://localhost/
# - App: http://localhost/app
# - Admin: http://localhost/admin
# - API: http://localhost/api
# - Swagger: http://localhost/swagger
```

## Implementation Checklist

### Landing Page ✅
- [x] Next.js setup
- [x] SEO optimization
- [x] Responsive design
- [x] Navigation
- [x] Hero section
- [x] Features section
- [x] Benefits section
- [x] CTA section
- [x] Footer
- [x] Dockerfile

### Main Application 🚧
- [x] Vite + React setup
- [x] TypeScript configuration
- [x] Tailwind CSS
- [ ] Auth0 integration
- [ ] API client setup
- [ ] React Query setup
- [ ] Routing (React Router)
- [ ] Layout component
- [ ] Dashboard page
- [ ] Groups CRUD
- [ ] Budgets CRUD
- [ ] Categories CRUD
- [ ] Expenses CRUD
- [ ] Charts (Recharts)
- [ ] Dockerfile

### Admin Application 🚧
- [ ] Vite + React setup
- [ ] Auth0 with role check
- [ ] Admin layout
- [ ] User management
- [ ] Group inspection
- [ ] Budget overview
- [ ] Expense inspection
- [ ] Soft delete UI
- [ ] Dockerfile

### Infrastructure ✅
- [x] Nginx configuration
- [x] Docker compose structure
- [ ] Environment variables setup
- [ ] Service dependencies
- [ ] Health checks

## Estimated Time to Complete

- **Main Application**: 8-10 hours
- **Admin Application**: 4-6 hours
- **Integration & Testing**: 2-3 hours
- **Total**: 14-19 hours

## Current File Structure

```
budgets/
├── core/                    # Go backend ✅
├── migrations/              # Python migrations ✅
├── nginx/                   # Nginx gateway ✅
├── landing/                 # Next.js landing ✅
│   ├── src/
│   │   ├── app/
│   │   └── components/
│   ├── public/
│   ├── Dockerfile
│   └── package.json
├── app/                     # React main app 🚧
│   ├── src/                 # Needs implementation
│   ├── Dockerfile           # Needs creation
│   └── package.json
├── admin/                   # React admin 🚧
│   └── (needs creation)
├── docker-compose.yml       # Needs update
└── secrets/
    ├── db_password.txt
    ├── encryption_key.txt
    ├── jwt_secret.txt
    └── auth0_client_secret.txt
```

## Notes

- Backend is fully functional with Auth0
- Landing page is production-ready
- Main app and admin need component implementation
- All infrastructure is configured
- Focus next on implementing React components for main app
