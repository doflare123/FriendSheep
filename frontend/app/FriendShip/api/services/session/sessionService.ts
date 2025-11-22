import apiClient from '@/api/apiClient';
import { getTokens } from '@/api/storage/tokenStorage';
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

      const response = await fetch(`${BASE_URL}/sessions/createSession`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
        body: formData,
      });

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