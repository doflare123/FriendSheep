import axios from 'axios';
import { RequestData } from '../types/RequestData'

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getGroupApplication(accessToken: string, groupId: number): Promise<RequestData[]> {
  try {
    const response = await axios.get(`${API_URL}/api/groups/requests/${groupId}`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    return response.data as RequestData[];
  } catch (error: any) {
    console.error('Ошибка при получении группы:', error);
    
    // Более подробная информация об ошибке для отладки
    if (error.response) {
      console.error('Статус ответа:', error.response.status);
      console.error('Данные ошибки:', error.response.data);
    }
    
    throw error;
  }
}
export async function approve(accessToken: string, requestId : number): Promise<any> {
  try {

    await axios.post(`${API_URL}/api/admin/groups/requests/${requestId}/approve`, {},
      {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при принятия приглашения в группу:', error);
    throw error;
  }
}

export async function approveAll(accessToken: string, groupId : number): Promise<any> {
  try {

    await axios.post(`${API_URL}/api/admin/groups/requests/all/${groupId}/approveAll`, {},
      {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при принятии всех приглашений в группу:', error);
    throw error;
  }
}

export async function reject(accessToken: string, requestId : number): Promise<any> {
  try {

    await axios.post(`${API_URL}/api/admin/groups/requests/${requestId}/reject`, {},
      {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при отклонении приглашения в группу:', error);
    throw error;
  }
}


export async function rejectAll(accessToken: string, groupId : number): Promise<any> {
  try {

    await axios.post(`${API_URL}/api/admin/groups/requests/all/${groupId}/rejectAll`, {},
      {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при отклонении всех приглашений в группу:', error);
    throw error;
  }
}
