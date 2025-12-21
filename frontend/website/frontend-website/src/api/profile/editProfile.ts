import axios from 'axios';
import { UpdateProfileRequest } from '@/types/apiTypes';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function editProfile(
  accessToken: string,
  data: UpdateProfileRequest
): Promise<void> {
  try {
    
    // Убираем пустые поля (undefined, null, пустые строки)
    const body = Object.fromEntries(
      Object.entries(data).filter(([_, v]) => {
        return v !== undefined && v !== null && v !== '';
      })
    );

    // Если нет полей для обновления, не выполняем запрос
    if (Object.keys(body).length === 0) {
      console.log('Нет данных для обновления профиля');
      return;
    }

    await axios.patch(`${API_URL}/api/users/user/profile`, body, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });
  } catch (error: any) {
    console.error('Ошибка при обновлении профиля:', error);
    throw error;
  }
}