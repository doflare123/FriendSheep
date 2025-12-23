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
  const hasRegistered = useRef(false);

  useEffect(() => {
    if (isAuthenticated && !hasRegistered.current) {
      const timer = setTimeout(() => {
        initializePushNotifications();
      }, 1000);
      
      return () => clearTimeout(timer);
    } else if (!isAuthenticated && hasRegistered.current) {
      cleanupPushNotifications();
    }

    return () => {
      notificationListener.current?.remove();
      responseListener.current?.remove();
    };
  }, [isAuthenticated]);

  const initializePushNotifications = async () => {
    try {
      console.log('[PushNotificationProvider] 🔔 Инициализация push-уведомлений...');

      await pushNotificationService.setupNotificationHandlers();

      const success = await pushNotificationService.registerForPushNotifications();
      
      if (!success) {
        console.warn('[PushNotificationProvider] ⚠️ Не удалось зарегистрировать push-уведомления');
        return;
      }

      hasRegistered.current = true;

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
          } else if (data.type === 'friend_request') {
            console.log('[PushNotificationProvider] Переход к запросам в друзья');
          }
        }
      );

      console.log('[PushNotificationProvider] ✅ Push-уведомления успешно инициализированы');
    } catch (error) {
      console.error('[PushNotificationProvider] ❌ Ошибка инициализации:', error);
    }
  };

  const cleanupPushNotifications = async () => {
    try {
      console.log('[PushNotificationProvider] 🗑️ Очистка push-уведомлений...');

      await pushNotificationService.removeTokenFromServer();

      notificationListener.current?.remove();
      responseListener.current?.remove();
      
      hasRegistered.current = false;
      
      console.log('[PushNotificationProvider] ✅ Push-уведомления очищены');
    } catch (error) {
      console.error('[PushNotificationProvider] ❌ Ошибка очистки:', error);
    }
  };

  return <>{children}</>;
};