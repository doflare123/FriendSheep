import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function delGroup(accessToken: string, groupId: number): Promise<void> {
  try {
    await axios.delete(`${API_URL}/api/admin/groups/${groupId}`, {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при получении группы:', error);
    throw error;
  }
}
