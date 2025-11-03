import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getNotif(accessToken: string): Promise<any> {
  try {

    const response = await axios.get(`${API_URL}/api/users/notify`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      }
    });

    //console.log("TEST:", response.data.)

    return response.data;
  } catch (error: any) {
    console.error('Ошибка при получении группы:', error);
    throw error;
  }
}
