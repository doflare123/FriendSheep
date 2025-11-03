import axios from 'axios';
import { UpdateProfileRequest } from '@/types/apiTypes';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function editProfile(
  accessToken: string,
  data: UpdateProfileRequest
): Promise<void> {
  try {
    const body = Object.fromEntries(
      Object.entries(data).filter(([_, v]) => v !== undefined)
    );

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
