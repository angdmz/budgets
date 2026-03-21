# Budget Management System - Admin Application

## Overview

Admin panel for managing users, groups, and system-wide operations. Built with React and Auth0 with admin-level permissions.

## Technology Stack

- **Framework**: React 18 + Vite
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Authentication**: Auth0 React SDK (with admin scope)
- **State Management**: React Query
- **HTTP Client**: Axios
- **Routing**: React Router v6

## Features

### Current Implementation
- Auth0 authentication with admin scope
- Basic scaffolding and layout
- Ready for feature expansion

### Planned Features
- **User Management**: View, edit, deactivate users
- **Group Inspection**: View all groups across system
- **Budget Overview**: System-wide budget statistics
- **Expense Inspection**: View all expenses
- **Soft Delete Management**: Restore deleted items
- **System Statistics**: Dashboard with metrics
- **Audit Logs**: Track system changes

## Setup & Development

### Prerequisites
- Node.js 20 or higher
- npm or yarn
- Auth0 account with admin role configured

### Installation

```bash
# Install dependencies
npm install
```

### Environment Variables

Create `.env.local` file:

```bash
# Auth0 Configuration
VITE_AUTH0_DOMAIN=your-tenant.auth0.com
VITE_AUTH0_CLIENT_ID=your-spa-client-id
VITE_AUTH0_AUDIENCE=https://api.budget.local

# API Configuration
VITE_API_URL=http://localhost:8080/api/v1
```

### Auth0 Configuration

**1. Create Admin Role in Auth0:**
- Go to User Management → Roles
- Create role: "admin"
- Add permissions: `admin:all`

**2. Assign Role to Users:**
- Go to User Management → Users
- Select user → Roles → Assign Role

**3. Configure Application:**
- Same SPA application as main app
- Add scope: `admin:all`
- Allowed Callback URLs: `http://localhost:3002/admin`
- Allowed Logout URLs: `http://localhost:3002/admin`

### Development Mode

```bash
# Start development server
npm run dev
```

Access at: `http://localhost:3002/admin`

### Build for Production

```bash
# Create production build
npm run build

# Preview production build
npm run preview
```

### Docker Build

```bash
# Build Docker image
docker build -t budget-admin .

# Run container
docker run -p 3002:3002 \
  -e VITE_AUTH0_DOMAIN=your-tenant.auth0.com \
  -e VITE_AUTH0_CLIENT_ID=your-client-id \
  -e VITE_AUTH0_AUDIENCE=https://api.budget.local \
  budget-admin
```

## Project Structure

```
admin/
├── src/
│   ├── main.tsx              # Entry point with Auth0Provider
│   ├── index.css             # Global styles
│   └── (to be implemented)
│       ├── components/
│       │   ├── AdminLayout.tsx
│       │   ├── DataTable.tsx
│       │   └── RoleGuard.tsx
│       ├── pages/
│       │   ├── Dashboard.tsx
│       │   ├── Users.tsx
│       │   ├── Groups.tsx
│       │   └── Expenses.tsx
│       └── lib/
│           ├── api.ts
│           └── types.ts
├── public/                   # Static assets
├── index.html                # HTML template
├── vite.config.ts            # Vite configuration
├── tailwind.config.js        # Tailwind configuration
├── tsconfig.json             # TypeScript configuration
├── package.json              # Dependencies
└── Dockerfile                # Production build
```

## Implementation Guide

### Adding Admin Features

**1. Create Admin Layout:**

```typescript
// src/components/AdminLayout.tsx
import { useAuth0 } from '@auth0/auth0-react';

export function AdminLayout() {
  const { user, logout } = useAuth0();
  
  // Check if user has admin role
  const isAdmin = user?.['https://api.budget.local/roles']?.includes('admin');
  
  if (!isAdmin) {
    return <div>Access Denied</div>;
  }
  
  return (
    <div>
      {/* Admin navigation and content */}
    </div>
  );
}
```

**2. Create User Management Page:**

```typescript
// src/pages/Users.tsx
import { useQuery } from '@tanstack/react-query';

export default function Users() {
  const { data: users } = useQuery({
    queryKey: ['admin', 'users'],
    queryFn: async () => {
      // Fetch all users (admin endpoint)
      const response = await api.get('/admin/users');
      return response.data;
    },
  });
  
  return (
    <div>
      <h1>User Management</h1>
      {/* User table */}
    </div>
  );
}
```

**3. Add Role Guard:**

```typescript
// src/components/RoleGuard.tsx
export function RoleGuard({ children, requiredRole }) {
  const { user } = useAuth0();
  const roles = user?.['https://api.budget.local/roles'] || [];
  
  if (!roles.includes(requiredRole)) {
    return <Navigate to="/unauthorized" />;
  }
  
  return children;
}
```

### Recommended Features to Implement

**1. User Management**
- List all users
- View user details
- Deactivate/activate users
- View user's groups and budgets
- Audit user activity

**2. Group Management**
- List all groups
- View group details
- View group members
- View group budgets
- Delete groups (soft delete)

