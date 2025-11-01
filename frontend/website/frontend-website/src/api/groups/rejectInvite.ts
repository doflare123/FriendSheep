import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function rejectInvite(accessToken: string, requestId : number): Promise<any> {
  try {

    await axios.put(`${API_URL}/api/users/invites/${requestId}/reject`, {},
      {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при отказе от приглашния в группу:', error);
    throw error;
  }
}
