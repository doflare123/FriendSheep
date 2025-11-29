import * as SecureStore from 'expo-secure-store';
import { AuthTokens } from '../types/auth';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';

const ACCESS_TOKEN_KEY = 'auth_access_token';
const REFRESH_TOKEN_KEY = 'auth_refresh_token';

const BASE_URL = API_BASE_URL || 'https://friendsheep.ru/api';

export const saveTokens = async (
  accessToken: string,
  refreshToken: string
): Promise<void> => {
  try {
    await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, accessToken);
    await SecureStore.setItemAsync(REFRESH_TOKEN_KEY, refreshToken);
  } catch (error) {
    console.error('Ошибка сохранения токенов:', error);
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
    console.error('Ошибка получения токенов:', error);
    return null;
  }
};

export const clearTokens = async (): Promise<void> => {
  try {
    await SecureStore.deleteItemAsync(ACCESS_TOKEN_KEY);
    await SecureStore.deleteItemAsync(REFRESH_TOKEN_KEY);
  } catch (error) {
    console.error('Ошибка очистки токенов:', error);
    throw error;
  }
};

export const hasTokens = async (): Promise<boolean> => {
  const tokens = await getTokens();
  return tokens !== null;
};

export const refreshAccessToken = async (): Promise<string | null> => {
  try {
    const tokens = await getTokens();
    
    if (!tokens?.refreshToken) {
      console.error('[TokenStorage] Refresh токен отсутствует');
      return null;
    }

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
      console.error('[TokenStorage] Ошибка обновления токена:', response.status);
  
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

    await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, data.access_token);

    if (data.refresh_token) {
      await SecureStore.setItemAsync(REFRESH_TOKEN_KEY, data.refresh_token);
    }

    return data.access_token;
  } catch (error) {
    console.error('[TokenStorage] Ошибка обновления токена:', error);
    return null;
  }
};

export const ensureValidToken = async (): Promise<string | null> => {
  try {
    const tokens = await getTokens();
    
    if (!tokens?.accessToken) {
      return null;
    }

    return tokens.accessToken;
  } catch (error) {
    console.error('[TokenStorage] Ошибка проверки токена:', error);
    return null;
  }
};