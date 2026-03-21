# Budget Management System - Main Application

## Overview

Full-featured React application for budget management with Auth0 authentication, real-time data visualization, and complete CRUD operations for groups, budgets, categories, and expenses.

## Technology Stack

- **Framework**: React 18 + Vite
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: React Query + Zustand
- **Authentication**: Auth0 React SDK
- **Charts**: Recharts
- **HTTP Client**: Axios
- **Routing**: React Router v6

## Features

### Authentication
- Auth0 integration with automatic token management
- Protected routes with auth guards
- Automatic redirect to login
- Token refresh handling
- Logout functionality

### Dashboard
- Budget selector dropdown
- Summary cards (Expected, Actual, Difference)
- Bar chart visualization (Expected vs Actual)
- Recent expenses table
- Real-time data updates

### CRUD Operations
- **Groups**: Create, list, update, delete
- **Budgets**: Create with date ranges, list, update, delete
- **Categories**: Create with colors, list, update, delete
- **Expected Expenses**: Create, list, update, delete
- **Actual Expenses**: Create with dates, list, update, delete

### User Experience
- Responsive design (mobile, tablet, desktop)
- Loading states
- Error handling
- Form validation
- Modal dialogs
- Toast notifications (can be added)

## Setup & Development

### Prerequisites
- Node.js 20 or higher
- npm or yarn
- Auth0 account

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

**Getting Auth0 Credentials:**

1. Go to https://manage.auth0.com
2. Create a new Application → Single Page Application
3. Copy Domain and Client ID
4. Configure:
   - Allowed Callback URLs: `http://localhost:3001/app`
   - Allowed Logout URLs: `http://localhost:3001/app`
   - Allowed Web Origins: `http://localhost:3001`

### Development Mode

```bash
# Start development server
npm run dev
```

Access at: `http://localhost:3001/app`

Features in development mode:
- Hot module replacement
- Fast refresh
- Source maps
- Error overlay
- React DevTools support

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
docker build -t budget-app .

# Run container
docker run -p 3001:3001 \
  -e VITE_AUTH0_DOMAIN=your-tenant.auth0.com \
  -e VITE_AUTH0_CLIENT_ID=your-client-id \
  -e VITE_AUTH0_AUDIENCE=https://api.budget.local \
  budget-app
```

## Project Structure

```
app/
├── src/
│   ├── lib/
│   │   ├── api.ts              # Axios instance with Auth0 token
│   │   └── types.ts            # TypeScript type definitions
│   ├── components/
│   │   ├── Layout.tsx          # Main layout with navigation
│   │   └── ProtectedRoute.tsx  # Auth guard component
│   ├── pages/
│   │   ├── Dashboard.tsx       # Dashboard with charts
│   │   ├── Groups.tsx          # Groups CRUD
│   │   ├── Budgets.tsx         # Budgets CRUD
│   │   ├── Categories.tsx      # Categories CRUD
│   │   └── Expenses.tsx        # Expenses CRUD
│   ├── App.tsx                 # Router configuration
│   ├── main.tsx                # Entry point with Auth0Provider
│   └── index.css               # Global styles
├── public/                     # Static assets
├── index.html                  # HTML template
├── vite.config.ts              # Vite configuration
├── tailwind.config.js          # Tailwind configuration
├── tsconfig.json               # TypeScript configuration
├── package.json                # Dependencies
└── Dockerfile                  # Production build
```

## Usage Guide

### First Time Setup

1. **Login**: Click "Get Started" or navigate to `/app`
2. **Auth0 Redirect**: You'll be redirected to Auth0 login
3. **Sign Up/Login**: Create account or login with existing credentials
4. **Callback**: After authentication, you'll be redirected back to the app

### Creating Your First Budget

1. **Create a Group**:
   - Navigate to "Groups"
   - Click "Add Group"
   - Enter name and description
   - Click "Create"

2. **Add Categories**:
   - Navigate to "Categories"
   - Select your group
   - Click "Add Category"
   - Choose name and color
   - Click "Create"

3. **Create a Budget**:
   - Navigate to "Budgets"
   - Select your group
   - Click "Add Budget"
   - Set name, start date, end date
   - Click "Create"

4. **Add Expected Expenses**:
   - Navigate to "Expenses"
   - Select your budget
   - Click "Add Expense"
   - Enter details
   - Click "Create"

5. **Track Actual Expenses**:
   - Same as expected expenses
   - Include expense date

6. **View Dashboard**:
   - Navigate to "Dashboard"
   - Select your budget
   - View charts and summaries

## API Integration

### Authentication Flow

```typescript
// Automatic token injection
const api = await createApiClient(getAccessTokenSilently);
const response = await api.get('/groups');
```

The API client automatically:
- Obtains Auth0 access token
- Adds Bearer token to requests
- Handles token refresh
- Manages token expiration

### Making API Calls

**Using React Query:**
```typescript
const { data: groups } = useQuery({
  queryKey: ['groups'],
  queryFn: async () => {
    const api = await createApiClient(getAccessTokenSilently);
    const response = await api.get<Group[]>('/groups');
    return response.data;
  },
});
```

**Using Mutations:**
```typescript
const createMutation = useMutation({
  mutationFn: async (data: CreateGroupRequest) => {
    const api = await createApiClient(getAccessTokenSilently);
    return api.post('/groups', data);
  },
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['groups'] });
  },
});
```

## State Management

### React Query (Server State)
- API data caching
- Automatic refetching
- Optimistic updates
- Background synchronization

### Zustand (Client State) - Optional
Can be added for:
- UI state (modals, filters)
- User preferences
- Temporary form data

## Customization

### Theming

Update colors in `tailwind.config.js`:

```javascript
theme: {
  extend: {
    colors: {
      primary: {
        500: '#your-color',
        600: '#your-color',
        // ...
      },
    },
  },
}
```

### Adding New Pages

1. Create component in `src/pages/`:
```typescript
// src/pages/Reports.tsx
export default function Reports() {
  return <div>Reports Page</div>;
}
```

2. Add route in `App.tsx`:
```typescript
<Route path="/reports" element={<Reports />} />
```

3. Add navigation link in `Layout.tsx`:
```typescript
{ name: 'Reports', href: '/reports', icon: '📊' }
```

### Adding New Features

**Example: Adding a filter**

```typescript
const [filter, setFilter] = useState('all');

