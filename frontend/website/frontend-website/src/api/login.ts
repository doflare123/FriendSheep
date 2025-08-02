import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function login(email: string, password: string): Promise<any> {
  try {
    const response = await axios.post(`${API_URL}/api/users/login`, {
      email,
      password
    });

    return response.data;
  } catch (error: any) {
    console.error('Ошибка при авторизации:', error);
    throw error;
  }
}
