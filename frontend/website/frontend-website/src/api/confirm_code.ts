import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function confirm_code(code: string, session_id: string, type: string): Promise<any> {
  try {
    console.log(code, session_id, type);
    const response = await axios.patch(`${API_URL}/api/sessions/verify`, {
      code,
      session_id,
      type
    });

    return response.data;
  } catch (error: any) {
    console.error('Ошибка при подтверждении кода:', error);
    throw error;
  }
}
