import axios from 'axios';
import {searchEvents} from '@/api/search/searchEvents'
import {EventCardProps} from '@/types/Events'
import {convertSessionsToEventCards} from '@/Constants'

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getGenreEvents(accessToken: string, genre: number): Promise<EventCardProps[]> {
  try {
    const response = await searchEvents(accessToken, 1, {categoryID: genre})

    return convertSessionsToEventCards(response.sessions);
  } catch (error: any) {
    console.error('Ошибка при получении событий из групп с жанром :', error);
    
    // Более подробная информация об ошибке для отладки
    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }
    
    throw error;
  }
}