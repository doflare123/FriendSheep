// src/api/user.ts

import axios from 'axios';
import { UserDataResponse } from '../../types/UserData';
import { EventCardProps } from '../../types/Events';
import { RawSession } from '../../types/RawEvents';
import { mapServerSessionToEvent, mapRawUserDataToUserData } from '../../Constants';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getOtherUserInfo(accessToken: string, id: string): Promise<UserDataResponse> {
  try {
    const response = await axios.get(`${API_URL}/api/users/inf/${id}`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    const data = response.data;

    // Конвертация сырых данных
    const recent_sessions: EventCardProps[] = data.recent_sessions.map((s: RawSession) =>
      mapServerSessionToEvent(s)
    );

    const upcoming_sessions: EventCardProps[] = data.upcoming_sessions.map((s: RawSession) =>
      mapServerSessionToEvent(s)
    );

    return mapRawUserDataToUserData({...data, recent_sessions, upcoming_sessions});
  } catch (error: any) {
    console.error('Ошибка при получении информации о пользователе:', error);

    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }

    throw error;
  }
}
