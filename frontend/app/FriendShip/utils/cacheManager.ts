import AsyncStorage from '@react-native-async-storage/async-storage';
import * as FileSystem from 'expo-file-system';

export const clearAppCache = async (): Promise<void> => {
  try {
    console.log('[CacheManager] 🧹 Начало очистки кеша...');

    const keys = await AsyncStorage.getAllKeys();
    const keysToRemove = keys.filter(key => 
      !key.includes('token') && 
      !key.includes('refresh') &&
      !key.includes('user_preferences')
    );
    
    if (keysToRemove.length > 0) {
      await AsyncStorage.multiRemove(keysToRemove);
      console.log(`[CacheManager] ✅ Удалено ${keysToRemove.length} ключей из AsyncStorage`);
    }

    if (FileSystem.cacheDirectory) {
      const cacheDir = FileSystem.cacheDirectory;
      const files = await FileSystem.readDirectoryAsync(cacheDir);
      
      for (const file of files) {
        try {
          await FileSystem.deleteAsync(`${cacheDir}${file}`, { idempotent: true });
        } catch (err) {
          console.log(`[CacheManager] ⚠️ Не удалось удалить файл: ${file}`);
        }
      }
      
      console.log(`[CacheManager] ✅ Очищено ${files.length} файлов из кеша`);
    }

    console.log('[CacheManager] 🎉 Кеш успешно очищен');
  } catch (error) {
    console.error('[CacheManager] ❌ Ошибка очистки кеша:', error);
    throw error;
  }
};

export const getCacheSize = async (): Promise<number> => {
  try {
    let totalSize = 0;

    const keys = await AsyncStorage.getAllKeys();
    for (const key of keys) {
      const value = await AsyncStorage.getItem(key);
      if (value) {
        totalSize += value.length;
      }
    }

    if (FileSystem.cacheDirectory) {
      const cacheDir = FileSystem.cacheDirectory;
      const files = await FileSystem.readDirectoryAsync(cacheDir);
      
      for (const file of files) {
        try {
          const info = await FileSystem.getInfoAsync(`${cacheDir}${file}`);
          if (info.exists && !info.isDirectory) {
            totalSize += info.size || 0;
          }
        } catch (err) {
        }
      }
    }

    return totalSize;
  } catch (error) {
    console.error('[CacheManager] ❌ Ошибка получения размера кеша:', error);
    return 0;
  }
};

export const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
};