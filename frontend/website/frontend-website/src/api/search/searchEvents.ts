import axios from 'axios';
import {EventCardProps} from '@/types/Events'

const API_URL = process.env.NEXT_PUBLIC_API_URL;

interface SearchEventsParams {
  query?: string;
  categoryID?: number;
  sessionType?: string;
  city?: string;
  sort_by?: string;
  order?: string;
  new_only?: boolean;
}

export async function searchEvents(
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

    const response = await axios.get(`${API_URL}/api/users/sessions/search?${queryString}`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    return response.data;
  } catch (error: any) {
    console.error('Ошибка при получении событий из подписанных групп:', error);
    
    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }
    
    throw error;
  }
}