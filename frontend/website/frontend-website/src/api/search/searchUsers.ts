import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

interface SearchEventsParams {
  name?: string;
}

export async function searchUsers(
  accessToken: string, 
  page: number,
  params: SearchEventsParams = {}
): Promise<any> {
  try {
    // Добавляем page к параметрам
    const allParams = {
      ...params,
      page: page
    };

    // Убираем undefined значения и приводим к строкам
    const filteredData = Object.fromEntries(
      Object.entries(allParams)
        .filter(([_, v]) => v !== undefined)
        .map(([k, v]) => [k, String(v)])
    );

    const queryString = new URLSearchParams(filteredData).toString();

    const response = await axios.get(`${API_URL}/api/users/search?${queryString}`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    return response.data;
  } catch (error: any) {
    console.error('Ошибка при получении результатов поиска групп:', error);
    
    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }
    
    throw error;
  }
}