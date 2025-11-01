import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function viewNotif(accessToken: string, id: number): Promise<void> {
  try {
    await axios.post(`${API_URL}/api/users/notifications/viewed`, {id}, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      }
    });
  } catch (error: any) {
    console.error('Ошибка при получении группы:', error);
    throw error;
  }
}
