import axios from 'axios';
import { RequestData } from '../types/RequestData'

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getGroupApplication(accessToken: string): Promise<RequestData> {
  try {
    const response = await axios.get(`${API_URL}/api/groups/requests`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    return response.data.requests as RequestData;
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

export async function approveApplication(accessToken: string, requestId: number): Promise<void> {
  try {
    await axios.get(`${API_URL}/api/admin/groups/requests/${requestId}/approve`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });
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

export async function rejectApplication(accessToken: string, requestId: number): Promise<void> {
  try {
    await axios.get(`${API_URL}/api/admin/groups/requests/${requestId}/reject`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });
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