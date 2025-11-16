import AsyncStorage from '@react-native-async-storage/async-storage';
import { AuthTokens } from '../types/auth';

const ACCESS_TOKEN_KEY = '@auth_access_token';
const REFRESH_TOKEN_KEY = '@auth_refresh_token';

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