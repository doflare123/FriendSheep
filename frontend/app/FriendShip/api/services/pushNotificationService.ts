import Constants from 'expo-constants';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldShowBanner: true,
    shouldShowList: true,
    shouldPlaySound: true,
    shouldSetBadge: true,
  }),
});

class PushNotificationService {
  private expoPushToken: string | null = null;
  private fcmToken: string | null = null;

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
        this.expoPushToken = tokenData.data;
        console.log('[PushNotificationService] ✅ Expo Push-токен:', this.expoPushToken);
      } else {
        console.warn('[PushNotificationService] ⚠️ projectId не найден, используется режим разработки');
        tokenData = await Notifications.getDevicePushTokenAsync();
        console.log('[PushNotificationService] 📱 Device Push Token:', tokenData.data);
        this.expoPushToken = JSON.stringify(tokenData.data);
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

      console.log('[PushNotificationService] ✅ Токены готовы:');
      console.log('  📱 Expo Token:', this.expoPushToken);
      console.log('  🔥 FCM Token:', fcmToken);

      return this.expoPushToken;
    } catch (error) {
      console.error('[PushNotificationService] ❌ Ошибка регистрации:', error);
      return null;
    }
  }

  async getFCMToken(): Promise<string | null> {
    try {
      if (Platform.OS === 'android') {
        const { getMessaging, getToken } = require('@react-native-firebase/messaging');
        
        const messaging = getMessaging();
        const token = await getToken(messaging);
        
        this.fcmToken = token;
        console.log('[PushNotificationService] ✅ FCM токен:', token);
        return token;
      }
      return null;
    } catch (error) {
      console.warn('[PushNotificationService] ⚠️ FCM токен недоступен:', error);
      return null;
    }
  }

  async sendTokenToServer(expoPushToken: string, fcmToken?: string | null): Promise<void> {

    console.log('[PushNotificationService] 📝 Токены (готовы к отправке на сервер):');
    console.log('  📱 Expo Token:', expoPushToken);
    console.log('  🔥 FCM Token:', fcmToken || 'не получен');
    console.log('  📲 Platform:', Platform.OS);
    console.log('  📱 Device:', Device.modelName);
    console.log('  🔢 OS Version:', Device.osVersion);
  }

  async removeTokenFromServer(): Promise<void> {
    if (!this.expoPushToken) return;

    try {
      this.expoPushToken = null;
      this.fcmToken = null;
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
        const { getMessaging, onBackgroundMessage } = require('@react-native-firebase/messaging');
        
        const messaging = getMessaging();
        
        onBackgroundMessage(messaging, async (remoteMessage: any) => {
          console.log('[PushNotificationService] 📨 Фоновое уведомление:', remoteMessage);
        });
        
        console.log('[PushNotificationService] ✅ Background handler настроен');
      } catch (error) {
        console.warn('[PushNotificationService] ⚠️ Background handler недоступен:', error);
      }
    }
  }

  async setupForegroundHandler(): Promise<void> {
    if (Platform.OS === 'android') {
      try {
        const { getMessaging, onMessage } = require('@react-native-firebase/messaging');
        
        const messaging = getMessaging();
        
        onMessage(messaging, async (remoteMessage: any) => {
          console.log('[PushNotificationService] 📨 Foreground уведомление:', remoteMessage);

          if (remoteMessage.notification) {
            await Notifications.scheduleNotificationAsync({
              content: {
                title: remoteMessage.notification.title || 'Уведомление',
                body: remoteMessage.notification.body || '',
                data: remoteMessage.data,
              },
              trigger: null,
            });
          }
        });
        
        console.log('[PushNotificationService] ✅ Foreground handler настроен');
      } catch (error) {
        console.warn('[PushNotificationService] ⚠️ Foreground handler недоступен:', error);
      }
    }
  }

  getExpoPushToken(): string | null {
    return this.expoPushToken;
  }

  getFCMTokenSync(): string | null {
    return this.fcmToken;
  }
}

export default new PushNotificationService();