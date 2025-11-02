import { clearTokens, getTokens, saveTokens } from '@/api/storage/tokenStorage';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';
import axios, { AxiosError, AxiosInstance } from 'axios';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

const apiClient: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

const PUBLIC_ENDPOINTS = [
  '/sessions/register',
  '/sessions/verify',
  '/users/login',
  '/users',
];

const isPublicEndpoint = (url?: string): boolean => {
  if (!url) return false;
  return PUBLIC_ENDPOINTS.some(endpoint => url.includes(endpoint));
};

apiClient.interceptors.request.use(
  async (config) => {
    if (isPublicEndpoint(config.url)) {
      return config;
    }

    const tokens = await getTokens();
    if (tokens?.accessToken) {
      config.headers.Authorization = `Bearer ${tokens.accessToken}`;
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as any;

    if (
      error.response?.status === 401 && 
      !originalRequest._retry && 
      !isPublicEndpoint(originalRequest?.url)
    ) {
      originalRequest._retry = true;

      try {
        const tokens = await getTokens();
        if (tokens?.refreshToken) {
          const response = await axios.post(
            `${BASE_URL}/users/refresh`,
            { refresh_token: tokens.refreshToken }
          );

          const { access_token, refresh_token } = response.data;
          await saveTokens(access_token, refresh_token);

          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return apiClient(originalRequest);
        }
      } catch (refreshError) {
        await clearTokens();
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default apiClient;