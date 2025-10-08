import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getGenres(accessToken: string): Promise<string> {
  try {
    const response = await axios.get(`${API_URL}/api/sessions/genres`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

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