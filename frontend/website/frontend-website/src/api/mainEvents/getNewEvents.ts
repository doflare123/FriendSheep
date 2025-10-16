import axios from 'axios';
import {searchEvents} from '@/api/search/searchEvents'
import {EventCardProps} from '@/types/Events'
import {convertSessionsToEventCards} from '@/Constants'

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getNewEvents(accessToken: string): Promise<EventCardProps[]> {
  try {
    const response = await searchEvents(accessToken, 1, {new_only: true})

    return convertSessionsToEventCards(response.sessions);
  } catch (error: any) {
    console.error('Ошибка при получении новых событий :', error);
    
    // Более подробная информация об ошибке для отладки
    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }
    
    throw error;
  }
}