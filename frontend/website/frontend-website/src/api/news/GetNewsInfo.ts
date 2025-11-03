import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getNewsInfo(id: number): Promise<any> {
  try {
    const response = await axios.get(`${API_URL}/api/news/${id}`);

    return response.data;
  } catch (error: any) {
    console.error('Ошибка при получении новости:', error);

    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }

    throw error;
  }
}