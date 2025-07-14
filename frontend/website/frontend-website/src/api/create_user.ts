import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function createUser(email: string, name: string, password: string, session_id: string): Promise<void> {
  try {
    await axios.post(`${API_URL}/api/users`, {
      email,
      name,
      password,
      session_id
    });
  } catch (error: any) {
    console.error('Ошибка при регистрации сессии:', error);
    throw error;
  }
}
