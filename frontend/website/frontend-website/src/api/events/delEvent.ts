import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function delEvent(
  accessToken: string,
  sessionId: number,
): Promise<void> {
  try {

    await axios.delete(`${API_URL}/api/sessions/sessions/${sessionId}`, {
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
