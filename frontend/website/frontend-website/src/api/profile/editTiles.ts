import axios from 'axios';
import {Counters} from '@/types/apiTypes';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function editTiles(accessToken: string, tiles: Counters): Promise<void> {
  try {

    console.log("ZZZZZZZZZZ", tiles)

    await axios.patch(`${API_URL}/api/users/tiles`, tiles, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      }
    });
  } catch (error: any) {
    console.error('Ошибка при получении группы:', error);
    throw error;
  }
}
