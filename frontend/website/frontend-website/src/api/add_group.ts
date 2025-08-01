import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export async function createGroup(name: string, description: string, smallDescription: string,  city?: string, categories: number[],  isPrivate: boolean,  image: File, contacts?: string, accessToken: string): Promise<void> {
  try {
    await axios.post(`${API_URL}/api/groups/createGroup`, {
      name,
      description,
      smallDescription,
      city,
      categories,
      isPrivate,
      image,
      contacts
    }, {headers: {'Authorization': `Bearer ${accessToken}`}});
  } catch (error: any) {
    console.error('Ошибка при создании группы:', error);
    throw error;
  }
}
