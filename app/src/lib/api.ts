import axios from 'axios';
import { useAuth0 } from '@auth0/auth0-react';
import { notifySessionExpired, SessionExpiredError } from './session';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

const AUTH_ERROR_CODES = [
  'login_required',
  'consent_required',
  'interaction_required',
  'invalid_grant',
  'missing_refresh_token',
];

export type GetAccessTokenSilently = ReturnType<typeof useAuth0>['getAccessTokenSilently'];

export async function createApiClient(getAccessTokenSilently: GetAccessTokenSilently) {
  let token: string;
  try {
    token = await getAccessTokenSilently({
      authorizationParams: {
        audience: import.meta.env.VITE_AUTH0_AUDIENCE,
      },
    }) as string;
  } catch (error: any) {
    if (
      AUTH_ERROR_CODES.includes(error?.error) ||
      error?.name === 'MissingRefreshTokenError'
    ) {
      notifySessionExpired();
      throw new SessionExpiredError();
    }
    throw error;
  }

  const client = axios.create({
    baseURL: API_BASE_URL,
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  client.interceptors.response.use(
    (response) => response,
    (error) => {
      if (axios.isAxiosError(error) && error.response?.status === 401) {
        notifySessionExpired();
        return Promise.reject(new SessionExpiredError());
      }
      return Promise.reject(error);
    }
  );

  return client;
}

export function getErrorMessage(error: unknown): string | undefined {
  if (error instanceof SessionExpiredError) return undefined;
  if (axios.isAxiosError(error)) {
    return error.response?.data?.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return undefined;
}
