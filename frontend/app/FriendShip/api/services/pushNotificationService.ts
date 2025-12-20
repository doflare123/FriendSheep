import apiClient from '@/api/apiClient';
import Constants from 'expo-constants';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: false,
    shouldPlaySound: true,
    shouldSetBadge: true,
    shouldShowBanner: true,
    shouldShowList: true,
  }),
});

class PushNotificationService {
  private expoPushToken: string | null = null;

  async registerForPushNotifications(): Promise<string | null> {
    if (!Device.isDevice) {
      console.warn('[PushNotificationService] Push-уведомления работают только на реальных устройствах');
      return null;
    }

    try {
      const { status: existingStatus } = await Notifications.getPermissionsAsync();
      let finalStatus = existingStatus;

      if (existingStatus !== 'granted') {
        const { status } = await Notifications.requestPermissionsAsync();
        finalStatus = status;
      }

      if (finalStatus !== 'granted') {
        console.warn('[PushNotificationService] Разрешение на уведомления не получено');
        return null;
      }

      const projectId = Constants.expoConfig?.extra?.eas?.projectId;

      let tokenData;
      if (projectId) {
        tokenData = await Notifications.getExpoPushTokenAsync({
          projectId: projectId,
        });
      } else {
        console.warn('[PushNotificationService] ⚠️ projectId не найден, используется режим разработки');
        tokenData = await Notifications.getDevicePushTokenAsync();
        console.log('[PushNotificationService] 📱 Device Push Token:', tokenData.data);
        this.expoPushToken = JSON.stringify(tokenData.data);
        return this.expoPushToken;
      }

      this.expoPushToken = tokenData.data;
      console.log('[PushNotificationService] ✅ Expo Push-токен:', this.expoPushToken);

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

      return this.expoPushToken;
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка регистрации:', error);
      return null;
    }
  }

  async getFCMToken(): Promise<string | null> {
    try {
      if (Platform.OS === 'android') {
        const messaging = require('@react-native-firebase/messaging');
        if (messaging && messaging.default) {
          const fcmToken = await messaging.default().getToken();
          console.log('[PushNotificationService] ✅ FCM токен:', fcmToken);
          return fcmToken;
        } else {
          console.warn('[PushNotificationService] ⚠️ Firebase messaging недоступен (требуется development build)');
          return null;
        }
      }
      return null;
    } catch (error) {
      console.warn('[PushNotificationService] ⚠️ FCM токен недоступен в Expo Go:', error);
      return null;
    }
  }

  async sendTokenToServer(expoPushToken: string, fcmToken?: string | null): Promise<void> {
    try {
      await apiClient.post('/users/device-token', {
        expo_token: expoPushToken,
        fcm_token: fcmToken,
        platform: Platform.OS,
        device_model: Device.modelName,
        device_os_version: Device.osVersion,
      });
      console.log('[PushNotificationService] ✅ Токен отправлен на сервер');
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка отправки токена:', error);
    }
  }

  async removeTokenFromServer(): Promise<void> {
    if (!this.expoPushToken) return;

    try {
      await apiClient.delete('/users/device-token', {
        data: { token: this.expoPushToken },
      });
      console.log('[PushNotificationService] ✅ Токен удален с сервера');
      this.expoPushToken = null;
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка удаления токена:', error);
    }
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

  async getBadgeCount(): Promise<number> {
    return await Notifications.getBadgeCountAsync();
  }

  async setBadgeCount(count: number): Promise<void> {
    await Notifications.setBadgeCountAsync(count);
  }

  async clearAllNotifications(): Promise<void> {
    await Notifications.dismissAllNotificationsAsync();
  }

  async setupBackgroundHandler(): Promise<void> {
    if (Platform.OS === 'android') {
      try {
        const messaging = require('@react-native-firebase/messaging');
        if (messaging && messaging.default) {
          messaging.default().setBackgroundMessageHandler(async (remoteMessage: any) => {
            console.log('[PushNotificationService] 📨 Фоновое уведомление:', remoteMessage);
          });
          console.log('[PushNotificationService] ✅ Background handler настроен');
        } else {
          console.warn('[PushNotificationService] ⚠️ Firebase messaging недоступен (требуется development build)');
        }
      } catch (error) {
        console.warn('[PushNotificationService] ⚠️ Background handler недоступен в Expo Go');
      }
    }
  }
}

export default new PushNotificationService();