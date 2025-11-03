// src/api/edit_group.ts
import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function editGroup(
  accessToken: string,
  groupId: number,
  name?: string, 
  description?: string, 
  small_description?: string,  
  city?: string, 
  categories?: number[],  
  is_private?: boolean,  
  image?: string | null, 
  contacts?: string, 
): Promise<void> {
  try {

    const data = {
      city,
      categories,
      contacts,
      description,
      image,
      is_private,
      name,
      small_description
    };

    // Убираем undefined значения
    const filteredData = Object.fromEntries(
      Object.entries(data).filter(([_, v]) => v !== undefined)
    );

    await axios.patch(`${API_URL}/api/admin/groups/${groupId}/`, filteredData, {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      }
    });
  } catch (error: any) {
    console.error('Ошибка при редактировании группы:', error);
    throw error;
  }
}