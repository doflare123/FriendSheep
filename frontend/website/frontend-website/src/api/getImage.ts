import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getImage(accessToken: string, image: File): Promise<string> {
  try {
    const data = new FormData();

    data.append('image', image);

    const response = await axios.post(`${API_URL}/api/admin/groups/UploadPhoto`, data, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      }
    });

    return response.data.image;
  } catch (error: any) {
    console.error('Ошибка при получении группы:', error);
    throw error;
  }
}
