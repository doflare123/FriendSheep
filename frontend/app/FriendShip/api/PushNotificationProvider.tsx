import pushNotificationService from '@/api/services/pushNotificationService';
import * as Notifications from 'expo-notifications';
import React, { useEffect, useRef } from 'react';

interface PushNotificationProviderProps {
  children: React.ReactNode;
  isAuthenticated: boolean;
}

export const PushNotificationProvider: React.FC<PushNotificationProviderProps> = ({ 
  children, 
  isAuthenticated 
}) => {
  const notificationListener = useRef<Notifications.Subscription | undefined>(undefined);
  const responseListener = useRef<Notifications.Subscription | undefined>(undefined);

  useEffect(() => {
    if (isAuthenticated) {
      initializePushNotifications();
    } else {
      notificationListener.current?.remove();
      responseListener.current?.remove();
    }

    return () => {
      notificationListener.current?.remove();
      responseListener.current?.remove();
    };
  }, [isAuthenticated]);

  const initializePushNotifications = async () => {
    try {
      console.log('[PushNotificationProvider] 🔔 Инициализация push-уведомлений...');

      await pushNotificationService.setupBackgroundHandler();

      const expoPushToken = await pushNotificationService.registerForPushNotifications();

      const fcmToken = await pushNotificationService.getFCMToken();

      if (expoPushToken) {
        await pushNotificationService.sendTokenToServer(expoPushToken, fcmToken);
      }

      await pushNotificationService.setupForegroundHandler();

      notificationListener.current = pushNotificationService.addNotificationListener(
        (notification) => {
          console.log('[PushNotificationProvider] 📨 Получено уведомление:', notification);
        }
      );

      responseListener.current = pushNotificationService.addNotificationResponseListener(
        (response) => {
          console.log('[PushNotificationProvider] 👆 Нажатие на уведомление:', response);
          
          const data = response.notification.request.content.data;

          if (data.type === 'group_invite') {
            console.log('[PushNotificationProvider] Переход к приглашениям в группу');
          } else if (data.type === 'event_update') {
            console.log('[PushNotificationProvider] Переход к событию:', data.eventId);
          } else if (data.type === 'notification') {
            console.log('[PushNotificationProvider] Общее уведомление');
          }
        }
      );

      console.log('[PushNotificationProvider] ✅ Push-уведомления успешно инициализированы');
    } catch (error) {
      console.error('[PushNotificationProvider] ❌ Ошибка инициализации push-уведомлений:', error);
    }
  };

  return <>{children}</>;
};