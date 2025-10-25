import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function approveAll(accessToken: string, groupId : number): Promise<any> {
  try {

    await axios.post(`${API_URL}/api/admin/groups/requests/all/${groupId}/approveAll`, {},
      {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при приглашении в группу:', error);
    throw error;
  }
}
