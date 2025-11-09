import apiClient from '@/api/apiClient';
import { getTokens } from '@/api/storage/tokenStorage';
// eslint-disable-next-line import/no-unresolved
import { API_BASE_URL } from '@env';

const BASE_URL = API_BASE_URL || 'http://localhost:8080/api';

export interface Contact {
  id: string;
  name: string;
  icon: any;
  description: string;
  link: string;
}

export interface CreateGroupData {
  name: string;
  description: string;
  smallDescription: string;
  city?: string;
  categories: number[];
  isPrivate: boolean;
  image: {
    uri: string;
    name: string;
    type: string;
  } | null;
  contacts?: Contact[];
}

export interface AdminGroup {
  id: number;
  name: string;
  small_description: string;
  image: string;
  member_count: number;
  category: string[];
  type: string;
}

export interface GroupSession {
  id: number;
  title: string;
  session_type: string;
  session_place: string;
  city: string;
  start_time: string;
  duration: number;
  count_users_max: number;
  current_users: number;
  image_url: string;
  group_name: string;
  genres: string[];
}

export interface GroupApplication {
  id: number;
  userId: number;
  name: string;
  us: string;
  image: string;
}

export interface GroupDetailResponse {
  id: number;
  name: string;
  small_description: string;
  description: string;
  image: string;
  city: string;
  private: boolean;
  categories: string[];
  contacts: {
    name: string;
    link: string;
  }[];
  sessions: GroupSession[];
  applications: GroupApplication[];
}

export interface PublicGroupResponse {
  id: number;
  name: string;
  description: string;
  image: string;
  city: string;
  categories: string[];
  contacts: {
    name: string;
    link: string;
  }[];
  count_members: number;
  creater: string;
  subscription: boolean;
  users: {
    name: string;
    image: string;
  }[];
  sessions: {
    session: {
      id: number;
      title: string;
      session_type: string;
      session_place: string;
      start_time: string;
      end_time: string;
      duration: number;
      count_users_max: number;
      current_users: number;
      image_url: string;
      group_id: number;
    };
    metadata: {
      sessionID: number;
      genres: string[];
      location: string;
      year: number;
      country: string;
      ageLimit: string;
      notes: string;
      fields: Record<string, any>;
    };
  }[];
}

class GroupService {
  async createGroup(groupData: CreateGroupData): Promise<any> {
    const tokens = await getTokens();
    if (!tokens?.accessToken) {
      throw new Error('Пользователь не авторизован');
    }

    console.log('=== Создание группы ===');
    console.log('Данные группы:', {
      name: groupData.name,
      description: groupData.description,
      smallDescription: groupData.smallDescription,
      city: groupData.city,
      isPrivate: groupData.isPrivate,
      categories: groupData.categories,
      hasImage: !!groupData.image,
      contactsCount: groupData.contacts?.length || 0,
    });

    const formData = new FormData();
    
    formData.append('name', groupData.name);
    formData.append('description', groupData.description);
    formData.append('smallDescription', groupData.smallDescription);
    
    if (groupData.city) {
      formData.append('city', groupData.city);
    }
    
    formData.append('isPrivate', groupData.isPrivate.toString());

    groupData.categories.forEach(categoryId => {
      formData.append('categories', categoryId.toString());
    });

    console.log('Категории для отправки:', groupData.categories);

    if (groupData.image) {
      formData.append('image', {
        uri: groupData.image.uri,
        name: groupData.image.name,
        type: groupData.image.type,
      } as any);
      console.log('Изображение добавлено:', groupData.image.name);
    }

    if (groupData.contacts && groupData.contacts.length > 0) {
      const contactsString = groupData.contacts
        .filter(contact => contact.link.trim() !== '')
        .map(contact => {
          const name = contact.description.trim() || contact.name.toLowerCase();
          return `${name}:${contact.link.trim()}`;
        })
        .join(', ');
      
      if (contactsString) {
        formData.append('contacts', contactsString);
        console.log('Контакты для отправки:', contactsString);
      }
    }

    console.log('Отправка запроса на:', `${BASE_URL}/groups/createGroup`);

    try {
      const response = await fetch(`${BASE_URL}/groups/createGroup`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${tokens.accessToken}`,
        },
        body: formData,
      });

      console.log('Статус ответа:', response.status);

      if (!response.ok) {
        const errorText = await response.text();
        console.error('Ошибка от сервера (текст):', errorText);
        
        try {
          const errorData = JSON.parse(errorText);
          console.error('Ошибка от сервера (JSON):', errorData);
          throw new Error(errorData.message || 'Ошибка создания группы');
        } catch (parseError) {
          throw new Error(`Ошибка создания группы: ${response.status} ${response.statusText}`);
        }
      }

      const result = await response.json();
      console.log('Группа успешно создана:', result);
      return result;
    } catch (error: any) {
      console.error('Ошибка при создании группы:', error);
      throw error;
    }
  }

  async getAdminGroups(): Promise<AdminGroup[]> {
    try {
      const response = await apiClient.get('/admin/groups');
      console.log('Ответ от API (admin/groups):', response.data);

      if (!response.data || !Array.isArray(response.data)) {
        console.warn('API вернул некорректные данные:', response.data);
        return [];
      }
      
      return response.data;
    } catch (error: any) {
      console.error('Ошибка получения групп администратора:', error);
      console.error('Детали ошибки:', error.response?.data);
      throw new Error(error.response?.data?.message || 'Ошибка получения групп');
    }
  }

  async getGroupDetail(groupId: string | number): Promise<GroupDetailResponse> {
    try {
      const endpoint = `/admin/groups/${groupId}/infGroup`;
      console.log(`Загрузка детальной информации о группе ${groupId}...`);
      console.log(`Полный URL: ${BASE_URL}${endpoint}`);
      
      const response = await apiClient.get(endpoint);
      console.log('Детальная информация о группе:', response.data);
      return response.data;
    } catch (error: any) {
      console.error('Ошибка получения детальной информации о группе:', error);
      console.error('Статус:', error.response?.status);
      console.error('Детали ошибки:', error.response?.data);
      
      throw new Error(error.response?.data?.message || 'Ошибка получения информации о группе');
    }
  }
}

export default new GroupService();