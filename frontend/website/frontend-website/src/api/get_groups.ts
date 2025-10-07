import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getGroups(accessToken: string, id?: number): Promise<any> {
  try {
    
    let response
    if(id)
      response = await axios.get(`${API_URL}/api/users/subscriptions?id=${id}`, {headers: {'Authorization': `Bearer ${accessToken}`}});
    else
      response = await axios.get(`${API_URL}/api/users/subscriptions`, {headers: {'Authorization': `Bearer ${accessToken}`}});

    return response.data;
  } catch (error: any) {
    console.error('Ошибка при получении группы:', error);
    throw error;
  }
}
