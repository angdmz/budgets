import { useAuth0 } from '@auth0/auth0-react';
import { Outlet, useLocation } from 'react-router-dom';

export default function ProtectedRoute() {
  const { isAuthenticated, isLoading, loginWithRedirect } = useAuth0();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Authenticating...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    loginWithRedirect({
      appState: { returnTo: location.pathname },
    });
    return null;
  }

  return <Outlet />;
}
