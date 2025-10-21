import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function inviteGroup(accessToken: string, group_id: number, user_id: number): Promise<any> {
  try {

    await axios.post(`${API_URL}/api/admin/groups/requestsForUser?group_id=${group_id}&user_id=${user_id}`, {},
      {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при приглашении в группу:', error);
    throw error;
  }
}
