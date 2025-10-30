import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function leaveGroup(accessToken: string, group_id: number): Promise<any> {
  try {
    
    await axios.delete(`${API_URL}/api/groups/${group_id}/leave`, {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при присоединении в группу:', error);
    throw error;
  }
}
