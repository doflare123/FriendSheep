import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function joinGroup(accessToken: string, group_id: number): Promise<any> {
  try {
    
    await axios.post(`${API_URL}/api/groups/joinToGroup`, {groupId: group_id}, {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при присоединении в группу:', error);
    throw error;
  }
}