**3. System Dashboard**
- Total users count
- Total groups count
- Total budgets count
- Total expenses (expected vs actual)
- Recent activity feed
- System health metrics

**4. Soft Delete Management**
- List deleted items
- Restore deleted items
- Permanently delete items
- Bulk operations

**5. Audit Logs**
- Track all system changes
- Filter by user, action, date
- Export audit logs
- View detailed change history

## API Endpoints (To be implemented)

### Admin-Only Endpoints

```
GET  /api/v1/admin/users              # List all users
GET  /api/v1/admin/users/:id          # Get user details
PUT  /api/v1/admin/users/:id          # Update user
POST /api/v1/admin/users/:id/deactivate

GET  /api/v1/admin/groups             # List all groups
GET  /api/v1/admin/groups/:id         # Get group details
GET  /api/v1/admin/groups/:id/members

GET  /api/v1/admin/budgets            # List all budgets
GET  /api/v1/admin/expenses           # List all expenses

GET  /api/v1/admin/deleted            # List soft-deleted items
POST /api/v1/admin/deleted/:id/restore

GET  /api/v1/admin/audit-logs         # Get audit logs
GET  /api/v1/admin/stats              # System statistics
```

## Security

### Role-Based Access Control

**Auth0 Configuration:**
```json
{
  "roles": ["admin"],
  "permissions": ["admin:all"]
}
```

**JWT Token Claims:**
```json
{
  "sub": "auth0|123456",
  "https://api.budget.local/roles": ["admin"],
  "https://api.budget.local/permissions": ["admin:all"]
}
```

**Backend Validation:**
```go
// Middleware to check admin role
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        user := GetAuth0UserFromContext(c)
        if !user.HasRole("admin") {
            c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
            return
        }
        c.Next()
    }
}
```

### Security Best Practices
- ✅ Admin routes protected by role check
- ✅ Tokens validated on every request
- ✅ No admin credentials in client code
- ✅ Audit logging for all admin actions
- ✅ Rate limiting on admin endpoints
- ✅ HTTPS enforced in production

## Customization

### Theming

Update `tailwind.config.js`:

```javascript
theme: {
  extend: {
    colors: {
      admin: {
        primary: '#1e40af',
        secondary: '#64748b',
      },
    },
  },
}
```

### Adding New Admin Pages

1. Create page component
2. Add route with RoleGuard
3. Add navigation link
4. Implement API calls
5. Add error handling

## Testing

### Manual Testing
- [ ] Login with admin user
- [ ] Login with non-admin user (should be denied)
- [ ] All admin features work
- [ ] Non-admin cannot access admin routes
- [ ] Audit logs record actions
- [ ] Error handling works

### Automated Testing (To be added)
```bash
npm install --save-dev vitest @testing-library/react
```

## Troubleshooting

### Common Issues

**1. "Access Denied" for admin user**
- Check Auth0 role assignment
- Verify JWT token includes admin role
- Check role claim namespace
- Verify backend validates role correctly

**2. Admin routes not protected**
- Ensure RoleGuard is implemented
- Check Auth0 token includes roles
- Verify middleware on backend

**3. Cannot see admin features**
- Check user has admin role in Auth0
- Verify token refresh
- Check browser console for errors

## Deployment

### Production Build

```bash
npm run build
```

### Docker Deployment

```bash
docker build -t budget-admin:latest .
docker run -p 3002:3002 budget-admin:latest
```

### Environment Variables for Production

```bash
VITE_AUTH0_DOMAIN=your-tenant.auth0.com
VITE_AUTH0_CLIENT_ID=your-production-client-id
VITE_AUTH0_AUDIENCE=https://api.budget.com
VITE_API_URL=https://api.budget.com/api/v1
```

### Deployment Checklist
- [ ] Update Auth0 with production URLs
- [ ] Configure admin roles in Auth0
- [ ] Set production environment variables
- [ ] Enable HTTPS
- [ ] Test admin access
- [ ] Verify role-based access control
- [ ] Test all admin features
- [ ] Set up monitoring
- [ ] Configure audit logging

## Roadmap

### Phase 1 (Current)
- ✅ Basic scaffolding
- ✅ Auth0 integration
- ✅ Docker configuration

### Phase 2 (Next)
- [ ] Admin layout
- [ ] User management
- [ ] Group inspection
- [ ] System dashboard

### Phase 3 (Future)
- [ ] Soft delete management
- [ ] Audit logs
- [ ] Advanced analytics
- [ ] Bulk operations
- [ ] Export functionality

## Contributing

### Code Style
- Use TypeScript
- Follow React best practices
- Implement proper error handling
- Add loading states
- Write meaningful comments

### Adding Features
1. Plan the feature
2. Design the UI
3. Implement API calls
4. Add error handling
5. Test thoroughly
6. Update documentation

## License

See LICENSE file in repository root.

## Support

For issues or questions:
- Check Auth0 role configuration
- Verify admin permissions
- Review API logs
- Check browser console for errors
