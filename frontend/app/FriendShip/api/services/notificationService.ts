import { validateNotificationData } from '@/utils/validators';
import apiClient from '../apiClient';
import {
  GroupInvite,
  MarkAsViewedResponse,
  Notification,
  NotificationsResponse,
  UnreadNotificationsResponse
} from '../types/notification';

class NotificationService {

  async getNotifications(): Promise<Notification[]> {
    try {
      const response = await apiClient.get<NotificationsResponse>('/users/notify');
      
      if (!response.data?.notifications) {
        return [];
      }

      const sanitized = response.data.notifications
        .map(validateNotificationData)
        .filter(n => n.id > 0);
      
      return sanitized;
    } catch (error: any) {
      console.error('[NotificationService] Ошибка получения уведомлений');
      throw error;
    }
  }

  async getGroupInvites(): Promise<GroupInvite[]> {
    try {
      console.log('[NotificationService] 📨 Загрузка приглашений в группы...');
      const response = await apiClient.get<NotificationsResponse>('/users/notify');
      
      console.log('[NotificationService] 📦 Полный ответ API:', JSON.stringify(response.data, null, 2));
      
      if (!response.data?.invites) {
        console.log('[NotificationService] ⚠️ Нет поля invites в ответе');
        return [];
      }
      
      console.log('[NotificationService] ✅ Получено приглашений:', response.data.invites.length);
      return response.data.invites;
    } catch (error: any) {
      console.error('[NotificationService] ❌ Ошибка получения приглашений:', error);
      return [];
    }
  }

  async checkUnreadNotifications(): Promise<UnreadNotificationsResponse> {
    try {
      const response = await apiClient.get<boolean>('/users/notify/inf');
      return { has_unread: response.data };
    } catch (error: any) {
      console.error('[NotificationService] Ошибка проверки уведомлений:', error);
      return { has_unread: false };
    }
  }

  async markAsViewed(notificationId: number): Promise<MarkAsViewedResponse> {
    try {
      console.log('[NotificationService] Отметка уведомления как просмотренное:', notificationId);
      
      const response = await apiClient.post<MarkAsViewedResponse>(
        '/users/notifications/viewed',
        { id: notificationId }
      );
      
      console.log('[NotificationService] Уведомление отмечено');
      return response.data;
    } catch (error: any) {
      console.error('[NotificationService] Ошибка отметки уведомления:', error);
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

  async updateDeviceToken(token: string, platform: string): Promise<void> {
    try {
      await apiClient.post('/users/device-token', {
        fcm_token: token,
        platform,
      });
      console.log('[NotificationService] ✅ Device token обновлен');
    } catch (error: any) {
      console.error('[NotificationService] ❌ Ошибка обновления токена:', error);
      throw this.handleError(error);
    }
  }
}

export default new NotificationService();