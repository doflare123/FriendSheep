import axios from 'axios';
import {decodeJWT} from '@/Constants'

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export function isTokenValid(token: string): boolean {
  const decoded = decodeJWT(token);
  if (!decoded || !decoded.exp) return false;
  return decoded.exp * 1000 > Date.now();
}

// Прямое обновление через ваш бэкенд
export async function refreshAccessToken(refreshToken: string): Promise<{
  access_token: string;
  refresh_token: string;
}> {
  try {
    const response = await axios.post(`${API_URL}/api/users/refresh`, {
      refresh_token: refreshToken
    });

    return response.data;
  } catch (error: any) {
    console.error('Ошибка при обновлении токена:', error);
    throw error;
  }
}

// Функции для работы с cookies
export function getCookie(name: string): string | undefined {
  if (typeof document === 'undefined') return undefined;
  
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    return parts.pop()?.split(';').shift();
  }
}

export function setCookie(name: string, value: string, days: number = 7) {
  if (typeof document === 'undefined') return;
  
  const expires = new Date();
  expires.setTime(expires.getTime() + (days * 24 * 60 * 60 * 1000));
  
  document.cookie = `${name}=${value}; expires=${expires.toUTCString()}; path=/; secure; samesite=Lax`;
}

export function deleteCookie(name: string) {
  if (typeof document === 'undefined') return;
  
  document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:01 GMT; path=/;`;
}

// Настройка axios interceptors
export function setupAxiosInterceptors() {
  axios.interceptors.request.use((config) => {
    const token = localStorage.getItem('access_token');
    if (token && isTokenValid(token)) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  axios.interceptors.response.use(
    (response) => response,
    async (error) => {
      const originalRequest = error.config;
      
      if (error.response?.status === 401 && !originalRequest._retry) {
        originalRequest._retry = true;
        
        try {
          const refreshToken = getCookie('refresh_token');
          if (refreshToken) {
            const tokens = await refreshAccessToken(refreshToken);
            localStorage.setItem('access_token', tokens.access_token);
            setCookie('refresh_token', tokens.refresh_token, 7);
            originalRequest.headers.Authorization = `Bearer ${tokens.access_token}`;
            return axios(originalRequest);
          }
        } catch (refreshError) {
          // ВАЖНО: редиректим ТОЛЬКО если мы НЕ на публичных страницах
          if (!['/login', '/register', '/'].includes(window.location.pathname)) {
            console.log("LOGIN1");
            window.location.href = '/login';
          }
        }
      }
      
      return Promise.reject(error);
    }
  );
}