const filteredExpenses = expenses?.filter(expense => {
  if (filter === 'all') return true;
  return expense.category_id === filter;
});
```

## Performance Optimization

### Current Optimizations
- Code splitting with React Router
- Lazy loading of routes
- React Query caching
- Memoization of expensive calculations
- Debounced search inputs

### Recommendations
- Add virtual scrolling for large lists
- Implement pagination
- Add service worker for offline support
- Optimize bundle size with tree shaking

## Testing

### Unit Tests (To be added)
```bash
npm install --save-dev vitest @testing-library/react
```

### E2E Tests (To be added)
```bash
npm install --save-dev playwright
```

### Manual Testing Checklist
- [ ] Login/logout flow works
- [ ] All CRUD operations work
- [ ] Dashboard displays correct data
- [ ] Charts render properly
- [ ] Forms validate correctly
- [ ] Error states display
- [ ] Loading states show
- [ ] Mobile responsive
- [ ] Token refresh works
- [ ] Logout clears state

## Troubleshooting

### Common Issues

**1. "Failed to fetch" errors**
- Check API is running
- Verify `VITE_API_URL` is correct
- Check CORS configuration
- Verify Auth0 token is valid

**2. Auth0 redirect loop**
- Check callback URLs in Auth0 dashboard
- Verify `VITE_AUTH0_DOMAIN` is correct
- Clear browser cookies/cache
- Check browser console for errors

**3. "Unauthorized" errors**
- Token may be expired (refresh page)
- Check Auth0 audience matches API
- Verify API is configured for Auth0

**4. Charts not rendering**
- Check data format matches Recharts requirements
- Verify budget has expenses
- Check browser console for errors

**5. Build fails**
- Clear node_modules and reinstall
- Check TypeScript errors
- Verify all environment variables are set

### Debug Mode

Enable React Query DevTools:
```typescript
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

<QueryClientProvider client={queryClient}>
  <App />
  <ReactQueryDevtools initialIsOpen={false} />
</QueryClientProvider>
```

## Security

### Best Practices Implemented
- Auth0 for authentication
- Tokens stored securely (not in localStorage)
- HTTPS enforced in production
- CORS configured properly
- No sensitive data in client code
- Environment variables for configuration

### Security Checklist
- [ ] Auth0 configured correctly
- [ ] Callback URLs restricted
- [ ] HTTPS enabled in production
- [ ] No API keys in code
- [ ] Content Security Policy configured
- [ ] XSS protection enabled

## Deployment

### Production Build

```bash
npm run build
```

Output in `dist/` directory.

### Docker Deployment

```bash
docker build -t budget-app:latest .
docker run -p 3001:3001 budget-app:latest
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
- [ ] Set production environment variables
- [ ] Enable HTTPS
- [ ] Configure CDN for assets
- [ ] Set up error tracking (Sentry)
- [ ] Configure analytics
- [ ] Test on production domain
- [ ] Verify API connectivity
- [ ] Check mobile responsiveness
- [ ] Run Lighthouse audit

## Contributing

### Code Style
- Use TypeScript for type safety
- Follow React best practices
- Use functional components with hooks
- Keep components small and focused
- Write meaningful variable names
- Add comments for complex logic

### Pull Request Process
1. Create feature branch
2. Make changes
3. Test thoroughly
4. Update documentation
5. Submit PR with description

## License

See LICENSE file in repository root.

## Support

For issues or questions:
- Check browser console for errors
- Review Auth0 dashboard for authentication issues
- Check API logs for backend errors
- Verify environment variables are set correctly
