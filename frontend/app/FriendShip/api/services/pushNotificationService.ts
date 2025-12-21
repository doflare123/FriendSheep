import apiClient from '@/api/apiClient';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';

// Настройка обработчика уведомлений
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowBanner: true,
    shouldShowList: true,
    shouldPlaySound: true,
    shouldSetBadge: true,
  }),
});

class PushNotificationService {
  private fcmToken: string | null = null;

  /**
   * Запросить разрешения и зарегистрировать устройство
   */
  async registerForPushNotifications(): Promise<boolean> {
    if (!Device.isDevice) {
      console.warn('[PushNotificationService] Push-уведомления работают только на реальных устройствах');
      return false;
    }

    try {
      // 1. Запросить разрешения
      const { status: existingStatus } = await Notifications.getPermissionsAsync();
      let finalStatus = existingStatus;

      if (existingStatus !== 'granted') {
        const { status } = await Notifications.requestPermissionsAsync();
        finalStatus = status;
      }

      if (finalStatus !== 'granted') {
        console.warn('[PushNotificationService] Разрешение на уведомления не получено');
        return false;
      }

      // 2. Настройка канала для Android
      if (Platform.OS === 'android') {
        await Notifications.setNotificationChannelAsync('default', {
          name: 'По умолчанию',
          importance: Notifications.AndroidImportance.MAX,
          vibrationPattern: [0, 250, 250, 250],
          lightColor: '#5DADE2',
          sound: 'default',
          enableVibrate: true,
          showBadge: true,
        });
      }

      // 3. Получить FCM токен
      const fcmToken = await this.getFCMToken();
      
      if (!fcmToken) {
        console.warn('[PushNotificationService] Не удалось получить FCM токен');
        return false;
      }

      // 4. Отправить токен на сервер
      await this.sendTokenToServer(fcmToken);

      console.log('[PushNotificationService] ✅ Push-уведомления зарегистрированы');
      return true;
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка регистрации:', error);
      return false;
    }
  }

  /**
   * Получить FCM токен
   */
  async getFCMToken(): Promise<string | null> {
    try {
      if (Platform.OS === 'android') {
        const { getMessaging, getToken } = require('@react-native-firebase/messaging');
        
        const messaging = getMessaging();
        const token = await getToken(messaging);
        
        this.fcmToken = token;
        console.log('[PushNotificationService] ✅ FCM токен получен');
        return token;
      }
      
      // Для iOS тоже можно получить FCM токен
      if (Platform.OS === 'ios') {
        const { getMessaging, getToken } = require('@react-native-firebase/messaging');
        
        const messaging = getMessaging();
        const token = await getToken(messaging);
        
        this.fcmToken = token;
        console.log('[PushNotificationService] ✅ FCM токен получен (iOS)');
        return token;
      }
      
      return null;
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка получения FCM токена:', error);
      return null;
    }
  }

  /**
   * Отправить токен на сервер
   */
  async sendTokenToServer(fcmToken: string): Promise<void> {
    try {
      const deviceInfo = {
        model: Device.modelName,
        os_version: Device.osVersion,
        brand: Device.brand,
        manufacturer: Device.manufacturer,
      };

      const response = await apiClient.post('/api/device-tokens/register', {
        device_token: fcmToken,
        platform: Platform.OS,
        device_info: JSON.stringify(deviceInfo),
      });

      console.log('[PushNotificationService] ✅ Токен отправлен на сервер:', response.data);
    } catch (error: any) {
      console.error('[PushNotificationService] ❌ Ошибка отправки токена:', error);
      
      if (error.response?.status === 401) {
        console.error('Пользователь не авторизован');
      } else if (error.response?.status === 400) {
        console.error('Некорректные данные:', error.response.data);
      }
      
      throw error;
    }
  }

  /**
   * Удалить токен с сервера (при выходе)
   */
  async removeTokenFromServer(): Promise<void> {
    if (!this.fcmToken) {
      console.warn('[PushNotificationService] Нет токена для удаления');
      return;
    }

    try {
      await apiClient.delete('/api/device-tokens', {
        params: { device_token: this.fcmToken },
      });
      
      console.log('[PushNotificationService] ✅ Токен удален с сервера');
      this.fcmToken = null;
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка удаления токена:', error);
    }
  }

  /**
   * Деактивировать токен (временно отключить уведомления)
   */
  async deactivateToken(): Promise<void> {
    if (!this.fcmToken) return;

    try {
      await apiClient.post('/api/device-tokens/deactivate', {
        device_token: this.fcmToken,
      });
      
      console.log('[PushNotificationService] ✅ Токен деактивирован');
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка деактивации токена:', error);
    }
  }

  /**
   * Настроить обработчики foreground/background уведомлений
   */
  async setupNotificationHandlers(): Promise<void> {
    if (Platform.OS === 'android' || Platform.OS === 'ios') {
      try {
        const { getMessaging, onMessage } = require('@react-native-firebase/messaging');
        
        const messaging = getMessaging();
        
        // Foreground handler
        onMessage(messaging, async (remoteMessage: any) => {
          console.log('[PushNotificationService] 📨 Foreground уведомление:', remoteMessage);
          
          // Показываем локальное уведомление
          if (remoteMessage.notification) {
            await Notifications.scheduleNotificationAsync({
              content: {
                title: remoteMessage.notification.title || 'Уведомление',
                body: remoteMessage.notification.body || '',
                data: remoteMessage.data || {},
              },
              trigger: null,
            });
          }
        });
        
        console.log('[PushNotificationService] ✅ Notification handlers настроены');
      } catch (error) {
        console.warn('[PushNotificationService] ⚠️ Не удалось настроить handlers:', error);
      }
    }
  }

  /**
   * Badge count
   */
  async getBadgeCount(): Promise<number> {
    return await Notifications.getBadgeCountAsync();
  }

  async setBadgeCount(count: number): Promise<void> {
    await Notifications.setBadgeCountAsync(count);
  }

  /**
   * Локальные уведомления
   */
  async sendLocalNotification(title: string, body: string, data?: any): Promise<void> {
    await Notifications.scheduleNotificationAsync({
      content: { title, body, data: data || {} },
      trigger: null,
    });
  }

  async clearAllNotifications(): Promise<void> {
    await Notifications.dismissAllNotificationsAsync();
  }

  /**
   * Слушатели
   */
  addNotificationListener(
    callback: (notification: Notifications.Notification) => void
  ): Notifications.Subscription {
    return Notifications.addNotificationReceivedListener(callback);
  }

  addNotificationResponseListener(
    callback: (response: Notifications.NotificationResponse) => void
  ): Notifications.Subscription {
    return Notifications.addNotificationResponseReceivedListener(callback);
  }

  /**
   * Получить сохраненный FCM токен
   */
  getFCMTokenSync(): string | null {
    return this.fcmToken;
  }
}

export default new PushNotificationService();