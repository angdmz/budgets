import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

type GetAccessTokenSilently = (options?: object) => Promise<string>;

export async function createApiClient(getAccessTokenSilently: GetAccessTokenSilently) {
  const token = await getAccessTokenSilently({
    authorizationParams: {
      audience: import.meta.env.VITE_AUTH0_AUDIENCE,
    },
  });

  return axios.create({
    baseURL: API_BASE_URL,
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });
}

export function getErrorMessage(error: unknown): string | undefined {
  if (axios.isAxiosError(error)) {
    return error.response?.data?.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return undefined;
}
