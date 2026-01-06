import apiClient from '@/api/apiClient';
import { fetchWithRetry } from '@/utils/errorHandler';
import { rateLimiter } from '@/utils/rateLimiter';
import { validateSessionId } from '@/utils/validators';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';
import AsyncStorage from '@react-native-async-storage/async-storage';
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
    try {
      console.log('[SessionService] 🚀 Создание сессии');

      let imageUrl: string;
      if (sessionData.image.uri.startsWith('http://') || 
          sessionData.image.uri.startsWith('https://')) {
        console.log('[SessionService] 🌐 Обнаружен внешний URL изображения');
        const localUri = await downloadImage(sessionData.image.uri);
        imageUrl = await this.uploadSessionImage(localUri);
      } else {
        console.log('[SessionService] 📸 Загрузка локального изображения');
        imageUrl = await this.uploadSessionImage(sessionData.image.uri);
      }

      logSessionData(sessionData, imageUrl);
      const formData = buildSessionFormData(sessionData, imageUrl);

      const result = await fetchWithRetry(
        `${BASE_URL}/sessions/createSession`,
        {
          method: 'POST',
          body: formData,
        },
        'SessionService.createSession'
      );

      console.log('[SessionService] ✅ Сессия создана:', result);
      return result;
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка:', error);

      let errorMessage = 'Ошибка создания сессии';
      
      try {
        if (error.message && typeof error.message === 'string') {
          if (error.message.startsWith('{') || error.message.startsWith('[')) {
            try {
              const parsed = JSON.parse(error.message);
              errorMessage = parsed.error || errorMessage;
            } catch {
              errorMessage = error.message;
            }
          } else {
            errorMessage = error.message;
          }
        }
        else if (error.response?.data?.error) {
          const errorData = error.response.data.error;
          
          if (typeof errorData === 'string' && (errorData.startsWith('{') || errorData.startsWith('['))) {
            try {
              const parsed = JSON.parse(errorData);
              errorMessage = parsed.error || errorMessage;
            } catch {
              errorMessage = errorData;
            }
          } else {
            errorMessage = errorData;
          }
        }
        else if (error.response?.data?.message) {
          errorMessage = error.response.data.message;
        }
      } catch (parseError) {
        console.error('[SessionService] ❌ Ошибка парсинга:', parseError);
      }
      
      throw new Error(errorMessage);
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
      const response = await apiClient.delete(`/sessions/${validSessionId}/leave`);
      console.log('[SessionService] ✅ Успешно покинули сессию');
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

  async saveCalendarEventId(sessionId: number, calendarEventId: string): Promise<void> {
    try {
      console.log('[SessionService] 💾 Сохранение calendarEventId:', calendarEventId);
      
      await AsyncStorage.setItem(
        `calendar_event_${sessionId}`, 
        calendarEventId
      );
      
      console.log('[SessionService] ✅ calendarEventId сохранён');
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка сохранения calendarEventId:', error);
      throw error;
    }
  }

  async getCalendarEventId(sessionId: number): Promise<string | null> {
    try {
      console.log('[SessionService] 📖 Загрузка calendarEventId для сессии:', sessionId);
      
      const calendarEventId = await AsyncStorage.getItem(
        `calendar_event_${sessionId}`
      );
      
      console.log('[SessionService] ✅ calendarEventId загружен:', calendarEventId);
      return calendarEventId;
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка загрузки calendarEventId:', error);
      return null;
    }
  }

  async removeCalendarEventId(sessionId: number): Promise<void> {
    try {
      console.log('[SessionService] 🗑️ Удаление calendarEventId для сессии:', sessionId);
      
      await AsyncStorage.removeItem(`calendar_event_${sessionId}`);
      
      console.log('[SessionService] ✅ calendarEventId удалён');
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка удаления calendarEventId:', error);
    }
  }

  async getUserGroupSessions(page: number = 1): Promise<any> {
    try {
      console.log('[SessionService] 👤 Загрузка событий пользователя, страница:', page);
      const response = await apiClient.get('/users/sessions/user-groups', {
        params: { page }
      });
      console.log('[SessionService] ✅ События пользователя получены:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка загрузки событий пользователя:', error);
      if (error.response?.status === 401) {
        throw new Error('Необходимо войти в систему');
      }
      if (error.response?.status === 404) {
        throw new Error('Пользователь не найден');
      }
      throw new Error(error.response?.data?.error || 'Ошибка загрузки событий пользователя');
    }
  }
}

export default new SessionService();