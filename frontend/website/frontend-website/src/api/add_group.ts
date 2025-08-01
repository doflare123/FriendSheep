import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function createGroup(
  name: string, 
  description: string, 
  smallDescription: string,  
  city?: string, 
  categories: number[],  
  isPrivate: boolean,  
  image?: File | null, 
  contacts?: string, 
  accessToken: string
): Promise<void> {
  try {
    // Создаем FormData один раз
    const data = new FormData();
    data.append('name', name);
    data.append('description', description);
    data.append('smallDescription', smallDescription);
    
    if (city) data.append('city', city);
    if (contacts) data.append('contacts', contacts);
    if (image) data.append('image', image);
    
    categories.forEach(category => data.append('categories', category.toString()));
    data.append('isPrivate', isPrivate.toString());

    await axios.post(`${API_URL}/api/groups/createGroup`, data, {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    });
  } catch (error: any) {
    console.error('Ошибка при создании группы:', error);
    throw error;
  }
}