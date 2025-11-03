import axios from 'axios';
import {convertSessionsToEventCards} from '@/Constants'
import {EventCardProps} from '@/types/Events'

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getPopular(): Promise<EventCardProps[]> {
  try {
    const response = await axios.get(`${API_URL}/api/users/sessions/popular`);

    return convertSessionsToEventCards(response.data.sessions);
  } catch (error: any) {
    console.error('Ошибка при получении популярных событий:', error);
    
    // Более подробная информация об ошибке для отладки
    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }
    
    throw error;
  }
}