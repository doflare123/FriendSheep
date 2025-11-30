import apiClient from '@/api/apiClient';
import { getTokens, refreshAccessToken } from '@/api/storage/tokenStorage';
 
import { rateLimiter } from '@/utils/rateLimiter';
import { validateSessionId } from '@/utils/validators';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';
import {
  buildSessionFormData,
  downloadImage,
  logSessionData,
  uploadSessionImage
} from './sessionHelpers';
import {
  CreateSessionData,
  UpdateSessionData
} from './sessionTypes';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

class SessionService {

  async uploadSessionImage(imageUri: string): Promise<string> {
    return uploadSessionImage(imageUri);
  }

  async createSession(sessionData: CreateSessionData): Promise<any> {
    let tokens = await getTokens();
    if (!tokens?.accessToken) {
      throw new Error('Пользователь не авторизован');
    }

    try {
      console.log('[SessionService] 🚀 Создание сессии');

      let imageUrl: string;
      
      if (sessionData.image.uri.startsWith('http://') || 
          sessionData.image.uri.startsWith('https://')) {
        console.log('[SessionService] 🌐 Обнаружен внешний URL изображения');
        
        const localUri = await downloadImage(sessionData.image.uri);
        imageUrl = await this.uploadSessionImage(localUri);
        
        console.log('[SessionService] ✅ Изображение с Кинопоиска загружено на сервер');
      } else {
        console.log('[SessionService] 📸 Загрузка локального изображения');
        imageUrl = await this.uploadSessionImage(sessionData.image.uri);
        console.log('[SessionService] ✅ Изображение загружено');
      }

      logSessionData(sessionData, imageUrl);
      const formData = buildSessionFormData(sessionData, imageUrl);

      console.log('[SessionService] 📋 Отправка FormData:');
      //@ts-ignore
      for (let [key, value] of formData.entries()) {
        if (typeof value === 'string') {
          console.log(`  ${key}: ${value.length > 100 ? value.substring(0, 100) + '...' : value}`);
        } else {
          console.log(`  ${key}: [object]`, value);
        }
      }

      tokens = await getTokens();
      if (!tokens?.accessToken) {
        throw new Error('Токен отсутствует');
      }

      let response = await fetch(`${BASE_URL}/sessions/createSession`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
        body: formData,
      });

      if (response.status === 401) {
        console.log('[SessionService] 🔄 Токен истёк, обновляем...');
        
        try {
          const newAccessToken = await refreshAccessToken();
          if (!newAccessToken) {
            throw new Error('Не удалось обновить токен');
          }

          console.log('[SessionService] ✅ Токен обновлён, повторная попытка создания...');

          response = await fetch(`${BASE_URL}/sessions/createSession`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${newAccessToken}`,
            },
            body: formData,
          });
        } catch (refreshError) {
          console.error('[SessionService] ❌ Ошибка обновления токена:', refreshError);
          throw new Error('Не удалось обновить токен. Пожалуйста, войдите снова.');
        }
      }

      if (!response.ok) {
        const errorText = await response.text();
        console.error('[SessionService] ❌ Статус ответа:', response.status);
        console.error('[SessionService] ❌ Тело ошибки:', errorText);
        
        try {
          const errorData = JSON.parse(errorText);
          console.error('[SessionService] ❌ Распарсенная ошибка:', errorData);
          throw new Error(errorData.message || errorData.error || 'Ошибка создания сессии');
        } catch (parseError) {
          console.error('[SessionService] ❌ Не удалось распарсить ошибку');
          throw new Error(errorText || 'Ошибка создания сессии');
        }
      }

      const result = await response.json();
      console.log('[SessionService] ✅ Сессия создана:', result);
      console.log('[SessionService] 🔍 Проверяем city в ответе:', result.city);

      return result;
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка:', error);
      throw error;
    }
  }

  async getGenres(): Promise<string[]> {
    try {
      const response = await apiClient.get<string[]>('/sessions/genres');
      return response.data;
    } catch (error: any) {
      throw new Error(error.response?.data?.message || 'Ошибка получения жанров');
    }
  }

  async joinSession(params: { group_id: number; session_id: number }): Promise<any> {
    const validSessionId = validateSessionId(params.session_id);

    const rateLimitKey = `join_session_${validSessionId}`;
    if (!rateLimiter.canPerformAction(rateLimitKey, 5, 60000)) {
      throw new Error('Слишком много попыток. Подождите минуту.');
    }

    try {
      console.log(`[SessionService] Вступление в событие ${validSessionId} в группе ${params.group_id}`);
      
      const response = await apiClient.post('/sessions/join', {
        group_id: params.group_id,
        session_id: validSessionId
      });
      
      console.log('[SessionService] ✅ Успешно присоединились к сессии');
      return response.data;
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка вступления в событие:', error);
      console.error('[SessionService] Детали ошибки:', error.response?.data);

      if (error.response?.status === 404) {
        throw new Error('Сессия или пользователь не найдены');
      }
      if (error.response?.status === 409) {
        throw new Error('Сессия заполнена или вы уже присоединились');
      }
      
      throw new Error(error.response?.data?.message || error.response?.data?.error || 'Не удалось присоединиться к сессии');
    }
  }

  private actionTimestamps: Map<string, number[]> = new Map();

  private canPerformAction(key: string, maxActions: number, windowMs: number): boolean {
    const now = Date.now();
    const timestamps = this.actionTimestamps.get(key) || [];
    
    const recentTimestamps = timestamps.filter(t => now - t < windowMs);
    
    if (recentTimestamps.length >= maxActions) {
      return false;
    }
    
    recentTimestamps.push(now);
    this.actionTimestamps.set(key, recentTimestamps);
    return true;
  }

  async leaveSession(sessionId: number): Promise<any> {
    const validSessionId = validateSessionId(sessionId);

    const rateLimitKey = `leave_session_${validSessionId}`;
    if (!rateLimiter.canPerformAction(rateLimitKey, 5, 60000)) {
      throw new Error('Слишком много попыток. Подождите минуту.');
    }

    try {
      console.log(`[SessionService] Выход из события ${validSessionId}`);
      const response = await apiClient.post(`/sessions/${validSessionId}/leave`);
      return response.data;
    } catch (error: any) {
      console.error('[SessionService] Ошибка выхода из события');
      throw error;
    }
  }

  async deleteSession(sessionId: number): Promise<void> {
    try {
      console.log(`[SessionService] 🗑️ Удаление сессии ${sessionId}...`);
      await apiClient.delete(`/sessions/sessions/${sessionId}`);
      console.log('[SessionService] ✅ Сессия успешно удалена');
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка удаления сессии:', error);
      console.error('Детали ошибки:', error.response?.data);
      
      if (error.response?.status === 403) {
        throw new Error('Недостаточно прав для удаления сессии');
      }
      if (error.response?.status === 404) {
        throw new Error('Сессия не найдена');
      }
      throw new Error(error.response?.data?.message || error.response?.data?.error || 'Ошибка удаления сессии');
    }
  }

  async updateSession(sessionId: number, data: UpdateSessionData): Promise<any> {
    try {
      const response = await apiClient.patch(`/admin/sessions/${sessionId}`, data);
      return response.data;
    } catch (error: any) {
      if (error.response?.status === 403) {
        throw new Error('Нет прав на редактирование сессии');
      }
      throw new Error(error.response?.data?.message || 'Ошибка обновления сессии');
    }
  }

  async getSessionDetail(sessionId: number): Promise<any> {
    try {
      console.log('[SessionService] 📋 Загрузка детальной информации о сессии:', sessionId);
      const response = await apiClient.get(`/users/sessions/${sessionId}`);
      console.log('[SessionService] ✅ Данные сессии получены:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка загрузки сессии:', error);
      if (error.response?.status === 404) {
        throw new Error('Сессия не найдена');
      }
      if (error.response?.status === 403) {
        throw new Error('Нет доступа к этой сессии');
      }
      throw new Error(error.response?.data?.message || 'Ошибка загрузки информации о сессии');
    }
  }
  async getPopularSessions(): Promise<any> {
    try {
      console.log('[SessionService] 📊 Загрузка популярных сессий');
      const response = await apiClient.get('/users/sessions/popular');
      console.log('[SessionService] ✅ Популярные сессии получены:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка загрузки популярных сессий:', error);
      throw new Error(error.response?.data?.message || 'Ошибка загрузки популярных сессий');
    }
  }

  async getNewSessions(): Promise<any> {
    try {
      console.log('[SessionService] 🆕 Загрузка новых сессий');
      const response = await apiClient.get('/users/sessions/search', {
        params: {
          page: 1,
          new_only: true,
          sort_by: 'date',
          order: 'desc'
        }
      });
      console.log('[SessionService] ✅ Новые сессии получены:', response.data);
      return {
        count: response.data.total || 0,
        sessions: response.data.sessions || [],
        updated_at: new Date().toISOString()
      };
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка загрузки новых сессий:', error);
      throw new Error(error.response?.data?.message || 'Ошибка загрузки новых сессий');
    }
  }

  async getAllSessions(): Promise<any> {
    try {
      console.log('[SessionService] 📋 Загрузка всех сессий');
      const response = await apiClient.get('/users/sessions/search', {
        params: {
          page: 1,
          sort_by: 'date',
          order: 'desc'
        }
      });
      console.log('[SessionService] ✅ Все сессии получены:', response.data);
      return {
        count: response.data.total || 0,
        sessions: response.data.sessions || [],
        updated_at: new Date().toISOString()
      };
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка загрузки всех сессий:', error);
      throw new Error(error.response?.data?.message || 'Ошибка загрузки сессий');
    }
  }
}

export default new SessionService();