import { clearTokens, getTokens, saveTokens } from '@/api/storage/tokenStorage';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';
import axios, { AxiosError, AxiosInstance } from 'axios';

const BASE_URL = API_BASE_URL || 'https://friendsheep.ru/api';

if (__DEV__ === false && !BASE_URL.startsWith('https://')) {
  console.error('⚠️ КРИТИЧНО: Production API должен использовать HTTPS!');
}

let requestCounter = 0;
const MAX_PENDING_REQUESTS = 50;

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
  '/users/request-reset',
  '/users/confirm-reset',
];

const isPublicEndpoint = (url?: string): boolean => {
  if (!url) return false;

  return PUBLIC_ENDPOINTS.some(endpoint => {
    if (endpoint === '/users') {
      return url === '/users' || url.startsWith('/users?');
    }
    return url.includes(endpoint);
  });
};

apiClient.interceptors.request.use(
  async (config) => {
    requestCounter++;
    
    if (requestCounter > MAX_PENDING_REQUESTS) {
      requestCounter--;
      throw new Error('Слишком много одновременных запросов. Подождите.');
    }

    console.log(`\n[API] → ${config.method?.toUpperCase()} ${config.url}`);

    if (isPublicEndpoint(config.url)) {
      console.log('[API] Публичный endpoint');
      return config;
    }

    const tokens = await getTokens();
    console.log('[API] Токены:', tokens ? 'OK' : 'отсутствуют');

    if (tokens?.accessToken) {
      config.headers.Authorization = `Bearer ${tokens.accessToken}`;
      console.log('[API] → Authorization добавлен');
    } else {
      console.warn('[API] ⚠ Токен отсутствует');
    }

    return config;
  },
  (error) => {
    requestCounter--;
    return Promise.reject(error);
  }
);

let isRefreshing = false;
let refreshQueue: ((token: string | null) => void)[] = [];

const processQueue = (token: string | null) => {
  refreshQueue.forEach((resolve) => resolve(token));
  refreshQueue = [];
};


apiClient.interceptors.response.use(
  (response) => {
    requestCounter--;
    return response;
  },
  async (error: AxiosError) => {
    requestCounter--;
    const originalRequest: any = error.config;

    if (error.response?.status !== 401 || isPublicEndpoint(originalRequest?.url)) {
      return Promise.reject(error);
    }

    if (originalRequest._retry) {
      console.log('[API] ❌ 401 после refresh — выходим из аккаунта');
      await clearTokens();
      return Promise.reject(error);
    }
    originalRequest._retry = true;

    const tokens = await getTokens();
    if (!tokens?.refreshToken) {
      console.log('[API] ❌ Нет refresh токена — logout');
      await clearTokens();
      return Promise.reject(error);
    }

    if (isRefreshing) {
      console.log('[API] ⏳ Ждём завершения refresh...');
      return new Promise(resolve => {
        refreshQueue.push((newToken) => {
          if (newToken) {
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            resolve(apiClient(originalRequest));
          } else {
            resolve(Promise.reject(error));
          }
        });
      });
    }

    isRefreshing = true;
    console.log('[API] 🔄 Refresh token...');

    try {
      const response = await axios.post(`${BASE_URL}/users/refresh`, {
        refresh_token: tokens.refreshToken,
      });

      const { access_token, refresh_token } = response.data;

      await saveTokens(access_token, refresh_token);

      await new Promise(res => setTimeout(res, 10));

      processQueue(access_token);
      console.log('[API] ✔ Refresh успешен');

      originalRequest.headers.Authorization = `Bearer ${access_token}`;
      return apiClient(originalRequest);

    } catch (refreshError) {
      console.log('[API] ❌ Refresh провалился — logout');
      processQueue(null);
      await clearTokens();
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  }
);

export default apiClient;