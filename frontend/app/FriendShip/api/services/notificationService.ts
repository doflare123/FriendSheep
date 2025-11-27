import apiClient from '../apiClient';
import {
    MarkAsViewedResponse,
    Notification,
    NotificationsResponse,
    normalizeInvite,
    normalizeNotification
} from '../types/notification';

class NotificationService {
  private viewedNotificationIds: Set<number> = new Set();

  async getNotifications(): Promise<NotificationsResponse> {
    try {
      console.log('[NotificationService] 📬 Загрузка уведомлений...');
      
      const response = await apiClient.get<any>('/users/notify');
      
      console.log('[NotificationService] ✅ Ответ сервера:', JSON.stringify(response.data, null, 2));
 
      const notifications = (response.data.notifications || [])
        .map(normalizeNotification)
        .filter((n: Notification) => !this.viewedNotificationIds.has(n.id));
      
      const invites = (response.data.invites || [])
        .map(normalizeInvite)
        .filter((inv: any) => inv.status === 'pending');
      
      console.log('[NotificationService] ✅ Загружено:', {
        notifications: notifications.length,
        invites: invites.length,
        viewedInMemory: this.viewedNotificationIds.size,
      });
      
      return {
        notifications,
        invites,
      };
    } catch (error: any) {
      console.error('[NotificationService] ❌ Ошибка загрузки уведомлений:', error);
      console.error('[NotificationService] 📋 Детали ошибки:', {
        status: error.response?.status,
        statusText: error.response?.statusText,
        data: error.response?.data,
        message: error.message,
      });
      throw this.handleError(error);
    }
  }

  async hasNotifications(): Promise<boolean> {
    try {
      const response = await apiClient.get<boolean>('/users/notify/inf');
      console.log('[NotificationService] 🔔 Есть уведомления:', response.data);
      return response.data;
    } catch (error: any) {
      try {
        const data = await this.getNotifications();
        const hasAny = (data.notifications.length + data.invites.length) > 0;
        console.log('[NotificationService] 🔔 Есть уведомления (fallback):', hasAny);
        return hasAny;
      } catch (fallbackError) {
        console.error('[NotificationService] ❌ Ошибка проверки уведомлений:', error);
        return false;
      }
    }
  }

  async markAsViewed(notificationId: number): Promise<MarkAsViewedResponse> {
    try {
      console.log('[NotificationService] 👁️ Отмечаю уведомление как просмотренное:', notificationId);
      
      this.viewedNotificationIds.add(notificationId);

      try {
        const response = await apiClient.post<MarkAsViewedResponse>(
          '/users/notifications/viewed',
          { id: notificationId }
        );
        
        console.log('[NotificationService] ✅ Уведомление отмечено на бэкенде');
        return response.data;
      } catch (backendError: any) {
        console.log('[NotificationService] ⚠️ Бэкенд не поддерживает viewed, используем локальное хранилище');
        return { success: 'true' };
      }
    } catch (error: any) {
      console.error('[NotificationService] ❌ Ошибка отметки уведомления:', error);
      throw this.handleError(error);
    }
  }

  clearViewedCache(): void {
    this.viewedNotificationIds.clear();
    console.log('[NotificationService] 🗑️ Кэш просмотренных уведомлений очищен');
  }

  getViewedCount(): number {
    return this.viewedNotificationIds.size;
  }

  private handleError(error: any): Error {
    if (error.response) {
      const status = error.response.status;
      const data = error.response.data;

      const errorMessage =
        typeof data === 'object' && data.error
          ? data.error
          : typeof data === 'object'
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
          return new Error(`Ошибка сервера: ${errorMessage}`);
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