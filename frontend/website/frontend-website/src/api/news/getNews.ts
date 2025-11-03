import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getNews(page: number): Promise<any> {
  try {
    const response = await axios.get(`${API_URL}/api/news?page=${page}`);

    return response.data;
  } catch (error: any) {
    console.error('Ошибка при получении новостей:', error);

    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }

    throw error;
  }
}