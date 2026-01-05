import { getTokens, refreshAccessToken } from '@/api/storage/tokenStorage';

export const handleApiError = <T>(
  error: any,
  context: string = 'Произошла ошибка'
): never => {
  console.error(`[${context}]`, error);
  
  const errorMessage = error.response?.data?.error || 
                      error.response?.data?.message || 
                      error.message || 
                      context;
  
  throw new Error(errorMessage);
};

export const fetchWithRetry = async <T>(
  url: string,
  options: RequestInit = {},
  context: string = 'API запрос'
): Promise<T> => {
  let tokens = await getTokens();
  if (!tokens?.accessToken) {
    throw new Error('Пользователь не авторизован');
  }

  const makeRequest = async (accessToken: string): Promise<Response> => {
    return fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${accessToken}`,
      },
    });
  };

  try {
    let response = await makeRequest(tokens.accessToken);

    if (response.status === 401) {
      console.log(`[${context}] 🔄 Токен истёк, обновляем...`);
      
      const newToken = await refreshAccessToken();
      if (!newToken) {
        throw new Error('Не удалось обновить токен');
      }

      console.log(`[${context}] ✅ Токен обновлён, повторная попытка...`);
      response = await makeRequest(newToken);
    }

    if (!response.ok) {
      const errorText = await response.text();
      console.error(`[${context}] ❌ Ошибка:`, response.status, errorText);
      
      try {
        const errorData = JSON.parse(errorText);
        throw new Error(errorData.message || errorData.error || context);
      } catch (parseError) {
        throw new Error(errorText || context);
      }
    }

    return await response.json();
  } catch (error: any) {
    console.error(`[${context}] ❌`, error);
    throw error;
  }
};

export const createErrorHandler = (serviceName: string) => {
  return (error: any): Error => {
    if (error.response) {
      const status = error.response.status;
      const data = error.response.data;

      const errorMessage =
        typeof data === 'object'
          ? Object.values(data).join(', ')
          : data || 'Произошла ошибка';

      switch (status) {
        case 400:
          return new Error(`Неверные данные: ${errorMessage}`);
        case 401:
          return new Error('Необходимо войти в систему');
        case 403:
          return new Error('Нет прав для выполнения действия');
        case 404:
          return new Error(`${serviceName}: не найдено`);
        case 500:
          return new Error('Ошибка сервера. Попробуйте позже');
        default:
          return new Error(errorMessage);
      }
    } else if (error.request) {
      return new Error('Нет связи с сервером');
    } else {
      return new Error(error.message || 'Неизвестная ошибка');
    }
  };
};