import apiClient from '../apiClient';
import {
  MarkAsViewedResponse,
  NotificationsResponse
} from '../types/notification';

class NotificationService {

  async getNotifications(): Promise<NotificationsResponse> {
    try {
      console.log('[NotificationService] 📬 Загрузка уведомлений...');
      
      const response = await apiClient.get<NotificationsResponse>('/users/notify');
      
      console.log('[NotificationService] ✅ Загружено:', {
        notifications: response.data.notifications?.length || 0,
        invites: response.data.invites?.length || 0,
      });
      
      return {
        notifications: response.data.notifications || [],
        invites: response.data.invites || [],
      };
    } catch (error: any) {
      console.error('[NotificationService] ❌ Ошибка загрузки уведомлений:', error);
      throw this.handleError(error);
    }
  }

  async hasNotifications(): Promise<boolean> {
    try {
      const response = await apiClient.get<boolean>('/users/notify/inf');
      console.log('[NotificationService] 🔔 Есть непросмотренные:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('[NotificationService] ❌ Ошибка проверки уведомлений:', error);
      throw this.handleError(error);
    }
  }

  async markAsViewed(notificationId: number): Promise<MarkAsViewedResponse> {
    try {
      console.log('[NotificationService] 👁️ Отмечаю уведомление как просмотренное:', notificationId);
      
      const response = await apiClient.post<MarkAsViewedResponse>(
        '/users/notifications/viewed',
        { id: notificationId }
      );
      
      console.log('[NotificationService] ✅ Уведомление отмечено');
      return response.data;
    } catch (error: any) {
      console.error('[NotificationService] ❌ Ошибка отметки уведомления:', error);
      throw this.handleError(error);
    }
  }

  private handleError(error: any): Error {
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
        case 404:
          return new Error('Уведомление не найдено');
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
  }
}

export default new NotificationService();