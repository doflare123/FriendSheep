import * as SecureStore from 'expo-secure-store';
import { AuthTokens } from '../types/auth';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';

const ACCESS_TOKEN_KEY = 'auth_access_token';
const REFRESH_TOKEN_KEY = 'auth_refresh_token';
const TOKEN_TIMESTAMP_KEY = 'auth_token_timestamp';

const BASE_URL = API_BASE_URL || 'https://friendsheep.ru/api';
const REFRESH_INTERVAL = 15 * 60 * 1000;

let refreshTimerId: number | null = null;

export const saveTokens = async (
  accessToken: string,
  refreshToken: string
): Promise<void> => {
  try {
    const timestamp = Date.now().toString();
    await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, accessToken);
    await SecureStore.setItemAsync(REFRESH_TOKEN_KEY, refreshToken);
    await SecureStore.setItemAsync(TOKEN_TIMESTAMP_KEY, timestamp);
    
    console.log('[TokenStorage] Токены сохранены');

    startTokenRefreshTimer();
  } catch (error) {
    console.error('[TokenStorage] Ошибка сохранения токенов:', error);
    throw error;
  }
};

export const getTokens = async (): Promise<AuthTokens | null> => {
  try {
    const accessToken = await SecureStore.getItemAsync(ACCESS_TOKEN_KEY);
    const refreshToken = await SecureStore.getItemAsync(REFRESH_TOKEN_KEY);

    if (accessToken && refreshToken) {
      return { accessToken, refreshToken };
    }
    return null;
  } catch (error) {
    console.error('[TokenStorage] Ошибка получения токенов:', error);
    return null;
  }
};

export const clearTokens = async (): Promise<void> => {
  try {
    await SecureStore.deleteItemAsync(ACCESS_TOKEN_KEY);
    await SecureStore.deleteItemAsync(REFRESH_TOKEN_KEY);
    await SecureStore.deleteItemAsync(TOKEN_TIMESTAMP_KEY);

    stopTokenRefreshTimer();
    
    console.log('[TokenStorage] Токены очищены');
  } catch (error) {
    console.error('[TokenStorage] Ошибка очистки токенов:', error);
    throw error;
  }
};

export const refreshAccessToken = async (): Promise<string | null> => {
  try {
    const tokens = await getTokens();
    
    if (!tokens?.refreshToken) {
      console.error('[TokenStorage] Refresh токен отсутствует');
      stopTokenRefreshTimer();
      return null;
    }

    console.log('[TokenStorage] 🔄 Обновление токена...');

    const response = await fetch(`${BASE_URL}/users/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refresh_token: tokens.refreshToken,
      }),
    });

    if (!response.ok) {
      console.error('[TokenStorage] ❌ Ошибка обновления токена:', response.status);
  
      if (response.status === 401) {
        await clearTokens();
      }
      
      return null;
    }

    const data = await response.json();
    
    if (!data.access_token) {
      console.error('[TokenStorage] В ответе нет access_token');
      return null;
    }

    await saveTokens(data.access_token, data.refresh_token || tokens.refreshToken);
    
    console.log('[TokenStorage] ✅ Токен успешно обновлен');

    return data.access_token;
  } catch (error) {
    console.error('[TokenStorage] Ошибка обновления токена:', error);
    return null;
  }
};

const startTokenRefreshTimer = () => {
  stopTokenRefreshTimer();
  
  console.log('[TokenStorage] ⏰ Запущен таймер автообновления токена (каждые 15 минут)');

  refreshTimerId = setInterval(async () => {
    console.log('[TokenStorage] ⏰ Время обновления токена');
    const newToken = await refreshAccessToken();
    
    if (!newToken) {
      console.error('[TokenStorage] ❌ Не удалось обновить токен автоматически');
      stopTokenRefreshTimer();
    }
  }, REFRESH_INTERVAL) as unknown as number;
};

const stopTokenRefreshTimer = () => {
  if (refreshTimerId !== null) {
    clearInterval(refreshTimerId);
    refreshTimerId = null;
    console.log('[TokenStorage] ⏰ Таймер автообновления остановлен');
  }
};

export const initializeTokenRefresh = async () => {
  const tokens = await getTokens();
  if (tokens) {
    console.log('[TokenStorage] Найдены сохраненные токены, запускаем автообновление');
    startTokenRefreshTimer();
  }
};