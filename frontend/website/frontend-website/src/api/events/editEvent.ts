import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function editEvent(
  accessToken: string,
  sessionId: number,
  year: number,
  count_users_max?: number,
  duration?: number,
  image_url?: string,
  start_time?: string,
  title?: string,
  age_limit?: string,
  genres?: string[],
  location?: string,
  session_place_id?: number,
  session_type_id?: number,
): Promise<void> {
  try {
    // Собираем объект с возможными полями
    const data = {
      count_users_max,
      duration,
      image_url,
      start_time,
      title,
      year,
      age_limit,
      genres,
      location,
      session_place_id,
      session_type_id,
    };

    // Убираем undefined значения
    const filteredData = Object.fromEntries(
      Object.entries(data).filter(([_, v]) => v !== undefined)
    );

    console.log("Отправляемые данные:", filteredData);

    await axios.patch(`${API_URL}/api/admin/sessions/${sessionId}`, filteredData, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });
  } catch (error: any) {
    console.error('Ошибка при редактировании события:', error);

    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }

    throw error;
  }
}
