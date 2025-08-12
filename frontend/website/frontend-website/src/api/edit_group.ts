// src/api/edit_group.ts
import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function editGroup(
  name: string, 
  description: string, 
  smallDescription: string,  
  city?: string, 
  categories: number[],  
  isPrivate: boolean,  
  image?: File | null, 
  contacts?: string, 
  accessToken: string,
  groupId: number,
): Promise<void> {
  try {

    await axios.patch(`${API_URL}/api/admin/groups/${groupId}/`, {
        city,
        contacts,
        description,
        image,
        isPrivate,
        name,
        smallDescription
    }, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      }
    });
  } catch (error: any) {
    console.error('Ошибка при редактировании группы:', error);
    throw error;
  }
}