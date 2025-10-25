import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function reject(accessToken: string, requestId : number): Promise<any> {
  try {

    await axios.post(`${API_URL}/api/admin/groups/requests/${requestId}/reject`, {},
      {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при приглашении в группу:', error);
    throw error;
  }
}
