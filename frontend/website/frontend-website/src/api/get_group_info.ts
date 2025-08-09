import axios from 'axios';
import { GroupData } from '../types/Groups';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getGroupInfo(accessToken: string, groupId: number): Promise<GroupData> {
  try {
    const response = await axios.get(`${API_URL}/api/groups/${groupId}`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    return response.data as GroupData;
  } catch (error: any) {
    console.error('Ошибка при получении группы:', error);
    
    // Более подробная информация об ошибке для отладки
    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }
    
    throw error;
  }
}