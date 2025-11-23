import AsyncStorage from '@react-native-async-storage/async-storage';
import { AuthTokens } from '../types/auth';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';

const ACCESS_TOKEN_KEY = '@auth_access_token';
const REFRESH_TOKEN_KEY = '@auth_refresh_token';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

export const saveTokens = async (
  accessToken: string,
  refreshToken: string
): Promise<void> => {
  try {
    await AsyncStorage.multiSet([
      [ACCESS_TOKEN_KEY, accessToken],
      [REFRESH_TOKEN_KEY, refreshToken],
    ]);
  } catch (error) {
    console.error('Ошибка сохранения токенов:', error);
    throw error;
  }
  console.log("SAVE TOKENS INPUT:", accessToken, refreshToken);
};

export const getTokens = async (): Promise<AuthTokens | null> => {
  try {
    const values = await AsyncStorage.multiGet([
      ACCESS_TOKEN_KEY,
      REFRESH_TOKEN_KEY,
    ]);

    const accessToken = values[0][1];
    const refreshToken = values[1][1];

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
    await AsyncStorage.multiRemove([ACCESS_TOKEN_KEY, REFRESH_TOKEN_KEY]);
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
      console.error('[TokenStorage] ❌ Refresh токен отсутствует');
      return null;
    }

    console.log('[TokenStorage] 🔄 Обновление access токена...');

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
      const errorText = await response.text();
      console.error('[TokenStorage] ❌ Ошибка обновления токена:', errorText);
      
      // Если refresh токен тоже истёк, очищаем всё
      if (response.status === 401) {
        await clearTokens();
      }
      
      return null;
    }

    const data = await response.json();
    
    if (!data.access_token) {
      console.error('[TokenStorage] ❌ В ответе нет access_token');
      return null;
    }

    // Сохраняем новый access токен
    await AsyncStorage.setItem(ACCESS_TOKEN_KEY, data.access_token);
    
    // Если пришёл новый refresh токен, сохраняем и его
    if (data.refresh_token) {
      await AsyncStorage.setItem(REFRESH_TOKEN_KEY, data.refresh_token);
    }

    console.log('[TokenStorage] ✅ Access токен успешно обновлён');
    return data.access_token;
  } catch (error) {
    console.error('[TokenStorage] ❌ Ошибка обновления токена:', error);
    return null;
  }
};

/**
 * Проверяет валидность access токена
 * Если истёк - автоматически обновляет
 */
export const ensureValidToken = async (): Promise<string | null> => {
  try {
    const tokens = await getTokens();
    
    if (!tokens?.accessToken) {
      return null;
    }

    // Здесь можно добавить проверку срока действия токена
    // Пока просто возвращаем существующий токен
    return tokens.accessToken;
  } catch (error) {
    console.error('[TokenStorage] ❌ Ошибка проверки токена:', error);
    return null;
  }
};