import axios from 'axios';
import { GroupData } from '../types/Groups';
import { convertCategRuToEng, convertSessionsToEventCards } from '../Constants'

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function getGroupInfoAdmin(accessToken: string, groupId: number): Promise<GroupData> {
  try {
    const response = await axios.get(`${API_URL}/api/admin/groups/${groupId}/infGroup`, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });

    const returnData: GroupData = {
      id: response.data.id || groupId,
      name: response.data.name,
      description: response.data.description,
      city: response.data.city,
      categories: convertCategRuToEng(response.data.categories),
      contacts: response.data.contacts || [],
      image: response.data.image,
      count_members: response.data.count_members || 0,
      users: response.data.users || [],
      sessions: convertSessionsToEventCards(response.data.sessions) || [],
      small_description: response.data.small_description,
      private: response.data.private
    }

    console.log("returnData", returnData);
    return returnData;
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