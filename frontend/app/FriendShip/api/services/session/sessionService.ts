import apiClient from '@/api/apiClient';
import { getTokens } from '@/api/storage/tokenStorage';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';
import {
  buildSessionFormData,
  logSessionData,
  uploadSessionImage
} from './sessionHelpers';
import {
  CreateSessionData,
  JoinSessionData,
  UpdateSessionData
} from './sessionTypes';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

class SessionService {

  async uploadSessionImage(imageUri: string): Promise<string> {
    return uploadSessionImage(imageUri);
  }

  async createSession(sessionData: CreateSessionData): Promise<any> {
    const tokens = await getTokens();
    if (!tokens?.accessToken) {
      throw new Error('Пользователь не авторизован');
    }

    try {
      console.log('[SessionService] 🚀 Создание сессии');

      const imageUrl = await this.uploadSessionImage(sessionData.image.uri);

      console.log('[SessionService] ✅ Изображение загружено');

      logSessionData(sessionData, imageUrl);

      const formData = buildSessionFormData(sessionData, imageUrl);

      const response = await fetch(`${BASE_URL}/sessions/createSession`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const errorText = await response.text();
        const errorData = JSON.parse(errorText);
        throw new Error(errorData.message || errorData.error || 'Ошибка создания сессии');
      }

      const result = await response.json();
      console.log('[SessionService] ✅ Сессия создана:', result);
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

  async joinSession(data: JoinSessionData): Promise<any> {
    try {
      const response = await apiClient.post('/sessions/join', data);
      return response.data;
    } catch (error: any) {
      if (error.response?.status === 409) {
        throw new Error('Сессия заполнена или вы уже присоединились');
      }
      throw new Error(error.response?.data?.message || 'Ошибка присоединения к сессии');
    }
  }

  async leaveSession(sessionId: number): Promise<void> {
    try {
      await apiClient.delete(`/sessions/${sessionId}/leave`);
    } catch (error: any) {
      if (error.response?.status === 403) {
        throw new Error('Вы не состоите в этой сессии');
      }
      throw new Error(error.response?.data?.message || 'Ошибка выхода из сессии');
    }
  }

  async deleteSession(sessionId: number): Promise<void> {
    try {
      await apiClient.delete(`/sessions/sessions/${sessionId}`);
    } catch (error: any) {
      if (error.response?.status === 403) {
        throw new Error('Недостаточно прав для удаления сессии');
      }
      throw new Error(error.response?.data?.message || 'Ошибка удаления сессии');
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
}

export default new SessionService();