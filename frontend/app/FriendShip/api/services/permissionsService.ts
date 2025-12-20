import AsyncStorage from '@react-native-async-storage/async-storage';
import * as ImagePicker from 'expo-image-picker';
import * as Location from 'expo-location';
import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';

const PERMISSIONS_REQUESTED_KEY = '@permissions_requested';

export type PermissionType = 'media' | 'notifications' | 'location';

export interface PermissionsStatus {
  media: boolean;
  notifications: boolean;
  location: boolean;
}

class PermissionsService {

  async hasRequestedPermissions(): Promise<boolean> {
    try {
      const requested = await AsyncStorage.getItem(PERMISSIONS_REQUESTED_KEY);
      return requested === 'true';
    } catch (error) {
      console.error('[PermissionsService] Ошибка проверки флага разрешений:', error);
      return false;
    }
  }

  async setPermissionsRequested(): Promise<void> {
    try {
      await AsyncStorage.setItem(PERMISSIONS_REQUESTED_KEY, 'true');
    } catch (error) {
      console.error('[PermissionsService] Ошибка установки флага разрешений:', error);
    }
  }

  async resetPermissionsFlag(): Promise<void> {
    try {
      await AsyncStorage.removeItem(PERMISSIONS_REQUESTED_KEY);
      console.log('[PermissionsService] Флаг разрешений сброшен');
    } catch (error) {
      console.error('[PermissionsService] Ошибка сброса флага:', error);
    }
  }

  async checkMediaPermission(): Promise<boolean> {
    try {
      const { status } = await ImagePicker.getMediaLibraryPermissionsAsync();
      return status === 'granted';
    } catch (error) {
      console.error('[PermissionsService] Ошибка проверки медиа разрешения:', error);
      return false;
    }
  }

  async requestMediaPermission(): Promise<boolean> {
    try {
      console.log('[PermissionsService] 📷 Запрос разрешения на медиатеку...');
      const { status } = await ImagePicker.requestMediaLibraryPermissionsAsync();
      const granted = status === 'granted';
      console.log('[PermissionsService] Медиатека:', granted ? '✅ Разрешено' : '❌ Отклонено');
      return granted;
    } catch (error) {
      console.error('[PermissionsService] Ошибка запроса медиа разрешения:', error);
      return false;
    }
  }

  async checkCameraPermission(): Promise<boolean> {
    try {
      const { status } = await ImagePicker.getCameraPermissionsAsync();
      return status === 'granted';
    } catch (error) {
      console.error('[PermissionsService] Ошибка проверки камеры:', error);
      return false;
    }
  }

  async requestCameraPermission(): Promise<boolean> {
    try {
      console.log('[PermissionsService] 📸 Запрос разрешения на камеру...');
      const { status } = await ImagePicker.requestCameraPermissionsAsync();
      const granted = status === 'granted';
      console.log('[PermissionsService] Камера:', granted ? '✅ Разрешено' : '❌ Отклонено');
      return granted;
    } catch (error) {
      console.error('[PermissionsService] Ошибка запроса камеры:', error);
      return false;
    }
  }

  async checkNotificationsPermission(): Promise<boolean> {
    try {
      const { status } = await Notifications.getPermissionsAsync();
      return status === 'granted';
    } catch (error) {
      console.error('[PermissionsService] Ошибка проверки уведомлений:', error);
      return false;
    }
  }

  async requestNotificationsPermission(): Promise<boolean> {
    try {
      console.log('[PermissionsService] 🔔 Запрос разрешения на уведомления...');
      
      const { status: existingStatus } = await Notifications.getPermissionsAsync();
      let finalStatus = existingStatus;

      if (existingStatus !== 'granted') {
        const { status } = await Notifications.requestPermissionsAsync();
        finalStatus = status;
      }

      const granted = finalStatus === 'granted';
      console.log('[PermissionsService] Уведомления:', granted ? '✅ Разрешено' : '❌ Отклонено');

      if (granted && Platform.OS === 'android') {
        await Notifications.setNotificationChannelAsync('default', {
          name: 'default',
          importance: Notifications.AndroidImportance.MAX,
          vibrationPattern: [0, 250, 250, 250],
          lightColor: '#5DADE2',
        });
      }

      return granted;
    } catch (error) {
      console.error('[PermissionsService] Ошибка запроса уведомлений:', error);
      return false;
    }
  }

  async checkLocationPermission(): Promise<boolean> {
    try {
      const { status } = await Location.getForegroundPermissionsAsync();
      return status === 'granted';
    } catch (error) {
      console.error('[PermissionsService] Ошибка проверки геолокации:', error);
      return false;
    }
  }

  async requestLocationPermission(): Promise<boolean> {
    try {
      console.log('[PermissionsService] 📍 Запрос разрешения на геолокацию...');
      const { status } = await Location.requestForegroundPermissionsAsync();
      const granted = status === 'granted';
      console.log('[PermissionsService] Геолокация:', granted ? '✅ Разрешено' : '❌ Отклонено');
      return granted;
    } catch (error) {
      console.error('[PermissionsService] Ошибка запроса геолокации:', error);
      return false;
    }
  }

  async requestInitialPermissions(): Promise<PermissionsStatus> {
    console.log('[PermissionsService] 🚀 Запрос начальных разрешений...');
    
    const mediaGranted = await this.requestMediaPermission();
    const notificationsGranted = await this.requestNotificationsPermission();

    await this.setPermissionsRequested();

    const status: PermissionsStatus = {
      media: mediaGranted,
      notifications: notificationsGranted,
      location: false,
    };

    console.log('[PermissionsService] ✅ Начальные разрешения запрошены:', status);
    return status;
  }

  async checkAllPermissions(): Promise<PermissionsStatus> {
    const media = await this.checkMediaPermission();
    const notifications = await this.checkNotificationsPermission();
    const location = await this.checkLocationPermission();

    return {
      media,
      notifications,
      location,
    };
  }
}

export default new PermissionsService();