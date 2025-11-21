import { getTokens } from '@/api/storage/tokenStorage';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';
import { CreateSessionData } from './sessionTypes';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

export async function uploadSessionImage(imageUri: string): Promise<string> {
  try {
    const tokens = await getTokens();
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

    const response = await fetch(`${BASE_URL}/admin/groups/UploadPhoto`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${tokens.accessToken}`,
      },
      body: formData,
    });

    console.log('[SessionHelpers] ✅ Статус загрузки:', response.status);

    if (!response.ok) {
      const errorText = await response.text();
      console.error('[SessionHelpers] ❌ Ошибка загрузки:', errorText);
      throw new Error('Ошибка загрузки изображения');
    }

    const result = await response.json();
    const imageUrl = result.url || result.image_url || result.image || Object.values(result)[0];
    
    if (!imageUrl || typeof imageUrl !== 'string') {
      throw new Error('Не удалось получить URL изображения');
    }
    
    return imageUrl as string;
  } catch (error: any) {
    console.error('[SessionHelpers] ❌ Ошибка:', error);
    throw error;
  }
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

  if (sessionData.duration !== undefined) {
    formData.append('duration', sessionData.duration.toString());
  }

  if (sessionData.genres) {
    formData.append('genres', sessionData.genres);
  }

  if (sessionData.fields) {
    formData.append('fields', sessionData.fields);
  }

  if (sessionData.location) {
    formData.append('location', sessionData.location);
  }

  if (sessionData.year !== undefined) {
    formData.append('year', sessionData.year.toString());
  }

  if (sessionData.country) {
    formData.append('country', sessionData.country);
  }

  if (sessionData.age_limit) {
    formData.append('age_limit', sessionData.age_limit);
  }

  if (sessionData.notes) {
    formData.append('notes', sessionData.notes);
  }

  return formData;
}

export function logSessionData(sessionData: CreateSessionData, imageUrl: string): void {
  console.log('[SessionHelpers] 📦 Данные сессии:');
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
  console.log('  - country:', sessionData.country);
  console.log('  - age_limit:', sessionData.age_limit);
  console.log('  - notes:', sessionData.notes);
}