import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function joinEvent(accessToken: string, group_id: number, session_id: number): Promise<void> {
  try {
    console.log("ZZZ", group_id, session_id);
    await axios.post(`${API_URL}/api/sessions/join`, {group_id, session_id}, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

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