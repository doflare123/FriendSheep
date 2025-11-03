import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function approveInvite(accessToken: string, requestId : number): Promise<any> {
  try {

    await axios.put(`${API_URL}/api/users/invites/${requestId}/approve`, {},
      {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при принятии приглашения в группу:', error);
    throw error;
  }
}
