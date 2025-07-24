import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

// Декодирование JWT токена
export function decodeJWT(token: string) {
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));
    return decoded;
  } catch (error) {
    console.error('Ошибка декодирования JWT:', error);
    return null;
  }
}

// Обновление access токена через refresh токен
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

// Получение cookie по имени
export function getCookie(name: string): string | undefined {
  if (typeof document === 'undefined') return undefined;
  
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    return parts.pop()?.split(';').shift();
  }
}

// Установка cookie
export function setCookie(name: string, value: string, days: number = 7) {
  if (typeof document === 'undefined') return;
  
  const expires = new Date();
  expires.setTime(expires.getTime() + (days * 24 * 60 * 60 * 1000));
  
  document.cookie = `${name}=${value}; expires=${expires.toUTCString()}; path=/; secure; samesite=strict`;
}

// Удаление cookie
export function deleteCookie(name: string) {
  if (typeof document === 'undefined') return;
  
  document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:01 GMT; path=/;`;
}

// Проверка валидности токена
export function isTokenValid(token: string): boolean {
  const decoded = decodeJWT(token);
  if (!decoded || !decoded.exp) return false;
  
  return decoded.exp * 1000 > Date.now();
}

// Автоматическое обновление токена при запросах
export function setupAxiosInterceptors() {
  // Interceptor для добавления токена к запросам
  axios.interceptors.request.use((config) => {
    const token = localStorage.getItem('access_token');
    if (token && isTokenValid(token)) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  // Interceptor для обработки ошибок авторизации
  axios.interceptors.response.use(
    (response) => response,
    async (error) => {
      const originalRequest = error.config;
      
      if (error.response?.status === 401 && !originalRequest._retry) {
        originalRequest._retry = true;
        
        const refreshToken = localStorage.getItem('refresh_token');
        // или const refreshToken = getCookie('refresh_token');
        
        if (refreshToken) {
          try {
            const tokens = await refreshAccessToken(refreshToken);
            localStorage.setItem('access_token', tokens.access_token);
            localStorage.setItem('refresh_token', tokens.refresh_token);
            
            originalRequest.headers.Authorization = `Bearer ${tokens.access_token}`;
            return axios(originalRequest);
          } catch (refreshError) {
            // Refresh токен тоже недействителен - разлогиниваем
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            // Перенаправление на страницу входа
            window.location.href = '/login';
          }
        }
      }
      
      return Promise.reject(error);
    }
  );
}