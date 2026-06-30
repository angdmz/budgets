import React from 'react';
import ReactDOM from 'react-dom/client';
import { Auth0Provider } from '@auth0/auth0-react';
import { BrowserRouter, useNavigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import App from './App';
import './index.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

const auth0Domain = import.meta.env.VITE_AUTH0_DOMAIN;
const auth0ClientId = import.meta.env.VITE_AUTH0_CLIENT_ID;
const auth0Audience = import.meta.env.VITE_AUTH0_AUDIENCE;

function Auth0ProviderWithNavigate({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  return (
    <Auth0Provider
      domain={auth0Domain}
      clientId={auth0ClientId}
      authorizationParams={{
        redirect_uri: window.location.origin + '/app',
        audience: auth0Audience,
        scope: 'openid profile email',
      }}
      useRefreshTokens={true}
      cacheLocation="localstorage"
      onRedirectCallback={(appState) => {
        navigate(appState?.returnTo || '/', { replace: true });
      }}
    >
      {children}
    </Auth0Provider>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter basename="/app">
      <Auth0ProviderWithNavigate>
        <QueryClientProvider client={queryClient}>
          <App />
        </QueryClientProvider>
      </Auth0ProviderWithNavigate>
    </BrowserRouter>
  </React.StrictMode>
);
