import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function addEvent(accessToken: string, 
    title: string,
    session_type: string,
    session_place: number,
    group_id: number,
    start_time: string,
    count_users: number,
    duration?: number,
    genres?: string,
    fields?: string,
    location?: string,
    year?: number,
    country?: string,
    age_limit?: string,
    notes?: string,
    image?: File,
): Promise<void> {
  try {
    const data = new FormData();

    data.append('title', title);
    data.append('session_type', session_type);
    data.append('session_place', session_place.toString());
    data.append('group_id', group_id.toString());
    data.append('start_time', start_time);
    data.append('count_users', count_users.toString());

    if (duration) data.append('duration', duration.toString());
    if (genres) data.append('genres', genres);
    if (fields) data.append('fields', fields);
    if (location) data.append('location', location);
    if (year) data.append('year', year.toString());
    if (country) data.append('country', country);
    if (age_limit) data.append('age_limit', age_limit);
    if (notes) data.append('notes', notes);
    if (image) data.append('image', image);

    await axios.post(`${API_URL}/api/sessions/createSession`, data, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    });

  } catch (error: any) {
    console.error('Ошибка при добавлении события:', error);

    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }

    throw error;
  }
}