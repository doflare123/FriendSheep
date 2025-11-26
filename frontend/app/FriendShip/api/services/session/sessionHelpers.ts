import { getTokens, refreshAccessToken } from '@/api/storage/tokenStorage';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';
import * as FileSystem from 'expo-file-system';
import { CreateSessionData } from './sessionTypes';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

export async function downloadImage(imageUrl: string): Promise<string> {
  const filename = `kinopoisk_${Date.now()}.jpg`;
  const localUri = `${FileSystem.cacheDirectory}${filename}`;
  
  const downloadResult = await FileSystem.downloadAsync(imageUrl, localUri);
  
  if (downloadResult.status !== 200) {
    throw new Error(`Ошибка скачивания: ${downloadResult.status}`);
  }
  
  return localUri;
}

export async function uploadSessionImage(imageUri: string): Promise<string> {
  try {
    let tokens = await getTokens();
    if (!tokens?.accessToken) {
      throw new Error('Пользователь не авторизован');
    }

    console.log('[SessionHelpers] 📸 Загрузка изображения...');

    const formData = new FormData();
    
    const filename = imageUri.split('/').pop() || `session_${Date.now()}.jpg`;
    const fileExtension = filename.split('.').pop()?.toLowerCase() || 'jpg';
    
    formData.append('image', {
      uri: imageUri,
      name: filename,
      type: `image/${fileExtension === 'jpg' ? 'jpeg' : fileExtension}`,
    } as any);

    let response = await fetch(`${BASE_URL}/admin/groups/UploadPhoto`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${tokens.accessToken}`,
      },
      body: formData,
    });

    console.log('[SessionHelpers] ✅ Статус загрузки:', response.status);

    if (response.status === 401) {
      console.log('[SessionHelpers] 🔄 Токен истёк, обновляем...');
      
      try {
        const newAccessToken = await refreshAccessToken();
        if (!newAccessToken) {
          throw new Error('Не удалось обновить токен');
        }

        console.log('[SessionHelpers] ✅ Токен обновлён, повторная попытка...');

        response = await fetch(`${BASE_URL}/admin/groups/UploadPhoto`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${newAccessToken}`,
          },
          body: formData,
        });

        console.log('[SessionHelpers] ✅ Статус после обновления токена:', response.status);
      } catch (refreshError) {
        console.error('[SessionHelpers] ❌ Ошибка обновления токена:', refreshError);
        throw new Error('Не удалось обновить токен. Пожалуйста, войдите снова.');
      }
    }

    if (!response.ok) {
      const errorText = await response.text();
      console.error('[SessionHelpers] ❌ Ошибка загрузки:', errorText);
      
      try {
        const errorData = JSON.parse(errorText);
        throw new Error(errorData.error || errorData.message || 'Ошибка загрузки изображения');
      } catch (parseError) {
        throw new Error('Ошибка загрузки изображения');
      }
    }

    const result = await response.json();
    const imageUrl = result.url || result.image_url || result.image || Object.values(result)[0];
    
    if (!imageUrl || typeof imageUrl !== 'string') {
      throw new Error('Не удалось получить URL изображения');
    }
    
    console.log('[SessionHelpers] ✅ Изображение загружено:', imageUrl);
    return imageUrl as string;
  } catch (error: any) {
    console.error('[SessionHelpers] ❌ Ошибка:', error);
    throw error;
  }
}

function extractCityFromAddress(address: string): string {
  if (!address || !address.trim()) return '';

  const cleaned = address.trim();

  const cityPrefixMatch = cleaned.match(/^(?:г\.\s*|город\s+)([^,]+)/i);
  if (cityPrefixMatch) {
    return cityPrefixMatch[1].trim();
  }

  const firstPart = cleaned.split(',')[0].trim();

  const notCityPrefixes = /^(ул\.|улица|пр\.|проспект|пер\.|переулок|д\.|дом|кв\.|квартира)/i;
  if (!notCityPrefixes.test(firstPart)) {
    return firstPart;
  }
  
  return '';
}

export function buildSessionFormData(
  sessionData: CreateSessionData, 
  imageUrl: string
): FormData {
  const formData = new FormData();

  formData.append('title', sessionData.title);
  formData.append('session_type', sessionData.session_type);
  formData.append('session_place', sessionData.session_place.toString());
  formData.append('group_id', sessionData.group_id.toString());
  formData.append('start_time', sessionData.start_time);
  formData.append('count_users', sessionData.count_users.toString());
  formData.append('image', imageUrl);

  if (sessionData.duration !== undefined && sessionData.duration !== null) {
    formData.append('duration', sessionData.duration.toString());
  }

  if (sessionData.genres && sessionData.genres.trim()) {
    formData.append('genres', sessionData.genres);
  }

  if (sessionData.location && sessionData.location.trim()) {
      formData.append('location', sessionData.location);
      
      if (sessionData.session_place === 2) {
        const city = extractCityFromAddress(sessionData.location);
        if (city) {
          const fieldsValue = `city:${city}`;
          formData.append('fields', fieldsValue);
          console.log('[SessionHelpers] 🏙️ Город добавлен в fields:', fieldsValue);
        }
      }
    }

  if (sessionData.year !== undefined && sessionData.year !== null && sessionData.year > 0) {
    formData.append('year', sessionData.year.toString());
  }

  if (sessionData.country && sessionData.country.trim()) {
    formData.append('country', sessionData.country);
    console.log('[SessionHelpers] 📝 Издатель записан в country:', sessionData.country);
  }

  if (sessionData.age_limit && sessionData.age_limit.trim()) {
    formData.append('age_limit', sessionData.age_limit);
  }

  if (sessionData.notes && sessionData.notes.trim()) {
    formData.append('notes', sessionData.notes);
    console.log('[SessionHelpers] 📝 Описание (notes) добавлено');
  }
  return formData;
}

export function logSessionData(sessionData: CreateSessionData, imageUrl: string): void {
  console.log('[SessionHelpers] 📦 Данные сессии для отправки:');
  console.log('  - title:', sessionData.title);
  console.log('  - session_type:', sessionData.session_type);
  console.log('  - session_place:', sessionData.session_place);
  console.log('  - group_id:', sessionData.group_id);
  console.log('  - start_time:', sessionData.start_time);
  console.log('  - count_users:', sessionData.count_users);
  console.log('  - image:', imageUrl);
  console.log('  - duration:', sessionData.duration);
  console.log('  - genres:', sessionData.genres);
  console.log('  - location:', sessionData.location);
  console.log('  - year:', sessionData.year);
  console.log('  - country (издатель):', sessionData.country);
  console.log('  - age_limit:', sessionData.age_limit);
  console.log('  - notes (описание):', sessionData.notes);
  console.log('  - fields:', sessionData.fields);
}