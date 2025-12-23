import apiClient from '@/api/apiClient';
import messaging from '@react-native-firebase/messaging';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';

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

  async registerForPushNotifications(): Promise<boolean> {
    if (!Device.isDevice) {
      console.warn('[PushNotificationService] Push-уведомления работают только на реальных устройствах');
      return false;
    }

    try {
      if (Platform.OS === 'ios') {
        const authStatus = await messaging().requestPermission();
        const enabled =
          authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
          authStatus === messaging.AuthorizationStatus.PROVISIONAL;

        if (!enabled) {
          console.warn('[PushNotificationService] iOS: разрешение Firebase не получено');
          return false;
        }
      }

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

      const fcmToken = await this.getFCMToken();
      
      if (!fcmToken) {
        console.warn('[PushNotificationService] Не удалось получить FCM токен');
        return false;
      }

      await this.sendTokenToServer(fcmToken);

      console.log('[PushNotificationService] ✅ Push-уведомления зарегистрированы');
      return true;
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка регистрации:', error);
      return false;
    }
  }

  async getFCMToken(): Promise<string | null> {
    try {
      const token = await messaging().getToken();
      
      this.fcmToken = token;
      console.log('[PushNotificationService] ✅ FCM токен получен:', token.substring(0, 30) + '...');
      return token;
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка получения FCM токена:', error);
      return null;
    }
  }

  async sendTokenToServer(fcmToken: string): Promise<void> {
    try {
      const deviceInfo = {
        model: Device.modelName || 'Unknown',
        os_version: Device.osVersion || 'Unknown',
        brand: Device.brand || 'Unknown',
        manufacturer: Device.manufacturer || 'Unknown',
      };

      console.log('[PushNotificationService] 📤 Отправка токена на сервер...');
      console.log('[PushNotificationService] Токен:', fcmToken.substring(0, 40) + '...');

      const response = await apiClient.post('/device-tokens/register', {
        device_token: fcmToken,
        platform: Platform.OS,
        device_info: JSON.stringify(deviceInfo),
      });

      console.log('[PushNotificationService] ✅ Токен отправлен на сервер:', response.data);
    } catch (error: any) {
      console.error('[PushNotificationService] ❌ Ошибка отправки токена:', error);
      
      if (error.response?.status === 401) {
        console.error('❌ Пользователь не авторизован - токен будет отправлен после логина');
      } else if (error.response?.status === 404) {
        console.error('❌ Endpoint не найден:', error.config?.url);
        console.error('❌ Полный URL:', error.config?.baseURL + error.config?.url);
        console.error('❌ Убедитесь что бэкенд запущен и endpoint существует');
      } else if (error.response?.status === 400) {
        console.error('❌ Некорректные данные:', error.response.data);
      } else if (error.code === 'ECONNREFUSED') {
        console.error('❌ Не удается подключиться к серверу');
        console.error('❌ Проверьте что бэкенд запущен на:', error.config?.baseURL);
      } else {
        console.error('❌ Неизвестная ошибка:', error.message);
      }
      
      throw error;
    }
  }

  async removeTokenFromServer(): Promise<void> {
    if (!this.fcmToken) {
      console.log('[PushNotificationService] ℹ️ Нет токена для удаления');
      return;
    }

    try {
      await apiClient.delete('/device-tokens', {
        params: { device_token: this.fcmToken },
      });
      
      console.log('[PushNotificationService] ✅ Токен удален с сервера');

      await messaging().deleteToken();
      this.fcmToken = null;
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка удаления токена:', error);
    }
  }

  async deactivateToken(): Promise<void> {
    if (!this.fcmToken) return;

    try {
      await apiClient.post('/device-tokens/deactivate', {
        device_token: this.fcmToken,
      });
      
      console.log('[PushNotificationService] ✅ Токен деактивирован');
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка деактивации токена:', error);
    }
  }

  async setupNotificationHandlers(): Promise<void> {
    try {
      messaging().onMessage(async (remoteMessage) => {
        console.log('[PushNotificationService] 📨 Foreground уведомление:', remoteMessage);

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

      messaging().onTokenRefresh(async (newToken) => {
        console.log('[PushNotificationService] 🔄 Токен обновлен:', newToken.substring(0, 30) + '...');
        this.fcmToken = newToken;
        try {
          await this.sendTokenToServer(newToken);
        } catch (error) {
          console.error('[PushNotificationService] ❌ Не удалось обновить токен на сервере:', error);
        }
      });
      
      console.log('[PushNotificationService] ✅ Notification handlers настроены');
    } catch (error) {
      console.warn('[PushNotificationService] ⚠️ Не удалось настроить handlers:', error);
    }
  }

  async getBadgeCount(): Promise<number> {
    return await Notifications.getBadgeCountAsync();
  }

  async setBadgeCount(count: number): Promise<void> {
    await Notifications.setBadgeCountAsync(count);
  }

  async sendLocalNotification(title: string, body: string, data?: any): Promise<void> {
    await Notifications.scheduleNotificationAsync({
      content: { title, body, data: data || {} },
      trigger: null,
    });
  }

  async clearAllNotifications(): Promise<void> {
    await Notifications.dismissAllNotificationsAsync();
  }

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

  getFCMTokenSync(): string | null {
    return this.fcmToken;
  }
}

export default new PushNotificationService();