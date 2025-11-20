import apiClient from '@/api/apiClient';
import { getTokens } from '@/api/storage/tokenStorage';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

export interface CreateSessionData {
  title: string;
  session_type: string;
  session_place: number;
  group_id: number;
  start_time: string;
  duration?: number;
  count_users: number;
  genres?: string;
  fields?: string;
  location?: string;
  year?: number;
  country?: string;
  age_limit?: string;
  notes?: string;
  image: {
    uri: string;
    name: string;
    type: string;
  };
}

export interface UpdateSessionData {
  title?: string;
  session_type_id?: number;
  session_place_id?: number;
  start_time?: string;
  end_time?: string;
  duration?: number;
  count_users_max?: number;
  genres?: string[];
  image_url?: string;
  location?: string;
  year?: number;
  country?: string;
  age_limit?: string;
  notes?: string;
}

export interface JoinSessionData {
  group_id: number;
  session_id: number;
}

class SessionService {
  async createSession(sessionData: CreateSessionData): Promise<any> {
    const tokens = await getTokens();
    if (!tokens?.accessToken) {
      throw new Error('Пользователь не авторизован');
    }

    console.log('[SessionService] 🚀 Создание сессии (правильный подход)');

    try {
      console.log('[SessionService] 📸 Шаг 1: Загрузка изображения...');
      const imageUrl = await this.uploadSessionImage(sessionData.image.uri);
      console.log('[SessionService] ✅ Изображение загружено:', imageUrl);

      console.log('[SessionService] 📝 Шаг 2: Создание сессии с FormData...');
      
      const formData = new FormData();

      formData.append('title', sessionData.title);
      formData.append('session_type', sessionData.session_type);
      formData.append('session_place', sessionData.session_place.toString());
      formData.append('group_id', sessionData.group_id.toString());
      formData.append('start_time', sessionData.start_time);
      formData.append('count_users', sessionData.count_users.toString());

      formData.append('image', imageUrl);

      if (sessionData.duration !== undefined) {
        formData.append('duration', sessionData.duration.toString());
      }

      if (sessionData.genres) {
        formData.append('genres', sessionData.genres);
      }

      if (sessionData.fields) {
        formData.append('fields', sessionData.fields);
      }

      if (sessionData.location) {
        formData.append('location', sessionData.location);
      }

      if (sessionData.year !== undefined) {
        formData.append('year', sessionData.year.toString());
      }

      if (sessionData.country) {
        formData.append('country', sessionData.country);
      }

      if (sessionData.age_limit) {
        formData.append('age_limit', sessionData.age_limit);
      }

      if (sessionData.notes) {
        formData.append('notes', sessionData.notes);
      }

      console.log('[SessionService] 📦 FormData подготовлена (image как URL-строка)');

      const response = await fetch(`${BASE_URL}/sessions/createSession`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
        body: formData,
      });

      console.log('[SessionService] ✅ Статус ответа:', response.status);

      if (!response.ok) {
        const errorText = await response.text();
        console.error('[SessionService] ❌ Ошибка от сервера:', errorText);
        
        try {
          const errorData = JSON.parse(errorText);
          throw new Error(errorData.message || errorData.error || 'Ошибка создания сессии');
        } catch (parseError) {
          throw new Error(`Ошибка создания сессии: ${response.status}`);
        }
      }

      const result = await response.json();
      console.log('[SessionService] ✅ Сессия успешно создана:', result);
      return result;
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка при создании сессии:', error);
      throw error;
    }
  }

  async getGenres(): Promise<string[]> {
    try {
      console.log('[SessionService] Загрузка жанров...');
      const response = await apiClient.get<string[]>('/sessions/genres');
      console.log('[SessionService] Жанры получены:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('[SessionService] Ошибка получения жанров:', error);
      throw new Error(error.response?.data?.message || 'Ошибка получения жанров');
    }
  }

  async joinSession(data: JoinSessionData): Promise<any> {
    try {
      console.log('[SessionService] Присоединение к сессии:', data);
      const response = await apiClient.post('/sessions/join', data);
      console.log('[SessionService] Присоединение успешно:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('[SessionService] Ошибка присоединения к сессии:', error);
      
      if (error.response?.status === 409) {
        throw new Error('Сессия заполнена или вы уже присоединились');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка присоединения к сессии');
    }
  }

  async leaveSession(sessionId: number): Promise<void> {
    try {
      console.log('[SessionService] Выход из сессии:', sessionId);
      const response = await apiClient.delete(`/sessions/${sessionId}/leave`);
      console.log('[SessionService] Выход успешен:', response.data);
    } catch (error: any) {
      console.error('[SessionService] Ошибка выхода из сессии:', error);
      
      if (error.response?.status === 403) {
        throw new Error('Вы не состоите в этой сессии');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка выхода из сессии');
    }
  }

  async deleteSession(sessionId: number): Promise<void> {
    try {
      console.log('[SessionService] Удаление сессии:', sessionId);
      const response = await apiClient.delete(`/sessions/sessions/${sessionId}`);
      console.log('[SessionService] Сессия удалена:', response.data);
    } catch (error: any) {
      console.error('[SessionService] Ошибка удаления сессии:', error);
      
      if (error.response?.status === 403) {
        throw new Error('Недостаточно прав для удаления сессии');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка удаления сессии');
    }
  }

  async updateSession(sessionId: number, data: UpdateSessionData): Promise<any> {
    try {
      console.log('[SessionService] Обновление сессии:', sessionId, data);
      const response = await apiClient.patch(`/admin/sessions/${sessionId}`, data);
      console.log('[SessionService] Сессия обновлена:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('[SessionService] Ошибка обновления сессии:', error);
      
      if (error.response?.status === 403) {
        throw new Error('Нет прав на редактирование сессии');
      }
      
      throw new Error(error.response?.data?.message || 'Ошибка обновления сессии');
    }
  }

async uploadSessionImage(imageUri: string): Promise<string> {
    try {
      const tokens = await getTokens();
      if (!tokens?.accessToken) {
        throw new Error('Пользователь не авторизован');
      }

      console.log('[SessionService] 📸 Загрузка изображения через fetch...');

      const formData = new FormData();
      
      const filename = imageUri.split('/').pop() || `session_${Date.now()}.jpg`;
      const fileExtension = filename.split('.').pop()?.toLowerCase() || 'jpg';
      
      formData.append('image', {
        uri: imageUri,
        name: filename,
        type: `image/${fileExtension === 'jpg' ? 'jpeg' : fileExtension}`,
      } as any);

      const response = await fetch(`${BASE_URL}/admin/groups/UploadPhoto`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
        body: formData,
      });

      console.log('[SessionService] ✅ Статус загрузки изображения:', response.status);

      if (!response.ok) {
        const errorText = await response.text();
        console.error('[SessionService] ❌ Ошибка загрузки:', errorText);
        throw new Error('Ошибка загрузки изображения');
      }

      const result = await response.json();
      console.log('[SessionService] ✅ Результат загрузки:', result);

      const imageUrl = result.url 
        || result.image_url 
        || result.image 
        || Object.values(result)[0];
      
      if (!imageUrl || typeof imageUrl !== 'string') {
        console.error('[SessionService] ❌ Не удалось найти URL в ответе:', result);
        throw new Error('Не удалось получить URL изображения');
      }
      
      console.log('[SessionService] ✅ URL изображения:', imageUrl);
      return imageUrl as string;
    } catch (error: any) {
      console.error('[SessionService] ❌ Ошибка загрузки изображения:', error);
      throw error;
    }
  }
}

export default new SessionService();