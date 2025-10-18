import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function leaveEvent(accessToken: string, session_id: number): Promise<void> {
  try {
    await axios.delete(`${API_URL}/api/sessions/${session_id}/leave`, {
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