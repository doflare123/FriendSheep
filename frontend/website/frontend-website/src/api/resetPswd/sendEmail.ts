import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function sendEmail(email: string): Promise<string> {
  try {
    const response = await axios.post(`${API_URL}/api/users/request-reset`, {
      email,
    });

    return response.data.session_id;
  } catch (error: any) {
    console.error('Ошибка при регистрации сессии:', error);
    throw error;
  }
}
