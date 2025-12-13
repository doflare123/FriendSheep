import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getGenres(accessToken: string): Promise<string> {
  try {
    const response = await axios.get(`${API_URL}/api/sessions/genres`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    const sortedData = response.data.sort((a: string, b: string) => {
      const aIsRussian = /[а-яА-ЯёЁ]/.test(a[0]);
      const bIsRussian = /[а-яА-ЯёЁ]/.test(b[0]);
      
      if (aIsRussian && !bIsRussian) return -1;
      if (!aIsRussian && bIsRussian) return 1;
      
      return a.localeCompare(b, aIsRussian ? 'ru' : 'en');
    });

    return sortedData;
  } catch (error: any) {
    console.error('Ошибка при получении группы:', error);
    
    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }
    
    throw error;
  }
}