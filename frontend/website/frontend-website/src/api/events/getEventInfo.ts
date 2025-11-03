import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getEventInfo(accessToken: string, sessionId: number): Promise<any> {
  try {
    const response = await axios.get(`${API_URL}/api/users/sessions/${sessionId}`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    console.log(response.data);
    return response.data;
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